package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"defense-allies-server/serverapp/timesquare/middleware"

	"github.com/redis/go-redis/v9"
)

// UserCreationHandler 신규 유저 생성 로직을 위한 핸들러 인터페이스
type UserCreationHandler func(userInfo *middleware.UserInfo) (*middleware.User, error)

// RedisUserService Redis 기반 유저 서비스 구현
type RedisUserService struct {
	redis               *redis.Client
	userCreationHandler UserCreationHandler
}

// Redis 키 패턴 상수들
const (
	UserKeyPrefix      = "user:"
	UserIndexKey       = "users:index"
	UserLastLoginKey   = "users:last_login"
	UserRolesKeyPrefix = "user:roles:"
	UserEmailIndexKey  = "users:email_index"
	UserNameIndexKey   = "users:username_index"
)

// NewRedisUserService 새로운 Redis 유저 서비스 생성
func NewRedisUserService(redisClient *redis.Client, creationHandler UserCreationHandler) middleware.UserService {
	return &RedisUserService{
		redis:               redisClient,
		userCreationHandler: creationHandler,
	}
}

// DefaultUserCreationHandler 기본 유저 생성 핸들러
func DefaultUserCreationHandler(userInfo *middleware.UserInfo) (*middleware.User, error) {
	now := time.Now()

	// 기본 게임 데이터 생성
	defaultGameData := &middleware.GameData{
		Level: 1,
		Score: 0,
		Resources: map[string]int64{
			"gold":   1000,
			"gems":   50,
			"energy": 100,
		},
		Settings: map[string]string{
			"language":      "en",
			"sound":         "on",
			"notifications": "on",
		},
	}

	return &middleware.User{
		ID:        userInfo.ID,
		Username:  userInfo.Username,
		Email:     userInfo.Email,
		CreatedAt: now,
		UpdatedAt: now,
		LastLogin: now,
		GameData:  defaultGameData,
	}, nil
}

// GetUser 유저 정보 조회
func (us *RedisUserService) GetUser(ctx context.Context, userID string) (*middleware.User, error) {
	userKey := UserKeyPrefix + userID

	// Redis에서 유저 데이터 조회
	userData, err := us.redis.HGetAll(ctx, userKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user from redis: %w", err)
	}

	if len(userData) == 0 {
		return nil, fmt.Errorf("user not found: %s", userID)
	}

	// 유저 객체 생성
	user := &middleware.User{
		ID:       userData["id"],
		Username: userData["username"],
		Email:    userData["email"],
	}

	// 시간 필드 파싱
	if createdAt, err := time.Parse(time.RFC3339, userData["created_at"]); err == nil {
		user.CreatedAt = createdAt
	}
	if updatedAt, err := time.Parse(time.RFC3339, userData["updated_at"]); err == nil {
		user.UpdatedAt = updatedAt
	}
	if lastLogin, err := time.Parse(time.RFC3339, userData["last_login"]); err == nil {
		user.LastLogin = lastLogin
	}

	// 게임 데이터 파싱
	if gameDataStr, exists := userData["game_data"]; exists && gameDataStr != "" {
		var gameData middleware.GameData
		if err := json.Unmarshal([]byte(gameDataStr), &gameData); err == nil {
			user.GameData = &gameData
		}
	}

	return user, nil
}

// FindAndUpsertUser 유저 조회 후 없으면 생성 (낙관적 동시성 제어)
func (us *RedisUserService) FindAndUpsertUser(ctx context.Context, userInfo *middleware.UserInfo) (*middleware.User, error) {
	const maxRetries = 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		// 1. 먼저 유저 조회 시도
		if user, err := us.GetUser(ctx, userInfo.ID); err == nil {
			// 기존 유저 발견 - 마지막 로그인 시간 업데이트
			if updateErr := us.UpdateLastLogin(ctx, userInfo.ID); updateErr != nil {
				fmt.Printf("Failed to update last login for user %s: %v\n", userInfo.ID, updateErr)
			}
			return user, nil
		}

		// 2. 유저가 없으면 생성 시도 (낙관적 동시성 제어)
		user, err := us.attemptCreateUser(ctx, userInfo)
		if err == nil {
			return user, nil
		}

		// 3. 동시성 충돌 발생 시 재시도
		if isConflictError(err) {
			fmt.Printf("Conflict detected on attempt %d for user %s, retrying...\n", attempt+1, userInfo.ID)
			time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond) // 지수 백오프
			continue
		}

		// 4. 다른 에러는 즉시 반환
		return nil, err
	}

	return nil, fmt.Errorf("failed to create user after %d attempts: %s", maxRetries, userInfo.ID)
}

// CreateUser 기존 인터페이스 호환성을 위한 래퍼
func (us *RedisUserService) CreateUser(ctx context.Context, userInfo *middleware.UserInfo) (*middleware.User, error) {
	return us.FindAndUpsertUser(ctx, userInfo)
}

// attemptCreateUser 유저 생성 시도 (원자적 연산)
func (us *RedisUserService) attemptCreateUser(ctx context.Context, userInfo *middleware.UserInfo) (*middleware.User, error) {
	userKey := UserKeyPrefix + userInfo.ID

	// 외부 핸들러로 유저 생성 로직 위임 (IoC)
	user, err := us.userCreationHandler(userInfo)
	if err != nil {
		return nil, fmt.Errorf("user creation handler failed: %w", err)
	}

	// 게임 데이터 JSON 변환
	gameDataJSON, err := json.Marshal(user.GameData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal game data: %w", err)
	}

	// 유저 데이터 준비
	userData := map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt.Format(time.RFC3339),
		"updated_at": user.UpdatedAt.Format(time.RFC3339),
		"last_login": user.LastLogin.Format(time.RFC3339),
		"game_data":  string(gameDataJSON),
	}

	// Redis WATCH를 사용한 낙관적 동시성 제어
	err = us.redis.Watch(ctx, func(tx *redis.Tx) error {
		// 유저가 이미 존재하는지 확인
		exists, err := tx.Exists(ctx, userKey).Result()
		if err != nil {
			return err
		}

		if exists > 0 {
			return fmt.Errorf("user already exists: %s", userInfo.ID)
		}

		// 이메일 중복 확인
		emailExists, err := tx.HExists(ctx, UserEmailIndexKey, user.Email).Result()
		if err != nil {
			return err
		}
		if emailExists {
			return fmt.Errorf("email already exists: %s", user.Email)
		}

		// 유저명 중복 확인
		usernameExists, err := tx.HExists(ctx, UserNameIndexKey, user.Username).Result()
		if err != nil {
			return err
		}
		if usernameExists {
			return fmt.Errorf("username already exists: %s", user.Username)
		}

		// 트랜잭션 파이프라인 시작
		pipe := tx.TxPipeline()

		// 유저 데이터 저장
		pipe.HMSet(ctx, userKey, userData)

		// 인덱스 업데이트
		pipe.SAdd(ctx, UserIndexKey, user.ID)
		pipe.HSet(ctx, UserEmailIndexKey, user.Email, user.ID)
		pipe.HSet(ctx, UserNameIndexKey, user.Username, user.ID)
		pipe.ZAdd(ctx, UserLastLoginKey, redis.Z{
			Score:  float64(user.LastLogin.Unix()),
			Member: user.ID,
		})

		// 유저 역할 저장
		if len(userInfo.Roles) > 0 {
			rolesKey := UserRolesKeyPrefix + user.ID
			for _, role := range userInfo.Roles {
				pipe.SAdd(ctx, rolesKey, role)
			}
		}

		// 파이프라인 실행
		_, err = pipe.Exec(ctx)
		return err
	}, userKey, UserEmailIndexKey, UserNameIndexKey)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// isConflictError 동시성 충돌 에러인지 확인
func isConflictError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return redis.TxFailedErr == err ||
		strings.Contains(errStr, "already exists") ||
		strings.Contains(errStr, "WATCH")
}

// UpdateLastLogin 마지막 로그인 시간 업데이트
func (us *RedisUserService) UpdateLastLogin(ctx context.Context, userID string) error {
	userKey := UserKeyPrefix + userID
	now := time.Now()

	// Redis 파이프라인 사용
	pipe := us.redis.Pipeline()

	// 유저 데이터의 last_login과 updated_at 업데이트
	pipe.HMSet(ctx, userKey, map[string]interface{}{
		"last_login": now.Format(time.RFC3339),
		"updated_at": now.Format(time.RFC3339),
	})

	// 마지막 로그인 시간 인덱스 업데이트
	pipe.ZAdd(ctx, UserLastLoginKey, redis.Z{
		Score:  float64(now.Unix()),
		Member: userID,
	})

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// GetUserGameData 유저 게임 데이터 조회
func (us *RedisUserService) GetUserGameData(ctx context.Context, userID string) (*middleware.GameData, error) {
	userKey := UserKeyPrefix + userID

	gameDataStr, err := us.redis.HGet(ctx, userKey, "game_data").Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("user not found: %s", userID)
		}
		return nil, fmt.Errorf("failed to get game data: %w", err)
	}

	if gameDataStr == "" {
		return nil, fmt.Errorf("no game data found for user: %s", userID)
	}

	var gameData middleware.GameData
	if err := json.Unmarshal([]byte(gameDataStr), &gameData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal game data: %w", err)
	}

	return &gameData, nil
}

// UpdateUserGameData 유저 게임 데이터 업데이트
func (us *RedisUserService) UpdateUserGameData(ctx context.Context, userID string, gameData *middleware.GameData) error {
	gameDataJSON, err := json.Marshal(gameData)
	if err != nil {
		return fmt.Errorf("failed to marshal game data: %w", err)
	}

	userKey := UserKeyPrefix + userID
	now := time.Now()

	// 게임 데이터와 업데이트 시간 저장
	err = us.redis.HMSet(ctx, userKey, map[string]interface{}{
		"game_data":  string(gameDataJSON),
		"updated_at": now.Format(time.RFC3339),
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to update game data: %w", err)
	}

	return nil
}

// GetUsersByLastLogin 마지막 로그인 기준으로 유저 조회
func (us *RedisUserService) GetUsersByLastLogin(ctx context.Context, since time.Time, limit int) ([]*middleware.User, error) {
	// 정렬된 셋에서 시간 범위로 유저 ID 조회
	sinceScore := float64(since.Unix())
	nowScore := float64(time.Now().Unix())

	userIDs, err := us.redis.ZRevRangeByScore(ctx, UserLastLoginKey, &redis.ZRangeBy{
		Min:   strconv.FormatFloat(sinceScore, 'f', 0, 64),
		Max:   strconv.FormatFloat(nowScore, 'f', 0, 64),
		Count: int64(limit),
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to query user IDs: %w", err)
	}

	// 각 유저 정보 조회
	var users []*middleware.User
	for _, userID := range userIDs {
		user, err := us.GetUser(ctx, userID)
		if err != nil {
			// 개별 유저 조회 실패는 로그만 남기고 계속 진행
			fmt.Printf("Failed to get user %s: %v\n", userID, err)
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUserByEmail 이메일로 유저 조회
func (us *RedisUserService) GetUserByEmail(ctx context.Context, email string) (*middleware.User, error) {
	userID, err := us.redis.HGet(ctx, UserEmailIndexKey, email).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("user not found with email: %s", email)
		}
		return nil, fmt.Errorf("failed to get user ID by email: %w", err)
	}

	return us.GetUser(ctx, userID)
}

// GetUserByUsername 유저명으로 유저 조회
func (us *RedisUserService) GetUserByUsername(ctx context.Context, username string) (*middleware.User, error) {
	userID, err := us.redis.HGet(ctx, UserNameIndexKey, username).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("user not found with username: %s", username)
		}
		return nil, fmt.Errorf("failed to get user ID by username: %w", err)
	}

	return us.GetUser(ctx, userID)
}

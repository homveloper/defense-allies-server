package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"defense-allies-server/pkg/gameauth/domain/authsession"

	"github.com/redis/go-redis/v9"
)

type RedisAuthSessionRepository struct {
	client *redis.Client
}

func NewRedisAuthSessionRepository(client *redis.Client) *RedisAuthSessionRepository {
	return &RedisAuthSessionRepository{
		client: client,
	}
}

const (
	authSessionKeyPrefix           = "gameauth:session:"
	authSessionByTokenPrefix       = "gameauth:session:token:"
	authSessionByRefreshPrefix     = "gameauth:session:refresh:"
	authSessionByGameAccountPrefix = "gameauth:session:gameaccount:"
	sessionTTL                     = 7 * 24 * time.Hour // 7일
)

func (r *RedisAuthSessionRepository) Save(ctx context.Context, session *authsession.AuthSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal auth session: %w", err)
	}

	pipe := r.client.Pipeline()

	// 메인 키로 저장 (TTL 적용)
	mainKey := authSessionKeyPrefix + session.ID()
	pipe.Set(ctx, mainKey, data, sessionTTL)

	// SessionToken으로 인덱싱 (TTL 적용)
	tokenKey := authSessionByTokenPrefix + session.SessionToken()
	pipe.Set(ctx, tokenKey, session.ID(), sessionTTL)

	// RefreshToken으로 인덱싱 (TTL 적용)
	refreshKey := authSessionByRefreshPrefix + session.RefreshToken()
	pipe.Set(ctx, refreshKey, session.ID(), sessionTTL)

	// GameAccountID로 인덱싱 (Set 사용하여 여러 세션 관리)
	gameAccountKey := authSessionByGameAccountPrefix + session.GameAccountID()
	pipe.SAdd(ctx, gameAccountKey, session.ID())
	pipe.Expire(ctx, gameAccountKey, sessionTTL)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save auth session: %w", err)
	}

	return nil
}

func (r *RedisAuthSessionRepository) Load(ctx context.Context, id string) (*authsession.AuthSession, error) {
	key := authSessionKeyPrefix + id
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load auth session: %w", err)
	}

	var session authsession.AuthSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal auth session: %w", err)
	}

	return &session, nil
}

func (r *RedisAuthSessionRepository) FindBySessionToken(ctx context.Context, sessionToken string) (*authsession.AuthSession, error) {
	key := authSessionByTokenPrefix + sessionToken
	sessionID, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find session by token: %w", err)
	}

	return r.Load(ctx, sessionID)
}

func (r *RedisAuthSessionRepository) FindByRefreshToken(ctx context.Context, refreshToken string) (*authsession.AuthSession, error) {
	key := authSessionByRefreshPrefix + refreshToken
	sessionID, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find session by refresh token: %w", err)
	}

	return r.Load(ctx, sessionID)
}

func (r *RedisAuthSessionRepository) FindActiveByGameAccountID(ctx context.Context, gameAccountID string) ([]*authsession.AuthSession, error) {
	key := authSessionByGameAccountPrefix + gameAccountID
	sessionIDs, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return []*authsession.AuthSession{}, nil
		}
		return nil, fmt.Errorf("failed to find sessions by game account ID: %w", err)
	}

	var sessions []*authsession.AuthSession
	for _, sessionID := range sessionIDs {
		session, err := r.Load(ctx, sessionID)
		if err != nil {
			continue // 만료된 세션은 스킵
		}
		if session != nil && session.IsActive() {
			sessions = append(sessions, session)
		} else if session != nil {
			// 비활성 세션은 Set에서 제거
			r.client.SRem(ctx, key, sessionID)
		}
	}

	return sessions, nil
}

func (r *RedisAuthSessionRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	// Redis TTL을 사용하므로 만료된 키들은 자동으로 삭제됨
	// 추가적인 정리가 필요한 경우에만 구현
	return nil
}

func (r *RedisAuthSessionRepository) Delete(ctx context.Context, id string) error {
	// 먼저 AuthSession 로드하여 인덱스 키들을 가져옴
	session, err := r.Load(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to load auth session for deletion: %w", err)
	}
	if session == nil {
		return nil // 이미 존재하지 않음
	}

	pipe := r.client.Pipeline()

	// 메인 키 삭제
	mainKey := authSessionKeyPrefix + id
	pipe.Del(ctx, mainKey)

	// SessionToken 인덱스 삭제
	tokenKey := authSessionByTokenPrefix + session.SessionToken()
	pipe.Del(ctx, tokenKey)

	// RefreshToken 인덱스 삭제
	refreshKey := authSessionByRefreshPrefix + session.RefreshToken()
	pipe.Del(ctx, refreshKey)

	// GameAccountID Set에서 제거
	gameAccountKey := authSessionByGameAccountPrefix + session.GameAccountID()
	pipe.SRem(ctx, gameAccountKey, id)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete auth session: %w", err)
	}

	return nil
}

func (r *RedisAuthSessionRepository) Exists(ctx context.Context, id string) (bool, error) {
	key := authSessionKeyPrefix + id
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check auth session existence: %w", err)
	}
	return exists > 0, nil
}

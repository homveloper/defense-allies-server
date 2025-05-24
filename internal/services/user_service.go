package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"defense-allies-server/internal/models"
	"defense-allies-server/pkg/redis"
)

// PlayerService 플레이어 비즈니스 로직을 처리합니다
type PlayerService struct {
	redisClient  *redis.Client
	eventService *EventService
}

// NewPlayerService 새로운 플레이어 서비스를 생성합니다
func NewPlayerService(redisClient *redis.Client, eventService *EventService) *PlayerService {
	return &PlayerService{
		redisClient:  redisClient,
		eventService: eventService,
	}
}

// generatePlayerID 새로운 플레이어 ID를 생성합니다
func (s *PlayerService) generatePlayerID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "player_" + hex.EncodeToString(bytes)
}

// CreatePlayer 새로운 플레이어를 생성합니다
func (s *PlayerService) CreatePlayer(req *models.CreatePlayerRequest) (*models.PlayerResponse, error) {
	// 유효성 검사
	if req.Username == "" {
		return nil, errors.New("username is required")
	}
	if req.Email == "" {
		return nil, errors.New("email is required")
	}
	if req.Password == "" {
		return nil, errors.New("password is required")
	}

	// 이메일 중복 확인
	exists, err := s.redisClient.Exists(fmt.Sprintf("player:email:%s", req.Email))
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists > 0 {
		return nil, errors.New("email already exists")
	}

	// 새 플레이어 생성
	playerID := s.generatePlayerID()
	now := time.Now()

	player := &models.Player{
		ID:        playerID,
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password, // TODO: 해싱 필요
		CreatedAt: now,
		UpdatedAt: now,
		LastSeen:  now,
		IsOnline:  false,
	}

	// Redis에 저장
	playerKey := fmt.Sprintf("player:%s", playerID)
	emailKey := fmt.Sprintf("player:email:%s", req.Email)

	// 플레이어 정보 저장
	if err := s.redisClient.HSet(playerKey,
		"id", player.ID,
		"username", player.Username,
		"email", player.Email,
		"password", player.Password,
		"created_at", player.CreatedAt.Format(time.RFC3339),
		"updated_at", player.UpdatedAt.Format(time.RFC3339),
		"last_seen", player.LastSeen.Format(time.RFC3339),
		"is_online", "false",
	); err != nil {
		return nil, fmt.Errorf("failed to save player: %w", err)
	}

	// 이메일 인덱스 저장
	if err := s.redisClient.Set(emailKey, playerID, 0); err != nil {
		return nil, fmt.Errorf("failed to save email index: %w", err)
	}

	// 응답 생성
	response := &models.PlayerResponse{
		ID:        player.ID,
		Username:  player.Username,
		Email:     player.Email,
		CreatedAt: player.CreatedAt,
		UpdatedAt: player.UpdatedAt,
		LastSeen:  player.LastSeen,
		IsOnline:  player.IsOnline,
	}

	// 플레이어 생성 이벤트 발행
	s.eventService.PublishToRedis("events:player", Event{
		Type: "player_created",
		Data: map[string]interface{}{
			"player_id": playerID,
			"username":  req.Username,
		},
		PlayerID: playerID,
	})

	return response, nil
}

// GetPlayerByID ID로 플레이어를 조회합니다
func (s *PlayerService) GetPlayerByID(playerID string) (*models.PlayerResponse, error) {
	playerKey := fmt.Sprintf("player:%s", playerID)

	data, err := s.redisClient.HGetAll(playerKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	if len(data) == 0 {
		return nil, errors.New("player not found")
	}

	// 시간 파싱
	createdAt, _ := time.Parse(time.RFC3339, data["created_at"])
	updatedAt, _ := time.Parse(time.RFC3339, data["updated_at"])
	lastSeen, _ := time.Parse(time.RFC3339, data["last_seen"])

	return &models.PlayerResponse{
		ID:        data["id"],
		Username:  data["username"],
		Email:     data["email"],
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		LastSeen:  lastSeen,
		IsOnline:  data["is_online"] == "true",
	}, nil
}

// GetPlayerByEmail 이메일로 플레이어를 조회합니다
func (s *PlayerService) GetPlayerByEmail(email string) (*models.PlayerResponse, error) {
	emailKey := fmt.Sprintf("player:email:%s", email)

	playerID, err := s.redisClient.Get(emailKey)
	if err != nil {
		return nil, errors.New("player not found")
	}

	return s.GetPlayerByID(playerID)
}

// UpdatePlayerOnlineStatus 플레이어 온라인 상태를 업데이트합니다
func (s *PlayerService) UpdatePlayerOnlineStatus(playerID string, isOnline bool) error {
	playerKey := fmt.Sprintf("player:%s", playerID)
	now := time.Now()

	updates := map[string]interface{}{
		"is_online":  fmt.Sprintf("%t", isOnline),
		"updated_at": now.Format(time.RFC3339),
	}

	if isOnline {
		updates["last_seen"] = now.Format(time.RFC3339)
	}

	if err := s.redisClient.HSet(playerKey, updates); err != nil {
		return fmt.Errorf("failed to update player status: %w", err)
	}

	// 상태 변경 이벤트 발행
	s.eventService.PublishToRedis("events:player", Event{
		Type: "player_status_changed",
		Data: map[string]interface{}{
			"player_id": playerID,
			"is_online": isOnline,
			"timestamp": now.Unix(),
		},
		PlayerID: playerID,
	})

	return nil
}

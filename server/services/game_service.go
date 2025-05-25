package services

import (
	"defense-allies-server/pkg/redis"
)

// GameService 게임 로직을 처리하는 서비스
type GameService struct {
	redisClient  *redis.Client
	eventService *EventService
}

// NewGameService 새로운 게임 서비스를 생성합니다
func NewGameService(redisClient *redis.Client, eventService *EventService) *GameService {
	return &GameService{
		redisClient:  redisClient,
		eventService: eventService,
	}
}

// TODO: 게임 로직 구현

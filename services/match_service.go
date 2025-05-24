package services

import (
	"defense-allies-server/pkg/redis"
)

// MatchService 매치메이킹을 처리하는 서비스
type MatchService struct {
	redisClient  *redis.Client
	eventService *EventService
}

// NewMatchService 새로운 매치 서비스를 생성합니다
func NewMatchService(redisClient *redis.Client, eventService *EventService) *MatchService {
	return &MatchService{
		redisClient:  redisClient,
		eventService: eventService,
	}
}

// TODO: 매치메이킹 로직 구현

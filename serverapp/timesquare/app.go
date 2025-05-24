package timesquare

import (
	"context"
	"log"
	"net/http"

	"defense-allies-server/pkg/redis"
	"defense-allies-server/serverapp"
)

// TimeSquareApp 게임 서버 - 모든 플레이어가 모이는 활동 중심지
type TimeSquareApp struct {
	*serverapp.BaseApp
	redisClient *redis.Client
}

// NewTimeSquareApp 새로운 TimeSquareApp을 생성합니다
func NewTimeSquareApp(redisClient *redis.Client) *TimeSquareApp {
	app := &TimeSquareApp{
		BaseApp:     serverapp.NewBaseApp("timesquare"),
		redisClient: redisClient,
	}

	// 핸들러 초기화

	return app
}

// RegisterRoutes HTTP Mux에 라우트를 등록합니다
func (t *TimeSquareApp) RegisterRoutes(mux *http.ServeMux) {

	log.Printf("[TimeSquare] Routes registered")
}

// onStart 시작 시 호출되는 훅
func (t *TimeSquareApp) onStart(ctx context.Context) error {

	// Redis 연결 테스트
	if t.redisClient != nil {
		if _, err := t.redisClient.Get("timesquare:startup"); err != nil && err.Error() != "redis: nil" {
			return err
		}
	}

	log.Printf("[TimeSquare] Started successfully - The city never sleeps! 🏙️")
	return nil
}

// onStop 종료 시 호출되는 훅
func (t *TimeSquareApp) onStop(ctx context.Context) error {
	log.Printf("[TimeSquare] Shutting down - Clearing the square... 🌃")

	return nil
}

// Health 서버앱의 상태를 확인합니다
func (t *TimeSquareApp) Health() serverapp.HealthStatus {
	baseHealth := t.BaseApp.Health()

	// TimeSquare 특화 헬스체크
	if t.redisClient == nil {
		baseHealth.Status = serverapp.HealthStatusUnhealthy
		baseHealth.Message = "Redis client not available"
		return baseHealth
	}

	return baseHealth
}

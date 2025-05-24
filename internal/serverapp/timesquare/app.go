package timesquare

import (
	"context"
	"log"
	"net/http"

	"defense-allies-server/internal/serverapp"
	"defense-allies-server/internal/services"
	"defense-allies-server/pkg/redis"
)

// TimeSquareApp 게임 서버 - 모든 플레이어가 모이는 활동 중심지
type TimeSquareApp struct {
	*serverapp.BaseApp
	redisClient  *redis.Client
	eventService *services.EventService
	gameService  *services.GameService
	matchService *services.MatchService
}

// NewTimeSquareApp 새로운 TimeSquareApp을 생성합니다
func NewTimeSquareApp(redisClient *redis.Client) *TimeSquareApp {
	app := &TimeSquareApp{
		BaseApp:     serverapp.NewBaseApp("timesquare"),
		redisClient: redisClient,
	}

	// 서비스 초기화
	app.eventService = services.NewEventService(redisClient)
	app.gameService = services.NewGameService(redisClient, app.eventService)
	app.matchService = services.NewMatchService(redisClient, app.eventService)

	return app
}

// RegisterRoutes HTTP Mux에 라우트를 등록합니다
func (t *TimeSquareApp) RegisterRoutes(mux *http.ServeMux) {
	// 매치메이킹 라우트
	mux.HandleFunc("/api/v1/game/match/join", t.handleMatchJoin)
	mux.HandleFunc("/api/v1/game/match/status", t.handleMatchStatus)
	mux.HandleFunc("/api/v1/game/match/leave", t.handleMatchLeave)

	// 게임 세션 라우트
	mux.HandleFunc("/api/v1/game/", t.handleGameRoutes)

	// SSE 이벤트 라우트
	mux.HandleFunc("/api/v1/events/subscribe", t.eventService.HandleSSE)

	log.Printf("[TimeSquare] Routes registered")
}

// onStart 시작 시 호출되는 훅
func (t *TimeSquareApp) onStart(ctx context.Context) error {
	// 이벤트 서비스 시작
	t.eventService.Start()

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
	
	// 진행 중인 게임 세션 정리
	if t.gameService != nil {
		// TODO: 게임 세션 정리 로직
	}

	// 매치메이킹 큐 정리
	if t.matchService != nil {
		// TODO: 매치메이킹 큐 정리 로직
	}

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

	// 활성 게임 수 확인
	activeGames := t.getActiveGameCount()
	if baseHealth.Details == nil {
		baseHealth.Details = make(map[string]string)
	}
	baseHealth.Details["active_games"] = string(rune(activeGames))
	baseHealth.Details["sse_clients"] = string(rune(t.eventService.GetClientCount()))

	// 게임 서버 상태 메시지
	if activeGames > 0 {
		baseHealth.Message = "TimeSquare is bustling with activity! 🎮"
	} else {
		baseHealth.Message = "TimeSquare is ready for players 🏙️"
	}

	return baseHealth
}

// getActiveGameCount 활성 게임 수를 반환합니다
func (t *TimeSquareApp) getActiveGameCount() int {
	// TODO: Redis에서 활성 게임 수 조회
	return 0
}

// handleMatchJoin 매치 참가 핸들러
func (t *TimeSquareApp) handleMatchJoin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: 매치 참가 로직 구현
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "joined", "message": "Welcome to TimeSquare! 🏙️"}`))
}

// handleMatchStatus 매치 상태 핸들러
func (t *TimeSquareApp) handleMatchStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: 매치 상태 조회 로직 구현
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "waiting", "queue_position": 1}`))
}

// handleMatchLeave 매치 떠나기 핸들러
func (t *TimeSquareApp) handleMatchLeave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: 매치 떠나기 로직 구현
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "left", "message": "See you later! 👋"}`))
}

// handleGameRoutes 게임 관련 라우트 핸들러
func (t *TimeSquareApp) handleGameRoutes(w http.ResponseWriter, r *http.Request) {
	// URL 파싱하여 게임 ID 추출
	// TODO: 게임 라우팅 로직 구현
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Game route - Coming soon! 🎮"}`))
}

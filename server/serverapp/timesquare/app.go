package timesquare

import (
	"context"
	"log"
	"net/http"
	"time"

	"defense-allies-server/pkg/gameauth/api"
	"defense-allies-server/pkg/gameauth/application/auth"
	"defense-allies-server/pkg/gameauth/application/providers"
	"defense-allies-server/pkg/gameauth/application/providers/guest"
	"defense-allies-server/pkg/gameauth/infrastructure/repositories"
	"defense-allies-server/pkg/gameauth/infrastructure/uuid"
	"defense-allies-server/serverapp"
	"defense-allies-server/serverapp/timesquare/gamedata"
	"defense-allies-server/serverapp/timesquare/handlers"
	"defense-allies-server/serverapp/timesquare/middleware"
	"defense-allies-server/serverapp/timesquare/service"

	redisClient "github.com/redis/go-redis/v9"
)

// TimeSquareApp 게임 서버 - 모든 플레이어가 모이는 활동 중심지
type TimeSquareApp struct {
	*serverapp.BaseApp
	config         *Config
	redisClient    *redisClient.Client
	gameDataRepo   gamedata.Repository
	userService    middleware.UserService
	authService    *auth.Service
	authMiddleware *middleware.AuthMiddleware
	gameHandler    *handlers.GameHandler
}

// NewTimeSquareApp 새로운 TimeSquareApp을 생성합니다 (설정 파일 경로로 생성)
func NewTimeSquareApp(configPath string) (*TimeSquareApp, error) {
	// TimeSquare 설정 로드
	config, err := NewConfigFromFile(configPath)
	if err != nil {
		return nil, err
	}

	// 설정 유효성 검사
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// 공개키 파싱
	publicKey, err := config.ParsePublicKey()
	if err != nil {
		return nil, err
	}

	// 사용자 데이터용 Redis 클라이언트 생성
	redisClient, err := config.CreateRedisClientForTopic("users")
	if err != nil {
		return nil, err
	}

	app := &TimeSquareApp{
		BaseApp:      serverapp.NewBaseApp("timesquare"),
		config:       config,
		redisClient:  redisClient,
		gameDataRepo: gamedata.GetRepository(),
	}

	// gameauth Redis 레포지토리 초기화
	authRedisClient, err := config.CreateRedisClientForTopic("sessions")
	if err != nil {
		return nil, err
	}
	accountRepo := repositories.NewRedisAccountLinkRepository(authRedisClient)
	sessionRepo := repositories.NewRedisAuthSessionRepository(authRedisClient)

	// UUID 생성기 초기화
	idGenerator := uuid.NewUUIDGenerator()

	// Provider Registry 초기화
	providerRegistry := providers.NewProviderRegistry()

	// Guest Provider 등록
	guestProvider := guest.NewProvider(accountRepo)
	providerRegistry.Register(guestProvider)

	// gameauth 서비스 초기화 (7일 세션 TTL)
	sessionTTL := 7 * 24 * time.Hour
	app.authService = auth.NewService(providerRegistry, accountRepo, sessionRepo, idGenerator, sessionTTL)

	// 유저 생성 핸들러 (게임 데이터 레포지토리 사용)
	userCreationHandler := func(userInfo *middleware.UserInfo) (*middleware.User, error) {
		defaults := app.gameDataRepo.GetNewUserDefaults()
		return &middleware.User{
			ID:        userInfo.ID,
			Username:  userInfo.Username,
			Email:     userInfo.Email,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			LastLogin: time.Now(),
			GameData: &middleware.GameData{
				Level:     defaults.Level,
				Score:     defaults.Score,
				Resources: defaults.Resources,
				Settings:  defaults.Settings,
			},
		}, nil
	}

	// 유저 서비스 초기화 (Redis 기반)
	app.userService = service.NewRedisUserService(redisClient, userCreationHandler)

	// 인증 미들웨어 초기화 (gameauth 서비스 포함)
	app.authMiddleware = middleware.NewAuthMiddleware(publicKey, config.GetGuardianURL(), app.userService, app.authService)

	// 게임 핸들러 초기화
	app.gameHandler = handlers.NewGameHandler(app.userService)

	return app, nil
}

// RegisterRoutes HTTP Mux에 라우트를 등록합니다
func (t *TimeSquareApp) RegisterRoutes(mux *http.ServeMux) {
	// 헬스체크 엔드포인트 (인증 불필요)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"timesquare"}`))
	})

	// gameauth API 라우트 등록 (인증 불필요)
	api.SetupRoutes(mux, t.authService)

	// 게임 API 라우트 (인증 필요)
	// 프로필 조회
	mux.Handle("/api/v1/game/profile", t.authMiddleware.Authenticate(
		http.HandlerFunc(t.gameHandler.GetProfile),
	))

	// 게임 데이터 업데이트
	mux.Handle("/api/v1/game/update", t.authMiddleware.Authenticate(
		http.HandlerFunc(t.gameHandler.UpdateGameData),
	))

	// 리더보드
	mux.Handle("/api/v1/game/leaderboard", t.authMiddleware.Authenticate(
		http.HandlerFunc(t.gameHandler.GetLeaderboard),
	))

	// 게임 세션 시작
	mux.Handle("/api/v1/game/session/start", t.authMiddleware.Authenticate(
		http.HandlerFunc(t.gameHandler.StartSession),
	))

	// 게임 세션 종료
	mux.Handle("/api/v1/game/session/end", t.authMiddleware.Authenticate(
		http.HandlerFunc(t.gameHandler.EndSession),
	))

	log.Printf("[TimeSquare] Routes registered - Game APIs and Auth APIs ready")
}

// onStart 시작 시 호출되는 훅
func (t *TimeSquareApp) onStart(ctx context.Context) error {
	// Redis 연결 테스트
	if t.redisClient != nil {
		if _, err := t.redisClient.Get(ctx, "timesquare:startup").Result(); err != nil && err != redisClient.Nil {
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

	// Redis 연결 상태 확인
	if _, err := t.redisClient.Ping(context.Background()).Result(); err != nil {
		baseHealth.Status = serverapp.HealthStatusUnhealthy
		baseHealth.Message = "Redis connection failed"
		return baseHealth
	}

	return baseHealth
}

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"defense-allies-server/configs"
	"defense-allies-server/internal/serverapp"
	"defense-allies-server/internal/serverapp/health"
	"defense-allies-server/pkg/redis"
)

func main() {
	// 설정 로드
	config := configs.LoadConfig()

	// Redis 클라이언트 초기화
	redisClient, err := redis.NewClient(&config.Redis)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		// Redis 연결 실패해도 서버는 계속 실행
	}

	// ServerApp Manager 생성
	manager := serverapp.NewManager()

	// HealthApp 등록
	healthApp := health.NewHealthApp(redisClient)
	if err := manager.Register(healthApp); err != nil {
		log.Fatalf("Failed to register HealthApp: %v", err)
	}

	// 기본 라우트 추가
	mux := manager.GetMux()
	mux.HandleFunc("/", homeHandler)

	// 모든 ServerApp 시작
	ctx := context.Background()
	if err := manager.StartAll(ctx); err != nil {
		log.Printf("Warning: Some apps failed to start: %v", err)
	}

	// HTTP 서버 설정
	server := &http.Server{
		Addr:    ":" + config.Server.Port,
		Handler: mux,
	}

	// 서버 시작
	go func() {
		fmt.Printf("Defense Allies Server starting on port %s\n", config.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown 설정
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	// 서버 종료
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// ServerApp들 종료
	if err := manager.StopAll(ctx); err != nil {
		log.Printf("Error stopping apps: %v", err)
	}

	// Redis 연결 종료
	if redisClient != nil {
		redisClient.Close()
	}

	log.Println("Server exited")
}

// 홈 핸들러
func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Welcome to Defense Allies Server", "status": "running"}`)
}

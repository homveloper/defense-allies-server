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
	"defense-allies-server/serverapp/timesquare"
)

func main() {
	fmt.Println("🏙️ Welcome to Metropolis - TimeSquare Server")
	fmt.Println("The city that never sleeps is starting up...")

	// 설정 파일 경로 결정
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.json"
	}

	// 전역 설정 로드 (서버 정보용)
	globalConfig, err := configs.LoadConfigFromPath(configPath)
	if err != nil {
		log.Fatalf("Failed to load global config: %v", err)
	}

	// TimeSquareApp 생성
	timeSquareApp, err := timesquare.NewTimeSquareApp(configPath)
	if err != nil {
		log.Fatalf("Failed to create TimeSquareApp: %v", err)
	}

	// HTTP Mux 생성
	mux := http.NewServeMux()

	// 기본 라우트 추가
	mux.HandleFunc("/", homeHandler)

	// TimeSquareApp 라우트 등록
	timeSquareApp.RegisterRoutes(mux)

	// TimeSquareApp 시작
	ctx := context.Background()
	if err := timeSquareApp.Start(ctx); err != nil {
		log.Fatalf("Failed to start TimeSquareApp: %v", err)
	}

	// HTTP 서버 설정
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", globalConfig.Server.Host, globalConfig.Server.Port),
		Handler: mux,
	}

	// 서버 시작
	go func() {
		fmt.Printf("🚀 Metropolis TimeSquare Server starting on port %d\n", globalConfig.Server.Port)
		fmt.Println("🎮 Ready to welcome players to the square!")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown 설정
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	fmt.Println("\n🌃 Metropolis is shutting down...")
	log.Println("Shutting down TimeSquare server...")

	// 서버 종료
	shutdownTimeout := time.Duration(globalConfig.Server.GracefulTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// TimeSquareApp 종료
	if err := timeSquareApp.Stop(ctx); err != nil {
		log.Printf("Error stopping TimeSquareApp: %v", err)
	}

	fmt.Println("🌙 Metropolis has gone to sleep. Good night!")
	log.Println("TimeSquare server exited")
}

// homeHandler 홈 핸들러
func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := `{
		"service": "Metropolis - TimeSquare Server",
		"message": "Welcome to the city that never sleeps! 🏙️",
		"status": "running",
		"description": "The heart of Defense Allies where all players gather"
	}`
	fmt.Fprint(w, response)
}

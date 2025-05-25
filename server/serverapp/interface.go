package serverapp

import (
	"context"
	"net/http"
)

// ServerApp 인터페이스는 가장 작은 단위의 서버 컴포넌트를 정의합니다
type ServerApp interface {
	// Name 서버앱의 이름을 반환합니다
	Name() string

	// RegisterRoutes HTTP Mux에 라우트를 등록합니다
	RegisterRoutes(mux *http.ServeMux)

	// Start 서버앱을 시작합니다
	Start(ctx context.Context) error

	// Stop 서버앱을 graceful하게 종료합니다
	Stop(ctx context.Context) error

	// Health 서버앱의 상태를 확인합니다
	Health() HealthStatus
}

// HealthStatus 서버앱의 상태 정보
type HealthStatus struct {
	Status  string            `json:"status"`  // "healthy", "unhealthy", "degraded"
	Message string            `json:"message,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// AppState 서버앱의 상태
type AppState int

const (
	StateCreated AppState = iota
	StateStarting
	StateRunning
	StateStopping
	StateStopped
	StateFailed
)

// String AppState의 문자열 표현
func (s AppState) String() string {
	switch s {
	case StateCreated:
		return "created"
	case StateStarting:
		return "starting"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	case StateStopped:
		return "stopped"
	case StateFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// HealthStatusHealthy 건강한 상태
const HealthStatusHealthy = "healthy"

// HealthStatusUnhealthy 비건강한 상태
const HealthStatusUnhealthy = "unhealthy"

// HealthStatusDegraded 성능 저하 상태
const HealthStatusDegraded = "degraded"

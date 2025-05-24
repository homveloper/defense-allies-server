package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"defense-allies-server/internal/serverapp"
	"defense-allies-server/pkg/redis"
)

// HealthApp 헬스체크 기능을 제공하는 ServerApp
type HealthApp struct {
	*serverapp.BaseApp
	redisClient *redis.Client
}

// NewHealthApp 새로운 HealthApp을 생성합니다
func NewHealthApp(redisClient *redis.Client) *HealthApp {
	app := &HealthApp{
		BaseApp:     serverapp.NewBaseApp("health"),
		redisClient: redisClient,
	}
	return app
}

// RegisterRoutes HTTP Mux에 라우트를 등록합니다
func (h *HealthApp) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/health/detailed", h.handleDetailedHealth)
	mux.HandleFunc("/health/redis", h.handleRedisHealth)
}

// handleHealth 기본 헬스체크 핸들러
func (h *HealthApp) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	health := h.Health()

	w.Header().Set("Content-Type", "application/json")

	// 상태에 따른 HTTP 상태 코드 설정
	switch health.Status {
	case serverapp.HealthStatusHealthy:
		w.WriteHeader(http.StatusOK)
	case serverapp.HealthStatusDegraded:
		w.WriteHeader(http.StatusOK) // 성능 저하지만 서비스 가능
	case serverapp.HealthStatusUnhealthy:
		w.WriteHeader(http.StatusServiceUnavailable)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(health)
}

// handleDetailedHealth 상세 헬스체크 핸들러
func (h *HealthApp) handleDetailedHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 상세 헬스체크 수행
	detailedHealth := h.performDetailedHealthCheck()

	w.Header().Set("Content-Type", "application/json")

	// 전체 상태 확인
	overallHealthy := true
	for _, check := range detailedHealth.Checks {
		if check.Status != serverapp.HealthStatusHealthy {
			overallHealthy = false
			break
		}
	}

	if overallHealthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(detailedHealth)
}

// handleRedisHealth Redis 헬스체크 핸들러
func (h *HealthApp) handleRedisHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	redisHealth := h.checkRedisHealth()

	w.Header().Set("Content-Type", "application/json")

	if redisHealth.Status == serverapp.HealthStatusHealthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(redisHealth)
}

// DetailedHealthResponse 상세 헬스체크 응답
type DetailedHealthResponse struct {
	Status    string                            `json:"status"`
	Timestamp time.Time                         `json:"timestamp"`
	Uptime    string                            `json:"uptime,omitempty"`
	Checks    map[string]serverapp.HealthStatus `json:"checks"`
}

// performDetailedHealthCheck 상세 헬스체크를 수행합니다
func (h *HealthApp) performDetailedHealthCheck() DetailedHealthResponse {
	checks := make(map[string]serverapp.HealthStatus)

	// 앱 자체 상태
	checks["app"] = h.Health()

	// Redis 상태
	checks["redis"] = h.checkRedisHealth()

	// 메모리 상태 (간단한 체크)
	checks["memory"] = h.checkMemoryHealth()

	// 전체 상태 결정
	overallStatus := serverapp.HealthStatusHealthy
	for _, check := range checks {
		if check.Status == serverapp.HealthStatusUnhealthy {
			overallStatus = serverapp.HealthStatusUnhealthy
			break
		} else if check.Status == serverapp.HealthStatusDegraded {
			overallStatus = serverapp.HealthStatusDegraded
		}
	}

	response := DetailedHealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Checks:    checks,
	}

	if h.IsRunning() {
		response.Uptime = h.GetUptime().String()
	}

	return response
}

// checkRedisHealth Redis 연결 상태를 확인합니다
func (h *HealthApp) checkRedisHealth() serverapp.HealthStatus {
	if h.redisClient == nil {
		return serverapp.HealthStatus{
			Status:  serverapp.HealthStatusUnhealthy,
			Message: "Redis client not initialized",
		}
	}

	// Redis 핑 테스트
	start := time.Now()
	_, err := h.redisClient.Get("health:ping")
	latency := time.Since(start)

	if err != nil && err.Error() != "redis: nil" { // nil은 키가 없다는 뜻이므로 정상
		return serverapp.HealthStatus{
			Status:  serverapp.HealthStatusUnhealthy,
			Message: "Redis connection failed",
			Details: map[string]string{
				"error": err.Error(),
			},
		}
	}

	status := serverapp.HealthStatusHealthy
	message := "Redis connection healthy"

	// 지연시간이 100ms 이상이면 성능 저하로 판단
	if latency > 100*time.Millisecond {
		status = serverapp.HealthStatusDegraded
		message = "Redis connection slow"
	}

	return serverapp.HealthStatus{
		Status:  status,
		Message: message,
		Details: map[string]string{
			"latency": latency.String(),
		},
	}
}

// checkMemoryHealth 메모리 상태를 확인합니다 (간단한 구현)
func (h *HealthApp) checkMemoryHealth() serverapp.HealthStatus {
	// 실제 구현에서는 runtime.MemStats를 사용하여 메모리 사용량 확인
	return serverapp.HealthStatus{
		Status:  serverapp.HealthStatusHealthy,
		Message: "Memory usage normal",
		Details: map[string]string{
			"note": "Memory check not implemented",
		},
	}
}

// onStart 시작 시 호출되는 훅
func (h *HealthApp) onStart(ctx context.Context) error {
	// Redis 연결 테스트
	if h.redisClient != nil {
		if _, err := h.redisClient.Get("health:startup"); err != nil && err.Error() != "redis: nil" {
			return err
		}
	}
	return nil
}

// onStop 종료 시 호출되는 훅
func (h *HealthApp) onStop(ctx context.Context) error {
	// 정리 작업이 필요하면 여기서 수행
	return nil
}

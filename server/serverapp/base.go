package serverapp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// BaseApp ServerApp의 기본 구현체
type BaseApp struct {
	name         string
	state        AppState
	stateMutex   sync.RWMutex
	startTime    time.Time
	stopTime     time.Time
	healthStatus HealthStatus
	healthMutex  sync.RWMutex
}

// NewBaseApp 새로운 BaseApp을 생성합니다
func NewBaseApp(name string) *BaseApp {
	return &BaseApp{
		name:  name,
		state: StateCreated,
		healthStatus: HealthStatus{
			Status:  HealthStatusHealthy,
			Message: "App created",
			Details: make(map[string]string),
		},
	}
}

// Name 서버앱의 이름을 반환합니다
func (b *BaseApp) Name() string {
	return b.name
}

// GetState 현재 상태를 반환합니다
func (b *BaseApp) GetState() AppState {
	b.stateMutex.RLock()
	defer b.stateMutex.RUnlock()
	return b.state
}

// SetState 상태를 설정합니다
func (b *BaseApp) SetState(state AppState) {
	b.stateMutex.Lock()
	defer b.stateMutex.Unlock()

	oldState := b.state
	b.state = state

	log.Printf("[%s] State changed: %s -> %s", b.name, oldState.String(), state.String())

	// 상태에 따른 시간 기록
	switch state {
	case StateRunning:
		b.startTime = time.Now()
	case StateStopped, StateFailed:
		b.stopTime = time.Now()
	}
}

// RegisterRoutes HTTP Mux에 라우트를 등록합니다 (기본 구현은 빈 구현)
func (b *BaseApp) RegisterRoutes(mux *http.ServeMux) {
	// 기본 구현은 아무것도 하지 않음
	// 하위 클래스에서 오버라이드해야 함
}

// Start 서버앱을 시작합니다
func (b *BaseApp) Start(ctx context.Context) error {
	if b.GetState() != StateCreated {
		return fmt.Errorf("app %s is not in created state, current state: %s", b.name, b.GetState().String())
	}

	b.SetState(StateStarting)

	// 시작 로직 실행
	if err := b.onStart(ctx); err != nil {
		b.SetState(StateFailed)
		b.updateHealthStatus(HealthStatusUnhealthy, fmt.Sprintf("Failed to start: %v", err), nil)
		return fmt.Errorf("failed to start app %s: %w", b.name, err)
	}

	b.SetState(StateRunning)
	b.updateHealthStatus(HealthStatusHealthy, "App is running", map[string]string{
		"start_time": b.startTime.Format(time.RFC3339),
		"uptime":     time.Since(b.startTime).String(),
	})

	log.Printf("[%s] Started successfully", b.name)
	return nil
}

// Stop 서버앱을 graceful하게 종료합니다
func (b *BaseApp) Stop(ctx context.Context) error {
	currentState := b.GetState()
	if currentState != StateRunning {
		return fmt.Errorf("app %s is not running, current state: %s", b.name, currentState.String())
	}

	b.SetState(StateStopping)

	// 종료 로직 실행
	if err := b.onStop(ctx); err != nil {
		b.SetState(StateFailed)
		b.updateHealthStatus(HealthStatusUnhealthy, fmt.Sprintf("Failed to stop: %v", err), nil)
		return fmt.Errorf("failed to stop app %s: %w", b.name, err)
	}

	b.SetState(StateStopped)
	b.updateHealthStatus(HealthStatusHealthy, "App stopped", map[string]string{
		"stop_time": b.stopTime.Format(time.RFC3339),
		"runtime":   b.stopTime.Sub(b.startTime).String(),
	})

	log.Printf("[%s] Stopped successfully", b.name)
	return nil
}

// Health 서버앱의 상태를 확인합니다
func (b *BaseApp) Health() HealthStatus {
	b.healthMutex.RLock()
	defer b.healthMutex.RUnlock()

	// 상태 정보 업데이트
	status := b.healthStatus
	if status.Details == nil {
		status.Details = make(map[string]string)
	}

	status.Details["state"] = b.GetState().String()

	if b.GetState() == StateRunning {
		status.Details["uptime"] = time.Since(b.startTime).String()
	}

	return status
}

// updateHealthStatus 헬스 상태를 업데이트합니다
func (b *BaseApp) updateHealthStatus(status, message string, details map[string]string) {
	b.healthMutex.Lock()
	defer b.healthMutex.Unlock()

	b.healthStatus.Status = status
	b.healthStatus.Message = message

	if details != nil {
		if b.healthStatus.Details == nil {
			b.healthStatus.Details = make(map[string]string)
		}
		for k, v := range details {
			b.healthStatus.Details[k] = v
		}
	}
}

// onStart 시작 시 호출되는 훅 메서드 (하위 클래스에서 오버라이드)
func (b *BaseApp) onStart(_ context.Context) error {
	// 기본 구현은 아무것도 하지 않음
	return nil
}

// onStop 종료 시 호출되는 훅 메서드 (하위 클래스에서 오버라이드)
func (b *BaseApp) onStop(_ context.Context) error {
	// 기본 구현은 아무것도 하지 않음
	return nil
}

// IsRunning 앱이 실행 중인지 확인합니다
func (b *BaseApp) IsRunning() bool {
	return b.GetState() == StateRunning
}

// GetUptime 앱의 실행 시간을 반환합니다
func (b *BaseApp) GetUptime() time.Duration {
	if b.GetState() == StateRunning {
		return time.Since(b.startTime)
	}
	return 0
}

// GetStartTime 시작 시간을 반환합니다
func (b *BaseApp) GetStartTime() time.Time {
	return b.startTime
}

// GetStopTime 종료 시간을 반환합니다
func (b *BaseApp) GetStopTime() time.Time {
	return b.stopTime
}

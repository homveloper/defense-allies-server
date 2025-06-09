package redisstream

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// HealthStatus represents the health status of a component
type HealthStatus int

const (
	HealthStatusHealthy HealthStatus = iota
	HealthStatusDegraded
	HealthStatusUnhealthy
)

func (hs HealthStatus) String() string {
	switch hs {
	case HealthStatusHealthy:
		return "healthy"
	case HealthStatusDegraded:
		return "degraded"
	case HealthStatusUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

// HealthCheckResult contains the result of a health check
type HealthCheckResult struct {
	Status       HealthStatus           `json:"status"`
	Message      string                 `json:"message"`
	Error        string                 `json:"error,omitempty"`
	ResponseTime time.Duration          `json:"response_time"`
	Timestamp    time.Time              `json:"timestamp"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// HealthCheckSummary contains overall health check results
type HealthCheckSummary struct {
	ServiceName   string                       `json:"service_name"`
	OverallStatus HealthStatus                 `json:"overall_status"`
	Timestamp     time.Time                    `json:"timestamp"`
	Checks        map[string]HealthCheckResult `json:"checks"`
	Summary       map[string]interface{}       `json:"summary,omitempty"`
}

// HealthCheck interface for implementing health checks
type HealthCheck interface {
	Check(ctx context.Context) HealthCheckResult
	GetName() string
}

// HealthChecker interface for managing health checks
type HealthChecker interface {
	// Check management
	AddCheck(name string, check HealthCheck) error
	RemoveCheck(name string)
	CheckHealth(ctx context.Context) *HealthCheckSummary

	// Periodic checking
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// Configuration
	SetCheckInterval(interval time.Duration)
	GetCheckInterval() time.Duration
	SetCheckTimeout(timeout time.Duration)
	GetCheckTimeout() time.Duration

	// History and monitoring
	GetHealthHistory(limit int) []*HealthCheckSummary
	GetLastHealthCheck() *HealthCheckSummary
}

// healthChecker implements HealthChecker interface
type healthChecker struct {
	serviceName string
	config      *RedisStreamConfig

	// Health checks
	checks map[string]HealthCheck

	// Configuration
	checkInterval time.Duration
	checkTimeout  time.Duration

	// Periodic checking
	running   bool
	ctx       context.Context
	cancel    context.CancelFunc
	waitGroup sync.WaitGroup

	// History
	history    []*HealthCheckSummary
	maxHistory int

	// Thread safety
	mu sync.RWMutex
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(serviceName string, config *RedisStreamConfig) (HealthChecker, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("service name cannot be empty")
	}

	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	return &healthChecker{
		serviceName:   serviceName,
		config:        config,
		checks:        make(map[string]HealthCheck),
		checkInterval: config.Monitoring.HealthCheckInterval,
		checkTimeout:  30 * time.Second, // Default timeout
		maxHistory:    100,              // Keep last 100 results
		history:       make([]*HealthCheckSummary, 0, 100),
	}, nil
}

// AddCheck adds a health check
func (hc *healthChecker) AddCheck(name string, check HealthCheck) error {
	if name == "" {
		return fmt.Errorf("check name cannot be empty")
	}

	if check == nil {
		return fmt.Errorf("check cannot be nil")
	}

	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.checks[name] = check
	return nil
}

// RemoveCheck removes a health check
func (hc *healthChecker) RemoveCheck(name string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	delete(hc.checks, name)
}

// CheckHealth performs all health checks
func (hc *healthChecker) CheckHealth(ctx context.Context) *HealthCheckSummary {
	hc.mu.RLock()
	checks := make(map[string]HealthCheck)
	for name, check := range hc.checks {
		checks[name] = check
	}
	timeout := hc.checkTimeout
	hc.mu.RUnlock()

	// Create context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	summary := &HealthCheckSummary{
		ServiceName: hc.serviceName,
		Timestamp:   time.Now(),
		Checks:      make(map[string]HealthCheckResult),
		Summary:     make(map[string]interface{}),
	}

	overallStatus := HealthStatusHealthy
	totalChecks := len(checks)
	healthyCount := 0
	degradedCount := 0
	unhealthyCount := 0

	// Perform all checks
	for name, check := range checks {
		result := hc.performCheck(checkCtx, check)
		summary.Checks[name] = result

		// Update overall status (worst wins)
		if result.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
			unhealthyCount++
		} else if result.Status == HealthStatusDegraded && overallStatus != HealthStatusUnhealthy {
			overallStatus = HealthStatusDegraded
			degradedCount++
		} else if result.Status == HealthStatusHealthy {
			healthyCount++
		}
	}

	summary.OverallStatus = overallStatus
	summary.Summary["total_checks"] = totalChecks
	summary.Summary["healthy_count"] = healthyCount
	summary.Summary["degraded_count"] = degradedCount
	summary.Summary["unhealthy_count"] = unhealthyCount

	// Add to history
	hc.addToHistory(summary)

	return summary
}

// Start starts periodic health checking
func (hc *healthChecker) Start(ctx context.Context) error {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if hc.running {
		return fmt.Errorf("health checker is already running")
	}

	hc.ctx, hc.cancel = context.WithCancel(ctx)
	hc.running = true

	hc.waitGroup.Add(1)
	go hc.periodicCheck()

	return nil
}

// Stop stops periodic health checking
func (hc *healthChecker) Stop(ctx context.Context) error {
	hc.mu.Lock()
	if !hc.running {
		hc.mu.Unlock()
		return fmt.Errorf("health checker is not running")
	}

	hc.running = false
	if hc.cancel != nil {
		hc.cancel()
	}
	hc.mu.Unlock()

	// Wait for periodic check to finish
	done := make(chan struct{})
	go func() {
		hc.waitGroup.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout waiting for health checker to stop")
	}
}

// IsRunning returns whether periodic checking is running
func (hc *healthChecker) IsRunning() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.running
}

// SetCheckInterval sets the check interval
func (hc *healthChecker) SetCheckInterval(interval time.Duration) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checkInterval = interval
}

// GetCheckInterval gets the check interval
func (hc *healthChecker) GetCheckInterval() time.Duration {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.checkInterval
}

// SetCheckTimeout sets the check timeout
func (hc *healthChecker) SetCheckTimeout(timeout time.Duration) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checkTimeout = timeout
}

// GetCheckTimeout gets the check timeout
func (hc *healthChecker) GetCheckTimeout() time.Duration {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.checkTimeout
}

// GetHealthHistory returns recent health check history
func (hc *healthChecker) GetHealthHistory(limit int) []*HealthCheckSummary {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	if limit <= 0 || limit > len(hc.history) {
		limit = len(hc.history)
	}

	result := make([]*HealthCheckSummary, limit)
	copy(result, hc.history[len(hc.history)-limit:])

	return result
}

// GetLastHealthCheck returns the last health check result
func (hc *healthChecker) GetLastHealthCheck() *HealthCheckSummary {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	if len(hc.history) == 0 {
		return nil
	}

	return hc.history[len(hc.history)-1]
}

// Helper methods

func (hc *healthChecker) performCheck(ctx context.Context, check HealthCheck) HealthCheckResult {
	start := time.Now()

	result := check.Check(ctx)
	result.ResponseTime = time.Since(start)
	result.Timestamp = time.Now()

	return result
}

func (hc *healthChecker) addToHistory(summary *HealthCheckSummary) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.history = append(hc.history, summary)

	// Keep only the last maxHistory items
	if len(hc.history) > hc.maxHistory {
		hc.history = hc.history[len(hc.history)-hc.maxHistory:]
	}
}

func (hc *healthChecker) periodicCheck() {
	defer hc.waitGroup.Done()

	ticker := time.NewTicker(hc.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hc.ctx.Done():
			return
		case <-ticker.C:
			hc.CheckHealth(hc.ctx)
		}
	}
}

// Specific Health Check Implementations

// RedisHealthCheck checks Redis connectivity and performance
type RedisHealthCheck struct {
	client          redis.UniversalClient
	maxLatency      time.Duration
	degradedLatency time.Duration
}

// NewRedisHealthCheck creates a new Redis health check
func NewRedisHealthCheck(client redis.UniversalClient) HealthCheck {
	return &RedisHealthCheck{
		client:          client,
		maxLatency:      1 * time.Second,
		degradedLatency: 500 * time.Millisecond,
	}
}

// Check performs Redis health check
func (rhc *RedisHealthCheck) Check(ctx context.Context) HealthCheckResult {
	start := time.Now()

	// Ping Redis
	pong, err := rhc.client.Ping(ctx).Result()
	pingTime := time.Since(start)

	if err != nil {
		return HealthCheckResult{
			Status:  HealthStatusUnhealthy,
			Message: "Redis connection failed",
			Error:   err.Error(),
		}
	}

	if pong != "PONG" {
		return HealthCheckResult{
			Status:  HealthStatusUnhealthy,
			Message: "Redis ping response invalid",
			Error:   fmt.Sprintf("expected PONG, got %s", pong),
		}
	}

	// Check latency
	status := HealthStatusHealthy
	message := "Redis connection healthy"

	if pingTime > rhc.maxLatency {
		status = HealthStatusUnhealthy
		message = "Redis latency too high"
	} else if pingTime > rhc.degradedLatency {
		status = HealthStatusDegraded
		message = "Redis latency high"
	}

	// Get additional Redis info
	details := map[string]interface{}{
		"ping_time": pingTime.String(),
		"response":  pong,
	}

	// Try to get server info (non-critical)
	if info, err := rhc.client.Info(ctx, "server").Result(); err == nil {
		lines := strings.Split(info, "\r\n")
		serverInfo := make(map[string]string)
		for _, line := range lines {
			if strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					serverInfo[parts[0]] = parts[1]
				}
			}
		}
		details["server_info"] = serverInfo
	}

	return HealthCheckResult{
		Status:  status,
		Message: message,
		Details: details,
	}
}

// GetName returns the name of this health check
func (rhc *RedisHealthCheck) GetName() string {
	return "redis"
}

// EventBusHealthCheck checks EventBus health and performance
type EventBusHealthCheck struct {
	eventBus             *RedisStreamEventBus
	maxErrorRate         float64
	maxLatency           time.Duration
	minActiveSubscribers int
}

// NewEventBusHealthCheck creates a new EventBus health check
func NewEventBusHealthCheck(eventBus *RedisStreamEventBus) HealthCheck {
	return &EventBusHealthCheck{
		eventBus:             eventBus,
		maxErrorRate:         0.05, // 5% error rate threshold
		maxLatency:           1 * time.Second,
		minActiveSubscribers: 1,
	}
}

// Check performs EventBus health check
func (ebhc *EventBusHealthCheck) Check(ctx context.Context) HealthCheckResult {
	if !ebhc.eventBus.IsRunning() {
		return HealthCheckResult{
			Status:  HealthStatusUnhealthy,
			Message: "EventBus not running",
		}
	}

	metrics := ebhc.eventBus.GetMetrics()

	status := HealthStatusHealthy
	message := "EventBus running"
	var issues []string

	// Check error rate
	totalEvents := metrics.PublishedEvents + metrics.ProcessedEvents
	if totalEvents > 0 {
		errorRate := float64(metrics.FailedEvents) / float64(totalEvents)
		if errorRate > ebhc.maxErrorRate {
			status = HealthStatusDegraded
			issues = append(issues, fmt.Sprintf("high error rate: %.2f%%", errorRate*100))
		}
	}

	// Check latency
	if metrics.AverageLatency > ebhc.maxLatency {
		status = HealthStatusDegraded
		issues = append(issues, fmt.Sprintf("high latency: %v", metrics.AverageLatency))
	}

	// Check active subscribers
	if metrics.ActiveSubscribers < ebhc.minActiveSubscribers {
		status = HealthStatusDegraded
		issues = append(issues, fmt.Sprintf("low subscriber count: %d", metrics.ActiveSubscribers))
	}

	if status == HealthStatusDegraded {
		message = "EventBus performance degraded: " + strings.Join(issues, ", ")
	}

	details := map[string]interface{}{
		"published_events":   metrics.PublishedEvents,
		"processed_events":   metrics.ProcessedEvents,
		"failed_events":      metrics.FailedEvents,
		"active_subscribers": metrics.ActiveSubscribers,
		"average_latency":    metrics.AverageLatency.String(),
		"last_event_time":    metrics.LastEventTime.Format(time.RFC3339),
	}

	return HealthCheckResult{
		Status:  status,
		Message: message,
		Details: details,
	}
}

// GetName returns the name of this health check
func (ebhc *EventBusHealthCheck) GetName() string {
	return "eventbus"
}

// CircuitBreakerHealthCheck checks circuit breaker status
type CircuitBreakerHealthCheck struct {
	cbManager CircuitBreakerManager
}

// NewCircuitBreakerHealthCheck creates a new circuit breaker health check
func NewCircuitBreakerHealthCheck(cbManager CircuitBreakerManager) HealthCheck {
	return &CircuitBreakerHealthCheck{
		cbManager: cbManager,
	}
}

// Check performs circuit breaker health check
func (cbhc *CircuitBreakerHealthCheck) Check(ctx context.Context) HealthCheckResult {
	metrics := cbhc.cbManager.GetAllMetrics()

	status := HealthStatusHealthy
	message := "All circuit breakers healthy"

	openCircuits := 0
	halfOpenCircuits := 0
	totalCircuits := len(metrics)

	for _, metric := range metrics {
		if metric.CurrentState == CircuitBreakerStateOpen {
			openCircuits++
		} else if metric.CurrentState == CircuitBreakerStateHalfOpen {
			halfOpenCircuits++
		}
	}

	if openCircuits > 0 {
		status = HealthStatusDegraded
		message = fmt.Sprintf("%d open circuit breakers detected", openCircuits)
	} else if halfOpenCircuits > 0 {
		status = HealthStatusDegraded
		message = fmt.Sprintf("%d circuit breakers in recovery", halfOpenCircuits)
	}

	details := map[string]interface{}{
		"total_circuits":     totalCircuits,
		"open_circuits":      openCircuits,
		"half_open_circuits": halfOpenCircuits,
		"circuit_details":    metrics,
	}

	return HealthCheckResult{
		Status:  status,
		Message: message,
		Details: details,
	}
}

// GetName returns the name of this health check
func (cbhc *CircuitBreakerHealthCheck) GetName() string {
	return "circuit_breakers"
}

// CustomHealthCheck allows for custom health check implementations
type CustomHealthCheck struct {
	name    string
	checkFn func(ctx context.Context) HealthCheckResult
}

// NewCustomHealthCheck creates a new custom health check
func NewCustomHealthCheck(name string, checkFn func(ctx context.Context) HealthCheckResult) HealthCheck {
	return &CustomHealthCheck{
		name:    name,
		checkFn: checkFn,
	}
}

// Check performs the custom health check
func (chc *CustomHealthCheck) Check(ctx context.Context) HealthCheckResult {
	return chc.checkFn(ctx)
}

// GetName returns the name of this health check
func (chc *CustomHealthCheck) GetName() string {
	return chc.name
}

package redisstream

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cqrs"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerStateClosed CircuitBreakerState = iota
	CircuitBreakerStateOpen
	CircuitBreakerStateHalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case CircuitBreakerStateClosed:
		return "closed"
	case CircuitBreakerStateOpen:
		return "open"
	case CircuitBreakerStateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker interface for implementing circuit breaker pattern
type CircuitBreaker interface {
	// Execution control
	Call(fn func() error) error
	RecordSuccess()
	RecordFailure(reason string) error

	// State management
	GetState() CircuitBreakerState
	IsEnabled() bool
	Reset()

	// Monitoring
	GetMetrics() *CircuitBreakerMetrics
}

// CircuitBreakerMetrics contains metrics for circuit breaker monitoring
type CircuitBreakerMetrics struct {
	ServiceName       string
	CurrentState      CircuitBreakerState
	TotalCalls        int64
	SuccessfulCalls   int64
	FailedCalls       int64
	RejectedCalls     int64
	SuccessRate       float64
	FailureRate       float64
	StateTransitions  int64
	LastStateChange   time.Time
	LastFailureReason string
	RecoveryTimeLeft  time.Duration
}

// circuitBreaker implements CircuitBreaker interface
type circuitBreaker struct {
	serviceName string
	config      *RedisStreamConfig
	enabled     bool

	// State management
	state           CircuitBreakerState
	failureCount    int
	lastFailureTime time.Time
	lastStateChange time.Time

	// Metrics
	totalCalls        int64
	successfulCalls   int64
	failedCalls       int64
	rejectedCalls     int64
	stateTransitions  int64
	lastFailureReason string

	// Thread safety
	mu sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(serviceName string, config *RedisStreamConfig) (CircuitBreaker, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("service name cannot be empty")
	}

	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.Monitoring.CircuitBreakerEnabled {
		if err := validateCircuitBreakerConfig(&config.Monitoring); err != nil {
			return nil, err
		}
	}

	return &circuitBreaker{
		serviceName:     serviceName,
		config:          config,
		enabled:         config.Monitoring.CircuitBreakerEnabled,
		state:           CircuitBreakerStateClosed,
		lastStateChange: time.Now(),
	}, nil
}

// Call executes a function with circuit breaker protection
func (cb *circuitBreaker) Call(fn func() error) error {
	if !cb.enabled {
		// When disabled, execute directly and track metrics
		cb.mu.Lock()
		cb.totalCalls++
		cb.mu.Unlock()

		err := cb.safeCall(fn)

		cb.mu.Lock()
		if err != nil {
			cb.failedCalls++
		} else {
			cb.successfulCalls++
		}
		cb.mu.Unlock()

		return err
	}

	// Check if call should be allowed
	if !cb.allowCall() {
		cb.mu.Lock()
		cb.rejectedCalls++
		cb.mu.Unlock()
		return ErrCircuitBreakerOpen
	}

	// Execute the call
	cb.mu.Lock()
	cb.totalCalls++
	cb.mu.Unlock()

	err := cb.safeCall(fn)

	// Record the result
	if err != nil {
		cb.RecordFailure(err.Error())
	} else {
		cb.RecordSuccess()
	}

	return err
}

// RecordSuccess records a successful operation
func (cb *circuitBreaker) RecordSuccess() {
	if !cb.enabled {
		return
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successfulCalls++

	// Reset failure count on success
	cb.failureCount = 0

	// If in half-open state, close the circuit
	if cb.state == CircuitBreakerStateHalfOpen {
		cb.setState(CircuitBreakerStateClosed)
	}
}

// RecordFailure records a failed operation
func (cb *circuitBreaker) RecordFailure(reason string) error {
	if !cb.enabled {
		return nil
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failedCalls++
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	cb.lastFailureReason = reason

	// Check if we should open the circuit
	if cb.shouldOpen() {
		cb.setState(CircuitBreakerStateOpen)
		return ErrCircuitBreakerOpen
	}

	return nil
}

// GetState returns the current circuit breaker state
func (cb *circuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	// Check if we should transition from open to half-open
	if cb.state == CircuitBreakerStateOpen && cb.shouldAttemptReset() {
		cb.mu.RUnlock()
		cb.mu.Lock()
		// Double-check after acquiring write lock
		if cb.state == CircuitBreakerStateOpen && cb.shouldAttemptReset() {
			cb.setState(CircuitBreakerStateHalfOpen)
		}
		cb.mu.Unlock()
		cb.mu.RLock()
	}

	return cb.state
}

// IsEnabled returns whether the circuit breaker is enabled
func (cb *circuitBreaker) IsEnabled() bool {
	return cb.enabled
}

// Reset resets the circuit breaker to closed state
func (cb *circuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0
	cb.setState(CircuitBreakerStateClosed)
}

// GetMetrics returns current circuit breaker metrics
func (cb *circuitBreaker) GetMetrics() *CircuitBreakerMetrics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	totalCalls := cb.totalCalls
	successRate := 0.0
	failureRate := 0.0

	if totalCalls > 0 {
		successRate = float64(cb.successfulCalls) / float64(totalCalls)
		failureRate = float64(cb.failedCalls) / float64(totalCalls)
	}

	var recoveryTimeLeft time.Duration
	if cb.state == CircuitBreakerStateOpen {
		recoveryTime := cb.lastFailureTime.Add(cb.config.Monitoring.RecoveryTimeout)
		recoveryTimeLeft = time.Until(recoveryTime)
		if recoveryTimeLeft < 0 {
			recoveryTimeLeft = 0
		}
	}

	return &CircuitBreakerMetrics{
		ServiceName:       cb.serviceName,
		CurrentState:      cb.state,
		TotalCalls:        totalCalls,
		SuccessfulCalls:   cb.successfulCalls,
		FailedCalls:       cb.failedCalls,
		RejectedCalls:     cb.rejectedCalls,
		SuccessRate:       successRate,
		FailureRate:       failureRate,
		StateTransitions:  cb.stateTransitions,
		LastStateChange:   cb.lastStateChange,
		LastFailureReason: cb.lastFailureReason,
		RecoveryTimeLeft:  recoveryTimeLeft,
	}
}

// Helper methods

func (cb *circuitBreaker) allowCall() bool {
	state := cb.GetState() // This handles state transitions
	return state == CircuitBreakerStateClosed || state == CircuitBreakerStateHalfOpen
}

func (cb *circuitBreaker) shouldOpen() bool {
	return cb.failureCount >= cb.config.Monitoring.FailureThreshold
}

func (cb *circuitBreaker) shouldAttemptReset() bool {
	return time.Since(cb.lastFailureTime) >= cb.config.Monitoring.RecoveryTimeout
}

func (cb *circuitBreaker) setState(newState CircuitBreakerState) {
	if cb.state != newState {
		cb.state = newState
		cb.lastStateChange = time.Now()
		cb.stateTransitions++

		// Reset failure count when closing
		if newState == CircuitBreakerStateClosed {
			cb.failureCount = 0
		}
	}
}

func (cb *circuitBreaker) safeCall(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in circuit breaker call: %v", r)
		}
	}()

	return fn()
}

func validateCircuitBreakerConfig(config *MonitoringConfig) error {
	if config.FailureThreshold <= 0 {
		return fmt.Errorf("failure threshold must be greater than 0")
	}
	if config.RecoveryTimeout <= 0 {
		return fmt.Errorf("recovery timeout must be greater than 0")
	}
	return nil
}

// CircuitBreakerProtectedHandler wraps an event handler with circuit breaker protection
type CircuitBreakerProtectedHandler struct {
	handler        cqrs.EventHandler
	circuitBreaker CircuitBreaker
}

// NewCircuitBreakerProtectedHandler creates a new circuit breaker protected handler
func NewCircuitBreakerProtectedHandler(handler cqrs.EventHandler, config *RedisStreamConfig) *CircuitBreakerProtectedHandler {
	serviceName := fmt.Sprintf("%s_%s", handler.GetHandlerName(), handler.GetHandlerType().String())
	cb, err := NewCircuitBreaker(serviceName, config)
	if err != nil {
		// If circuit breaker creation fails, create a disabled one
		cb = &disabledCircuitBreaker{}
	}

	return &CircuitBreakerProtectedHandler{
		handler:        handler,
		circuitBreaker: cb,
	}
}

// Handle implements EventHandler interface with circuit breaker protection
func (cbh *CircuitBreakerProtectedHandler) Handle(ctx context.Context, event cqrs.EventMessage) error {
	return cbh.circuitBreaker.Call(func() error {
		return cbh.handler.Handle(ctx, event)
	})
}

// CanHandle implements EventHandler interface
func (cbh *CircuitBreakerProtectedHandler) CanHandle(eventType string) bool {
	return cbh.handler.CanHandle(eventType)
}

// GetHandlerName implements EventHandler interface
func (cbh *CircuitBreakerProtectedHandler) GetHandlerName() string {
	return cbh.handler.GetHandlerName()
}

// GetHandlerType implements EventHandler interface
func (cbh *CircuitBreakerProtectedHandler) GetHandlerType() cqrs.HandlerType {
	return cbh.handler.GetHandlerType()
}

// GetCircuitBreaker returns the circuit breaker for monitoring
func (cbh *CircuitBreakerProtectedHandler) GetCircuitBreaker() CircuitBreaker {
	return cbh.circuitBreaker
}

// disabledCircuitBreaker is a no-op circuit breaker for fallback scenarios
type disabledCircuitBreaker struct{}

func (dcb *disabledCircuitBreaker) Call(fn func() error) error {
	return fn()
}

func (dcb *disabledCircuitBreaker) RecordSuccess() {}

func (dcb *disabledCircuitBreaker) RecordFailure(reason string) error {
	return nil
}

func (dcb *disabledCircuitBreaker) GetState() CircuitBreakerState {
	return CircuitBreakerStateClosed
}

func (dcb *disabledCircuitBreaker) IsEnabled() bool {
	return false
}

func (dcb *disabledCircuitBreaker) Reset() {}

func (dcb *disabledCircuitBreaker) GetMetrics() *CircuitBreakerMetrics {
	return &CircuitBreakerMetrics{
		ServiceName:  "disabled",
		CurrentState: CircuitBreakerStateClosed,
	}
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager interface {
	GetCircuitBreaker(serviceName string) CircuitBreaker
	CreateCircuitBreaker(serviceName string) CircuitBreaker
	GetAllMetrics() map[string]*CircuitBreakerMetrics
	ResetAll()
}

// circuitBreakerManager implements CircuitBreakerManager
type circuitBreakerManager struct {
	config   *RedisStreamConfig
	breakers map[string]CircuitBreaker
	mu       sync.RWMutex
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(config *RedisStreamConfig) CircuitBreakerManager {
	return &circuitBreakerManager{
		config:   config,
		breakers: make(map[string]CircuitBreaker),
	}
}

// GetCircuitBreaker gets or creates a circuit breaker for a service
func (cbm *circuitBreakerManager) GetCircuitBreaker(serviceName string) CircuitBreaker {
	cbm.mu.RLock()
	if cb, exists := cbm.breakers[serviceName]; exists {
		cbm.mu.RUnlock()
		return cb
	}
	cbm.mu.RUnlock()

	return cbm.CreateCircuitBreaker(serviceName)
}

// CreateCircuitBreaker creates a new circuit breaker for a service
func (cbm *circuitBreakerManager) CreateCircuitBreaker(serviceName string) CircuitBreaker {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	// Double-check if another goroutine created it
	if cb, exists := cbm.breakers[serviceName]; exists {
		return cb
	}

	cb, err := NewCircuitBreaker(serviceName, cbm.config)
	if err != nil {
		// Return disabled circuit breaker on error
		cb = &disabledCircuitBreaker{}
	}

	cbm.breakers[serviceName] = cb
	return cb
}

// GetAllMetrics returns metrics for all circuit breakers
func (cbm *circuitBreakerManager) GetAllMetrics() map[string]*CircuitBreakerMetrics {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	metrics := make(map[string]*CircuitBreakerMetrics)
	for serviceName, cb := range cbm.breakers {
		metrics[serviceName] = cb.GetMetrics()
	}

	return metrics
}

// ResetAll resets all circuit breakers
func (cbm *circuitBreakerManager) ResetAll() {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	for _, cb := range cbm.breakers {
		cb.Reset()
	}
}

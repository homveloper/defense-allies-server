package redisstream

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cqrs"
)

// TestCircuitBreaker_Creation tests circuit breaker creation
func TestCircuitBreaker_Creation(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Monitoring.CircuitBreakerEnabled = true
	config.Monitoring.FailureThreshold = 5
	config.Monitoring.RecoveryTimeout = 30 * time.Second

	t.Run("should create circuit breaker with valid config", func(t *testing.T) {
		cb, err := NewCircuitBreaker("test_service", config)
		assert.NoError(t, err)
		assert.NotNil(t, cb)
		assert.True(t, cb.IsEnabled())
		assert.Equal(t, CircuitBreakerStateClosed, cb.GetState())
	})

	t.Run("should handle disabled circuit breaker", func(t *testing.T) {
		configDisabled := DefaultRedisStreamConfig()
		configDisabled.Monitoring.CircuitBreakerEnabled = false

		cb, err := NewCircuitBreaker("test_service", configDisabled)
		assert.NoError(t, err)
		assert.NotNil(t, cb)
		assert.False(t, cb.IsEnabled())
	})

	t.Run("should fail with invalid config", func(t *testing.T) {
		invalidConfig := DefaultRedisStreamConfig()
		invalidConfig.Monitoring.FailureThreshold = 0

		_, err := NewCircuitBreaker("test_service", invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failure threshold must be greater than 0")
	})

	t.Run("should fail with empty service name", func(t *testing.T) {
		_, err := NewCircuitBreaker("", config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "service name cannot be empty")
	})
}

// TestCircuitBreaker_StateTransitions tests state transitions
func TestCircuitBreaker_StateTransitions(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Monitoring.CircuitBreakerEnabled = true
	config.Monitoring.FailureThreshold = 3
	config.Monitoring.RecoveryTimeout = 100 * time.Millisecond

	cb, err := NewCircuitBreaker("test_service", config)
	require.NoError(t, err)

	t.Run("should transition from Closed to Open on failures", func(t *testing.T) {
		// Initially closed
		assert.Equal(t, CircuitBreakerStateClosed, cb.GetState())

		// Record failures
		for i := 0; i < 3; i++ {
			err := cb.RecordFailure("test error")
			if i < 2 {
				// Should remain closed for first 2 failures
				assert.NoError(t, err)
				assert.Equal(t, CircuitBreakerStateClosed, cb.GetState())
			} else {
				// Should open on 3rd failure
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrCircuitBreakerOpen)
				assert.Equal(t, CircuitBreakerStateOpen, cb.GetState())
			}
		}
	})

	t.Run("should reject calls when Open", func(t *testing.T) {
		// Ensure circuit is open
		for i := 0; i < 3; i++ {
			cb.RecordFailure("test error")
		}
		assert.Equal(t, CircuitBreakerStateOpen, cb.GetState())

		// All calls should be rejected
		for i := 0; i < 5; i++ {
			err := cb.Call(func() error {
				return nil // This should not be executed
			})
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrCircuitBreakerOpen)
		}
	})

	t.Run("should transition from Open to HalfOpen after timeout", func(t *testing.T) {
		// Create new circuit breaker with short timeout
		cb, _ := NewCircuitBreaker("test_service", config)

		// Trip the circuit
		for i := 0; i < 3; i++ {
			cb.RecordFailure("test error")
		}
		assert.Equal(t, CircuitBreakerStateOpen, cb.GetState())

		// Wait for recovery timeout
		time.Sleep(150 * time.Millisecond)

		// Next call should transition to half-open
		err := cb.Call(func() error {
			return nil // Success
		})
		assert.NoError(t, err)
		assert.Equal(t, CircuitBreakerStateClosed, cb.GetState()) // Should close after success
	})

	t.Run("should transition from HalfOpen to Closed on success", func(t *testing.T) {
		cb, _ := NewCircuitBreaker("test_service", config)

		// Trip the circuit
		for i := 0; i < 3; i++ {
			cb.RecordFailure("test error")
		}

		// Wait for recovery timeout
		time.Sleep(150 * time.Millisecond)

		// Successful call should close the circuit
		err := cb.Call(func() error {
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, CircuitBreakerStateClosed, cb.GetState())
	})

	t.Run("should transition from HalfOpen to Open on failure", func(t *testing.T) {
		cb, _ := NewCircuitBreaker("test_service", config)

		// Trip the circuit
		for i := 0; i < 3; i++ {
			cb.RecordFailure("test error")
		}

		// Wait for recovery timeout
		time.Sleep(150 * time.Millisecond)

		// Failed call should re-open the circuit
		err := cb.Call(func() error {
			return assert.AnError
		})
		assert.Error(t, err)
		assert.Equal(t, CircuitBreakerStateOpen, cb.GetState())
	})
}

// TestCircuitBreaker_CallExecution tests call execution logic
func TestCircuitBreaker_CallExecution(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Monitoring.CircuitBreakerEnabled = true
	config.Monitoring.FailureThreshold = 5

	cb, err := NewCircuitBreaker("test_service", config)
	require.NoError(t, err)

	t.Run("should execute calls when circuit is closed", func(t *testing.T) {
		callExecuted := false
		err := cb.Call(func() error {
			callExecuted = true
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, callExecuted)
		assert.Equal(t, CircuitBreakerStateClosed, cb.GetState())
	})

	t.Run("should record success and update metrics", func(t *testing.T) {
		err := cb.Call(func() error {
			return nil
		})

		assert.NoError(t, err)
		metrics := cb.GetMetrics()
		assert.Greater(t, metrics.TotalCalls, int64(0))
		assert.Greater(t, metrics.SuccessfulCalls, int64(0))
	})

	t.Run("should record failure and update metrics", func(t *testing.T) {
		initialFailures := cb.GetMetrics().FailedCalls

		err := cb.Call(func() error {
			return assert.AnError
		})

		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)

		metrics := cb.GetMetrics()
		assert.Equal(t, initialFailures+1, metrics.FailedCalls)
	})

	t.Run("should handle panic in called function", func(t *testing.T) {
		err := cb.Call(func() error {
			panic("test panic")
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "panic")

		// Circuit should record this as a failure
		metrics := cb.GetMetrics()
		assert.Greater(t, metrics.FailedCalls, int64(0))
	})
}

// TestCircuitBreaker_Metrics tests metrics collection
func TestCircuitBreaker_Metrics(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Monitoring.CircuitBreakerEnabled = true
	config.Monitoring.FailureThreshold = 5

	cb, err := NewCircuitBreaker("test_service", config)
	require.NoError(t, err)

	t.Run("should track basic metrics", func(t *testing.T) {
		// Execute some successful calls
		for i := 0; i < 10; i++ {
			cb.Call(func() error { return nil })
		}

		// Execute some failed calls
		for i := 0; i < 3; i++ {
			cb.Call(func() error { return assert.AnError })
		}

		metrics := cb.GetMetrics()
		assert.Equal(t, int64(13), metrics.TotalCalls)
		assert.Equal(t, int64(10), metrics.SuccessfulCalls)
		assert.Equal(t, int64(3), metrics.FailedCalls)
		assert.InDelta(t, 0.769, metrics.SuccessRate, 0.01) // 10/13 ≈ 0.769
	})

	t.Run("should track state transition metrics", func(t *testing.T) {
		cb, _ := NewCircuitBreaker("test_service", config)

		// Trip the circuit
		for i := 0; i < 5; i++ {
			cb.RecordFailure("test error")
		}

		metrics := cb.GetMetrics()
		assert.Greater(t, metrics.StateTransitions, int64(0))
		assert.Equal(t, CircuitBreakerStateOpen, metrics.CurrentState)
		assert.NotZero(t, metrics.LastStateChange)
	})

	t.Run("should calculate failure rate correctly", func(t *testing.T) {
		cb, _ := NewCircuitBreaker("test_service", config)

		// 7 successes, 3 failures = 30% failure rate
		for i := 0; i < 7; i++ {
			cb.RecordSuccess()
		}
		for i := 0; i < 3; i++ {
			cb.RecordFailure("test error")
		}

		metrics := cb.GetMetrics()
		assert.InDelta(t, 0.3, metrics.FailureRate, 0.01)
	})
}

// TestCircuitBreaker_EventHandlerIntegration tests integration with event handlers
func TestCircuitBreaker_EventHandlerIntegration(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Monitoring.CircuitBreakerEnabled = true
	config.Monitoring.FailureThreshold = 2

	t.Run("should protect event handler execution", func(t *testing.T) {
		handler := &TestEventHandler{
			name:        "circuit-breaker-test-handler",
			handlerType: cqrs.ProjectionHandler,
			eventTypes:  []string{"TestEvent"},
		}

		protectedHandler := NewCircuitBreakerProtectedHandler(handler, config)
		require.NotNil(t, protectedHandler)

		baseOptions := cqrs.Options().
			WithAggregateID("cb-test-123").
			WithAggregateType("CBTestAggregate")

		event := cqrs.NewBaseDomainEventMessage(
			"TestEvent",
			map[string]interface{}{"data": "test"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		// Should work normally when circuit is closed
		err := protectedHandler.Handle(context.Background(), event)
		assert.NoError(t, err)

		// Verify handler metadata
		assert.Equal(t, handler.GetHandlerName(), protectedHandler.GetHandlerName())
		assert.Equal(t, handler.GetHandlerType(), protectedHandler.GetHandlerType())
		assert.True(t, protectedHandler.CanHandle("TestEvent"))
	})

	t.Run("should open circuit on handler failures", func(t *testing.T) {
		failingHandler := &FailingEventHandler{
			name:        "failing-cb-handler",
			handlerType: cqrs.ProjectionHandler,
			eventTypes:  []string{"FailingEvent"},
		}

		protectedHandler := NewCircuitBreakerProtectedHandler(failingHandler, config)

		baseOptions := cqrs.Options().
			WithAggregateID("fail-test-123").
			WithAggregateType("FailTestAggregate")

		event := cqrs.NewBaseDomainEventMessage(
			"FailingEvent",
			map[string]interface{}{"data": "test"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		// First two failures should execute
		err1 := protectedHandler.Handle(context.Background(), event)
		assert.Error(t, err1)

		err2 := protectedHandler.Handle(context.Background(), event)
		assert.Error(t, err2)

		// Third call should be circuit breaker error
		err3 := protectedHandler.Handle(context.Background(), event)
		assert.Error(t, err3)
		assert.ErrorIs(t, err3, ErrCircuitBreakerOpen)
	})
}

// TestCircuitBreaker_Configuration tests different configuration scenarios
func TestCircuitBreaker_Configuration(t *testing.T) {
	t.Run("should handle different failure thresholds", func(t *testing.T) {
		testCases := []struct {
			threshold  int
			failures   int
			shouldOpen bool
		}{
			{1, 1, true},
			{5, 4, false},
			{5, 5, true},
			{10, 8, false},
		}

		for _, tc := range testCases {
			config := DefaultRedisStreamConfig()
			config.Monitoring.CircuitBreakerEnabled = true
			config.Monitoring.FailureThreshold = tc.threshold

			cb, err := NewCircuitBreaker("test_service", config)
			require.NoError(t, err)

			// Record failures
			for i := 0; i < tc.failures; i++ {
				cb.RecordFailure("test error")
			}

			if tc.shouldOpen {
				assert.Equal(t, CircuitBreakerStateOpen, cb.GetState(),
					"Circuit should be open after %d failures with threshold %d", tc.failures, tc.threshold)
			} else {
				assert.Equal(t, CircuitBreakerStateClosed, cb.GetState(),
					"Circuit should be closed after %d failures with threshold %d", tc.failures, tc.threshold)
			}
		}
	})

	t.Run("should handle different recovery timeouts", func(t *testing.T) {
		shortTimeout := 50 * time.Millisecond
		longTimeout := 500 * time.Millisecond

		config := DefaultRedisStreamConfig()
		config.Monitoring.CircuitBreakerEnabled = true
		config.Monitoring.FailureThreshold = 1

		// Test short timeout
		config.Monitoring.RecoveryTimeout = shortTimeout
		cb1, _ := NewCircuitBreaker("test_service_1", config)
		cb1.RecordFailure("test error")

		time.Sleep(shortTimeout + 10*time.Millisecond)

		// Should allow call after short timeout
		err := cb1.Call(func() error { return nil })
		assert.NoError(t, err)

		// Test long timeout
		config.Monitoring.RecoveryTimeout = longTimeout
		cb2, _ := NewCircuitBreaker("test_service_2", config)
		cb2.RecordFailure("test error")

		time.Sleep(shortTimeout) // Wait less than long timeout

		// Should still reject calls
		err = cb2.Call(func() error { return nil })
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrCircuitBreakerOpen)
	})
}

// TestCircuitBreaker_DisabledMode tests behavior when circuit breaker is disabled
func TestCircuitBreaker_DisabledMode(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Monitoring.CircuitBreakerEnabled = false

	cb, err := NewCircuitBreaker("test_service", config)
	require.NoError(t, err)

	t.Run("should always execute calls when disabled", func(t *testing.T) {
		// Record many failures
		for i := 0; i < 100; i++ {
			err := cb.Call(func() error {
				return assert.AnError
			})
			assert.Error(t, err)
			assert.Equal(t, assert.AnError, err) // Should get original error, not circuit breaker error
		}

		// Should still execute calls
		callExecuted := false
		err := cb.Call(func() error {
			callExecuted = true
			return nil
		})
		assert.NoError(t, err)
		assert.True(t, callExecuted)
	})

	t.Run("should still track metrics when disabled", func(t *testing.T) {
		initialMetrics := cb.GetMetrics()

		cb.Call(func() error { return nil })
		cb.Call(func() error { return assert.AnError })

		metrics := cb.GetMetrics()
		assert.Greater(t, metrics.TotalCalls, initialMetrics.TotalCalls)
	})
}

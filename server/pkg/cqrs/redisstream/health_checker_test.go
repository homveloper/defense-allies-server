package redisstream

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthChecker_Creation tests health checker creation
func TestHealthChecker_Creation(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Monitoring.HealthCheckInterval = 10 * time.Second

	t.Run("should create health checker with valid config", func(t *testing.T) {
		hc, err := NewHealthChecker("test_service", config)
		assert.NoError(t, err)
		assert.NotNil(t, hc)
		assert.False(t, hc.IsRunning()) // Should not be running initially
	})

	t.Run("should fail with empty service name", func(t *testing.T) {
		_, err := NewHealthChecker("", config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "service name cannot be empty")
	})

	t.Run("should fail with nil config", func(t *testing.T) {
		_, err := NewHealthChecker("test_service", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})
}

// TestHealthChecker_BasicChecks tests basic health check functionality
func TestHealthChecker_BasicChecks(t *testing.T) {
	ctx := context.Background()
	testContainer, err := NewTestRedisContainer(ctx)
	require.NoError(t, err)
	defer testContainer.Close(ctx)

	config := testContainer.config
	config.Monitoring.HealthCheckInterval = 100 * time.Millisecond

	hc, err := NewHealthChecker("test_service", config)
	require.NoError(t, err)

	t.Run("should perform individual health checks", func(t *testing.T) {
		// Add Redis client check
		err := hc.AddCheck("redis", NewRedisHealthCheck(testContainer.client))
		assert.NoError(t, err)

		// Add custom check
		err = hc.AddCheck("custom", NewCustomHealthCheck("custom_test", func(ctx context.Context) HealthCheckResult {
			return HealthCheckResult{
				Status:  HealthStatusHealthy,
				Message: "Custom check passed",
			}
		}))
		assert.NoError(t, err)

		// Perform checks
		results := hc.CheckHealth(ctx)
		
		assert.Len(t, results.Checks, 2)
		assert.Contains(t, results.Checks, "redis")
		assert.Contains(t, results.Checks, "custom")
		
		// Redis should be healthy
		redisResult := results.Checks["redis"]
		assert.Equal(t, HealthStatusHealthy, redisResult.Status)
		
		// Custom check should be healthy
		customResult := results.Checks["custom"]
		assert.Equal(t, HealthStatusHealthy, customResult.Status)
		
		// Overall status should be healthy
		assert.Equal(t, HealthStatusHealthy, results.OverallStatus)
	})

	t.Run("should detect unhealthy services", func(t *testing.T) {
		hc, _ := NewHealthChecker("test_service", config)
		
		// Add failing check
		err := hc.AddCheck("failing", NewCustomHealthCheck("failing_test", func(ctx context.Context) HealthCheckResult {
			return HealthCheckResult{
				Status:  HealthStatusUnhealthy,
				Message: "Service is down",
				Error:   "Connection failed",
			}
		}))
		assert.NoError(t, err)

		results := hc.CheckHealth(ctx)
		
		failingResult := results.Checks["failing"]
		assert.Equal(t, HealthStatusUnhealthy, failingResult.Status)
		assert.Equal(t, "Service is down", failingResult.Message)
		assert.Equal(t, "Connection failed", failingResult.Error)
		
		// Overall status should be unhealthy
		assert.Equal(t, HealthStatusUnhealthy, results.OverallStatus)
	})

	t.Run("should handle degraded services", func(t *testing.T) {
		hc, _ := NewHealthChecker("test_service", config)
		
		// Add healthy check
		hc.AddCheck("healthy", NewCustomHealthCheck("healthy_test", func(ctx context.Context) HealthCheckResult {
			return HealthCheckResult{Status: HealthStatusHealthy, Message: "OK"}
		}))
		
		// Add degraded check
		hc.AddCheck("degraded", NewCustomHealthCheck("degraded_test", func(ctx context.Context) HealthCheckResult {
			return HealthCheckResult{
				Status:  HealthStatusDegraded,
				Message: "Performance issues detected",
			}
		}))

		results := hc.CheckHealth(ctx)
		
		// Overall status should be degraded (worst of healthy and degraded)
		assert.Equal(t, HealthStatusDegraded, results.OverallStatus)
	})
}

// TestHealthChecker_RedisHealthCheck tests Redis-specific health checks
func TestHealthChecker_RedisHealthCheck(t *testing.T) {
	ctx := context.Background()

	t.Run("should detect healthy Redis connection", func(t *testing.T) {
		testContainer, err := NewTestRedisContainer(ctx)
		require.NoError(t, err)
		defer testContainer.Close(ctx)

		redisCheck := NewRedisHealthCheck(testContainer.client)
		result := redisCheck.Check(ctx)
		
		assert.Equal(t, HealthStatusHealthy, result.Status)
		assert.Contains(t, result.Message, "Redis connection healthy")
		assert.NotZero(t, result.ResponseTime)
		assert.Contains(t, result.Details, "ping_time")
		assert.Contains(t, result.Details, "server_info")
	})

	t.Run("should detect unhealthy Redis connection", func(t *testing.T) {
		// Create client with invalid address
		invalidClient := redis.NewClient(&redis.Options{
			Addr: "localhost:9999", // Non-existent port
		})
		defer invalidClient.Close()

		redisCheck := NewRedisHealthCheck(invalidClient)
		result := redisCheck.Check(ctx)
		
		assert.Equal(t, HealthStatusUnhealthy, result.Status)
		assert.Contains(t, result.Message, "Redis connection failed")
		assert.NotEmpty(t, result.Error)
	})

	t.Run("should detect degraded Redis performance", func(t *testing.T) {
		testContainer, err := NewTestRedisContainer(ctx)
		require.NoError(t, err)
		defer testContainer.Close(ctx)

		// Create Redis check with very low latency threshold
		redisCheck := &RedisHealthCheck{
			client:           testContainer.client,
			maxLatency:       1 * time.Nanosecond, // Unrealistically low
			degradedLatency:  1 * time.Nanosecond,
		}
		
		result := redisCheck.Check(ctx)
		
		// Should be degraded due to high latency
		assert.Equal(t, HealthStatusDegraded, result.Status)
		assert.Contains(t, result.Message, "Redis latency high")
	})
}

// TestHealthChecker_EventBusHealthCheck tests EventBus health checks
func TestHealthChecker_EventBusHealthCheck(t *testing.T) {
	ctx := context.Background()
	testContainer, err := NewTestRedisContainer(ctx)
	require.NoError(t, err)
	defer testContainer.Close(ctx)

	eventBus, err := NewRedisStreamEventBus(testContainer.client, testContainer.config)
	require.NoError(t, err)

	t.Run("should detect stopped EventBus", func(t *testing.T) {
		eventBusCheck := NewEventBusHealthCheck(eventBus)
		result := eventBusCheck.Check(ctx)
		
		assert.Equal(t, HealthStatusUnhealthy, result.Status)
		assert.Contains(t, result.Message, "EventBus not running")
	})

	t.Run("should detect running EventBus", func(t *testing.T) {
		err := eventBus.Start(ctx)
		require.NoError(t, err)
		defer eventBus.Stop(ctx)

		eventBusCheck := NewEventBusHealthCheck(eventBus)
		result := eventBusCheck.Check(ctx)
		
		assert.Equal(t, HealthStatusHealthy, result.Status)
		assert.Contains(t, result.Message, "EventBus running")
		assert.Contains(t, result.Details, "published_events")
		assert.Contains(t, result.Details, "processed_events")
		assert.Contains(t, result.Details, "active_subscribers")
	})

	t.Run("should detect degraded EventBus performance", func(t *testing.T) {
		err := eventBus.Start(ctx)
		require.NoError(t, err)
		defer eventBus.Stop(ctx)

		// Create check with low thresholds to trigger degraded status
		eventBusCheck := &EventBusHealthCheck{
			eventBus:              eventBus,
			maxErrorRate:          0.0, // Any errors will trigger degraded
			maxLatency:            1 * time.Nanosecond,
			minActiveSubscribers:  100, // Unrealistically high
		}
		
		result := eventBusCheck.Check(ctx)
		
		// Should be degraded due to low subscriber count
		assert.Equal(t, HealthStatusDegraded, result.Status)
		assert.Contains(t, result.Message, "EventBus performance degraded")
	})
}

// TestHealthChecker_PeriodicChecks tests periodic health checking
func TestHealthChecker_PeriodicChecks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping periodic health check test in short mode")
	}

	ctx := context.Background()
	testContainer, err := NewTestRedisContainer(ctx)
	require.NoError(t, err)
	defer testContainer.Close(ctx)

	config := testContainer.config
	config.Monitoring.HealthCheckInterval = 50 * time.Millisecond

	hc, err := NewHealthChecker("test_service", config)
	require.NoError(t, err)

	t.Run("should run periodic health checks", func(t *testing.T) {
		checkCount := 0
		hc.AddCheck("counter", NewCustomHealthCheck("counter", func(ctx context.Context) HealthCheckResult {
			checkCount++
			return HealthCheckResult{
				Status:  HealthStatusHealthy,
				Message: fmt.Sprintf("Check #%d", checkCount),
			}
		}))

		// Start periodic checking
		err := hc.Start(ctx)
		assert.NoError(t, err)
		assert.True(t, hc.IsRunning())

		// Wait for multiple checks
		time.Sleep(200 * time.Millisecond)

		// Stop periodic checking
		err = hc.Stop(ctx)
		assert.NoError(t, err)
		assert.False(t, hc.IsRunning())

		// Should have performed multiple checks
		assert.Greater(t, checkCount, 2)
	})

	t.Run("should track health check history", func(t *testing.T) {
		hc, _ := NewHealthChecker("test_service", config)
		
		hc.AddCheck("redis", NewRedisHealthCheck(testContainer.client))

		// Perform several checks
		for i := 0; i < 5; i++ {
			hc.CheckHealth(ctx)
			time.Sleep(10 * time.Millisecond)
		}

		history := hc.GetHealthHistory(5)
		assert.Len(t, history, 5)
		
		// All checks should be recent
		for _, result := range history {
			assert.WithinDuration(t, time.Now(), result.Timestamp, 1*time.Second)
		}
	})
}

// TestHealthChecker_CircuitBreakerIntegration tests circuit breaker integration
func TestHealthChecker_CircuitBreakerIntegration(t *testing.T) {
	ctx := context.Background()
	config := DefaultRedisStreamConfig()
	config.Monitoring.CircuitBreakerEnabled = true
	config.Monitoring.FailureThreshold = 2

	hc, err := NewHealthChecker("test_service", config)
	require.NoError(t, err)

	t.Run("should include circuit breaker status in health checks", func(t *testing.T) {
		// Create circuit breaker manager
		cbManager := NewCircuitBreakerManager(config)
		cb := cbManager.CreateCircuitBreaker("test_service")

		// Add circuit breaker health check
		err := hc.AddCheck("circuit_breakers", NewCircuitBreakerHealthCheck(cbManager))
		assert.NoError(t, err)

		// Initially should be healthy
		results := hc.CheckHealth(ctx)
		cbResult := results.Checks["circuit_breakers"]
		assert.Equal(t, HealthStatusHealthy, cbResult.Status)

		// Trip the circuit breaker
		cb.RecordFailure("test error")
		cb.RecordFailure("test error")

		// Should now show degraded status
		results = hc.CheckHealth(ctx)
		cbResult = results.Checks["circuit_breakers"]
		assert.Equal(t, HealthStatusDegraded, cbResult.Status)
		assert.Contains(t, cbResult.Message, "open circuit breakers detected")
	})
}

// TestHealthChecker_Configuration tests different configuration scenarios
func TestHealthChecker_Configuration(t *testing.T) {
	t.Run("should handle different check intervals", func(t *testing.T) {
		shortInterval := 10 * time.Millisecond
		longInterval := 1 * time.Second

		config := DefaultRedisStreamConfig()
		config.Monitoring.HealthCheckInterval = shortInterval

		hc, err := NewHealthChecker("test_service", config)
		require.NoError(t, err)

		// Verify interval is set correctly
		assert.Equal(t, shortInterval, hc.GetCheckInterval())

		// Test updating interval
		hc.SetCheckInterval(longInterval)
		assert.Equal(t, longInterval, hc.GetCheckInterval())
	})

	t.Run("should support check timeouts", func(t *testing.T) {
		config := DefaultRedisStreamConfig()
		hc, err := NewHealthChecker("test_service", config)
		require.NoError(t, err)

		// Add slow check
		slowCheck := NewCustomHealthCheck("slow", func(ctx context.Context) HealthCheckResult {
			select {
			case <-time.After(200 * time.Millisecond):
				return HealthCheckResult{Status: HealthStatusHealthy, Message: "Slow check completed"}
			case <-ctx.Done():
				return HealthCheckResult{Status: HealthStatusUnhealthy, Message: "Check timed out", Error: ctx.Err().Error()}
			}
		})

		hc.AddCheck("slow", slowCheck)

		// Set short timeout
		hc.SetCheckTimeout(50 * time.Millisecond)

		// Check should timeout
		ctx := context.Background()
		results := hc.CheckHealth(ctx)
		slowResult := results.Checks["slow"]
		assert.Equal(t, HealthStatusUnhealthy, slowResult.Status)
		assert.Contains(t, slowResult.Error, "context deadline exceeded")
	})
}

// TestHealthChecker_CustomChecks tests custom health check implementations
func TestHealthChecker_CustomChecks(t *testing.T) {
	ctx := context.Background()
	config := DefaultRedisStreamConfig()
	hc, err := NewHealthChecker("test_service", config)
	require.NoError(t, err)

	t.Run("should support custom health checks", func(t *testing.T) {
		// Add custom database check
		dbCheck := NewCustomHealthCheck("database", func(ctx context.Context) HealthCheckResult {
			// Simulate database connectivity check
			return HealthCheckResult{
				Status:  HealthStatusHealthy,
				Message: "Database connection established",
				Details: map[string]interface{}{
					"connection_pool_size": 10,
					"active_connections":   3,
					"last_query_time":      time.Now().Add(-5 * time.Second).Format(time.RFC3339),
				},
			}
		})

		hc.AddCheck("database", dbCheck)

		results := hc.CheckHealth(ctx)
		dbResult := results.Checks["database"]
		
		assert.Equal(t, HealthStatusHealthy, dbResult.Status)
		assert.Equal(t, "Database connection established", dbResult.Message)
		assert.Contains(t, dbResult.Details, "connection_pool_size")
		assert.Contains(t, dbResult.Details, "active_connections")
	})

	t.Run("should support removing checks", func(t *testing.T) {
		hc.AddCheck("temporary", NewCustomHealthCheck("temp", func(ctx context.Context) HealthCheckResult {
			return HealthCheckResult{Status: HealthStatusHealthy}
		}))

		// Check is present
		results := hc.CheckHealth(ctx)
		assert.Contains(t, results.Checks, "temporary")

		// Remove check
		hc.RemoveCheck("temporary")

		// Check is no longer present
		results = hc.CheckHealth(ctx)
		assert.NotContains(t, results.Checks, "temporary")
	})
}

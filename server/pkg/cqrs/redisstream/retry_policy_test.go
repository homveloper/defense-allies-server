package redisstream

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"defense-allies-server/pkg/cqrs"
)

// TestRetryPolicyManager_Creation tests retry policy manager creation
func TestRetryPolicyManager_Creation(t *testing.T) {
	config := DefaultRedisStreamConfig()

	t.Run("should create retry policy manager with valid config", func(t *testing.T) {
		manager, err := NewRetryPolicyManager(config)
		assert.NoError(t, err)
		assert.NotNil(t, manager)
	})

	t.Run("should fail with nil config", func(t *testing.T) {
		_, err := NewRetryPolicyManager(nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidConfiguration)
	})

	t.Run("should validate retry configuration", func(t *testing.T) {
		invalidConfig := DefaultRedisStreamConfig()
		invalidConfig.Retry.MaxAttempts = 0

		_, err := NewRetryPolicyManager(invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max attempts must be greater than 0")
	})
}

// TestRetryPolicyManager_DelayCalculation tests delay calculation for different backoff types
func TestRetryPolicyManager_DelayCalculation(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Retry.InitialDelay = 100 * time.Millisecond
	config.Retry.MaxDelay = 5 * time.Second
	config.Retry.BackoffFactor = 2.0

	manager, err := NewRetryPolicyManager(config)
	require.NoError(t, err)

	t.Run("should calculate fixed backoff delay", func(t *testing.T) {
		policy := &RetryPolicy{
			MaxAttempts:   3,
			InitialDelay:  100 * time.Millisecond,
			MaxDelay:      5 * time.Second,
			BackoffType:   cqrs.FixedBackoff,
			BackoffFactor: 2.0,
		}

		testCases := []struct {
			attempt  int
			expected time.Duration
		}{
			{1, 100 * time.Millisecond},
			{2, 100 * time.Millisecond},
			{3, 100 * time.Millisecond},
			{10, 100 * time.Millisecond}, // Should remain constant
		}

		for _, tc := range testCases {
			delay := manager.CalculateDelay(policy, tc.attempt)
			assert.Equal(t, tc.expected, delay,
				"Failed for attempt %d", tc.attempt)
		}
	})

	t.Run("should calculate exponential backoff delay", func(t *testing.T) {
		policy := &RetryPolicy{
			MaxAttempts:   5,
			InitialDelay:  100 * time.Millisecond,
			MaxDelay:      5 * time.Second,
			BackoffType:   cqrs.ExponentialBackoff,
			BackoffFactor: 2.0,
		}

		testCases := []struct {
			attempt  int
			expected time.Duration
		}{
			{1, 100 * time.Millisecond},  // 100ms
			{2, 200 * time.Millisecond},  // 100ms * 2^1
			{3, 400 * time.Millisecond},  // 100ms * 2^2
			{4, 800 * time.Millisecond},  // 100ms * 2^3
			{5, 1600 * time.Millisecond}, // 100ms * 2^4
			{10, 5 * time.Second},        // Should cap at MaxDelay
		}

		for _, tc := range testCases {
			delay := manager.CalculateDelay(policy, tc.attempt)
			assert.Equal(t, tc.expected, delay,
				"Failed for attempt %d", tc.attempt)
		}
	})

	t.Run("should calculate linear backoff delay", func(t *testing.T) {
		policy := &RetryPolicy{
			MaxAttempts:   5,
			InitialDelay:  100 * time.Millisecond,
			MaxDelay:      5 * time.Second,
			BackoffType:   cqrs.LinearBackoff,
			BackoffFactor: 2.0,
		}

		testCases := []struct {
			attempt  int
			expected time.Duration
		}{
			{1, 100 * time.Millisecond}, // 100ms
			{2, 300 * time.Millisecond}, // 100ms + (100ms * 2.0 * 1)
			{3, 500 * time.Millisecond}, // 100ms + (100ms * 2.0 * 2)
			{4, 700 * time.Millisecond}, // 100ms + (100ms * 2.0 * 3)
			{5, 900 * time.Millisecond}, // 100ms + (100ms * 2.0 * 4)
		}

		for _, tc := range testCases {
			delay := manager.CalculateDelay(policy, tc.attempt)
			assert.Equal(t, tc.expected, delay,
				"Failed for attempt %d", tc.attempt)
		}
	})

	t.Run("should respect max delay limit", func(t *testing.T) {
		policy := &RetryPolicy{
			MaxAttempts:   10,
			InitialDelay:  1 * time.Second,
			MaxDelay:      3 * time.Second,
			BackoffType:   cqrs.ExponentialBackoff,
			BackoffFactor: 3.0,
		}

		// High attempt number should be capped at MaxDelay
		delay := manager.CalculateDelay(policy, 10)
		assert.Equal(t, 3*time.Second, delay)
	})
}

// TestRetryPolicyManager_ShouldRetry tests retry decision logic
func TestRetryPolicyManager_ShouldRetry(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Retry.MaxAttempts = 3

	manager, err := NewRetryPolicyManager(config)
	require.NoError(t, err)

	t.Run("should retry within max attempts", func(t *testing.T) {
		baseOptions := cqrs.Options().
			WithAggregateID("retry-123").
			WithAggregateType("RetryAggregate").
			WithMetadata(map[string]interface{}{
				"retry_count": 1,
				"max_retries": 3,
			})

		event := cqrs.NewBaseDomainEventMessage(
			"RetryEvent",
			map[string]interface{}{"data": "test"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		error := &ProcessingError{
			Error:      "temporary network error",
			Handler:    "TestHandler",
			Timestamp:  time.Now(),
			RetryCount: 1,
		}

		shouldRetry := manager.ShouldRetry(event, error)
		assert.True(t, shouldRetry)
	})

	t.Run("should not retry when max attempts reached", func(t *testing.T) {
		baseOptions := cqrs.Options().
			WithAggregateID("max-retry-123").
			WithAggregateType("MaxRetryAggregate").
			WithMetadata(map[string]interface{}{
				"retry_count": 3,
				"max_retries": 3,
			})

		event := cqrs.NewBaseDomainEventMessage(
			"MaxRetryEvent",
			map[string]interface{}{"data": "test"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		error := &ProcessingError{
			Error:      "persistent error",
			Handler:    "TestHandler",
			Timestamp:  time.Now(),
			RetryCount: 3,
		}

		shouldRetry := manager.ShouldRetry(event, error)
		assert.False(t, shouldRetry)
	})

	t.Run("should handle events without retry metadata", func(t *testing.T) {
		baseOptions := cqrs.Options().
			WithAggregateID("no-metadata-123").
			WithAggregateType("NoMetadataAggregate")

		event := cqrs.NewBaseDomainEventMessage(
			"NoMetadataEvent",
			map[string]interface{}{"data": "test"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		error := &ProcessingError{
			Error:     "first error",
			Handler:   "TestHandler",
			Timestamp: time.Now(),
		}

		shouldRetry := manager.ShouldRetry(event, error)
		assert.True(t, shouldRetry) // Should retry first time
	})

	t.Run("should consider error type for retry decisions", func(t *testing.T) {
		baseOptions := cqrs.Options().
			WithAggregateID("error-type-123").
			WithAggregateType("ErrorTypeAggregate")

		event := cqrs.NewBaseDomainEventMessage(
			"ErrorTypeEvent",
			map[string]interface{}{"data": "test"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		// Non-retryable error
		nonRetryableError := &ProcessingError{
			Error:     "validation failed: invalid data format",
			Handler:   "ValidationHandler",
			Timestamp: time.Now(),
		}

		shouldRetry := manager.ShouldRetry(event, nonRetryableError)
		assert.False(t, shouldRetry, "Should not retry validation errors")

		// Retryable error
		retryableError := &ProcessingError{
			Error:     "connection timeout",
			Handler:   "NetworkHandler",
			Timestamp: time.Now(),
		}

		shouldRetry = manager.ShouldRetry(event, retryableError)
		assert.True(t, shouldRetry, "Should retry network errors")
	})
}

// TestRetryPolicyManager_EventEnrichment tests retry event enrichment
func TestRetryPolicyManager_EventEnrichment(t *testing.T) {
	config := DefaultRedisStreamConfig()
	manager, err := NewRetryPolicyManager(config)
	require.NoError(t, err)

	t.Run("should enrich event with retry metadata", func(t *testing.T) {
		baseOptions := cqrs.Options().
			WithAggregateID("enrich-123").
			WithAggregateType("EnrichAggregate")

		originalEvent := cqrs.NewBaseDomainEventMessage(
			"EnrichEvent",
			map[string]interface{}{"data": "test"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		error := &ProcessingError{
			Error:      "network timeout",
			Handler:    "NetworkHandler",
			Timestamp:  time.Now(),
			RetryCount: 1,
			StreamName: "events:domain:normal:user",
		}

		enrichedEvent := manager.EnrichEventForRetry(originalEvent, error)

		// Verify retry metadata was added
		metadata := enrichedEvent.Metadata()
		assert.Contains(t, metadata, "retry_count")
		assert.Contains(t, metadata, "max_retries")
		assert.Contains(t, metadata, "last_error")
		assert.Contains(t, metadata, "last_retry_timestamp")
		assert.Contains(t, metadata, "retry_handler")

		assert.Equal(t, 2, metadata["retry_count"]) // Should increment
		assert.Equal(t, config.Retry.MaxAttempts, metadata["max_retries"])
		assert.Equal(t, "network timeout", metadata["last_error"])
		assert.Equal(t, "NetworkHandler", metadata["retry_handler"])
	})

	t.Run("should preserve original event data", func(t *testing.T) {
		originalData := map[string]interface{}{
			"user_id": "user-456",
			"action":  "update_profile",
			"details": map[string]interface{}{
				"field": "email",
				"value": "new@example.com",
			},
		}

		baseOptions := cqrs.Options().
			WithAggregateID("preserve-123").
			WithAggregateType("PreserveAggregate")

		originalEvent := cqrs.NewBaseDomainEventMessage(
			"PreserveEvent",
			originalData,
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		error := &ProcessingError{
			Error:     "temporary error",
			Handler:   "TestHandler",
			Timestamp: time.Now(),
		}

		enrichedEvent := manager.EnrichEventForRetry(originalEvent, error)

		// Verify original data is preserved
		assert.Equal(t, originalEvent.EventType(), enrichedEvent.EventType())
		assert.Equal(t, originalEvent.ID(), enrichedEvent.ID())
		assert.Equal(t, originalEvent.EventData(), enrichedEvent.EventData())
	})

	t.Run("should track retry history", func(t *testing.T) {
		baseOptions := cqrs.Options().
			WithAggregateID("history-123").
			WithAggregateType("HistoryAggregate").
			WithMetadata(map[string]interface{}{
				"retry_count": 1,
				"retry_history": []map[string]interface{}{
					{
						"attempt":   1,
						"error":     "first error",
						"timestamp": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
						"handler":   "FirstHandler",
					},
				},
			})

		event := cqrs.NewBaseDomainEventMessage(
			"HistoryEvent",
			map[string]interface{}{"data": "test"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		error := &ProcessingError{
			Error:     "second error",
			Handler:   "SecondHandler",
			Timestamp: time.Now(),
		}

		enrichedEvent := manager.EnrichEventForRetry(event, error)

		metadata := enrichedEvent.Metadata()
		retryHistory, exists := metadata["retry_history"]
		assert.True(t, exists)

		history, ok := retryHistory.([]map[string]interface{})
		assert.True(t, ok)
		assert.Len(t, history, 2) // Should have both attempts

		// Verify latest retry entry
		latestRetry := history[1]
		assert.Equal(t, 2, latestRetry["attempt"])
		assert.Equal(t, "second error", latestRetry["error"])
		assert.Equal(t, "SecondHandler", latestRetry["handler"])
	})
}

// TestRetryPolicyManager_Statistics tests retry statistics tracking
func TestRetryPolicyManager_Statistics(t *testing.T) {
	config := DefaultRedisStreamConfig()
	manager, err := NewRetryPolicyManager(config)
	require.NoError(t, err)

	t.Run("should track retry statistics", func(t *testing.T) {
		// Record some retry events
		manager.RecordRetryAttempt("stream1", "handler1", 1, "connection_error")
		manager.RecordRetryAttempt("stream1", "handler1", 2, "timeout_error")
		manager.RecordRetryAttempt("stream2", "handler2", 1, "validation_error")
		manager.RecordRetrySuccess("stream1", "handler1", 3)
		manager.RecordRetryExhausted("stream2", "handler2", 3, "persistent_error")

		stats := manager.GetRetryStatistics()

		assert.Equal(t, int64(3), stats.TotalRetryAttempts)
		assert.Equal(t, int64(1), stats.SuccessfulRetries)
		assert.Equal(t, int64(1), stats.ExhaustedRetries)

		// Verify per-stream stats
		assert.Contains(t, stats.RetriesByStream, "stream1")
		assert.Contains(t, stats.RetriesByStream, "stream2")

		// Verify per-handler stats
		assert.Contains(t, stats.RetriesByHandler, "handler1")
		assert.Contains(t, stats.RetriesByHandler, "handler2")
	})

	t.Run("should calculate retry success rate", func(t *testing.T) {
		// Reset manager for clean test
		manager, _ = NewRetryPolicyManager(config)

		// Record test data: 7 attempts, 3 successes, 2 exhausted
		for i := 0; i < 7; i++ {
			manager.RecordRetryAttempt("test_stream", "test_handler", 1, "error")
		}
		for i := 0; i < 3; i++ {
			manager.RecordRetrySuccess("test_stream", "test_handler", 2)
		}
		for i := 0; i < 2; i++ {
			manager.RecordRetryExhausted("test_stream", "test_handler", 3, "error")
		}

		successRate := manager.GetRetrySuccessRate("test_stream")
		// Success rate = successful retries / (successful + exhausted)
		// 3 / (3 + 2) = 0.6
		assert.InDelta(t, 0.6, successRate, 0.001)

		overallRate := manager.GetOverallRetrySuccessRate()
		assert.InDelta(t, 0.6, overallRate, 0.001)
	})

	t.Run("should identify most common retry reasons", func(t *testing.T) {
		manager, _ = NewRetryPolicyManager(config)

		// Record various retry reasons
		manager.RecordRetryAttempt("stream", "handler", 1, "connection_error")
		manager.RecordRetryAttempt("stream", "handler", 1, "connection_error")
		manager.RecordRetryAttempt("stream", "handler", 1, "connection_error")
		manager.RecordRetryAttempt("stream", "handler", 1, "timeout_error")
		manager.RecordRetryAttempt("stream", "handler", 1, "timeout_error")
		manager.RecordRetryAttempt("stream", "handler", 1, "validation_error")

		topReasons := manager.GetTopRetryReasons(2)

		assert.Len(t, topReasons, 2)
		assert.Equal(t, "connection_error", topReasons[0].Reason)
		assert.Equal(t, int64(3), topReasons[0].Count)
		assert.Equal(t, "timeout_error", topReasons[1].Reason)
		assert.Equal(t, int64(2), topReasons[1].Count)
	})
}

// TestRetryPolicyManager_Integration tests retry policy integration
func TestRetryPolicyManager_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	testContainer, err := NewTestRedisContainer(ctx)
	require.NoError(t, err)
	defer testContainer.Close(ctx)

	t.Run("should handle retry workflow end-to-end", func(t *testing.T) {
		config := testContainer.config
		config.Retry.MaxAttempts = 2
		config.Retry.InitialDelay = 50 * time.Millisecond

		retryManager, err := NewRetryPolicyManager(config)
		require.NoError(t, err)

		// Create an event
		baseOptions := cqrs.Options().
			WithAggregateID("integration-123").
			WithAggregateType("IntegrationTest")

		event := cqrs.NewBaseDomainEventMessage(
			"IntegrationEvent",
			map[string]interface{}{"data": "test"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		// Simulate first failure
		firstError := &ProcessingError{
			Error:     "first failure",
			Handler:   "IntegrationHandler",
			Timestamp: time.Now(),
		}

		// Should retry
		shouldRetry := retryManager.ShouldRetry(event, firstError)
		assert.True(t, shouldRetry)

		// Enrich for retry
		retryEvent := retryManager.EnrichEventForRetry(event, firstError)
		assert.Equal(t, 1, retryEvent.Metadata()["retry_count"])

		// Simulate second failure
		secondError := &ProcessingError{
			Error:     "second failure",
			Handler:   "IntegrationHandler",
			Timestamp: time.Now(),
		}

		// Should still retry (attempt 2 of 2)
		shouldRetry = retryManager.ShouldRetry(retryEvent, secondError)
		assert.True(t, shouldRetry)

		// Enrich for second retry
		secondRetryEvent := retryManager.EnrichEventForRetry(retryEvent, secondError)
		assert.Equal(t, 2, secondRetryEvent.Metadata()["retry_count"])

		// Simulate third failure
		thirdError := &ProcessingError{
			Error:     "third failure",
			Handler:   "IntegrationHandler",
			Timestamp: time.Now(),
		}

		// Should not retry (exhausted)
		shouldRetry = retryManager.ShouldRetry(secondRetryEvent, thirdError)
		assert.False(t, shouldRetry)

		// Verify statistics
		stats := retryManager.GetRetryStatistics()
		assert.Greater(t, stats.TotalRetryAttempts, int64(0))
	})
}

// TestRetryPolicyManager_CustomPolicies tests custom retry policies
func TestRetryPolicyManager_CustomPolicies(t *testing.T) {
	config := DefaultRedisStreamConfig()
	manager, err := NewRetryPolicyManager(config)
	require.NoError(t, err)

	t.Run("should support per-handler retry policies", func(t *testing.T) {
		// Define custom policy for specific handler
		customPolicy := &RetryPolicy{
			MaxAttempts:   5,
			InitialDelay:  200 * time.Millisecond,
			MaxDelay:      10 * time.Second,
			BackoffType:   cqrs.ExponentialBackoff,
			BackoffFactor: 1.5,
		}

		err := manager.SetHandlerRetryPolicy("CriticalHandler", customPolicy)
		assert.NoError(t, err)

		retrievedPolicy := manager.GetHandlerRetryPolicy("CriticalHandler")
		assert.Equal(t, customPolicy.MaxAttempts, retrievedPolicy.MaxAttempts)
		assert.Equal(t, customPolicy.InitialDelay, retrievedPolicy.InitialDelay)
		assert.Equal(t, customPolicy.BackoffType, retrievedPolicy.BackoffType)

		// Non-existent handler should return default policy
		defaultPolicy := manager.GetHandlerRetryPolicy("NonExistentHandler")
		assert.Equal(t, config.Retry.MaxAttempts, defaultPolicy.MaxAttempts)
	})

	t.Run("should support per-event-type retry policies", func(t *testing.T) {
		customPolicy := &RetryPolicy{
			MaxAttempts:   1, // Critical events get only one retry
			InitialDelay:  1 * time.Second,
			MaxDelay:      1 * time.Second,
			BackoffType:   cqrs.FixedBackoff,
			BackoffFactor: 1.0,
		}

		err := manager.SetEventTypeRetryPolicy("CriticalEvent", customPolicy)
		assert.NoError(t, err)

		retrievedPolicy := manager.GetEventTypeRetryPolicy("CriticalEvent")
		assert.Equal(t, 1, retrievedPolicy.MaxAttempts)
		assert.Equal(t, cqrs.FixedBackoff, retrievedPolicy.BackoffType)
	})
}

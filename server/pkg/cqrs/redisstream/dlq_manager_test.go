package redisstream

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cqrs"
)

// TestDLQManager_Creation tests DLQ manager creation
func TestDLQManager_Creation(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Stream.DLQEnabled = true

	t.Run("should create DLQ manager with valid config", func(t *testing.T) {
		manager, err := NewDLQManager(config)
		assert.NoError(t, err)
		assert.NotNil(t, manager)
		assert.True(t, manager.IsDLQEnabled())
	})

	t.Run("should handle disabled DLQ", func(t *testing.T) {
		configNoDLQ := DefaultRedisStreamConfig()
		configNoDLQ.Stream.DLQEnabled = false

		manager, err := NewDLQManager(configNoDLQ)
		assert.NoError(t, err)
		assert.NotNil(t, manager)
		assert.False(t, manager.IsDLQEnabled())
	})

	t.Run("should fail with nil config", func(t *testing.T) {
		_, err := NewDLQManager(nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidConfiguration)
	})
}

// TestDLQManager_StreamNaming tests DLQ stream naming
func TestDLQManager_StreamNaming(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Stream.DLQEnabled = true
	config.Stream.StreamPrefix = "test_events"
	config.Stream.DLQSuffix = "dlq"

	manager, err := NewDLQManager(config)
	require.NoError(t, err)

	t.Run("should generate correct DLQ stream names", func(t *testing.T) {
		testCases := []struct {
			originalStream string
			expectedDLQ    string
		}{
			{
				originalStream: "test_events:domain_event:normal:user",
				expectedDLQ:    "test_events:domain_event:normal:user:dlq",
			},
			{
				originalStream: "test_events:system_event:high:system",
				expectedDLQ:    "test_events:system_event:high:system:dlq",
			},
			{
				originalStream: "simple_stream",
				expectedDLQ:    "simple_stream:dlq",
			},
		}

		for _, tc := range testCases {
			dlqStream := manager.GetDLQStreamName(tc.originalStream)
			assert.Equal(t, tc.expectedDLQ, dlqStream,
				"Failed for original stream: %s", tc.originalStream)
		}
	})

	t.Run("should generate DLQ consumer group names", func(t *testing.T) {
		originalGroup := "service_projection_normal_cg"
		dlqGroup := manager.GetDLQConsumerGroupName(originalGroup)
		expectedDLQGroup := "service_projection_normal_cg_dlq"
		assert.Equal(t, expectedDLQGroup, dlqGroup)
	})
}

// TestDLQManager_FailureDetection tests failure detection logic
func TestDLQManager_FailureDetection(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Stream.DLQEnabled = true
	config.Retry.MaxAttempts = 3

	manager, err := NewDLQManager(config)
	require.NoError(t, err)

	t.Run("should detect when event should be moved to DLQ", func(t *testing.T) {
		// Create a failed event
		baseOptions := cqrs.Options().
			WithAggregateID("failed-123").
			WithAggregateType("FailedAggregate").
			WithMetadata(map[string]interface{}{
				"retry_count":   3,
				"max_retries":   3,
				"last_error":    "processing failed",
				"first_failure": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			})

		event := cqrs.NewBaseDomainEventMessage(
			"FailedEvent",
			map[string]interface{}{"data": "test"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		shouldMoveToDLQ := manager.ShouldMoveToDLQ(event)
		assert.True(t, shouldMoveToDLQ)
	})

	t.Run("should not move event to DLQ if retries not exhausted", func(t *testing.T) {
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

		shouldMoveToDLQ := manager.ShouldMoveToDLQ(event)
		assert.False(t, shouldMoveToDLQ)
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

		shouldMoveToDLQ := manager.ShouldMoveToDLQ(event)
		assert.False(t, shouldMoveToDLQ)
	})
}

// TestDLQManager_EventEnrichment tests DLQ event enrichment
func TestDLQManager_EventEnrichment(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Stream.DLQEnabled = true

	manager, err := NewDLQManager(config)
	require.NoError(t, err)

	t.Run("should enrich event with DLQ metadata", func(t *testing.T) {
		originalError := &ProcessingError{
			Error:      "database connection failed",
			Handler:    "UserProjectionHandler",
			Timestamp:  time.Now(),
			RetryCount: 3,
			StreamName: "events:domain:normal:user",
			MessageID:  "1234567890-0",
		}

		baseOptions := cqrs.Options().
			WithAggregateID("enrich-123").
			WithAggregateType("EnrichAggregate")

		originalEvent := cqrs.NewBaseDomainEventMessage(
			"EnrichEvent",
			map[string]interface{}{"data": "test"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		enrichedEvent := manager.EnrichEventForDLQ(originalEvent, originalError)

		// Verify DLQ metadata was added
		metadata := enrichedEvent.Metadata()
		assert.Contains(t, metadata, "dlq_reason")
		assert.Contains(t, metadata, "dlq_timestamp")
		assert.Contains(t, metadata, "dlq_original_stream")
		assert.Contains(t, metadata, "dlq_original_handler")
		assert.Contains(t, metadata, "dlq_retry_count")
		assert.Contains(t, metadata, "dlq_original_error")

		assert.Equal(t, "max_retries_exceeded", metadata["dlq_reason"])
		assert.Equal(t, "UserProjectionHandler", metadata["dlq_original_handler"])
		assert.Equal(t, "events:domain:normal:user", metadata["dlq_original_stream"])
		assert.Equal(t, 3, metadata["dlq_retry_count"])
	})

	t.Run("should preserve original event data", func(t *testing.T) {
		originalData := map[string]interface{}{
			"user_id": "user-123",
			"action":  "login",
			"metadata": map[string]interface{}{
				"ip_address": "192.168.1.1",
				"user_agent": "Mozilla/5.0",
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
			Error:      "test error",
			Handler:    "TestHandler",
			Timestamp:  time.Now(),
			RetryCount: 1,
		}

		enrichedEvent := manager.EnrichEventForDLQ(originalEvent, error)

		// Verify original data is preserved
		assert.Equal(t, originalEvent.EventType(), enrichedEvent.EventType())
		assert.Equal(t, originalEvent.ID(), enrichedEvent.ID())
		assert.Equal(t, originalEvent.EventData(), enrichedEvent.EventData())

		// Verify original metadata is preserved and DLQ metadata is added
		originalMetadata := originalEvent.Metadata()
		enrichedMetadata := enrichedEvent.Metadata()

		for key, value := range originalMetadata {
			assert.Equal(t, value, enrichedMetadata[key])
		}
	})
}

// TestDLQManager_Statistics tests DLQ statistics tracking
func TestDLQManager_Statistics(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Stream.DLQEnabled = true

	manager, err := NewDLQManager(config)
	require.NoError(t, err)

	t.Run("should track DLQ statistics", func(t *testing.T) {
		// Record some DLQ events
		manager.RecordDLQEvent("stream1", "handler1", "connection_error")
		manager.RecordDLQEvent("stream1", "handler2", "timeout_error")
		manager.RecordDLQEvent("stream2", "handler1", "validation_error")

		stats := manager.GetDLQStatistics()

		assert.Equal(t, int64(3), stats.TotalDLQEvents)
		assert.Len(t, stats.EventsByStream, 2)
		assert.Len(t, stats.EventsByHandler, 2)
		assert.Len(t, stats.EventsByReason, 3)

		// Verify stream statistics
		assert.Equal(t, int64(2), stats.EventsByStream["stream1"])
		assert.Equal(t, int64(1), stats.EventsByStream["stream2"])

		// Verify handler statistics
		assert.Equal(t, int64(2), stats.EventsByHandler["handler1"])
		assert.Equal(t, int64(1), stats.EventsByHandler["handler2"])

		// Verify reason statistics
		assert.Equal(t, int64(1), stats.EventsByReason["connection_error"])
		assert.Equal(t, int64(1), stats.EventsByReason["timeout_error"])
		assert.Equal(t, int64(1), stats.EventsByReason["validation_error"])
	})

	t.Run("should calculate DLQ rates", func(t *testing.T) {
		manager = &dlqManager{
			config: config,
			stats:  &DLQStatistics{},
		}

		// Record some events
		for i := 0; i < 100; i++ {
			manager.RecordProcessedEvent("stream1")
		}

		for i := 0; i < 5; i++ {
			manager.RecordDLQEvent("stream1", "handler1", "error")
		}

		rate := manager.GetDLQRate("stream1")
		expectedRate := 5.0 / 105.0 // 5 DLQ events out of 105 total events
		assert.InDelta(t, expectedRate, rate, 0.001)

		overallRate := manager.GetOverallDLQRate()
		assert.InDelta(t, expectedRate, overallRate, 0.001)
	})

	t.Run("should identify top error reasons", func(t *testing.T) {
		manager = &dlqManager{
			config: config,
			stats: &DLQStatistics{
				EventsByReason: make(map[string]int64),
			},
		}

		// Record various errors with different frequencies
		for i := 0; i < 10; i++ {
			manager.RecordDLQEvent("stream1", "handler1", "connection_error")
		}
		for i := 0; i < 5; i++ {
			manager.RecordDLQEvent("stream1", "handler1", "timeout_error")
		}
		for i := 0; i < 3; i++ {
			manager.RecordDLQEvent("stream1", "handler1", "validation_error")
		}
		for i := 0; i < 1; i++ {
			manager.RecordDLQEvent("stream1", "handler1", "unknown_error")
		}

		topReasons := manager.GetTopErrorReasons(3)

		assert.Len(t, topReasons, 3)
		assert.Equal(t, "connection_error", topReasons[0].Reason)
		assert.Equal(t, int64(10), topReasons[0].Count)
		assert.Equal(t, "timeout_error", topReasons[1].Reason)
		assert.Equal(t, int64(5), topReasons[1].Count)
		assert.Equal(t, "validation_error", topReasons[2].Reason)
		assert.Equal(t, int64(3), topReasons[2].Count)
	})
}

// TestDLQManager_Integration tests DLQ integration with EventBus
func TestDLQManager_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	testContainer, err := NewTestRedisContainer(ctx)
	require.NoError(t, err)
	defer testContainer.Close(ctx)

	t.Run("should move failed events to DLQ", func(t *testing.T) {
		config := testContainer.config
		config.Stream.DLQEnabled = true
		config.Retry.MaxAttempts = 2

		dlqManager, err := NewDLQManager(config)
		require.NoError(t, err)

		// Create EventBus with DLQ support
		eventBus, err := NewRedisStreamEventBus(testContainer.client, config)
		require.NoError(t, err)

		err = eventBus.Start(ctx)
		require.NoError(t, err)
		defer eventBus.Stop(ctx)

		// Create a handler that always fails
		failingHandler := &FailingEventHandler{
			name:        "failing-handler",
			handlerType: cqrs.ProjectionHandler,
			eventTypes:  []string{"FailingEvent"},
			failCount:   0,
		}

		subID, err := eventBus.Subscribe("FailingEvent", failingHandler)
		require.NoError(t, err)
		defer eventBus.Unsubscribe(subID)

		time.Sleep(100 * time.Millisecond)

		// Publish an event that will fail
		baseOptions := cqrs.Options().
			WithAggregateID("failing-123").
			WithAggregateType("FailingAggregate")

		event := cqrs.NewBaseDomainEventMessage(
			"FailingEvent",
			map[string]interface{}{"data": "will_fail"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		err = eventBus.Publish(ctx, event)
		require.NoError(t, err)

		// Wait for processing attempts and eventual DLQ
		time.Sleep(5 * time.Second)

		// Verify the event was moved to DLQ
		stats := dlqManager.GetDLQStatistics()
		assert.Greater(t, stats.TotalDLQEvents, int64(0))
	})
}

// TestDLQManager_DisabledMode tests behavior when DLQ is disabled
func TestDLQManager_DisabledMode(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Stream.DLQEnabled = false

	manager, err := NewDLQManager(config)
	require.NoError(t, err)

	t.Run("should handle disabled DLQ gracefully", func(t *testing.T) {
		// All DLQ operations should be no-ops when disabled
		assert.False(t, manager.IsDLQEnabled())

		// These should not panic or cause errors
		dlqStream := manager.GetDLQStreamName("test_stream")
		assert.Equal(t, "", dlqStream) // Should return empty when disabled

		baseOptions := cqrs.Options().
			WithAggregateID("disabled-123").
			WithAggregateType("DisabledAggregate")

		event := cqrs.NewBaseDomainEventMessage(
			"DisabledEvent",
			map[string]interface{}{"data": "test"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		shouldMove := manager.ShouldMoveToDLQ(event)
		assert.False(t, shouldMove) // Should never move to DLQ when disabled

		// Statistics should be empty
		stats := manager.GetDLQStatistics()
		assert.Equal(t, int64(0), stats.TotalDLQEvents)
	})
}

// FailingEventHandler is a test handler that always fails
type FailingEventHandler struct {
	name        string
	handlerType cqrs.HandlerType
	eventTypes  []string
	failCount   int
}

func (h *FailingEventHandler) Handle(ctx context.Context, event cqrs.EventMessage) error {
	h.failCount++
	return fmt.Errorf("handler intentionally failed (attempt %d)", h.failCount)
}

func (h *FailingEventHandler) CanHandle(eventType string) bool {
	for _, et := range h.eventTypes {
		if et == eventType {
			return true
		}
	}
	return false
}

func (h *FailingEventHandler) GetHandlerName() string {
	return h.name
}

func (h *FailingEventHandler) GetHandlerType() cqrs.HandlerType {
	return h.handlerType
}

func (h *FailingEventHandler) GetFailCount() int {
	return h.failCount
}

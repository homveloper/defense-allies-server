package redisstream

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"defense-allies-server/pkg/cqrs"
)

// TestPriorityStreamManager tests priority-based stream management
func TestPriorityStreamManager_Creation(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Stream.EnablePriorityStreams = true

	t.Run("should create priority stream manager with valid config", func(t *testing.T) {
		manager, err := NewPriorityStreamManager(config)
		assert.NoError(t, err)
		assert.NotNil(t, manager)
		assert.True(t, manager.IsPriorityEnabled())
	})

	t.Run("should handle disabled priority streams", func(t *testing.T) {
		configNoPriority := DefaultRedisStreamConfig()
		configNoPriority.Stream.EnablePriorityStreams = false

		manager, err := NewPriorityStreamManager(configNoPriority)
		assert.NoError(t, err)
		assert.NotNil(t, manager)
		assert.False(t, manager.IsPriorityEnabled())
	})

	t.Run("should fail with nil config", func(t *testing.T) {
		_, err := NewPriorityStreamManager(nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidConfiguration)
	})
}

// TestPriorityStreamManager_StreamNaming tests stream name generation
func TestPriorityStreamManager_StreamNaming(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Stream.EnablePriorityStreams = true
	config.Stream.StreamPrefix = "test_events"

	manager, err := NewPriorityStreamManager(config)
	require.NoError(t, err)

	t.Run("should generate correct stream names for different priorities", func(t *testing.T) {
		testCases := []struct {
			priority     cqrs.EventPriority
			category     cqrs.EventCategory
			partitionKey string
			expectedName string
		}{
			{
				priority:     cqrs.PriorityCritical,
				category:     cqrs.DomainEvent,
				partitionKey: "user",
				expectedName: "test_events:domain_event:critical:user",
			},
			{
				priority:     cqrs.PriorityHigh,
				category:     cqrs.SystemEvent,
				partitionKey: "system",
				expectedName: "test_events:system_event:high:system",
			},
			{
				priority:     cqrs.PriorityNormal,
				category:     cqrs.UserAction,
				partitionKey: "default",
				expectedName: "test_events:user_action:normal:default",
			},
			{
				priority:     cqrs.PriorityLow,
				category:     cqrs.IntegrationEvent,
				partitionKey: "integration",
				expectedName: "test_events:integration_event:low:integration",
			},
		}

		for _, tc := range testCases {
			streamName := manager.GetStreamName(tc.priority, tc.category, tc.partitionKey)
			assert.Equal(t, tc.expectedName, streamName,
				"Failed for priority=%s, category=%s, partition=%s",
				tc.priority, tc.category, tc.partitionKey)
		}
	})

	t.Run("should handle empty partition key", func(t *testing.T) {
		streamName := manager.GetStreamName(cqrs.PriorityNormal, cqrs.DomainEvent, "")
		assert.Equal(t, "test_events:domain_event:normal:default", streamName)
	})

	t.Run("should generate consumer group names", func(t *testing.T) {
		groupName := manager.GetConsumerGroupName(cqrs.PriorityHigh, "test_service", cqrs.ProjectionHandler)
		expectedName := "test_service_projection_high_cg"
		assert.Equal(t, expectedName, groupName)
	})
}

// TestPriorityStreamManager_StreamRouting tests event routing logic
func TestPriorityStreamManager_StreamRouting(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Stream.EnablePriorityStreams = true

	manager, err := NewPriorityStreamManager(config)
	require.NoError(t, err)

	t.Run("should route events to correct streams based on priority", func(t *testing.T) {
		// Create test events with different priorities
		baseOptions := cqrs.Options().
			WithAggregateID("test-123").
			WithAggregateType("TestAggregate")

		events := []struct {
			priority       cqrs.EventPriority
			expectedStream string
		}{
			{cqrs.PriorityCritical, "events:domain_event:critical:TestAggregate"},
			{cqrs.PriorityHigh, "events:domain_event:high:TestAggregate"},
			{cqrs.PriorityNormal, "events:domain_event:normal:TestAggregate"},
			{cqrs.PriorityLow, "events:domain_event:low:TestAggregate"},
		}

		for _, eventTest := range events {
			domainOptions := &cqrs.BaseDomainEventMessageOptions{}
			domainOptions.Priority = &eventTest.priority

			event := cqrs.NewBaseDomainEventMessage(
				"TestEvent",
				map[string]interface{}{"test": "data"},
				[]*cqrs.BaseEventMessageOptions{baseOptions},
				domainOptions,
			)

			routingInfo := manager.GetRoutingInfo(event)
			assert.Equal(t, eventTest.expectedStream, routingInfo.StreamName)
			assert.Equal(t, eventTest.priority, routingInfo.Priority)
		}
	})

	t.Run("should handle events without explicit priority", func(t *testing.T) {
		baseOptions := cqrs.Options().
			WithAggregateID("test-123").
			WithAggregateType("TestAggregate")

		// Event without explicit priority should default to Normal
		event := cqrs.NewBaseDomainEventMessage(
			"TestEvent",
			map[string]interface{}{"test": "data"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		routingInfo := manager.GetRoutingInfo(event)
		assert.Equal(t, cqrs.PriorityNormal, routingInfo.Priority)
		assert.Contains(t, routingInfo.StreamName, ":normal:")
	})
}

// TestPriorityStreamManager_ProcessingOrder tests priority-based processing
func TestPriorityStreamManager_ProcessingOrder(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Stream.EnablePriorityStreams = true

	manager, err := NewPriorityStreamManager(config)
	require.NoError(t, err)

	t.Run("should return streams in priority order", func(t *testing.T) {
		partitionKey := "test"
		category := cqrs.DomainEvent

		priorityOrder := manager.GetStreamsByPriority(category, partitionKey)

		// Should return streams in order: Critical, High, Normal, Low
		expectedOrder := []cqrs.EventPriority{
			cqrs.PriorityCritical,
			cqrs.PriorityHigh,
			cqrs.PriorityNormal,
			cqrs.PriorityLow,
		}

		assert.Len(t, priorityOrder, len(expectedOrder))

		for i, stream := range priorityOrder {
			assert.Equal(t, expectedOrder[i], stream.Priority)
			assert.Contains(t, stream.StreamName, expectedOrder[i].String())
		}
	})

	t.Run("should filter streams by minimum priority", func(t *testing.T) {
		partitionKey := "test"
		category := cqrs.DomainEvent

		// Get streams with minimum priority High (should exclude Low and Normal)
		filteredStreams := manager.GetStreamsWithMinPriority(category, partitionKey, cqrs.PriorityHigh)

		assert.Len(t, filteredStreams, 2) // Critical and High only
		assert.Equal(t, cqrs.PriorityCritical, filteredStreams[0].Priority)
		assert.Equal(t, cqrs.PriorityHigh, filteredStreams[1].Priority)
	})
}

// TestPriorityStreamManager_ConsumerManagement tests consumer group management
func TestPriorityStreamManager_ConsumerManagement(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Stream.EnablePriorityStreams = true
	config.Consumer.ServiceName = "test_service"

	manager, err := NewPriorityStreamManager(config)
	require.NoError(t, err)

	t.Run("should create consumer configurations for all priorities", func(t *testing.T) {
		handlerType := cqrs.ProjectionHandler
		partitionKey := "user"
		category := cqrs.DomainEvent

		consumerConfigs := manager.GetConsumerConfigurations(category, partitionKey, handlerType)

		// Should have one configuration per priority level
		assert.Len(t, consumerConfigs, 4)

		// Verify each priority is covered
		priorities := make(map[cqrs.EventPriority]bool)
		for _, config := range consumerConfigs {
			priorities[config.Priority] = true

			// Verify stream and group names are correct
			expectedStreamName := manager.GetStreamName(config.Priority, category, partitionKey)
			assert.Equal(t, expectedStreamName, config.StreamName)

			expectedGroupName := manager.GetConsumerGroupName(config.Priority, "test_service", handlerType)
			assert.Equal(t, expectedGroupName, config.ConsumerGroup)
		}

		// All priorities should be covered
		assert.True(t, priorities[cqrs.PriorityCritical])
		assert.True(t, priorities[cqrs.PriorityHigh])
		assert.True(t, priorities[cqrs.PriorityNormal])
		assert.True(t, priorities[cqrs.PriorityLow])
	})

	t.Run("should handle different handler types", func(t *testing.T) {
		handlerTypes := []cqrs.HandlerType{
			cqrs.ProjectionHandler,
			cqrs.ProcessManagerHandler,
			cqrs.SagaHandler,
			cqrs.NotificationHandler,
		}

		for _, handlerType := range handlerTypes {
			consumerConfigs := manager.GetConsumerConfigurations(
				cqrs.DomainEvent,
				"test",
				handlerType,
			)

			assert.Len(t, consumerConfigs, 4) // One per priority

			// Verify handler type is reflected in consumer group name
			for _, config := range consumerConfigs {
				assert.Contains(t, config.ConsumerGroup, handlerType.String())
			}
		}
	})
}

// TestPriorityStreamManager_Metrics tests priority-specific metrics
func TestPriorityStreamManager_Metrics(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Stream.EnablePriorityStreams = true

	manager, err := NewPriorityStreamManager(config)
	require.NoError(t, err)

	t.Run("should track metrics per priority", func(t *testing.T) {
		// Simulate some activity
		manager.RecordPublishedEvent(cqrs.PriorityCritical, "test_stream")
		manager.RecordPublishedEvent(cqrs.PriorityHigh, "test_stream")
		manager.RecordPublishedEvent(cqrs.PriorityHigh, "test_stream")
		manager.RecordProcessedEvent(cqrs.PriorityCritical, "test_stream", 10*time.Millisecond)

		metrics := manager.GetPriorityMetrics()

		// Verify metrics are tracked correctly
		assert.Equal(t, int64(1), metrics[cqrs.PriorityCritical].PublishedEvents)
		assert.Equal(t, int64(2), metrics[cqrs.PriorityHigh].PublishedEvents)
		assert.Equal(t, int64(1), metrics[cqrs.PriorityCritical].ProcessedEvents)
		assert.Equal(t, int64(0), metrics[cqrs.PriorityHigh].ProcessedEvents)

		// Verify latency tracking
		assert.Equal(t, 10*time.Millisecond, metrics[cqrs.PriorityCritical].AverageLatency)
	})

	t.Run("should calculate priority ratios", func(t *testing.T) {
		// Reset and add some test data
		manager = &priorityStreamManager{config: config, metrics: make(map[cqrs.EventPriority]*PriorityMetrics)}

		manager.RecordPublishedEvent(cqrs.PriorityCritical, "test")
		manager.RecordPublishedEvent(cqrs.PriorityHigh, "test")
		manager.RecordPublishedEvent(cqrs.PriorityHigh, "test")
		manager.RecordPublishedEvent(cqrs.PriorityNormal, "test")
		manager.RecordPublishedEvent(cqrs.PriorityNormal, "test")
		manager.RecordPublishedEvent(cqrs.PriorityNormal, "test")
		manager.RecordPublishedEvent(cqrs.PriorityNormal, "test")

		ratios := manager.GetPriorityRatios()

		// Total: 8 events
		// Critical: 1/8 = 0.125
		// High: 2/8 = 0.25
		// Normal: 4/8 = 0.5
		// Low: 0/8 = 0

		assert.InDelta(t, 0.125, ratios[cqrs.PriorityCritical], 0.001)
		assert.InDelta(t, 0.25, ratios[cqrs.PriorityHigh], 0.001)
		assert.InDelta(t, 0.5, ratios[cqrs.PriorityNormal], 0.001)
		assert.InDelta(t, 0.0, ratios[cqrs.PriorityLow], 0.001)
	})
}

// TestPriorityStreamManager_DisabledMode tests behavior when priority is disabled
func TestPriorityStreamManager_DisabledMode(t *testing.T) {
	config := DefaultRedisStreamConfig()
	config.Stream.EnablePriorityStreams = false

	manager, err := NewPriorityStreamManager(config)
	require.NoError(t, err)

	t.Run("should fallback to single stream when disabled", func(t *testing.T) {
		// All events should go to the same stream regardless of priority
		partitionKey := "test"
		category := cqrs.DomainEvent

		streamName1 := manager.GetStreamName(cqrs.PriorityCritical, category, partitionKey)
		streamName2 := manager.GetStreamName(cqrs.PriorityHigh, category, partitionKey)
		streamName3 := manager.GetStreamName(cqrs.PriorityNormal, category, partitionKey)
		streamName4 := manager.GetStreamName(cqrs.PriorityLow, category, partitionKey)

		// All should be the same (no priority in name)
		assert.Equal(t, streamName1, streamName2)
		assert.Equal(t, streamName2, streamName3)
		assert.Equal(t, streamName3, streamName4)

		// Should not contain priority information
		assert.NotContains(t, streamName1, ":critical:")
		assert.NotContains(t, streamName1, ":high:")
		assert.NotContains(t, streamName1, ":normal:")
		assert.NotContains(t, streamName1, ":low:")
	})

	t.Run("should return single consumer configuration when disabled", func(t *testing.T) {
		consumerConfigs := manager.GetConsumerConfigurations(
			cqrs.DomainEvent,
			"test",
			cqrs.ProjectionHandler,
		)

		// Should return only one configuration when priority is disabled
		assert.Len(t, consumerConfigs, 1)
		assert.Equal(t, cqrs.PriorityNormal, consumerConfigs[0].Priority) // Default priority
	})
}

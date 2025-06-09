package redisstream

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"cqrs"
)

// TestRedisContainer manages Redis container for integration tests
type TestRedisContainer struct {
	container testcontainers.Container
	client    redis.UniversalClient
	config    *RedisStreamConfig
}

// NewTestRedisContainer creates a new Redis container for testing
func NewTestRedisContainer(ctx context.Context) (*TestRedisContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		return nil, err
	}

	config := DefaultRedisStreamConfig()
	config.Redis.Addr = host + ":" + port.Port()

	client := config.CreateRedisClient()

	return &TestRedisContainer{
		container: container,
		client:    client,
		config:    config,
	}, nil
}

// Close cleans up the test container
func (tc *TestRedisContainer) Close(ctx context.Context) error {
	if tc.client != nil {
		tc.client.Close()
	}
	if tc.container != nil {
		return tc.container.Terminate(ctx)
	}
	return nil
}

// TestEventBusCreation tests event bus creation
func TestRedisStreamEventBus_Creation(t *testing.T) {
	ctx := context.Background()

	// Test with invalid configuration
	t.Run("should fail with invalid config", func(t *testing.T) {
		invalidConfig := &RedisStreamConfig{}
		_, err := NewRedisStreamEventBus(nil, invalidConfig)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidConfiguration)
	})

	// Test with valid configuration
	t.Run("should succeed with valid config", func(t *testing.T) {
		testContainer, err := NewTestRedisContainer(ctx)
		require.NoError(t, err)
		defer testContainer.Close(ctx)

		eventBus, err := NewRedisStreamEventBus(testContainer.client, testContainer.config)
		assert.NoError(t, err)
		assert.NotNil(t, eventBus)
		assert.False(t, eventBus.IsRunning())
	})
}

// TestEventBusLifecycle tests event bus lifecycle
func TestRedisStreamEventBus_Lifecycle(t *testing.T) {
	ctx := context.Background()

	testContainer, err := NewTestRedisContainer(ctx)
	require.NoError(t, err)
	defer testContainer.Close(ctx)

	eventBus, err := NewRedisStreamEventBus(testContainer.client, testContainer.config)
	require.NoError(t, err)

	t.Run("should start successfully", func(t *testing.T) {
		err := eventBus.Start(ctx)
		assert.NoError(t, err)
		assert.True(t, eventBus.IsRunning())
	})

	t.Run("should not start if already running", func(t *testing.T) {
		err := eventBus.Start(ctx)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrEventBusAlreadyRunning)
	})

	t.Run("should stop successfully", func(t *testing.T) {
		err := eventBus.Stop(ctx)
		assert.NoError(t, err)
		assert.False(t, eventBus.IsRunning())
	})

	t.Run("should not stop if not running", func(t *testing.T) {
		err := eventBus.Stop(ctx)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrEventBusNotRunning)
	})
}

// TestEventPublishing tests event publishing functionality
func TestRedisStreamEventBus_Publishing(t *testing.T) {
	ctx := context.Background()

	testContainer, err := NewTestRedisContainer(ctx)
	require.NoError(t, err)
	defer testContainer.Close(ctx)

	eventBus, err := NewRedisStreamEventBus(testContainer.client, testContainer.config)
	require.NoError(t, err)

	err = eventBus.Start(ctx)
	require.NoError(t, err)
	defer eventBus.Stop(ctx)

	t.Run("should publish single event", func(t *testing.T) {
		baseOptions := cqrs.Options().
			WithAggregateID("test-aggregate-123").
			WithAggregateType("TestAggregate")
		event := cqrs.NewBaseDomainEventMessage(
			"TestEvent",
			map[string]interface{}{"test": "data"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		err := eventBus.Publish(ctx, event)
		assert.NoError(t, err)
	})

	t.Run("should publish event with options", func(t *testing.T) {
		baseOptions := cqrs.Options().
			WithAggregateID("test-aggregate-456").
			WithAggregateType("TestAggregate")
		event := cqrs.NewBaseDomainEventMessage(
			"TestEventWithOptions",
			map[string]interface{}{"test": "data"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		options := cqrs.EventPublishOptions{
			Persistent:   true,
			Immediate:    false,
			Async:        true,
			Priority:     cqrs.PriorityHigh,
			PartitionKey: "test-partition",
		}

		err := eventBus.Publish(ctx, event, options)
		assert.NoError(t, err)
	})

	t.Run("should publish batch events", func(t *testing.T) {
		events := make([]cqrs.EventMessage, 3)
		for i := 0; i < 3; i++ {
			baseOptions := cqrs.Options().
				WithAggregateID(fmt.Sprintf("batch-aggregate-%d", i)).
				WithAggregateType("BatchTestAggregate")
			event := cqrs.NewBaseDomainEventMessage(
				"BatchTestEvent",
				map[string]interface{}{"index": i},
				[]*cqrs.BaseEventMessageOptions{baseOptions},
			)
			events[i] = event
		}

		err := eventBus.PublishBatch(ctx, events)
		assert.NoError(t, err)
	})

	t.Run("should fail to publish when not running", func(t *testing.T) {
		eventBus.Stop(ctx)

		event := cqrs.NewBaseDomainEventMessage(
			"FailTestEvent",
			map[string]interface{}{"test": "data"},
			nil,
		)

		err := eventBus.Publish(ctx, event)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrEventBusNotRunning)
	})
}

// TestEventSubscription tests event subscription functionality
func TestRedisStreamEventBus_Subscription(t *testing.T) {
	ctx := context.Background()

	testContainer, err := NewTestRedisContainer(ctx)
	require.NoError(t, err)
	defer testContainer.Close(ctx)

	eventBus, err := NewRedisStreamEventBus(testContainer.client, testContainer.config)
	require.NoError(t, err)

	err = eventBus.Start(ctx)
	require.NoError(t, err)
	defer eventBus.Stop(ctx)

	t.Run("should subscribe to specific event type", func(t *testing.T) {
		handler := &TestEventHandler{
			name:        "test-handler",
			handlerType: cqrs.ProjectionHandler,
			eventTypes:  []string{"TestSubscriptionEvent"},
		}

		subID, err := eventBus.Subscribe("TestSubscriptionEvent", handler)
		assert.NoError(t, err)
		assert.NotEmpty(t, subID)

		// Cleanup
		err = eventBus.Unsubscribe(subID)
		assert.NoError(t, err)
	})

	t.Run("should subscribe to all events", func(t *testing.T) {
		handler := &TestEventHandler{
			name:        "test-all-handler",
			handlerType: cqrs.NotificationHandler,
			eventTypes:  []string{},
		}

		subID, err := eventBus.SubscribeAll(handler)
		assert.NoError(t, err)
		assert.NotEmpty(t, subID)

		// Cleanup
		err = eventBus.Unsubscribe(subID)
		assert.NoError(t, err)
	})

	t.Run("should fail to unsubscribe non-existent subscription", func(t *testing.T) {
		err := eventBus.Unsubscribe(cqrs.SubscriptionID("non-existent"))
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSubscriptionNotFound)
	})
}

// TestEventHandling tests end-to-end event handling
func TestRedisStreamEventBus_EventHandling(t *testing.T) {
	ctx := context.Background()

	testContainer, err := NewTestRedisContainer(ctx)
	require.NoError(t, err)
	defer testContainer.Close(ctx)

	eventBus, err := NewRedisStreamEventBus(testContainer.client, testContainer.config)
	require.NoError(t, err)

	err = eventBus.Start(ctx)
	require.NoError(t, err)
	defer eventBus.Stop(ctx)

	t.Run("should handle published event", func(t *testing.T) {
		handler := &TestEventHandler{
			name:        "integration-test-handler",
			handlerType: cqrs.ProjectionHandler,
			eventTypes:  []string{"IntegrationTestEvent"},
		}

		subID, err := eventBus.Subscribe("IntegrationTestEvent", handler)
		require.NoError(t, err)
		defer eventBus.Unsubscribe(subID)

		// Give some time for subscription to be ready
		time.Sleep(100 * time.Millisecond)

		baseOptions := cqrs.Options().
			WithAggregateID("integration-test-123").
			WithAggregateType("IntegrationTestAggregate")
		event := cqrs.NewBaseDomainEventMessage(
			"IntegrationTestEvent",
			map[string]interface{}{"message": "hello world"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		err = eventBus.Publish(ctx, event)
		require.NoError(t, err)

		// Wait for event processing
		assert.Eventually(t, func() bool {
			return handler.ProcessedEventCount() > 0
		}, 5*time.Second, 100*time.Millisecond, "Event should be processed")

		assert.Equal(t, 1, handler.ProcessedEventCount())
		processedEvent := handler.LastProcessedEvent()
		assert.Equal(t, "IntegrationTestEvent", processedEvent.EventType())
		assert.Equal(t, "integration-test-123", processedEvent.ID())
	})
}

// TestEventBusMetrics tests metrics collection
func TestRedisStreamEventBus_Metrics(t *testing.T) {
	ctx := context.Background()

	testContainer, err := NewTestRedisContainer(ctx)
	require.NoError(t, err)
	defer testContainer.Close(ctx)

	eventBus, err := NewRedisStreamEventBus(testContainer.client, testContainer.config)
	require.NoError(t, err)

	err = eventBus.Start(ctx)
	require.NoError(t, err)
	defer eventBus.Stop(ctx)

	t.Run("should collect basic metrics", func(t *testing.T) {
		metrics := eventBus.GetMetrics()
		assert.NotNil(t, metrics)
		assert.GreaterOrEqual(t, metrics.PublishedEvents, int64(0))
		assert.GreaterOrEqual(t, metrics.ProcessedEvents, int64(0))
		assert.GreaterOrEqual(t, metrics.ActiveSubscribers, 0)
	})

	t.Run("should update metrics after publishing", func(t *testing.T) {
		initialMetrics := eventBus.GetMetrics()

		event := cqrs.NewBaseDomainEventMessage(
			"MetricsTestEvent",
			map[string]interface{}{"test": "metrics"},
			nil,
		)

		err := eventBus.Publish(ctx, event)
		require.NoError(t, err)

		updatedMetrics := eventBus.GetMetrics()
		assert.Greater(t, updatedMetrics.PublishedEvents, initialMetrics.PublishedEvents)
	})
}

// TestRedisStreamEventBus_Serialization tests event serialization integration
func TestRedisStreamEventBus_Serialization(t *testing.T) {
	ctx := context.Background()

	testContainer, err := NewTestRedisContainer(ctx)
	require.NoError(t, err)
	defer testContainer.Close(ctx)

	t.Run("should serialize and deserialize events correctly", func(t *testing.T) {
		// Create EventBus with JSON serializer
		jsonSerializer := NewJSONEventSerializer()
		eventBus, err := NewRedisStreamEventBusWithSerializer(testContainer.client, testContainer.config, jsonSerializer)
		require.NoError(t, err)

		err = eventBus.Start(ctx)
		require.NoError(t, err)
		defer eventBus.Stop(ctx)

		// Create a complex event
		complexData := map[string]interface{}{
			"user": map[string]interface{}{
				"name": "John Doe",
				"age":  30,
				"preferences": map[string]interface{}{
					"theme":         "dark",
					"notifications": []string{"email", "push"},
				},
			},
			"scores": []float64{1.1, 2.2, 3.3},
			"metadata": map[string]interface{}{
				"source":  "test",
				"version": 2.0,
			},
		}

		baseOptions := cqrs.Options().
			WithAggregateID("complex-user-123").
			WithAggregateType("ComplexUser").
			WithVersion(1)

		originalEvent := cqrs.NewBaseDomainEventMessage(
			"ComplexUserEvent",
			complexData,
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		// Set up handler to capture the deserialized event
		handler := &TestEventHandler{
			name:        "serialization-test-handler",
			handlerType: cqrs.ProjectionHandler,
			eventTypes:  []string{"ComplexUserEvent"},
		}

		subID, err := eventBus.Subscribe("ComplexUserEvent", handler)
		require.NoError(t, err)
		defer eventBus.Unsubscribe(subID)

		// Give some time for subscription to be ready
		time.Sleep(100 * time.Millisecond)

		// Publish the event
		err = eventBus.Publish(ctx, originalEvent)
		require.NoError(t, err)

		// Wait for event processing
		assert.Eventually(t, func() bool {
			return handler.ProcessedEventCount() > 0
		}, 5*time.Second, 100*time.Millisecond)

		// Verify the deserialized event
		processedEvent := handler.LastProcessedEvent()
		assert.Equal(t, originalEvent.EventType(), processedEvent.EventType())
		assert.Equal(t, originalEvent.ID(), processedEvent.ID())
		assert.Equal(t, originalEvent.Type(), processedEvent.Type())

		// Verify complex event data was preserved
		processedData := processedEvent.EventData().(map[string]interface{})
		user := processedData["user"].(map[string]interface{})
		assert.Equal(t, "John Doe", user["name"])
		assert.Equal(t, float64(30), user["age"]) // JSON numbers are float64

		preferences := user["preferences"].(map[string]interface{})
		assert.Equal(t, "dark", preferences["theme"])

		notifications := preferences["notifications"].([]interface{})
		assert.Len(t, notifications, 2)
		assert.Contains(t, notifications, "email")
		assert.Contains(t, notifications, "push")
	})

	t.Run("should handle different serialization formats", func(t *testing.T) {
		// Test that we can create EventBus with different serializers
		jsonSerializer := NewJSONEventSerializer()
		assert.Equal(t, SerializationFormatJSON, jsonSerializer.Format())

		eventBus, err := NewRedisStreamEventBusWithSerializer(testContainer.client, testContainer.config, jsonSerializer)
		require.NoError(t, err)

		// Verify serializer is set correctly
		assert.Equal(t, jsonSerializer, eventBus.GetSerializer())

		// Test changing serializer before start
		newJsonSerializer := NewJSONEventSerializer()
		err = eventBus.SetSerializer(newJsonSerializer)
		assert.NoError(t, err)
		assert.Equal(t, newJsonSerializer, eventBus.GetSerializer())

		// Start the bus
		err = eventBus.Start(ctx)
		require.NoError(t, err)
		defer eventBus.Stop(ctx)

		// Should not allow changing serializer while running
		err = eventBus.SetSerializer(NewJSONEventSerializer())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot change serializer while event bus is running")
	})

	t.Run("should handle domain event metadata correctly", func(t *testing.T) {
		eventBus, err := NewRedisStreamEventBus(testContainer.client, testContainer.config)
		require.NoError(t, err)

		err = eventBus.Start(ctx)
		require.NoError(t, err)
		defer eventBus.Stop(ctx)

		// Create domain event with metadata
		baseOptions := cqrs.Options().
			WithAggregateID("metadata-test-123").
			WithAggregateType("MetadataTest")

		domainOptions := &cqrs.BaseDomainEventMessageOptions{}
		issuerID := "user-456"
		issuerType := cqrs.UserIssuer
		causationID := "cmd-789"
		correlationID := "corr-101"
		category := cqrs.DomainEvent
		priority := cqrs.PriorityHigh

		domainOptions.IssuerID = &issuerID
		domainOptions.IssuerType = &issuerType
		domainOptions.CausationID = &causationID
		domainOptions.CorrelationID = &correlationID
		domainOptions.Category = &category
		domainOptions.Priority = &priority

		originalEvent := cqrs.NewBaseDomainEventMessage(
			"DomainMetadataEvent",
			map[string]interface{}{"action": "test"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
			domainOptions,
		)

		// Set up handler
		handler := &TestEventHandler{
			name:        "metadata-test-handler",
			handlerType: cqrs.ProjectionHandler,
			eventTypes:  []string{"DomainMetadataEvent"},
		}

		subID, err := eventBus.Subscribe("DomainMetadataEvent", handler)
		require.NoError(t, err)
		defer eventBus.Unsubscribe(subID)

		time.Sleep(100 * time.Millisecond)

		// Publish and verify
		err = eventBus.Publish(ctx, originalEvent)
		require.NoError(t, err)

		assert.Eventually(t, func() bool {
			return handler.ProcessedEventCount() > 0
		}, 5*time.Second, 100*time.Millisecond)

		processedEvent := handler.LastProcessedEvent()

		// Verify it's a domain event
		if domainEvent, ok := processedEvent.(cqrs.DomainEventMessage); ok {
			assert.Equal(t, "user-456", domainEvent.IssuerID())
			assert.Equal(t, cqrs.UserIssuer, domainEvent.IssuerType())
			assert.Equal(t, "cmd-789", domainEvent.CausationID())
			assert.Equal(t, "corr-101", domainEvent.CorrelationID())
			assert.Equal(t, cqrs.DomainEvent, domainEvent.GetEventCategory())
			assert.Equal(t, cqrs.PriorityHigh, domainEvent.GetPriority())
		} else {
			t.Fatal("Expected DomainEventMessage interface")
		}
	})
}

// TestSerializationRegistry_Integration tests serialization registry integration
func TestSerializationRegistry_Integration(t *testing.T) {
	registry := NewSerializationRegistry()

	t.Run("should manage multiple serializers", func(t *testing.T) {
		jsonSerializer := NewJSONEventSerializer()

		// Register JSON serializer
		err := registry.Register(SerializationFormatJSON, jsonSerializer)
		assert.NoError(t, err)

		// Verify registration
		retrievedSerializer, err := registry.Get(SerializationFormatJSON)
		assert.NoError(t, err)
		assert.Equal(t, jsonSerializer, retrievedSerializer)

		// Check supported formats
		formats := registry.SupportedFormats()
		assert.Contains(t, formats, SerializationFormatJSON)
	})

	t.Run("should handle unknown formats gracefully", func(t *testing.T) {
		_, err := registry.Get(SerializationFormat("unknown"))
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSerializationFormatNotSupported)
	})
}

type TestEventHandler struct {
	name               string
	handlerType        cqrs.HandlerType
	eventTypes         []string
	processedEvents    []cqrs.EventMessage
	processedCount     int
	lastProcessedEvent cqrs.EventMessage
}

func (h *TestEventHandler) Handle(ctx context.Context, event cqrs.EventMessage) error {
	h.processedEvents = append(h.processedEvents, event)
	h.processedCount++
	h.lastProcessedEvent = event
	return nil
}

func (h *TestEventHandler) CanHandle(eventType string) bool {
	if len(h.eventTypes) == 0 {
		return true // Handle all events
	}
	for _, et := range h.eventTypes {
		if et == eventType {
			return true
		}
	}
	return false
}

func (h *TestEventHandler) GetHandlerName() string {
	return h.name
}

func (h *TestEventHandler) GetHandlerType() cqrs.HandlerType {
	return h.handlerType
}

func (h *TestEventHandler) ProcessedEventCount() int {
	return h.processedCount
}

func (h *TestEventHandler) LastProcessedEvent() cqrs.EventMessage {
	return h.lastProcessedEvent
}

func (h *TestEventHandler) ProcessedEvents() []cqrs.EventMessage {
	return h.processedEvents
}

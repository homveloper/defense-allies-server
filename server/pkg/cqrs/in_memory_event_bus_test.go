package cqrs

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test event handler implementation
type TestEventHandler struct {
	*BaseEventHandler
	HandledEvents []EventMessage
	HandleFunc    func(ctx context.Context, event EventMessage) error
	mutex         sync.Mutex
}

func NewTestEventHandler(name string, eventTypes []string) *TestEventHandler {
	return &TestEventHandler{
		BaseEventHandler: NewBaseEventHandler(name, ProjectionHandler, eventTypes),
		HandledEvents:    make([]EventMessage, 0),
	}
}

func (h *TestEventHandler) Handle(ctx context.Context, event EventMessage) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.HandledEvents = append(h.HandledEvents, event)

	if h.HandleFunc != nil {
		return h.HandleFunc(ctx, event)
	}

	return nil
}

func (h *TestEventHandler) GetHandledEventCount() int {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	return len(h.HandledEvents)
}

func (h *TestEventHandler) GetLastHandledEvent() EventMessage {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if len(h.HandledEvents) == 0 {
		return nil
	}
	return h.HandledEvents[len(h.HandledEvents)-1]
}

func TestNewInMemoryEventBus(t *testing.T) {
	// Act
	bus := NewInMemoryEventBus()

	// Assert
	assert.NotNil(t, bus)
	assert.False(t, bus.IsRunning())
	assert.Equal(t, 0, bus.GetSubscriptionCount())

	metrics := bus.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.PublishedEvents)
	assert.Equal(t, int64(0), metrics.ProcessedEvents)
}

func TestEventBus_StartStop(t *testing.T) {
	// Arrange
	bus := NewInMemoryEventBus()

	// Initially not running
	assert.False(t, bus.IsRunning())

	// Act - Start
	err := bus.Start(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.True(t, bus.IsRunning())

	// Act - Start again (should error)
	err = bus.Start(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Act - Stop
	err = bus.Stop(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.False(t, bus.IsRunning())

	// Act - Stop again (should error)
	err = bus.Stop(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestEventBus_Subscribe(t *testing.T) {
	// Arrange
	bus := NewInMemoryEventBus()
	handler := NewTestEventHandler("TestHandler", []string{"TestEvent"})

	// Act
	subID, err := bus.Subscribe("TestEvent", handler)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, subID)
	assert.Equal(t, 1, bus.GetSubscriptionCount())
	assert.Equal(t, 1, bus.GetEventTypeSubscriptions("TestEvent"))
}

func TestEventBus_Subscribe_EmptyEventType(t *testing.T) {
	// Arrange
	bus := NewInMemoryEventBus()
	handler := NewTestEventHandler("TestHandler", []string{"TestEvent"})

	// Act
	subID, err := bus.Subscribe("", handler)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, subID)
	assert.Contains(t, err.Error(), "event type cannot be empty")
}

func TestEventBus_Subscribe_NilHandler(t *testing.T) {
	// Arrange
	bus := NewInMemoryEventBus()

	// Act
	subID, err := bus.Subscribe("TestEvent", nil)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, subID)
	assert.Contains(t, err.Error(), "handler cannot be nil")
}

func TestEventBus_SubscribeAll(t *testing.T) {
	// Arrange
	bus := NewInMemoryEventBus()
	handler := NewTestEventHandler("AllEventsHandler", []string{"TestEvent1", "TestEvent2"})

	// Act
	subID, err := bus.SubscribeAll(handler)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, subID)
	assert.Equal(t, 1, bus.GetSubscriptionCount())
}

func TestEventBus_Publish_Success(t *testing.T) {
	// Arrange
	bus := NewInMemoryEventBus()
	handler := NewTestEventHandler("TestHandler", []string{"TestEvent"})
	event := NewBaseEventMessage("TestEvent", "test-id", "TestAggregate", 1, "test data")

	bus.Subscribe("TestEvent", handler)

	// Act
	err := bus.Publish(context.Background(), event)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, handler.GetHandledEventCount())
	assert.Equal(t, event, handler.GetLastHandledEvent())

	metrics := bus.GetMetrics()
	assert.Equal(t, int64(1), metrics.PublishedEvents)
	assert.Equal(t, int64(1), metrics.ProcessedEvents)
}

func TestEventBus_Publish_NilEvent(t *testing.T) {
	// Arrange
	bus := NewInMemoryEventBus()

	// Act
	err := bus.Publish(context.Background(), nil)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event cannot be nil")
}

func TestEventBus_Publish_NoHandlers(t *testing.T) {
	// Arrange
	bus := NewInMemoryEventBus()
	event := NewBaseEventMessage("TestEvent", "test-id", "TestAggregate", 1, "test data")

	// Act
	err := bus.Publish(context.Background(), event)

	// Assert
	assert.NoError(t, err) // Should not error even if no handlers

	metrics := bus.GetMetrics()
	assert.Equal(t, int64(1), metrics.PublishedEvents)
	assert.Equal(t, int64(1), metrics.ProcessedEvents)
}

func TestEventBus_Publish_HandlerError(t *testing.T) {
	// Arrange
	bus := NewInMemoryEventBus()
	handler := NewTestEventHandler("TestHandler", []string{"TestEvent"})
	event := NewBaseEventMessage("TestEvent", "test-id", "TestAggregate", 1, "test data")

	// Set up handler to return error
	handler.HandleFunc = func(ctx context.Context, event EventMessage) error {
		return NewCQRSError(ErrCodeEventValidation.String(), "handler error", nil)
	}

	bus.Subscribe("TestEvent", handler)

	// Act
	err := bus.Publish(context.Background(), event)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler error")

	metrics := bus.GetMetrics()
	assert.Equal(t, int64(1), metrics.PublishedEvents)
	assert.Equal(t, int64(1), metrics.FailedEvents)
}

func TestEventBus_Publish_AllHandlers(t *testing.T) {
	// Arrange
	bus := NewInMemoryEventBus()
	specificHandler := NewTestEventHandler("SpecificHandler", []string{"TestEvent"})
	allHandler := NewTestEventHandler("AllHandler", []string{"TestEvent", "OtherEvent"})
	event := NewBaseEventMessage("TestEvent", "test-id", "TestAggregate", 1, "test data")

	bus.Subscribe("TestEvent", specificHandler)
	bus.SubscribeAll(allHandler)

	// Act
	err := bus.Publish(context.Background(), event)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, specificHandler.GetHandledEventCount())
	assert.Equal(t, 1, allHandler.GetHandledEventCount())
}

func TestEventBus_PublishBatch(t *testing.T) {
	// Arrange
	bus := NewInMemoryEventBus()
	handler := NewTestEventHandler("TestHandler", []string{"TestEvent"})
	events := []EventMessage{
		NewBaseEventMessage("TestEvent", "test-id-1", "TestAggregate", 1, "data 1"),
		NewBaseEventMessage("TestEvent", "test-id-2", "TestAggregate", 2, "data 2"),
		NewBaseEventMessage("TestEvent", "test-id-3", "TestAggregate", 3, "data 3"),
	}

	bus.Subscribe("TestEvent", handler)

	// Act
	err := bus.PublishBatch(context.Background(), events)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 3, handler.GetHandledEventCount())

	metrics := bus.GetMetrics()
	assert.Equal(t, int64(3), metrics.PublishedEvents)
	assert.Equal(t, int64(3), metrics.ProcessedEvents)
}

func TestEventBus_PublishBatch_EmptyEvents(t *testing.T) {
	// Arrange
	bus := NewInMemoryEventBus()

	// Act
	err := bus.PublishBatch(context.Background(), []EventMessage{})

	// Assert
	assert.NoError(t, err)
}

func TestEventBus_Publish_Async(t *testing.T) {
	// Arrange
	bus := NewInMemoryEventBus()
	handler := NewTestEventHandler("TestHandler", []string{"TestEvent"})
	event := NewBaseEventMessage("TestEvent", "test-id", "TestAggregate", 1, "test data")

	bus.Subscribe("TestEvent", handler)

	// Act
	err := bus.Publish(context.Background(), event, EventPublishOptions{Async: true})

	// Assert
	assert.NoError(t, err)

	// Wait a bit for async processing
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, handler.GetHandledEventCount())
}

func TestEventBus_Clear(t *testing.T) {
	// Arrange
	bus := NewInMemoryEventBus()
	handler := NewTestEventHandler("TestHandler", []string{"TestEvent"})
	event := NewBaseEventMessage("TestEvent", "test-id", "TestAggregate", 1, "test data")

	bus.Subscribe("TestEvent", handler)
	bus.Publish(context.Background(), event)

	// Verify initial state
	assert.Equal(t, 1, bus.GetSubscriptionCount())
	metrics := bus.GetMetrics()
	assert.Equal(t, int64(1), metrics.PublishedEvents)

	// Act
	bus.Clear()

	// Assert
	assert.Equal(t, 0, bus.GetSubscriptionCount())
	metrics = bus.GetMetrics()
	assert.Equal(t, int64(0), metrics.PublishedEvents)
	assert.Equal(t, int64(0), metrics.ProcessedEvents)
}

func TestBaseEventHandler_GetHandlerType(t *testing.T) {
	// Arrange
	handler := NewBaseEventHandler("TestHandler", SagaHandler, []string{"TestEvent"})

	// Act & Assert
	assert.Equal(t, SagaHandler, handler.GetHandlerType())
}

func TestHandlerType_String(t *testing.T) {
	tests := []struct {
		handlerType HandlerType
		expected    string
	}{
		{ProjectionHandler, "projection"},
		{ProcessManagerHandler, "process_manager"},
		{SagaHandler, "saga"},
		{NotificationHandler, "notification"},
		{HandlerType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.handlerType.String())
		})
	}
}

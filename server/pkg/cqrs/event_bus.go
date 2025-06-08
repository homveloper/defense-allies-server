package cqrs

import (
	"context"
	"time"
)

// SubscriptionID represents a subscription identifier
type SubscriptionID string

// HandlerType represents different types of event handlers
type HandlerType int

const (
	ProjectionHandler HandlerType = iota
	ProcessManagerHandler
	SagaHandler
	NotificationHandler
)

func (ht HandlerType) String() string {
	switch ht {
	case ProjectionHandler:
		return "projection"
	case ProcessManagerHandler:
		return "process_manager"
	case SagaHandler:
		return "saga"
	case NotificationHandler:
		return "notification"
	default:
		return "unknown"
	}
}

// BackoffType represents retry backoff strategies
type BackoffType int

const (
	FixedBackoff BackoffType = iota
	ExponentialBackoff
	LinearBackoff
)

func (bt BackoffType) String() string {
	switch bt {
	case FixedBackoff:
		return "fixed"
	case ExponentialBackoff:
		return "exponential"
	case LinearBackoff:
		return "linear"
	default:
		return "fixed"
	}
}

// RetryPolicy defines retry behavior for event processing
type RetryPolicy struct {
	MaxAttempts int
	Delay       time.Duration
	BackoffType BackoffType
}

// EventPublishOptions defines options for event publishing
type EventPublishOptions struct {
	Persistent   bool          // Store event permanently for event sourcing
	Immediate    bool          // Publish immediately
	Async        bool          // Process asynchronously
	Retry        *RetryPolicy  // Retry policy
	Timeout      time.Duration // Timeout
	Priority     EventPriority // Priority
	PartitionKey string        // Partition key for ordering
}

// EventBusMetrics represents event bus performance metrics
type EventBusMetrics struct {
	PublishedEvents   int64
	ProcessedEvents   int64
	FailedEvents      int64
	ActiveSubscribers int
	AverageLatency    time.Duration
	LastEventTime     time.Time
}

// StreamPosition represents position in an event stream
type StreamPosition struct {
	Offset    int64
	Timestamp time.Time
}

// EventBus interface for event publishing and subscription
type EventBus interface {
	// Event publishing
	Publish(ctx context.Context, event EventMessage, options ...EventPublishOptions) error
	PublishBatch(ctx context.Context, events []EventMessage, options ...EventPublishOptions) error

	// Subscription management
	Subscribe(eventType string, handler EventHandler) (SubscriptionID, error)
	SubscribeAll(handler EventHandler) (SubscriptionID, error)
	Unsubscribe(subscriptionID SubscriptionID) error

	// Lifecycle
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// Status
	IsRunning() bool
	GetMetrics() *EventBusMetrics
}

// EventHandler interface for handling events (go.cqrs extended)
type EventHandler interface {
	Handle(ctx context.Context, event EventMessage) error
	CanHandle(eventType string) bool
	GetHandlerName() string
	GetHandlerType() HandlerType
}

// EventStream interface for streaming events
type EventStream interface {
	Subscribe(ctx context.Context, fromPosition StreamPosition) (<-chan EventMessage, error)
	GetPosition() StreamPosition
	Close() error
}

// BaseEventHandler provides a base implementation of EventHandler
type BaseEventHandler struct {
	name        string
	handlerType HandlerType
	eventTypes  map[string]bool
}

// NewBaseEventHandler creates a new BaseEventHandler
func NewBaseEventHandler(name string, handlerType HandlerType, eventTypes []string) *BaseEventHandler {
	typeMap := make(map[string]bool)
	for _, eventType := range eventTypes {
		typeMap[eventType] = true
	}

	return &BaseEventHandler{
		name:        name,
		handlerType: handlerType,
		eventTypes:  typeMap,
	}
}

// EventHandler interface implementation

func (h *BaseEventHandler) GetHandlerName() string {
	return h.name
}

func (h *BaseEventHandler) GetHandlerType() HandlerType {
	return h.handlerType
}

func (h *BaseEventHandler) CanHandle(eventType string) bool {
	return h.eventTypes[eventType]
}

// Handle method should be implemented by concrete handlers
func (h *BaseEventHandler) Handle(ctx context.Context, event EventMessage) error {
	// Base implementation - should be overridden
	return nil
}

// Helper methods

// AddEventType adds an event type that this handler can process
func (h *BaseEventHandler) AddEventType(eventType string) {
	h.eventTypes[eventType] = true
}

// RemoveEventType removes an event type from this handler
func (h *BaseEventHandler) RemoveEventType(eventType string) {
	delete(h.eventTypes, eventType)
}

// GetSupportedEventTypes returns all supported event types
func (h *BaseEventHandler) GetSupportedEventTypes() []string {
	types := make([]string, 0, len(h.eventTypes))
	for eventType := range h.eventTypes {
		types = append(types, eventType)
	}
	return types
}

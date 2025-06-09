package cqrs

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// InMemoryEventBus provides an in-memory implementation of EventBus
type InMemoryEventBus struct {
	subscriptions map[string][]EventHandler
	allHandlers   []EventHandler
	metrics       *EventBusMetrics
	running       bool
	mutex         sync.RWMutex
	nextSubID     int64
	subIDMutex    sync.Mutex
}

// NewInMemoryEventBus creates a new in-memory event bus
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		subscriptions: make(map[string][]EventHandler),
		allHandlers:   make([]EventHandler, 0),
		metrics: &EventBusMetrics{
			PublishedEvents:   0,
			ProcessedEvents:   0,
			FailedEvents:      0,
			ActiveSubscribers: 0,
			AverageLatency:    0,
			LastEventTime:     time.Time{},
		},
		running: false,
	}
}

// EventBus interface implementation

func (bus *InMemoryEventBus) Publish(ctx context.Context, event EventMessage, options ...EventPublishOptions) error {
	if event == nil {
		return NewCQRSError(ErrCodeEventValidation.String(), "event cannot be nil", nil)
	}

	start := time.Now()

	bus.mutex.Lock()
	bus.metrics.PublishedEvents++
	bus.metrics.LastEventTime = start
	bus.mutex.Unlock()

	// Get options
	opts := EventPublishOptions{
		Persistent: false,
		Immediate:  true,
		Async:      false,
	}
	if len(options) > 0 {
		opts = options[0]
	}

	// Process event
	if opts.Async {
		go bus.processEvent(ctx, event)
	} else {
		if err := bus.processEvent(ctx, event); err != nil {
			bus.mutex.Lock()
			bus.metrics.FailedEvents++
			bus.mutex.Unlock()
			return err
		}
	}

	// Update metrics
	bus.mutex.Lock()
	latency := time.Since(start)
	if bus.metrics.ProcessedEvents == 0 {
		bus.metrics.AverageLatency = latency
	} else {
		bus.metrics.AverageLatency = (bus.metrics.AverageLatency + latency) / 2
	}
	bus.metrics.ProcessedEvents++
	bus.mutex.Unlock()

	return nil
}

func (bus *InMemoryEventBus) PublishBatch(ctx context.Context, events []EventMessage, options ...EventPublishOptions) error {
	if len(events) == 0 {
		return nil
	}

	for _, event := range events {
		if err := bus.Publish(ctx, event, options...); err != nil {
			return err
		}
	}

	return nil
}

func (bus *InMemoryEventBus) Subscribe(eventType string, handler EventHandler) (SubscriptionID, error) {
	if eventType == "" {
		return "", NewCQRSError(ErrCodeEventValidation.String(), "event type cannot be empty", nil)
	}
	if handler == nil {
		return "", NewCQRSError(ErrCodeEventValidation.String(), "handler cannot be nil", nil)
	}

	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	if _, exists := bus.subscriptions[eventType]; !exists {
		bus.subscriptions[eventType] = make([]EventHandler, 0)
	}

	bus.subscriptions[eventType] = append(bus.subscriptions[eventType], handler)
	bus.metrics.ActiveSubscribers++

	return bus.generateSubscriptionID(), nil
}

func (bus *InMemoryEventBus) SubscribeAll(handler EventHandler) (SubscriptionID, error) {
	if handler == nil {
		return "", NewCQRSError(ErrCodeEventValidation.String(), "handler cannot be nil", nil)
	}

	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	bus.allHandlers = append(bus.allHandlers, handler)
	bus.metrics.ActiveSubscribers++

	return bus.generateSubscriptionID(), nil
}

func (bus *InMemoryEventBus) Unsubscribe(subscriptionID SubscriptionID) error {
	// Note: In a real implementation, you would track subscription IDs
	// For simplicity, this implementation doesn't track individual subscriptions
	return nil
}

func (bus *InMemoryEventBus) Start(ctx context.Context) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	if bus.running {
		return NewCQRSError(ErrCodeEventBusError.String(), "event bus is already running", nil)
	}

	bus.running = true
	return nil
}

func (bus *InMemoryEventBus) Stop(ctx context.Context) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	if !bus.running {
		return NewCQRSError(ErrCodeEventBusError.String(), "event bus is not running", nil)
	}

	bus.running = false
	return nil
}

func (bus *InMemoryEventBus) IsRunning() bool {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()

	return bus.running
}

func (bus *InMemoryEventBus) GetMetrics() *EventBusMetrics {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()

	// Return a copy of metrics
	return &EventBusMetrics{
		PublishedEvents:   bus.metrics.PublishedEvents,
		ProcessedEvents:   bus.metrics.ProcessedEvents,
		FailedEvents:      bus.metrics.FailedEvents,
		ActiveSubscribers: bus.metrics.ActiveSubscribers,
		AverageLatency:    bus.metrics.AverageLatency,
		LastEventTime:     bus.metrics.LastEventTime,
	}
}

// Helper methods

func (bus *InMemoryEventBus) processEvent(ctx context.Context, event EventMessage) error {
	bus.mutex.RLock()

	// Get handlers for specific event type
	handlers := make([]EventHandler, 0)
	if eventHandlers, exists := bus.subscriptions[event.EventType()]; exists {
		handlers = append(handlers, eventHandlers...)
	}

	// Add all-event handlers
	handlers = append(handlers, bus.allHandlers...)

	bus.mutex.RUnlock()

	// Process handlers
	for _, handler := range handlers {
		if handler.CanHandle(event.EventType()) {
			if err := handler.Handle(ctx, event); err != nil {
				return NewCQRSError(ErrCodeEventValidation.String(),
					fmt.Sprintf("handler %s failed to process event %s", handler.GetHandlerName(), event.EventType()), err)
			}
		}
	}

	return nil
}

func (bus *InMemoryEventBus) generateSubscriptionID() SubscriptionID {
	bus.subIDMutex.Lock()
	defer bus.subIDMutex.Unlock()

	bus.nextSubID++
	return SubscriptionID(fmt.Sprintf("sub_%d", bus.nextSubID))
}

// GetSubscriptionCount returns the number of active subscriptions
func (bus *InMemoryEventBus) GetSubscriptionCount() int {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()

	count := 0
	for _, handlers := range bus.subscriptions {
		count += len(handlers)
	}
	count += len(bus.allHandlers)

	return count
}

// GetEventTypeSubscriptions returns the number of subscriptions for a specific event type
func (bus *InMemoryEventBus) GetEventTypeSubscriptions(eventType string) int {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()

	if handlers, exists := bus.subscriptions[eventType]; exists {
		return len(handlers)
	}

	return 0
}

// Clear removes all subscriptions and resets metrics
func (bus *InMemoryEventBus) Clear() {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	bus.subscriptions = make(map[string][]EventHandler)
	bus.allHandlers = make([]EventHandler, 0)
	bus.metrics = &EventBusMetrics{
		PublishedEvents:   0,
		ProcessedEvents:   0,
		FailedEvents:      0,
		ActiveSubscribers: 0,
		AverageLatency:    0,
		LastEventTime:     time.Time{},
	}
}

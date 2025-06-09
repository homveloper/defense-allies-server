package redisstream

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"cqrs"
)

// RedisStreamEventBus implements EventBus interface using Redis Streams
type RedisStreamEventBus struct {
	// Dependencies
	client     redis.UniversalClient
	config     *RedisStreamConfig
	serializer EventSerializer

	// State management
	running     int32                                 // atomic boolean
	subscribers map[cqrs.SubscriptionID]*subscription // subscription registry
	mu          sync.RWMutex                          // protects subscribers map

	// Lifecycle management
	ctx       context.Context
	cancel    context.CancelFunc
	waitGroup sync.WaitGroup

	// Metrics
	metrics   *cqrs.EventBusMetrics
	metricsMu sync.RWMutex
}

// subscription represents an active event subscription
type subscription struct {
	id           cqrs.SubscriptionID
	eventType    string
	handler      cqrs.EventHandler
	consumerName string
	streamName   string
	isActive     bool
	cancel       context.CancelFunc
}

// NewRedisStreamEventBus creates a new Redis Stream EventBus with default JSON serializer
func NewRedisStreamEventBus(client redis.UniversalClient, config *RedisStreamConfig) (*RedisStreamEventBus, error) {
	return NewRedisStreamEventBusWithSerializer(client, config, NewJSONEventSerializer())
}

// NewRedisStreamEventBusWithSerializer creates a new Redis Stream EventBus with custom serializer
func NewRedisStreamEventBusWithSerializer(client redis.UniversalClient, config *RedisStreamConfig, serializer EventSerializer) (*RedisStreamEventBus, error) {
	if client == nil {
		return nil, ErrConfigInvalid("redis client cannot be nil")
	}

	if config == nil {
		return nil, ErrConfigInvalid("config cannot be nil")
	}

	if serializer == nil {
		return nil, ErrConfigInvalid("serializer cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &RedisStreamEventBus{
		client:      client,
		config:      config,
		serializer:  serializer,
		subscribers: make(map[cqrs.SubscriptionID]*subscription),
		metrics:     &cqrs.EventBusMetrics{},
	}, nil
}

// EventBus interface implementation

// Publish publishes a single event
func (bus *RedisStreamEventBus) Publish(ctx context.Context, event cqrs.EventMessage, options ...cqrs.EventPublishOptions) error {
	if !bus.IsRunning() {
		return ErrEventBusNotRunning
	}

	streamName := bus.getStreamName(event, options...)

	// Convert event to Redis stream format
	fields, err := bus.eventToFields(event)
	if err != nil {
		return ErrStreamOperation("convert_event", err)
	}

	// Publish to Redis stream
	result := bus.client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		MaxLen: bus.config.Stream.MaxLen,
		Approx: bus.config.Stream.MaxLenApprox > 0,
		ID:     "*",
		Values: fields,
	})

	if err := result.Err(); err != nil {
		return ErrStreamOperation("xadd", err)
	}

	// Update metrics
	bus.updatePublishedMetrics(1)

	return nil
}

// PublishBatch publishes multiple events in a batch
func (bus *RedisStreamEventBus) PublishBatch(ctx context.Context, events []cqrs.EventMessage, options ...cqrs.EventPublishOptions) error {
	if !bus.IsRunning() {
		return ErrEventBusNotRunning
	}

	// Use pipeline for batch publishing
	pipe := bus.client.Pipeline()

	for _, event := range events {
		streamName := bus.getStreamName(event, options...)

		fields, err := bus.eventToFields(event)
		if err != nil {
			return ErrStreamOperation("convert_event_batch", err)
		}

		pipe.XAdd(ctx, &redis.XAddArgs{
			Stream: streamName,
			MaxLen: bus.config.Stream.MaxLen,
			Approx: bus.config.Stream.MaxLenApprox > 0,
			ID:     "*",
			Values: fields,
		})
	}

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return ErrStreamOperation("pipeline_exec", err)
	}

	// Update metrics
	bus.updatePublishedMetrics(int64(len(events)))

	return nil
}

// Subscribe subscribes to a specific event type
func (bus *RedisStreamEventBus) Subscribe(eventType string, handler cqrs.EventHandler) (cqrs.SubscriptionID, error) {
	if !bus.IsRunning() {
		return "", ErrEventBusNotRunning
	}

	subID := cqrs.SubscriptionID(uuid.New().String())
	streamName := bus.getStreamNameForEventType(eventType)
	consumerName := bus.getConsumerName(handler)

	// Create consumer group if not exists
	err := bus.ensureConsumerGroup(streamName, consumerName)
	if err != nil {
		return "", ErrHandlerRegistrationFailed
	}

	// Create subscription context
	subCtx, cancel := context.WithCancel(bus.ctx)

	sub := &subscription{
		id:           subID,
		eventType:    eventType,
		handler:      handler,
		consumerName: consumerName,
		streamName:   streamName,
		isActive:     true,
		cancel:       cancel,
	}

	// Register subscription
	bus.mu.Lock()
	bus.subscribers[subID] = sub
	bus.mu.Unlock()

	// Start consuming events
	bus.waitGroup.Add(1)
	go bus.consumeEvents(subCtx, sub)

	return subID, nil
}

// SubscribeAll subscribes to all events
func (bus *RedisStreamEventBus) SubscribeAll(handler cqrs.EventHandler) (cqrs.SubscriptionID, error) {
	// For simplicity, we'll use a wildcard approach
	return bus.Subscribe("*", handler)
}

// Unsubscribe unsubscribes from events
func (bus *RedisStreamEventBus) Unsubscribe(subscriptionID cqrs.SubscriptionID) error {
	bus.mu.Lock()
	sub, exists := bus.subscribers[subscriptionID]
	if !exists {
		bus.mu.Unlock()
		return ErrSubscriptionNotFound
	}

	delete(bus.subscribers, subscriptionID)
	bus.mu.Unlock()

	// Cancel subscription
	if sub.cancel != nil {
		sub.cancel()
	}

	return nil
}

// Start starts the event bus
func (bus *RedisStreamEventBus) Start(ctx context.Context) error {
	if bus.IsRunning() {
		return ErrEventBusAlreadyRunning
	}

	// Test Redis connection
	err := bus.client.Ping(ctx).Err()
	if err != nil {
		return ErrConnectionFailed
	}

	// Set up context
	bus.ctx, bus.cancel = context.WithCancel(ctx)

	// Mark as running
	atomic.StoreInt32(&bus.running, 1)

	// Initialize metrics
	bus.metricsMu.Lock()
	bus.metrics.LastEventTime = time.Now()
	bus.metricsMu.Unlock()

	return nil
}

// Stop stops the event bus
func (bus *RedisStreamEventBus) Stop(ctx context.Context) error {
	if !bus.IsRunning() {
		return ErrEventBusNotRunning
	}

	// Mark as not running
	atomic.StoreInt32(&bus.running, 0)

	// Cancel all operations
	if bus.cancel != nil {
		bus.cancel()
	}

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		bus.waitGroup.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All goroutines finished
	case <-time.After(bus.config.Consumer.ShutdownTimeout):
		// Timeout - force shutdown
	}

	// Clear subscribers
	bus.mu.Lock()
	bus.subscribers = make(map[cqrs.SubscriptionID]*subscription)
	bus.mu.Unlock()

	return nil
}

// IsRunning returns true if the event bus is running
func (bus *RedisStreamEventBus) IsRunning() bool {
	return atomic.LoadInt32(&bus.running) == 1
}

// GetMetrics returns current metrics
func (bus *RedisStreamEventBus) GetMetrics() *cqrs.EventBusMetrics {
	bus.metricsMu.RLock()
	defer bus.metricsMu.RUnlock()

	// Create a copy to avoid data races
	return &cqrs.EventBusMetrics{
		PublishedEvents:   bus.metrics.PublishedEvents,
		ProcessedEvents:   bus.metrics.ProcessedEvents,
		FailedEvents:      bus.metrics.FailedEvents,
		ActiveSubscribers: len(bus.subscribers),
		AverageLatency:    bus.metrics.AverageLatency,
		LastEventTime:     bus.metrics.LastEventTime,
	}
}

// SetSerializer sets the event serializer (should be called before Start)
func (bus *RedisStreamEventBus) SetSerializer(serializer EventSerializer) error {
	if bus.IsRunning() {
		return fmt.Errorf("cannot change serializer while event bus is running")
	}

	if serializer == nil {
		return fmt.Errorf("serializer cannot be nil")
	}

	bus.serializer = serializer
	return nil
}

// GetSerializer returns the current event serializer
func (bus *RedisStreamEventBus) GetSerializer() EventSerializer {
	return bus.serializer
}

// Helper methods

// getStreamName generates stream name for an event
func (bus *RedisStreamEventBus) getStreamName(event cqrs.EventMessage, options ...cqrs.EventPublishOptions) string {
	category := "domain"
	priority := "normal"
	partitionKey := event.Type()

	if len(options) > 0 {
		opt := options[0]
		if opt.PartitionKey != "" {
			partitionKey = opt.PartitionKey
		}
		priority = opt.Priority.String()
	}

	return fmt.Sprintf("%s%s%s%s%s%s%s",
		bus.config.Stream.StreamPrefix,
		bus.config.Stream.NamespaceDelim,
		category,
		bus.config.Stream.NamespaceDelim,
		priority,
		bus.config.Stream.NamespaceDelim,
		partitionKey,
	)
}

// getStreamNameForEventType generates stream name for event type subscription
func (bus *RedisStreamEventBus) getStreamNameForEventType(eventType string) string {
	if eventType == "*" {
		return fmt.Sprintf("%s%s*", bus.config.Stream.StreamPrefix, bus.config.Stream.NamespaceDelim)
	}
	return fmt.Sprintf("%s%s%s", bus.config.Stream.StreamPrefix, bus.config.Stream.NamespaceDelim, eventType)
}

// getConsumerName generates consumer name for a handler
func (bus *RedisStreamEventBus) getConsumerName(handler cqrs.EventHandler) string {
	return fmt.Sprintf("%s_%s_%s",
		bus.config.Consumer.ServiceName,
		handler.GetHandlerType().String(),
		bus.config.Consumer.InstanceID,
	)
}

// ensureConsumerGroup creates consumer group if it doesn't exist
func (bus *RedisStreamEventBus) ensureConsumerGroup(streamName, groupName string) error {
	result := bus.client.XGroupCreateMkStream(context.Background(), streamName, groupName, "0")
	err := result.Err()

	// Ignore error if group already exists
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return err
	}

	return nil
}

// eventToFields converts event to Redis stream fields using serialization
func (bus *RedisStreamEventBus) eventToFields(event cqrs.EventMessage) (map[string]interface{}, error) {
	// Serialize the entire event
	serializedData, err := bus.serializer.Serialize(event)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize event: %w", err)
	}

	// Store serialized data as a single field
	fields := map[string]interface{}{
		"event_type":      event.EventType(),
		"aggregate_id":    event.ID(),
		"aggregate_type":  event.Type(),
		"event_id":        event.EventID(),
		"version":         event.Version(),
		"timestamp":       event.Timestamp().Format(time.RFC3339Nano),
		"serialized_data": string(serializedData),
		"format":          string(bus.serializer.Format()),
	}

	// Add partition key for routing
	if domainEvent, ok := event.(cqrs.DomainEventMessage); ok {
		fields["priority"] = domainEvent.GetPriority().String()
		fields["category"] = domainEvent.GetEventCategory().String()
		fields["issuer_id"] = domainEvent.IssuerID()
		fields["causation_id"] = domainEvent.CausationID()
		fields["correlation_id"] = domainEvent.CorrelationID()
	}

	return fields, nil
}

// consumeEvents consumes events for a subscription
func (bus *RedisStreamEventBus) consumeEvents(ctx context.Context, sub *subscription) {
	defer bus.waitGroup.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Read events from stream
			result := bus.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    bus.getConsumerGroupName(),
				Consumer: sub.consumerName,
				Streams:  []string{sub.streamName, ">"},
				Count:    bus.config.Stream.Count,
				Block:    bus.config.Stream.BlockTime,
			})

			if err := result.Err(); err != nil {
				if err == redis.Nil {
					continue // No new messages
				}
				// Log error and continue
				continue
			}

			// Process messages
			for _, stream := range result.Val() {
				for _, message := range stream.Messages {
					bus.processMessage(ctx, sub, message)
				}
			}
		}
	}
}

// getConsumerGroupName returns the consumer group name
func (bus *RedisStreamEventBus) getConsumerGroupName() string {
	return fmt.Sprintf("%s_cg", bus.config.Stream.ConsumerGroupPrefix)
}

// processMessage processes a single Redis stream message
func (bus *RedisStreamEventBus) processMessage(ctx context.Context, sub *subscription, message redis.XMessage) {
	// Convert Redis message to EventMessage
	event, err := bus.fieldsToEvent(message.Values)
	if err != nil {
		bus.updateFailedMetrics(1)
		return
	}

	// Check if handler can handle this event
	if !sub.handler.CanHandle(event.EventType()) {
		// Acknowledge message even if not handled
		bus.client.XAck(ctx, sub.streamName, bus.getConsumerGroupName(), message.ID)
		return
	}

	// Handle the event
	err = sub.handler.Handle(ctx, event)
	if err != nil {
		bus.updateFailedMetrics(1)
		// In a real implementation, you'd implement retry logic here
		return
	}

	// Acknowledge successful processing
	bus.client.XAck(ctx, sub.streamName, bus.getConsumerGroupName(), message.ID)
	bus.updateProcessedMetrics(1)
}

// fieldsToEvent converts Redis stream fields to EventMessage using deserialization
func (bus *RedisStreamEventBus) fieldsToEvent(fields map[string]interface{}) (cqrs.EventMessage, error) {
	// Get serialized data from fields
	serializedDataStr, ok := fields["serialized_data"].(string)
	if !ok {
		return nil, fmt.Errorf("serialized_data field not found or invalid")
	}

	// Get format (for future compatibility with multiple formats)
	format, ok := fields["format"].(string)
	if !ok {
		// Default to JSON for backward compatibility
		format = string(SerializationFormatJSON)
	}

	// Verify format matches current serializer
	if SerializationFormat(format) != bus.serializer.Format() {
		return nil, fmt.Errorf("unsupported serialization format: %s", format)
	}

	// Deserialize the event
	event, err := bus.serializer.Deserialize([]byte(serializedDataStr))
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize event: %w", err)
	}

	return event, nil
}

// updatePublishedMetrics updates published event metrics
func (bus *RedisStreamEventBus) updatePublishedMetrics(count int64) {
	bus.metricsMu.Lock()
	defer bus.metricsMu.Unlock()

	bus.metrics.PublishedEvents += count
	bus.metrics.LastEventTime = time.Now()
}

// updateProcessedMetrics updates processed event metrics
func (bus *RedisStreamEventBus) updateProcessedMetrics(count int64) {
	bus.metricsMu.Lock()
	defer bus.metricsMu.Unlock()

	bus.metrics.ProcessedEvents += count
}

// updateFailedMetrics updates failed event metrics
func (bus *RedisStreamEventBus) updateFailedMetrics(count int64) {
	bus.metricsMu.Lock()
	defer bus.metricsMu.Unlock()

	bus.metrics.FailedEvents += count
}

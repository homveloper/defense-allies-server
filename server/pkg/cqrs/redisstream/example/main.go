package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"cqrs"
	"cqrs/redisstream"
)

// ExampleEventHandler implements a simple event handler for demonstration
type ExampleEventHandler struct {
	*cqrs.BaseEventHandler
}

func NewExampleEventHandler() *ExampleEventHandler {
	base := cqrs.NewBaseEventHandler(
		"example-handler",
		cqrs.ProjectionHandler,
		[]string{"UserRegistered", "UserUpdated"},
	)

	return &ExampleEventHandler{
		BaseEventHandler: base,
	}
}

func (h *ExampleEventHandler) Handle(ctx context.Context, event cqrs.EventMessage) error {
	fmt.Printf("Handling event: %s (ID: %s, Aggregate: %s)\n",
		event.EventType(),
		event.EventID(),
		event.ID(),
	)

	// Simulate some processing time
	time.Sleep(100 * time.Millisecond)

	return nil
}

func main() {
	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Test Redis connection
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Create Redis Stream EventBus configuration
	config := redisstream.DefaultRedisStreamConfig()
	config.Consumer.ServiceName = "example-service"
	config.Consumer.InstanceID = "example-instance-1"

	// Create EventBus
	eventBus, err := redisstream.NewRedisStreamEventBus(rdb, config)
	if err != nil {
		log.Fatalf("Failed to create EventBus: %v", err)
	}

	// Start EventBus
	err = eventBus.Start(ctx)
	if err != nil {
		log.Fatalf("Failed to start EventBus: %v", err)
	}
	defer eventBus.Stop(ctx)

	fmt.Println("EventBus started successfully!")

	// Create and register event handler
	handler := NewExampleEventHandler()
	subID, err := eventBus.Subscribe("UserRegistered", handler)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	defer eventBus.Unsubscribe(subID)

	fmt.Printf("Subscribed with ID: %s\n", subID)

	// Give some time for subscription to be ready
	time.Sleep(1 * time.Second)

	// Publish some test events
	for i := 0; i < 3; i++ {
		baseOptions := cqrs.Options().
			WithAggregateID(fmt.Sprintf("user-%d", i+1)).
			WithAggregateType("User").
			WithVersion(1)

		event := cqrs.NewBaseDomainEventMessage(
			"UserRegistered",
			map[string]interface{}{
				"name":  fmt.Sprintf("User %d", i+1),
				"email": fmt.Sprintf("user%d@example.com", i+1),
			},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		err = eventBus.Publish(ctx, event)
		if err != nil {
			log.Printf("Failed to publish event %d: %v", i+1, err)
		} else {
			fmt.Printf("Published event %d: UserRegistered for user-%d\n", i+1, i+1)
		}

		// Small delay between events
		time.Sleep(500 * time.Millisecond)
	}

	// Wait a bit for event processing
	time.Sleep(3 * time.Second)

	// Print metrics
	metrics := eventBus.GetMetrics()
	fmt.Printf("\nEventBus Metrics:\n")
	fmt.Printf("- Published Events: %d\n", metrics.PublishedEvents)
	fmt.Printf("- Processed Events: %d\n", metrics.ProcessedEvents)
	fmt.Printf("- Failed Events: %d\n", metrics.FailedEvents)
	fmt.Printf("- Active Subscribers: %d\n", metrics.ActiveSubscribers)
	fmt.Printf("- Last Event Time: %s\n", metrics.LastEventTime.Format(time.RFC3339))

	fmt.Println("\nExample completed successfully!")
}

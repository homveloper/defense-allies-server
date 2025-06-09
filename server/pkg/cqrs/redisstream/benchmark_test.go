package redisstream

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"cqrs"
)

// BenchmarkEventBusPublishing benchmarks event publishing performance
func BenchmarkEventBusPublishing(b *testing.B) {
	ctx := context.Background()
	testContainer, err := NewTestRedisContainer(ctx)
	if err != nil {
		b.Fatalf("Failed to create test container: %v", err)
	}
	defer testContainer.Close(ctx)

	eventBus, err := NewRedisStreamEventBus(testContainer.client, testContainer.config)
	if err != nil {
		b.Fatalf("Failed to create EventBus: %v", err)
	}

	err = eventBus.Start(ctx)
	if err != nil {
		b.Fatalf("Failed to start EventBus: %v", err)
	}
	defer eventBus.Stop(ctx)

	// Pre-create events to avoid allocation overhead in benchmark
	events := make([]cqrs.EventMessage, b.N)
	for i := 0; i < b.N; i++ {
		baseOptions := cqrs.Options().
			WithAggregateID(fmt.Sprintf("bench-test-%d", i)).
			WithAggregateType("BenchmarkTest").
			WithVersion(1)

		events[i] = cqrs.NewBaseDomainEventMessage(
			"BenchmarkTest",
			map[string]interface{}{
				"index": i,
				"data":  fmt.Sprintf("Benchmark data %d", i),
			},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := eventBus.Publish(ctx, events[i])
		if err != nil {
			b.Fatalf("Failed to publish event: %v", err)
		}
	}
}

// BenchmarkEventBusBatchPublishing benchmarks batch publishing performance
func BenchmarkEventBusBatchPublishing(b *testing.B) {
	ctx := context.Background()
	testContainer, err := NewTestRedisContainer(ctx)
	if err != nil {
		b.Fatalf("Failed to create test container: %v", err)
	}
	defer testContainer.Close(ctx)

	eventBus, err := NewRedisStreamEventBus(testContainer.client, testContainer.config)
	if err != nil {
		b.Fatalf("Failed to create EventBus: %v", err)
	}

	err = eventBus.Start(ctx)
	if err != nil {
		b.Fatalf("Failed to start EventBus: %v", err)
	}
	defer eventBus.Stop(ctx)

	batchSizes := []int{1, 10, 50, 100}

	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("BatchSize%d", batchSize), func(b *testing.B) {
			// Pre-create event batches
			batches := make([][]cqrs.EventMessage, 0, b.N)
			for i := 0; i < b.N; i++ {
				batch := make([]cqrs.EventMessage, batchSize)
				for j := 0; j < batchSize; j++ {
					baseOptions := cqrs.Options().
						WithAggregateID(fmt.Sprintf("batch-test-%d-%d", i, j)).
						WithAggregateType("BatchTest")

					batch[j] = cqrs.NewBaseDomainEventMessage(
						"BatchTest",
						map[string]interface{}{
							"batch": i,
							"index": j,
						},
						[]*cqrs.BaseEventMessageOptions{baseOptions},
					)
				}
				batches = append(batches, batch)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				err := eventBus.PublishBatch(ctx, batches[i])
				if err != nil {
					b.Fatalf("Failed to publish batch: %v", err)
				}
			}
		})
	}
}

// BenchmarkSerialization benchmarks serialization performance
func BenchmarkSerialization(b *testing.B) {
	serializers := map[string]EventSerializer{
		"JSON": NewJSONEventSerializer(),
	}

	// Create test event with complex data
	complexData := map[string]interface{}{
		"user": map[string]interface{}{
			"id":   "user-123",
			"name": "Benchmark User",
			"profile": map[string]interface{}{
				"age":         30,
				"preferences": []string{"tech", "gaming", "music", "sports", "travel"},
				"metadata": map[string]interface{}{
					"last_login":   time.Now().Format(time.RFC3339),
					"login_count":  1000,
					"premium_user": true,
					"balance":      12345.67,
					"settings": map[string]interface{}{
						"theme":         "dark",
						"notifications": true,
						"language":      "en",
					},
				},
			},
		},
		"action": map[string]interface{}{
			"type":       "profile_update",
			"timestamp":  time.Now().Unix(),
			"changes":    []string{"email", "preferences", "settings"},
			"ip_address": "192.168.1.100",
			"user_agent": "Mozilla/5.0 (compatible benchmark)",
		},
	}

	baseOptions := cqrs.Options().
		WithAggregateID("benchmark-user-123").
		WithAggregateType("BenchmarkUser").
		WithVersion(1)

	event := cqrs.NewBaseDomainEventMessage(
		"BenchmarkEvent",
		complexData,
		[]*cqrs.BaseEventMessageOptions{baseOptions},
	)

	for serializerName, serializer := range serializers {
		b.Run(fmt.Sprintf("Serialize_%s", serializerName), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := serializer.Serialize(event)
				if err != nil {
					b.Fatalf("Serialization failed: %v", err)
				}
			}
		})

		b.Run(fmt.Sprintf("Deserialize_%s", serializerName), func(b *testing.B) {
			// Pre-serialize the event
			data, err := serializer.Serialize(event)
			if err != nil {
				b.Fatalf("Pre-serialization failed: %v", err)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := serializer.Deserialize(data)
				if err != nil {
					b.Fatalf("Deserialization failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkPriorityStreamManager benchmarks priority stream operations
func BenchmarkPriorityStreamManager(b *testing.B) {
	config := DefaultRedisStreamConfig()
	config.Stream.EnablePriorityStreams = true

	manager, err := NewPriorityStreamManager(config)
	if err != nil {
		b.Fatalf("Failed to create priority manager: %v", err)
	}

	priorities := []cqrs.EventPriority{cqrs.PriorityCritical, cqrs.PriorityHigh, cqrs.PriorityNormal, cqrs.PriorityLow}
	categories := []cqrs.EventCategory{cqrs.DomainEvent, cqrs.SystemEvent, cqrs.UserAction}
	partitionKeys := []string{"user", "order", "product", "payment", "notification"}

	b.Run("GetStreamName", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			priority := priorities[i%len(priorities)]
			category := categories[i%len(categories)]
			partitionKey := partitionKeys[i%len(partitionKeys)]

			_ = manager.GetStreamName(priority, category, partitionKey)
		}
	})

	b.Run("GetStreamsByPriority", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			category := categories[i%len(categories)]
			partitionKey := partitionKeys[i%len(partitionKeys)]

			_ = manager.GetStreamsByPriority(category, partitionKey)
		}
	})

	b.Run("GetConsumerConfigurations", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			category := categories[i%len(categories)]
			partitionKey := partitionKeys[i%len(partitionKeys)]
			handlerType := cqrs.ProjectionHandler

			_ = manager.GetConsumerConfigurations(category, partitionKey, handlerType)
		}
	})
}

// BenchmarkCircuitBreaker benchmarks circuit breaker operations
func BenchmarkCircuitBreaker(b *testing.B) {
	config := DefaultRedisStreamConfig()
	config.Monitoring.CircuitBreakerEnabled = true
	config.Monitoring.FailureThreshold = 5

	cb, err := NewCircuitBreaker("benchmark-service", config)
	if err != nil {
		b.Fatalf("Failed to create circuit breaker: %v", err)
	}

	b.Run("SuccessfulCall", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := cb.Call(func() error {
				return nil // Always succeed
			})
			if err != nil {
				b.Fatalf("Unexpected error: %v", err)
			}
		}
	})

	b.Run("RecordSuccess", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cb.RecordSuccess()
		}
	})

	b.Run("RecordFailure", func(b *testing.B) {
		// Reset circuit breaker state before each run
		cb.Reset()

		for i := 0; i < b.N; i++ {
			if i%10 == 0 {
				cb.Reset() // Reset every 10 operations to prevent circuit opening
			}
			cb.RecordFailure(fmt.Sprintf("benchmark error %d", i))
		}
	})

	b.Run("GetMetrics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = cb.GetMetrics()
		}
	})
}

// BenchmarkRetryPolicyManager benchmarks retry policy operations
func BenchmarkRetryPolicyManager(b *testing.B) {
	config := DefaultRedisStreamConfig()
	manager, err := NewRetryPolicyManager(config)
	if err != nil {
		b.Fatalf("Failed to create retry manager: %v", err)
	}

	// Pre-create test events
	events := make([]cqrs.EventMessage, 1000)
	for i := range events {
		baseOptions := cqrs.Options().
			WithAggregateID(fmt.Sprintf("retry-test-%d", i)).
			WithAggregateType("RetryTest")

		events[i] = cqrs.NewBaseDomainEventMessage(
			"RetryTest",
			map[string]interface{}{"index": i},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)
	}

	error := &ProcessingError{
		Error:     "benchmark error",
		Handler:   "BenchmarkHandler",
		Timestamp: time.Now(),
	}

	b.Run("ShouldRetry", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			event := events[i%len(events)]
			_ = manager.ShouldRetry(event, error)
		}
	})

	b.Run("CalculateDelay", func(b *testing.B) {
		policy := manager.GetDefaultRetryPolicy()

		for i := 0; i < b.N; i++ {
			attempt := (i % 10) + 1 // 1-10 attempts
			_ = manager.CalculateDelay(policy, attempt)
		}
	})

	b.Run("EnrichEventForRetry", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			event := events[i%len(events)]
			_ = manager.EnrichEventForRetry(event, error)
		}
	})
}

// BenchmarkDLQManager benchmarks DLQ operations
func BenchmarkDLQManager(b *testing.B) {
	config := DefaultRedisStreamConfig()
	config.Stream.DLQEnabled = true

	manager, err := NewDLQManager(config)
	if err != nil {
		b.Fatalf("Failed to create DLQ manager: %v", err)
	}

	streamNames := []string{
		"events:domain:normal:user",
		"events:domain:high:order",
		"events:system:critical:payment",
	}

	error := &ProcessingError{
		Error:     "benchmark DLQ error",
		Handler:   "BenchmarkHandler",
		Timestamp: time.Now(),
	}

	// Pre-create test events
	events := make([]cqrs.EventMessage, 1000)
	for i := range events {
		baseOptions := cqrs.Options().
			WithAggregateID(fmt.Sprintf("dlq-test-%d", i)).
			WithAggregateType("DLQTest")

		events[i] = cqrs.NewBaseDomainEventMessage(
			"DLQTest",
			map[string]interface{}{"index": i},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)
	}

	b.Run("GetDLQStreamName", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			streamName := streamNames[i%len(streamNames)]
			_ = manager.GetDLQStreamName(streamName)
		}
	})

	b.Run("ShouldMoveToDLQ", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			event := events[i%len(events)]
			_ = manager.ShouldMoveToDLQ(event)
		}
	})

	b.Run("EnrichEventForDLQ", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			event := events[i%len(events)]
			_ = manager.EnrichEventForDLQ(event, error)
		}
	})

	b.Run("RecordDLQEvent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			streamName := streamNames[i%len(streamNames)]
			manager.RecordDLQEvent(streamName, "BenchmarkHandler", "benchmark_error")
		}
	})
}

// BenchmarkHealthChecker benchmarks health check operations
func BenchmarkHealthChecker(b *testing.B) {
	ctx := context.Background()
	config := DefaultRedisStreamConfig()

	hc, err := NewHealthChecker("benchmark-service", config)
	if err != nil {
		b.Fatalf("Failed to create health checker: %v", err)
	}

	// Add some health checks
	hc.AddCheck("fast", NewCustomHealthCheck("fast", func(ctx context.Context) HealthCheckResult {
		return HealthCheckResult{
			Status:  HealthStatusHealthy,
			Message: "Fast check",
		}
	}))

	hc.AddCheck("medium", NewCustomHealthCheck("medium", func(ctx context.Context) HealthCheckResult {
		time.Sleep(1 * time.Millisecond) // Simulate some work
		return HealthCheckResult{
			Status:  HealthStatusHealthy,
			Message: "Medium check",
		}
	}))

	b.Run("CheckHealth", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = hc.CheckHealth(ctx)
		}
	})

	b.Run("AddRemoveCheck", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			checkName := fmt.Sprintf("temp-check-%d", i)
			check := NewCustomHealthCheck(checkName, func(ctx context.Context) HealthCheckResult {
				return HealthCheckResult{Status: HealthStatusHealthy}
			})

			hc.AddCheck(checkName, check)
			hc.RemoveCheck(checkName)
		}
	})
}

// BenchmarkConcurrentOperations benchmarks concurrent operations
func BenchmarkConcurrentOperations(b *testing.B) {
	ctx := context.Background()
	testContainer, err := NewTestRedisContainer(ctx)
	if err != nil {
		b.Fatalf("Failed to create test container: %v", err)
	}
	defer testContainer.Close(ctx)

	config := testContainer.config
	config.Stream.EnablePriorityStreams = true

	eventBus, err := NewRedisStreamEventBus(testContainer.client, config)
	if err != nil {
		b.Fatalf("Failed to create EventBus: %v", err)
	}

	err = eventBus.Start(ctx)
	if err != nil {
		b.Fatalf("Failed to start EventBus: %v", err)
	}
	defer eventBus.Stop(ctx)

	b.Run("ConcurrentPublish", func(b *testing.B) {
		var wg sync.WaitGroup
		numGoroutines := 10
		eventsPerGoroutine := b.N / numGoroutines

		b.ResetTimer()

		for g := 0; g < numGoroutines; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for i := 0; i < eventsPerGoroutine; i++ {
					baseOptions := cqrs.Options().
						WithAggregateID(fmt.Sprintf("concurrent-test-%d-%d", goroutineID, i)).
						WithAggregateType("ConcurrentTest")

					event := cqrs.NewBaseDomainEventMessage(
						"ConcurrentTest",
						map[string]interface{}{
							"goroutine": goroutineID,
							"index":     i,
						},
						[]*cqrs.BaseEventMessageOptions{baseOptions},
					)

					err := eventBus.Publish(ctx, event)
					if err != nil {
						b.Errorf("Failed to publish event: %v", err)
					}
				}
			}(g)
		}

		wg.Wait()
	})
}

// BenchmarkMemoryUsage benchmarks memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	ctx := context.Background()
	testContainer, err := NewTestRedisContainer(ctx)
	if err != nil {
		b.Fatalf("Failed to create test container: %v", err)
	}
	defer testContainer.Close(ctx)

	eventBus, err := NewRedisStreamEventBus(testContainer.client, testContainer.config)
	if err != nil {
		b.Fatalf("Failed to create EventBus: %v", err)
	}

	err = eventBus.Start(ctx)
	if err != nil {
		b.Fatalf("Failed to start EventBus: %v", err)
	}
	defer eventBus.Stop(ctx)

	b.Run("LargeEventData", func(b *testing.B) {
		// Create large event data to test memory handling
		largeData := make(map[string]interface{})
		for i := 0; i < 1000; i++ {
			largeData[fmt.Sprintf("field_%d", i)] = fmt.Sprintf("value_%d_with_some_longer_text_to_increase_size", i)
		}

		for i := 0; i < b.N; i++ {
			baseOptions := cqrs.Options().
				WithAggregateID(fmt.Sprintf("large-event-%d", i)).
				WithAggregateType("LargeEvent")

			event := cqrs.NewBaseDomainEventMessage(
				"LargeEvent",
				largeData,
				[]*cqrs.BaseEventMessageOptions{baseOptions},
			)

			err := eventBus.Publish(ctx, event)
			if err != nil {
				b.Fatalf("Failed to publish large event: %v", err)
			}
		}
	})
}

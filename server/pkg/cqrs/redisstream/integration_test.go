package redisstream

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cqrs"
)

// IntegrationTestSuite runs comprehensive integration tests
func TestIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := context.Background()
	testContainer, err := NewTestRedisContainer(ctx)
	require.NoError(t, err)
	defer testContainer.Close(ctx)

	t.Run("FullSystemIntegration", func(t *testing.T) {
		testFullSystemIntegration(t, testContainer)
	})

	t.Run("SerializationIntegration", func(t *testing.T) {
		testSerializationIntegration(t, testContainer)
	})

	t.Run("PriorityStreamIntegration", func(t *testing.T) {
		testPriorityStreamIntegration(t, testContainer)
	})

	t.Run("DLQIntegration", func(t *testing.T) {
		testDLQIntegration(t, testContainer)
	})

	t.Run("RetryPolicyIntegration", func(t *testing.T) {
		testRetryPolicyIntegration(t, testContainer)
	})

	t.Run("CircuitBreakerIntegration", func(t *testing.T) {
		testCircuitBreakerIntegration(t, testContainer)
	})

	t.Run("HealthCheckIntegration", func(t *testing.T) {
		testHealthCheckIntegration(t, testContainer)
	})

	t.Run("PerformanceIntegration", func(t *testing.T) {
		testPerformanceIntegration(t, testContainer)
	})
}

// testFullSystemIntegration tests all components working together
func testFullSystemIntegration(t *testing.T, testContainer *TestRedisContainer) {
	ctx := context.Background()

	// Setup configuration with all features enabled
	config := testContainer.config
	config.Stream.EnablePriorityStreams = true
	config.Stream.DLQEnabled = true
	config.Retry.MaxAttempts = 2
	config.Monitoring.CircuitBreakerEnabled = true
	config.Monitoring.FailureThreshold = 3

	// Create all managers
	eventBus, err := NewRedisStreamEventBus(testContainer.client, config)
	require.NoError(t, err)

	priorityManager, err := NewPriorityStreamManager(config)
	require.NoError(t, err)

	dlqManager, err := NewDLQManager(config)
	require.NoError(t, err)

	retryManager, err := NewRetryPolicyManager(config)
	require.NoError(t, err)

	cbManager := NewCircuitBreakerManager(config)

	healthChecker, err := NewHealthChecker("integration-test", config)
	require.NoError(t, err)

	// Start EventBus
	err = eventBus.Start(ctx)
	require.NoError(t, err)
	defer eventBus.Stop(ctx)

	// Setup health checks
	healthChecker.AddCheck("redis", NewRedisHealthCheck(testContainer.client))
	healthChecker.AddCheck("eventbus", NewEventBusHealthCheck(eventBus))
	healthChecker.AddCheck("circuit_breakers", NewCircuitBreakerHealthCheck(cbManager))

	err = healthChecker.Start(ctx)
	require.NoError(t, err)
	defer healthChecker.Stop(ctx)

	// Create test handlers
	successHandler := &IntegrationTestHandler{
		name:        "success-handler",
		handlerType: cqrs.ProjectionHandler,
		eventTypes:  []string{"IntegrationTest"},
		shouldFail:  false,
	}

	failingHandler := &IntegrationTestHandler{
		name:        "failing-handler",
		handlerType: cqrs.ProcessManagerHandler,
		eventTypes:  []string{"IntegrationTest"},
		shouldFail:  true,
	}

	// Wrap with circuit breakers
	protectedSuccessHandler := NewCircuitBreakerProtectedHandler(successHandler, config)
	protectedFailingHandler := NewCircuitBreakerProtectedHandler(failingHandler, config)

	// Subscribe handlers
	successSubID, err := eventBus.Subscribe("IntegrationTest", protectedSuccessHandler)
	require.NoError(t, err)
	defer eventBus.Unsubscribe(successSubID)

	failingSubID, err := eventBus.Subscribe("IntegrationTest", protectedFailingHandler)
	require.NoError(t, err)
	defer eventBus.Unsubscribe(failingSubID)

	time.Sleep(200 * time.Millisecond) // Allow subscriptions to be ready

	// Test scenario: Publish events with different priorities
	priorities := []cqrs.EventPriority{cqrs.Critical, cqrs.High, cqrs.Normal, cqrs.Low}

	for i, priority := range priorities {
		baseOptions := cqrs.Options().
			WithAggregateID(fmt.Sprintf("integration-test-%d", i)).
			WithAggregateType("IntegrationTest")

		domainOptions := &cqrs.BaseDomainEventMessageOptions{}
		domainOptions.Priority = &priority

		event := cqrs.NewBaseDomainEventMessage(
			"IntegrationTest",
			map[string]interface{}{
				"test_id":  i,
				"priority": priority.String(),
				"data":     fmt.Sprintf("Integration test data %d", i),
			},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
			domainOptions,
		)

		err := eventBus.Publish(ctx, event)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond) // Allow processing
	}

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Verify results
	assert.Greater(t, successHandler.GetProcessedCount(), 0, "Success handler should have processed events")

	// Check EventBus metrics
	busMetrics := eventBus.GetMetrics()
	assert.Greater(t, busMetrics.PublishedEvents, int64(0))
	assert.Greater(t, busMetrics.ProcessedEvents, int64(0))

	// Check priority metrics
	if priorityManager.IsPriorityEnabled() {
		priorityMetrics := priorityManager.GetPriorityMetrics()
		assert.Greater(t, priorityMetrics[cqrs.Critical].PublishedEvents, int64(0))
	}

	// Check health status
	healthSummary := healthChecker.CheckHealth(ctx)
	assert.NotNil(t, healthSummary)
	assert.Equal(t, HealthStatusHealthy, healthSummary.OverallStatus)

	// Check circuit breaker status
	cbMetrics := cbManager.GetAllMetrics()
	assert.Greater(t, len(cbMetrics), 0)

	t.Logf("Integration test completed successfully")
	t.Logf("Published events: %d", busMetrics.PublishedEvents)
	t.Logf("Processed events: %d", busMetrics.ProcessedEvents)
	t.Logf("Failed events: %d", busMetrics.FailedEvents)
	t.Logf("Success handler processed: %d", successHandler.GetProcessedCount())
	t.Logf("Failing handler processed: %d", failingHandler.GetProcessedCount())
}

// testSerializationIntegration tests serialization with complex data
func testSerializationIntegration(t *testing.T, testContainer *TestRedisContainer) {
	ctx := context.Background()

	// Test different serialization formats
	serializers := map[string]EventSerializer{
		"json": NewJSONEventSerializer(),
	}

	for serializerName, serializer := range serializers {
		t.Run(fmt.Sprintf("serializer_%s", serializerName), func(t *testing.T) {
			eventBus, err := NewRedisStreamEventBusWithSerializer(testContainer.client, testContainer.config, serializer)
			require.NoError(t, err)

			err = eventBus.Start(ctx)
			require.NoError(t, err)
			defer eventBus.Stop(ctx)

			// Create complex event data
			complexData := map[string]interface{}{
				"user": map[string]interface{}{
					"id":   "user-123",
					"name": "Complex User",
					"profile": map[string]interface{}{
						"age":         30,
						"preferences": []string{"tech", "gaming", "music"},
						"metadata": map[string]interface{}{
							"last_login":   time.Now().Format(time.RFC3339),
							"login_count":  42,
							"premium_user": true,
							"balance":      123.45,
						},
					},
				},
				"action": map[string]interface{}{
					"type":       "profile_update",
					"timestamp":  time.Now().Unix(),
					"changes":    []string{"email", "preferences"},
					"ip_address": "192.168.1.100",
				},
			}

			testHandler := &IntegrationTestHandler{
				name:        "serialization-test-handler",
				handlerType: cqrs.ProjectionHandler,
				eventTypes:  []string{"ComplexSerializationTest"},
				shouldFail:  false,
			}

			subID, err := eventBus.Subscribe("ComplexSerializationTest", testHandler)
			require.NoError(t, err)
			defer eventBus.Unsubscribe(subID)

			time.Sleep(100 * time.Millisecond)

			baseOptions := cqrs.Options().
				WithAggregateID("complex-123").
				WithAggregateType("ComplexAggregate").
				WithVersion(1)

			event := cqrs.NewBaseDomainEventMessage(
				"ComplexSerializationTest",
				complexData,
				[]*cqrs.BaseEventMessageOptions{baseOptions},
			)

			err = eventBus.Publish(ctx, event)
			require.NoError(t, err)

			// Wait for processing
			assert.Eventually(t, func() bool {
				return testHandler.GetProcessedCount() > 0
			}, 3*time.Second, 100*time.Millisecond)

			// Verify complex data was preserved
			processedEvent := testHandler.GetLastProcessedEvent()
			assert.NotNil(t, processedEvent)

			processedData := processedEvent.EventData().(map[string]interface{})
			user := processedData["user"].(map[string]interface{})
			assert.Equal(t, "Complex User", user["name"])
		})
	}
}

// testPriorityStreamIntegration tests priority stream functionality
func testPriorityStreamIntegration(t *testing.T, testContainer *TestRedisContainer) {
	ctx := context.Background()

	config := testContainer.config
	config.Stream.EnablePriorityStreams = true

	eventBus, err := NewRedisStreamEventBus(testContainer.client, config)
	require.NoError(t, err)

	priorityManager, err := NewPriorityStreamManager(config)
	require.NoError(t, err)

	err = eventBus.Start(ctx)
	require.NoError(t, err)
	defer eventBus.Stop(ctx)

	// Create priority-aware handler
	priorityHandler := &PriorityAwareHandler{
		processedByPriority: make(map[cqrs.EventPriority]int),
	}

	subID, err := eventBus.Subscribe("PriorityTest", priorityHandler)
	require.NoError(t, err)
	defer eventBus.Unsubscribe(subID)

	time.Sleep(100 * time.Millisecond)

	// Publish events with different priorities
	priorities := []cqrs.EventPriority{cqrs.Low, cqrs.Normal, cqrs.High, cqrs.Critical}
	eventsPerPriority := 5

	for _, priority := range priorities {
		for i := 0; i < eventsPerPriority; i++ {
			baseOptions := cqrs.Options().
				WithAggregateID(fmt.Sprintf("priority-test-%s-%d", priority.String(), i)).
				WithAggregateType("PriorityTest")

			domainOptions := &cqrs.BaseDomainEventMessageOptions{}
			domainOptions.Priority = &priority

			event := cqrs.NewBaseDomainEventMessage(
				"PriorityTest",
				map[string]interface{}{
					"priority": priority.String(),
					"index":    i,
				},
				[]*cqrs.BaseEventMessageOptions{baseOptions},
				domainOptions,
			)

			err := eventBus.Publish(ctx, event)
			require.NoError(t, err)
		}
	}

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Verify priority metrics
	priorityMetrics := priorityManager.GetPriorityMetrics()
	for _, priority := range priorities {
		assert.Equal(t, int64(eventsPerPriority), priorityMetrics[priority].PublishedEvents,
			"Priority %s should have %d published events", priority.String(), eventsPerPriority)
	}

	// Verify handler processed events by priority
	for _, priority := range priorities {
		count := priorityHandler.GetProcessedCountByPriority(priority)
		assert.Greater(t, count, 0, "Priority %s should have processed events", priority.String())
	}
}

// testDLQIntegration tests DLQ functionality
func testDLQIntegration(t *testing.T, testContainer *TestRedisContainer) {
	ctx := context.Background()

	config := testContainer.config
	config.Stream.DLQEnabled = true
	config.Retry.MaxAttempts = 2

	eventBus, err := NewRedisStreamEventBus(testContainer.client, config)
	require.NoError(t, err)

	dlqManager, err := NewDLQManager(config)
	require.NoError(t, err)

	err = eventBus.Start(ctx)
	require.NoError(t, err)
	defer eventBus.Stop(ctx)

	// Create always-failing handler
	failingHandler := &IntegrationTestHandler{
		name:        "dlq-test-handler",
		handlerType: cqrs.ProcessManagerHandler,
		eventTypes:  []string{"DLQTest"},
		shouldFail:  true,
	}

	subID, err := eventBus.Subscribe("DLQTest", failingHandler)
	require.NoError(t, err)
	defer eventBus.Unsubscribe(subID)

	time.Sleep(100 * time.Millisecond)

	// Publish events that will fail
	numEvents := 3
	for i := 0; i < numEvents; i++ {
		baseOptions := cqrs.Options().
			WithAggregateID(fmt.Sprintf("dlq-test-%d", i)).
			WithAggregateType("DLQTest")

		event := cqrs.NewBaseDomainEventMessage(
			"DLQTest",
			map[string]interface{}{
				"test_id": i,
				"data":    fmt.Sprintf("This will fail %d", i),
			},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		err := eventBus.Publish(ctx, event)
		require.NoError(t, err)
	}

	// Wait for retries and DLQ processing
	time.Sleep(5 * time.Second)

	// Check DLQ statistics
	dlqStats := dlqManager.GetDLQStatistics()
	assert.Greater(t, dlqStats.TotalDLQEvents, int64(0), "Some events should be moved to DLQ")

	t.Logf("DLQ Stats: Total=%d", dlqStats.TotalDLQEvents)
}

// testRetryPolicyIntegration tests retry policy functionality
func testRetryPolicyIntegration(t *testing.T, testContainer *TestRedisContainer) {
	ctx := context.Background()

	config := testContainer.config
	config.Retry.MaxAttempts = 3
	config.Retry.InitialDelay = 50 * time.Millisecond
	config.Retry.BackoffType = "exponential"

	retryManager, err := NewRetryPolicyManager(config)
	require.NoError(t, err)

	// Test different retry policies
	t.Run("exponential_backoff", func(t *testing.T) {
		policy := &RetryPolicy{
			MaxAttempts:   3,
			InitialDelay:  100 * time.Millisecond,
			MaxDelay:      1 * time.Second,
			BackoffType:   ExponentialBackoff,
			BackoffFactor: 2.0,
		}

		delays := []time.Duration{
			retryManager.CalculateDelay(policy, 1),
			retryManager.CalculateDelay(policy, 2),
			retryManager.CalculateDelay(policy, 3),
		}

		assert.Equal(t, 100*time.Millisecond, delays[0])
		assert.Equal(t, 200*time.Millisecond, delays[1])
		assert.Equal(t, 400*time.Millisecond, delays[2])
	})

	// Test custom retry policies
	customPolicy := &RetryPolicy{
		MaxAttempts:   5,
		InitialDelay:  50 * time.Millisecond,
		MaxDelay:      500 * time.Millisecond,
		BackoffType:   LinearBackoff,
		BackoffFactor: 1.5,
	}

	err = retryManager.SetHandlerRetryPolicy("custom-handler", customPolicy)
	require.NoError(t, err)

	retrievedPolicy := retryManager.GetHandlerRetryPolicy("custom-handler")
	assert.Equal(t, customPolicy.MaxAttempts, retrievedPolicy.MaxAttempts)
	assert.Equal(t, customPolicy.BackoffType, retrievedPolicy.BackoffType)
}

// testCircuitBreakerIntegration tests circuit breaker functionality
func testCircuitBreakerIntegration(t *testing.T, testContainer *TestRedisContainer) {
	ctx := context.Background()

	config := testContainer.config
	config.Monitoring.CircuitBreakerEnabled = true
	config.Monitoring.FailureThreshold = 2
	config.Monitoring.RecoveryTimeout = 100 * time.Millisecond

	cbManager := NewCircuitBreakerManager(config)
	cb := cbManager.CreateCircuitBreaker("integration-test-service")

	// Test circuit breaker state transitions
	assert.Equal(t, CircuitBreakerStateClosed, cb.GetState())

	// Trip the circuit breaker
	for i := 0; i < 2; i++ {
		err := cb.Call(func() error {
			return fmt.Errorf("test failure %d", i)
		})
		if i < 1 {
			assert.Error(t, err)
			assert.NotErrorIs(t, err, ErrCircuitBreakerOpen)
		} else {
			assert.ErrorIs(t, err, ErrCircuitBreakerOpen)
		}
	}

	assert.Equal(t, CircuitBreakerStateOpen, cb.GetState())

	// Wait for recovery
	time.Sleep(150 * time.Millisecond)

	// Should transition to half-open and then closed on success
	err := cb.Call(func() error {
		return nil // Success
	})
	assert.NoError(t, err)
	assert.Equal(t, CircuitBreakerStateClosed, cb.GetState())

	// Verify metrics
	metrics := cb.GetMetrics()
	assert.Greater(t, metrics.TotalCalls, int64(0))
	assert.Greater(t, metrics.StateTransitions, int64(0))
}

// testHealthCheckIntegration tests health check functionality
func testHealthCheckIntegration(t *testing.T, testContainer *TestRedisContainer) {
	ctx := context.Background()

	config := testContainer.config
	config.Monitoring.HealthCheckInterval = 100 * time.Millisecond

	healthChecker, err := NewHealthChecker("integration-test", config)
	require.NoError(t, err)

	// Add various health checks
	healthChecker.AddCheck("redis", NewRedisHealthCheck(testContainer.client))

	healthChecker.AddCheck("custom_healthy", NewCustomHealthCheck("custom_healthy", func(ctx context.Context) HealthCheckResult {
		return HealthCheckResult{
			Status:  HealthStatusHealthy,
			Message: "All good",
		}
	}))

	healthChecker.AddCheck("custom_degraded", NewCustomHealthCheck("custom_degraded", func(ctx context.Context) HealthCheckResult {
		return HealthCheckResult{
			Status:  HealthStatusDegraded,
			Message: "Performance issues",
		}
	}))

	// Perform health check
	summary := healthChecker.CheckHealth(ctx)

	assert.Equal(t, HealthStatusDegraded, summary.OverallStatus) // Worst of all checks
	assert.Len(t, summary.Checks, 3)
	assert.Equal(t, HealthStatusHealthy, summary.Checks["redis"].Status)
	assert.Equal(t, HealthStatusHealthy, summary.Checks["custom_healthy"].Status)
	assert.Equal(t, HealthStatusDegraded, summary.Checks["custom_degraded"].Status)

	// Test periodic checking
	err = healthChecker.Start(ctx)
	require.NoError(t, err)

	time.Sleep(250 * time.Millisecond) // Let it run a few checks

	err = healthChecker.Stop(ctx)
	require.NoError(t, err)

	history := healthChecker.GetHealthHistory(5)
	assert.Greater(t, len(history), 1, "Should have multiple health check results")
}

// testPerformanceIntegration tests system performance under load
func testPerformanceIntegration(t *testing.T, testContainer *TestRedisContainer) {
	ctx := context.Background()

	config := testContainer.config
	config.Stream.EnablePriorityStreams = true

	eventBus, err := NewRedisStreamEventBus(testContainer.client, config)
	require.NoError(t, err)

	err = eventBus.Start(ctx)
	require.NoError(t, err)
	defer eventBus.Stop(ctx)

	// Create performance test handler
	perfHandler := &PerformanceTestHandler{
		processedEvents: make(map[string]int),
	}

	subID, err := eventBus.Subscribe("PerformanceTest", perfHandler)
	require.NoError(t, err)
	defer eventBus.Unsubscribe(subID)

	time.Sleep(100 * time.Millisecond)

	// Performance test parameters
	numEvents := 100
	batchSize := 10

	start := time.Now()

	// Publish events in batches
	for batch := 0; batch < numEvents/batchSize; batch++ {
		events := make([]cqrs.EventMessage, batchSize)

		for i := 0; i < batchSize; i++ {
			eventIndex := batch*batchSize + i
			baseOptions := cqrs.Options().
				WithAggregateID(fmt.Sprintf("perf-test-%d", eventIndex)).
				WithAggregateType("PerformanceTest")

			events[i] = cqrs.NewBaseDomainEventMessage(
				"PerformanceTest",
				map[string]interface{}{
					"batch":       batch,
					"index":       i,
					"event_index": eventIndex,
					"timestamp":   time.Now().UnixNano(),
				},
				[]*cqrs.BaseEventMessageOptions{baseOptions},
			)
		}

		err := eventBus.PublishBatch(ctx, events)
		require.NoError(t, err)
	}

	publishDuration := time.Since(start)

	// Wait for all events to be processed
	assert.Eventually(t, func() bool {
		return perfHandler.GetTotalProcessed() >= numEvents
	}, 10*time.Second, 100*time.Millisecond, "All events should be processed")

	totalDuration := time.Since(start)

	// Performance assertions
	assert.LessOrEqual(t, publishDuration, 5*time.Second, "Publishing should complete within 5 seconds")
	assert.LessOrEqual(t, totalDuration, 10*time.Second, "Processing should complete within 10 seconds")

	// Calculate throughput
	publishThroughput := float64(numEvents) / publishDuration.Seconds()
	processingThroughput := float64(numEvents) / totalDuration.Seconds()

	t.Logf("Performance Results:")
	t.Logf("  Events: %d", numEvents)
	t.Logf("  Publish Duration: %v", publishDuration)
	t.Logf("  Total Duration: %v", totalDuration)
	t.Logf("  Publish Throughput: %.2f events/sec", publishThroughput)
	t.Logf("  Processing Throughput: %.2f events/sec", processingThroughput)

	// Verify metrics
	metrics := eventBus.GetMetrics()
	assert.Equal(t, int64(numEvents), metrics.PublishedEvents)
	assert.GreaterOrEqual(t, metrics.ProcessedEvents, int64(numEvents))
}

// Test helper types

type IntegrationTestHandler struct {
	*cqrs.BaseEventHandler
	processedEvents []cqrs.EventMessage
	processedCount  int
	shouldFail      bool
	mu              sync.RWMutex
}

func (h *IntegrationTestHandler) Handle(ctx context.Context, event cqrs.EventMessage) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.shouldFail {
		return fmt.Errorf("intentional failure for event %s", event.EventID())
	}

	h.processedEvents = append(h.processedEvents, event)
	h.processedCount++
	return nil
}

func (h *IntegrationTestHandler) GetProcessedCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.processedCount
}

func (h *IntegrationTestHandler) GetLastProcessedEvent() cqrs.EventMessage {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.processedEvents) == 0 {
		return nil
	}
	return h.processedEvents[len(h.processedEvents)-1]
}

type PriorityAwareHandler struct {
	*cqrs.BaseEventHandler
	processedByPriority map[cqrs.EventPriority]int
	mu                  sync.RWMutex
}

func (h *PriorityAwareHandler) Handle(ctx context.Context, event cqrs.EventMessage) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if domainEvent, ok := event.(cqrs.DomainEventMessage); ok {
		priority := domainEvent.GetPriority()
		h.processedByPriority[priority]++
	}

	return nil
}

func (h *PriorityAwareHandler) CanHandle(eventType string) bool {
	return eventType == "PriorityTest"
}

func (h *PriorityAwareHandler) GetHandlerName() string {
	return "priority-aware-handler"
}

func (h *PriorityAwareHandler) GetHandlerType() cqrs.HandlerType {
	return cqrs.ProjectionHandler
}

func (h *PriorityAwareHandler) GetProcessedCountByPriority(priority cqrs.EventPriority) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.processedByPriority[priority]
}

type PerformanceTestHandler struct {
	processedEvents map[string]int
	mu              sync.RWMutex
}

func (h *PerformanceTestHandler) Handle(ctx context.Context, event cqrs.EventMessage) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.processedEvents[event.EventType()]++
	return nil
}

func (h *PerformanceTestHandler) CanHandle(eventType string) bool {
	return eventType == "PerformanceTest"
}

func (h *PerformanceTestHandler) GetHandlerName() string {
	return "performance-test-handler"
}

func (h *PerformanceTestHandler) GetHandlerType() cqrs.HandlerType {
	return cqrs.ProjectionHandler
}

func (h *PerformanceTestHandler) GetTotalProcessed() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	total := 0
	for _, count := range h.processedEvents {
		total += count
	}
	return total
}

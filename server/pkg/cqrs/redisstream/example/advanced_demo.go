package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	
	"defense-allies-server/pkg/cqrs"
	"defense-allies-server/pkg/cqrs/redisstream"
)

// Advanced example demonstrating all Phase 2 & 3 features
// This example shows:
// - Enhanced Serialization
// - Priority Stream Management  
// - Dead Letter Queue
// - Retry Policy Management
// - Circuit Breaker Pattern
// - Health Check System

// UserRegistrationHandler handles user registration events
type UserRegistrationHandler struct {
	*cqrs.BaseEventHandler
	processedEvents int
	shouldFail      bool
}

func NewUserRegistrationHandler() *UserRegistrationHandler {
	base := cqrs.NewBaseEventHandler(
		"user-registration-handler",
		cqrs.ProjectionHandler,
		[]string{"UserRegistered", "UserActivated"},
	)
	
	return &UserRegistrationHandler{
		BaseEventHandler: base,
	}
}

func (h *UserRegistrationHandler) Handle(ctx context.Context, event cqrs.EventMessage) error {
	h.processedEvents++
	
	// Simulate occasional failures for retry/DLQ demonstration
	if h.shouldFail && h.processedEvents%5 == 0 {
		return fmt.Errorf("simulated processing failure for event %s", event.EventID())
	}
	
	fmt.Printf("✅ [UserRegistrationHandler] Processed %s: %s (Total: %d)\n", 
		event.EventType(), 
		event.ID(), 
		h.processedEvents,
	)
	
	// Simulate processing time
	time.Sleep(10 * time.Millisecond)
	
	return nil
}

func (h *UserRegistrationHandler) SetFailureMode(shouldFail bool) {
	h.shouldFail = shouldFail
}

func (h *UserRegistrationHandler) GetProcessedCount() int {
	return h.processedEvents
}

// NotificationHandler handles notification events
type NotificationHandler struct {
	*cqrs.BaseEventHandler
	sentNotifications int
}

func NewNotificationHandler() *NotificationHandler {
	base := cqrs.NewBaseEventHandler(
		"notification-handler",
		cqrs.NotificationHandler,
		[]string{"UserRegistered", "OrderCreated"},
	)
	
	return &NotificationHandler{
		BaseEventHandler: base,
	}
}

func (h *NotificationHandler) Handle(ctx context.Context, event cqrs.EventMessage) error {
	h.sentNotifications++
	
	fmt.Printf("📧 [NotificationHandler] Sent notification for %s: %s (Total: %d)\n", 
		event.EventType(), 
		event.ID(), 
		h.sentNotifications,
	)
	
	return nil
}

func (h *NotificationHandler) GetSentCount() int {
	return h.sentNotifications
}

func main() {
	fmt.Println("🚀 Starting Advanced Redis Stream EventBus Demo")
	fmt.Println("===================================================")
	
	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutdown signal received, cleaning up...")
		cancel()
	}()
	
	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer rdb.Close()
	
	// Test Redis connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	fmt.Println("✅ Connected to Redis successfully")
	
	// Create advanced configuration
	config := createAdvancedConfig()
	
	// Create EventBus with JSON serializer
	jsonSerializer := redisstream.NewJSONEventSerializer()
	eventBus, err := redisstream.NewRedisStreamEventBusWithSerializer(rdb, config, jsonSerializer)
	if err != nil {
		log.Fatalf("❌ Failed to create EventBus: %v", err)
	}
	fmt.Println("✅ Created EventBus with JSON serialization")
	
	// Start EventBus
	if err := eventBus.Start(ctx); err != nil {
		log.Fatalf("❌ Failed to start EventBus: %v", err)
	}
	defer eventBus.Stop(ctx)
	fmt.Println("✅ EventBus started successfully")
	
	// Create Priority Stream Manager
	priorityManager, err := redisstream.NewPriorityStreamManager(config)
	if err != nil {
		log.Fatalf("❌ Failed to create Priority Manager: %v", err)
	}
	fmt.Println("✅ Priority Stream Manager created")
	
	// Create DLQ Manager
	dlqManager, err := redisstream.NewDLQManager(config)
	if err != nil {
		log.Fatalf("❌ Failed to create DLQ Manager: %v", err)
	}
	fmt.Println("✅ Dead Letter Queue Manager created")
	
	// Create Retry Policy Manager
	retryManager, err := redisstream.NewRetryPolicyManager(config)
	if err != nil {
		log.Fatalf("❌ Failed to create Retry Manager: %v", err)
	}
	fmt.Println("✅ Retry Policy Manager created")
	
	// Create Circuit Breaker Manager
	cbManager := redisstream.NewCircuitBreakerManager(config)
	fmt.Println("✅ Circuit Breaker Manager created")
	
	// Create Health Checker
	healthChecker, err := redisstream.NewHealthChecker("demo-service", config)
	if err != nil {
		log.Fatalf("❌ Failed to create Health Checker: %v", err)
	}
	fmt.Println("✅ Health Checker created")
	
	// Setup health checks
	setupHealthChecks(healthChecker, rdb, eventBus, cbManager)
	
	// Start health checking
	if err := healthChecker.Start(ctx); err != nil {
		log.Fatalf("❌ Failed to start Health Checker: %v", err)
	}
	defer healthChecker.Stop(ctx)
	fmt.Println("✅ Health monitoring started")
	
	// Create event handlers
	userHandler := NewUserRegistrationHandler()
	notificationHandler := NewNotificationHandler()
	
	// Wrap handlers with circuit breaker protection
	protectedUserHandler := redisstream.NewCircuitBreakerProtectedHandler(userHandler, config)
	protectedNotificationHandler := redisstream.NewCircuitBreakerProtectedHandler(notificationHandler, config)
	
	// Subscribe handlers
	userSubID, err := eventBus.Subscribe("UserRegistered", protectedUserHandler)
	if err != nil {
		log.Fatalf("❌ Failed to subscribe user handler: %v", err)
	}
	defer eventBus.Unsubscribe(userSubID)
	
	notifSubID, err := eventBus.Subscribe("UserRegistered", protectedNotificationHandler)
	if err != nil {
		log.Fatalf("❌ Failed to subscribe notification handler: %v", err)
	}
	defer eventBus.Unsubscribe(notifSubID)
	
	orderSubID, err := eventBus.Subscribe("OrderCreated", protectedNotificationHandler)
	if err != nil {
		log.Fatalf("❌ Failed to subscribe order handler: %v", err)
	}
	defer eventBus.Unsubscribe(orderSubID)
	
	fmt.Println("✅ Event handlers subscribed with circuit breaker protection")
	
	// Give time for subscriptions to be ready
	time.Sleep(500 * time.Millisecond)
	
	fmt.Println("\n🎯 Starting event publishing demonstration...")
	
	// Demonstrate different scenarios
	demonstrateBasicPublishing(ctx, eventBus)
	demonstratePriorityPublishing(ctx, eventBus)
	demonstrateFailureScenarios(ctx, eventBus, userHandler)
	demonstrateHealthMonitoring(healthChecker)
	
	// Show final statistics
	showFinalStatistics(eventBus, priorityManager, dlqManager, retryManager, cbManager, healthChecker)
	
	fmt.Println("\n✅ Demo completed successfully!")
	fmt.Println("Press Ctrl+C to exit...")
	
	// Keep running until shutdown
	<-ctx.Done()
}

func createAdvancedConfig() *redisstream.RedisStreamConfig {
	config := redisstream.DefaultRedisStreamConfig()
	
	// Enable all advanced features
	config.Stream.EnablePriorityStreams = true
	config.Stream.DLQEnabled = true
	config.Retry.MaxAttempts = 3
	config.Retry.InitialDelay = 100 * time.Millisecond
	config.Retry.BackoffType = "exponential"
	config.Consumer.ServiceName = "demo-service"
	config.Consumer.InstanceID = "demo-instance-1"
	config.Monitoring.CircuitBreakerEnabled = true
	config.Monitoring.FailureThreshold = 3
	config.Monitoring.RecoveryTimeout = 10 * time.Second
	config.Monitoring.HealthCheckInterval = 5 * time.Second
	
	return config
}

func setupHealthChecks(hc redisstream.HealthChecker, rdb redis.UniversalClient, eventBus *redisstream.RedisStreamEventBus, cbManager redisstream.CircuitBreakerManager) {
	// Add Redis health check
	hc.AddCheck("redis", redisstream.NewRedisHealthCheck(rdb))
	
	// Add EventBus health check
	hc.AddCheck("eventbus", redisstream.NewEventBusHealthCheck(eventBus))
	
	// Add Circuit Breaker health check
	hc.AddCheck("circuit_breakers", redisstream.NewCircuitBreakerHealthCheck(cbManager))
	
	// Add custom application health check
	hc.AddCheck("application", redisstream.NewCustomHealthCheck("application", func(ctx context.Context) redisstream.HealthCheckResult {
		return redisstream.HealthCheckResult{
			Status:  redisstream.HealthStatusHealthy,
			Message: "Application is running normally",
			Details: map[string]interface{}{
				"uptime":    time.Since(time.Now().Add(-30 * time.Second)).String(),
				"version":   "1.0.0",
				"build":     "demo-build",
				"go_version": "1.22",
			},
		}
	}))
}

func demonstrateBasicPublishing(ctx context.Context, eventBus *redisstream.RedisStreamEventBus) {
	fmt.Println("\n📤 Demonstrating basic event publishing...")
	
	for i := 1; i <= 5; i++ {
		baseOptions := cqrs.Options().
			WithAggregateID(fmt.Sprintf("user-%d", i)).
			WithAggregateType("User").
			WithVersion(1)
			
		event := cqrs.NewBaseDomainEventMessage(
			"UserRegistered",
			map[string]interface{}{
				"name":  fmt.Sprintf("User %d", i),
				"email": fmt.Sprintf("user%d@example.com", i),
				"profile": map[string]interface{}{
					"age":      25 + i,
					"country":  "South Korea",
					"interests": []string{"technology", "gaming"},
				},
			},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)
		
		if err := eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("❌ Failed to publish event: %v\n", err)
		} else {
			fmt.Printf("📤 Published UserRegistered for user-%d\n", i)
		}
		
		time.Sleep(200 * time.Millisecond)
	}
}

func demonstratePriorityPublishing(ctx context.Context, eventBus *redisstream.RedisStreamEventBus) {
	fmt.Println("\n⚡ Demonstrating priority-based event publishing...")
	
	priorities := []cqrs.EventPriority{cqrs.Critical, cqrs.High, cqrs.Normal, cqrs.Low}
	
	for i, priority := range priorities {
		baseOptions := cqrs.Options().
			WithAggregateID(fmt.Sprintf("order-%d", i+1)).
			WithAggregateType("Order").
			WithVersion(1)
			
		domainOptions := &cqrs.BaseDomainEventMessageOptions{}
		domainOptions.Priority = &priority
		domainOptions.Category = &[]cqrs.EventCategory{cqrs.DomainEvent}[0]
		
		event := cqrs.NewBaseDomainEventMessage(
			"OrderCreated",
			map[string]interface{}{
				"order_id": fmt.Sprintf("order-%d", i+1),
				"amount":   100.00 * float64(i+1),
				"priority": priority.String(),
			},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
			domainOptions,
		)
		
		if err := eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("❌ Failed to publish priority event: %v\n", err)
		} else {
			fmt.Printf("📤 Published %s priority OrderCreated for order-%d\n", priority.String(), i+1)
		}
		
		time.Sleep(100 * time.Millisecond)
	}
}

func demonstrateFailureScenarios(ctx context.Context, eventBus *redisstream.RedisStreamEventBus, userHandler *UserRegistrationHandler) {
	fmt.Println("\n💥 Demonstrating failure scenarios (retry/DLQ)...")
	
	// Enable failure mode
	userHandler.SetFailureMode(true)
	
	// Publish events that will trigger failures
	for i := 1; i <= 8; i++ {
		baseOptions := cqrs.Options().
			WithAggregateID(fmt.Sprintf("failing-user-%d", i)).
			WithAggregateType("User").
			WithVersion(1)
			
		event := cqrs.NewBaseDomainEventMessage(
			"UserRegistered",
			map[string]interface{}{
				"name":  fmt.Sprintf("Failing User %d", i),
				"email": fmt.Sprintf("failing-user%d@example.com", i),
			},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)
		
		if err := eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("❌ Failed to publish failing event: %v\n", err)
		} else {
			fmt.Printf("📤 Published failing UserRegistered for failing-user-%d\n", i)
		}
		
		time.Sleep(300 * time.Millisecond) // Give time for processing and retries
	}
	
	// Disable failure mode
	userHandler.SetFailureMode(false)
	fmt.Println("✅ Disabled failure mode")
}

func demonstrateHealthMonitoring(healthChecker redisstream.HealthChecker) {
	fmt.Println("\n🔍 Demonstrating health monitoring...")
	
	// Perform manual health check
	ctx := context.Background()
	summary := healthChecker.CheckHealth(ctx)
	
	fmt.Printf("🏥 Overall Health Status: %s\n", summary.OverallStatus.String())
	fmt.Printf("📊 Health Check Summary:\n")
	fmt.Printf("   - Total Checks: %v\n", summary.Summary["total_checks"])
	fmt.Printf("   - Healthy: %v\n", summary.Summary["healthy_count"])
	fmt.Printf("   - Degraded: %v\n", summary.Summary["degraded_count"])
	fmt.Printf("   - Unhealthy: %v\n", summary.Summary["unhealthy_count"])
	
	for checkName, result := range summary.Checks {
		status := result.Status.String()
		emoji := "✅"
		if result.Status == redisstream.HealthStatusDegraded {
			emoji = "⚠️"
		} else if result.Status == redisstream.HealthStatusUnhealthy {
			emoji = "❌"
		}
		
		fmt.Printf("   %s %s: %s - %s\n", emoji, checkName, status, result.Message)
	}
}

func showFinalStatistics(
	eventBus *redisstream.RedisStreamEventBus,
	priorityManager redisstream.PriorityStreamManager,
	dlqManager redisstream.DLQManager,
	retryManager redisstream.RetryPolicyManager,
	cbManager redisstream.CircuitBreakerManager,
	healthChecker redisstream.HealthChecker,
) {
	fmt.Println("\n📊 FINAL STATISTICS")
	fmt.Println("===================")
	
	// EventBus metrics
	busMetrics := eventBus.GetMetrics()
	fmt.Printf("EventBus Metrics:\n")
	fmt.Printf("  📤 Published Events: %d\n", busMetrics.PublishedEvents)
	fmt.Printf("  ✅ Processed Events: %d\n", busMetrics.ProcessedEvents)
	fmt.Printf("  ❌ Failed Events: %d\n", busMetrics.FailedEvents)
	fmt.Printf("  👥 Active Subscribers: %d\n", busMetrics.ActiveSubscribers)
	fmt.Printf("  ⏱️  Average Latency: %v\n", busMetrics.AverageLatency)
	
	// Priority metrics
	if priorityManager.IsPriorityEnabled() {
		fmt.Printf("\nPriority Metrics:\n")
		priorityMetrics := priorityManager.GetPriorityMetrics()
		ratios := priorityManager.GetPriorityRatios()
		
		for _, priority := range []cqrs.EventPriority{cqrs.Critical, cqrs.High, cqrs.Normal, cqrs.Low} {
			metrics := priorityMetrics[priority]
			ratio := ratios[priority]
			fmt.Printf("  %s: %d events (%.1f%%)\n", 
				priority.String(), metrics.PublishedEvents, ratio*100)
		}
	}
	
	// DLQ metrics
	if dlqManager.IsDLQEnabled() {
		dlqStats := dlqManager.GetDLQStatistics()
		fmt.Printf("\nDLQ Metrics:\n")
		fmt.Printf("  💀 Total DLQ Events: %d\n", dlqStats.TotalDLQEvents)
		fmt.Printf("  📈 Overall DLQ Rate: %.2f%%\n", dlqManager.GetOverallDLQRate()*100)
		
		if len(dlqStats.EventsByReason) > 0 {
			fmt.Printf("  Top Error Reasons:\n")
			topReasons := dlqManager.GetTopErrorReasons(3)
			for _, reason := range topReasons {
				fmt.Printf("    - %s: %d (%.1f%%)\n", reason.Reason, reason.Count, reason.Rate*100)
			}
		}
	}
	
	// Retry metrics
	retryStats := retryManager.GetRetryStatistics()
	fmt.Printf("\nRetry Metrics:\n")
	fmt.Printf("  🔄 Total Retry Attempts: %d\n", retryStats.TotalRetryAttempts)
	fmt.Printf("  ✅ Successful Retries: %d\n", retryStats.SuccessfulRetries)
	fmt.Printf("  ❌ Exhausted Retries: %d\n", retryStats.ExhaustedRetries)
	fmt.Printf("  📈 Retry Success Rate: %.2f%%\n", retryManager.GetOverallRetrySuccessRate()*100)
	
	// Circuit Breaker metrics
	cbMetrics := cbManager.GetAllMetrics()
	fmt.Printf("\nCircuit Breaker Metrics:\n")
	fmt.Printf("  ⚡ Total Circuit Breakers: %d\n", len(cbMetrics))
	
	openCount := 0
	for serviceName, metrics := range cbMetrics {
		if metrics.CurrentState == redisstream.CircuitBreakerStateOpen {
			openCount++
		}
		if metrics.TotalCalls > 0 {
			fmt.Printf("  📊 %s: %d calls, %.1f%% success, %s\n", 
				serviceName, metrics.TotalCalls, metrics.SuccessRate*100, metrics.CurrentState.String())
		}
	}
	
	if openCount > 0 {
		fmt.Printf("  ⚠️  Open Circuit Breakers: %d\n", openCount)
	}
	
	// Health Check summary
	lastHealth := healthChecker.GetLastHealthCheck()
	if lastHealth != nil {
		fmt.Printf("\nHealth Status: %s (%s)\n", 
			lastHealth.OverallStatus.String(), 
			lastHealth.Timestamp.Format("15:04:05"))
	}
	
	fmt.Println("\n🎉 All systems demonstrated successfully!")
}

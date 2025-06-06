package infrastructure

import (
	"context"
	"defense-allies-server/pkg/cqrs"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/04-read-models/projections"
	"fmt"
	"log"
	"sync"
	"time"
)

// EventHandlerRegistry manages event handlers and their lifecycle
type EventHandlerRegistry struct {
	handlers    map[string]cqrs.EventHandler
	eventBus    cqrs.EventBus
	readStore   cqrs.ReadStore
	projManager *projections.ProjectionManager
	mutex       sync.RWMutex
	running     bool
}

// NewEventHandlerRegistry creates a new EventHandlerRegistry
func NewEventHandlerRegistry(eventBus cqrs.EventBus, readStore cqrs.ReadStore) *EventHandlerRegistry {
	projManager := projections.NewProjectionManager(eventBus, readStore)

	return &EventHandlerRegistry{
		handlers:    make(map[string]cqrs.EventHandler),
		eventBus:    eventBus,
		readStore:   readStore,
		projManager: projManager,
	}
}

// RegisterProjections registers all projection handlers
func (ehr *EventHandlerRegistry) RegisterProjections() error {
	ehr.mutex.Lock()
	defer ehr.mutex.Unlock()

	// Create and register UserProjection
	userProjection := projections.NewUserProjection(ehr.readStore)
	if err := ehr.projManager.RegisterProjection(userProjection); err != nil {
		return fmt.Errorf("failed to register UserProjection: %w", err)
	}

	// Create and register OrderProjection
	orderProjection := projections.NewOrderProjection(ehr.readStore)
	if err := ehr.projManager.RegisterProjection(orderProjection); err != nil {
		return fmt.Errorf("failed to register OrderProjection: %w", err)
	}

	// Create and register AnalyticsProjection
	analyticsProjection := projections.NewAnalyticsProjection(ehr.readStore)
	if err := ehr.projManager.RegisterProjection(analyticsProjection); err != nil {
		return fmt.Errorf("failed to register AnalyticsProjection: %w", err)
	}

	log.Printf("EventHandlerRegistry: Registered %d projections", len(ehr.projManager.GetProjectionNames()))
	return nil
}

// RegisterCustomHandler registers a custom event handler
func (ehr *EventHandlerRegistry) RegisterCustomHandler(name string, handler cqrs.EventHandler) error {
	ehr.mutex.Lock()
	defer ehr.mutex.Unlock()

	if _, exists := ehr.handlers[name]; exists {
		return fmt.Errorf("handler %s already registered", name)
	}

	ehr.handlers[name] = handler
	log.Printf("EventHandlerRegistry: Registered custom handler %s", name)
	return nil
}

// UnregisterHandler unregisters an event handler
func (ehr *EventHandlerRegistry) UnregisterHandler(name string) error {
	ehr.mutex.Lock()
	defer ehr.mutex.Unlock()

	if _, exists := ehr.handlers[name]; !exists {
		return fmt.Errorf("handler %s not found", name)
	}

	delete(ehr.handlers, name)
	log.Printf("EventHandlerRegistry: Unregistered handler %s", name)
	return nil
}

// Start starts all registered handlers
func (ehr *EventHandlerRegistry) Start(ctx context.Context) error {
	ehr.mutex.Lock()
	defer ehr.mutex.Unlock()

	if ehr.running {
		return fmt.Errorf("event handler registry is already running")
	}

	// Start projection manager
	if err := ehr.projManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start projection manager: %w", err)
	}

	// Subscribe custom handlers to event bus
	for name, handler := range ehr.handlers {
		if ehr.eventBus != nil {
			// Subscribe to all events for now - in a real implementation,
			// you might want to be more selective
			subscriptionID, err := ehr.eventBus.SubscribeAll(handler)
			if err != nil {
				log.Printf("EventHandlerRegistry: Warning - failed to subscribe handler %s: %v", name, err)
			} else {
				log.Printf("EventHandlerRegistry: Subscribed handler %s with ID %s", name, subscriptionID)
			}
		}
	}

	ehr.running = true
	log.Printf("EventHandlerRegistry: Started with %d custom handlers and projection manager", len(ehr.handlers))
	return nil
}

// Stop stops all registered handlers
func (ehr *EventHandlerRegistry) Stop(ctx context.Context) error {
	ehr.mutex.Lock()
	defer ehr.mutex.Unlock()

	if !ehr.running {
		return nil
	}

	// Stop projection manager
	if err := ehr.projManager.Stop(ctx); err != nil {
		log.Printf("EventHandlerRegistry: Warning - failed to stop projection manager: %v", err)
	}

	ehr.running = false
	log.Printf("EventHandlerRegistry: Stopped")
	return nil
}

// GetProjectionManager returns the projection manager
func (ehr *EventHandlerRegistry) GetProjectionManager() *projections.ProjectionManager {
	return ehr.projManager
}

// GetHandlers returns all registered custom handlers
func (ehr *EventHandlerRegistry) GetHandlers() map[string]cqrs.EventHandler {
	ehr.mutex.RLock()
	defer ehr.mutex.RUnlock()

	result := make(map[string]cqrs.EventHandler)
	for name, handler := range ehr.handlers {
		result[name] = handler
	}
	return result
}

// IsRunning returns whether the registry is running
func (ehr *EventHandlerRegistry) IsRunning() bool {
	ehr.mutex.RLock()
	defer ehr.mutex.RUnlock()
	return ehr.running
}

// GetStatistics returns statistics about the event handlers
func (ehr *EventHandlerRegistry) GetStatistics() map[string]interface{} {
	ehr.mutex.RLock()
	defer ehr.mutex.RUnlock()

	stats := map[string]interface{}{
		"custom_handlers":    len(ehr.handlers),
		"running":            ehr.running,
		"projection_manager": ehr.projManager.GetStatistics(),
	}

	return stats
}

// DashboardUpdateHandler handles events to update dashboard views
type DashboardUpdateHandler struct {
	readStore cqrs.ReadStore
	name      string
}

// NewDashboardUpdateHandler creates a new DashboardUpdateHandler
func NewDashboardUpdateHandler(readStore cqrs.ReadStore) *DashboardUpdateHandler {
	return &DashboardUpdateHandler{
		readStore: readStore,
		name:      "DashboardUpdateHandler",
	}
}

// Handle handles events to update dashboard
func (h *DashboardUpdateHandler) Handle(ctx context.Context, event cqrs.EventMessage) error {
	// For now, just log the event
	// In a real implementation, this would update dashboard metrics
	log.Printf("DashboardUpdateHandler: Processing event %s for dashboard update", event.EventType())

	// TODO: Implement dashboard update logic
	// This would typically:
	// 1. Check if dashboard data needs refresh
	// 2. Aggregate data from various read models
	// 3. Update dashboard view with new metrics

	return nil
}

// CanHandle checks if the handler can handle the event type
func (h *DashboardUpdateHandler) CanHandle(eventType string) bool {
	// Dashboard handler is interested in all events that might affect metrics
	supportedEvents := []string{
		"UserCreated",
		"OrderCreated",
		"OrderCompleted",
		"OrderCancelled",
		"ProductCreated",
	}

	for _, supported := range supportedEvents {
		if supported == eventType {
			return true
		}
	}
	return false
}

// GetHandlerName returns the handler name
func (h *DashboardUpdateHandler) GetHandlerName() string {
	return h.name
}

// GetHandlerType returns the handler type
func (h *DashboardUpdateHandler) GetHandlerType() cqrs.HandlerType {
	return cqrs.ProjectionHandler
}

// MetricsCollectionHandler collects metrics from events
type MetricsCollectionHandler struct {
	metrics map[string]int64
	mutex   sync.RWMutex
	name    string
}

// NewMetricsCollectionHandler creates a new MetricsCollectionHandler
func NewMetricsCollectionHandler() *MetricsCollectionHandler {
	return &MetricsCollectionHandler{
		metrics: make(map[string]int64),
		name:    "MetricsCollectionHandler",
	}
}

// Handle handles events to collect metrics
func (h *MetricsCollectionHandler) Handle(ctx context.Context, event cqrs.EventMessage) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	eventType := event.EventType()
	h.metrics[eventType]++
	h.metrics["total_events"]++

	log.Printf("MetricsCollectionHandler: Collected metric for event type %s (total: %d)",
		eventType, h.metrics[eventType])

	return nil
}

// CanHandle checks if the handler can handle the event type
func (h *MetricsCollectionHandler) CanHandle(eventType string) bool {
	// Metrics handler is interested in all events
	return true
}

// GetHandlerName returns the handler name
func (h *MetricsCollectionHandler) GetHandlerName() string {
	return h.name
}

// GetHandlerType returns the handler type
func (h *MetricsCollectionHandler) GetHandlerType() cqrs.HandlerType {
	return cqrs.ProjectionHandler
}

// GetMetrics returns collected metrics
func (h *MetricsCollectionHandler) GetMetrics() map[string]int64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	result := make(map[string]int64)
	for key, value := range h.metrics {
		result[key] = value
	}
	return result
}

// ResetMetrics resets all collected metrics
func (h *MetricsCollectionHandler) ResetMetrics() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.metrics = make(map[string]int64)
	log.Printf("MetricsCollectionHandler: Reset all metrics")
}

// AuditLogHandler logs all events for audit purposes
type AuditLogHandler struct {
	name string
}

// NewAuditLogHandler creates a new AuditLogHandler
func NewAuditLogHandler() *AuditLogHandler {
	return &AuditLogHandler{
		name: "AuditLogHandler",
	}
}

// Handle handles events for audit logging
func (h *AuditLogHandler) Handle(ctx context.Context, event cqrs.EventMessage) error {
	// Log event details for audit
	log.Printf("AUDIT: Event %s | Type: %s | Aggregate: %s (%s) | Version: %d | Timestamp: %s",
		event.EventID(),
		event.EventType(),
		event.ID(),
		event.Type(),
		event.Version(),
		event.Timestamp().Format(time.RFC3339),
	)

	return nil
}

// CanHandle checks if the handler can handle the event type
func (h *AuditLogHandler) CanHandle(eventType string) bool {
	// Audit handler logs all events
	return true
}

// GetHandlerName returns the handler name
func (h *AuditLogHandler) GetHandlerName() string {
	return h.name
}

// GetHandlerType returns the handler type
func (h *AuditLogHandler) GetHandlerType() cqrs.HandlerType {
	return cqrs.ProjectionHandler
}

// SetupDefaultHandlers sets up default event handlers
func (ehr *EventHandlerRegistry) SetupDefaultHandlers() error {
	// Register dashboard update handler
	dashboardHandler := NewDashboardUpdateHandler(ehr.readStore)
	if err := ehr.RegisterCustomHandler("dashboard", dashboardHandler); err != nil {
		return fmt.Errorf("failed to register dashboard handler: %w", err)
	}

	// Register metrics collection handler
	metricsHandler := NewMetricsCollectionHandler()
	if err := ehr.RegisterCustomHandler("metrics", metricsHandler); err != nil {
		return fmt.Errorf("failed to register metrics handler: %w", err)
	}

	// Register audit log handler
	auditHandler := NewAuditLogHandler()
	if err := ehr.RegisterCustomHandler("audit", auditHandler); err != nil {
		return fmt.Errorf("failed to register audit handler: %w", err)
	}

	log.Printf("EventHandlerRegistry: Set up %d default handlers", 3)
	return nil
}

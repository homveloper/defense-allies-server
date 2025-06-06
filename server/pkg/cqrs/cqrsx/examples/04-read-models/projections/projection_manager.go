package projections

import (
	"context"
	"defense-allies-server/pkg/cqrs"
	"fmt"
	"log"
	"sync"
	"time"
)

// ProjectionHandler defines the interface for projection handlers
type ProjectionHandler interface {
	// GetName returns the projection name
	GetName() string

	// Handle handles a single event
	Handle(ctx context.Context, event cqrs.EventMessage) error

	// GetSupportedEvents returns the list of events this projection handles
	GetSupportedEvents() []string

	// IsEventSupported checks if the projection supports a specific event type
	IsEventSupported(eventType string) bool

	// Rebuild rebuilds the projection from events
	Rebuild(ctx context.Context, eventStore interface{}) error

	// Validate validates the projection state
	Validate() error
}

// ProjectionStatus represents the status of a projection
type ProjectionStatus string

const (
	ProjectionStatusStopped    ProjectionStatus = "stopped"
	ProjectionStatusRunning    ProjectionStatus = "running"
	ProjectionStatusError      ProjectionStatus = "error"
	ProjectionStatusRebuilding ProjectionStatus = "rebuilding"
)

// ProjectionInfo contains information about a projection
type ProjectionInfo struct {
	Name            string           `json:"name"`
	Status          ProjectionStatus `json:"status"`
	LastProcessedAt time.Time        `json:"last_processed_at"`
	EventsProcessed int64            `json:"events_processed"`
	ErrorCount      int64            `json:"error_count"`
	LastError       string           `json:"last_error,omitempty"`
	SupportedEvents []string         `json:"supported_events"`
}

// ProjectionManager manages multiple projections
type ProjectionManager struct {
	projections map[string]ProjectionHandler
	eventBus    cqrs.EventBus
	readStore   cqrs.ReadStore
	info        map[string]*ProjectionInfo
	mutex       sync.RWMutex
	running     bool
	stopChan    chan struct{}
}

// NewProjectionManager creates a new ProjectionManager
func NewProjectionManager(eventBus cqrs.EventBus, readStore cqrs.ReadStore) *ProjectionManager {
	return &ProjectionManager{
		projections: make(map[string]ProjectionHandler),
		eventBus:    eventBus,
		readStore:   readStore,
		info:        make(map[string]*ProjectionInfo),
		stopChan:    make(chan struct{}),
	}
}

// RegisterProjection registers a projection handler
func (pm *ProjectionManager) RegisterProjection(projection ProjectionHandler) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if err := projection.Validate(); err != nil {
		return fmt.Errorf("projection validation failed: %w", err)
	}

	name := projection.GetName()
	if _, exists := pm.projections[name]; exists {
		return fmt.Errorf("projection %s already registered", name)
	}

	pm.projections[name] = projection
	pm.info[name] = &ProjectionInfo{
		Name:            name,
		Status:          ProjectionStatusStopped,
		SupportedEvents: projection.GetSupportedEvents(),
	}

	log.Printf("ProjectionManager: Registered projection %s", name)
	return nil
}

// UnregisterProjection unregisters a projection handler
func (pm *ProjectionManager) UnregisterProjection(name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if _, exists := pm.projections[name]; !exists {
		return fmt.Errorf("projection %s not found", name)
	}

	delete(pm.projections, name)
	delete(pm.info, name)

	log.Printf("ProjectionManager: Unregistered projection %s", name)
	return nil
}

// Start starts the projection manager and all registered projections
func (pm *ProjectionManager) Start(ctx context.Context) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if pm.running {
		return fmt.Errorf("projection manager is already running")
	}

	// Subscribe to all events
	if pm.eventBus != nil {
		subscriptionID, err := pm.eventBus.SubscribeAll(pm)
		if err != nil {
			return fmt.Errorf("failed to subscribe to event bus: %w", err)
		}
		log.Printf("ProjectionManager: Subscribed to event bus with ID %s", subscriptionID)
	}

	// Update all projections status to running
	for name, info := range pm.info {
		info.Status = ProjectionStatusRunning
		log.Printf("ProjectionManager: Started projection %s", name)
	}

	pm.running = true
	log.Printf("ProjectionManager: Started with %d projections", len(pm.projections))
	return nil
}

// Stop stops the projection manager and all projections
func (pm *ProjectionManager) Stop(ctx context.Context) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if !pm.running {
		return nil
	}

	// Signal stop
	close(pm.stopChan)

	// Update all projections status to stopped
	for name, info := range pm.info {
		info.Status = ProjectionStatusStopped
		log.Printf("ProjectionManager: Stopped projection %s", name)
	}

	pm.running = false
	log.Printf("ProjectionManager: Stopped")
	return nil
}

// Handle handles events from the event bus (implements EventHandler interface)
func (pm *ProjectionManager) Handle(ctx context.Context, event cqrs.EventMessage) error {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	if !pm.running {
		return nil
	}

	eventType := event.EventType()
	log.Printf("ProjectionManager: Processing event %s (type: %s)", event.EventID(), eventType)

	// Process event with all interested projections
	for name, projection := range pm.projections {
		if projection.IsEventSupported(eventType) {
			info := pm.info[name]

			if err := pm.processEventWithProjection(ctx, projection, event, info); err != nil {
				log.Printf("ProjectionManager: Error processing event %s with projection %s: %v",
					event.EventID(), name, err)
				// Continue processing with other projections
			}
		}
	}

	return nil
}

// processEventWithProjection processes an event with a specific projection
func (pm *ProjectionManager) processEventWithProjection(ctx context.Context, projection ProjectionHandler, event cqrs.EventMessage, info *ProjectionInfo) error {
	start := time.Now()

	defer func() {
		info.LastProcessedAt = time.Now()
		info.EventsProcessed++
	}()

	if err := projection.Handle(ctx, event); err != nil {
		info.ErrorCount++
		info.LastError = err.Error()
		info.Status = ProjectionStatusError
		return err
	}

	// Reset error status if successful
	if info.Status == ProjectionStatusError {
		info.Status = ProjectionStatusRunning
		info.LastError = ""
	}

	log.Printf("ProjectionManager: Processed event %s with projection %s in %v",
		event.EventID(), projection.GetName(), time.Since(start))
	return nil
}

// CanHandle checks if any projection can handle the event type
func (pm *ProjectionManager) CanHandle(eventType string) bool {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	for _, projection := range pm.projections {
		if projection.IsEventSupported(eventType) {
			return true
		}
	}
	return false
}

// GetHandlerName returns the handler name
func (pm *ProjectionManager) GetHandlerName() string {
	return "ProjectionManager"
}

// GetHandlerType returns the handler type
func (pm *ProjectionManager) GetHandlerType() cqrs.HandlerType {
	return cqrs.ProjectionHandler
}

// RebuildProjection rebuilds a specific projection
func (pm *ProjectionManager) RebuildProjection(ctx context.Context, projectionName string, eventStore interface{}) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	projection, exists := pm.projections[projectionName]
	if !exists {
		return fmt.Errorf("projection %s not found", projectionName)
	}

	info := pm.info[projectionName]
	info.Status = ProjectionStatusRebuilding

	log.Printf("ProjectionManager: Starting rebuild for projection %s", projectionName)

	if err := projection.Rebuild(ctx, eventStore); err != nil {
		info.Status = ProjectionStatusError
		info.LastError = err.Error()
		return fmt.Errorf("failed to rebuild projection %s: %w", projectionName, err)
	}

	info.Status = ProjectionStatusRunning
	info.LastError = ""
	log.Printf("ProjectionManager: Successfully rebuilt projection %s", projectionName)
	return nil
}

// RebuildAllProjections rebuilds all projections
func (pm *ProjectionManager) RebuildAllProjections(ctx context.Context, eventStore interface{}) error {
	pm.mutex.RLock()
	projectionNames := make([]string, 0, len(pm.projections))
	for name := range pm.projections {
		projectionNames = append(projectionNames, name)
	}
	pm.mutex.RUnlock()

	log.Printf("ProjectionManager: Starting rebuild for all %d projections", len(projectionNames))

	for _, name := range projectionNames {
		if err := pm.RebuildProjection(ctx, name, eventStore); err != nil {
			log.Printf("ProjectionManager: Failed to rebuild projection %s: %v", name, err)
			return err
		}
	}

	log.Printf("ProjectionManager: Successfully rebuilt all projections")
	return nil
}

// GetProjectionInfo returns information about a specific projection
func (pm *ProjectionManager) GetProjectionInfo(projectionName string) (*ProjectionInfo, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	info, exists := pm.info[projectionName]
	if !exists {
		return nil, fmt.Errorf("projection %s not found", projectionName)
	}

	// Return a copy to avoid race conditions
	infoCopy := *info
	return &infoCopy, nil
}

// GetAllProjectionInfo returns information about all projections
func (pm *ProjectionManager) GetAllProjectionInfo() map[string]*ProjectionInfo {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	result := make(map[string]*ProjectionInfo)
	for name, info := range pm.info {
		// Return copies to avoid race conditions
		infoCopy := *info
		result[name] = &infoCopy
	}

	return result
}

// GetProjectionNames returns the names of all registered projections
func (pm *ProjectionManager) GetProjectionNames() []string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	names := make([]string, 0, len(pm.projections))
	for name := range pm.projections {
		names = append(names, name)
	}

	return names
}

// IsRunning returns whether the projection manager is running
func (pm *ProjectionManager) IsRunning() bool {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return pm.running
}

// GetStatistics returns overall statistics for all projections
func (pm *ProjectionManager) GetStatistics() map[string]interface{} {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var totalEvents, totalErrors int64
	runningCount := 0
	errorCount := 0

	for _, info := range pm.info {
		totalEvents += info.EventsProcessed
		totalErrors += info.ErrorCount

		switch info.Status {
		case ProjectionStatusRunning:
			runningCount++
		case ProjectionStatusError:
			errorCount++
		}
	}

	return map[string]interface{}{
		"total_projections":   len(pm.projections),
		"running_projections": runningCount,
		"error_projections":   errorCount,
		"total_events":        totalEvents,
		"total_errors":        totalErrors,
		"manager_running":     pm.running,
	}
}

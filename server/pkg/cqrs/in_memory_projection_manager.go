package cqrs

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// InMemoryProjectionManager provides an in-memory implementation of ProjectionManager
type InMemoryProjectionManager struct {
	projections map[string]Projection
	metrics     *ProjectionMetrics
	running     bool
	mutex       sync.RWMutex
}

// NewInMemoryProjectionManager creates a new in-memory projection manager
func NewInMemoryProjectionManager() *InMemoryProjectionManager {
	return &InMemoryProjectionManager{
		projections: make(map[string]Projection),
		metrics: &ProjectionMetrics{
			TotalProjections:      0,
			RunningProjections:    0,
			FaultedProjections:    0,
			ProcessedEvents:       0,
			AverageProcessingTime: 0,
			LastProcessedEvent:    time.Time{},
			Errors:                make([]ProjectionError, 0),
		},
		running: false,
	}
}

// ProjectionManager interface implementation

func (pm *InMemoryProjectionManager) RegisterProjection(projection Projection) error {
	if projection == nil {
		return NewCQRSError(ErrCodeEventValidation.String(), "projection cannot be nil", nil)
	}

	name := projection.GetProjectionName()
	if name == "" {
		return NewCQRSError(ErrCodeEventValidation.String(), "projection name cannot be empty", nil)
	}

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if _, exists := pm.projections[name]; exists {
		return NewCQRSError(ErrCodeEventValidation.String(), fmt.Sprintf("projection already registered: %s", name), nil)
	}

	pm.projections[name] = projection
	pm.metrics.TotalProjections++

	if projection.GetState() == ProjectionRunning {
		pm.metrics.RunningProjections++
	} else if projection.GetState() == ProjectionFaulted {
		pm.metrics.FaultedProjections++
	}

	return nil
}

func (pm *InMemoryProjectionManager) UnregisterProjection(projectionName string) error {
	if projectionName == "" {
		return NewCQRSError(ErrCodeEventValidation.String(), "projection name cannot be empty", nil)
	}

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	projection, exists := pm.projections[projectionName]
	if !exists {
		return NewCQRSError(ErrCodeEventValidation.String(), fmt.Sprintf("projection not found: %s", projectionName), nil)
	}

	delete(pm.projections, projectionName)
	pm.metrics.TotalProjections--

	if projection.GetState() == ProjectionRunning {
		pm.metrics.RunningProjections--
	} else if projection.GetState() == ProjectionFaulted {
		pm.metrics.FaultedProjections--
	}

	return nil
}

func (pm *InMemoryProjectionManager) Start(ctx context.Context) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if pm.running {
		return NewCQRSError(ErrCodeEventValidation.String(), "projection manager is already running", nil)
	}

	pm.running = true

	// Start all projections
	for _, projection := range pm.projections {
		oldState := projection.GetState()

		// Try to set state through different projection types
		if baseProjection, ok := projection.(*BaseProjection); ok {
			baseProjection.SetState(ProjectionRunning)
			pm.updateStateCounters(oldState, ProjectionRunning)
		} else {
			// For embedded BaseProjection (like TestProjection)
			projection.Rebuild(context.Background()) // This sets state to running
			pm.updateStateCounters(oldState, projection.GetState())
		}
	}

	return nil
}

func (pm *InMemoryProjectionManager) Stop(ctx context.Context) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if !pm.running {
		return NewCQRSError(ErrCodeEventValidation.String(), "projection manager is not running", nil)
	}

	pm.running = false

	// Stop all projections
	for _, projection := range pm.projections {
		if projection.GetState() == ProjectionRunning {
			oldState := projection.GetState()

			// Try to set state through different projection types
			if baseProjection, ok := projection.(*BaseProjection); ok {
				baseProjection.SetState(ProjectionStopped)
				pm.updateStateCounters(oldState, ProjectionStopped)
			} else {
				// For embedded BaseProjection (like TestProjection)
				projection.Reset(context.Background()) // This sets state to stopped
				pm.updateStateCounters(oldState, projection.GetState())
			}
		}
	}

	return nil
}

func (pm *InMemoryProjectionManager) GetProjectionState(projectionName string) (ProjectionState, error) {
	if projectionName == "" {
		return ProjectionStopped, NewCQRSError(ErrCodeEventValidation.String(), "projection name cannot be empty", nil)
	}

	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	projection, exists := pm.projections[projectionName]
	if !exists {
		return ProjectionStopped, NewCQRSError(ErrCodeEventValidation.String(), fmt.Sprintf("projection not found: %s", projectionName), nil)
	}

	return projection.GetState(), nil
}

func (pm *InMemoryProjectionManager) ResetProjection(ctx context.Context, projectionName string) error {
	if projectionName == "" {
		return NewCQRSError(ErrCodeEventValidation.String(), "projection name cannot be empty", nil)
	}

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	projection, exists := pm.projections[projectionName]
	if !exists {
		return NewCQRSError(ErrCodeEventValidation.String(), fmt.Sprintf("projection not found: %s", projectionName), nil)
	}

	// Update state counters
	oldState := projection.GetState()
	if err := projection.Reset(ctx); err != nil {
		return err
	}

	pm.updateStateCounters(oldState, projection.GetState())
	return nil
}

func (pm *InMemoryProjectionManager) RebuildProjection(ctx context.Context, projectionName string) error {
	if projectionName == "" {
		return NewCQRSError(ErrCodeEventValidation.String(), "projection name cannot be empty", nil)
	}

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	projection, exists := pm.projections[projectionName]
	if !exists {
		return NewCQRSError(ErrCodeEventValidation.String(), fmt.Sprintf("projection not found: %s", projectionName), nil)
	}

	// Update state counters
	oldState := projection.GetState()
	if err := projection.Rebuild(ctx); err != nil {
		return err
	}

	pm.updateStateCounters(oldState, projection.GetState())
	return nil
}

func (pm *InMemoryProjectionManager) GetMetrics() *ProjectionMetrics {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	// Return a copy of metrics
	errorsCopy := make([]ProjectionError, len(pm.metrics.Errors))
	copy(errorsCopy, pm.metrics.Errors)

	return &ProjectionMetrics{
		TotalProjections:      pm.metrics.TotalProjections,
		RunningProjections:    pm.metrics.RunningProjections,
		FaultedProjections:    pm.metrics.FaultedProjections,
		ProcessedEvents:       pm.metrics.ProcessedEvents,
		AverageProcessingTime: pm.metrics.AverageProcessingTime,
		LastProcessedEvent:    pm.metrics.LastProcessedEvent,
		Errors:                errorsCopy,
	}
}

// Helper methods

func (pm *InMemoryProjectionManager) updateStateCounters(oldState, newState ProjectionState) {
	// Decrement old state counter
	switch oldState {
	case ProjectionRunning:
		pm.metrics.RunningProjections--
	case ProjectionFaulted:
		pm.metrics.FaultedProjections--
	}

	// Increment new state counter
	switch newState {
	case ProjectionRunning:
		pm.metrics.RunningProjections++
	case ProjectionFaulted:
		pm.metrics.FaultedProjections++
	}
}

// ProcessEvent processes an event through all registered projections
func (pm *InMemoryProjectionManager) ProcessEvent(ctx context.Context, event EventMessage) error {
	if event == nil {
		return NewCQRSError(ErrCodeEventValidation.String(), "event cannot be nil", nil)
	}

	pm.mutex.RLock()
	projections := make([]Projection, 0, len(pm.projections))
	for _, projection := range pm.projections {
		if projection.CanHandle(event.EventType()) && projection.GetState() == ProjectionRunning {
			projections = append(projections, projection)
		}
	}
	pm.mutex.RUnlock()

	start := time.Now()

	for _, projection := range projections {
		if err := projection.Project(ctx, event); err != nil {
			// Record error
			projectionError := ProjectionError{
				ProjectionName: projection.GetProjectionName(),
				EventID:        event.EventID(),
				EventType:      event.EventType(),
				Error:          err,
				Timestamp:      time.Now(),
				RetryCount:     0,
			}

			pm.mutex.Lock()
			pm.metrics.Errors = append(pm.metrics.Errors, projectionError)

			// Mark projection as faulted
			if projection.GetState() == ProjectionRunning {
				oldState := projection.GetState()

				if baseProjection, ok := projection.(*BaseProjection); ok {
					baseProjection.SetState(ProjectionFaulted)
				}
				// For other projection types, we can't directly set to faulted
				// but we can track the state change in metrics
				pm.updateStateCounters(oldState, ProjectionFaulted)
			}
			pm.mutex.Unlock()

			return err
		}
	}

	// Update metrics
	pm.mutex.Lock()
	pm.metrics.ProcessedEvents++
	pm.metrics.LastProcessedEvent = time.Now()

	processingTime := time.Since(start)
	if pm.metrics.ProcessedEvents == 1 {
		pm.metrics.AverageProcessingTime = processingTime
	} else {
		pm.metrics.AverageProcessingTime = (pm.metrics.AverageProcessingTime + processingTime) / 2
	}
	pm.mutex.Unlock()

	return nil
}

// GetProjection returns a projection by name
func (pm *InMemoryProjectionManager) GetProjection(projectionName string) (Projection, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	projection, exists := pm.projections[projectionName]
	return projection, exists
}

// GetAllProjections returns all registered projections
func (pm *InMemoryProjectionManager) GetAllProjections() map[string]Projection {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	projections := make(map[string]Projection)
	for name, projection := range pm.projections {
		projections[name] = projection
	}
	return projections
}

// IsRunning returns whether the projection manager is running
func (pm *InMemoryProjectionManager) IsRunning() bool {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return pm.running
}

// Clear removes all projections and resets metrics
func (pm *InMemoryProjectionManager) Clear() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.projections = make(map[string]Projection)
	pm.metrics = &ProjectionMetrics{
		TotalProjections:      0,
		RunningProjections:    0,
		FaultedProjections:    0,
		ProcessedEvents:       0,
		AverageProcessingTime: 0,
		LastProcessedEvent:    time.Time{},
		Errors:                make([]ProjectionError, 0),
	}
	pm.running = false
}

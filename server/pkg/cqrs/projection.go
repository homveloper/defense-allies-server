package cqrs

import (
	"context"
	"time"
)

// ProjectionState represents the state of a projection
type ProjectionState int

const (
	ProjectionStopped ProjectionState = iota
	ProjectionRunning
	ProjectionCatchingUp
	ProjectionFaulted
	ProjectionRebuilding
)

func (ps ProjectionState) String() string {
	switch ps {
	case ProjectionStopped:
		return "stopped"
	case ProjectionRunning:
		return "running"
	case ProjectionCatchingUp:
		return "catching_up"
	case ProjectionFaulted:
		return "faulted"
	case ProjectionRebuilding:
		return "rebuilding"
	default:
		return "unknown"
	}
}

// ProjectionError represents an error that occurred during projection processing
type ProjectionError struct {
	ProjectionName string
	EventID        string
	EventType      string
	Error          error
	Timestamp      time.Time
	RetryCount     int
}

// ProjectionMetrics represents projection performance metrics
type ProjectionMetrics struct {
	TotalProjections      int
	RunningProjections    int
	FaultedProjections    int
	ProcessedEvents       int64
	AverageProcessingTime time.Duration
	LastProcessedEvent    time.Time
	Errors                []ProjectionError
}

// Projection interface for event projections
type Projection interface {
	// Projection information
	GetProjectionName() string
	GetVersion() string
	GetLastProcessedEvent() string

	// Event processing
	CanHandle(eventType string) bool
	Project(ctx context.Context, event EventMessage) error

	// State management
	GetState() ProjectionState
	Reset(ctx context.Context) error
	Rebuild(ctx context.Context) error
}

// ProjectionManager interface for managing projections
type ProjectionManager interface {
	// Projection management
	RegisterProjection(projection Projection) error
	UnregisterProjection(projectionName string) error

	// Lifecycle
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// State management
	GetProjectionState(projectionName string) (ProjectionState, error)
	ResetProjection(ctx context.Context, projectionName string) error
	RebuildProjection(ctx context.Context, projectionName string) error

	// Monitoring
	GetMetrics() *ProjectionMetrics
}

// ReadModel interface for query-optimized data models
type ReadModel interface {
	GetID() string
	GetType() string
	GetVersion() int
	GetData() interface{}
	GetLastUpdated() time.Time

	// Validation
	Validate() error
}

// ReadStore interface for read model storage
type ReadStore interface {
	// Basic CRUD
	Save(ctx context.Context, readModel ReadModel) error
	GetByID(ctx context.Context, id string, modelType string) (ReadModel, error)
	Delete(ctx context.Context, id string, modelType string) error

	// Query
	Query(ctx context.Context, criteria QueryCriteria) ([]ReadModel, error)
	Count(ctx context.Context, criteria QueryCriteria) (int64, error)

	// Batch operations
	SaveBatch(ctx context.Context, readModels []ReadModel) error
	DeleteBatch(ctx context.Context, ids []string, modelType string) error

	// Index management
	CreateIndex(ctx context.Context, modelType string, fields []string) error
	DropIndex(ctx context.Context, modelType string, indexName string) error
}

// BaseProjection provides a base implementation of Projection interface
type BaseProjection struct {
	name                string
	version             string
	lastProcessedEvent  string
	state               ProjectionState
	supportedEventTypes map[string]bool
}

// NewBaseProjection creates a new BaseProjection
func NewBaseProjection(name, version string, eventTypes []string) *BaseProjection {
	typeMap := make(map[string]bool)
	for _, eventType := range eventTypes {
		typeMap[eventType] = true
	}

	return &BaseProjection{
		name:                name,
		version:             version,
		state:               ProjectionStopped,
		supportedEventTypes: typeMap,
	}
}

// Projection interface implementation

func (p *BaseProjection) GetProjectionName() string {
	return p.name
}

func (p *BaseProjection) GetVersion() string {
	return p.version
}

func (p *BaseProjection) GetLastProcessedEvent() string {
	return p.lastProcessedEvent
}

func (p *BaseProjection) CanHandle(eventType string) bool {
	return p.supportedEventTypes[eventType]
}

func (p *BaseProjection) GetState() ProjectionState {
	return p.state
}

// Project method should be implemented by concrete projections
func (p *BaseProjection) Project(ctx context.Context, event EventMessage) error {
	// Base implementation - should be overridden
	p.lastProcessedEvent = event.EventID()
	return nil
}

func (p *BaseProjection) Reset(ctx context.Context) error {
	p.lastProcessedEvent = ""
	p.state = ProjectionStopped
	return nil
}

func (p *BaseProjection) Rebuild(ctx context.Context) error {
	p.state = ProjectionRebuilding
	// Rebuilding logic should be implemented by concrete projections
	p.state = ProjectionRunning
	return nil
}

// Helper methods

// SetState sets the projection state
func (p *BaseProjection) SetState(state ProjectionState) {
	p.state = state
}

// SetLastProcessedEvent sets the last processed event ID
func (p *BaseProjection) SetLastProcessedEvent(eventID string) {
	p.lastProcessedEvent = eventID
}

// AddEventType adds an event type that this projection can handle
func (p *BaseProjection) AddEventType(eventType string) {
	p.supportedEventTypes[eventType] = true
}

// RemoveEventType removes an event type from this projection
func (p *BaseProjection) RemoveEventType(eventType string) {
	delete(p.supportedEventTypes, eventType)
}

// GetSupportedEventTypes returns all supported event types
func (p *BaseProjection) GetSupportedEventTypes() []string {
	types := make([]string, 0, len(p.supportedEventTypes))
	for eventType := range p.supportedEventTypes {
		types = append(types, eventType)
	}
	return types
}

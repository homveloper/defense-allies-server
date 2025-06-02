package cqrs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test projection implementation
type TestProjection struct {
	*BaseProjection
	ProjectFunc func(ctx context.Context, event EventMessage) error
}

func NewTestProjection(name, version string, eventTypes []string) *TestProjection {
	return &TestProjection{
		BaseProjection: NewBaseProjection(name, version, eventTypes),
	}
}

func (p *TestProjection) Project(ctx context.Context, event EventMessage) error {
	if p.ProjectFunc != nil {
		return p.ProjectFunc(ctx, event)
	}

	// Default implementation - just update last processed event
	p.SetLastProcessedEvent(event.EventID())
	return nil
}

// Expose BaseProjection methods for testing
func (p *TestProjection) SetState(state ProjectionState) {
	p.BaseProjection.SetState(state)
}

func TestNewInMemoryProjectionManager(t *testing.T) {
	// Act
	pm := NewInMemoryProjectionManager()

	// Assert
	assert.NotNil(t, pm)
	assert.False(t, pm.IsRunning())

	metrics := pm.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, 0, metrics.TotalProjections)
	assert.Equal(t, 0, metrics.RunningProjections)
	assert.Equal(t, 0, metrics.FaultedProjections)
}

func TestProjectionManager_RegisterProjection(t *testing.T) {
	// Arrange
	pm := NewInMemoryProjectionManager()
	projection := NewTestProjection("TestProjection", "1.0", []string{"TestEvent"})

	// Act
	err := pm.RegisterProjection(projection)

	// Assert
	assert.NoError(t, err)

	metrics := pm.GetMetrics()
	assert.Equal(t, 1, metrics.TotalProjections)

	// Verify projection can be retrieved
	retrievedProjection, exists := pm.GetProjection("TestProjection")
	assert.True(t, exists)
	assert.Equal(t, projection, retrievedProjection)
}

func TestProjectionManager_RegisterProjection_NilProjection(t *testing.T) {
	// Arrange
	pm := NewInMemoryProjectionManager()

	// Act
	err := pm.RegisterProjection(nil)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "projection cannot be nil")
}

func TestProjectionManager_RegisterProjection_EmptyName(t *testing.T) {
	// Arrange
	pm := NewInMemoryProjectionManager()
	projection := NewTestProjection("", "1.0", []string{"TestEvent"})

	// Act
	err := pm.RegisterProjection(projection)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "projection name cannot be empty")
}

func TestProjectionManager_RegisterProjection_Duplicate(t *testing.T) {
	// Arrange
	pm := NewInMemoryProjectionManager()
	projection1 := NewTestProjection("TestProjection", "1.0", []string{"TestEvent"})
	projection2 := NewTestProjection("TestProjection", "2.0", []string{"TestEvent"})

	// Act
	err1 := pm.RegisterProjection(projection1)
	err2 := pm.RegisterProjection(projection2)

	// Assert
	assert.NoError(t, err1)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "projection already registered")
}

func TestProjectionManager_UnregisterProjection(t *testing.T) {
	// Arrange
	pm := NewInMemoryProjectionManager()
	projection := NewTestProjection("TestProjection", "1.0", []string{"TestEvent"})
	pm.RegisterProjection(projection)

	// Act
	err := pm.UnregisterProjection("TestProjection")

	// Assert
	assert.NoError(t, err)

	metrics := pm.GetMetrics()
	assert.Equal(t, 0, metrics.TotalProjections)

	// Verify projection is removed
	_, exists := pm.GetProjection("TestProjection")
	assert.False(t, exists)
}

func TestProjectionManager_UnregisterProjection_NotFound(t *testing.T) {
	// Arrange
	pm := NewInMemoryProjectionManager()

	// Act
	err := pm.UnregisterProjection("NonExistentProjection")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "projection not found")
}

func TestProjectionManager_StartStop(t *testing.T) {
	// Arrange
	pm := NewInMemoryProjectionManager()
	projection := NewTestProjection("TestProjection", "1.0", []string{"TestEvent"})
	pm.RegisterProjection(projection)

	// Initially not running
	assert.False(t, pm.IsRunning())
	assert.Equal(t, ProjectionStopped, projection.GetState())

	// Act - Start
	err := pm.Start(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.True(t, pm.IsRunning())
	assert.Equal(t, ProjectionRunning, projection.GetState())

	metrics := pm.GetMetrics()
	assert.Equal(t, 1, metrics.RunningProjections)

	// Act - Start again (should error)
	err = pm.Start(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Act - Stop
	err = pm.Stop(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.False(t, pm.IsRunning())
	assert.Equal(t, ProjectionStopped, projection.GetState())

	metrics = pm.GetMetrics()
	assert.Equal(t, 0, metrics.RunningProjections)
}

func TestProjectionManager_GetProjectionState(t *testing.T) {
	// Arrange
	pm := NewInMemoryProjectionManager()
	projection := NewTestProjection("TestProjection", "1.0", []string{"TestEvent"})
	pm.RegisterProjection(projection)

	// Act
	state, err := pm.GetProjectionState("TestProjection")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, ProjectionStopped, state)
}

func TestProjectionManager_GetProjectionState_NotFound(t *testing.T) {
	// Arrange
	pm := NewInMemoryProjectionManager()

	// Act
	state, err := pm.GetProjectionState("NonExistentProjection")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ProjectionStopped, state)
	assert.Contains(t, err.Error(), "projection not found")
}

func TestProjectionManager_ResetProjection(t *testing.T) {
	// Arrange
	pm := NewInMemoryProjectionManager()
	projection := NewTestProjection("TestProjection", "1.0", []string{"TestEvent"})
	projection.SetState(ProjectionRunning)
	pm.RegisterProjection(projection)

	// Act
	err := pm.ResetProjection(context.Background(), "TestProjection")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, ProjectionStopped, projection.GetState())
}

func TestProjectionManager_RebuildProjection(t *testing.T) {
	// Arrange
	pm := NewInMemoryProjectionManager()
	projection := NewTestProjection("TestProjection", "1.0", []string{"TestEvent"})
	pm.RegisterProjection(projection)

	// Act
	err := pm.RebuildProjection(context.Background(), "TestProjection")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, ProjectionRunning, projection.GetState())
}

func TestProjectionManager_ProcessEvent_Success(t *testing.T) {
	// Arrange
	pm := NewInMemoryProjectionManager()
	projection := NewTestProjection("TestProjection", "1.0", []string{"TestEvent"})
	projection.SetState(ProjectionRunning)
	pm.RegisterProjection(projection)

	event := NewBaseEventMessage("TestEvent", "test-id", "TestAggregate", 1, "test data")

	// Act
	err := pm.ProcessEvent(context.Background(), event)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, event.EventID(), projection.GetLastProcessedEvent())

	metrics := pm.GetMetrics()
	assert.Equal(t, int64(1), metrics.ProcessedEvents)
}

func TestProjectionManager_ProcessEvent_ProjectionError(t *testing.T) {
	// Arrange
	pm := NewInMemoryProjectionManager()
	projection := NewTestProjection("TestProjection", "1.0", []string{"TestEvent"})
	projection.SetState(ProjectionRunning)

	// Set up projection to return error
	projection.ProjectFunc = func(ctx context.Context, event EventMessage) error {
		return NewCQRSError(ErrCodeEventValidation.String(), "projection error", nil)
	}

	pm.RegisterProjection(projection)

	event := NewBaseEventMessage("TestEvent", "test-id", "TestAggregate", 1, "test data")

	// Act
	err := pm.ProcessEvent(context.Background(), event)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "projection error")

	// Note: TestProjection doesn't automatically change to faulted state
	// but the metrics should still track the error
	metrics := pm.GetMetrics()
	assert.Equal(t, 1, metrics.FaultedProjections)
	assert.Len(t, metrics.Errors, 1)
}

func TestProjectionManager_ProcessEvent_NilEvent(t *testing.T) {
	// Arrange
	pm := NewInMemoryProjectionManager()

	// Act
	err := pm.ProcessEvent(context.Background(), nil)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event cannot be nil")
}

func TestProjectionManager_GetAllProjections(t *testing.T) {
	// Arrange
	pm := NewInMemoryProjectionManager()
	projection1 := NewTestProjection("Projection1", "1.0", []string{"Event1"})
	projection2 := NewTestProjection("Projection2", "1.0", []string{"Event2"})

	pm.RegisterProjection(projection1)
	pm.RegisterProjection(projection2)

	// Act
	projections := pm.GetAllProjections()

	// Assert
	assert.Len(t, projections, 2)
	assert.Contains(t, projections, "Projection1")
	assert.Contains(t, projections, "Projection2")
	assert.Equal(t, projection1, projections["Projection1"])
	assert.Equal(t, projection2, projections["Projection2"])
}

func TestProjectionManager_Clear(t *testing.T) {
	// Arrange
	pm := NewInMemoryProjectionManager()
	projection := NewTestProjection("TestProjection", "1.0", []string{"TestEvent"})
	pm.RegisterProjection(projection)
	pm.Start(context.Background())

	// Verify initial state
	assert.True(t, pm.IsRunning())
	assert.Equal(t, 1, pm.GetMetrics().TotalProjections)

	// Act
	pm.Clear()

	// Assert
	assert.False(t, pm.IsRunning())
	metrics := pm.GetMetrics()
	assert.Equal(t, 0, metrics.TotalProjections)
	assert.Equal(t, 0, metrics.RunningProjections)
	assert.Equal(t, 0, metrics.FaultedProjections)
}

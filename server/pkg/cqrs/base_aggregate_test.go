package cqrs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBaseAggregate(t *testing.T) {
	// Arrange
	id := "test-id"
	aggregateType := "TestAggregate"

	// Act
	aggregate := NewBaseAggregate(id, aggregateType)

	// Assert
	assert.NotNil(t, aggregate)
	assert.Equal(t, id, aggregate.ID())
	assert.Equal(t, aggregateType, aggregate.Type())
	assert.Equal(t, 0, aggregate.OriginalVersion())
	assert.Equal(t, 0, aggregate.Version())
	assert.Empty(t, aggregate.Changes())
	assert.NotZero(t, aggregate.createdAt)
	assert.NotZero(t, aggregate.updatedAt)
	assert.False(t, aggregate.deleted)
}

func TestNewBaseAggregateWithOptions(t *testing.T) {
	// Arrange
	id := "test-id"
	aggregateType := "TestAggregate"
	originalVersion := 5
	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now().Add(-1 * time.Hour)
	deleted := true

	// Act
	aggregate := NewBaseAggregate(id, aggregateType,
		WithOriginalVersion(originalVersion),
		WithCreatedAt(createdAt),
		WithUpdatedAt(updatedAt),
		WithDeleted(deleted),
	)

	// Assert
	assert.NotNil(t, aggregate)
	assert.Equal(t, id, aggregate.ID())
	assert.Equal(t, aggregateType, aggregate.Type())
	assert.Equal(t, originalVersion, aggregate.OriginalVersion())
	assert.Equal(t, originalVersion, aggregate.Version()) // Version should match OriginalVersion initially
	assert.Equal(t, createdAt, aggregate.createdAt)
	assert.Equal(t, updatedAt, aggregate.updatedAt)
	assert.Equal(t, deleted, aggregate.deleted)
	assert.Empty(t, aggregate.Changes())
}

func TestBaseAggregate_ApplyEvent(t *testing.T) {
	// Arrange
	aggregate := NewBaseAggregate("test-id", "TestAggregate")
	event := NewBaseEventMessage("TestEvent", "test data")
	initialVersion := aggregate.Version()
	initialUpdatedAt := aggregate.updatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)

	// Act - Apply new event
	err := aggregate.ApplyEvent(event)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, initialVersion+1, aggregate.Version())
	changes := aggregate.Changes()
	assert.Len(t, changes, 1)
	assert.Equal(t, event, changes[0])
	assert.True(t, aggregate.updatedAt.After(initialUpdatedAt))
}

func TestBaseAggregate_ReplayEvent(t *testing.T) {
	// Arrange
	aggregate := NewBaseAggregate("test-id", "TestAggregate")
	event := NewBaseEventMessage("TestEvent", "test data")
	initialVersion := aggregate.Version()
	initialUpdatedAt := aggregate.updatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)

	// Act - Replay existing event (not new)
	err := aggregate.ReplayEvent(event)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, initialVersion+1, aggregate.Version())
	changes := aggregate.Changes()
	assert.Empty(t, changes) // Should not track replayed events
	assert.True(t, aggregate.updatedAt.After(initialUpdatedAt))
}

func TestBaseAggregate_ClearChanges(t *testing.T) {
	// Arrange
	aggregate := NewBaseAggregate("test-id", "TestAggregate")
	event := NewBaseEventMessage("TestEvent", "test data")

	// Apply event to track it
	err := aggregate.ApplyEvent(event)
	assert.NoError(t, err)

	// Verify event is tracked
	assert.Len(t, aggregate.Changes(), 1)

	// Act
	aggregate.ClearChanges()

	// Assert
	assert.Empty(t, aggregate.Changes())
}

func TestBaseAggregate_ApplyEventWithNilEvent(t *testing.T) {
	// Arrange
	aggregate := NewBaseAggregate("test-id", "TestAggregate")

	// Act
	err := aggregate.ApplyEvent(nil)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event cannot be nil")
}

func TestBaseAggregate_ReplayEventWithNilEvent(t *testing.T) {
	// Arrange
	aggregate := NewBaseAggregate("test-id", "TestAggregate")

	// Act
	err := aggregate.ReplayEvent(nil)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event cannot be nil")
}

func TestBaseAggregate_Validate(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		aggregateType string
		expectError   bool
	}{
		{
			name:          "Valid aggregate",
			id:            "test-id",
			aggregateType: "TestAggregate",
			expectError:   false,
		},
		{
			name:          "Empty ID",
			id:            "",
			aggregateType: "TestAggregate",
			expectError:   true,
		},
		{
			name:          "Empty aggregate type",
			id:            "test-id",
			aggregateType: "",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			aggregate := NewBaseAggregate(tt.id, tt.aggregateType)

			// Act
			err := aggregate.Validate()

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBaseAggregate_HasUncommittedChanges(t *testing.T) {
	// Arrange
	aggregate := NewBaseAggregate("test-id", "TestAggregate")

	// Initially no changes
	assert.False(t, aggregate.HasUncommittedChanges())

	// Add a change by applying an event
	event := NewBaseEventMessage("TestEvent", "test data")
	err := aggregate.ApplyEvent(event)
	assert.NoError(t, err)

	// Should have changes now
	assert.True(t, aggregate.HasUncommittedChanges())
	assert.Equal(t, 1, aggregate.GetUncommittedChangeCount())

	// Clear changes
	aggregate.ClearChanges()

	// Should not have changes
	assert.False(t, aggregate.HasUncommittedChanges())
	assert.Equal(t, 0, aggregate.GetUncommittedChangeCount())
}

func TestBaseAggregate_LoadFromHistory(t *testing.T) {
	// Arrange
	aggregate := NewBaseAggregate("test-id", "TestAggregate")
	events := []EventMessage{
		NewBaseEventMessage("Event1", "data1"),
		NewBaseEventMessage("Event2", "data2"),
		NewBaseEventMessage("Event3", "data3"),
	}

	// Act
	err := aggregate.LoadFromHistory(events)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 3, aggregate.Version()) // Version should be incremented for each event
	assert.Empty(t, aggregate.Changes())    // LoadFromHistory should not track changes
}

func TestBaseAggregate_JSONSerialization(t *testing.T) {
	// Arrange - Create aggregate without events to avoid JSON serialization issues
	originalAggregate := NewBaseAggregate("test-id", "TestAggregate",
		WithOriginalVersion(5),
		WithCreatedAt(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)),
		WithUpdatedAt(time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC)),
		WithDeleted(true),
	)

	// Act - Marshal to JSON
	jsonData, err := originalAggregate.MarshalJSON()
	assert.NoError(t, err)

	// Act - Unmarshal from JSON
	newAggregate := &BaseAggregate{}
	err = newAggregate.UnmarshalJSON(jsonData)
	assert.NoError(t, err)

	// Assert
	assert.Equal(t, originalAggregate.ID(), newAggregate.ID())
	assert.Equal(t, originalAggregate.Type(), newAggregate.Type())
	assert.Equal(t, originalAggregate.OriginalVersion(), newAggregate.OriginalVersion())
	assert.Equal(t, originalAggregate.Version(), newAggregate.Version())
	assert.Equal(t, originalAggregate.createdAt, newAggregate.createdAt)
	assert.Equal(t, originalAggregate.updatedAt, newAggregate.updatedAt)
	assert.Equal(t, originalAggregate.deleted, newAggregate.deleted)
	assert.Empty(t, newAggregate.Changes()) // No events added
}

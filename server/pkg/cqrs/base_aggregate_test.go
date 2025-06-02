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
	assert.Equal(t, id, aggregate.AggregateID())
	assert.Equal(t, aggregateType, aggregate.AggregateType())
	assert.Equal(t, 0, aggregate.OriginalVersion())
	assert.Equal(t, 0, aggregate.CurrentVersion())
	assert.False(t, aggregate.IsDeleted())
	assert.Empty(t, aggregate.GetChanges())
	assert.NotZero(t, aggregate.CreatedAt())
	assert.NotZero(t, aggregate.UpdatedAt())
}

func TestBaseAggregate_IncrementVersion(t *testing.T) {
	// Arrange
	aggregate := NewBaseAggregate("test-id", "TestAggregate")
	initialVersion := aggregate.CurrentVersion()
	initialUpdatedAt := aggregate.UpdatedAt()

	// Wait a bit to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)

	// Act
	aggregate.IncrementVersion()

	// Assert
	assert.Equal(t, initialVersion+1, aggregate.CurrentVersion())
	assert.True(t, aggregate.UpdatedAt().After(initialUpdatedAt))
}

func TestBaseAggregate_TrackChange(t *testing.T) {
	// Arrange
	aggregate := NewBaseAggregate("test-id", "TestAggregate")
	event := NewBaseEventMessage("TestEvent", "test-id", "TestAggregate", 1, "test data")

	// Act
	aggregate.TrackChange(event)

	// Assert
	changes := aggregate.GetChanges()
	assert.Len(t, changes, 1)
	assert.Equal(t, event, changes[0])
}

func TestBaseAggregate_Apply(t *testing.T) {
	// Arrange
	aggregate := NewBaseAggregate("test-id", "TestAggregate")
	event := NewBaseEventMessage("TestEvent", "test-id", "TestAggregate", 1, "test data")
	initialVersion := aggregate.CurrentVersion()

	// Act - Apply new event
	aggregate.Apply(event, true)

	// Assert
	assert.Equal(t, initialVersion+1, aggregate.CurrentVersion())
	changes := aggregate.GetChanges()
	assert.Len(t, changes, 1)
	assert.Equal(t, event, changes[0])
}

func TestBaseAggregate_Apply_ExistingEvent(t *testing.T) {
	// Arrange
	aggregate := NewBaseAggregate("test-id", "TestAggregate")
	event := NewBaseEventMessage("TestEvent", "test-id", "TestAggregate", 1, "test data")
	initialVersion := aggregate.CurrentVersion()

	// Act - Apply existing event (not new)
	aggregate.Apply(event, false)

	// Assert
	assert.Equal(t, initialVersion+1, aggregate.CurrentVersion())
	changes := aggregate.GetChanges()
	assert.Empty(t, changes) // Should not track existing events
}

func TestBaseAggregate_ClearChanges(t *testing.T) {
	// Arrange
	aggregate := NewBaseAggregate("test-id", "TestAggregate")
	event := NewBaseEventMessage("TestEvent", "test-id", "TestAggregate", 1, "test data")
	aggregate.TrackChange(event)

	// Verify event is tracked
	assert.Len(t, aggregate.GetChanges(), 1)

	// Act
	aggregate.ClearChanges()

	// Assert
	assert.Empty(t, aggregate.GetChanges())
}

func TestBaseAggregate_MarkAsDeleted(t *testing.T) {
	// Arrange
	aggregate := NewBaseAggregate("test-id", "TestAggregate")
	initialUpdatedAt := aggregate.UpdatedAt()

	// Wait a bit to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)

	// Act
	aggregate.MarkAsDeleted()

	// Assert
	assert.True(t, aggregate.IsDeleted())
	assert.True(t, aggregate.UpdatedAt().After(initialUpdatedAt))
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

func TestBaseAggregate_SetOriginalVersion(t *testing.T) {
	// Arrange
	aggregate := NewBaseAggregate("test-id", "TestAggregate")
	version := 5

	// Act
	aggregate.SetOriginalVersion(version)

	// Assert
	assert.Equal(t, version, aggregate.OriginalVersion())
	assert.Equal(t, version, aggregate.CurrentVersion())
}

func TestBaseAggregate_HasUncommittedChanges(t *testing.T) {
	// Arrange
	aggregate := NewBaseAggregate("test-id", "TestAggregate")

	// Initially no changes
	assert.False(t, aggregate.HasUncommittedChanges())

	// Add a change
	event := NewBaseEventMessage("TestEvent", "test-id", "TestAggregate", 1, "test data")
	aggregate.TrackChange(event)

	// Should have changes now
	assert.True(t, aggregate.HasUncommittedChanges())
	assert.Equal(t, 1, aggregate.GetUncommittedChangeCount())

	// Clear changes
	aggregate.ClearChanges()

	// Should not have changes
	assert.False(t, aggregate.HasUncommittedChanges())
	assert.Equal(t, 0, aggregate.GetUncommittedChangeCount())
}

func TestBaseAggregate_SetTimestamps(t *testing.T) {
	// Arrange
	aggregate := NewBaseAggregate("test-id", "TestAggregate")
	createdAt := time.Now().Add(-1 * time.Hour)
	updatedAt := time.Now().Add(-30 * time.Minute)

	// Act
	aggregate.SetCreatedAt(createdAt)
	aggregate.SetUpdatedAt(updatedAt)

	// Assert
	assert.Equal(t, createdAt, aggregate.CreatedAt())
	assert.Equal(t, updatedAt, aggregate.UpdatedAt())
}

func TestBaseAggregate_SetDeleted(t *testing.T) {
	// Arrange
	aggregate := NewBaseAggregate("test-id", "TestAggregate")

	// Initially not deleted
	assert.False(t, aggregate.IsDeleted())

	// Act
	aggregate.SetDeleted(true)

	// Assert
	assert.True(t, aggregate.IsDeleted())

	// Act - Set back to false
	aggregate.SetDeleted(false)

	// Assert
	assert.False(t, aggregate.IsDeleted())
}

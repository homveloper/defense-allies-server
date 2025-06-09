package cqrs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBaseEventMessage(t *testing.T) {
	// Arrange
	eventType := "TestEvent"
	aggregateID := "test-id"
	aggregateType := "TestAggregate"
	version := 1
	eventData := "test data"

	// Act
	event := NewBaseEventMessage(eventType, eventData)

	// Assert
	assert.NotNil(t, event)
	assert.NotEmpty(t, event.EventID())
	assert.Equal(t, eventType, event.EventType())
	assert.Equal(t, aggregateID, event.ID())
	assert.Equal(t, aggregateType, event.Type())
	assert.Equal(t, version, event.Version())
	assert.Equal(t, eventData, event.EventData())
	assert.NotNil(t, event.Metadata())
	assert.NotZero(t, event.Timestamp())
}

func TestBaseEventMessage_Metadata(t *testing.T) {
	// Arrange
	event := NewBaseEventMessage("TestEvent", "test data")

	// Act & Assert - Add metadata
	event.AddMetadata("key1", "value1")
	event.AddMetadata("key2", 42)

	metadata := event.Metadata()
	assert.Len(t, metadata, 2)
	assert.Equal(t, "value1", metadata["key1"])
	assert.Equal(t, 42, metadata["key2"])

	// Test GetMetadata
	value, exists := event.GetMetadata("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)

	value, exists = event.GetMetadata("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, value)
}

func TestBaseEventMessage_SetTimestamp(t *testing.T) {
	// Arrange
	event := NewBaseEventMessage("TestEvent", "test data")
	newTimestamp := time.Now().Add(-1 * time.Hour)

	// Act
	event.SetTimestamp(newTimestamp)

	// Assert
	assert.Equal(t, newTimestamp, event.Timestamp())
}

func TestBaseEventMessage_SetEventID(t *testing.T) {
	// Arrange
	event := NewBaseEventMessage("TestEvent", "test data")
	newEventID := "custom-event-id"

	// Act
	event.SetEventID(newEventID)

	// Assert
	assert.Equal(t, newEventID, event.EventID())
}

func TestNewBaseDomainEvent(t *testing.T) {
	// Arrange
	eventType := "TestDomainEvent"
	aggregateID := "test-id"
	aggregateType := "TestAggregate"
	version := 1
	eventData := "test data"

	// Act
	event := NewBaseDomainEventMessage(eventType, eventData, []*BaseEventMessageOptions{
		Options().WithAggregateID(aggregateID).WithAggregateType(aggregateType).WithVersion(version),
	})

	// Assert
	assert.NotNil(t, event)
	assert.NotEmpty(t, event.EventID())
	assert.Equal(t, eventType, event.EventType())
	assert.Equal(t, aggregateID, event.ID())
	assert.Equal(t, aggregateType, event.Type())
	assert.Equal(t, version, event.Version())
	assert.Equal(t, eventData, event.EventData())

	// Domain event specific fields
	assert.Equal(t, DomainEvent, event.GetEventCategory())
	assert.Equal(t, PriorityNormal, event.GetPriority())
	assert.Equal(t, UserIssuer, event.IssuerType()) // Default issuer type
	assert.Empty(t, event.CausationID())
	assert.Empty(t, event.CorrelationID())
	assert.Empty(t, event.IssuerID())
}

func TestBaseDomainEvent_SetFields(t *testing.T) {
	// Arrange
	event := NewBaseDomainEventMessage("TestEvent", "test data", []*BaseEventMessageOptions{
		Options().WithAggregateID("test-id").WithAggregateType("TestAggregate").WithVersion(1),
	})

	// Act
	event.SetCausationID("causation-123")
	event.SetCorrelationID("correlation-456")
	event.SetIssuerID("issuer-789")
	event.SetIssuerType(SystemIssuer)
	event.SetCategory(SystemEvent)
	event.SetPriority(PriorityHigh)

	// Assert
	assert.Equal(t, "causation-123", event.CausationID())
	assert.Equal(t, "correlation-456", event.CorrelationID())
	assert.Equal(t, "issuer-789", event.IssuerID())
	assert.Equal(t, SystemIssuer, event.IssuerType())
	assert.Equal(t, SystemEvent, event.GetEventCategory())
	assert.Equal(t, PriorityHigh, event.GetPriority())
}

func TestBaseDomainEvent_ValidateEvent(t *testing.T) {
	tests := []struct {
		name        string
		eventID     string
		eventType   string
		aggregateID string
		expectError bool
	}{
		{
			name:        "Valid event",
			eventID:     "event-123",
			eventType:   "TestEvent",
			aggregateID: "aggregate-456",
			expectError: false,
		},
		{
			name:        "Empty event ID",
			eventID:     "",
			eventType:   "TestEvent",
			aggregateID: "aggregate-456",
			expectError: true,
		},
		{
			name:        "Empty event type",
			eventID:     "event-123",
			eventType:   "",
			aggregateID: "aggregate-456",
			expectError: true,
		},
		{
			name:        "Empty aggregate ID",
			eventID:     "event-123",
			eventType:   "TestEvent",
			aggregateID: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			event := NewBaseDomainEventMessage(TestedEventDataType, &testedEventData{}, []*BaseEventMessageOptions{Options()})
			event.SetEventID(tt.eventID)
			event.BaseEventMessage.eventType = tt.eventType
			event.BaseEventMessage.aggregateID = tt.aggregateID

			// Act
			err := event.ValidateEvent()

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBaseDomainEvent_GetChecksum(t *testing.T) {
	// Arrange
	event1 := NewBaseDomainEventMessage(TestedEventDataType, &testedEventData{Field1: "test", Field2: 123}, []*BaseEventMessageOptions{Options()})
	event2 := NewBaseDomainEventMessage(TestedEventDataType, &testedEventData{Field1: "test", Field2: 123}, []*BaseEventMessageOptions{Options()})
	event3 := NewBaseDomainEventMessage(TestedEventDataType, &testedEventData{Field1: "test", Field2: 456}, []*BaseEventMessageOptions{Options()})

	// Set same event IDs for consistent checksums
	event1.SetEventID("same-id")
	event2.SetEventID("same-id")
	event3.SetEventID("same-id")

	// Act
	checksum1 := event1.GetChecksum()
	checksum2 := event2.GetChecksum()
	checksum3 := event3.GetChecksum()

	// Assert
	assert.NotEmpty(t, checksum1)
	assert.Equal(t, checksum1, checksum2)    // Same data should produce same checksum
	assert.NotEqual(t, checksum1, checksum3) // Different data should produce different checksum
}

func TestEventCategory_String(t *testing.T) {
	tests := []struct {
		category EventCategory
		expected string
	}{
		{UserAction, "user_action"},
		{SystemEvent, "system_event"},
		{IntegrationEvent, "integration_event"},
		{DomainEvent, "domain_event"},
		{EventCategory(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.category.String())
		})
	}
}

func TestEventPriority_String(t *testing.T) {
	tests := []struct {
		priority EventPriority
		expected string
	}{
		{PriorityLow, "low"},
		{PriorityNormal, "normal"},
		{PriorityHigh, "high"},
		{PriorityCritical, "critical"},
		{EventPriority(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.priority.String())
		})
	}
}

func TestIssuerType_String(t *testing.T) {
	tests := []struct {
		issuerType IssuerType
		expected   string
	}{
		{UserIssuer, "user"},
		{SystemIssuer, "system"},
		{AdminIssuer, "admin"},
		{ServiceIssuer, "service"},
		{SchedulerIssuer, "scheduler"},
		{IssuerType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.issuerType.String())
		})
	}
}

func TestNewBaseDomainEventMessageWithIssuer(t *testing.T) {
	// Arrange
	eventType := "TestDomainEvent"
	aggregateID := "test-id"
	aggregateType := "TestAggregate"
	version := 1
	eventData := "test data"
	issuerID := "system-engine"
	issuerType := SystemIssuer

	// Act
	event := NewBaseDomainEventMessage(
		eventType, eventData, []*BaseEventMessageOptions{
			Options().WithAggregateID(aggregateID).WithAggregateType(aggregateType).WithVersion(version),
		}, &BaseDomainEventMessageOptions{
			IssuerID:   &issuerID,
			IssuerType: &issuerType,
		},
	)

	// Assert
	assert.NotNil(t, event)
	assert.NotEmpty(t, event.EventID())
	assert.Equal(t, eventType, event.EventType())
	assert.Equal(t, aggregateID, event.ID())
	assert.Equal(t, aggregateType, event.Type())
	assert.Equal(t, version, event.Version())
	assert.Equal(t, eventData, event.EventData())

	// Domain event specific fields
	assert.Equal(t, issuerID, event.IssuerID())
	assert.Equal(t, issuerType, event.IssuerType())
	assert.Equal(t, DomainEvent, event.GetEventCategory())
	assert.Equal(t, PriorityNormal, event.GetPriority())
}

func TestBaseDomainEvent_IssuerTypes(t *testing.T) {
	tests := []struct {
		name       string
		issuerType IssuerType
		issuerID   string
	}{
		{
			name:       "User issuer",
			issuerType: UserIssuer,
			issuerID:   "player123",
		},
		{
			name:       "System issuer",
			issuerType: SystemIssuer,
			issuerID:   "game-engine",
		},
		{
			name:       "Admin issuer",
			issuerType: AdminIssuer,
			issuerID:   "admin456",
		},
		{
			name:       "Service issuer",
			issuerType: ServiceIssuer,
			issuerID:   "external-api",
		},
		{
			name:       "Scheduler issuer",
			issuerType: SchedulerIssuer,
			issuerID:   "cron-job",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			event := NewBaseDomainEventMessage(
				"TestEvent", "test data", []*BaseEventMessageOptions{Options()}, &BaseDomainEventMessageOptions{
					IssuerID:   &tt.issuerID,
					IssuerType: &tt.issuerType,
				},
			)

			// Assert
			assert.Equal(t, tt.issuerID, event.IssuerID())
			assert.Equal(t, tt.issuerType, event.IssuerType())
		})
	}
}

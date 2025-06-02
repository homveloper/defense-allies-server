package cqrs

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// BaseEventMessage provides a base implementation of EventMessage interface
type BaseEventMessage struct {
	eventID       string
	eventType     string
	aggregateID   string
	aggregateType string
	version       int
	eventData     interface{}
	metadata      map[string]interface{}
	timestamp     time.Time
}

// NewBaseEventMessage creates a new BaseEventMessage
func NewBaseEventMessage(eventType, aggregateID, aggregateType string, version int, eventData interface{}) *BaseEventMessage {
	return &BaseEventMessage{
		eventID:       uuid.New().String(),
		eventType:     eventType,
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
		version:       version,
		eventData:     eventData,
		metadata:      make(map[string]interface{}),
		timestamp:     time.Now(),
	}
}

// EventMessage interface implementation

func (e *BaseEventMessage) EventID() string {
	return e.eventID
}

func (e *BaseEventMessage) EventType() string {
	return e.eventType
}

func (e *BaseEventMessage) AggregateID() string {
	return e.aggregateID
}

func (e *BaseEventMessage) AggregateType() string {
	return e.aggregateType
}

func (e *BaseEventMessage) Version() int {
	return e.version
}

func (e *BaseEventMessage) EventData() interface{} {
	return e.eventData
}

func (e *BaseEventMessage) Metadata() map[string]interface{} {
	return e.metadata
}

func (e *BaseEventMessage) Timestamp() time.Time {
	return e.timestamp
}

// Helper methods

// SetEventID sets the event ID (used when loading from storage)
func (e *BaseEventMessage) SetEventID(eventID string) {
	e.eventID = eventID
}

// SetTimestamp sets the timestamp (used when loading from storage)
func (e *BaseEventMessage) SetTimestamp(timestamp time.Time) {
	e.timestamp = timestamp
}

// AddMetadata adds metadata to the event
func (e *BaseEventMessage) AddMetadata(key string, value interface{}) {
	if e.metadata == nil {
		e.metadata = make(map[string]interface{})
	}
	e.metadata[key] = value
}

// GetMetadata gets metadata value by key
func (e *BaseEventMessage) GetMetadata(key string) (interface{}, bool) {
	if e.metadata == nil {
		return nil, false
	}
	value, exists := e.metadata[key]
	return value, exists
}

// BaseDomainEventMessage extends BaseEventMessage with DomainEventMessage features
type BaseDomainEventMessage struct {
	*BaseEventMessage
	causationID   string
	correlationID string
	userID        string
	category      EventCategory
	priority      EventPriority
	isSystem      bool
}

// NewBaseDomainEventMessage creates a new BaseDomainEventMessage
func NewBaseDomainEventMessage(eventType, aggregateID, aggregateType string, version int, eventData interface{}) *BaseDomainEventMessage {
	return &BaseDomainEventMessage{
		BaseEventMessage: NewBaseEventMessage(eventType, aggregateID, aggregateType, version, eventData),
		category:         DomainEvent,
		priority:         Normal,
		isSystem:         false,
	}
}

// DomainEventMessage interface implementation

func (e *BaseDomainEventMessage) CausationID() string {
	return e.causationID
}

func (e *BaseDomainEventMessage) CorrelationID() string {
	return e.correlationID
}

func (e *BaseDomainEventMessage) UserID() string {
	return e.userID
}

func (e *BaseDomainEventMessage) IsSystemEvent() bool {
	return e.isSystem
}

func (e *BaseDomainEventMessage) GetEventCategory() EventCategory {
	return e.category
}

func (e *BaseDomainEventMessage) GetPriority() EventPriority {
	return e.priority
}

func (e *BaseDomainEventMessage) ValidateEvent() error {
	if e.eventID == "" {
		return fmt.Errorf("event ID cannot be empty")
	}
	if e.eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	if e.aggregateID == "" {
		return fmt.Errorf("aggregate ID cannot be empty")
	}
	return nil
}

func (e *BaseDomainEventMessage) GetChecksum() string {
	data := fmt.Sprintf("%s:%s:%s:%d:%v",
		e.eventID, e.eventType, e.aggregateID, e.version, e.eventData)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// Helper methods for BaseDomainEventMessage

func (e *BaseDomainEventMessage) SetCausationID(causationID string) {
	e.causationID = causationID
}

func (e *BaseDomainEventMessage) SetCorrelationID(correlationID string) {
	e.correlationID = correlationID
}

func (e *BaseDomainEventMessage) SetUserID(userID string) {
	e.userID = userID
}

func (e *BaseDomainEventMessage) SetCategory(category EventCategory) {
	e.category = category
}

func (e *BaseDomainEventMessage) SetPriority(priority EventPriority) {
	e.priority = priority
}

func (e *BaseDomainEventMessage) SetIsSystem(isSystem bool) {
	e.isSystem = isSystem
}

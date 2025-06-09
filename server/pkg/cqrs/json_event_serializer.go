package cqrs

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

// EventSerializer interface for event serialization in the core CQRS package
type EventSerializer interface {
	Serialize(event EventMessage) ([]byte, error)
	Deserialize(data []byte) (EventMessage, error)
}

// JSONEventSerializer provides JSON serialization for BaseEventMessage with EventDataRegistry support
type JSONEventSerializer struct {
	registry *EventDataRegistry
}

// NewJSONEventSerializer creates a new JSON event serializer with the given registry
func NewJSONEventSerializer(registry *EventDataRegistry) *JSONEventSerializer {
	if registry == nil {
		registry = NewEventDataRegistry()
	}

	return &JSONEventSerializer{
		registry: registry,
	}
}

// SerializableEventMessage represents the JSON structure for serialized events
type SerializableEventMessage struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int                    `json:"version"`
	EventData     interface{}            `json:"event_data"`
	Metadata      map[string]interface{} `json:"metadata"`
	Timestamp     time.Time              `json:"timestamp"`
}

// Serialize serializes a BaseEventMessage to JSON bytes
func (s *JSONEventSerializer) Serialize(event EventMessage) ([]byte, error) {
	if event == nil {
		return nil, NewCQRSError(ErrCodeSerializationError.String(), "event cannot be nil", nil)
	}

	serializableEvent := SerializableEventMessage{
		EventID:       event.EventID(),
		EventType:     event.EventType(),
		AggregateID:   event.ID(),
		AggregateType: event.Type(),
		Version:       event.Version(),
		EventData:     event.EventData(),
		Metadata:      event.Metadata(),
		Timestamp:     event.Timestamp(),
	}

	data, err := json.Marshal(serializableEvent)
	if err != nil {
		return nil, NewCQRSError(ErrCodeSerializationError.String(),
			fmt.Sprintf("failed to serialize event to JSON: %v", err), err)
	}

	return data, nil
}

// Deserialize deserializes JSON bytes to a BaseEventMessage
func (s *JSONEventSerializer) Deserialize(data []byte) (EventMessage, error) {
	if len(data) == 0 {
		return nil, NewCQRSError(ErrCodeSerializationError.String(), "data cannot be empty", nil)
	}

	var serializableEvent SerializableEventMessage
	if err := json.Unmarshal(data, &serializableEvent); err != nil {
		return nil, NewCQRSError(ErrCodeSerializationError.String(),
			fmt.Sprintf("failed to unmarshal event from JSON: %v", err), err)
	}

	// Deserialize event data using registry if available
	eventData, err := s.deserializeEventData(serializableEvent.EventData, serializableEvent.EventType)
	if err != nil {
		return nil, NewCQRSError(ErrCodeSerializationError.String(),
			fmt.Sprintf("failed to deserialize event data: %v", err), err)
	}

	// Create BaseEventMessage with all the data
	event := NewBaseEventMessage(
		serializableEvent.EventType,
		eventData,
		Options().WithLoad(
			serializableEvent.EventID,
			serializableEvent.Timestamp,
			serializableEvent.Metadata,
			serializableEvent.AggregateID,
			serializableEvent.AggregateType,
			serializableEvent.Version,
		),
	)

	return event, nil
}

// deserializeEventData deserializes event data using the registry
func (s *JSONEventSerializer) deserializeEventData(data interface{}, eventType string) (interface{}, error) {
	// If no registry or event type not registered, return data as-is
	if s.registry == nil || !s.registry.IsRegistered(eventType) {
		return data, nil
	}

	// Get the registered type (always a pointer type)
	targetType, err := s.registry.GetEventDataType(eventType)
	if err != nil {
		return data, nil // Fallback to original data
	}

	// Convert data to JSON and back to the target type
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event data to JSON: %w", err)
	}

	// Create new instance of the target type (always a pointer)
	target := reflect.New(targetType.Elem()).Interface()

	if err := json.Unmarshal(jsonData, target); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event data to target type: %w", err)
	}

	return target, nil
}

// GetRegistry returns the event data registry
func (s *JSONEventSerializer) GetRegistry() *EventDataRegistry {
	return s.registry
}

// SetRegistry sets the event data registry
func (s *JSONEventSerializer) SetRegistry(registry *EventDataRegistry) {
	s.registry = registry
}

// RegisterEventData is a convenience method to register event data types
func (s *JSONEventSerializer) RegisterEventData(eventType string, eventData interface{}) error {
	if s.registry == nil {
		s.registry = NewEventDataRegistry()
	}

	return s.registry.RegisterEventData(eventType, eventData)
}

// ValidateEventData validates that the given event data can be properly serialized and deserialized
func (s *JSONEventSerializer) ValidateEventData(eventData interface{}) error {
	// Test serialization
	_, err := json.Marshal(eventData)
	if err != nil {
		return NewCQRSError(ErrCodeValidationError.String(),
			fmt.Sprintf("event data cannot be marshaled to JSON: %v", err), err)
	}

	// Test deserialization
	dataType := reflect.TypeOf(eventData)
	var target interface{}

	if dataType.Kind() == reflect.Ptr {
		target = reflect.New(dataType.Elem()).Interface()
	} else {
		target = reflect.New(dataType).Interface()
	}

	jsonData, _ := json.Marshal(eventData)
	if err := json.Unmarshal(jsonData, target); err != nil {
		return NewCQRSError(ErrCodeValidationError.String(),
			fmt.Sprintf("event data cannot be unmarshaled from JSON: %v", err), err)
	}

	return nil
}

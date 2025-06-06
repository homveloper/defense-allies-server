package cqrsx

import (
	"defense-allies-server/pkg/cqrs"
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// EventSerializer interface for event serialization
// This interface provides a clean abstraction for different event serialization strategies
type EventSerializer interface {
	Serialize(event cqrs.EventMessage) ([]byte, error)
	Deserialize(data []byte) (cqrs.EventMessage, error)
}

// Note: EventSerializerWithType was removed as it's unnecessary - event type is stored within the serialized data

// EventData represents serialized event data structure
// This is the standard format used for JSON serialization of events
type EventData struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int                    `json:"version"`
	Data          interface{}            `json:"data"`
	Metadata      map[string]interface{} `json:"metadata"`
	Timestamp     time.Time              `json:"timestamp"`
}

// JSONEventSerializer implements EventSerializer using JSON format
// This serializer stores the event type within the serialized data
type JSONEventSerializer struct{}

// Serialize serializes an event to JSON bytes
func (s *JSONEventSerializer) Serialize(event cqrs.EventMessage) ([]byte, error) {
	eventData := EventData{
		EventID:       event.EventID(),
		EventType:     event.EventType(),
		AggregateID:   event.ID(),
		AggregateType: event.Type(),
		Version:       event.Version(),
		Data:          event.EventData(),
		Metadata:      event.Metadata(),
		Timestamp:     event.Timestamp(),
	}

	return json.Marshal(eventData)
}

// Deserialize deserializes JSON bytes to an event
func (s *JSONEventSerializer) Deserialize(data []byte) (cqrs.EventMessage, error) {
	var eventData EventData
	if err := json.Unmarshal(data, &eventData); err != nil {
		return nil, err
	}

	event := cqrs.NewBaseEventMessage(
		eventData.EventType,
		eventData.Data,
		cqrs.Options().WithLoad(
			eventData.EventID,
			eventData.Timestamp,
			eventData.Metadata,
			eventData.AggregateID,
			eventData.AggregateType,
			eventData.Version,
		),
	)

	// Set metadata
	for key, value := range eventData.Metadata {
		event.AddMetadata(key, value)
	}

	return event, nil
}

// Note: JSONEventSerializerWithType was removed as it's unnecessary

// CompactEventSerializer implements a more compact JSON serialization
// This serializer reduces the size of serialized events by using shorter field names
type CompactEventSerializer struct{}

// CompactEventData represents a more compact event data structure
type CompactEventData struct {
	ID  string                 `json:"i"`  // event_id
	T   string                 `json:"t"`  // event_type
	AID string                 `json:"ai"` // aggregate_id
	AT  string                 `json:"at"` // aggregate_type
	V   int                    `json:"v"`  // version
	D   interface{}            `json:"d"`  // data
	M   map[string]interface{} `json:"m"`  // metadata
	TS  time.Time              `json:"ts"` // timestamp
}

// Serialize serializes an event to compact JSON bytes
func (s *CompactEventSerializer) Serialize(event cqrs.EventMessage) ([]byte, error) {
	eventData := CompactEventData{
		ID:  event.EventID(),
		T:   event.EventType(),
		AID: event.ID(),
		AT:  event.Type(),
		V:   event.Version(),
		D:   event.EventData(),
		M:   event.Metadata(),
		TS:  event.Timestamp(),
	}

	return json.Marshal(eventData)
}

// Deserialize deserializes compact JSON bytes to an event
func (s *CompactEventSerializer) Deserialize(data []byte) (cqrs.EventMessage, error) {
	var eventData CompactEventData
	if err := json.Unmarshal(data, &eventData); err != nil {
		return nil, err
	}

	event := cqrs.NewBaseEventMessage(
		eventData.T,
		eventData.D,
		cqrs.Options().WithLoad(
			eventData.ID,
			eventData.TS,
			eventData.M,
			eventData.AID,
			eventData.AT,
			eventData.V,
		),
	)

	event.SetEventID(eventData.ID)
	event.SetTimestamp(eventData.TS)

	// Set metadata
	for key, value := range eventData.M {
		event.AddMetadata(key, value)
	}

	return event, nil
}

// BinaryEventSerializer implements binary serialization for events
// This can be implemented later for performance-critical scenarios
type BinaryEventSerializer struct{}

// Serialize serializes an event to binary format (placeholder implementation)
func (s *BinaryEventSerializer) Serialize(event cqrs.EventMessage) ([]byte, error) {
	// For now, fallback to JSON
	jsonSerializer := &JSONEventSerializer{}
	return jsonSerializer.Serialize(event)
}

// Deserialize deserializes binary data to an event (placeholder implementation)
func (s *BinaryEventSerializer) Deserialize(data []byte) (cqrs.EventMessage, error) {
	// For now, fallback to JSON
	jsonSerializer := &JSONEventSerializer{}
	return jsonSerializer.Deserialize(data)
}

// EventSerializerFactory creates event serializers based on format
type EventSerializerFactory struct{}

// SerializerFormat represents different serialization formats
type SerializerFormat string

const (
	JSONFormat    SerializerFormat = "json"
	CompactFormat SerializerFormat = "compact"
	BinaryFormat  SerializerFormat = "binary"
)

// CreateEventSerializer creates an event serializer for the specified format
func (f *EventSerializerFactory) CreateEventSerializer(format SerializerFormat) EventSerializer {
	switch format {
	case JSONFormat:
		return &JSONEventSerializer{}
	case CompactFormat:
		return &CompactEventSerializer{}
	case BinaryFormat:
		return &BinaryEventSerializer{}
	default:
		return &JSONEventSerializer{} // Default to JSON
	}
}

// Note: CreateEventSerializerWithType was removed as it's unnecessary

// GetSupportedFormats returns all supported serialization formats
func (f *EventSerializerFactory) GetSupportedFormats() []SerializerFormat {
	return []SerializerFormat{JSONFormat, CompactFormat, BinaryFormat}
}

// BSONEventSerializer implements EventSerializer using BSON format for MongoDB
type BSONEventSerializer struct{}

// Serialize serializes an event to BSON bytes
func (s *BSONEventSerializer) Serialize(event cqrs.EventMessage) ([]byte, error) {
	eventData := EventData{
		EventID:       event.EventID(),
		EventType:     event.EventType(),
		AggregateID:   event.ID(),
		AggregateType: event.Type(),
		Version:       event.Version(),
		Data:          event.EventData(),
		Metadata:      event.Metadata(),
		Timestamp:     event.Timestamp(),
	}

	return bson.Marshal(eventData)
}

// Deserialize deserializes BSON bytes to an event
func (s *BSONEventSerializer) Deserialize(data []byte) (cqrs.EventMessage, error) {
	var eventData EventData
	if err := bson.Unmarshal(data, &eventData); err != nil {
		return nil, err
	}

	event := cqrs.NewBaseEventMessage(
		eventData.EventType,
		eventData.Data,
		cqrs.Options().WithLoad(
			eventData.EventID,
			eventData.Timestamp,
			eventData.Metadata,
			eventData.AggregateID,
			eventData.AggregateType,
			eventData.Version,
		),
	)

	event.SetEventID(eventData.EventID)
	event.SetTimestamp(eventData.Timestamp)

	// Set metadata
	for key, value := range eventData.Metadata {
		event.AddMetadata(key, value)
	}

	return event, nil
}

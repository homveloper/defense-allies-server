package cqrsx

import (
	"cqrs"
)

// EventMarshaler interface for event serialization
// This interface provides a clean abstraction for different event serialization strategies
type EventMarshaler interface {
	Marshal(event cqrs.EventMessage) ([]byte, error)
	Unmarshal(data []byte) (cqrs.EventMessage, error)
}

// JSONEventMarshaler implements EventSerializer using JSON format
// This serializer stores the event type within the serialized data
type JSONEventMarshaler struct {
	registry EventRegistry
}

func NewJSONEventMarshaler(registry EventRegistry) *JSONEventMarshaler {
	return &JSONEventMarshaler{
		registry: registry,
	}
}

// Marshal serializes an event to JSON bytes
func (s *JSONEventMarshaler) Marshal(event cqrs.EventMessage) ([]byte, error) {
	return MarshalEventJSON(event)
}

// Unmarshal deserializes JSON bytes to an event
func (s *JSONEventMarshaler) Unmarshal(data []byte) (cqrs.EventMessage, error) {
	return UnmarshalEventJSON(data, s.registry)
}

// EventSerializerFactory creates event serializers based on format
type EventSerializerFactory struct{}

// SerializerFormat represents different serialization formats
type SerializerFormat string

const (
	JSONFormat SerializerFormat = "json"
	BSONFormat SerializerFormat = "bson"
)

// CreateEventSerializer creates an event serializer for the specified format
func (f *EventSerializerFactory) CreateEventSerializer(format SerializerFormat) EventMarshaler {
	switch format {
	case JSONFormat:
		return &JSONEventMarshaler{}
	case BSONFormat:
		return &BSONEventMarshaler{}
	default:
		return &JSONEventMarshaler{} // Default to JSON
	}
}

// Note: CreateEventSerializerWithType was removed as it's unnecessary

// GetSupportedFormats returns all supported serialization formats
func (f *EventSerializerFactory) GetSupportedFormats() []SerializerFormat {
	return []SerializerFormat{JSONFormat, BSONFormat}
}

// BSONEventMarshaler implements EventSerializer using BSON format for MongoDB
type BSONEventMarshaler struct {
	registry EventRegistry
}

func NewBSONEventMarshaler(registry EventRegistry) *BSONEventMarshaler {
	return &BSONEventMarshaler{
		registry: registry,
	}
}

// Marshal serializes an event to BSON bytes
func (s *BSONEventMarshaler) Marshal(event cqrs.EventMessage) ([]byte, error) {
	return MarshalEventBSON(event)
}

// Unmarshal deserializes BSON bytes to an event
func (s *BSONEventMarshaler) Unmarshal(data []byte) (cqrs.EventMessage, error) {
	return UnmarshalEventBSON(data, s.registry)
}

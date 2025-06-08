package redisstream

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"defense-allies-server/pkg/cqrs"
)

// SerializationFormat represents different serialization formats
type SerializationFormat string

const (
	SerializationFormatJSON     SerializationFormat = "json"
	SerializationFormatProtobuf SerializationFormat = "protobuf"
	SerializationFormatAvro     SerializationFormat = "avro"
)

// EventSerializer interface for serializing/deserializing events
type EventSerializer interface {
	Serialize(event cqrs.EventMessage) ([]byte, error)
	Deserialize(data []byte) (cqrs.EventMessage, error)
	Format() SerializationFormat
}

// SerializationRegistry manages different serialization formats
type SerializationRegistry interface {
	Register(format SerializationFormat, serializer EventSerializer) error
	Get(format SerializationFormat) (EventSerializer, error)
	SupportedFormats() []SerializationFormat
}

// SerializedEventData represents the structure of serialized event data
type SerializedEventData struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int                    `json:"version"`
	EventData     interface{}            `json:"event_data"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`

	// Domain event specific fields
	IssuerID      string `json:"issuer_id,omitempty"`
	IssuerType    string `json:"issuer_type,omitempty"`
	CausationID   string `json:"causation_id,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
	Category      string `json:"category,omitempty"`
	Priority      string `json:"priority,omitempty"`

	// Serialization metadata
	SerializationVersion string `json:"serialization_version"`
	SerializationFormat  string `json:"serialization_format"`
	Checksum             string `json:"checksum,omitempty"`
}

// JSONEventSerializer implements EventSerializer for JSON format
type JSONEventSerializer struct {
	format SerializationFormat
}

// NewJSONEventSerializer creates a new JSON event serializer
func NewJSONEventSerializer() EventSerializer {
	return &JSONEventSerializer{
		format: SerializationFormatJSON,
	}
}

// Serialize converts an event to JSON bytes
func (s *JSONEventSerializer) Serialize(event cqrs.EventMessage) ([]byte, error) {
	if event == nil {
		return nil, fmt.Errorf("%w: event cannot be nil", ErrSerializationFailed)
	}

	// Create serialized data structure
	serializedData := &SerializedEventData{
		EventID:              event.EventID(),
		EventType:            event.EventType(),
		AggregateID:          event.ID(),
		AggregateType:        event.Type(),
		Version:              event.Version(),
		EventData:            event.EventData(),
		Metadata:             event.Metadata(),
		Timestamp:            event.Timestamp(),
		SerializationVersion: "1.0",
		SerializationFormat:  string(s.format),
	}

	// Handle domain event specific fields
	if domainEvent, ok := event.(cqrs.DomainEventMessage); ok {
		serializedData.IssuerID = domainEvent.IssuerID()
		serializedData.IssuerType = domainEvent.IssuerType().String()
		serializedData.CausationID = domainEvent.CausationID()
		serializedData.CorrelationID = domainEvent.CorrelationID()
		serializedData.Category = domainEvent.GetEventCategory().String()
		serializedData.Priority = domainEvent.GetPriority().String()
		serializedData.Checksum = domainEvent.GetChecksum()
	}

	// Serialize to JSON
	data, err := json.Marshal(serializedData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSerializationFailed, err)
	}

	return data, nil
}

// Deserialize converts JSON bytes back to an event
func (s *JSONEventSerializer) Deserialize(data []byte) (cqrs.EventMessage, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("%w: data cannot be empty", ErrDeserializationFailed)
	}

	var serializedData SerializedEventData
	if err := json.Unmarshal(data, &serializedData); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDeserializationFailed, err)
	}

	// Validate required fields
	if err := s.validateSerializedData(&serializedData); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDeserializationFailed, err)
	}

	// Create base event options
	baseOptions := cqrs.Options().
		WithEventID(serializedData.EventID).
		WithAggregateID(serializedData.AggregateID).
		WithAggregateType(serializedData.AggregateType).
		WithVersion(serializedData.Version).
		WithTimestamp(serializedData.Timestamp).
		WithMetadata(serializedData.Metadata)

	// Check if this is a domain event
	if s.isDomainEvent(&serializedData) {
		// Create domain event options
		domainOptions := &cqrs.BaseDomainEventMessageOptions{}

		if serializedData.IssuerID != "" {
			domainOptions.IssuerID = &serializedData.IssuerID
		}

		if serializedData.IssuerType != "" {
			if issuerType, err := s.parseIssuerType(serializedData.IssuerType); err == nil {
				domainOptions.IssuerType = &issuerType
			}
		}

		if serializedData.CausationID != "" {
			domainOptions.CausationID = &serializedData.CausationID
		}

		if serializedData.CorrelationID != "" {
			domainOptions.CorrelationID = &serializedData.CorrelationID
		}

		if serializedData.Category != "" {
			if category, err := s.parseEventCategory(serializedData.Category); err == nil {
				domainOptions.Category = &category
			}
		}

		if serializedData.Priority != "" {
			if priority, err := s.parseEventPriority(serializedData.Priority); err == nil {
				domainOptions.Priority = &priority
			}
		}

		return cqrs.NewBaseDomainEventMessage(
			serializedData.EventType,
			serializedData.EventData,
			[]*cqrs.BaseEventMessageOptions{baseOptions},
			domainOptions,
		), nil
	}

	// Create regular event
	return cqrs.NewBaseEventMessage(
		serializedData.EventType,
		serializedData.EventData,
		baseOptions,
	), nil
}

// Format returns the serialization format
func (s *JSONEventSerializer) Format() SerializationFormat {
	return s.format
}

// Helper methods

func (s *JSONEventSerializer) validateSerializedData(data *SerializedEventData) error {
	if data.EventID == "" {
		return fmt.Errorf("event_id is required")
	}
	if data.EventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if data.AggregateID == "" {
		return fmt.Errorf("aggregate_id is required")
	}
	if data.AggregateType == "" {
		return fmt.Errorf("aggregate_type is required")
	}
	return nil
}

func (s *JSONEventSerializer) isDomainEvent(data *SerializedEventData) bool {
	return data.IssuerType != "" || data.Category != "" || data.Priority != ""
}

func (s *JSONEventSerializer) parseIssuerType(issuerTypeStr string) (cqrs.IssuerType, error) {
	switch issuerTypeStr {
	case "user":
		return cqrs.UserIssuer, nil
	case "system":
		return cqrs.SystemIssuer, nil
	case "admin":
		return cqrs.AdminIssuer, nil
	case "service":
		return cqrs.ServiceIssuer, nil
	case "scheduler":
		return cqrs.SchedulerIssuer, nil
	default:
		return cqrs.UserIssuer, fmt.Errorf("unknown issuer type: %s", issuerTypeStr)
	}
}

func (s *JSONEventSerializer) parseEventCategory(categoryStr string) (cqrs.EventCategory, error) {
	switch categoryStr {
	case "user_action":
		return cqrs.UserAction, nil
	case "system_event":
		return cqrs.SystemEvent, nil
	case "integration_event":
		return cqrs.IntegrationEvent, nil
	case "domain_event":
		return cqrs.DomainEvent, nil
	default:
		return cqrs.DomainEvent, fmt.Errorf("unknown event category: %s", categoryStr)
	}
}

func (s *JSONEventSerializer) parseEventPriority(priorityStr string) (cqrs.EventPriority, error) {
	switch priorityStr {
	case "low":
		return cqrs.PriorityLow, nil
	case "normal":
		return cqrs.PriorityNormal, nil
	case "high":
		return cqrs.PriorityHigh, nil
	case "critical":
		return cqrs.PriorityCritical, nil
	default:
		return cqrs.PriorityNormal, fmt.Errorf("unknown event priority: %s", priorityStr)
	}
}

// defaultSerializationRegistry implements SerializationRegistry
type defaultSerializationRegistry struct {
	serializers map[SerializationFormat]EventSerializer
	mu          sync.RWMutex
}

// NewSerializationRegistry creates a new serialization registry
func NewSerializationRegistry() SerializationRegistry {
	return &defaultSerializationRegistry{
		serializers: make(map[SerializationFormat]EventSerializer),
	}
}

// Register adds a serializer for a specific format
func (r *defaultSerializationRegistry) Register(format SerializationFormat, serializer EventSerializer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.serializers[format]; exists {
		return fmt.Errorf("%w: format %s", ErrSerializationFormatAlreadyRegistered, format)
	}

	r.serializers[format] = serializer
	return nil
}

// Get retrieves a serializer for a specific format
func (r *defaultSerializationRegistry) Get(format SerializationFormat) (EventSerializer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	serializer, exists := r.serializers[format]
	if !exists {
		return nil, fmt.Errorf("%w: format %s", ErrSerializationFormatNotSupported, format)
	}

	return serializer, nil
}

// SupportedFormats returns all supported formats
func (r *defaultSerializationRegistry) SupportedFormats() []SerializationFormat {
	r.mu.RLock()
	defer r.mu.RUnlock()

	formats := make([]SerializationFormat, 0, len(r.serializers))
	for format := range r.serializers {
		formats = append(formats, format)
	}

	return formats
}

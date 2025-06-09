package cqrs

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var _ EventMessage = (*BaseEventMessage)(nil)

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

func Options() *BaseEventMessageOptions {
	return &BaseEventMessageOptions{}
}

// BaseEventMessageOptions defines options for creating BaseEventMessage
type BaseEventMessageOptions struct {
	EventID       *string                `json:"event_id,omitempty"`
	Timestamp     *time.Time             `json:"timestamp,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	string        *string                `json:"aggregate_id,omitempty"`
	AggregateType *string                `json:"aggregate_type,omitempty"`
	Version       *int                   `json:"version,omitempty"`
}

func (opts *BaseEventMessageOptions) WithLoad(
	eventID string,
	timestamp time.Time,
	metadata map[string]interface{},
	aggregateID string,
	aggregateType string,
	version int,
) *BaseEventMessageOptions {
	opts.EventID = &eventID
	opts.Timestamp = &timestamp
	opts.Metadata = metadata
	opts.string = &aggregateID
	opts.AggregateType = &aggregateType
	opts.Version = &version
	return opts
}

func (opts *BaseEventMessageOptions) WithEventID(eventID string) *BaseEventMessageOptions {
	opts.EventID = &eventID
	return opts
}

func (opts *BaseEventMessageOptions) WithTimestamp(timestamp time.Time) *BaseEventMessageOptions {
	opts.Timestamp = &timestamp
	return opts
}

func (opts *BaseEventMessageOptions) WithMetadata(metadata map[string]interface{}) *BaseEventMessageOptions {
	opts.Metadata = metadata
	return opts
}

func (opts *BaseEventMessageOptions) WithAggregateID(aggregateID string) *BaseEventMessageOptions {
	opts.string = &aggregateID
	return opts
}

func (opts *BaseEventMessageOptions) WithAggregateType(aggregateType string) *BaseEventMessageOptions {
	opts.AggregateType = &aggregateType
	return opts
}

func (opts *BaseEventMessageOptions) WithVersion(version int) *BaseEventMessageOptions {
	opts.Version = &version
	return opts
}

// Merge combines multiple options, with later options overriding earlier ones
func mergeOptions(options ...*BaseEventMessageOptions) *BaseEventMessageOptions {

	merged := &BaseEventMessageOptions{}

	for _, opt := range options {
		if opt.EventID != nil {
			merged.EventID = opt.EventID
		}
		if opt.Timestamp != nil {
			merged.Timestamp = opt.Timestamp
		}
		if opt.Metadata != nil {
			merged.Metadata = opt.Metadata
		}
		if opt.string != nil {
			merged.string = opt.string
		}
		if opt.AggregateType != nil {
			merged.AggregateType = opt.AggregateType
		}
		if opt.Version != nil {
			merged.Version = opt.Version
		}
	}

	return merged
}

func applyOptions(event *BaseEventMessage, opts *BaseEventMessageOptions) {
	if opts == nil {
		return
	}

	if opts.EventID != nil {
		event.eventID = *opts.EventID
	}
	if opts.Timestamp != nil {
		event.timestamp = *opts.Timestamp
	}
	if opts.Metadata != nil {
		event.metadata = opts.Metadata
	}
	if opts.string != nil {
		event.aggregateID = *opts.string
	}
	if opts.AggregateType != nil {
		event.aggregateType = *opts.AggregateType
	}
	if opts.Version != nil {
		event.version = *opts.Version
	}
}

// NewBaseEventMessage creates a new BaseEventMessage with optional configuration
// Note: aggregateID, aggregateType, version, and timestamp will be auto-populated by BaseAggregate.ApplyEvent unless specified
func NewBaseEventMessage(eventType string, eventData interface{}, options ...*BaseEventMessageOptions) *BaseEventMessage {

	opt := mergeOptions(options...)

	event := &BaseEventMessage{
		eventID:       uuid.New().String(),
		eventType:     eventType,
		aggregateID:   "", // Will be set by ApplyEvent unless overridden
		aggregateType: "", // Will be set by ApplyEvent unless overridden
		version:       0,  // Will be set by ApplyEvent unless overridden
		eventData:     eventData,
		metadata:      make(map[string]interface{}),
		timestamp:     time.Time{}, // Will be set by ApplyEvent unless overridden
	}

	applyOptions(event, opt)

	return event
}

// EventMessage interface implementation

func (e *BaseEventMessage) EventID() string {
	return e.eventID
}

func (e *BaseEventMessage) EventType() string {
	return e.eventType
}

func (e *BaseEventMessage) ID() string {
	return e.aggregateID
}

func (e *BaseEventMessage) Type() string {
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

func (e *BaseEventMessage) Clone() EventMessage {
	clone := *e
	clone.metadata = make(map[string]interface{})
	for k, v := range e.metadata {
		clone.metadata[k] = v
	}
	return &clone
}

// CloneWithOptions creates a clone with optional modifications
func (e *BaseEventMessage) CloneWithOptions(options ...*BaseEventMessageOptions) EventMessage {
	clone := *e
	clone.metadata = make(map[string]interface{})
	for k, v := range e.metadata {
		clone.metadata[k] = v
	}

	opt := mergeOptions(options...)
	applyOptions(&clone, opt)

	return &clone
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
	issuerID      string
	issuerType    IssuerType
	causationID   string
	correlationID string
	category      EventCategory
	priority      EventPriority
}

// BaseDomainEventMessageOptions defines options for creating BaseDomainEventMessage
type BaseDomainEventMessageOptions struct {
	IssuerID      *string        `json:"issuer_id,omitempty"`
	IssuerType    *IssuerType    `json:"issuer_type,omitempty"`
	CausationID   *string        `json:"causation_id,omitempty"`
	CorrelationID *string        `json:"correlation_id,omitempty"`
	Category      *EventCategory `json:"category,omitempty"`
	Priority      *EventPriority `json:"priority,omitempty"`
}

// Merge combines multiple domain options, with later options overriding earlier ones
func (opts *BaseDomainEventMessageOptions) Merge(other *BaseDomainEventMessageOptions) *BaseDomainEventMessageOptions {
	if other == nil {
		return opts
	}

	merged := &BaseDomainEventMessageOptions{}

	// Copy from current options
	if opts != nil {
		if opts.IssuerID != nil {
			merged.IssuerID = opts.IssuerID
		}
		if opts.IssuerType != nil {
			merged.IssuerType = opts.IssuerType
		}
		if opts.CausationID != nil {
			merged.CausationID = opts.CausationID
		}
		if opts.CorrelationID != nil {
			merged.CorrelationID = opts.CorrelationID
		}
		if opts.Category != nil {
			merged.Category = opts.Category
		}
		if opts.Priority != nil {
			merged.Priority = opts.Priority
		}
	}

	// Override with other options
	if other.IssuerID != nil {
		merged.IssuerID = other.IssuerID
	}
	if other.IssuerType != nil {
		merged.IssuerType = other.IssuerType
	}
	if other.CausationID != nil {
		merged.CausationID = other.CausationID
	}
	if other.CorrelationID != nil {
		merged.CorrelationID = other.CorrelationID
	}
	if other.Category != nil {
		merged.Category = other.Category
	}
	if other.Priority != nil {
		merged.Priority = other.Priority
	}

	return merged
}

// Apply applies the options to a BaseDomainEventMessage
func (opts *BaseDomainEventMessageOptions) Apply(event *BaseDomainEventMessage) {
	if opts == nil {
		return
	}

	if opts.IssuerID != nil {
		event.issuerID = *opts.IssuerID
	}
	if opts.IssuerType != nil {
		event.issuerType = *opts.IssuerType
	}
	if opts.CausationID != nil {
		event.causationID = *opts.CausationID
	}
	if opts.CorrelationID != nil {
		event.correlationID = *opts.CorrelationID
	}
	if opts.Category != nil {
		event.category = *opts.Category
	}
	if opts.Priority != nil {
		event.priority = *opts.Priority
	}
}

// NewBaseDomainEventMessage creates a new BaseDomainEventMessage with optional configuration
// Note: aggregateID, aggregateType, version will be auto-populated by BaseAggregate.ApplyEvent unless specified
func NewBaseDomainEventMessage(eventType string, eventData interface{}, baseOptions []*BaseEventMessageOptions, domainOptions ...*BaseDomainEventMessageOptions) *BaseDomainEventMessage {
	domainEvent := &BaseDomainEventMessage{
		BaseEventMessage: NewBaseEventMessage(eventType, eventData, baseOptions...),
		issuerType:       UserIssuer, // Default to user issuer
		category:         DomainEvent,
		priority:         PriorityNormal,
	}

	// Merge and apply domain-specific options
	var mergedDomainOptions *BaseDomainEventMessageOptions
	for _, option := range domainOptions {
		mergedDomainOptions = mergedDomainOptions.Merge(option)
	}
	mergedDomainOptions.Apply(domainEvent)

	return domainEvent
}

// DomainEventMessage interface implementation

func (e *BaseDomainEventMessage) CausationID() string {
	return e.causationID
}

func (e *BaseDomainEventMessage) CorrelationID() string {
	return e.correlationID
}

func (e *BaseDomainEventMessage) IssuerID() string {
	return e.issuerID
}

func (e *BaseDomainEventMessage) IssuerType() IssuerType {
	return e.issuerType
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

func (e *BaseDomainEventMessage) SetIssuerID(issuerID string) {
	e.issuerID = issuerID
}

func (e *BaseDomainEventMessage) SetIssuerType(issuerType IssuerType) {
	e.issuerType = issuerType
}

func (e *BaseDomainEventMessage) SetCategory(category EventCategory) {
	e.category = category
}

func (e *BaseDomainEventMessage) SetPriority(priority EventPriority) {
	e.priority = priority
}

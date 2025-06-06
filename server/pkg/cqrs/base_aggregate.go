package cqrs

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var _ AggregateRoot = (*BaseAggregate)(nil)

// BaseAggregate provides a base implementation of the AggregateRoot interface
// Optimized for Defense Allies with clean and simple API
type BaseAggregate struct {
	id              string
	aggregateType   string
	originalVersion int
	currentVersion  int
	changes         []EventMessage
	createdAt       time.Time
	updatedAt       time.Time
	deleted         bool
}

// BaseAggregateOption defines options for creating BaseAggregate
type BaseAggregateOption func(*BaseAggregate)

// WithOriginalVersion sets the original version (used when loading from storage)
func WithOriginalVersion(version int) BaseAggregateOption {
	return func(a *BaseAggregate) {
		a.originalVersion = version
		a.currentVersion = version
	}
}

// WithCreatedAt sets the creation timestamp (used when loading from storage)
func WithCreatedAt(createdAt time.Time) BaseAggregateOption {
	return func(a *BaseAggregate) {
		a.createdAt = createdAt
	}
}

// WithUpdatedAt sets the update timestamp (used when loading from storage)
func WithUpdatedAt(updatedAt time.Time) BaseAggregateOption {
	return func(a *BaseAggregate) {
		a.updatedAt = updatedAt
	}
}

// WithDeleted sets the deleted flag (used when loading from storage)
func WithDeleted(deleted bool) BaseAggregateOption {
	return func(a *BaseAggregate) {
		a.deleted = deleted
	}
}

// NewBaseAggregate creates a new BaseAggregate with optional configuration
func NewBaseAggregate(id, aggregateType string, options ...BaseAggregateOption) *BaseAggregate {
	now := time.Now()
	aggregate := &BaseAggregate{
		id:              id,
		aggregateType:   aggregateType,
		originalVersion: 0,
		currentVersion:  0,
		changes:         make([]EventMessage, 0),
		createdAt:       now,
		updatedAt:       now,
		deleted:         false,
	}

	// Apply options
	for _, option := range options {
		option(aggregate)
	}

	return aggregate
}

// AggregateRoot interface implementation

func (a *BaseAggregate) ID() string {
	return a.id
}

func (a *BaseAggregate) Type() string {
	return a.aggregateType
}

func (a *BaseAggregate) Version() int {
	return a.currentVersion
}

func (a *BaseAggregate) OriginalVersion() int {
	return a.originalVersion
}

func (a *BaseAggregate) SetOriginalVersion(version int) {
	a.originalVersion = version
}

func (a *BaseAggregate) nextVersion() int {
	a.updatedAt = time.Now()
	a.currentVersion++
	return a.currentVersion
}

// ApplyEvent applies a newly generated event and tracks it as an uncommitted change
func (a *BaseAggregate) ApplyEvent(event EventMessage) error {
	// Validate event
	if event == nil {
		return errors.New("event cannot be nil")
	}

	// Update version and timestamp first
	// Auto-populate aggregate metadata in the event
	if baseEvent, ok := event.(*BaseEventMessage); ok {
		version := a.nextVersion()
		timestamp := time.Now()

		event = baseEvent.CloneWithOptions(
			&BaseEventMessageOptions{
				AggregateID:   &a.id,
				AggregateType: &a.aggregateType,
				Version:       &version,
				Timestamp:     &timestamp,
			},
		)
	}

	// Track new events for persistence
	a.changes = append(a.changes, event)

	return nil
}

// ReplayEvent applies an existing event during state reconstruction
func (a *BaseAggregate) ReplayEvent(event EventMessage) error {
	// Validate event
	if event == nil {
		return errors.New("event cannot be nil")
	}

	// Update version and timestamp (but don't track as new change)
	a.nextVersion()

	return nil
}

func (a *BaseAggregate) Changes() []EventMessage {
	return a.changes
}

func (a *BaseAggregate) ClearChanges() {
	a.changes = nil
}

// LoadFromHistory reconstructs the aggregate state by replaying events
func (a *BaseAggregate) LoadFromHistory(events []EventMessage) error {
	for _, event := range events {
		if err := a.ReplayEvent(event); err != nil {
			return fmt.Errorf("failed to replay event: %w", err)
		}
	}
	return nil
}

func (a *BaseAggregate) Validate() error {
	if a.id == "" {
		return errors.New("aggregate ID cannot be empty")
	}
	if a.aggregateType == "" {
		return errors.New("aggregate type cannot be empty")
	}
	return nil
}

// Helper methods

// HasUncommittedChanges returns true if there are uncommitted changes
func (a *BaseAggregate) HasUncommittedChanges() bool {
	return len(a.changes) > 0
}

// GetUncommittedChangeCount returns the number of uncommitted changes
func (a *BaseAggregate) GetUncommittedChangeCount() int {
	return len(a.changes)
}

// JSON Serialization Support

// baseAggregateJSON represents the JSON structure for BaseAggregate
type baseAggregateJSON struct {
	ID              string         `json:"id"`
	Type            string         `json:"type"`
	OriginalVersion int            `json:"original_version"`
	Version         int            `json:"version"`
	Changes         []EventMessage `json:"changes,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	Deleted         bool           `json:"deleted"`
}

// MarshalJSON implements custom JSON marshaling for BaseAggregate
func (a *BaseAggregate) MarshalJSON() ([]byte, error) {
	return json.Marshal(&baseAggregateJSON{
		ID:              a.id,
		Type:            a.aggregateType,
		OriginalVersion: a.originalVersion,
		Version:         a.currentVersion,
		Changes:         a.changes,
		CreatedAt:       a.createdAt,
		UpdatedAt:       a.updatedAt,
		Deleted:         a.deleted,
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for BaseAggregate
func (a *BaseAggregate) UnmarshalJSON(data []byte) error {
	var jsonData baseAggregateJSON
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return err
	}

	a.id = jsonData.ID
	a.aggregateType = jsonData.Type
	a.originalVersion = jsonData.OriginalVersion
	a.currentVersion = jsonData.Version
	a.changes = jsonData.Changes
	a.createdAt = jsonData.CreatedAt
	a.updatedAt = jsonData.UpdatedAt
	a.deleted = jsonData.Deleted

	// Initialize changes slice if nil
	if a.changes == nil {
		a.changes = make([]EventMessage, 0)
	}

	return nil
}

package cqrs

import (
	"errors"
	"time"
)

// BaseAggregate provides a base implementation of the Aggregate interface
// Compatible with go.cqrs style
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

// NewBaseAggregate creates a new BaseAggregate
func NewBaseAggregate(id, aggregateType string) *BaseAggregate {
	now := time.Now()
	return &BaseAggregate{
		id:              id,
		aggregateType:   aggregateType,
		originalVersion: 0,
		currentVersion:  0,
		changes:         make([]EventMessage, 0),
		createdAt:       now,
		updatedAt:       now,
		deleted:         false,
	}
}

// AggregateRoot interface implementation

func (a *BaseAggregate) AggregateID() string {
	return a.id
}

func (a *BaseAggregate) OriginalVersion() int {
	return a.originalVersion
}

func (a *BaseAggregate) CurrentVersion() int {
	return a.currentVersion
}

func (a *BaseAggregate) IncrementVersion() {
	a.currentVersion++
	a.updatedAt = time.Now()
}

func (a *BaseAggregate) Apply(event EventMessage, isNew bool) {
	// Apply event to aggregate
	if isNew {
		a.TrackChange(event)
	}
	a.IncrementVersion()
}

func (a *BaseAggregate) TrackChange(event EventMessage) {
	a.changes = append(a.changes, event)
}

func (a *BaseAggregate) GetChanges() []EventMessage {
	return a.changes
}

func (a *BaseAggregate) ClearChanges() {
	a.changes = nil
}

// Defense Allies Aggregate interface implementation

func (a *BaseAggregate) AggregateType() string {
	return a.aggregateType
}

func (a *BaseAggregate) CreatedAt() time.Time {
	return a.createdAt
}

func (a *BaseAggregate) UpdatedAt() time.Time {
	return a.updatedAt
}

func (a *BaseAggregate) IsDeleted() bool {
	return a.deleted
}

func (a *BaseAggregate) MarkAsDeleted() {
	a.deleted = true
	a.updatedAt = time.Now()
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

// SetOriginalVersion sets the original version (used when loading from storage)
func (a *BaseAggregate) SetOriginalVersion(version int) {
	a.originalVersion = version
	a.currentVersion = version
}

// SetCreatedAt sets the creation time (used when loading from storage)
func (a *BaseAggregate) SetCreatedAt(createdAt time.Time) {
	a.createdAt = createdAt
}

// SetUpdatedAt sets the update time (used when loading from storage)
func (a *BaseAggregate) SetUpdatedAt(updatedAt time.Time) {
	a.updatedAt = updatedAt
}

// SetDeleted sets the deleted flag (used when loading from storage)
func (a *BaseAggregate) SetDeleted(deleted bool) {
	a.deleted = deleted
}

// HasUncommittedChanges returns true if there are uncommitted changes
func (a *BaseAggregate) HasUncommittedChanges() bool {
	return len(a.changes) > 0
}

// GetUncommittedChangeCount returns the number of uncommitted changes
func (a *BaseAggregate) GetUncommittedChangeCount() int {
	return len(a.changes)
}

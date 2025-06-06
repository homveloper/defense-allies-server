package cqrsx

import (
	"context"
	"fmt"
	"log"
	"time"

	"defense-allies-server/pkg/cqrs"
)

// SnapshotConfiguration holds snapshot management configuration
type SnapshotConfiguration struct {
	Enabled                  bool          `json:"enabled"`
	MaxSnapshotsPerAggregate int           `json:"max_snapshots_per_aggregate"`
	AsyncCreation            bool          `json:"async_creation"`
	CleanupInterval          time.Duration `json:"cleanup_interval"`
	CompressionEnabled       bool          `json:"compression_enabled"`
	CompressionType          string        `json:"compression_type"`
	SerializationType        string        `json:"serialization_type"`
}

// DefaultSnapshotConfiguration returns default configuration
func DefaultSnapshotConfiguration() *SnapshotConfiguration {
	return &SnapshotConfiguration{
		Enabled:                  true,
		MaxSnapshotsPerAggregate: 5,
		AsyncCreation:            false,
		CleanupInterval:          24 * time.Hour,
		CompressionEnabled:       true,
		CompressionType:          "gzip",
		SerializationType:        "json",
	}
}

// SnapshotManager manages snapshot creation, restoration, and cleanup
type SnapshotManager interface {
	// CreateSnapshot creates a snapshot for the given aggregate
	CreateSnapshot(ctx context.Context, aggregate cqrs.AggregateRoot) error

	// RestoreFromSnapshot restores an aggregate from snapshot
	RestoreFromSnapshot(ctx context.Context, aggregateID string, maxVersion int) (cqrs.AggregateRoot, int, error)

	// ShouldCreateSnapshot determines if a snapshot should be created
	ShouldCreateSnapshot(aggregate cqrs.AggregateRoot, eventCount int) bool

	// CleanupOldSnapshots removes old snapshots
	CleanupOldSnapshots(ctx context.Context, aggregateID string) error

	// GetSnapshotInfo returns snapshot information
	GetSnapshotInfo(ctx context.Context, aggregateID string) ([]SnapshotInfo, error)

	// GetConfiguration returns current configuration
	GetConfiguration() *SnapshotConfiguration

	// UpdateConfiguration updates the configuration
	UpdateConfiguration(config *SnapshotConfiguration) error
}

// SnapshotInfo contains information about a snapshot
type SnapshotInfo struct {
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int                    `json:"version"`
	Size          int64                  `json:"size"`
	Timestamp     time.Time              `json:"timestamp"`
	ContentType   string                 `json:"content_type"`
	Compression   string                 `json:"compression"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// SnapshotEvent represents a snapshot-related event for monitoring
type SnapshotEvent struct {
	Type          string                 `json:"type"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int                    `json:"version"`
	Timestamp     time.Time              `json:"timestamp"`
	Duration      time.Duration          `json:"duration"`
	Size          int64                  `json:"size"`
	Success       bool                   `json:"success"`
	Error         string                 `json:"error,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// Snapshot event types
const (
	SnapshotEventCreated  = "created"
	SnapshotEventRestored = "restored"
	SnapshotEventFailed   = "failed"
	SnapshotEventDeleted  = "deleted"
)

// SnapshotError represents snapshot-related errors
type SnapshotError struct {
	Code      string
	Message   string
	Operation string
	Cause     error
}

func (e *SnapshotError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *SnapshotError) Unwrap() error {
	return e.Cause
}

// Error codes
const (
	ErrCodeSnapshotNotFound       = "SNAPSHOT_NOT_FOUND"
	ErrCodeSerializationFailed    = "SERIALIZATION_FAILED"
	ErrCodeDeserializationFailed  = "DESERIALIZATION_FAILED"
	ErrCodeStorageFailed          = "STORAGE_FAILED"
	ErrCodeConfigurationInvalid   = "CONFIGURATION_INVALID"
	ErrCodePolicyEvaluationFailed = "POLICY_EVALUATION_FAILED"
)

// AdvancedSnapshotStore extends the basic snapshot store with advanced features
type AdvancedSnapshotStore interface {
	// Basic operations
	SaveSnapshot(ctx context.Context, aggregate cqrs.AggregateRoot) error
	LoadSnapshot(ctx context.Context, aggregateID, aggregateType string) (cqrs.AggregateRoot, error)

	// Advanced operations
	GetSnapshot(ctx context.Context, aggregateID string, maxVersion int) (SnapshotData, error)
	GetSnapshotByVersion(ctx context.Context, aggregateID string, version int) (SnapshotData, error)
	DeleteSnapshot(ctx context.Context, aggregateID string, version int) error
	DeleteOldSnapshots(ctx context.Context, aggregateID string, keepCount int) error
	ListSnapshotsForAggregate(ctx context.Context, aggregateID string) ([]SnapshotData, error)
	GetSnapshotStats(ctx context.Context) (map[string]interface{}, error)
}

// SnapshotData represents snapshot data with metadata
type SnapshotData interface {
	AggregateID() string
	AggregateType() string
	Version() int
	Data() []byte
	Timestamp() time.Time
	Metadata() map[string]interface{}
	Size() int64
	ContentType() string
	Compression() string
}

// DefaultSnapshotManager implements SnapshotManager with advanced features
type DefaultSnapshotManager struct {
	store      AdvancedSnapshotStore
	serializer AdvancedSnapshotSerializer
	policy     SnapshotPolicy
	config     *SnapshotConfiguration
}

// NewDefaultSnapshotManager creates a new snapshot manager
func NewDefaultSnapshotManager(
	store AdvancedSnapshotStore,
	serializer AdvancedSnapshotSerializer,
	policy SnapshotPolicy,
	config *SnapshotConfiguration,
) *DefaultSnapshotManager {
	if config == nil {
		config = DefaultSnapshotConfiguration()
	}

	return &DefaultSnapshotManager{
		store:      store,
		serializer: serializer,
		policy:     policy,
		config:     config,
	}
}

// CreateSnapshot creates a snapshot for the given aggregate
func (m *DefaultSnapshotManager) CreateSnapshot(ctx context.Context, aggregate cqrs.AggregateRoot) error {
	if !m.config.Enabled {
		return nil
	}

	start := time.Now()

	// Serialize aggregate
	data, err := m.serializer.SerializeSnapshot(aggregate)
	if err != nil {
		m.logSnapshotEvent(SnapshotEventFailed, aggregate, 0, time.Since(start), 0, err)
		return &SnapshotError{
			Code:      ErrCodeSerializationFailed,
			Message:   "failed to serialize aggregate",
			Operation: "CreateSnapshot",
			Cause:     err,
		}
	}

	// Save snapshot
	if err := m.store.SaveSnapshot(ctx, aggregate); err != nil {
		m.logSnapshotEvent(SnapshotEventFailed, aggregate, int64(len(data)), time.Since(start), 0, err)
		return &SnapshotError{
			Code:      ErrCodeStorageFailed,
			Message:   "failed to save snapshot",
			Operation: "CreateSnapshot",
			Cause:     err,
		}
	}

	// Log success
	m.logSnapshotEvent(SnapshotEventCreated, aggregate, int64(len(data)), time.Since(start), 0, nil)

	// Cleanup old snapshots asynchronously
	if m.config.MaxSnapshotsPerAggregate > 0 {
		go func() {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := m.store.DeleteOldSnapshots(cleanupCtx, aggregate.AggregateID(), m.config.MaxSnapshotsPerAggregate); err != nil {
				log.Printf("Failed to cleanup old snapshots for %s: %v", aggregate.AggregateID(), err)
			}
		}()
	}

	return nil
}

// RestoreFromSnapshot restores an aggregate from snapshot
func (m *DefaultSnapshotManager) RestoreFromSnapshot(ctx context.Context, aggregateID string, maxVersion int) (cqrs.AggregateRoot, int, error) {
	start := time.Now()

	// Get snapshot
	snapshot, err := m.store.GetSnapshot(ctx, aggregateID, maxVersion)
	if err != nil {
		if snapshotErr, ok := err.(*SnapshotError); ok && snapshotErr.Code == ErrCodeSnapshotNotFound {
			return nil, 0, err // snapshot not found, return as-is
		}
		return nil, 0, &SnapshotError{
			Code:      ErrCodeStorageFailed,
			Message:   "failed to get snapshot",
			Operation: "RestoreFromSnapshot",
			Cause:     err,
		}
	}

	// Deserialize
	aggregate, err := m.serializer.DeserializeSnapshot(snapshot.Data(), snapshot.AggregateType())
	if err != nil {
		m.logSnapshotEvent(SnapshotEventFailed, nil, int64(len(snapshot.Data())), time.Since(start), 0, err)
		return nil, 0, &SnapshotError{
			Code:      ErrCodeDeserializationFailed,
			Message:   "failed to deserialize snapshot",
			Operation: "RestoreFromSnapshot",
			Cause:     err,
		}
	}

	// Log success
	m.logSnapshotEvent(SnapshotEventRestored, aggregate, int64(len(snapshot.Data())), time.Since(start), 0, nil)

	return aggregate, snapshot.Version(), nil
}

// ShouldCreateSnapshot determines if a snapshot should be created
func (m *DefaultSnapshotManager) ShouldCreateSnapshot(aggregate cqrs.AggregateRoot, eventCount int) bool {
	if !m.config.Enabled {
		return false
	}

	return m.policy.ShouldCreateSnapshot(aggregate, eventCount)
}

// CleanupOldSnapshots removes old snapshots
func (m *DefaultSnapshotManager) CleanupOldSnapshots(ctx context.Context, aggregateID string) error {
	if m.config.MaxSnapshotsPerAggregate <= 0 {
		return nil
	}

	start := time.Now()

	err := m.store.DeleteOldSnapshots(ctx, aggregateID, m.config.MaxSnapshotsPerAggregate)
	if err != nil {
		return &SnapshotError{
			Code:      ErrCodeStorageFailed,
			Message:   "failed to cleanup old snapshots",
			Operation: "CleanupOldSnapshots",
			Cause:     err,
		}
	}

	log.Printf("Cleaned up old snapshots for %s in %v", aggregateID, time.Since(start))
	return nil
}

// GetSnapshotInfo returns snapshot information
func (m *DefaultSnapshotManager) GetSnapshotInfo(ctx context.Context, aggregateID string) ([]SnapshotInfo, error) {
	snapshots, err := m.store.ListSnapshotsForAggregate(ctx, aggregateID)
	if err != nil {
		return nil, &SnapshotError{
			Code:      ErrCodeStorageFailed,
			Message:   "failed to list snapshots",
			Operation: "GetSnapshotInfo",
			Cause:     err,
		}
	}

	var infos []SnapshotInfo
	for _, snapshot := range snapshots {
		info := SnapshotInfo{
			AggregateID:   snapshot.AggregateID(),
			AggregateType: snapshot.AggregateType(),
			Version:       snapshot.Version(),
			Size:          snapshot.Size(),
			Timestamp:     snapshot.Timestamp(),
			ContentType:   snapshot.ContentType(),
			Compression:   snapshot.Compression(),
			Metadata:      snapshot.Metadata(),
		}
		infos = append(infos, info)
	}

	return infos, nil
}

// GetConfiguration returns current configuration
func (m *DefaultSnapshotManager) GetConfiguration() *SnapshotConfiguration {
	return m.config
}

// UpdateConfiguration updates the configuration
func (m *DefaultSnapshotManager) UpdateConfiguration(config *SnapshotConfiguration) error {
	if config == nil {
		return &SnapshotError{
			Code:      ErrCodeConfigurationInvalid,
			Message:   "configuration cannot be nil",
			Operation: "UpdateConfiguration",
		}
	}

	m.config = config
	return nil
}

// logSnapshotEvent logs snapshot events
func (m *DefaultSnapshotManager) logSnapshotEvent(eventType string, aggregate cqrs.AggregateRoot, size int64, duration time.Duration, version int, err error) {
	event := SnapshotEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Duration:  duration,
		Size:      size,
		Success:   err == nil,
		Metadata:  make(map[string]interface{}),
	}

	if aggregate != nil {
		event.AggregateID = aggregate.AggregateID()
		event.AggregateType = aggregate.AggregateType()
		event.Version = aggregate.CurrentVersion()
	} else {
		event.Version = version
	}

	if err != nil {
		event.Error = err.Error()
	}

	// Add metadata
	event.Metadata["policy"] = m.policy.GetPolicyName()
	event.Metadata["serializer"] = m.serializer.GetContentType()
	event.Metadata["compression"] = m.serializer.GetCompressionType()

	// Log output
	if err != nil {
		log.Printf("Snapshot %s failed for %s v%d: %v (took %v)",
			eventType, event.AggregateID, event.Version, err, duration)
	} else {
		log.Printf("Snapshot %s for %s v%d: %d bytes (took %v)",
			eventType, event.AggregateID, event.Version, size, duration)
	}
}

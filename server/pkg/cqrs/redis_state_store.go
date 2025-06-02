package cqrs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStateStore implements state-based aggregate storage using Redis
type RedisStateStore struct {
	client     *RedisClientManager
	keyBuilder *RedisKeyBuilder
	serializer StateSerializer
}

// StateSerializer interface for aggregate state serialization
type StateSerializer interface {
	SerializeAggregate(aggregate Aggregate) ([]byte, error)
	DeserializeAggregate(data []byte, aggregateType string) (AggregateRoot, error)
}

// JSONStateSerializer implements JSON-based state serialization
type JSONStateSerializer struct{}

// AggregateData represents serialized aggregate data
type AggregateData struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"`
	Version         int                    `json:"version"`
	OriginalVersion int                    `json:"original_version"`
	Data            interface{}            `json:"data"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	IsDeleted       bool                   `json:"is_deleted"`
}

// NewRedisStateStore creates a new Redis state store
func NewRedisStateStore(client *RedisClientManager, keyPrefix string) *RedisStateStore {
	return &RedisStateStore{
		client:     client,
		keyBuilder: NewRedisKeyBuilder(keyPrefix),
		serializer: &JSONStateSerializer{},
	}
}

// Save saves an aggregate to Redis
func (ss *RedisStateStore) Save(ctx context.Context, aggregate interface{}, expectedVersion int) error {
	if aggregate == nil {
		return NewCQRSError(ErrCodeRepositoryError.String(), "aggregate cannot be nil", nil)
	}

	// Cast to Aggregate interface for additional methods
	agg, ok := aggregate.(Aggregate)
	if !ok {
		return NewCQRSError(ErrCodeRepositoryError.String(), "aggregate must implement Aggregate interface", nil)
	}

	if err := agg.Validate(); err != nil {
		return NewCQRSError(ErrCodeRepositoryError.String(), "aggregate validation failed", err)
	}

	aggregateKey := ss.keyBuilder.AggregateKey(agg.AggregateType(), agg.AggregateID())
	lockKey := ss.keyBuilder.LockKey(agg.AggregateType(), agg.AggregateID())

	return ss.client.ExecuteCommand(ctx, func() error {
		// Acquire distributed lock
		lockAcquired, err := ss.acquireLock(ctx, lockKey, 30*time.Second)
		if err != nil {
			return err
		}
		if !lockAcquired {
			return NewCQRSError(ErrCodeConcurrencyConflict.String(), "failed to acquire lock", nil)
		}

		defer ss.releaseLock(ctx, lockKey)

		// Check current version for optimistic concurrency control
		currentVersion, err := ss.getCurrentVersion(ctx, aggregateKey)
		if err != nil {
			return err
		}

		if currentVersion != expectedVersion {
			return NewCQRSError(ErrCodeConcurrencyConflict.String(),
				fmt.Sprintf("expected version %d, but current version is %d", expectedVersion, currentVersion),
				ErrConcurrencyConflict)
		}

		// Serialize aggregate
		data, err := ss.serializer.SerializeAggregate(agg)
		if err != nil {
			return NewCQRSError(ErrCodeSerializationError.String(), "failed to serialize aggregate", err)
		}

		// Save to Redis with expiration
		err = ss.client.GetClient().Set(ctx, aggregateKey, data, 24*time.Hour).Err()
		if err != nil {
			return NewCQRSError(ErrCodeRepositoryError.String(), "failed to save aggregate", err)
		}

		return nil
	})
}

// GetByID retrieves an aggregate by ID
func (ss *RedisStateStore) GetByID(ctx context.Context, aggregateType, aggregateID string) (AggregateRoot, error) {
	if aggregateID == "" {
		return nil, NewCQRSError(ErrCodeRepositoryError.String(), "aggregate ID cannot be empty", nil)
	}
	if aggregateType == "" {
		return nil, NewCQRSError(ErrCodeRepositoryError.String(), "aggregate type cannot be empty", nil)
	}

	aggregateKey := ss.keyBuilder.AggregateKey(aggregateType, aggregateID)

	var aggregate AggregateRoot

	err := ss.client.ExecuteCommand(ctx, func() error {
		data, err := ss.client.GetClient().Get(ctx, aggregateKey).Result()
		if err != nil {
			if err == redis.Nil {
				return NewCQRSError(ErrCodeAggregateNotFound.String(),
					fmt.Sprintf("aggregate not found: %s:%s", aggregateType, aggregateID),
					ErrAggregateNotFound)
			}
			return NewCQRSError(ErrCodeRepositoryError.String(), "failed to get aggregate", err)
		}

		aggregate, err = ss.serializer.DeserializeAggregate([]byte(data), aggregateType)
		if err != nil {
			return NewCQRSError(ErrCodeSerializationError.String(), "failed to deserialize aggregate", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return aggregate, nil
}

// GetVersion gets the current version of an aggregate
func (ss *RedisStateStore) GetVersion(ctx context.Context, aggregateType, aggregateID string) (int, error) {
	if aggregateID == "" {
		return 0, NewCQRSError(ErrCodeRepositoryError.String(), "aggregate ID cannot be empty", nil)
	}
	if aggregateType == "" {
		return 0, NewCQRSError(ErrCodeRepositoryError.String(), "aggregate type cannot be empty", nil)
	}

	aggregateKey := ss.keyBuilder.AggregateKey(aggregateType, aggregateID)

	return ss.getCurrentVersion(ctx, aggregateKey)
}

// Exists checks if an aggregate exists
func (ss *RedisStateStore) Exists(ctx context.Context, aggregateType, aggregateID string) bool {
	if aggregateID == "" || aggregateType == "" {
		return false
	}

	aggregateKey := ss.keyBuilder.AggregateKey(aggregateType, aggregateID)

	exists := false
	ss.client.ExecuteCommand(ctx, func() error {
		result, err := ss.client.GetClient().Exists(ctx, aggregateKey).Result()
		if err == nil {
			exists = result > 0
		}
		return nil // Don't propagate error for existence check
	})

	return exists
}

// Delete removes an aggregate
func (ss *RedisStateStore) Delete(ctx context.Context, aggregateType, aggregateID string) error {
	if aggregateID == "" {
		return NewCQRSError(ErrCodeRepositoryError.String(), "aggregate ID cannot be empty", nil)
	}
	if aggregateType == "" {
		return NewCQRSError(ErrCodeRepositoryError.String(), "aggregate type cannot be empty", nil)
	}

	aggregateKey := ss.keyBuilder.AggregateKey(aggregateType, aggregateID)
	lockKey := ss.keyBuilder.LockKey(aggregateType, aggregateID)

	return ss.client.ExecuteCommand(ctx, func() error {
		// Acquire distributed lock
		lockAcquired, err := ss.acquireLock(ctx, lockKey, 30*time.Second)
		if err != nil {
			return err
		}
		if !lockAcquired {
			return NewCQRSError(ErrCodeConcurrencyConflict.String(), "failed to acquire lock", nil)
		}

		defer ss.releaseLock(ctx, lockKey)

		// Delete aggregate
		err = ss.client.GetClient().Del(ctx, aggregateKey).Err()
		if err != nil {
			return NewCQRSError(ErrCodeRepositoryError.String(), "failed to delete aggregate", err)
		}

		return nil
	})
}

// Helper methods

func (ss *RedisStateStore) getCurrentVersion(ctx context.Context, aggregateKey string) (int, error) {
	data, err := ss.client.GetClient().Get(ctx, aggregateKey).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // Aggregate doesn't exist, version is 0
		}
		return 0, NewCQRSError(ErrCodeRepositoryError.String(), "failed to get aggregate for version check", err)
	}

	// Parse version from serialized data
	var aggregateData AggregateData
	if err := json.Unmarshal([]byte(data), &aggregateData); err != nil {
		return 0, NewCQRSError(ErrCodeSerializationError.String(), "failed to parse aggregate data", err)
	}

	return aggregateData.Version, nil
}

func (ss *RedisStateStore) acquireLock(ctx context.Context, lockKey string, expiration time.Duration) (bool, error) {
	result, err := ss.client.GetClient().SetNX(ctx, lockKey, "locked", expiration).Result()
	if err != nil {
		return false, NewCQRSError(ErrCodeRepositoryError.String(), "failed to acquire lock", err)
	}
	return result, nil
}

func (ss *RedisStateStore) releaseLock(ctx context.Context, lockKey string) error {
	return ss.client.GetClient().Del(ctx, lockKey).Err()
}

// JSONStateSerializer implementation

func (s *JSONStateSerializer) SerializeAggregate(aggregate Aggregate) ([]byte, error) {
	// Note: In a real implementation, you would need to handle different aggregate types
	// For now, we'll create a generic representation
	aggregateData := AggregateData{
		ID:              aggregate.AggregateID(),
		Type:            aggregate.AggregateType(),
		Version:         aggregate.CurrentVersion(),
		OriginalVersion: aggregate.OriginalVersion(),
		Data:            aggregate, // This would need proper handling in real implementation
		Metadata:        make(map[string]interface{}),
		CreatedAt:       aggregate.CreatedAt(),
		UpdatedAt:       aggregate.UpdatedAt(),
		IsDeleted:       aggregate.IsDeleted(),
	}

	return json.Marshal(aggregateData)
}

func (s *JSONStateSerializer) DeserializeAggregate(data []byte, aggregateType string) (AggregateRoot, error) {
	var aggregateData AggregateData
	if err := json.Unmarshal(data, &aggregateData); err != nil {
		return nil, err
	}

	// Note: In a real implementation, you would need a factory to create specific aggregate types
	// For now, we'll create a BaseAggregate
	aggregate := NewBaseAggregate(aggregateData.ID, aggregateData.Type)
	aggregate.SetOriginalVersion(aggregateData.OriginalVersion)
	aggregate.SetCreatedAt(aggregateData.CreatedAt)
	aggregate.SetUpdatedAt(aggregateData.UpdatedAt)
	aggregate.SetDeleted(aggregateData.IsDeleted)

	// Set current version to match the stored version
	for i := aggregate.CurrentVersion(); i < aggregateData.Version; i++ {
		aggregate.IncrementVersion()
	}

	return aggregate, nil
}

package cqrs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisReadStore implements ReadStore using Redis
type RedisReadStore struct {
	client     *RedisClientManager
	keyBuilder *RedisKeyBuilder
	serializer ReadModelSerializer
	factory    ReadModelFactory
}

// ReadModelFactory interface for creating read models by type
type ReadModelFactory interface {
	CreateReadModel(modelType string, id string, data interface{}) (ReadModel, error)
}

// ReadModelSerializer interface for read model serialization
type ReadModelSerializer interface {
	SerializeReadModel(model ReadModel) ([]byte, error)
	DeserializeReadModel(data []byte, modelType string, factory ReadModelFactory) (ReadModel, error)
}

// JSONReadModelSerializer implements JSON-based read model serialization
type JSONReadModelSerializer struct{}

// ReadModelData represents serialized read model data
type ReadModelData struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	Version     int         `json:"version"`
	Data        interface{} `json:"data"`
	LastUpdated time.Time   `json:"last_updated"`
}

// NewRedisReadStore creates a new Redis read store
func NewRedisReadStore(client *RedisClientManager, keyPrefix string, factory ReadModelFactory) *RedisReadStore {
	return &RedisReadStore{
		client:     client,
		keyBuilder: NewRedisKeyBuilder(keyPrefix),
		serializer: &JSONReadModelSerializer{},
		factory:    factory,
	}
}

// Save saves a read model to Redis
func (rs *RedisReadStore) Save(ctx context.Context, readModel ReadModel) error {
	if readModel == nil {
		return NewCQRSError(ErrCodeRepositoryError.String(), "read model cannot be nil", nil)
	}

	if err := readModel.Validate(); err != nil {
		return NewCQRSError(ErrCodeRepositoryError.String(), "read model validation failed", err)
	}

	modelKey := rs.keyBuilder.ReadModelKey(readModel.GetType(), readModel.GetID())

	return rs.client.ExecuteCommand(ctx, func() error {
		// Serialize read model
		data, err := rs.serializer.SerializeReadModel(readModel)
		if err != nil {
			return NewCQRSError(ErrCodeSerializationError.String(), "failed to serialize read model", err)
		}

		// Save to Redis with expiration
		err = rs.client.GetClient().Set(ctx, modelKey, data, 24*time.Hour).Err()
		if err != nil {
			return NewCQRSError(ErrCodeRepositoryError.String(), "failed to save read model", err)
		}

		// Update indexes
		return rs.updateIndexes(ctx, readModel)
	})
}

// GetByID retrieves a read model by ID
func (rs *RedisReadStore) GetByID(ctx context.Context, id string, modelType string) (ReadModel, error) {
	if id == "" {
		return nil, NewCQRSError(ErrCodeRepositoryError.String(), "id cannot be empty", nil)
	}
	if modelType == "" {
		return nil, NewCQRSError(ErrCodeRepositoryError.String(), "model type cannot be empty", nil)
	}

	modelKey := rs.keyBuilder.ReadModelKey(modelType, id)

	var readModel ReadModel

	err := rs.client.ExecuteCommand(ctx, func() error {
		data, err := rs.client.GetClient().Get(ctx, modelKey).Result()
		if err != nil {
			if err == redis.Nil {
				return NewCQRSError(ErrCodeRepositoryError.String(),
					fmt.Sprintf("read model not found: %s:%s", modelType, id), nil)
			}
			return NewCQRSError(ErrCodeRepositoryError.String(), "failed to get read model", err)
		}

		readModel, err = rs.serializer.DeserializeReadModel([]byte(data), modelType, rs.factory)
		if err != nil {
			return NewCQRSError(ErrCodeSerializationError.String(), "failed to deserialize read model", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return readModel, nil
}

// Delete removes a read model
func (rs *RedisReadStore) Delete(ctx context.Context, id string, modelType string) error {
	if id == "" {
		return NewCQRSError(ErrCodeRepositoryError.String(), "id cannot be empty", nil)
	}
	if modelType == "" {
		return NewCQRSError(ErrCodeRepositoryError.String(), "model type cannot be empty", nil)
	}

	modelKey := rs.keyBuilder.ReadModelKey(modelType, id)

	return rs.client.ExecuteCommand(ctx, func() error {
		// Get the model first to update indexes
		readModel, err := rs.GetByID(ctx, id, modelType)
		if err != nil {
			// If model doesn't exist, consider it already deleted
			return nil
		}

		// Remove from indexes
		err = rs.removeFromIndexes(ctx, readModel)
		if err != nil {
			return err
		}

		// Delete the model
		err = rs.client.GetClient().Del(ctx, modelKey).Err()
		if err != nil {
			return NewCQRSError(ErrCodeRepositoryError.String(), "failed to delete read model", err)
		}

		return nil
	})
}

// Query performs a query on read models
func (rs *RedisReadStore) Query(ctx context.Context, criteria QueryCriteria) ([]ReadModel, error) {
	var results []ReadModel

	err := rs.client.ExecuteCommand(ctx, func() error {
		// Simple implementation using pattern matching
		// In a real implementation, you'd use proper indexing

		pattern := rs.keyBuilder.ReadModelKey("*", "*")
		keys, err := rs.client.GetClient().Keys(ctx, pattern).Result()
		if err != nil {
			return NewCQRSError(ErrCodeRepositoryError.String(), "failed to get keys", err)
		}

		for _, key := range keys {
			data, err := rs.client.GetClient().Get(ctx, key).Result()
			if err != nil {
				continue // Skip invalid entries
			}

			// Extract model type from key
			parts := strings.Split(key, ":")
			if len(parts) < 4 {
				continue
			}
			modelType := parts[len(parts)-2]

			readModel, err := rs.serializer.DeserializeReadModel([]byte(data), modelType, rs.factory)
			if err != nil {
				continue // Skip invalid entries
			}

			// Apply filters (simple implementation)
			if rs.matchesCriteria(readModel, criteria) {
				results = append(results, readModel)
			}
		}

		// Apply sorting and pagination
		results = rs.applySortingAndPagination(results, criteria)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// Count returns the count of read models matching criteria
func (rs *RedisReadStore) Count(ctx context.Context, criteria QueryCriteria) (int64, error) {
	results, err := rs.Query(ctx, criteria)
	if err != nil {
		return 0, err
	}
	return int64(len(results)), nil
}

// SaveBatch saves multiple read models
func (rs *RedisReadStore) SaveBatch(ctx context.Context, readModels []ReadModel) error {
	if len(readModels) == 0 {
		return nil
	}

	return rs.client.ExecuteCommand(ctx, func() error {
		pipe := rs.client.GetClient().Pipeline()

		for _, readModel := range readModels {
			if readModel == nil {
				continue
			}

			if err := readModel.Validate(); err != nil {
				return NewCQRSError(ErrCodeRepositoryError.String(), "read model validation failed", err)
			}

			modelKey := rs.keyBuilder.ReadModelKey(readModel.GetType(), readModel.GetID())

			data, err := rs.serializer.SerializeReadModel(readModel)
			if err != nil {
				return NewCQRSError(ErrCodeSerializationError.String(), "failed to serialize read model", err)
			}

			pipe.Set(ctx, modelKey, data, 24*time.Hour)
		}

		_, err := pipe.Exec(ctx)
		if err != nil {
			return NewCQRSError(ErrCodeRepositoryError.String(), "failed to save read models batch", err)
		}

		// Update indexes for all models
		for _, readModel := range readModels {
			if readModel != nil {
				rs.updateIndexes(ctx, readModel)
			}
		}

		return nil
	})
}

// DeleteBatch deletes multiple read models
func (rs *RedisReadStore) DeleteBatch(ctx context.Context, ids []string, modelType string) error {
	if len(ids) == 0 {
		return nil
	}
	if modelType == "" {
		return NewCQRSError(ErrCodeRepositoryError.String(), "model type cannot be empty", nil)
	}

	return rs.client.ExecuteCommand(ctx, func() error {
		pipe := rs.client.GetClient().Pipeline()

		for _, id := range ids {
			if id == "" {
				continue
			}

			modelKey := rs.keyBuilder.ReadModelKey(modelType, id)
			pipe.Del(ctx, modelKey)
		}

		_, err := pipe.Exec(ctx)
		if err != nil {
			return NewCQRSError(ErrCodeRepositoryError.String(), "failed to delete read models batch", err)
		}

		return nil
	})
}

// CreateIndex creates an index for a model type and fields
func (rs *RedisReadStore) CreateIndex(ctx context.Context, modelType string, fields []string) error {
	// Note: This is a simplified implementation
	// In a real implementation, you'd create proper secondary indexes
	return nil
}

// DropIndex drops an index
func (rs *RedisReadStore) DropIndex(ctx context.Context, modelType string, indexName string) error {
	// Note: This is a simplified implementation
	return nil
}

// Helper methods

func (rs *RedisReadStore) updateIndexes(ctx context.Context, readModel ReadModel) error {
	// Simple indexing by type
	typeIndexKey := rs.keyBuilder.IndexKey(readModel.GetType(), "all")
	return rs.client.GetClient().SAdd(ctx, typeIndexKey, readModel.GetID()).Err()
}

func (rs *RedisReadStore) removeFromIndexes(ctx context.Context, readModel ReadModel) error {
	// Remove from type index
	typeIndexKey := rs.keyBuilder.IndexKey(readModel.GetType(), "all")
	return rs.client.GetClient().SRem(ctx, typeIndexKey, readModel.GetID()).Err()
}

func (rs *RedisReadStore) matchesCriteria(readModel ReadModel, criteria QueryCriteria) bool {
	// Simple criteria matching - in real implementation, this would be more sophisticated
	if len(criteria.Filters) == 0 {
		return true
	}

	for key, value := range criteria.Filters {
		if key == "type" && readModel.GetType() != fmt.Sprintf("%v", value) {
			return false
		}
		if key == "id" && readModel.GetID() != fmt.Sprintf("%v", value) {
			return false
		}
	}

	return true
}

func (rs *RedisReadStore) applySortingAndPagination(results []ReadModel, criteria QueryCriteria) []ReadModel {
	// Simple pagination implementation
	if criteria.Limit > 0 {
		start := criteria.Offset
		end := start + criteria.Limit

		if start >= len(results) {
			return []ReadModel{}
		}

		if end > len(results) {
			end = len(results)
		}

		return results[start:end]
	}

	return results
}

// JSONReadModelSerializer implementation

func (s *JSONReadModelSerializer) SerializeReadModel(model ReadModel) ([]byte, error) {
	modelData := ReadModelData{
		ID:          model.GetID(),
		Type:        model.GetType(),
		Version:     model.GetVersion(),
		Data:        model.GetData(),
		LastUpdated: model.GetLastUpdated(),
	}

	return json.Marshal(modelData)
}

func (s *JSONReadModelSerializer) DeserializeReadModel(data []byte, modelType string, factory ReadModelFactory) (ReadModel, error) {
	var modelData ReadModelData
	if err := json.Unmarshal(data, &modelData); err != nil {
		return nil, err
	}

	// Use factory to create the correct type
	if factory != nil {
		readModel, err := factory.CreateReadModel(modelType, modelData.ID, modelData.Data)
		if err != nil {
			// Fallback to BaseReadModel if factory fails
			readModel = NewBaseReadModel(modelData.ID, modelData.Type, modelData.Data)
		}

		// Set version and last updated if it's a BaseReadModel
		if baseModel, ok := readModel.(*BaseReadModel); ok {
			baseModel.SetVersion(modelData.Version)
			baseModel.SetLastUpdated(modelData.LastUpdated)
		}
		return readModel, nil
	}

	// Fallback to BaseReadModel
	readModel := NewBaseReadModel(modelData.ID, modelData.Type, modelData.Data)
	readModel.SetVersion(modelData.Version)
	readModel.SetLastUpdated(modelData.LastUpdated)

	return readModel, nil
}

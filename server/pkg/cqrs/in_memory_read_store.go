package cqrs

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// InMemoryReadStore provides an in-memory implementation of ReadStore
type InMemoryReadStore struct {
	models  map[string]ReadModel // key: "type:id"
	indexes map[string]map[string][]string // type -> field -> values -> ids
	mutex   sync.RWMutex
}

// NewInMemoryReadStore creates a new in-memory read store
func NewInMemoryReadStore() *InMemoryReadStore {
	return &InMemoryReadStore{
		models:  make(map[string]ReadModel),
		indexes: make(map[string]map[string][]string),
	}
}

// ReadStore interface implementation

func (rs *InMemoryReadStore) Save(ctx context.Context, readModel ReadModel) error {
	if readModel == nil {
		return NewCQRSError(ErrCodeRepositoryError.String(), "read model cannot be nil", nil)
	}

	if err := readModel.Validate(); err != nil {
		return NewCQRSError(ErrCodeRepositoryError.String(), "read model validation failed", err)
	}

	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	key := rs.getModelKey(readModel.GetType(), readModel.GetID())
	rs.models[key] = readModel

	return nil
}

func (rs *InMemoryReadStore) GetByID(ctx context.Context, id string, modelType string) (ReadModel, error) {
	if id == "" {
		return nil, NewCQRSError(ErrCodeRepositoryError.String(), "id cannot be empty", nil)
	}
	if modelType == "" {
		return nil, NewCQRSError(ErrCodeRepositoryError.String(), "model type cannot be empty", nil)
	}

	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	key := rs.getModelKey(modelType, id)
	if model, exists := rs.models[key]; exists {
		return model, nil
	}

	return nil, NewCQRSError(ErrCodeRepositoryError.String(), fmt.Sprintf("read model not found: %s:%s", modelType, id), nil)
}

func (rs *InMemoryReadStore) Delete(ctx context.Context, id string, modelType string) error {
	if id == "" {
		return NewCQRSError(ErrCodeRepositoryError.String(), "id cannot be empty", nil)
	}
	if modelType == "" {
		return NewCQRSError(ErrCodeRepositoryError.String(), "model type cannot be empty", nil)
	}

	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	key := rs.getModelKey(modelType, id)
	if _, exists := rs.models[key]; !exists {
		return NewCQRSError(ErrCodeRepositoryError.String(), fmt.Sprintf("read model not found: %s:%s", modelType, id), nil)
	}

	delete(rs.models, key)
	return nil
}

func (rs *InMemoryReadStore) Query(ctx context.Context, criteria QueryCriteria) ([]ReadModel, error) {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	var results []ReadModel

	// Simple filtering implementation
	for _, model := range rs.models {
		if rs.matchesCriteria(model, criteria) {
			results = append(results, model)
		}
	}

	// Apply sorting (simple implementation)
	if criteria.SortBy != "" {
		// Note: In a real implementation, you would implement proper sorting
		// For now, we'll just return the results as-is
	}

	// Apply pagination
	if criteria.Limit > 0 {
		start := criteria.Offset
		end := start + criteria.Limit

		if start >= len(results) {
			return []ReadModel{}, nil
		}

		if end > len(results) {
			end = len(results)
		}

		results = results[start:end]
	}

	return results, nil
}

func (rs *InMemoryReadStore) Count(ctx context.Context, criteria QueryCriteria) (int64, error) {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	count := int64(0)

	for _, model := range rs.models {
		if rs.matchesCriteria(model, criteria) {
			count++
		}
	}

	return count, nil
}

func (rs *InMemoryReadStore) SaveBatch(ctx context.Context, readModels []ReadModel) error {
	if len(readModels) == 0 {
		return nil
	}

	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	for _, model := range readModels {
		if model == nil {
			return NewCQRSError(ErrCodeRepositoryError.String(), "read model cannot be nil", nil)
		}

		if err := model.Validate(); err != nil {
			return NewCQRSError(ErrCodeRepositoryError.String(), "read model validation failed", err)
		}

		key := rs.getModelKey(model.GetType(), model.GetID())
		rs.models[key] = model
	}

	return nil
}

func (rs *InMemoryReadStore) DeleteBatch(ctx context.Context, ids []string, modelType string) error {
	if len(ids) == 0 {
		return nil
	}
	if modelType == "" {
		return NewCQRSError(ErrCodeRepositoryError.String(), "model type cannot be empty", nil)
	}

	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	for _, id := range ids {
		if id == "" {
			continue
		}

		key := rs.getModelKey(modelType, id)
		delete(rs.models, key)
	}

	return nil
}

func (rs *InMemoryReadStore) CreateIndex(ctx context.Context, modelType string, fields []string) error {
	if modelType == "" {
		return NewCQRSError(ErrCodeRepositoryError.String(), "model type cannot be empty", nil)
	}
	if len(fields) == 0 {
		return NewCQRSError(ErrCodeRepositoryError.String(), "fields cannot be empty", nil)
	}

	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	if _, exists := rs.indexes[modelType]; !exists {
		rs.indexes[modelType] = make(map[string][]string)
	}

	// Simple index creation (just mark fields as indexed)
	for _, field := range fields {
		rs.indexes[modelType][field] = make([]string, 0)
	}

	return nil
}

func (rs *InMemoryReadStore) DropIndex(ctx context.Context, modelType string, indexName string) error {
	if modelType == "" {
		return NewCQRSError(ErrCodeRepositoryError.String(), "model type cannot be empty", nil)
	}
	if indexName == "" {
		return NewCQRSError(ErrCodeRepositoryError.String(), "index name cannot be empty", nil)
	}

	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	if typeIndexes, exists := rs.indexes[modelType]; exists {
		delete(typeIndexes, indexName)
	}

	return nil
}

// Helper methods

func (rs *InMemoryReadStore) getModelKey(modelType, id string) string {
	return fmt.Sprintf("%s:%s", modelType, id)
}

func (rs *InMemoryReadStore) matchesCriteria(model ReadModel, criteria QueryCriteria) bool {
	// Simple criteria matching implementation
	// In a real implementation, you would have more sophisticated filtering

	if len(criteria.Filters) == 0 {
		return true
	}

	// For demonstration, we'll check if the model data contains the filter values
	// This is a very basic implementation
	for key, value := range criteria.Filters {
		if !rs.modelContainsValue(model, key, value) {
			return false
		}
	}

	return true
}

func (rs *InMemoryReadStore) modelContainsValue(model ReadModel, key string, value interface{}) bool {
	// Very basic implementation - in reality, you'd need proper field access
	// This is just for demonstration purposes
	
	// Check if the key matches model type or ID
	if key == "type" && model.GetType() == fmt.Sprintf("%v", value) {
		return true
	}
	if key == "id" && model.GetID() == fmt.Sprintf("%v", value) {
		return true
	}

	// For other fields, we'd need to inspect the model data
	// This would require reflection or a more sophisticated approach
	return false
}

// GetModelCount returns the total number of stored models
func (rs *InMemoryReadStore) GetModelCount() int {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	return len(rs.models)
}

// GetModelsByType returns all models of a specific type
func (rs *InMemoryReadStore) GetModelsByType(modelType string) []ReadModel {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	var models []ReadModel
	prefix := modelType + ":"

	for key, model := range rs.models {
		if strings.HasPrefix(key, prefix) {
			models = append(models, model)
		}
	}

	return models
}

// Clear removes all models and indexes
func (rs *InMemoryReadStore) Clear() {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	rs.models = make(map[string]ReadModel)
	rs.indexes = make(map[string]map[string][]string)
}

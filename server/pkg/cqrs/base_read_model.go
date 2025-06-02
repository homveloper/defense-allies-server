package cqrs

import (
	"fmt"
	"time"
)

// BaseReadModel provides a base implementation of ReadModel interface
type BaseReadModel struct {
	id          string
	modelType   string
	version     int
	data        interface{}
	lastUpdated time.Time
}

// NewBaseReadModel creates a new BaseReadModel
func NewBaseReadModel(id, modelType string, data interface{}) *BaseReadModel {
	return &BaseReadModel{
		id:          id,
		modelType:   modelType,
		version:     1,
		data:        data,
		lastUpdated: time.Now(),
	}
}

// ReadModel interface implementation

func (rm *BaseReadModel) GetID() string {
	return rm.id
}

func (rm *BaseReadModel) GetType() string {
	return rm.modelType
}

func (rm *BaseReadModel) GetVersion() int {
	return rm.version
}

func (rm *BaseReadModel) GetData() interface{} {
	return rm.data
}

func (rm *BaseReadModel) GetLastUpdated() time.Time {
	return rm.lastUpdated
}

func (rm *BaseReadModel) Validate() error {
	if rm.id == "" {
		return fmt.Errorf("read model ID cannot be empty")
	}
	if rm.modelType == "" {
		return fmt.Errorf("read model type cannot be empty")
	}
	if rm.data == nil {
		return fmt.Errorf("read model data cannot be nil")
	}
	return nil
}

// Helper methods

// SetData sets the read model data and updates the version and timestamp
func (rm *BaseReadModel) SetData(data interface{}) {
	rm.data = data
	rm.version++
	rm.lastUpdated = time.Now()
}

// SetVersion sets the version (used when loading from storage)
func (rm *BaseReadModel) SetVersion(version int) {
	rm.version = version
}

// SetLastUpdated sets the last updated timestamp (used when loading from storage)
func (rm *BaseReadModel) SetLastUpdated(lastUpdated time.Time) {
	rm.lastUpdated = lastUpdated
}

// UpdateData updates the read model data without incrementing version
func (rm *BaseReadModel) UpdateData(data interface{}) {
	rm.data = data
	rm.lastUpdated = time.Now()
}

// IncrementVersion increments the version and updates timestamp
func (rm *BaseReadModel) IncrementVersion() {
	rm.version++
	rm.lastUpdated = time.Now()
}

// GetModelInfo returns basic read model information as a map
func (rm *BaseReadModel) GetModelInfo() map[string]interface{} {
	return map[string]interface{}{
		"id":           rm.id,
		"type":         rm.modelType,
		"version":      rm.version,
		"last_updated": rm.lastUpdated,
	}
}

// Clone creates a copy of the read model
func (rm *BaseReadModel) Clone() *BaseReadModel {
	return &BaseReadModel{
		id:          rm.id,
		modelType:   rm.modelType,
		version:     rm.version,
		data:        rm.data, // Note: This is a shallow copy
		lastUpdated: rm.lastUpdated,
	}
}

package cqrs

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
)

// EventDataRegistry manages event data type registration and validation
type EventDataRegistry struct {
	eventTypes map[string]reflect.Type
	mu         sync.RWMutex
}

// NewEventDataRegistry creates a new event data registry
func NewEventDataRegistry() *EventDataRegistry {
	return &EventDataRegistry{
		eventTypes: make(map[string]reflect.Type),
	}
}

// RegisterEventData registers an event data type for the given event type
// The eventData parameter should be a zero value of the event data struct
// The function will store the pointer type for consistent deserialization
func (r *EventDataRegistry) RegisterEventData(eventType string, eventData interface{}) error {
	if eventType == "" {
		return NewCQRSError(ErrCodeValidationError.String(), "event type cannot be empty", nil)
	}

	if eventData == nil {
		return NewCQRSError(ErrCodeValidationError.String(), "event data cannot be nil", nil)
	}

	// Validate that the event data can be marshaled and unmarshaled
	if err := r.validateJSONSerialization(eventData); err != nil {
		return NewCQRSError(ErrCodeValidationError.String(),
			fmt.Sprintf("event data validation failed for type %s: %v", eventType, err), err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	dataType := reflect.TypeOf(eventData)

	// Always store as pointer type for consistent deserialization
	if dataType.Kind() != reflect.Ptr {
		dataType = reflect.PointerTo(dataType)
	}

	r.eventTypes[eventType] = dataType

	return nil
}

// GetEventDataType returns the registered type for the given event type
func (r *EventDataRegistry) GetEventDataType(eventType string) (reflect.Type, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dataType, exists := r.eventTypes[eventType]
	if !exists {
		return nil, NewCQRSError(ErrCodeNotFoundError.String(),
			fmt.Sprintf("event type not registered: %s", eventType), nil)
	}

	return dataType, nil
}

// IsRegistered checks if an event type is registered
func (r *EventDataRegistry) IsRegistered(eventType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.eventTypes[eventType]
	return exists
}

// GetRegisteredEventTypes returns all registered event types
func (r *EventDataRegistry) GetRegisteredEventTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	eventTypes := make([]string, 0, len(r.eventTypes))
	for eventType := range r.eventTypes {
		eventTypes = append(eventTypes, eventType)
	}

	return eventTypes
}

// CreateEventDataInstance creates a new instance of the registered event data type
// Always returns a pointer to the struct
func (r *EventDataRegistry) CreateEventDataInstance(eventType string) (interface{}, error) {
	dataType, err := r.GetEventDataType(eventType)
	if err != nil {
		return nil, err
	}

	// Since we always store pointer types, create new instance
	return reflect.New(dataType.Elem()).Interface(), nil
}

// validateJSONSerialization validates that the event data can be properly serialized and deserialized
func (r *EventDataRegistry) validateJSONSerialization(eventData interface{}) error {
	// Test JSON marshaling
	jsonData, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	// Test JSON unmarshaling
	dataType := reflect.TypeOf(eventData)
	var target interface{}

	if dataType.Kind() == reflect.Ptr {
		target = reflect.New(dataType.Elem()).Interface()
	} else {
		target = reflect.New(dataType).Interface()
	}

	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("failed to unmarshal from JSON: %w", err)
	}

	return nil
}

// UnregisterEventData removes an event type from the registry
func (r *EventDataRegistry) UnregisterEventData(eventType string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.eventTypes, eventType)
}

// Clear removes all registered event types
func (r *EventDataRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.eventTypes = make(map[string]reflect.Type)
}

// GetRegistryStats returns statistics about the registry
func (r *EventDataRegistry) GetRegistryStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return map[string]interface{}{
		"total_registered_types": len(r.eventTypes),
		"registered_types":       r.GetRegisteredEventTypes(),
	}
}

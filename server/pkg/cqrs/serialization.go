package cqrs

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// SerializeToJSON serializes an object to JSON bytes
func SerializeToJSON(obj interface{}) ([]byte, error) {
	if obj == nil {
		return nil, NewCQRSError(ErrCodeSerializationError.String(), "object cannot be nil", nil)
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, NewCQRSError(ErrCodeSerializationError.String(),
			fmt.Sprintf("failed to serialize to JSON: %v", err), err)
	}

	return data, nil
}

// DeserializeFromJSON deserializes JSON bytes to an object of the specified type
func DeserializeFromJSON(data []byte, targetType reflect.Type) (interface{}, error) {
	if len(data) == 0 {
		return nil, NewCQRSError(ErrCodeSerializationError.String(), "data cannot be empty", nil)
	}

	if targetType == nil {
		return nil, NewCQRSError(ErrCodeSerializationError.String(), "target type cannot be nil", nil)
	}

	// Create a new instance of the target type
	var target interface{}
	if targetType.Kind() == reflect.Ptr {
		target = reflect.New(targetType.Elem()).Interface()
	} else {
		target = reflect.New(targetType).Interface()
	}

	err := json.Unmarshal(data, target)
	if err != nil {
		return nil, NewCQRSError(ErrCodeSerializationError.String(),
			fmt.Sprintf("failed to deserialize from JSON: %v", err), err)
	}

	// Return the value if target type is not a pointer
	if targetType.Kind() != reflect.Ptr {
		return reflect.ValueOf(target).Elem().Interface(), nil
	}

	return target, nil
}

// DeserializeFromJSONInto deserializes JSON bytes into an existing object
func DeserializeFromJSONInto(data []byte, target interface{}) error {
	if len(data) == 0 {
		return NewCQRSError(ErrCodeSerializationError.String(), "data cannot be empty", nil)
	}

	if target == nil {
		return NewCQRSError(ErrCodeSerializationError.String(), "target cannot be nil", nil)
	}

	err := json.Unmarshal(data, target)
	if err != nil {
		return NewCQRSError(ErrCodeSerializationError.String(),
			fmt.Sprintf("failed to deserialize into target: %v", err), err)
	}

	return nil
}

// Note: ReadModel, QueryCriteria, and SortCriteria interfaces are defined elsewhere

// ReadModelType registry for read model types
var readModelTypes = make(map[string]reflect.Type)

// RegisterReadModelType registers a read model type
func RegisterReadModelType(typeName string, modelType reflect.Type) {
	readModelTypes[typeName] = modelType
}

// GetReadModelType gets a registered read model type
func GetReadModelType(typeName string) (reflect.Type, error) {
	modelType, exists := readModelTypes[typeName]
	if !exists {
		return nil, NewCQRSError(ErrCodeSerializationError.String(),
			fmt.Sprintf("read model type not registered: %s", typeName), nil)
	}
	return modelType, nil
}

// AggregateFactory function type for creating aggregate instances
type AggregateFactory func(id string) (AggregateRoot, error)

// AggregateType registry for aggregate factories
var aggregateFactories = make(map[string]AggregateFactory)

// RegisterAggregateType registers an aggregate factory
func RegisterAggregateType(typeName string, factory AggregateFactory) {
	aggregateFactories[typeName] = factory
}

// CreateAggregateInstance creates an aggregate instance of the specified type
func CreateAggregateInstance(aggregateType, id string) (AggregateRoot, error) {
	factory, exists := aggregateFactories[aggregateType]
	if !exists {
		return nil, NewCQRSError(ErrCodeSerializationError.String(),
			fmt.Sprintf("aggregate type not registered: %s", aggregateType), nil)
	}

	aggregate, err := factory(id)
	if err != nil {
		return nil, NewCQRSError(ErrCodeSerializationError.String(),
			fmt.Sprintf("failed to create aggregate instance: %v", err), err)
	}

	return aggregate, nil
}

// Note: EventSerializer interfaces and implementations are now in cqrsx/event_serializer.go

// Note: Event serialization implementations are now in cqrsx/event_serializer.go

// Note: Event factory functions are available in the core CQRS package if needed

package cqrsx

import (
	"fmt"
	"reflect"
	"sync"
)

// import (
// 	"encoding/json"
// 	"fmt"
// 	"reflect"
// )

// =============================================================================
// BucketDataRegistry - 이벤트 타입별 강타입 저장소
// =============================================================================

type EventRegistry interface {
	// 이벤트 타입과 데이터 구조체 타입 등록
	RegisterEventDataType(eventType string, dataType reflect.Type) error

	// 이벤트 타입으로 새 데이터 인스턴스 생성
	CreateDataInstance(eventType string) (interface{}, error)

	// 이벤트 타입의 데이터 타입 조회
	GetDataType(eventType string) (reflect.Type, error)

	// 등록된 모든 이벤트 타입 조회
	GetRegisteredEventTypes() []string

	// 이벤트 타입이 등록되어 있는지 확인
	IsRegistered(eventType string) bool
}

type InMemoryEventRegistry struct {
	mu             sync.RWMutex            // 읽기/쓰기 동시성 보장
	eventDataTypes map[string]reflect.Type // eventType -> reflect.Type
}

func NewInMemoryEventRegistry() *InMemoryEventRegistry {
	return &InMemoryEventRegistry{
		eventDataTypes: make(map[string]reflect.Type),
	}
}

func (r *InMemoryEventRegistry) RegisterEventDataType(eventType string, dataType reflect.Type) error {
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}

	// 포인터 타입인 경우 실제 타입 추출
	if dataType.Kind() == reflect.Ptr {
		dataType = dataType.Elem()
	}

	if dataType.Kind() != reflect.Struct {
		return fmt.Errorf("data type must be a struct, got %v", dataType.Kind())
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 이미 등록된 이벤트 타입인지 확인
	if existingType, exists := r.eventDataTypes[eventType]; exists {
		if existingType == dataType {
			// 같은 타입으로 재등록하는 경우는 무시 (멱등성 보장)
			return nil
		}
		// 다른 타입으로 재등록하려는 경우 에러 반환
		return fmt.Errorf("event type '%s' is already registered with type %v, cannot register with different type %v",
			eventType, existingType, dataType)
	}

	r.eventDataTypes[eventType] = dataType
	return nil
}

func (r *InMemoryEventRegistry) CreateDataInstance(eventType string) (interface{}, error) {
	r.mu.RLock()
	dataType, exists := r.eventDataTypes[eventType]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}

	// 새 포인터 인스턴스 생성
	instance := reflect.New(dataType).Interface()
	return instance, nil
}

func (r *InMemoryEventRegistry) GetDataType(eventType string) (reflect.Type, error) {
	r.mu.RLock()
	dataType, exists := r.eventDataTypes[eventType]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}
	return dataType, nil
}

func (r *InMemoryEventRegistry) GetRegisteredEventTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var eventTypes []string
	for eventType := range r.eventDataTypes {
		eventTypes = append(eventTypes, eventType)
	}
	return eventTypes
}

func (r *InMemoryEventRegistry) IsRegistered(eventType string) bool {
	r.mu.RLock()
	_, exists := r.eventDataTypes[eventType]
	r.mu.RUnlock()
	return exists
}

// 편의 메서드: 구조체로 직접 등록
func (r *InMemoryEventRegistry) RegisterDataStruct(eventType string, dataStruct interface{}) error {
	dataType := reflect.TypeOf(dataStruct)
	return r.RegisterEventDataType(eventType, dataType)
}

// // =============================================================================
// // BucketSerializer - EventBucketData.Data 필드 강타입 변환
// // =============================================================================

// type BucketSerializer interface {
// 	// EventBucketData.Data 필드 강타입 복원
// 	DeserializeEventData(eventBucketData EventBucketMessage) (EventBucketMessage, error)

// 	// 강타입 데이터를 interface{}로 직렬화 (저장용)
// 	SerializeEventData(eventType string, typedData interface{}) (EventBucketData, error)

// 	// 배치 처리
// 	DeserializeEventDataBatch(eventBucketDataList []EventBucketMessage) ([]EventBucketMessage, error)
// }

// type JSONBucketSerializer struct {
// 	registry BucketDataRegistry
// }

// func NewJSONBucketSerializer(registry BucketDataRegistry) *JSONBucketSerializer {
// 	return &JSONBucketSerializer{
// 		registry: registry,
// 	}
// }

// // EventBucketData의 Data 필드를 강타입으로 복원
// func (s *JSONBucketSerializer) DeserializeEventData(eventBucketData EventBucketMessage) (EventBucketMessage, error) {
// 	if eventBucketData.Data == nil {
// 		return eventBucketData, nil // Data가 nil이면 그대로 반환
// 	}

// 	// 등록된 타입으로 새 인스턴스 생성
// 	typedDataInstance, err := s.registry.CreateDataInstance(eventBucketData.EventType)
// 	if err != nil {
// 		return eventBucketData, fmt.Errorf("failed to create data instance for %s: %w",
// 			eventBucketData.EventType, err)
// 	}

// 	// JSON을 통한 역직렬화
// 	jsonBytes, err := json.Marshal(eventBucketData.Data)
// 	if err != nil {
// 		return eventBucketData, fmt.Errorf("failed to marshal data for %s: %w",
// 			eventBucketData.EventType, err)
// 	}

// 	if err := json.Unmarshal(jsonBytes, typedDataInstance); err != nil {
// 		return eventBucketData, fmt.Errorf("failed to unmarshal to %s: %w",
// 			eventBucketData.EventType, err)
// 	}

// 	// 새로운 EventBucketData 생성 (Data 필드만 강타입으로 교체)
// 	result := NewEventBucketMessage(
// 		eventBucketData.EventID,
// 		eventBucketData.EventType,
// 		eventBucketData.AggregateID,
// 		typedDataInstance, // 강타입 데이터로 교체
// 		eventBucketData.Metadata,
// 		eventBucketData.OccurredAt,
// 		eventBucketData.Version,
// 	)

// 	return result, nil
// }

// // 강타입 데이터를 저장용 interface{}로 직렬화
// func (s *JSONBucketSerializer) SerializeEventData(eventType string, typedData interface{}) (EventBucketData, error) {
// 	if typedData == nil {
// 		return nil, fmt.Errorf("typed data cannot be nil")
// 	}

// 	// 이벤트 타입이 등록되어 있는지 확인
// 	if !s.registry.IsRegistered(eventType) {
// 		return nil, fmt.Errorf("unregistered event type: %s", eventType)
// 	}

// 	// JSON을 통한 직렬화 (map[string]interface{} 형태로 변환)
// 	jsonBytes, err := json.Marshal(typedData)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal typed data: %w", err)
// 	}

// 	var serializedData interface{}
// 	if err := json.Unmarshal(jsonBytes, &serializedData); err != nil {
// 		return nil, fmt.Errorf("failed to unmarshal to interface{}: %w", err)
// 	}

// 	return serializedData, nil
// }

// // 배치 처리
// func (s *JSONBucketSerializer) DeserializeEventDataBatch(eventBucketDataList []EventBucketMessage) ([]EventBucketMessage, error) {
// 	var results = make([]EventBucketMessage, 0, len(eventBucketDataList))

// 	for i, eventBucketData := range eventBucketDataList {
// 		deserializedEvent, err := s.DeserializeEventData(eventBucketData)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to deserialize event at index %d: %w", i, err)
// 		}
// 		results = append(results, deserializedEvent)
// 	}

// 	return results, nil
// }

// // BsonBucketSerializer는 BSON을 사용하는 버킷 직렬화 구현체입니다.
// // MongoDB의 BSON을 사용하므로 MongoDB와의 호환성이 보장됩니다.
// type BsonBucketSerializer struct {
// 	registry BucketDataRegistry
// }

// func NewBsonBucketSerializer(registry BucketDataRegistry) *BsonBucketSerializer {
// 	return &BsonBucketSerializer{
// 		registry: registry,
// 	}
// }

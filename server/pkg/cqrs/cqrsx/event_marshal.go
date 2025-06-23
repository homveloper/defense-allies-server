package cqrsx

import (
	"cqrs"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

type metadata struct {
	EventID       string                 `json:"eventId" bson:"eventId"`
	EventType     string                 `json:"eventType" bson:"eventType"`
	AggregateID   string                 `json:"aggregateId" bson:"aggregateId"`
	AggregateType string                 `json:"aggregateType" bson:"aggregateType"`
	Version       int                    `json:"version" bson:"version"`
	Metadata      map[string]interface{} `json:"metadata" bson:"metadata"`
	Timestamp     time.Time              `json:"timestamp" bson:"timestamp"`
}

func MapEvent(event cqrs.EventMessage) (map[string]interface{}, error) {
	meta := metadata{
		EventID:       event.EventID(),
		EventType:     event.EventType(),
		AggregateID:   event.AggregateID(),
		AggregateType: event.AggregateType(),
		Version:       event.Version(),
		Metadata:      event.Metadata(),
		Timestamp:     event.Timestamp(),
	}

	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal metadata")
	}

	var metadataMap map[string]interface{}
	if err := json.Unmarshal(metaBytes, &metadataMap); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal metadata to map for merging")
	}

	// event data payload to marshal
	payloadBytes, err := json.Marshal(event)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal payload")
	}

	var payloadMap map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payloadMap); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal payload to map for merging")
	}

	// 메타데이터 맵과 페이로드 맵을 병합
	// 페이로드의 키가 메타데이터 키와 충돌하면, 페이로드가 우선됨.
	for k, v := range payloadMap {
		metadataMap[k] = v
	}

	return metadataMap, nil
}

func MarshalEventJSON(event cqrs.EventMessage) ([]byte, error) {
	// // temporal metadata
	// meta := metadata{
	// 	EventID:       event.EventID(),
	// 	EventType:     event.EventType(),
	// 	AggregateID:   event.ID(),
	// 	AggregateType: event.Type(),
	// 	Version:       event.Version(),
	// 	Metadata:      event.Metadata(),
	// 	Timestamp:     event.Timestamp(),
	// }

	// metaBytes, err := json.Marshal(meta)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "failed to marshal metadata")
	// }

	// var metadataMap map[string]interface{}
	// if err := json.Unmarshal(metaBytes, &metadataMap); err != nil {
	// 	return nil, errors.Wrap(err, "failed to unmarshal metadata to map for merging")
	// }

	// // event data payload to marshal
	// payloadBytes, err := json.Marshal(event)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "failed to marshal payload")
	// }

	// var payloadMap map[string]interface{}
	// if err := json.Unmarshal(payloadBytes, &payloadMap); err != nil {
	// 	return nil, errors.Wrap(err, "failed to unmarshal payload to map for merging")
	// }

	// // 메타데이터 맵과 페이로드 맵을 병합
	// // 페이로드의 키가 메타데이터 키와 충돌하면, 페이로드가 우선됨.
	// for k, v := range payloadMap {
	// 	metadataMap[k] = v
	// }

	// // 4. 최종 병합된 맵을 직렬화

	// return json.Marshal(metadataMap)

	return json.Marshal(event)
}

var typeExtractor struct {
	EventType string `json:"eventType" bson:"eventType"`
}

// UnmarshalEventJSON은 JSON 바이트를 올바른 타입의 이벤트 객체로 역직렬화합니다.
// BucketDataRegistry를 사용하여 동적으로 타입을 결정합니다.
func UnmarshalEventJSON(data []byte, registry EventRegistry) (cqrs.EventMessage, error) {
	// 1. eventType 필드만 추출

	if err := json.Unmarshal(data, &typeExtractor); err != nil {
		return nil, fmt.Errorf("failed to extract eventType from JSON: %w", err)
	}
	if typeExtractor.EventType == "" {
		return nil, fmt.Errorf("eventType is missing in the JSON data")
	}

	// 2. 레지스트리를 사용하여 해당 eventType에 맞는 빈 이벤트 객체(포인터)를 생성
	instance, err := registry.CreateDataInstance(typeExtractor.EventType)
	if err != nil {
		return nil, err // 레지스트리에 등록되지 않은 이벤트 타입
	}

	// 3. 완전한 데이터를 새로 생성된 구체적인 타입의 객체에 역직렬화
	if err := json.Unmarshal(data, instance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal full JSON data into event struct: %w", err)
	}

	// // 4. 결과가 EventMessage 인터페이스를 만족하는지 확인하고 반환
	// if eventMessage, ok := instance.(cqrs.EventMessage); ok {
	// 	cqrs.RehydrateEventMessage(eventMessage,
	// 		eventMessage.EventID(),
	// 		eventMessage.EventType(),
	// 		eventMessage.ID(),
	// 		eventMessage.Type(),
	// 		eventMessage.Version(),
	// 		eventMessage.Metadata(),
	// 		eventMessage.Timestamp(),
	// 	)

	// 	return eventMessage, nil
	// }

	if eventMessage, ok := instance.(cqrs.EventMessage); ok {
		return eventMessage, nil
	}

	return nil, fmt.Errorf("unmarshaled event of type '%s' does not implement EventMessage interface", typeExtractor.EventType)
}

// MarshalEventBSON은 EventMessage를 평평한(flat) BSON 구조로 직렬화합니다.
func MarshalEventBSON(e cqrs.EventMessage) ([]byte, error) {
	// 1. 메타데이터를 생성
	meta := metadata{
		EventID:       e.EventID(),
		EventType:     e.EventType(),
		AggregateID:   e.AggregateID(),
		AggregateType: e.AggregateType(),
		Version:       e.Version(),
		Metadata:      e.Metadata(),
		Timestamp:     e.Timestamp(),
	}

	metaBytes, err := bson.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata to BSON: %w", err)
	}

	var metadataMap map[string]interface{}
	if err := bson.Unmarshal(metaBytes, &metadataMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata to map for BSON merging: %w", err)
	}

	// 2. 페이로드(EventData)를 가져와 맵으로 변환
	payloadBytes, err := bson.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event payload to BSON: %w", err)
	}

	var payloadMap map[string]interface{}
	if err := bson.Unmarshal(payloadBytes, &payloadMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload to map for BSON merging: %w", err)
	}

	// 3. 메타데이터 맵과 페이로드 맵을 병합
	for k, v := range payloadMap {
		metadataMap[k] = v
	}

	// 4. 최종 병합된 맵을 직렬화
	return bson.Marshal(metadataMap)
}

// UnmarshalEventBSON은 BSON 바이트를 올바른 타입의 이벤트 객체로 역직렬화합니다.
// 개선된 버전: bson.D 문제를 해결하기 위해 JSON 변환을 통한 우회 방식을 사용합니다.
func UnmarshalEventBSON(data []byte, registry EventRegistry) (cqrs.EventMessage, error) {
	// eventType 필드 추출
	meta := metadata{}
	if err := bson.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to extract eventType from BSON: %w", err)
	}
	if meta.EventType == "" {
		return nil, fmt.Errorf("eventType is missing in the BSON data")
	}

	// 레지스트리를 사용하여 인스턴스 생성
	instance, err := registry.CreateDataInstance(meta.EventType)
	if err != nil {
		return nil, err
	}

	// bson Unmarshal
	if err := bson.Unmarshal(data, instance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal BSON to typed instance: %w", err)
	}

	// 인터페이스 만족 확인 및 반환
	if eventMessage, ok := instance.(cqrs.EventMessage); ok {
		return eventMessage, nil
	}

	return nil, fmt.Errorf("unmarshaled BSON event of type '%s' does not implement EventMessage interface", typeExtractor.EventType)
}

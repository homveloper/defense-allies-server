package cqrs

import (
	"time"

	"github.com/google/uuid"
)

var _ EventMessage = (*BaseEventMessage)(nil)

// BaseEventMessage는 EventMessage 인터페이스의 기본 구현을 제공합니다.
// 구체적인 이벤트 구조체에 '값으로' 임베딩하여 사용합니다.
type BaseEventMessage struct {
	EventID_       string                 `json:"eventId" bson:"eventId"`
	EventType_     string                 `json:"eventType" bson:"eventType"`
	AggregateID_   string                 `json:"aggregateId" bson:"aggregateId"`
	AggregateType_ string                 `json:"aggregateType" bson:"aggregateType"`
	Version_       int                    `json:"version" bson:"version"`
	Metadata_      map[string]interface{} `json:"metadata" bson:"metadata"`
	Timestamp_     time.Time              `json:"timestamp" bson:"timestamp"`
}

// func (b *BaseEventMessage) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(EventMetaData{
// 		EventID:       b.eventID,
// 		EventType:     b.eventType,
// 		AggregateID:   b.aggregateID,
// 		AggregateType: b.aggregateType,
// 		Version:       b.version,
// 		Metadata:      b.metadata,
// 		Timestamp:     b.timestamp,
// 	})
// }

// func (b *BaseEventMessage) UnmarshalJSON(data []byte) error {
// 	var meta EventMetaData
// 	if err := json.Unmarshal(data, &meta); err != nil {
// 		return err
// 	}

// 	type Alias BaseEventMessage
// 	alias := &struct {
// 		*Alias `json:",inline" bson:",inline"`
// 	}{
// 		Alias: (*Alias)(b),
// 	}

// 	if err := json.Unmarshal(data, &alias); err != nil {
// 		return err
// 	}

// 	if b == nil {
// 		b = &BaseEventMessage{}
// 	}

// 	b.rehydrate(
// 		meta.EventID,
// 		meta.EventType,
// 		meta.AggregateID,
// 		meta.AggregateType,
// 		meta.Version,
// 		meta.Metadata,
// 		meta.Timestamp,
// 	)

// 	return nil
// }

// NewBaseEventMessage는 이벤트 타입만으로 기본 메타데이터를 가진 메시지를 생성합니다.
// 나머지 메타데이터(Aggregate 정보)는 Aggregate.Apply에서 채워집니다.
func NewBaseEventMessage(eventType string) *BaseEventMessage {
	return &BaseEventMessage{
		EventID_:   uuid.NewString(), // New().String() 보다 빠름
		EventType_: eventType,
		Timestamp_: time.Now().UTC(), // 항상 UTC 사용 권장
		Metadata_:  make(map[string]interface{}),

		// 나머지 필드는 기본값으로 초기화됩니다.
		// Aggregate 정보는 Apply에서 채워집니다.
		AggregateID_:   "",
		AggregateType_: "",
		Version_:       0,
	}
}

// EventMessage interface implementation
// --- EventMessage 인터페이스 구현 ---

func (e BaseEventMessage) EventID() string {
	return e.EventID_
}

func (e BaseEventMessage) EventType() string {
	return e.EventType_
}

func (e BaseEventMessage) AggregateID() string {
	return e.AggregateID_
}

func (e BaseEventMessage) AggregateType() string {
	return e.AggregateType_
}

func (e BaseEventMessage) Version() int {
	return e.Version_
}

// EventData는 기본적으로 nil을 반환합니다.
// 구체적인 이벤트 구조체에서 이 메서드를 오버라이드(override)해야 합니다.
func (e BaseEventMessage) EventData() interface{} {
	return nil
}

func (e BaseEventMessage) Metadata() map[string]interface{} {
	return e.Metadata_
}

func (e BaseEventMessage) Timestamp() time.Time {
	return e.Timestamp_
}

func (e *BaseEventMessage) setAggregateInfo(aggregateID string, aggregateType string, version int) {
	e.AggregateID_ = aggregateID
	e.AggregateType_ = aggregateType
	e.Version_ = version
}

// AddMetadata adds metadata to the event
func (e *BaseEventMessage) AddMetadata(key string, value interface{}) {
	if e.Metadata_ == nil {
		e.Metadata_ = make(map[string]interface{})
	}
	e.Metadata_[key] = value
}

// func (e *BaseEventMessage) rehydrate(
// 	eventID string,
// 	eventType string,
// 	aggregateID string,
// 	aggregateType string,
// 	version int,
// 	metadata map[string]interface{},
// 	timestamp time.Time,
// ) {
// 	e.EventID_ = eventID
// 	e.EventType_ = eventType
// 	e.AggregateID_ = aggregateID
// 	e.AggregateType_ = aggregateType
// 	e.Version_ = version
// 	e.Metadata_ = metadata
// 	e.Timestamp_ = timestamp
// }

// func RehydrateEventMessage(
// 	e EventMessage,
// 	eventID string,
// 	eventType string,
// 	aggregateID string,
// 	aggregateType string,
// 	version int,
// 	metadata map[string]interface{},
// 	timestamp time.Time,
// ) {
// 	e.rehydrate(eventID, eventType, aggregateID, aggregateType, version, metadata, timestamp)
// }

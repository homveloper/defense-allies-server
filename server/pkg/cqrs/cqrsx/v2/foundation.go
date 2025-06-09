// foundation.go - 기본 타입과 인터페이스
package cqrsx

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// StorageStrategy는 이벤트 저장 전략을 정의합니다
type StorageStrategy string

const (
	StrategyStream   StorageStrategy = "stream"   // 이벤트 스트림 방식
	StrategyDocument StorageStrategy = "document" // 단일 이벤트 방식
	StrategyHybrid   StorageStrategy = "hybrid"   // 하이브리드 방식
)

// EventStore는 이벤트 저장소의 기본 인터페이스입니다
type EventStore interface {
	// Save는 이벤트들을 저장합니다
	Save(ctx context.Context, events []Event, expectedVersion int) error

	// Load는 집합체의 모든 이벤트를 로드합니다
	Load(ctx context.Context, aggregateID uuid.UUID) ([]Event, error)

	// LoadFrom은 특정 버전부터 이벤트를 로드합니다
	LoadFrom(ctx context.Context, aggregateID uuid.UUID, fromVersion int) ([]Event, error)

	// GetMetrics는 성능 메트릭을 반환합니다
	GetMetrics() StoreMetrics

	// Close는 연결을 정리합니다
	Close() error
}

// Event는 이벤트의 기본 인터페이스입니다
type Event interface {
	string() uuid.UUID
	EventType() EventType
	Data() EventData
	Version() int
	Timestamp() time.Time
	Metadata() Metadata
}

// EventData는 이벤트의 실제 데이터를 나타냅니다
type EventData interface{}

// EventType은 이벤트 타입을 나타냅니다
type EventType string

// Metadata는 이벤트의 메타데이터를 나타냅니다
type Metadata map[string]interface{}

// BaseEvent는 Event 인터페이스의 기본 구현입니다
type BaseEvent struct {
	aggregateID uuid.UUID
	eventType   EventType
	data        EventData
	version     int
	timestamp   time.Time
	metadata    Metadata
}

func NewBaseEvent(aggregateID uuid.UUID, eventType EventType, data EventData, version int) *BaseEvent {
	return &BaseEvent{
		aggregateID: aggregateID,
		eventType:   eventType,
		data:        data,
		version:     version,
		timestamp:   time.Now(),
		metadata:    make(Metadata),
	}
}

func (e *BaseEvent) string() uuid.UUID    { return e.aggregateID }
func (e *BaseEvent) EventType() EventType { return e.eventType }
func (e *BaseEvent) Data() EventData      { return e.data }
func (e *BaseEvent) Version() int         { return e.version }
func (e *BaseEvent) Timestamp() time.Time { return e.timestamp }
func (e *BaseEvent) Metadata() Metadata   { return e.metadata }

// StoreMetrics는 저장소 성능 메트릭을 나타냅니다
type StoreMetrics struct {
	SaveOperations  int64           `json:"saveOperations"`
	LoadOperations  int64           `json:"loadOperations"`
	AverageSaveTime time.Duration   `json:"averageSaveTime"`
	AverageLoadTime time.Duration   `json:"averageLoadTime"`
	ErrorCount      int64           `json:"errorCount"`
	LastOperation   time.Time       `json:"lastOperation"`
	StorageStrategy StorageStrategy `json:"storageStrategy"`
}

// EventSerializer는 이벤트 직렬화를 담당합니다
type EventSerializer interface {
	Serialize(event Event) ([]byte, error)
	Deserialize(data []byte, eventType EventType) (Event, error)
}

// 에러 정의
var (
	ErrConcurrencyConflict = NewEventStoreError("concurrency conflict detected")
	ErrStreamNotFound      = NewEventStoreError("event stream not found")
	ErrInvalidVersion      = NewEventStoreError("invalid version specified")
	ErrSnapshotNotFound    = NewEventStoreError("snapshot not found")
	ErrSnapshotInvalid     = NewEventStoreError("snapshot data is invalid")
	ErrCompressionFailed   = NewEventStoreError("compression operation failed")
	ErrEncryptionFailed    = NewEventStoreError("encryption operation failed")
)

// EventStoreError는 이벤트 저장소 전용 에러입니다
type EventStoreError struct {
	message string
}

func NewEventStoreError(message string) *EventStoreError {
	return &EventStoreError{message: message}
}

func (e *EventStoreError) Error() string {
	return e.message
}

// EventQuery는 복잡한 이벤트 쿼리를 위한 구조체입니다
type EventQuery struct {
	EventTypes   []EventType `json:"eventTypes,omitempty"`
	StartTime    *time.Time  `json:"startTime,omitempty"`
	EndTime      *time.Time  `json:"endTime,omitempty"`
	AggregateIDs []uuid.UUID `json:"aggregateIds,omitempty"`
	Limit        int         `json:"limit,omitempty"`
	Offset       int         `json:"offset,omitempty"`
}

// QueryableEventStore는 복잡한 쿼리를 지원하는 이벤트 저장소입니다
type QueryableEventStore interface {
	EventStore
	FindEvents(ctx context.Context, query EventQuery) ([]Event, error)
	CountEvents(ctx context.Context, query EventQuery) (int64, error)
}

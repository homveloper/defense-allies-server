// aggregate_state.go - 집합체 상태 데이터 구조
package cqrsx

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AggregateState는 집합체의 특정 시점 상태를 나타냅니다
type AggregateState struct {
	AggregateID   uuid.UUID      `json:"aggregateId" bson:"aggregateId"`
	AggregateType string         `json:"aggregateType" bson:"aggregateType"`
	Version       int            `json:"version" bson:"version"`
	Data          []byte         `json:"data" bson:"data"`
	Metadata      map[string]any `json:"metadata" bson:"metadata"`
	Timestamp     time.Time      `json:"timestamp" bson:"timestamp"`
}

// NewAggregateState는 새로운 집합체 상태를 생성합니다
func NewAggregateState(aggregateID uuid.UUID, aggregateType string, version int, data []byte) *AggregateState {
	return &AggregateState{
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		Version:       version,
		Data:          data,
		Metadata:      make(map[string]any),
		Timestamp:     time.Now(),
	}
}

// Size는 데이터 크기를 반환합니다
func (as *AggregateState) Size() int64 {
	return int64(len(as.Data))
}

// IsEmpty는 데이터가 비어있는지 확인합니다
func (as *AggregateState) IsEmpty() bool {
	return len(as.Data) == 0
}

// Clone은 깊은 복사를 수행합니다
func (as *AggregateState) Clone() *AggregateState {
	cloned := &AggregateState{
		AggregateID:   as.AggregateID,
		AggregateType: as.AggregateType,
		Version:       as.Version,
		Data:          make([]byte, len(as.Data)),
		Metadata:      make(map[string]any),
		Timestamp:     as.Timestamp,
	}

	copy(cloned.Data, as.Data)

	// 메타데이터 깊은 복사
	for k, v := range as.Metadata {
		cloned.Metadata[k] = v
	}

	return cloned
}

// SetMetadata는 메타데이터를 설정합니다
func (as *AggregateState) SetMetadata(key string, value any) {
	if as.Metadata == nil {
		as.Metadata = make(map[string]any)
	}
	as.Metadata[key] = value
}

// GetMetadata는 메타데이터를 조회합니다
func (as *AggregateState) GetMetadata(key string) (any, bool) {
	if as.Metadata == nil {
		return nil, false
	}
	value, exists := as.Metadata[key]
	return value, exists
}

// ToJSON은 JSON 형태로 직렬화합니다 (디버깅용)
func (as *AggregateState) ToJSON() ([]byte, error) {
	return json.Marshal(as)
}

// FromJSON은 JSON에서 역직렬화합니다 (디버깅용)
func FromJSON(data []byte) (*AggregateState, error) {
	var state AggregateState
	err := json.Unmarshal(data, &state)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// StateQuery는 상태 조회를 위한 쿼리 구조체입니다
type StateQuery struct {
	AggregateIDs  []uuid.UUID `json:"aggregateIds,omitempty"`
	AggregateType string      `json:"aggregateType,omitempty"`
	MinVersion    *int        `json:"minVersion,omitempty"`
	MaxVersion    *int        `json:"maxVersion,omitempty"`
	StartTime     *time.Time  `json:"startTime,omitempty"`
	EndTime       *time.Time  `json:"endTime,omitempty"`
	Limit         int         `json:"limit,omitempty"`
	Offset        int         `json:"offset,omitempty"`
}

// StateMetrics는 상태 저장소의 메트릭을 나타냅니다
type StateMetrics struct {
	TotalStates       int64     `json:"totalStates"`
	TotalStorageBytes int64     `json:"totalStorageBytes"`
	AverageSize       int64     `json:"averageSize"`
	MaxSize           int64     `json:"maxSize"`
	MinSize           int64     `json:"minSize"`
	AggregateTypes    []string  `json:"aggregateTypes"`
	OldestState       time.Time `json:"oldestState"`
	NewestState       time.Time `json:"newestState"`
}

// StateValidationError는 상태 검증 에러입니다
type StateValidationError struct {
	Field   string
	Message string
}

func (e *StateValidationError) Error() string {
	return "state validation error: " + e.Field + " - " + e.Message
}

// Validate는 상태 데이터를 검증합니다
func (as *AggregateState) Validate() error {
	if as.AggregateID == uuid.Nil {
		return &StateValidationError{
			Field:   "AggregateID",
			Message: "aggregate ID cannot be nil",
		}
	}

	if as.AggregateType == "" {
		return &StateValidationError{
			Field:   "AggregateType",
			Message: "aggregate type cannot be empty",
		}
	}

	if as.Version < 0 {
		return &StateValidationError{
			Field:   "Version",
			Message: "version cannot be negative",
		}
	}

	if as.Timestamp.IsZero() {
		return &StateValidationError{
			Field:   "Timestamp",
			Message: "timestamp cannot be zero",
		}
	}

	return nil
}

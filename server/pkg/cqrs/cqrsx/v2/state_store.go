// state_store.go - 집합체 상태 저장소 인터페이스
package cqrsx

import (
	"context"

	"github.com/google/uuid"
)

// StateStore는 집합체 상태 저장소의 핵심 인터페이스입니다
type StateStore interface {
	// Save는 집합체 상태를 저장합니다
	Save(ctx context.Context, state *AggregateState) error

	// Load는 집합체의 최신 상태를 로드합니다
	Load(ctx context.Context, aggregateID uuid.UUID) (*AggregateState, error)

	// LoadVersion은 특정 버전의 상태를 로드합니다
	LoadVersion(ctx context.Context, aggregateID uuid.UUID, version int) (*AggregateState, error)

	// Delete는 특정 버전의 상태를 삭제합니다
	Delete(ctx context.Context, aggregateID uuid.UUID, version int) error

	// DeleteAll은 집합체의 모든 상태를 삭제합니다
	DeleteAll(ctx context.Context, aggregateID uuid.UUID) error

	// List는 집합체의 모든 상태 버전을 조회합니다
	List(ctx context.Context, aggregateID uuid.UUID) ([]*AggregateState, error)

	// Count는 저장된 상태 개수를 반환합니다
	Count(ctx context.Context, aggregateID uuid.UUID) (int64, error)

	// Exists는 특정 상태가 존재하는지 확인합니다
	Exists(ctx context.Context, aggregateID uuid.UUID, version int) (bool, error)

	// Close는 저장소 연결을 정리합니다
	Close() error
}

// QueryableStateStore는 복잡한 쿼리를 지원하는 상태 저장소입니다
type QueryableStateStore interface {
	StateStore

	// Query는 조건에 맞는 상태들을 조회합니다
	Query(ctx context.Context, query StateQuery) ([]*AggregateState, error)

	// CountByQuery는 쿼리 조건에 맞는 상태 개수를 반환합니다
	CountByQuery(ctx context.Context, query StateQuery) (int64, error)

	// GetAggregateTypes는 저장된 모든 집합체 타입을 반환합니다
	GetAggregateTypes(ctx context.Context) ([]string, error)

	// GetVersions는 집합체의 모든 버전을 반환합니다
	GetVersions(ctx context.Context, aggregateID uuid.UUID) ([]int, error)
}

// MetricsStateStore는 메트릭을 제공하는 상태 저장소입니다
type MetricsStateStore interface {
	StateStore

	// GetMetrics는 저장소 메트릭을 반환합니다
	GetMetrics(ctx context.Context) (*StateMetrics, error)

	// GetAggregateMetrics는 특정 집합체의 메트릭을 반환합니다
	GetAggregateMetrics(ctx context.Context, aggregateID uuid.UUID) (*StateMetrics, error)
}

// StateStoreOption은 상태 저장소 설정을 위한 옵션입니다
type StateStoreOption func(*StateStoreConfig)

// StateStoreConfig는 상태 저장소 설정입니다
type StateStoreConfig struct {
	// 압축 설정
	CompressionEnabled bool
	CompressionType    CompressionType
	CompressionMinSize int // 최소 압축 크기 (바이트)

	// 암호화 설정
	EncryptionEnabled bool
	Encryptor         Encryptor

	// 보존 정책
	RetentionPolicy RetentionPolicy

	// 성능 설정
	BatchSize       int  // 배치 처리 크기
	IndexingEnabled bool // 인덱스 생성 여부
	MetricsEnabled  bool // 메트릭 수집 여부
}

// RetentionPolicy는 상태 보존 정책입니다
type RetentionPolicy interface {
	// ShouldKeep은 특정 상태를 보존할지 결정합니다
	ShouldKeep(state *AggregateState) bool

	// GetCleanupCandidates는 정리 대상 상태들을 반환합니다
	GetCleanupCandidates(ctx context.Context, states []*AggregateState) []*AggregateState
}

// 기본 옵션 함수들

// WithCompression은 압축을 활성화합니다
func WithCompression(compressionType CompressionType) StateStoreOption {
	return func(config *StateStoreConfig) {
		config.CompressionEnabled = true
		config.CompressionType = compressionType
		config.CompressionMinSize = 1024 // 1KB 이상만 압축
	}
}

// WithEncryption은 암호화를 활성화합니다
func WithEncryption(encryptor Encryptor) StateStoreOption {
	return func(config *StateStoreConfig) {
		config.EncryptionEnabled = true
		config.Encryptor = encryptor
	}
}

// WithRetentionPolicy는 보존 정책을 설정합니다
func WithRetentionPolicy(policy RetentionPolicy) StateStoreOption {
	return func(config *StateStoreConfig) {
		config.RetentionPolicy = policy
	}
}

// WithBatchSize는 배치 크기를 설정합니다
func WithBatchSize(size int) StateStoreOption {
	return func(config *StateStoreConfig) {
		config.BatchSize = size
	}
}

// WithMetrics는 메트릭 수집을 활성화합니다
func WithMetrics() StateStoreOption {
	return func(config *StateStoreConfig) {
		config.MetricsEnabled = true
	}
}

// WithIndexing은 인덱싱을 활성화합니다
func WithIndexing() StateStoreOption {
	return func(config *StateStoreConfig) {
		config.IndexingEnabled = true
	}
}

// 기본 설정 생성
func NewDefaultStateStoreConfig() *StateStoreConfig {
	return &StateStoreConfig{
		CompressionEnabled: false,
		CompressionType:    CompressionNone,
		CompressionMinSize: 1024,
		EncryptionEnabled:  false,
		BatchSize:          100,
		IndexingEnabled:    true,
		MetricsEnabled:     false,
	}
}

// 에러 정의
var (
	ErrStateNotFound    = NewEventStoreError("aggregate state not found")
	ErrStateExists      = NewEventStoreError("aggregate state already exists")
	ErrInvalidState     = NewEventStoreError("invalid aggregate state")
	ErrVersionConflict  = NewEventStoreError("version conflict")
	ErrStoreUnavailable = NewEventStoreError("state store unavailable")
)

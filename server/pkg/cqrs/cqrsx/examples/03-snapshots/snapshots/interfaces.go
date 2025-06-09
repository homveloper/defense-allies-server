package snapshots

import (
	"context"
	"fmt"
	"time"
)

// Snapshot 스냅샷 인터페이스
type Snapshot interface {
	ID() string
	Type() string
	Version() int
	Data() []byte
	Timestamp() time.Time
	Metadata() map[string]interface{}
}

// SnapshotStore 스냅샷 저장소 인터페이스
type SnapshotStore interface {
	// SaveSnapshot 스냅샷 저장
	SaveSnapshot(ctx context.Context, snapshot Snapshot) error

	// GetSnapshot 최신 스냅샷 조회 (maxVersion 이하의 가장 최신)
	GetSnapshot(ctx context.Context, aggregateID string, maxVersion int) (Snapshot, error)

	// GetSnapshotByVersion 특정 버전의 스냅샷 조회
	GetSnapshotByVersion(ctx context.Context, aggregateID string, version int) (Snapshot, error)

	// DeleteSnapshot 특정 버전의 스냅샷 삭제
	DeleteSnapshot(ctx context.Context, aggregateID string, version int) error

	// DeleteOldSnapshots 오래된 스냅샷들 삭제 (keepCount개만 유지)
	DeleteOldSnapshots(ctx context.Context, aggregateID string, keepCount int) error

	// ListSnapshots 스냅샷 목록 조회
	ListSnapshots(ctx context.Context, aggregateID string) ([]Snapshot, error)

	// GetSnapshotStats 스냅샷 통계 조회
	GetSnapshotStats(ctx context.Context) (map[string]interface{}, error)
}

// Aggregate 인터페이스 (임시 정의)
type Aggregate interface {
	ID() string
	Type() string
	Version() int
	OriginalVersion() int
}

// SnapshotPolicy 스냅샷 정책 인터페이스
type SnapshotPolicy interface {
	// ShouldCreateSnapshot 스냅샷을 생성해야 하는지 판단
	ShouldCreateSnapshot(aggregate Aggregate, eventCount int) bool

	// GetSnapshotInterval 스냅샷 생성 간격 반환
	GetSnapshotInterval() int

	// GetPolicyName 정책 이름 반환
	GetPolicyName() string
}

// SnapshotSerializer 스냅샷 직렬화 인터페이스
type SnapshotSerializer interface {
	// Serialize Aggregate를 바이트 배열로 직렬화
	Serialize(aggregate Aggregate) ([]byte, error)

	// Deserialize 바이트 배열을 Aggregate로 역직렬화
	Deserialize(data []byte, aggregateType string) (Aggregate, error)

	// GetContentType 직렬화 형식 반환 (예: "application/json", "application/bson")
	GetContentType() string

	// GetCompressionType 압축 형식 반환 (예: "none", "gzip", "lz4")
	GetCompressionType() string
}

// SnapshotManager 스냅샷 관리자 인터페이스
type SnapshotManager interface {
	// CreateSnapshot 스냅샷 생성
	CreateSnapshot(ctx context.Context, aggregate Aggregate) error

	// RestoreFromSnapshot 스냅샷에서 Aggregate 복원
	RestoreFromSnapshot(ctx context.Context, aggregateID string, maxVersion int) (Aggregate, int, error)

	// ShouldCreateSnapshot 스냅샷 생성 여부 판단
	ShouldCreateSnapshot(aggregate Aggregate, eventCount int) bool

	// CleanupOldSnapshots 오래된 스냅샷 정리
	CleanupOldSnapshots(ctx context.Context, aggregateID string) error

	// GetSnapshotInfo 스냅샷 정보 조회
	GetSnapshotInfo(ctx context.Context, aggregateID string) ([]SnapshotInfo, error)
}

// SnapshotInfo 스냅샷 정보
type SnapshotInfo struct {
	string        string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int                    `json:"version"`
	Size          int64                  `json:"size"`
	ContentType   string                 `json:"content_type"`
	Compression   string                 `json:"compression"`
	Timestamp     time.Time              `json:"timestamp"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// SnapshotMetrics 스냅샷 메트릭
type SnapshotMetrics struct {
	TotalSnapshots    int64              `json:"total_snapshots"`
	TotalSize         int64              `json:"total_size"`
	AverageSize       float64            `json:"average_size"`
	CompressionRatio  float64            `json:"compression_ratio"`
	SnapshotsByType   map[string]int64   `json:"snapshots_by_type"`
	SnapshotsByPolicy map[string]int64   `json:"snapshots_by_policy"`
	OldestSnapshot    *time.Time         `json:"oldest_snapshot,omitempty"`
	NewestSnapshot    *time.Time         `json:"newest_snapshot,omitempty"`
	CreationRate      map[string]float64 `json:"creation_rate"` // per hour/day/week
}

// SnapshotConfiguration 스냅샷 설정
type SnapshotConfiguration struct {
	// 기본 설정
	Enabled            bool   `json:"enabled"`
	DefaultPolicy      string `json:"default_policy"`
	DefaultSerializer  string `json:"default_serializer"`
	DefaultCompression string `json:"default_compression"`

	// 정책 설정
	EventCountThreshold int   `json:"event_count_threshold"`
	TimeIntervalMinutes int   `json:"time_interval_minutes"`
	SizeThresholdBytes  int64 `json:"size_threshold_bytes"`

	// 정리 설정
	MaxSnapshotsPerAggregate int `json:"max_snapshots_per_aggregate"`
	CleanupIntervalHours     int `json:"cleanup_interval_hours"`
	RetentionDays            int `json:"retention_days"`

	// 성능 설정
	AsyncCreation    bool `json:"async_creation"`
	BatchSize        int  `json:"batch_size"`
	CompressionLevel int  `json:"compression_level"`

	// 모니터링 설정
	EnableMetrics   bool               `json:"enable_metrics"`
	MetricsInterval int                `json:"metrics_interval"`
	AlertThresholds map[string]float64 `json:"alert_thresholds"`
}

// SnapshotEvent 스냅샷 관련 이벤트
type SnapshotEvent struct {
	Type          string                 `json:"type"`
	string        string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int                    `json:"version"`
	Timestamp     time.Time              `json:"timestamp"`
	Duration      time.Duration          `json:"duration"`
	Size          int64                  `json:"size"`
	Success       bool                   `json:"success"`
	Error         string                 `json:"error,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// SnapshotEventType 스냅샷 이벤트 타입
const (
	SnapshotEventCreated  = "snapshot_created"
	SnapshotEventRestored = "snapshot_restored"
	SnapshotEventDeleted  = "snapshot_deleted"
	SnapshotEventCleaned  = "snapshot_cleaned"
	SnapshotEventFailed   = "snapshot_failed"
)

// SnapshotError 스냅샷 관련 에러
type SnapshotError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Operation string `json:"operation"`
	Cause     error  `json:"cause,omitempty"`
}

func (e *SnapshotError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s in %s: %v", e.Code, e.Message, e.Operation, e.Cause)
	}
	return fmt.Sprintf("[%s] %s in %s", e.Code, e.Message, e.Operation)
}

// 스냅샷 에러 코드
const (
	ErrCodeSnapshotNotFound      = "SNAPSHOT_NOT_FOUND"
	ErrCodeSerializationFailed   = "SERIALIZATION_FAILED"
	ErrCodeDeserializationFailed = "DESERIALIZATION_FAILED"
	ErrCodeCompressionFailed     = "COMPRESSION_FAILED"
	ErrCodeStorageFailed         = "STORAGE_FAILED"
	ErrCodePolicyViolation       = "POLICY_VIOLATION"
	ErrCodeInvalidConfiguration  = "INVALID_CONFIGURATION"
)

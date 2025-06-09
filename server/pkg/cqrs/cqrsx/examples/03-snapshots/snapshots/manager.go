package snapshots

import (
	"context"
	"fmt"
	"log"
	"time"

	"cqrs/cqrsx/examples/03-snapshots/domain"
)

// DefaultSnapshotManager 기본 스냅샷 매니저 구현
type DefaultSnapshotManager struct {
	store      SnapshotStore
	serializer SnapshotSerializer
	policy     SnapshotPolicy
	config     *SnapshotConfiguration
}

// NewDefaultSnapshotManager 기본 스냅샷 매니저 생성
func NewDefaultSnapshotManager(
	store SnapshotStore,
	serializer SnapshotSerializer,
	policy SnapshotPolicy,
	config *SnapshotConfiguration,
) *DefaultSnapshotManager {
	if config == nil {
		config = &SnapshotConfiguration{
			Enabled:                  true,
			MaxSnapshotsPerAggregate: 5,
			AsyncCreation:            false,
		}
	}

	return &DefaultSnapshotManager{
		store:      store,
		serializer: serializer,
		policy:     policy,
		config:     config,
	}
}

// CreateSnapshot 스냅샷 생성
func (m *DefaultSnapshotManager) CreateSnapshot(ctx context.Context, aggregate Aggregate) error {
	if !m.config.Enabled {
		return nil
	}

	start := time.Now()

	// 직렬화
	data, err := m.serializer.Serialize(aggregate)
	if err != nil {
		m.logSnapshotEvent(SnapshotEventFailed, aggregate, 0, time.Since(start), 0, err)
		return &SnapshotError{
			Code:      ErrCodeSerializationFailed,
			Message:   "failed to serialize aggregate",
			Operation: "CreateSnapshot",
			Cause:     err,
		}
	}

	// 메타데이터 생성
	metadata := map[string]interface{}{
		"content_type":  m.serializer.GetContentType(),
		"compression":   m.serializer.GetCompressionType(),
		"policy":        m.policy.GetPolicyName(),
		"created_by":    "DefaultSnapshotManager",
		"creation_time": start,
	}

	// 스냅샷 객체 생성
	snapshot := &MongoSnapshot{
		aggregateID:   aggregate.ID(),
		aggregateType: aggregate.Type(),
		version:       aggregate.Version(),
		data:          data,
		contentType:   m.serializer.GetContentType(),
		compression:   m.serializer.GetCompressionType(),
		timestamp:     time.Now(),
		metadata:      metadata,
	}

	// 저장
	if err := m.store.SaveSnapshot(ctx, snapshot); err != nil {
		m.logSnapshotEvent(SnapshotEventFailed, aggregate, int64(len(data)), time.Since(start), 0, err)
		return &SnapshotError{
			Code:      ErrCodeStorageFailed,
			Message:   "failed to save snapshot",
			Operation: "CreateSnapshot",
			Cause:     err,
		}
	}

	// 성공 로그
	m.logSnapshotEvent(SnapshotEventCreated, aggregate, int64(len(data)), time.Since(start), 0, nil)

	// 오래된 스냅샷 정리 (비동기)
	if m.config.MaxSnapshotsPerAggregate > 0 {
		go func() {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := m.store.DeleteOldSnapshots(cleanupCtx, aggregate.ID(), m.config.MaxSnapshotsPerAggregate); err != nil {
				log.Printf("Failed to cleanup old snapshots for %s: %v", aggregate.ID(), err)
			}
		}()
	}

	return nil
}

// RestoreFromSnapshot 스냅샷에서 Aggregate 복원
func (m *DefaultSnapshotManager) RestoreFromSnapshot(ctx context.Context, aggregateID string, maxVersion int) (Aggregate, int, error) {
	start := time.Now()

	// 스냅샷 조회
	snapshot, err := m.store.GetSnapshot(ctx, aggregateID, maxVersion)
	if err != nil {
		if snapshotErr, ok := err.(*SnapshotError); ok && snapshotErr.Code == ErrCodeSnapshotNotFound {
			return nil, 0, err // 스냅샷이 없으면 그대로 반환
		}
		return nil, 0, &SnapshotError{
			Code:      ErrCodeStorageFailed,
			Message:   "failed to get snapshot",
			Operation: "RestoreFromSnapshot",
			Cause:     err,
		}
	}

	// 역직렬화
	aggregate, err := m.serializer.Deserialize(snapshot.Data(), snapshot.Type())
	if err != nil {
		m.logSnapshotEvent(SnapshotEventFailed, nil, int64(len(snapshot.Data())), time.Since(start), 0, err)
		return nil, 0, &SnapshotError{
			Code:      ErrCodeDeserializationFailed,
			Message:   "failed to deserialize snapshot",
			Operation: "RestoreFromSnapshot",
			Cause:     err,
		}
	}

	// 성공 로그
	m.logSnapshotEvent(SnapshotEventRestored, aggregate, int64(len(snapshot.Data())), time.Since(start), 0, nil)

	return aggregate, snapshot.Version(), nil
}

// ShouldCreateSnapshot 스냅샷 생성 여부 판단
func (m *DefaultSnapshotManager) ShouldCreateSnapshot(aggregate Aggregate, eventCount int) bool {
	if !m.config.Enabled {
		return false
	}

	return m.policy.ShouldCreateSnapshot(aggregate, eventCount)
}

// CleanupOldSnapshots 오래된 스냅샷 정리
func (m *DefaultSnapshotManager) CleanupOldSnapshots(ctx context.Context, aggregateID string) error {
	if m.config.MaxSnapshotsPerAggregate <= 0 {
		return nil
	}

	start := time.Now()

	err := m.store.DeleteOldSnapshots(ctx, aggregateID, m.config.MaxSnapshotsPerAggregate)
	if err != nil {
		return &SnapshotError{
			Code:      ErrCodeStorageFailed,
			Message:   "failed to cleanup old snapshots",
			Operation: "CleanupOldSnapshots",
			Cause:     err,
		}
	}

	// 정리 완료 로그
	log.Printf("Cleaned up old snapshots for %s in %v", aggregateID, time.Since(start))

	return nil
}

// GetSnapshotInfo 스냅샷 정보 조회
func (m *DefaultSnapshotManager) GetSnapshotInfo(ctx context.Context, aggregateID string) ([]SnapshotInfo, error) {
	snapshots, err := m.store.ListSnapshots(ctx, aggregateID)
	if err != nil {
		return nil, &SnapshotError{
			Code:      ErrCodeStorageFailed,
			Message:   "failed to list snapshots",
			Operation: "GetSnapshotInfo",
			Cause:     err,
		}
	}

	var infos []SnapshotInfo
	for _, snapshot := range snapshots {
		metadata := snapshot.Metadata()

		info := SnapshotInfo{
			string:        snapshot.ID(),
			AggregateType: snapshot.Type(),
			Version:       snapshot.Version(),
			Size:          int64(len(snapshot.Data())),
			Timestamp:     snapshot.Timestamp(),
			Metadata:      metadata,
		}

		// 메타데이터에서 ContentType과 Compression 추출
		if ct, ok := metadata["content_type"].(string); ok {
			info.ContentType = ct
		}
		if comp, ok := metadata["compression"].(string); ok {
			info.Compression = comp
		}

		infos = append(infos, info)
	}

	return infos, nil
}

// logSnapshotEvent 스냅샷 이벤트 로깅
func (m *DefaultSnapshotManager) logSnapshotEvent(eventType string, aggregate Aggregate, size int64, duration time.Duration, version int, err error) {
	event := SnapshotEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Duration:  duration,
		Size:      size,
		Success:   err == nil,
		Metadata:  make(map[string]interface{}),
	}

	if aggregate != nil {
		event.string = aggregate.ID()
		event.AggregateType = aggregate.Type()
		event.Version = aggregate.Version()
	} else {
		event.Version = version
	}

	if err != nil {
		event.Error = err.Error()
	}

	// 메타데이터 추가
	event.Metadata["policy"] = m.policy.GetPolicyName()
	event.Metadata["serializer"] = m.serializer.GetContentType()
	event.Metadata["compression"] = m.serializer.GetCompressionType()

	// 로그 출력
	if err != nil {
		log.Printf("Snapshot %s failed for %s v%d: %v (took %v)",
			eventType, event.string, event.Version, err, duration)
	} else {
		log.Printf("Snapshot %s for %s v%d: %d bytes (took %v)",
			eventType, event.string, event.Version, size, duration)
	}
}

// CreateSnapshotFromData 데이터에서 직접 스냅샷 생성 (테스트용)
func CreateSnapshotFromData(aggregateID, aggregateType string, version int, data []byte, contentType, compression string) Snapshot {
	metadata := map[string]interface{}{
		"content_type": contentType,
		"compression":  compression,
		"size":         int64(len(data)),
	}

	return &MongoSnapshot{
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
		version:       version,
		data:          data,
		contentType:   contentType,
		compression:   compression,
		timestamp:     time.Now(),
		metadata:      metadata,
	}
}

// RestoreOrderFromSnapshot Order 전용 복원 헬퍼 함수
func RestoreOrderFromSnapshot(ctx context.Context, manager SnapshotManager, orderID string, maxVersion int) (*domain.Order, int, error) {
	aggregate, version, err := manager.RestoreFromSnapshot(ctx, orderID, maxVersion)
	if err != nil {
		return nil, 0, err
	}

	if aggregate == nil {
		return nil, 0, nil
	}

	order, ok := aggregate.(*domain.Order)
	if !ok {
		return nil, 0, fmt.Errorf("restored aggregate is not an Order: %T", aggregate)
	}

	return order, version, nil
}

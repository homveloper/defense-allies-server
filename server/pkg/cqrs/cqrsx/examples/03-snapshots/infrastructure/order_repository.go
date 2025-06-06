package infrastructure

import (
	"context"
	"fmt"
	"log"
	"time"

	"defense-allies-server/pkg/cqrs"
	"defense-allies-server/pkg/cqrs/cqrsx"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/03-snapshots/domain"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/03-snapshots/snapshots"
)

// OrderRepository 스냅샷 지원 주문 리포지토리
type OrderRepository struct {
	eventStore      cqrsx.EventStore
	snapshotManager snapshots.SnapshotManager
	enableSnapshots bool
}

// NewOrderRepository 주문 리포지토리 생성
func NewOrderRepository(eventStore cqrsx.EventStore, snapshotManager snapshots.SnapshotManager) *OrderRepository {
	return &OrderRepository{
		eventStore:      eventStore,
		snapshotManager: snapshotManager,
		enableSnapshots: snapshotManager != nil,
	}
}

// Save 주문 저장 (이벤트 + 스냅샷)
func (r *OrderRepository) Save(ctx context.Context, order *domain.Order) error {
	// 커밋되지 않은 변경사항 가져오기
	uncommittedChanges := order.GetUncommittedChanges()
	if len(uncommittedChanges) == 0 {
		return nil // 변경사항이 없으면 저장하지 않음
	}

	log.Printf("Saving %d events for order %s", len(uncommittedChanges), order.ID())

	// Event Store에 이벤트들 저장
	expectedVersion := order.OriginalVersion()

	log.Printf("Attempting to save %d events for order %s with expectedVersion %d (original: %d, current: %d)",
		len(uncommittedChanges), order.ID(), expectedVersion, order.OriginalVersion(), order.Version())

	err := r.eventStore.SaveEvents(ctx, order.ID(), uncommittedChanges, expectedVersion)
	if err != nil {
		return fmt.Errorf("failed to save events for order %s: %w", order.ID(), err)
	}

	// 변경사항 클리어
	order.ClearChanges()

	// 저장 성공 후 원본 버전을 현재 버전으로 업데이트 (다음 저장을 위해)
	order.SetOriginalVersion(order.Version())

	log.Printf("Successfully saved order %s with version %d", order.ID(), order.Version())

	// 스냅샷 생성 여부 확인 및 생성
	if r.enableSnapshots {
		eventCount := order.Version()
		if r.snapshotManager.ShouldCreateSnapshot(order, eventCount) {
			log.Printf("Creating snapshot for order %s at version %d", order.ID(), order.Version())

			if err := r.snapshotManager.CreateSnapshot(ctx, order); err != nil {
				// 스냅샷 생성 실패는 치명적이지 않음 (로그만 남김)
				log.Printf("Failed to create snapshot for order %s: %v", order.ID(), err)
			}
		}
	}

	return nil
}

// GetByID ID로 주문 조회 (스냅샷 + 이벤트)
func (r *OrderRepository) GetByID(ctx context.Context, orderID string) (*domain.Order, error) {
	start := time.Now()

	var order *domain.Order
	var fromVersion int = 1
	var restoredFromSnapshot bool

	// 스냅샷에서 복원 시도
	if r.enableSnapshots {
		snapshotOrder, snapshotVersion, err := snapshots.RestoreOrderFromSnapshot(ctx, r.snapshotManager, orderID, int(^uint(0)>>1)) // max int
		if err != nil {
			if !isSnapshotNotFoundError(err) {
				log.Printf("Failed to restore order %s from snapshot: %v", orderID, err)
			}
		} else if snapshotOrder != nil {
			order = snapshotOrder
			fromVersion = snapshotVersion + 1
			restoredFromSnapshot = true
			log.Printf("Restored order %s from snapshot at version %d", orderID, snapshotVersion)
		}
	}

	// 스냅샷이 없으면 새 주문 생성
	if order == nil {
		order = domain.NewOrderWithID(orderID)
		fromVersion = 1
	}

	// 스냅샷 이후 (또는 처음부터) 이벤트들 로드
	events, err := r.eventStore.LoadEvents(ctx, orderID, "Order", fromVersion, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to load events for order %s: %w", orderID, err)
	}

	if len(events) == 0 && !restoredFromSnapshot {
		return nil, fmt.Errorf("order %s not found", orderID)
	}

	log.Printf("Loading order %s from event store (from version %d)", orderID, fromVersion)
	log.Printf("Found %d events for order %s", len(events), orderID)

	// 이벤트들을 순서대로 적용하여 상태 복원
	for _, event := range events {
		err := order.ApplyEvent(event)
		if err != nil {
			return nil, fmt.Errorf("failed to apply event %s for order %s: %w",
				event.EventType(), orderID, err)
		}
	}

	// 로드 완료 후 원본 버전을 현재 버전으로 설정 (동시성 제어용)
	order.SetOriginalVersion(order.Version())

	loadTime := time.Since(start)
	log.Printf("Successfully loaded order %s with version %d in %v (snapshot: %v, events: %d)",
		orderID, order.Version(), loadTime, restoredFromSnapshot, len(events))

	return order, nil
}

// GetEventHistory 주문의 이벤트 히스토리 조회
func (r *OrderRepository) GetEventHistory(ctx context.Context, orderID string) ([]cqrs.EventMessage, error) {
	return r.eventStore.GetEventHistory(ctx, orderID, "Order", 1)
}

// GetSnapshotInfo 주문의 스냅샷 정보 조회
func (r *OrderRepository) GetSnapshotInfo(ctx context.Context, orderID string) ([]snapshots.SnapshotInfo, error) {
	if !r.enableSnapshots {
		return nil, fmt.Errorf("snapshots are not enabled")
	}

	return r.snapshotManager.GetSnapshotInfo(ctx, orderID)
}

// CreateSnapshot 수동으로 스냅샷 생성
func (r *OrderRepository) CreateSnapshot(ctx context.Context, orderID string) error {
	if !r.enableSnapshots {
		return fmt.Errorf("snapshots are not enabled")
	}

	// 주문 로드
	order, err := r.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to load order for snapshot: %w", err)
	}

	// 스냅샷 생성
	return r.snapshotManager.CreateSnapshot(ctx, order)
}

// CleanupSnapshots 오래된 스냅샷 정리
func (r *OrderRepository) CleanupSnapshots(ctx context.Context, orderID string) error {
	if !r.enableSnapshots {
		return fmt.Errorf("snapshots are not enabled")
	}

	return r.snapshotManager.CleanupOldSnapshots(ctx, orderID)
}

// GetPerformanceMetrics 성능 메트릭 조회
func (r *OrderRepository) GetPerformanceMetrics(ctx context.Context, orderID string) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// 이벤트 개수 조회
	events, err := r.GetEventHistory(ctx, orderID)
	if err != nil {
		return nil, err
	}
	metrics["event_count"] = len(events)

	// 스냅샷 정보 조회
	if r.enableSnapshots {
		snapshotInfos, err := r.GetSnapshotInfo(ctx, orderID)
		if err == nil {
			metrics["snapshot_count"] = len(snapshotInfos)
			if len(snapshotInfos) > 0 {
				latest := snapshotInfos[len(snapshotInfos)-1]
				metrics["latest_snapshot_version"] = latest.Version
				metrics["latest_snapshot_size"] = latest.Size
				metrics["latest_snapshot_time"] = latest.Timestamp
			}
		}
	}

	return metrics, nil
}

// 헬퍼 함수들

// isSnapshotNotFoundError 스냅샷 없음 에러인지 확인
func isSnapshotNotFoundError(err error) bool {
	if snapshotErr, ok := err.(*snapshots.SnapshotError); ok {
		return snapshotErr.Code == snapshots.ErrCodeSnapshotNotFound
	}
	return false
}

// EnableSnapshots 스냅샷 기능 활성화/비활성화
func (r *OrderRepository) EnableSnapshots(enable bool) {
	r.enableSnapshots = enable && r.snapshotManager != nil
}

// IsSnapshotsEnabled 스냅샷 기능 활성화 여부 확인
func (r *OrderRepository) IsSnapshotsEnabled() bool {
	return r.enableSnapshots
}

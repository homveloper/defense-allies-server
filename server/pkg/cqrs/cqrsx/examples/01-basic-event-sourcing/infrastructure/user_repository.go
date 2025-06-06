package infrastructure

import (
	"context"
	"defense-allies-server/pkg/cqrs"
	"defense-allies-server/pkg/cqrs/cqrsx"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/01-basic-event-sourcing/domain"
	"fmt"
	"log"
)

// UserRepository 사용자 저장소 - Event Store를 사용한 구현
type UserRepository struct {
	eventStore *cqrsx.MongoEventStore
}

// NewUserRepository 새로운 UserRepository 생성
func NewUserRepository(eventStore *cqrsx.MongoEventStore) *UserRepository {
	return &UserRepository{
		eventStore: eventStore,
	}
}

// Save 사용자 저장 - 변경된 이벤트들을 Event Store에 저장
func (r *UserRepository) Save(ctx context.Context, user *domain.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	// 변경된 이벤트들 가져오기
	uncommittedChanges := user.GetUncommittedChanges()
	if len(uncommittedChanges) == 0 {
		log.Printf("No uncommitted changes for user %s", user.ID())
		return nil
	}

	log.Printf("Saving %d events for user %s", len(uncommittedChanges), user.ID())

	// Event Store에 이벤트들 저장
	// 동시성 제어를 위해 사용자의 원본 버전을 사용 (로드 시점의 버전)
	expectedVersion := user.OriginalVersion()

	log.Printf("Attempting to save %d events for user %s with expectedVersion %d (original: %d, current: %d)",
		len(uncommittedChanges), user.ID(), expectedVersion, user.OriginalVersion(), user.CurrentVersion())

	err := r.eventStore.SaveEvents(ctx, user.ID(), uncommittedChanges, expectedVersion)
	if err != nil {
		return fmt.Errorf("failed to save events for user %s: %w", user.ID(), err)
	}

	// 변경사항 클리어
	user.ClearChanges()

	// 저장 성공 후 원본 버전을 현재 버전으로 업데이트 (다음 저장을 위해)
	user.SetOriginalVersion(user.CurrentVersion())

	log.Printf("Successfully saved user %s with version %d", user.ID(), user.Version())
	return nil
}

// GetByID ID로 사용자 조회 - Event Store에서 이벤트들을 조회하여 상태 복원
func (r *UserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	log.Printf("Loading user %s from event store", userID)

	// Event Store에서 이벤트 히스토리 조회
	events, err := r.eventStore.GetEventHistory(ctx, userID, "User", 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get event history for user %s: %w", userID, err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("user %s not found", userID)
	}

	log.Printf("Found %d events for user %s", len(events), userID)

	// 새로운 User 인스턴스 생성
	user := domain.NewUserWithID(userID)

	// 이벤트들을 순서대로 적용하여 상태 복원
	for _, event := range events {
		err := user.ReplayEvent(event)
		if err != nil {
			return nil, fmt.Errorf("failed to apply event %s for user %s: %w",
				event.EventType(), userID, err)
		}
	}

	// 로드 완료 후 원본 버전을 현재 버전으로 설정 (동시성 제어용)
	user.SetOriginalVersion(user.CurrentVersion())

	log.Printf("Successfully loaded user %s with version %d", userID, user.Version())
	return user, nil
}

// GetEventHistory 사용자의 이벤트 히스토리 조회
func (r *UserRepository) GetEventHistory(ctx context.Context, userID string) ([]cqrs.EventMessage, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	log.Printf("Getting event history for user %s", userID)

	events, err := r.eventStore.GetEventHistory(ctx, userID, "User", 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get event history for user %s: %w", userID, err)
	}

	log.Printf("Found %d events in history for user %s", len(events), userID)
	return events, nil
}

// GetLastEventVersion 사용자의 마지막 이벤트 버전 조회
func (r *UserRepository) GetLastEventVersion(ctx context.Context, userID string) (int, error) {
	if userID == "" {
		return 0, fmt.Errorf("user ID cannot be empty")
	}

	// 이벤트 히스토리를 조회해서 마지막 버전 확인
	events, err := r.eventStore.GetEventHistory(ctx, userID, "User", 1)
	if err != nil {
		return 0, fmt.Errorf("failed to get event history for user %s: %w", userID, err)
	}

	if len(events) == 0 {
		return 0, nil
	}

	// 마지막 이벤트의 버전 반환
	lastEvent := events[len(events)-1]
	return lastEvent.Version(), nil
}

// Exists 사용자 존재 여부 확인
func (r *UserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	if userID == "" {
		return false, fmt.Errorf("user ID cannot be empty")
	}

	version, err := r.GetLastEventVersion(ctx, userID)
	if err != nil {
		// 사용자가 존재하지 않는 경우
		if cqrsErr, ok := err.(*cqrs.CQRSError); ok {
			if cqrsErr.Code == cqrs.ErrCodeAggregateNotFound.String() {
				return false, nil
			}
		}
		return false, err
	}

	return version > 0, nil
}

// ListAllUsers 모든 사용자 목록 조회 (데모용 - 실제 운영에서는 페이징 필요)
func (r *UserRepository) ListAllUsers(ctx context.Context) ([]*domain.User, error) {
	log.Printf("Loading all users from event store")

	// 이 메서드는 실제 운영에서는 비효율적이므로 간단한 구현만 제공
	// 실제로는 별도의 사용자 목록 관리나 Read Model을 사용해야 함

	// 현재는 빈 목록 반환 (실제 구현에서는 Read Model이나 별도 인덱스 사용)
	log.Printf("ListAllUsers: This method should use Read Models in production")
	return []*domain.User{}, nil
}

// DeleteEventHistory 사용자의 이벤트 히스토리 삭제 (개발/테스트용)
func (r *UserRepository) DeleteEventHistory(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	log.Printf("Deleting event history for user %s", userID)

	// MongoEventStore에는 DeleteEvents 메서드가 없으므로
	// 실제 구현에서는 MongoDB 컬렉션에서 직접 삭제하거나
	// 별도의 삭제 메서드를 구현해야 함
	log.Printf("DeleteEventHistory: Not implemented - would require direct MongoDB access")
	return fmt.Errorf("delete event history not implemented")
}

// GetUserStats 사용자 통계 정보 조회 (데모용)
func (r *UserRepository) GetUserStats(ctx context.Context) (map[string]interface{}, error) {
	log.Printf("Getting user statistics")

	users, err := r.ListAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user statistics: %w", err)
	}

	stats := map[string]interface{}{
		"total_users":    len(users),
		"active_users":   0,
		"inactive_users": 0,
		"deleted_users":  0,
	}

	for _, user := range users {
		if user.IsDeleted() {
			stats["deleted_users"] = stats["deleted_users"].(int) + 1
		} else if user.IsActive() {
			stats["active_users"] = stats["active_users"].(int) + 1
		} else {
			stats["inactive_users"] = stats["inactive_users"].(int) + 1
		}
	}

	log.Printf("User statistics: %+v", stats)
	return stats, nil
}

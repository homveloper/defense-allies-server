package demo

import (
	"context"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/01-basic-event-sourcing/domain"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/01-basic-event-sourcing/infrastructure"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DemoScenarios 데모 시나리오 실행기
type DemoScenarios struct {
	infra *infrastructure.Infrastructure
}

// NewDemoScenarios 새로운 데모 시나리오 실행기 생성
func NewDemoScenarios(infra *infrastructure.Infrastructure) *DemoScenarios {
	return &DemoScenarios{
		infra: infra,
	}
}

// RunBasicCRUDScenario 기본 CRUD 시나리오 실행
func (d *DemoScenarios) RunBasicCRUDScenario(ctx context.Context) error {
	fmt.Println("\n=== Running Basic CRUD Scenario ===")

	// 1. 사용자 생성
	userID := uuid.New().String()
	user := domain.NewUser()

	err := user.CreateUser(userID, "John Doe", "john.doe@example.com")
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	err = d.infra.UserRepo.Save(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	fmt.Printf("✓ Created user: %s\n", user.String())

	// 2. 사용자 조회
	loadedUser, err := d.infra.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to load user: %w", err)
	}

	fmt.Printf("✓ Loaded user: %s\n", loadedUser.String())

	// 3. 사용자 활성화
	err = loadedUser.ActivateUser("system")
	if err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	err = d.infra.UserRepo.Save(ctx, loadedUser)
	if err != nil {
		return fmt.Errorf("failed to save activated user: %w", err)
	}

	fmt.Printf("✓ Activated user: %s\n", loadedUser.String())

	// 4. 사용자 정보 업데이트
	err = loadedUser.UpdateUser("John Smith", "john.smith@example.com")
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	err = d.infra.UserRepo.Save(ctx, loadedUser)
	if err != nil {
		return fmt.Errorf("failed to save updated user: %w", err)
	}

	fmt.Printf("✓ Updated user: %s\n", loadedUser.String())

	// 5. 이벤트 히스토리 확인
	events, err := d.infra.UserRepo.GetEventHistory(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get event history: %w", err)
	}

	fmt.Printf("✓ Event history (%d events):\n", len(events))
	for i, event := range events {
		fmt.Printf("  %d. %s (v%d) - %v\n", i+1, event.EventType(), event.Version(), event.Timestamp())
	}

	return nil
}

// RunEventRestorationScenario 이벤트 기반 복원 시나리오 실행
func (d *DemoScenarios) RunEventRestorationScenario(ctx context.Context) error {
	fmt.Println("\n=== Running Event Restoration Scenario ===")

	// 1. 여러 이벤트가 있는 사용자 생성
	userID := uuid.New().String()
	user := domain.NewUser()

	// 사용자 생성
	err := user.CreateUser(userID, "Alice Johnson", "alice@example.com")
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	err = d.infra.UserRepo.Save(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	// 여러 번의 업데이트
	updates := []struct {
		name  string
		email string
	}{
		{"Alice Smith", "alice.smith@example.com"},
		{"Alice Brown", "alice.brown@example.com"},
		{"Alice Wilson", "alice.wilson@example.com"},
	}

	for i, update := range updates {
		loadedUser, err := d.infra.UserRepo.GetByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to load user for update %d: %w", i+1, err)
		}

		err = loadedUser.UpdateUser(update.name, update.email)
		if err != nil {
			return fmt.Errorf("failed to update user %d: %w", i+1, err)
		}

		err = d.infra.UserRepo.Save(ctx, loadedUser)
		if err != nil {
			return fmt.Errorf("failed to save user update %d: %w", i+1, err)
		}

		fmt.Printf("✓ Update %d: %s\n", i+1, loadedUser.String())
	}

	// 활성화
	loadedUser, err := d.infra.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to load user for activation: %w", err)
	}

	err = loadedUser.ActivateUser("admin")
	if err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	err = d.infra.UserRepo.Save(ctx, loadedUser)
	if err != nil {
		return fmt.Errorf("failed to save activated user: %w", err)
	}

	// 최신 상태로 다시 로드하여 정확한 상태 확인
	finalUser, err := d.infra.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to reload final user state: %w", err)
	}

	fmt.Printf("✓ Final state: %s\n", finalUser.String())

	// 2. 이벤트 히스토리 확인
	events, err := d.infra.UserRepo.GetEventHistory(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get event history: %w", err)
	}

	fmt.Printf("\n✓ Complete event history (%d events):\n", len(events))
	for i, event := range events {
		fmt.Printf("  %d. %s (v%d) at %v\n", i+1, event.EventType(), event.Version(),
			event.Timestamp().Format("15:04:05"))
	}

	// 3. 새로운 인스턴스에서 복원 테스트
	fmt.Println("\n✓ Testing restoration from events...")
	restoredUser := domain.NewUserWithID(userID)

	for _, event := range events {
		err := restoredUser.Apply(event)
		if err != nil {
			return fmt.Errorf("failed to apply event %s: %w", event.EventType(), err)
		}
	}

	fmt.Printf("✓ Restored user: %s\n", restoredUser.String())

	// 4. 원본과 복원된 사용자 비교 (최신 상태와 비교)
	if finalUser.Name() != restoredUser.Name() ||
		finalUser.Email() != restoredUser.Email() ||
		finalUser.IsActive() != restoredUser.IsActive() ||
		finalUser.Version() != restoredUser.Version() {
		return fmt.Errorf("restored user state does not match original")
	}

	fmt.Println("✓ Restoration verification passed!")
	return nil
}

// RunConcurrencyScenario 동시성 처리 시나리오 실행
func (d *DemoScenarios) RunConcurrencyScenario(ctx context.Context) error {
	fmt.Println("\n=== Running Concurrency Scenario ===")

	// 1. 사용자 생성
	userID := uuid.New().String()
	user := domain.NewUser()

	err := user.CreateUser(userID, "Bob Wilson", "bob@example.com")
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	err = d.infra.UserRepo.Save(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	fmt.Printf("✓ Created user: %s\n", user.String())

	// 2. 두 개의 독립적인 인스턴스로 동시 업데이트 시뮬레이션
	user1, err := d.infra.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to load user1: %w", err)
	}

	user2, err := d.infra.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to load user2: %w", err)
	}

	fmt.Printf("✓ Loaded two instances: v%d and v%d\n", user1.Version(), user2.Version())

	// 3. 첫 번째 인스턴스 업데이트 (성공해야 함)
	err = user1.UpdateUser("Bob Smith", "bob.smith@example.com")
	if err != nil {
		return fmt.Errorf("failed to update user1: %w", err)
	}

	err = d.infra.UserRepo.Save(ctx, user1)
	if err != nil {
		return fmt.Errorf("failed to save user1: %w", err)
	}

	fmt.Printf("✓ First update successful: %s\n", user1.String())

	// 4. 두 번째 인스턴스 업데이트 (버전 충돌로 실패해야 함)
	err = user2.UpdateUser("Bob Johnson", "bob.johnson@example.com")
	if err != nil {
		return fmt.Errorf("failed to update user2: %w", err)
	}

	err = d.infra.UserRepo.Save(ctx, user2)
	if err != nil {
		fmt.Printf("✓ Expected concurrency error: %v\n", err)

		// 최신 상태로 다시 로드하여 업데이트
		latestUser, err := d.infra.UserRepo.GetByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to reload user: %w", err)
		}

		err = latestUser.UpdateUser("Bob Johnson", "bob.johnson@example.com")
		if err != nil {
			return fmt.Errorf("failed to update latest user: %w", err)
		}

		err = d.infra.UserRepo.Save(ctx, latestUser)
		if err != nil {
			return fmt.Errorf("failed to save latest user: %w", err)
		}

		fmt.Printf("✓ Retry successful: %s\n", latestUser.String())
	} else {
		return fmt.Errorf("expected concurrency error but save succeeded")
	}

	return nil
}

// RunPerformanceScenario 성능 테스트 시나리오 실행
func (d *DemoScenarios) RunPerformanceScenario(ctx context.Context) error {
	fmt.Println("\n=== Running Performance Scenario ===")

	userCount := 10
	eventsPerUser := 5

	fmt.Printf("Creating %d users with %d events each...\n", userCount, eventsPerUser)

	start := time.Now()

	// 여러 사용자 생성 및 이벤트 생성
	userIDs := make([]string, userCount)
	for i := 0; i < userCount; i++ {
		userID := uuid.New().String()
		userIDs[i] = userID

		user := domain.NewUser()
		err := user.CreateUser(userID, fmt.Sprintf("User %d", i+1), fmt.Sprintf("user%d@example.com", i+1))
		if err != nil {
			return fmt.Errorf("failed to create user %d: %w", i+1, err)
		}

		err = d.infra.UserRepo.Save(ctx, user)
		if err != nil {
			return fmt.Errorf("failed to save user %d: %w", i+1, err)
		}

		// 각 사용자에 대해 여러 이벤트 생성
		for j := 1; j < eventsPerUser; j++ {
			loadedUser, err := d.infra.UserRepo.GetByID(ctx, userID)
			if err != nil {
				return fmt.Errorf("failed to load user %d for event %d: %w", i+1, j+1, err)
			}

			err = loadedUser.UpdateUser(fmt.Sprintf("User %d Updated %d", i+1, j),
				fmt.Sprintf("user%d.%d@example.com", i+1, j))
			if err != nil {
				return fmt.Errorf("failed to update user %d event %d: %w", i+1, j+1, err)
			}

			err = d.infra.UserRepo.Save(ctx, loadedUser)
			if err != nil {
				return fmt.Errorf("failed to save user %d event %d: %w", i+1, j+1, err)
			}
		}
	}

	creationTime := time.Since(start)
	fmt.Printf("✓ Created %d users with %d events in %v\n", userCount, userCount*eventsPerUser, creationTime)

	// 모든 사용자 로드 성능 테스트
	start = time.Now()
	users, err := d.infra.UserRepo.ListAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list all users: %w", err)
	}

	loadTime := time.Since(start)
	fmt.Printf("✓ Loaded %d users in %v\n", len(users), loadTime)

	// 개별 사용자 로드 성능 테스트
	start = time.Now()
	for _, userID := range userIDs {
		_, err := d.infra.UserRepo.GetByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to load user %s: %w", userID, err)
		}
	}

	individualLoadTime := time.Since(start)
	fmt.Printf("✓ Individually loaded %d users in %v (avg: %v per user)\n",
		len(userIDs), individualLoadTime, individualLoadTime/time.Duration(len(userIDs)))

	return nil
}

// RunAllScenarios 모든 시나리오 실행
func (d *DemoScenarios) RunAllScenarios(ctx context.Context) error {
	scenarios := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"Basic CRUD", d.RunBasicCRUDScenario},
		{"Event Restoration", d.RunEventRestorationScenario},
		{"Concurrency", d.RunConcurrencyScenario},
		{"Performance", d.RunPerformanceScenario},
	}

	for _, scenario := range scenarios {
		fmt.Printf("\n" + strings.Repeat("=", 50))
		fmt.Printf("\nRunning scenario: %s\n", scenario.name)
		fmt.Printf(strings.Repeat("=", 50))

		err := scenario.fn(ctx)
		if err != nil {
			return fmt.Errorf("scenario '%s' failed: %w", scenario.name, err)
		}

		fmt.Printf("\n✅ Scenario '%s' completed successfully!\n", scenario.name)
	}

	return nil
}

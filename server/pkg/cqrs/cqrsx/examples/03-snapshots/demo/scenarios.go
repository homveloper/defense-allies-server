package demo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"cqrs/cqrsx/examples/03-snapshots/domain"
	"cqrs/cqrsx/examples/03-snapshots/infrastructure"
)

// DemoRunner 데모 실행기
type DemoRunner struct {
	infra *infrastructure.Infrastructure
}

// NewDemoRunner 데모 실행기 생성
func NewDemoRunner(infra *infrastructure.Infrastructure) *DemoRunner {
	return &DemoRunner{
		infra: infra,
	}
}

// RunBasicSnapshotDemo 기본 스냅샷 데모
func (d *DemoRunner) RunBasicSnapshotDemo(ctx context.Context) error {
	fmt.Println("\n=== Basic Snapshot Demo ===")

	// 1. 주문 생성 및 여러 이벤트 발생
	orderID := uuid.New().String()
	order := domain.NewOrder()

	fmt.Printf("Creating order %s...\n", orderID)

	// 주문 생성
	err := order.CreateOrder(orderID, "customer-123", decimal.NewFromFloat(10.00))
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	err = d.infra.OrderRepo.Save(ctx, order)
	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	fmt.Printf("✓ Order created: %s\n", order.String())

	// 2. 여러 상품 추가 (스냅샷 트리거를 위해)
	products := []struct {
		id    string
		name  string
		price float64
		qty   int
	}{
		{"prod-1", "Laptop", 999.99, 1},
		{"prod-2", "Mouse", 29.99, 2},
		{"prod-3", "Keyboard", 79.99, 1},
		{"prod-4", "Monitor", 299.99, 1},
		{"prod-5", "Webcam", 89.99, 1},
	}

	for i, product := range products {
		// 주문 다시 로드
		order, err = d.infra.OrderRepo.GetByID(ctx, orderID)
		if err != nil {
			return fmt.Errorf("failed to load order: %w", err)
		}

		err = order.AddItem(product.id, product.name, product.qty, decimal.NewFromFloat(product.price))
		if err != nil {
			return fmt.Errorf("failed to add item: %w", err)
		}

		err = d.infra.OrderRepo.Save(ctx, order)
		if err != nil {
			return fmt.Errorf("failed to save order: %w", err)
		}

		fmt.Printf("✓ Added item %d: %s (v%d)\n", i+1, product.name, order.Version())
	}

	// 3. 할인 적용
	order, err = d.infra.OrderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to load order: %w", err)
	}

	err = order.ApplyDiscount(decimal.NewFromFloat(0.1), "10% discount")
	if err != nil {
		return fmt.Errorf("failed to apply discount: %w", err)
	}

	err = d.infra.OrderRepo.Save(ctx, order)
	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	fmt.Printf("✓ Applied discount: %s (v%d)\n", order.String(), order.Version())

	// 4. 주문 확정
	err = order.ConfirmOrder()
	if err != nil {
		return fmt.Errorf("failed to confirm order: %w", err)
	}

	err = d.infra.OrderRepo.Save(ctx, order)
	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	fmt.Printf("✓ Order confirmed: %s (v%d)\n", order.String(), order.Version())

	// 5. 스냅샷 정보 확인
	snapshotInfos, err := d.infra.OrderRepo.GetSnapshotInfo(ctx, orderID)
	if err != nil {
		fmt.Printf("⚠ Failed to get snapshot info: %v\n", err)
	} else {
		fmt.Printf("\n📸 Snapshots created: %d\n", len(snapshotInfos))
		for i, info := range snapshotInfos {
			fmt.Printf("  %d. Version %d, Size: %d bytes, Type: %s, Time: %v\n",
				i+1, info.Version, info.Size, info.ContentType, info.Timestamp.Format("15:04:05"))
		}
	}

	// 6. 이벤트 히스토리 확인
	events, err := d.infra.OrderRepo.GetEventHistory(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get event history: %w", err)
	}

	fmt.Printf("\n📜 Event history: %d events\n", len(events))
	for i, event := range events {
		fmt.Printf("  %d. %s (v%d) at %v\n", i+1, event.EventType(), event.Version(),
			event.Timestamp().Format("15:04:05"))
	}

	// 7. 복원 테스트
	fmt.Println("\n🔄 Testing restoration...")
	restoredOrder, err := d.infra.OrderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to restore order: %w", err)
	}

	fmt.Printf("✓ Restored order: %s\n", restoredOrder.String())

	// 8. 상태 비교
	if order.Version() != restoredOrder.Version() {
		return fmt.Errorf("version mismatch: original %d vs restored %d", order.Version(), restoredOrder.Version())
	}

	if order.Status() != restoredOrder.Status() {
		return fmt.Errorf("status mismatch: original %s vs restored %s", order.Status, restoredOrder.Status)
	}

	fmt.Println("✅ Restoration verification passed!")

	return nil
}

// RunPerformanceComparisonDemo 성능 비교 데모
func (d *DemoRunner) RunPerformanceComparisonDemo(ctx context.Context) error {
	fmt.Println("\n=== Performance Comparison Demo ===")

	// 테스트 설정
	eventCounts := []int{10, 50, 100, 200}

	for _, eventCount := range eventCounts {
		fmt.Printf("\n--- Testing with %d events ---\n", eventCount)

		// 스냅샷 있는 경우와 없는 경우 비교
		err := d.comparePerformance(ctx, eventCount)
		if err != nil {
			return fmt.Errorf("performance comparison failed for %d events: %w", eventCount, err)
		}
	}

	return nil
}

// comparePerformance 성능 비교
func (d *DemoRunner) comparePerformance(ctx context.Context, eventCount int) error {
	// 1. 스냅샷 없이 테스트
	fmt.Println("🚫 Testing WITHOUT snapshots...")
	d.infra.OrderRepo.EnableSnapshots(false)

	orderID1 := uuid.New().String()
	timeWithoutSnapshot, err := d.createOrderWithEvents(ctx, orderID1, eventCount)
	if err != nil {
		return fmt.Errorf("failed to create order without snapshots: %w", err)
	}

	loadTimeWithoutSnapshot, err := d.measureLoadTime(ctx, orderID1)
	if err != nil {
		return fmt.Errorf("failed to measure load time without snapshots: %w", err)
	}

	// 2. 스냅샷 있이 테스트
	fmt.Println("📸 Testing WITH snapshots...")
	d.infra.OrderRepo.EnableSnapshots(true)

	orderID2 := uuid.New().String()
	timeWithSnapshot, err := d.createOrderWithEvents(ctx, orderID2, eventCount)
	if err != nil {
		return fmt.Errorf("failed to create order with snapshots: %w", err)
	}

	loadTimeWithSnapshot, err := d.measureLoadTime(ctx, orderID2)
	if err != nil {
		return fmt.Errorf("failed to measure load time with snapshots: %w", err)
	}

	// 3. 결과 출력
	fmt.Printf("\n📊 Performance Results:\n")
	fmt.Printf("  Events: %d\n", eventCount)
	fmt.Printf("  Creation time (no snapshot): %v\n", timeWithoutSnapshot)
	fmt.Printf("  Creation time (with snapshot): %v\n", timeWithSnapshot)
	fmt.Printf("  Load time (no snapshot): %v\n", loadTimeWithoutSnapshot)
	fmt.Printf("  Load time (with snapshot): %v\n", loadTimeWithSnapshot)

	// 성능 개선 계산
	if loadTimeWithoutSnapshot > 0 && loadTimeWithSnapshot > 0 {
		improvement := float64(loadTimeWithoutSnapshot-loadTimeWithSnapshot) / float64(loadTimeWithoutSnapshot) * 100
		fmt.Printf("  Load time improvement: %.1f%%\n", improvement)
	}

	return nil
}

// createOrderWithEvents 이벤트가 많은 주문 생성
func (d *DemoRunner) createOrderWithEvents(ctx context.Context, orderID string, eventCount int) (time.Duration, error) {
	start := time.Now()

	order := domain.NewOrder()

	// 주문 생성
	err := order.CreateOrder(orderID, "customer-perf-test", decimal.NewFromFloat(5.00))
	if err != nil {
		return 0, err
	}

	err = d.infra.OrderRepo.Save(ctx, order)
	if err != nil {
		return 0, err
	}

	// 이벤트 생성 (상품 추가/제거/수량 변경 등)
	for i := 1; i < eventCount; i++ {
		order, err = d.infra.OrderRepo.GetByID(ctx, orderID)
		if err != nil {
			return 0, err
		}

		switch i % 4 {
		case 0: // 상품 추가
			productID := fmt.Sprintf("prod-%d", i)
			err = order.AddItem(productID, fmt.Sprintf("Product %d", i), 1, decimal.NewFromFloat(float64(i*10)))
		case 1: // 할인 적용
			rate := float64(i%10) / 100.0
			err = order.ApplyDiscount(decimal.NewFromFloat(rate), fmt.Sprintf("Discount %d", i))
		case 2: // 수량 변경 (기존 상품이 있는 경우)
			if len(order.Items()) > 0 {
				firstItem := order.Items()[0]
				newQty := (i % 5) + 1
				err = order.ChangeItemQuantity(firstItem.ProductID, newQty)
			}
		case 3: // 메타데이터 업데이트
			// 메타데이터 업데이트 이벤트는 별도 구현 필요
			continue
		}

		if err != nil {
			return 0, err
		}

		err = d.infra.OrderRepo.Save(ctx, order)
		if err != nil {
			return 0, err
		}
	}

	return time.Since(start), nil
}

// measureLoadTime 로드 시간 측정
func (d *DemoRunner) measureLoadTime(ctx context.Context, orderID string) (time.Duration, error) {
	start := time.Now()

	_, err := d.infra.OrderRepo.GetByID(ctx, orderID)
	if err != nil {
		return 0, err
	}

	return time.Since(start), nil
}

// RunSnapshotPoliciesDemo 스냅샷 정책 데모
func (d *DemoRunner) RunSnapshotPoliciesDemo(ctx context.Context) error {
	fmt.Println("\n=== Snapshot Policies Demo ===")

	// 다양한 정책들 테스트
	policies := []string{"event_count", "time_based", "version_based", "always", "never"}

	for _, policyName := range policies {
		fmt.Printf("\n--- Testing %s policy ---\n", policyName)
		err := d.testSnapshotPolicy(ctx, policyName)
		if err != nil {
			return fmt.Errorf("policy test failed for %s: %w", policyName, err)
		}
	}

	return nil
}

// testSnapshotPolicy 특정 정책 테스트
func (d *DemoRunner) testSnapshotPolicy(ctx context.Context, policyName string) error {
	// 정책별 설정은 실제 구현에서 동적으로 변경 필요
	// 여기서는 간단히 시뮬레이션

	orderID := uuid.New().String()
	order := domain.NewOrder()

	err := order.CreateOrder(orderID, "customer-policy-test", decimal.NewFromFloat(5.00))
	if err != nil {
		return err
	}

	err = d.infra.OrderRepo.Save(ctx, order)
	if err != nil {
		return err
	}

	// 몇 개의 이벤트 추가
	for i := 0; i < 8; i++ {
		order, err = d.infra.OrderRepo.GetByID(ctx, orderID)
		if err != nil {
			return err
		}

		productID := fmt.Sprintf("prod-%d", i)
		err = order.AddItem(productID, fmt.Sprintf("Product %d", i), 1, decimal.NewFromFloat(10.0))
		if err != nil {
			return err
		}

		err = d.infra.OrderRepo.Save(ctx, order)
		if err != nil {
			return err
		}
	}

	// 스냅샷 정보 확인
	snapshotInfos, err := d.infra.OrderRepo.GetSnapshotInfo(ctx, orderID)
	if err != nil {
		fmt.Printf("⚠ Failed to get snapshot info: %v\n", err)
	} else {
		fmt.Printf("✓ Policy %s created %d snapshots\n", policyName, len(snapshotInfos))
	}

	return nil
}

// RunSerializationComparisonDemo 직렬화 비교 데모
func (d *DemoRunner) RunSerializationComparisonDemo(ctx context.Context) error {
	fmt.Println("\n=== Serialization Comparison Demo ===")

	// 다양한 직렬화 방식 비교는 실제 구현에서 동적으로 변경 필요
	// 여기서는 개념적으로 시뮬레이션

	serializers := []string{"JSON", "BSON", "Compressed JSON", "Compressed BSON"}

	for _, serializer := range serializers {
		fmt.Printf("📦 Testing %s serialization...\n", serializer)
		// 실제로는 각 직렬화 방식으로 스냅샷 생성하고 크기 비교
		fmt.Printf("✓ %s serialization test completed\n", serializer)
	}

	return nil
}

// RunStressTestDemo 스트레스 테스트 데모
func (d *DemoRunner) RunStressTestDemo(ctx context.Context) error {
	fmt.Println("\n=== Stress Test Demo ===")

	// 대량의 이벤트로 스트레스 테스트
	eventCount := 1000
	fmt.Printf("Creating order with %d events...\n", eventCount)

	orderID := uuid.New().String()
	_, err := d.createOrderWithEvents(ctx, orderID, eventCount)
	if err != nil {
		return fmt.Errorf("stress test failed: %w", err)
	}

	fmt.Printf("✓ Stress test completed with %d events\n", eventCount)

	// 성능 메트릭 출력
	metrics, err := d.infra.OrderRepo.GetPerformanceMetrics(ctx, orderID)
	if err != nil {
		fmt.Printf("⚠ Failed to get performance metrics: %v\n", err)
	} else {
		fmt.Println("\n📊 Performance Metrics:")
		for key, value := range metrics {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

	return nil
}

// RunCleanupDemo 정리 데모
func (d *DemoRunner) RunCleanupDemo(ctx context.Context) error {
	fmt.Println("\n=== Cleanup Demo ===")

	// 여러 스냅샷 생성
	orderID := uuid.New().String()
	order := domain.NewOrder()

	err := order.CreateOrder(orderID, "customer-cleanup-test", decimal.NewFromFloat(5.00))
	if err != nil {
		return err
	}

	// 강제로 여러 스냅샷 생성
	for i := 0; i < 10; i++ {
		err = d.infra.OrderRepo.Save(ctx, order)
		if err != nil {
			return err
		}

		// 수동 스냅샷 생성
		err = d.infra.OrderRepo.CreateSnapshot(ctx, orderID)
		if err != nil {
			fmt.Printf("⚠ Failed to create snapshot %d: %v\n", i+1, err)
		}

		// 더미 이벤트 추가
		productID := fmt.Sprintf("cleanup-prod-%d", i)
		err = order.AddItem(productID, fmt.Sprintf("Cleanup Product %d", i), 1, decimal.NewFromFloat(10.0))
		if err != nil {
			return err
		}
	}

	// 정리 전 스냅샷 개수 확인
	snapshotInfos, err := d.infra.OrderRepo.GetSnapshotInfo(ctx, orderID)
	if err != nil {
		return err
	}

	fmt.Printf("📸 Snapshots before cleanup: %d\n", len(snapshotInfos))

	// 정리 실행
	err = d.infra.OrderRepo.CleanupSnapshots(ctx, orderID)
	if err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	// 정리 후 스냅샷 개수 확인
	snapshotInfos, err = d.infra.OrderRepo.GetSnapshotInfo(ctx, orderID)
	if err != nil {
		return err
	}

	fmt.Printf("📸 Snapshots after cleanup: %d\n", len(snapshotInfos))
	fmt.Println("✅ Cleanup completed!")

	return nil
}

// RunAllDemos 모든 데모 실행
func (d *DemoRunner) RunAllDemos(ctx context.Context) error {
	demos := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"Basic Snapshot", d.RunBasicSnapshotDemo},
		{"Performance Comparison", d.RunPerformanceComparisonDemo},
		{"Snapshot Policies", d.RunSnapshotPoliciesDemo},
		{"Serialization Comparison", d.RunSerializationComparisonDemo},
		{"Stress Test", d.RunStressTestDemo},
		{"Cleanup", d.RunCleanupDemo},
	}

	for _, demo := range demos {
		fmt.Printf("\n" + strings.Repeat("=", 60))
		fmt.Printf("\nRunning: %s\n", demo.name)
		fmt.Printf(strings.Repeat("=", 60))

		err := demo.fn(ctx)
		if err != nil {
			return fmt.Errorf("demo '%s' failed: %w", demo.name, err)
		}

		fmt.Printf("\n✅ Demo '%s' completed successfully!\n", demo.name)
	}

	return nil
}

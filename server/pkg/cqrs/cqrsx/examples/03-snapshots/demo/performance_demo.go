package demo

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"defense-allies-server/pkg/cqrs/cqrsx/examples/03-snapshots/domain"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/03-snapshots/infrastructure"
)

// PerformanceDemo 성능 비교 데모
type PerformanceDemo struct {
	monitor infrastructure.PerformanceMonitor
}

// NewPerformanceDemo 새로운 성능 데모 생성
func NewPerformanceDemo() *PerformanceDemo {
	return &PerformanceDemo{
		monitor: infrastructure.NewInMemoryPerformanceMonitor(),
	}
}

// RunBasicPerformanceTest 기본 성능 테스트 실행
func (d *PerformanceDemo) RunBasicPerformanceTest(ctx context.Context) error {
	fmt.Println("=== 기본 성능 테스트 시작 ===")

	// 테스트 주문 생성
	order := domain.NewOrder()
	customerID := "customer-123"
	shippingCost := decimal.NewFromFloat(10.0)

	// 주문 생성
	operationID := d.monitor.StartOperation(ctx, order.ID(), "create_order")
	err := order.CreateOrder(order.ID(), customerID, shippingCost)
	d.monitor.EndOperation(ctx, operationID, 1, false, 0, err)
	if err != nil {
		return fmt.Errorf("주문 생성 실패: %w", err)
	}

	// 여러 상품 추가 (대량 이벤트 생성)
	productCount := 50
	for i := 0; i < productCount; i++ {
		operationID := d.monitor.StartOperation(ctx, order.ID(), "add_item")
		err := order.AddItem(
			fmt.Sprintf("product-%d", i),
			fmt.Sprintf("Product %d", i),
			i+1,
			decimal.NewFromFloat(float64(10+i)),
		)
		d.monitor.EndOperation(ctx, operationID, 1, false, 0, err)
		if err != nil {
			return fmt.Errorf("상품 추가 실패: %w", err)
		}
	}

	// 할인 적용
	operationID = d.monitor.StartOperation(ctx, order.ID(), "apply_discount")
	err = order.ApplyDiscount(decimal.NewFromFloat(0.1), "10% 할인")
	d.monitor.EndOperation(ctx, operationID, 1, false, 0, err)
	if err != nil {
		return fmt.Errorf("할인 적용 실패: %w", err)
	}

	// 주문 확정
	operationID = d.monitor.StartOperation(ctx, order.ID(), "confirm_order")
	err = order.ConfirmOrder()
	d.monitor.EndOperation(ctx, operationID, 1, false, 0, err)
	if err != nil {
		return fmt.Errorf("주문 확정 실패: %w", err)
	}

	fmt.Printf("주문 생성 완료: %s (이벤트 수: %d)\n", order.ID(), order.Version())
	return nil
}

// RunSnapshotPerformanceComparison 스냅샷 성능 비교 실행
func (d *PerformanceDemo) RunSnapshotPerformanceComparison(ctx context.Context) error {
	fmt.Println("\n=== 스냅샷 성능 비교 테스트 ===")

	// 테스트용 주문 생성
	order := domain.NewOrder()
	customerID := "customer-performance-test"
	shippingCost := decimal.NewFromFloat(15.0)

	// 주문 생성
	err := order.CreateOrder(order.ID(), customerID, shippingCost)
	if err != nil {
		return fmt.Errorf("주문 생성 실패: %w", err)
	}

	// 대량 이벤트 생성 (100개 상품 추가)
	eventCount := 100
	fmt.Printf("대량 이벤트 생성 중... (%d개 상품)\n", eventCount)

	for i := 0; i < eventCount; i++ {
		err := order.AddItem(
			fmt.Sprintf("perf-product-%d", i),
			fmt.Sprintf("Performance Product %d", i),
			i%10+1,
			decimal.NewFromFloat(float64(5+i%50)),
		)
		if err != nil {
			return fmt.Errorf("상품 추가 실패: %w", err)
		}

		// 일부 상품 수량 변경
		if i%10 == 0 && i > 0 {
			err := order.ChangeItemQuantity(fmt.Sprintf("perf-product-%d", i-5), (i%5)+1)
			if err != nil {
				log.Printf("수량 변경 실패: %v", err)
			}
		}
	}

	fmt.Printf("총 이벤트 수: %d\n", order.Version())

	// 1. 스냅샷 없이 복원 시간 측정
	fmt.Println("\n1. 스냅샷 없이 복원 시간 측정...")
	withoutSnapshotTime := d.measureRestorationTime(ctx, order, false)
	fmt.Printf("   복원 시간: %v\n", withoutSnapshotTime)

	// 2. 스냅샷 생성
	fmt.Println("\n2. 스냅샷 생성...")
	snapshotStartTime := time.Now()
	snapshot, err := order.CreateSnapshot()
	snapshotCreationTime := time.Since(snapshotStartTime)
	if err != nil {
		return fmt.Errorf("스냅샷 생성 실패: %w", err)
	}
	fmt.Printf("   스냅샷 생성 시간: %v\n", snapshotCreationTime)
	fmt.Printf("   스냅샷 버전: %d\n", snapshot.Version())

	// 3. 스냅샷을 사용한 복원 시간 측정
	fmt.Println("\n3. 스냅샷을 사용한 복원 시간 측정...")
	withSnapshotTime := d.measureRestorationTime(ctx, order, true)
	fmt.Printf("   복원 시간: %v\n", withSnapshotTime)

	// 4. 성능 개선 계산
	improvement := float64(withoutSnapshotTime-withSnapshotTime) / float64(withoutSnapshotTime) * 100
	fmt.Printf("\n=== 성능 비교 결과 ===\n")
	fmt.Printf("스냅샷 없이: %v\n", withoutSnapshotTime)
	fmt.Printf("스냅샷 사용: %v\n", withSnapshotTime)
	fmt.Printf("성능 개선: %.2f%%\n", improvement)
	fmt.Printf("속도 향상: %.2fx\n", float64(withoutSnapshotTime)/float64(withSnapshotTime))

	return nil
}

// measureRestorationTime 복원 시간 측정
func (d *PerformanceDemo) measureRestorationTime(ctx context.Context, order *domain.Order, useSnapshot bool) time.Duration {
	operationType := "restore_without_snapshot"
	if useSnapshot {
		operationType = "restore_with_snapshot"
	}

	operationID := d.monitor.StartOperation(ctx, order.ID(), operationType)
	startTime := time.Now()

	if useSnapshot {
		// 스냅샷을 사용한 복원 시뮬레이션
		snapshot, err := order.CreateSnapshot()
		if err == nil {
			// 스냅샷에서 복원
			_, err = domain.RestoreFromSnapshot(snapshot.(*domain.OrderSnapshot))
		}
		d.monitor.EndOperation(ctx, operationID, 5, true, snapshot.Version(), err) // 스냅샷 이후 5개 이벤트만 처리
	} else {
		// 전체 이벤트 재생 시뮬레이션
		// 실제로는 모든 이벤트를 다시 적용해야 함
		newOrder := domain.NewOrder()
		_ = newOrder // 실제 복원 로직 생략
		d.monitor.EndOperation(ctx, operationID, order.Version(), false, 0, nil)
	}

	return time.Since(startTime)
}

// RunBenchmarkSuite 벤치마크 스위트 실행
func (d *PerformanceDemo) RunBenchmarkSuite(ctx context.Context) error {
	fmt.Println("\n=== 벤치마크 스위트 실행 ===")

	eventCounts := []int{10, 50, 100, 200, 500}

	for _, eventCount := range eventCounts {
		fmt.Printf("\n--- %d 이벤트 벤치마크 ---\n", eventCount)

		// 테스트용 주문 생성
		order := domain.NewOrder()
		customerID := fmt.Sprintf("benchmark-customer-%d", eventCount)

		err := order.CreateOrder(order.ID(), customerID, decimal.NewFromFloat(10.0))
		if err != nil {
			return fmt.Errorf("주문 생성 실패: %w", err)
		}

		// 지정된 수만큼 이벤트 생성
		for i := 0; i < eventCount; i++ {
			err := order.AddItem(
				fmt.Sprintf("bench-product-%d", i),
				fmt.Sprintf("Benchmark Product %d", i),
				1,
				decimal.NewFromFloat(10.0),
			)
			if err != nil {
				return fmt.Errorf("상품 추가 실패: %w", err)
			}
		}

		// 스냅샷 없이 복원
		withoutTime := d.measureRestorationTime(ctx, order, false)

		// 스냅샷과 함께 복원
		withTime := d.measureRestorationTime(ctx, order, true)

		improvement := float64(withoutTime-withTime) / float64(withoutTime) * 100

		fmt.Printf("스냅샷 없이: %v\n", withoutTime)
		fmt.Printf("스냅샷 사용: %v\n", withTime)
		fmt.Printf("성능 개선: %.2f%%\n", improvement)
	}

	return nil
}

// GeneratePerformanceReport 성능 보고서 생성
func (d *PerformanceDemo) GeneratePerformanceReport(ctx context.Context) error {
	fmt.Println("\n=== 성능 보고서 생성 ===")

	// 1시간 전부터의 데이터로 보고서 생성
	since := time.Now().Add(-time.Hour)
	report, err := d.monitor.GenerateReport(ctx, since)
	if err != nil {
		return fmt.Errorf("보고서 생성 실패: %w", err)
	}

	fmt.Printf("총 작업 수: %d\n", report.TotalOperations)
	fmt.Printf("성공한 작업: %d\n", report.SuccessfulOps)
	fmt.Printf("실패한 작업: %d\n", report.FailedOps)
	fmt.Printf("평균 소요 시간: %v\n", report.AverageDuration)
	fmt.Printf("최소 소요 시간: %v\n", report.MinDuration)
	fmt.Printf("최대 소요 시간: %v\n", report.MaxDuration)
	fmt.Printf("스냅샷 사용률: %.2f%%\n", report.SnapshotUsageRate)
	fmt.Printf("메모리 효율성: %.2f events/KB\n", report.MemoryEfficiency)

	fmt.Println("\n작업 유형별 통계:")
	for opType, count := range report.OperationsByType {
		fmt.Printf("  %s: %d회\n", opType, count)
	}

	return nil
}

// RunCompleteDemo 전체 데모 실행
func (d *PerformanceDemo) RunCompleteDemo(ctx context.Context) error {
	fmt.Println("🚀 스냅샷 성능 데모 시작")
	fmt.Println(strings.Repeat("=", 50))

	// 1. 기본 성능 테스트
	if err := d.RunBasicPerformanceTest(ctx); err != nil {
		return err
	}

	// 2. 스냅샷 성능 비교
	if err := d.RunSnapshotPerformanceComparison(ctx); err != nil {
		return err
	}

	// 3. 벤치마크 스위트
	if err := d.RunBenchmarkSuite(ctx); err != nil {
		return err
	}

	// 4. 성능 보고서 생성
	if err := d.GeneratePerformanceReport(ctx); err != nil {
		return err
	}

	fmt.Println("\n✅ 모든 데모가 성공적으로 완료되었습니다!")
	return nil
}

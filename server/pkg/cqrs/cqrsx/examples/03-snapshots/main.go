package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cqrs/cqrsx/examples/03-snapshots/demo"
	"cqrs/cqrsx/examples/03-snapshots/domain"
	"cqrs/cqrsx/examples/03-snapshots/infrastructure"

	"github.com/shopspring/decimal"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := strings.ToLower(os.Args[1])

	switch command {
	case "demo":
		if len(os.Args) < 3 {
			printDemoUsage()
			return
		}
		runDemo(os.Args[2])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Snapshots Example - Event Sourcing with Snapshots")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run . demo <scenario>")
	fmt.Println("  go run . help")
	fmt.Println()
	fmt.Println("Demo scenarios:")
	fmt.Println("  basic        - Basic snapshot creation and restoration")
	fmt.Println("  performance  - Performance comparison with/without snapshots")
	fmt.Println("  policies     - Different snapshot policies demonstration")
	fmt.Println("  serialization - Serialization methods comparison")
	fmt.Println("  stress       - Stress test with large number of events")
	fmt.Println("  cleanup      - Snapshot cleanup and maintenance")
	fmt.Println("  all          - Run all scenarios")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run . demo basic")
	fmt.Println("  go run . demo performance")
	fmt.Println("  go run . demo all")
}

func printDemoUsage() {
	fmt.Println("Available demo scenarios:")
	fmt.Println("  basic        - Basic snapshot functionality")
	fmt.Println("  performance  - Performance comparison")
	fmt.Println("  policies     - Snapshot policies")
	fmt.Println("  serialization - Serialization comparison")
	fmt.Println("  stress       - Stress testing")
	fmt.Println("  cleanup      - Cleanup operations")
	fmt.Println("  all          - Run all scenarios")
}

func runDemo(scenario string) {
	ctx := context.Background()

	// 시나리오 실행
	switch strings.ToLower(scenario) {
	case "basic":
		err := runBasicDemo(ctx)
		if err != nil {
			log.Fatalf("Basic demo failed: %v", err)
		}
	case "performance":
		err := runPerformanceDemo(ctx)
		if err != nil {
			log.Fatalf("Performance demo failed: %v", err)
		}
	case "policies":
		fmt.Println("Snapshot policies demo - Coming soon!")
	case "serialization":
		fmt.Println("Serialization comparison demo - Coming soon!")
	case "stress":
		fmt.Println("Stress test demo - Coming soon!")
	case "cleanup":
		fmt.Println("Cleanup demo - Coming soon!")
	case "all":
		fmt.Println("Running all demos...")
		if err := runBasicDemo(ctx); err != nil {
			log.Fatalf("Basic demo failed: %v", err)
		}
		if err := runPerformanceDemo(ctx); err != nil {
			log.Fatalf("Performance demo failed: %v", err)
		}
	default:
		fmt.Printf("Unknown scenario: %s\n", scenario)
		printDemoUsage()
		return
	}

	fmt.Println("\n🎉 Demo completed successfully!")
}

func runBasicDemo(ctx context.Context) error {
	fmt.Println("=== Basic Snapshot Demo ===")
	fmt.Println("스냅샷 기본 기능 시연을 시작합니다...\n")

	// 1. 주문 생성 및 이벤트 추가
	fmt.Println("1️⃣ 주문 생성 및 이벤트 추가")
	order := domain.NewOrder()
	customerID := "customer-snapshot-demo"
	shippingCost := decimal.NewFromFloat(15.0)

	// 주문 생성
	err := order.CreateOrder(order.ID(), customerID, shippingCost)
	if err != nil {
		return fmt.Errorf("주문 생성 실패: %w", err)
	}
	fmt.Printf("   ✅ 주문 생성: %s\n", order.ID())

	// 여러 상품 추가
	products := []struct {
		id    string
		name  string
		qty   int
		price float64
	}{
		{"laptop-001", "Gaming Laptop", 1, 1500.00},
		{"mouse-002", "Wireless Mouse", 2, 25.00},
		{"keyboard-003", "Mechanical Keyboard", 1, 120.00},
		{"monitor-004", "4K Monitor", 1, 400.00},
		{"headset-005", "Gaming Headset", 1, 80.00},
	}

	for _, product := range products {
		err := order.AddItem(product.id, product.name, product.qty, decimal.NewFromFloat(product.price))
		if err != nil {
			return fmt.Errorf("상품 추가 실패: %w", err)
		}
		fmt.Printf("   ✅ 상품 추가: %s (수량: %d, 가격: $%.2f)\n", product.name, product.qty, product.price)
	}

	// 할인 적용
	err = order.ApplyDiscount(decimal.NewFromFloat(0.1), "신규 고객 10% 할인")
	if err != nil {
		return fmt.Errorf("할인 적용 실패: %w", err)
	}
	fmt.Printf("   ✅ 할인 적용: 10%%\n")

	fmt.Printf("   📊 현재 상태: 버전 %d, 상품 %d개, 총액 $%.2f\n\n",
		order.Version(), order.ItemCount(), order.FinalAmount())

	// 2. 스냅샷 생성
	fmt.Println("2️⃣ 스냅샷 생성")
	snapshot, err := order.CreateSnapshot()
	if err != nil {
		return fmt.Errorf("스냅샷 생성 실패: %w", err)
	}

	orderSnapshot := snapshot.(*domain.OrderSnapshot)
	fmt.Printf("   ✅ 스냅샷 생성 완료\n")
	fmt.Printf("   📸 스냅샷 정보:\n")
	fmt.Printf("      - Aggregate ID: %s\n", orderSnapshot.ID())
	fmt.Printf("      - 버전: %d\n", orderSnapshot.Version())
	fmt.Printf("      - 고객 ID: %s\n", orderSnapshot.CustomerID)
	fmt.Printf("      - 상품 개수: %d\n", orderSnapshot.GetItemCount())
	fmt.Printf("      - 상태: %s\n", orderSnapshot.Status)

	totalAmount, _ := orderSnapshot.GetTotalAmountDecimal()
	finalAmount, _ := orderSnapshot.GetFinalAmountDecimal()
	fmt.Printf("      - 총액: $%.2f\n", totalAmount)
	fmt.Printf("      - 최종 금액: $%.2f\n", finalAmount)
	fmt.Printf("      - 생성 시간: %s\n\n", orderSnapshot.Timestamp().Format("2006-01-02 15:04:05"))

	// 3. 스냅샷 직렬화/역직렬화 테스트
	fmt.Println("3️⃣ 스냅샷 직렬화/역직렬화 테스트")
	serializedData, err := orderSnapshot.Serialize()
	if err != nil {
		return fmt.Errorf("스냅샷 직렬화 실패: %w", err)
	}
	fmt.Printf("   ✅ 직렬화 완료: %d bytes\n", len(serializedData))

	// 새로운 스냅샷 인스턴스에 역직렬화
	newSnapshot := &domain.OrderSnapshot{}
	err = newSnapshot.Deserialize(serializedData)
	if err != nil {
		return fmt.Errorf("스냅샷 역직렬화 실패: %w", err)
	}
	fmt.Printf("   ✅ 역직렬화 완료\n")
	fmt.Printf("   🔍 검증: Aggregate ID = %s, 버전 = %d\n\n",
		newSnapshot.ID(), newSnapshot.Version())

	// 4. 스냅샷에서 주문 복원
	fmt.Println("4️⃣ 스냅샷에서 주문 복원")
	restoredOrder, err := domain.RestoreFromSnapshot(newSnapshot)
	if err != nil {
		return fmt.Errorf("주문 복원 실패: %w", err)
	}

	fmt.Printf("   ✅ 주문 복원 완료\n")
	fmt.Printf("   🔍 복원된 주문 정보:\n")
	fmt.Printf("      - ID: %s\n", restoredOrder.ID())
	fmt.Printf("      - 버전: %d\n", restoredOrder.Version())
	fmt.Printf("      - 고객 ID: %s\n", restoredOrder.CustomerID())
	fmt.Printf("      - 상품 개수: %d\n", restoredOrder.ItemCount())
	fmt.Printf("      - 상태: %s\n", restoredOrder.Status())
	fmt.Printf("      - 총액: $%.2f\n", restoredOrder.TotalAmount())
	fmt.Printf("      - 최종 금액: $%.2f\n\n", restoredOrder.FinalAmount())

	// 5. 복원된 주문에 추가 이벤트 적용
	fmt.Println("5️⃣ 복원된 주문에 추가 이벤트 적용")
	err = restoredOrder.AddItem("cable-006", "USB-C Cable", 3, decimal.NewFromFloat(15.00))
	if err != nil {
		return fmt.Errorf("상품 추가 실패: %w", err)
	}
	fmt.Printf("   ✅ 새 상품 추가: USB-C Cable\n")

	err = restoredOrder.ConfirmOrder()
	if err != nil {
		return fmt.Errorf("주문 확정 실패: %w", err)
	}
	fmt.Printf("   ✅ 주문 확정 완료\n")

	fmt.Printf("   📊 최종 상태: 버전 %d, 상품 %d개, 상태 %s, 총액 $%.2f\n\n",
		restoredOrder.Version(), restoredOrder.ItemCount(), restoredOrder.Status(), restoredOrder.FinalAmount())

	// 6. 스냅샷 복사 및 검증
	fmt.Println("6️⃣ 스냅샷 복사 및 검증")
	clonedSnapshot := orderSnapshot.Clone()
	fmt.Printf("   ✅ 스냅샷 복사 완료\n")
	fmt.Printf("   🔍 원본과 복사본 비교:\n")
	fmt.Printf("      - 원본 ID: %s\n", orderSnapshot.ID())
	fmt.Printf("      - 복사본 ID: %s\n", clonedSnapshot.ID())
	fmt.Printf("      - ID 일치: %t\n", orderSnapshot.ID() == clonedSnapshot.ID())
	fmt.Printf("      - 버전 일치: %t\n\n", orderSnapshot.Version() == clonedSnapshot.Version())

	// 7. 스냅샷 유효성 검증
	fmt.Println("7️⃣ 스냅샷 유효성 검증")
	err = orderSnapshot.Validate()
	if err != nil {
		return fmt.Errorf("스냅샷 유효성 검증 실패: %w", err)
	}
	fmt.Printf("   ✅ 스냅샷 유효성 검증 통과\n")

	// 만료 여부 확인 (1시간 TTL로 테스트)
	isExpired := orderSnapshot.IsExpired(time.Hour)
	fmt.Printf("   🕐 스냅샷 만료 여부 (1시간 TTL): %t\n\n", isExpired)

	fmt.Println("🎉 기본 스냅샷 데모가 성공적으로 완료되었습니다!")
	fmt.Println("   - 스냅샷 생성 ✅")
	fmt.Println("   - 직렬화/역직렬화 ✅")
	fmt.Println("   - 주문 복원 ✅")
	fmt.Println("   - 추가 이벤트 적용 ✅")
	fmt.Println("   - 스냅샷 복사 ✅")
	fmt.Println("   - 유효성 검증 ✅")

	return nil
}

func runPerformanceDemo(ctx context.Context) error {
	fmt.Println("=== Performance Demo ===")

	// 성능 데모 실행
	perfDemo := demo.NewPerformanceDemo()
	return perfDemo.RunCompleteDemo(ctx)
}

// 개발용 헬퍼 함수들

func runQuickTest() {
	ctx := context.Background()

	// 테스트용 설정
	config := infrastructure.TestInfraConfig()
	infra, err := infrastructure.SetupInfrastructure(config)
	if err != nil {
		log.Fatalf("Failed to setup test infrastructure: %v", err)
	}
	defer infra.Cleanup()

	demoRunner := demo.NewDemoRunner(infra)

	fmt.Println("Running quick test...")
	if err := demoRunner.RunBasicSnapshotDemo(ctx); err != nil {
		log.Fatalf("Quick test failed: %v", err)
	}

	fmt.Println("✅ Quick test passed!")
}

func runBenchmark() {
	ctx := context.Background()

	// 성능 테스트용 설정
	config := infrastructure.PerformanceTestInfraConfig()
	infra, err := infrastructure.SetupInfrastructure(config)
	if err != nil {
		log.Fatalf("Failed to setup performance test infrastructure: %v", err)
	}
	defer infra.Cleanup()

	demoRunner := demo.NewDemoRunner(infra)

	fmt.Println("Running performance benchmark...")
	if err := demoRunner.RunPerformanceComparisonDemo(ctx); err != nil {
		log.Fatalf("Benchmark failed: %v", err)
	}

	fmt.Println("✅ Benchmark completed!")
}

// 환경별 설정 예제
func getConfigForEnvironment(env string) *infrastructure.InfraConfig {
	switch strings.ToLower(env) {
	case "development", "dev":
		config := infrastructure.DefaultInfraConfig()
		config.SnapshotConfig = infrastructure.ConfigPresets["development"]
		return config

	case "production", "prod":
		config := infrastructure.DefaultInfraConfig()
		config.MongoURI = os.Getenv("MONGO_URI") // 환경변수에서 읽기
		if config.MongoURI == "" {
			config.MongoURI = "mongodb://localhost:27017"
		}
		config.Database = "cqrs_snapshots_prod"
		config.SnapshotConfig = infrastructure.ConfigPresets["production"]
		return config

	case "testing", "test":
		return infrastructure.TestInfraConfig()

	default:
		return infrastructure.DefaultInfraConfig()
	}
}

// 설정 검증
func validateConfig(config *infrastructure.InfraConfig) error {
	if config.MongoURI == "" {
		return fmt.Errorf("MongoDB URI is required")
	}

	if config.Database == "" {
		return fmt.Errorf("Database name is required")
	}

	if config.SnapshotConfig == nil {
		return fmt.Errorf("Snapshot configuration is required")
	}

	if config.SnapshotConfig.EventCountThreshold <= 0 {
		return fmt.Errorf("Event count threshold must be positive")
	}

	return nil
}

// 설정 출력
func printConfig(config *infrastructure.InfraConfig) {
	fmt.Println("Configuration:")
	fmt.Printf("  MongoDB URI: %s\n", config.MongoURI)
	fmt.Printf("  Database: %s\n", config.Database)
	fmt.Printf("  Events Collection: %s\n", config.EventsCollection)
	fmt.Printf("  Snapshots Collection: %s\n", config.SnapshotsCollection)
	fmt.Printf("  Snapshot Policy: %s\n", config.SnapshotConfig.DefaultPolicy)
	fmt.Printf("  Serializer: %s\n", config.SnapshotConfig.DefaultSerializer)
	fmt.Printf("  Compression: %s\n", config.SnapshotConfig.DefaultCompression)
	fmt.Printf("  Event Threshold: %d\n", config.SnapshotConfig.EventCountThreshold)
	fmt.Printf("  Max Snapshots: %d\n", config.SnapshotConfig.MaxSnapshotsPerAggregate)
	fmt.Println()
}

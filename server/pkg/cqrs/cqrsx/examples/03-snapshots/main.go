package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"defense-allies-server/pkg/cqrs/cqrsx/examples/03-snapshots/demo"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/03-snapshots/infrastructure"
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

	// 인프라 설정
	config := infrastructure.DefaultInfraConfig()
	infra, err := infrastructure.SetupInfrastructure(config)
	if err != nil {
		log.Fatalf("Failed to setup infrastructure: %v", err)
	}
	defer infra.Cleanup()

	// 데모 시나리오 실행기 생성
	demoRunner := demo.NewDemoRunner(infra)

	// 시나리오 실행
	switch strings.ToLower(scenario) {
	case "basic":
		err = demoRunner.RunBasicSnapshotDemo(ctx)
	case "performance":
		err = demoRunner.RunPerformanceComparisonDemo(ctx)
	case "policies":
		err = demoRunner.RunSnapshotPoliciesDemo(ctx)
	case "serialization":
		err = demoRunner.RunSerializationComparisonDemo(ctx)
	case "stress":
		err = demoRunner.RunStressTestDemo(ctx)
	case "cleanup":
		err = demoRunner.RunCleanupDemo(ctx)
	case "all":
		err = demoRunner.RunAllDemos(ctx)
	default:
		fmt.Printf("Unknown scenario: %s\n", scenario)
		printDemoUsage()
		return
	}

	if err != nil {
		log.Fatalf("Demo failed: %v", err)
	}

	fmt.Println("\n🎉 Demo completed successfully!")
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

package main

import (
	"context"
	"fmt"
	"time"

	"defense-allies-server/pkg/cqrs"
	"defense-allies-server/pkg/cqrs/cqrsx"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/03-snapshots/domain"

	"github.com/shopspring/decimal"
)

// Performance demo demonstrates performance monitoring
func main() {
	fmt.Println("⚡ Snapshot Performance Demo")
	fmt.Println("============================")

	serializer := cqrsx.NewCompressedJSONSnapshotSerializer("gzip", false)
	policy := cqrsx.NewEventCountPolicy(2) // 2개 이벤트마다 스냅샷 생성
	config := cqrsx.DefaultSnapshotConfiguration()

	store := &InMemoryAdvancedSnapshotStore{
		snapshots: make(map[string][]cqrsx.SnapshotData),
	}

	manager := cqrsx.NewDefaultSnapshotManager(store, serializer, policy, config)

	var totalDuration time.Duration
	var successCount int
	var totalOperations int
	var createOperations int
	var restoreOperations int

	// Create multiple orders and snapshots
	fmt.Println("Creating and testing snapshots...")
	for i := 0; i < 10; i++ {
		start := time.Now()

		order := createSampleOrder()
		addItemToOrder(order, fmt.Sprintf("Item %d", i), decimal.NewFromInt(int64(100+i*10)), 1)
		addItemToOrder(order, fmt.Sprintf("Extra %d", i), decimal.NewFromInt(int64(50+i*5)), 2)

		// 2개 이벤트가 있으므로 스냅샷 생성 조건 확인
		if manager.ShouldCreateSnapshot(order, order.Version()) {
			err := manager.CreateSnapshot(context.Background(), order)
			duration := time.Since(start)
			totalDuration += duration
			totalOperations++
			createOperations++
			if err == nil {
				successCount++
				fmt.Printf("✅ Snapshot created for order %s (v%d) in %v\n",
					order.ID(), order.Version(), duration)
			} else {
				fmt.Printf("❌ Snapshot creation failed: %v\n", err)
			}
		} else {
			// 스냅샷이 생성되지 않은 경우도 카운트
			totalOperations++
			createOperations++
			fmt.Printf("⏭️  Snapshot not created for order %s (v%d) - policy condition not met\n",
				order.ID(), order.Version())
		}

		// Test restoration (스냅샷이 있는 경우에만)
		start = time.Now()
		_, _, err := manager.RestoreFromSnapshot(context.Background(), order.ID(), order.Version())
		duration := time.Since(start)
		totalDuration += duration
		totalOperations++
		restoreOperations++
		if err == nil {
			successCount++
			fmt.Printf("🔄 Snapshot restored for order %s in %v\n", order.ID(), duration)
		} else {
			// 스냅샷이 없는 경우는 정상적인 상황
			if manager.ShouldCreateSnapshot(order, order.Version()) {
				fmt.Printf("❌ Snapshot restoration failed: %v\n", err)
			} else {
				successCount++ // 스냅샷이 없는 것은 정상
				fmt.Printf("ℹ️  No snapshot to restore for order %s (policy condition not met)\n", order.ID())
			}
		}
	}

	// Generate performance report
	fmt.Printf("\n" + repeatString("=", 60))
	fmt.Printf("\n📊 Performance Report:\n")
	fmt.Printf(repeatString("=", 60) + "\n")
	fmt.Printf("   - Total operations: %d\n", totalOperations)
	fmt.Printf("   - Create operations: %d\n", createOperations)
	fmt.Printf("   - Restore operations: %d\n", restoreOperations)
	fmt.Printf("   - Success rate: %.2f%%\n", float64(successCount)/float64(totalOperations)*100)
	fmt.Printf("   - Average duration: %v\n", totalDuration/time.Duration(totalOperations))
	fmt.Printf("   - Total duration: %v\n", totalDuration)

	// Test different serializers performance
	fmt.Printf("\n" + repeatString("=", 60))
	fmt.Printf("\nSerializer Performance Comparison:\n")
	fmt.Printf(repeatString("=", 60) + "\n")

	testOrder := createComplexOrder()
	serializers := map[string]cqrsx.AdvancedSnapshotSerializer{
		"JSON":            cqrsx.NewJSONSnapshotSerializer(false),
		"Pretty JSON":     cqrsx.NewJSONSnapshotSerializer(true),
		"Compressed JSON": cqrsx.NewCompressedJSONSnapshotSerializer("gzip", false),
		"BSON":            cqrsx.NewBSONSnapshotSerializer(),
		"Compressed BSON": cqrsx.NewCompressedBSONSnapshotSerializer("gzip"),
	}

	for name, ser := range serializers {
		start := time.Now()
		data, err := ser.SerializeSnapshot(testOrder)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("%-20s: ERROR - %v\n", name, err)
			continue
		}

		fmt.Printf("%-20s: %d bytes in %v (%.2f MB/s)\n",
			name, len(data), duration,
			float64(len(data))/duration.Seconds()/1024/1024)
	}

	// Test policy performance
	fmt.Printf("\n" + repeatString("=", 60))
	fmt.Printf("\nPolicy Performance Comparison:\n")
	fmt.Printf(repeatString("=", 60) + "\n")

	policies := map[string]cqrsx.SnapshotPolicy{
		"EventCount":   cqrsx.NewEventCountPolicy(5),
		"VersionBased": cqrsx.NewVersionBasedPolicy(3),
		"TimeBased":    cqrsx.NewTimeBasedPolicy(1 * time.Hour),
		"Adaptive":     cqrsx.NewAdaptivePolicy(8, 0.7),
		"Always":       cqrsx.NewAlwaysPolicy(),
		"Never":        cqrsx.NewNeverPolicy(),
	}

	testIterations := 10000
	for name, pol := range policies {
		start := time.Now()
		for i := 0; i < testIterations; i++ {
			pol.ShouldCreateSnapshot(testOrder, i%10)
		}
		duration := time.Since(start)

		fmt.Printf("%-15s: %d evaluations in %v (%.0f ops/sec)\n",
			name, testIterations, duration,
			float64(testIterations)/duration.Seconds())
	}

	// Memory usage simulation
	fmt.Printf("\n" + repeatString("=", 60))
	fmt.Printf("\nMemory Usage Simulation:\n")
	fmt.Printf(repeatString("=", 60) + "\n")

	// Simulate different snapshot retention policies
	retentionPolicies := []int{1, 3, 5, 10, 20}
	for _, retention := range retentionPolicies {
		totalSize := int64(0)
		for aggregateID := range store.snapshots {
			snapshots := store.snapshots[aggregateID]
			keepCount := retention
			if len(snapshots) < keepCount {
				keepCount = len(snapshots)
			}
			for i := len(snapshots) - keepCount; i < len(snapshots); i++ {
				totalSize += snapshots[i].Size()
			}
		}
		fmt.Printf("Retention %2d snapshots: %d bytes (%.2f KB)\n",
			retention, totalSize, float64(totalSize)/1024)
	}

	fmt.Println("\n✅ Performance demo completed!")
}

// Helper functions

func createSampleOrder() *domain.Order {
	order := domain.NewOrder()
	// Create order with customer ID
	order.CreateOrder(order.ID(), "customer-123", decimal.Zero)
	return order
}

func createComplexOrder() *domain.Order {
	order := domain.NewOrder()
	order.CreateOrder(order.ID(), "customer-complex-test", decimal.NewFromInt(15))

	// Add multiple items
	order.AddItem("laptop-001", "Gaming Laptop", 1, decimal.NewFromInt(1500))
	order.AddItem("mouse-001", "Gaming Mouse", 2, decimal.NewFromInt(75))
	order.AddItem("keyboard-001", "Mechanical Keyboard", 1, decimal.NewFromInt(120))
	order.AddItem("monitor-001", "4K Monitor", 2, decimal.NewFromInt(400))
	order.AddItem("headset-001", "Gaming Headset", 1, decimal.NewFromInt(200))

	// Apply discount
	order.ApplyDiscount(decimal.NewFromFloat(0.15), "Bulk purchase discount")

	// Confirm order
	order.ConfirmOrder()

	return order
}

func addItemToOrder(order *domain.Order, name string, price decimal.Decimal, quantity int) {
	order.AddItem(fmt.Sprintf("prod-%s", name), name, quantity, price)
}

// String repeat helper
func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// InMemoryAdvancedSnapshotStore implements AdvancedSnapshotStore for demo
type InMemoryAdvancedSnapshotStore struct {
	snapshots map[string][]cqrsx.SnapshotData
}

func (s *InMemoryAdvancedSnapshotStore) SaveSnapshot(ctx context.Context, aggregate cqrs.AggregateRoot) error {
	// This would normally serialize the aggregate, but for demo we'll create a simple snapshot
	snapshot := &InMemorySnapshotData{
		aggregateID:   aggregate.ID(),
		aggregateType: aggregate.Type(),
		version:       aggregate.Version(),
		data:          []byte(fmt.Sprintf("snapshot-data-%s-v%d", aggregate.ID(), aggregate.Version())),
		timestamp:     time.Now(),
		metadata:      map[string]interface{}{"demo": true},
		size:          100,
		contentType:   "application/json",
		compression:   "gzip",
	}

	if s.snapshots[aggregate.ID()] == nil {
		s.snapshots[aggregate.ID()] = []cqrsx.SnapshotData{}
	}
	s.snapshots[aggregate.ID()] = append(s.snapshots[aggregate.ID()], snapshot)
	return nil
}

func (s *InMemoryAdvancedSnapshotStore) LoadSnapshot(ctx context.Context, aggregateID, aggregateType string) (cqrs.AggregateRoot, error) {
	snapshots := s.snapshots[aggregateID]
	if len(snapshots) == 0 {
		return nil, fmt.Errorf("snapshot not found")
	}
	// Return the latest snapshot (simplified for demo)
	order := domain.NewOrderWithID(aggregateID)
	order.CreateOrder(aggregateID, "customer-123", decimal.Zero)
	return order, nil
}

func (s *InMemoryAdvancedSnapshotStore) GetSnapshot(ctx context.Context, aggregateID string, maxVersion int) (cqrsx.SnapshotData, error) {
	snapshots := s.snapshots[aggregateID]
	for i := len(snapshots) - 1; i >= 0; i-- {
		if snapshots[i].Version() <= maxVersion {
			return snapshots[i], nil
		}
	}
	return nil, fmt.Errorf("snapshot not found")
}

func (s *InMemoryAdvancedSnapshotStore) GetSnapshotByVersion(ctx context.Context, aggregateID string, version int) (cqrsx.SnapshotData, error) {
	snapshots := s.snapshots[aggregateID]
	for _, snapshot := range snapshots {
		if snapshot.Version() == version {
			return snapshot, nil
		}
	}
	return nil, fmt.Errorf("snapshot not found")
}

func (s *InMemoryAdvancedSnapshotStore) DeleteSnapshot(ctx context.Context, aggregateID string, version int) error {
	snapshots := s.snapshots[aggregateID]
	for i, snapshot := range snapshots {
		if snapshot.Version() == version {
			s.snapshots[aggregateID] = append(snapshots[:i], snapshots[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("snapshot not found")
}

func (s *InMemoryAdvancedSnapshotStore) DeleteOldSnapshots(ctx context.Context, aggregateID string, keepCount int) error {
	snapshots := s.snapshots[aggregateID]
	if len(snapshots) <= keepCount {
		return nil
	}
	s.snapshots[aggregateID] = snapshots[len(snapshots)-keepCount:]
	return nil
}

func (s *InMemoryAdvancedSnapshotStore) ListSnapshotsForAggregate(ctx context.Context, aggregateID string) ([]cqrsx.SnapshotData, error) {
	return s.snapshots[aggregateID], nil
}

func (s *InMemoryAdvancedSnapshotStore) GetSnapshotStats(ctx context.Context) (map[string]interface{}, error) {
	totalCount := 0
	for _, snapshots := range s.snapshots {
		totalCount += len(snapshots)
	}
	return map[string]interface{}{
		"total_snapshots": totalCount,
		"aggregates":      len(s.snapshots),
		"generated_at":    time.Now(),
	}, nil
}

// InMemorySnapshotData implements SnapshotData for demo
type InMemorySnapshotData struct {
	aggregateID   string
	aggregateType string
	version       int
	data          []byte
	timestamp     time.Time
	metadata      map[string]interface{}
	size          int64
	contentType   string
	compression   string
}

func (s *InMemorySnapshotData) ID() string                       { return s.aggregateID }
func (s *InMemorySnapshotData) Type() string                     { return s.aggregateType }
func (s *InMemorySnapshotData) Version() int                     { return s.version }
func (s *InMemorySnapshotData) Data() []byte                     { return s.data }
func (s *InMemorySnapshotData) Timestamp() time.Time             { return s.timestamp }
func (s *InMemorySnapshotData) Metadata() map[string]interface{} { return s.metadata }
func (s *InMemorySnapshotData) Size() int64                      { return s.size }
func (s *InMemorySnapshotData) ContentType() string              { return s.contentType }
func (s *InMemorySnapshotData) Compression() string              { return s.compression }

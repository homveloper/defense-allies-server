package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"defense-allies-server/pkg/cqrs"
	"defense-allies-server/pkg/cqrs/cqrsx"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/03-snapshots/domain"

	"github.com/shopspring/decimal"
)

// Basic demo using new cqrsx snapshot features
func main() {
	fmt.Println("🚀 Enhanced Snapshot Basic Demo")
	fmt.Println("================================")

	// Create enhanced snapshot manager with JSON serializer and event count policy
	serializer := cqrsx.NewJSONSnapshotSerializer(true) // pretty print
	policy := cqrsx.NewEventCountPolicy(3)              // snapshot every 3 events
	config := cqrsx.DefaultSnapshotConfiguration()

	// Create in-memory store for demo
	store := &InMemoryAdvancedSnapshotStore{
		snapshots: make(map[string][]cqrsx.SnapshotData),
	}

	manager := cqrsx.NewDefaultSnapshotManager(store, serializer, policy, config)

	// Create and modify order
	order := createSampleOrder()
	fmt.Printf("📦 Created order: %s\n", order.ID())

	// Add events to trigger snapshot
	addItemToOrder(order, "Laptop", decimal.NewFromInt(1200), 1)
	addItemToOrder(order, "Mouse", decimal.NewFromInt(25), 2)
	addItemToOrder(order, "Keyboard", decimal.NewFromInt(75), 1)

	// Check if snapshot should be created
	eventCount := 3
	if manager.ShouldCreateSnapshot(order, eventCount) {
		fmt.Printf("✅ Creating snapshot (policy: %s)\n", policy.GetPolicyName())

		err := manager.CreateSnapshot(context.Background(), order)
		if err != nil {
			log.Fatalf("Failed to create snapshot: %v", err)
		}

		fmt.Printf("📸 Snapshot created successfully\n")
	}

	// Get snapshot info
	infos, err := manager.GetSnapshotInfo(context.Background(), order.ID())
	if err != nil {
		log.Fatalf("Failed to get snapshot info: %v", err)
	}

	fmt.Printf("\n📊 Snapshot Information:\n")
	for _, info := range infos {
		fmt.Printf("   - Version: %d, Size: %d bytes, Type: %s\n",
			info.Version, info.Size, info.ContentType)
		fmt.Printf("   - Created: %s\n", info.Timestamp.Format("2006-01-02 15:04:05"))
	}

	fmt.Println("\n✅ Basic demo completed!")
}

// Helper functions

func createSampleOrder() *domain.Order {
	order := domain.NewOrder()
	// Create order with customer ID
	order.CreateOrder(order.ID(), "customer-123", decimal.Zero)
	return order
}

func addItemToOrder(order *domain.Order, name string, price decimal.Decimal, quantity int) {
	order.AddItem(fmt.Sprintf("prod-%s", name), name, quantity, price)
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
		compression:   "none",
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

// snapshot_store_test.go - 스냅샷 저장소 테스트
package cqrsx

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"testing"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// // TestAggregate는 테스트용 집합체 구현입니다
// type TestAggregate struct {
// 	id      uuid.UUID
// 	version int
// 	data    map[string]interface{}
// }

// func NewTestAggregate(id uuid.UUID) *TestAggregate {
// 	return &TestAggregate{
// 		id:      id,
// 		version: 0,
// 		data:    make(map[string]interface{}),
// 	}
// }

// func (a *TestAggregate) GetID() uuid.UUID      { return a.id }
// func (a *TestAggregate) GetType() string       { return "TestAggregate" }
// func (a *TestAggregate) GetVersion() int       { return a.version }
// func (a *TestAggregate) GetState() interface{} { return a.data }

// func (a *TestAggregate) Apply(event Event) {
// 	a.version = event.Version()
// 	// 간단한 데이터 업데이트
// 	if eventData, ok := event.Data().(map[string]interface{}); ok {
// 		for k, v := range eventData {
// 			a.data[k] = v
// 		}
// 	}
// }

// func (a *TestAggregate) LoadFromSnapshot(data []byte, version int) error {
// 	var state map[string]interface{}
// 	if err := json.Unmarshal(data, &state); err != nil {
// 		return err
// 	}
// 	a.data = state
// 	a.version = version
// 	return nil
// }

// // MockEventStore는 테스트용 이벤트 저장소입니다
// type MockEventStore struct {
// 	events map[string][]Event
// }

// func NewMockEventStore() *MockEventStore {
// 	return &MockEventStore{
// 		events: make(map[string][]Event),
// 	}
// }

// func (m *MockEventStore) Save(ctx context.Context, events []Event, expectedVersion int) error {
// 	if len(events) == 0 {
// 		return nil
// 	}

// 	aggregateID := events[0].string().String()
// 	m.events[aggregateID] = append(m.events[aggregateID], events...)
// 	return nil
// }

// func (m *MockEventStore) Load(ctx context.Context, aggregateID uuid.UUID) ([]Event, error) {
// 	return m.events[aggregateID.String()], nil
// }

// func (m *MockEventStore) LoadFrom(ctx context.Context, aggregateID uuid.UUID, fromVersion int) ([]Event, error) {
// 	allEvents := m.events[aggregateID.String()]
// 	var filteredEvents []Event
// 	for _, event := range allEvents {
// 		if event.Version() >= fromVersion {
// 			filteredEvents = append(filteredEvents, event)
// 		}
// 	}
// 	return filteredEvents, nil
// }

// func (m *MockEventStore) GetMetrics() StoreMetrics {
// 	return StoreMetrics{}
// }

// func (m *MockEventStore) Close() error {
// 	return nil
// }

// // 테스트용 MongoDB 설정
// func setupTestMongoDB(t *testing.T) *mongo.Collection {
// 	// 실제 테스트에서는 testcontainers 사용 권장
// 	// 여기서는 로컬 MongoDB 가정
// 	ctx := context.Background()
// 	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
// 	require.NoError(t, err)

// 	collection := client.Database("test_cqrsx").Collection("snapshots")

// 	// 테스트 시작 전 컬렉션 정리
// 	collection.Drop(ctx)

// 	t.Cleanup(func() {
// 		collection.Drop(ctx)
// 		client.Disconnect(ctx)
// 	})

// 	return collection
// }

// func TestSnapshotManager_CreateSnapshot(t *testing.T) {
// 	// Given
// 	collection := setupTestMongoDB(t)
// 	eventStore := NewMockEventStore()
// 	manager := NewSnapshotManager(collection, eventStore, 3)

// 	aggregateID := uuid.New()
// 	aggregate := NewTestAggregate(aggregateID)
// 	aggregate.data["field1"] = "value1"
// 	aggregate.data["field2"] = 42
// 	aggregate.version = 3

// 	// When
// 	err := manager.CreateSnapshot(context.Background(), aggregate)

// 	// Then
// 	assert.NoError(t, err)

// 	// 스냅샷이 저장되었는지 확인
// 	snapshot, err := manager.GetLatestSnapshot(context.Background(), aggregateID)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, snapshot)
// 	assert.Equal(t, aggregateID.String(), snapshot.string)
// 	assert.Equal(t, 3, snapshot.Version)
// 	assert.Equal(t, "TestAggregate", snapshot.AggregateType)
// }

// func TestSnapshotManager_ShouldCreateSnapshot(t *testing.T) {
// 	// Given
// 	collection := setupTestMongoDB(t)
// 	eventStore := NewMockEventStore()
// 	frequency := 5
// 	manager := NewSnapshotManager(collection, eventStore, frequency)
// 	aggregateID := uuid.New()

// 	tests := []struct {
// 		name           string
// 		currentVersion int
// 		expectedResult bool
// 	}{
// 		{"Version 5 should create", 5, true},
// 		{"Version 10 should create", 10, true},
// 		{"Version 3 should not create", 3, false},
// 		{"Version 7 should not create", 7, false},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// When
// 			shouldCreate, err := manager.ShouldCreateSnapshot(aggregateID, tt.currentVersion)

// 			// Then
// 			assert.NoError(t, err)
// 			assert.Equal(t, tt.expectedResult, shouldCreate)
// 		})
// 	}
// }

// func TestSnapshotManager_LoadFromSnapshot(t *testing.T) {
// 	// Given
// 	collection := setupTestMongoDB(t)
// 	eventStore := NewMockEventStore()
// 	manager := NewSnapshotManager(collection, eventStore, 3)

// 	aggregateID := uuid.New()
// 	originalAggregate := NewTestAggregate(aggregateID)
// 	originalAggregate.data["field1"] = "value1"
// 	originalAggregate.data["field2"] = 42
// 	originalAggregate.version = 3

// 	// 스냅샷 생성
// 	err := manager.CreateSnapshot(context.Background(), originalAggregate)
// 	require.NoError(t, err)

// 	// When
// 	newAggregate := NewTestAggregate(aggregateID)
// 	err = manager.LoadFromSnapshot(context.Background(), aggregateID, newAggregate)

// 	// Then
// 	assert.NoError(t, err)
// 	assert.Equal(t, 3, newAggregate.version)
// 	assert.Equal(t, "value1", newAggregate.data["field1"])
// 	assert.Equal(t, 42, newAggregate.data["field2"])
// }

// func TestSnapshotManager_LoadAggregateFromSnapshot(t *testing.T) {
// 	// Given
// 	collection := setupTestMongoDB(t)
// 	eventStore := NewMockEventStore()
// 	manager := NewSnapshotManager(collection, eventStore, 3)

// 	aggregateID := uuid.New()

// 	// 스냅샷 생성 (버전 3)
// 	originalAggregate := NewTestAggregate(aggregateID)
// 	originalAggregate.data["field1"] = "snapshot_value"
// 	originalAggregate.version = 3
// 	err := manager.CreateSnapshot(context.Background(), originalAggregate)
// 	require.NoError(t, err)

// 	// 스냅샷 이후 이벤트 추가 (버전 4, 5)
// 	events := []Event{
// 		NewBaseEvent(aggregateID, "TestEvent", map[string]interface{}{"field2": "event_value"}, 4),
// 		NewBaseEvent(aggregateID, "TestEvent", map[string]interface{}{"field3": "latest_value"}, 5),
// 	}
// 	err = eventStore.Save(context.Background(), events, 3)
// 	require.NoError(t, err)

// 	// When
// 	newAggregate := NewTestAggregate(aggregateID)
// 	err = manager.LoadAggregateFromSnapshot(context.Background(), aggregateID, newAggregate)

// 	// Then
// 	assert.NoError(t, err)
// 	assert.Equal(t, 5, newAggregate.version)                       // 마지막 이벤트 버전
// 	assert.Equal(t, "snapshot_value", newAggregate.data["field1"]) // 스냅샷에서
// 	assert.Equal(t, "event_value", newAggregate.data["field2"])    // 이벤트에서
// 	assert.Equal(t, "latest_value", newAggregate.data["field3"])   // 이벤트에서
// }

// func TestSnapshotManager_WithCompression(t *testing.T) {
// 	// Given
// 	collection := setupTestMongoDB(t)
// 	eventStore := NewMockEventStore()
// 	manager := NewSnapshotManager(collection, eventStore, 1)
// 	manager.compressionType = CompressionGzip

// 	aggregateID := uuid.New()
// 	aggregate := NewTestAggregate(aggregateID)
// 	// 큰 데이터 생성 (압축 효과를 위해)
// 	largeData := make(map[string]interface{})
// 	for i := 0; i < 1000; i++ {
// 		largeData[fmt.Sprintf("field_%d", i)] = "this is a repeated value for compression testing"
// 	}
// 	aggregate.data = largeData
// 	aggregate.version = 1

// 	// When
// 	err := manager.CreateSnapshot(context.Background(), aggregate)
// 	require.NoError(t, err)

// 	// Then
// 	snapshot, err := manager.GetLatestSnapshot(context.Background(), aggregateID)
// 	assert.NoError(t, err)
// 	assert.True(t, snapshot.Compressed)

// 	// 압축된 스냅샷에서 로드 가능한지 확인
// 	newAggregate := NewTestAggregate(aggregateID)
// 	err = manager.LoadFromSnapshot(context.Background(), aggregateID, newAggregate)
// 	assert.NoError(t, err)
// 	assert.Equal(t, 1, newAggregate.version)
// 	assert.Len(t, newAggregate.data, 1000)
// }

// func TestSnapshotManager_WithEncryption(t *testing.T) {
// 	// Given
// 	collection := setupTestMongoDB(t)
// 	eventStore := NewMockEventStore()
// 	manager := NewSnapshotManager(collection, eventStore, 1)
// 	manager.encryptor = NewAESEncryptor("test-passphrase")

// 	aggregateID := uuid.New()
// 	aggregate := NewTestAggregate(aggregateID)
// 	aggregate.data["sensitive"] = "secret information"
// 	aggregate.version = 1

// 	// When
// 	err := manager.CreateSnapshot(context.Background(), aggregate)
// 	require.NoError(t, err)

// 	// Then
// 	snapshot, err := manager.GetLatestSnapshot(context.Background(), aggregateID)
// 	assert.NoError(t, err)
// 	assert.True(t, snapshot.Encrypted)

// 	// 암호화된 스냅샷에서 로드 가능한지 확인
// 	newAggregate := NewTestAggregate(aggregateID)
// 	err = manager.LoadFromSnapshot(context.Background(), aggregateID, newAggregate)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "secret information", newAggregate.data["sensitive"])
// }

// func TestSnapshotManager_ErrorCases(t *testing.T) {
// 	collection := setupTestMongoDB(t)
// 	eventStore := NewMockEventStore()
// 	manager := NewSnapshotManager(collection, eventStore, 3)

// 	t.Run("Load non-existent snapshot", func(t *testing.T) {
// 		aggregateID := uuid.New()
// 		_, err := manager.GetLatestSnapshot(context.Background(), aggregateID)
// 		assert.ErrorIs(t, err, ErrSnapshotNotFound)
// 	})

// 	t.Run("Load aggregate without snapshot", func(t *testing.T) {
// 		aggregateID := uuid.New()
// 		aggregate := NewTestAggregate(aggregateID)

// 		// 스냅샷 없이 이벤트만 있는 경우
// 		events := []Event{
// 			NewBaseEvent(aggregateID, "TestEvent", map[string]interface{}{"field1": "value1"}, 1),
// 		}
// 		err := eventStore.Save(context.Background(), events, 0)
// 		require.NoError(t, err)

// 		err = manager.LoadAggregateFromSnapshot(context.Background(), aggregateID, aggregate)
// 		assert.NoError(t, err)
// 		assert.Equal(t, 1, aggregate.version)
// 		assert.Equal(t, "value1", aggregate.data["field1"])
// 	})
// }

// func TestSnapshotManager_Performance(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping performance test in short mode")
// 	}

// 	// Given
// 	collection := setupTestMongoDB(t)
// 	eventStore := NewMockEventStore()
// 	manager := NewSnapshotManager(collection, eventStore, 1)

// 	aggregateID := uuid.New()
// 	aggregate := NewTestAggregate(aggregateID)

// 	// 큰 데이터 준비
// 	largeData := make(map[string]interface{})
// 	for i := 0; i < 10000; i++ {
// 		largeData[fmt.Sprintf("field_%d", i)] = fmt.Sprintf("value_%d", i)
// 	}
// 	aggregate.data = largeData
// 	aggregate.version = 1

// 	// When & Then
// 	start := time.Now()
// 	err := manager.CreateSnapshot(context.Background(), aggregate)
// 	createDuration := time.Since(start)

// 	assert.NoError(t, err)
// 	t.Logf("Create snapshot took: %v", createDuration)

// 	start = time.Now()
// 	newAggregate := NewTestAggregate(aggregateID)
// 	err = manager.LoadFromSnapshot(context.Background(), aggregateID, newAggregate)
// 	loadDuration := time.Since(start)

// 	assert.NoError(t, err)
// 	assert.Equal(t, 10000, len(newAggregate.data))
// 	t.Logf("Load snapshot took: %v", loadDuration)

// 	// 성능 임계값 검증 (예시)
// 	assert.Less(t, createDuration, 5*time.Second, "Create snapshot should be fast")
// 	assert.Less(t, loadDuration, 2*time.Second, "Load snapshot should be fast")
// }

// func TestSnapshotManager_Concurrency(t *testing.T) {
// 	// Given
// 	collection := setupTestMongoDB(t)
// 	eventStore := NewMockEventStore()
// 	manager := NewSnapshotManager(collection, eventStore, 1)

// 	aggregateID := uuid.New()

// 	// When - 동시에 여러 스냅샷 생성 시도
// 	const numGoroutines = 10
// 	errors := make(chan error, numGoroutines)

// 	for i := 0; i < numGoroutines; i++ {
// 		go func(version int) {
// 			aggregate := NewTestAggregate(aggregateID)
// 			aggregate.data["version"] = version
// 			aggregate.version = version

// 			err := manager.CreateSnapshot(context.Background(), aggregate)
// 			errors <- err
// 		}(i + 1)
// 	}

// 	// Then - 모든 고루틴이 완료될 때까지 대기
// 	for i := 0; i < numGoroutines; i++ {
// 		err := <-errors
// 		// 일부는 성공하고 일부는 중복 키 에러가 날 수 있음 (정상)
// 		if err != nil {
// 			t.Logf("Concurrent operation error (expected): %v", err)
// 		}
// 	}

// 	// 적어도 하나의 스냅샷은 저장되어야 함
// 	snapshot, err := manager.GetLatestSnapshot(context.Background(), aggregateID)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, snapshot)
// }

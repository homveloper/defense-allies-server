// state_store_extended_test.go - 확장 기능 테스트
package cqrsx

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupQueryableStateStore(t *testing.T) QueryableStateStore {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	collection := client.Database("test_cqrsx_query").Collection("states")
	collection.Drop(ctx)

	t.Cleanup(func() {
		collection.Drop(ctx)
		client.Disconnect(ctx)
	})

	return NewMongoStateStore(collection, client, WithIndexing(), WithMetrics())
}

func TestQueryableStateStore_Query(t *testing.T) {
	// Given
	store := setupQueryableStateStore(t)
	defer store.Close()

	ctx := context.Background()

	// 다양한 집합체 타입과 버전의 상태들 생성
	aggregateID1 := uuid.New()
	aggregateID2 := uuid.New()

	baseTime := time.Now().Add(-24 * time.Hour)

	states := []*AggregateState{
		{
			string:        aggregateID1,
			AggregateType: "Guild",
			Version:       1,
			Data:          []byte("guild data 1"),
			Timestamp:     baseTime,
			Metadata:      map[string]any{"level": 1},
		},
		{
			string:        aggregateID1,
			AggregateType: "Guild",
			Version:       2,
			Data:          []byte("guild data 2"),
			Timestamp:     baseTime.Add(1 * time.Hour),
			Metadata:      map[string]any{"level": 2},
		},
		{
			string:        aggregateID2,
			AggregateType: "User",
			Version:       1,
			Data:          []byte("user data 1"),
			Timestamp:     baseTime.Add(2 * time.Hour),
			Metadata:      map[string]any{"level": 1},
		},
	}

	for _, state := range states {
		err := store.Save(ctx, state)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		query         StateQuery
		expectedCount int
	}{
		{
			name: "Query by aggregate type",
			query: StateQuery{
				AggregateType: "Guild",
			},
			expectedCount: 2,
		},
		{
			name: "Query by aggregate IDs",
			query: StateQuery{
				AggregateIDs: []uuid.UUID{aggregateID1},
			},
			expectedCount: 2,
		},
		{
			name: "Query by version range",
			query: StateQuery{
				MinVersion: intPtr(1),
				MaxVersion: intPtr(1),
			},
			expectedCount: 2,
		},
		{
			name: "Query by time range",
			query: StateQuery{
				StartTime: &baseTime,
				EndTime:   timePtr(baseTime.Add(90 * time.Minute)),
			},
			expectedCount: 2,
		},
		{
			name: "Query with limit",
			query: StateQuery{
				Limit: 1,
			},
			expectedCount: 1,
		},
		{
			name: "Complex query",
			query: StateQuery{
				AggregateType: "Guild",
				MinVersion:    intPtr(2),
			},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			results, err := store.Query(ctx, tt.query)

			// Then
			assert.NoError(t, err)
			assert.Len(t, results, tt.expectedCount)
		})
	}
}

func TestQueryableStateStore_CountByQuery(t *testing.T) {
	// Given
	store := setupQueryableStateStore(t)
	defer store.Close()

	ctx := context.Background()
	aggregateID := uuid.New()

	// 5개 버전 저장
	for i := 1; i <= 5; i++ {
		state := createTestState(aggregateID, i)
		err := store.Save(ctx, state)
		require.NoError(t, err)
	}

	// When
	count, err := store.CountByQuery(ctx, StateQuery{
		AggregateIDs: []uuid.UUID{aggregateID},
		MinVersion:   intPtr(3),
	})

	// Then
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count) // 버전 3, 4, 5
}

func TestQueryableStateStore_GetAggregateTypes(t *testing.T) {
	// Given
	store := setupQueryableStateStore(t)
	defer store.Close()

	ctx := context.Background()

	// 다양한 타입의 집합체 저장
	types := []string{"Guild", "User", "Order"}
	for _, aggregateType := range types {
		state := NewAggregateState(uuid.New(), aggregateType, 1, []byte("test"))
		err := store.Save(ctx, state)
		require.NoError(t, err)
	}

	// When
	result, err := store.GetAggregateTypes(ctx)

	// Then
	assert.NoError(t, err)
	assert.ElementsMatch(t, types, result)
}

func TestQueryableStateStore_GetVersions(t *testing.T) {
	// Given
	store := setupQueryableStateStore(t)
	defer store.Close()

	ctx := context.Background()
	aggregateID := uuid.New()

	// 버전 1, 3, 5 저장 (연속되지 않음)
	versions := []int{1, 3, 5}
	for _, version := range versions {
		state := createTestState(aggregateID, version)
		err := store.Save(ctx, state)
		require.NoError(t, err)
	}

	// When
	result, err := store.GetVersions(ctx, aggregateID)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, versions, result)
}

func TestMetricsStateStore_GetMetrics(t *testing.T) {
	// Given
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)
	defer client.Disconnect(ctx)

	collection := client.Database("test_cqrsx_metrics").Collection("states")
	collection.Drop(ctx)

	store := NewMongoStateStore(collection, client, WithMetrics())
	defer store.Close()

	metricsStore := store

	// 테스트 데이터 저장
	for i := 1; i <= 10; i++ {
		for j := 1; j <= 3; j++ {
			aggregateID := uuid.New()
			state := createTestState(aggregateID, j)
			state.Data = make([]byte, 100*i) // 크기 다양화
			err := store.Save(ctx, state)
			require.NoError(t, err)
		}
	}

	// When
	metrics, err := metricsStore.GetMetrics(ctx)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, int64(30), metrics.TotalStates) // 10 aggregates * 3 versions
	assert.Greater(t, metrics.TotalStorageBytes, int64(0))
	assert.Greater(t, metrics.AverageSize, int64(0))
	assert.Greater(t, metrics.MaxSize, int64(0))
	assert.Contains(t, metrics.AggregateTypes, "TestAggregate")
}

func TestMetricsStateStore_GetAggregateMetrics(t *testing.T) {
	// Given
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)
	defer client.Disconnect(ctx)

	collection := client.Database("test_cqrsx_aggregate_metrics").Collection("states")
	collection.Drop(ctx)

	store := NewMongoStateStore(collection, client, WithMetrics())
	defer store.Close()

	metricsStore := store

	aggregateID := uuid.New()

	// 특정 집합체의 여러 버전 저장
	for i := 1; i <= 5; i++ {
		state := createTestState(aggregateID, i)
		state.Data = make([]byte, 100*i)
		err := store.Save(ctx, state)
		require.NoError(t, err)
	}

	// When
	metrics, err := metricsStore.GetAggregateMetrics(ctx, aggregateID)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, int64(5), metrics.TotalStates)
	assert.Greater(t, metrics.TotalStorageBytes, int64(0))
	assert.Equal(t, []string{"TestAggregate"}, metrics.AggregateTypes)
}

func TestRetentionPolicy_KeepLastN(t *testing.T) {
	// Given
	policy := NewKeepLastNPolicy(3)
	now := time.Now()

	states := []*AggregateState{
		{Version: 1, Timestamp: now.Add(-4 * time.Hour)},
		{Version: 2, Timestamp: now.Add(-3 * time.Hour)},
		{Version: 3, Timestamp: now.Add(-2 * time.Hour)},
		{Version: 4, Timestamp: now.Add(-1 * time.Hour)},
		{Version: 5, Timestamp: now},
	}

	// When
	candidates := policy.GetCleanupCandidates(context.Background(), states)

	// Then
	assert.Len(t, candidates, 2) // 가장 오래된 2개 (버전 1, 2)
	assert.Equal(t, 1, candidates[0].Version)
	assert.Equal(t, 2, candidates[1].Version)
}

func TestRetentionPolicy_TimeBased(t *testing.T) {
	// Given
	policy := NewTimeBasedPolicy(2 * time.Hour)
	now := time.Now()

	states := []*AggregateState{
		{Version: 1, Timestamp: now.Add(-3 * time.Hour)}, // 너무 오래됨
		{Version: 2, Timestamp: now.Add(-1 * time.Hour)}, // 보존
		{Version: 3, Timestamp: now},                     // 보존
	}

	// When
	candidates := policy.GetCleanupCandidates(context.Background(), states)

	// Then
	assert.Len(t, candidates, 1)
	assert.Equal(t, 1, candidates[0].Version)
}

func TestRetentionPolicy_SizeBased(t *testing.T) {
	// Given
	policy := NewSizeBasedPolicy(250) // 250바이트 제한

	states := []*AggregateState{
		{Version: 1, Data: make([]byte, 100)}, // 100바이트
		{Version: 2, Data: make([]byte, 100)}, // 100바이트
		{Version: 3, Data: make([]byte, 100)}, // 100바이트 (총 300바이트)
	}

	// When
	candidates := policy.GetCleanupCandidates(context.Background(), states)

	// Then
	assert.Greater(t, len(candidates), 0) // 크기 제한을 초과하므로 일부 정리됨
}

func TestRetentionPolicy_Composite(t *testing.T) {
	// Given - KeepLast(2) AND TimeBased(1시간)
	keepLastPolicy := NewKeepLastNPolicy(2)
	timePolicy := NewTimeBasedPolicy(1 * time.Hour)
	policy := NewCompositePolicy(CompositeModeAND, keepLastPolicy, timePolicy)

	now := time.Now()
	states := []*AggregateState{
		{Version: 1, Timestamp: now.Add(-2 * time.Hour)},    // 시간 초과, 개수 초과
		{Version: 2, Timestamp: now.Add(-30 * time.Minute)}, // 시간 OK, 개수 OK
		{Version: 3, Timestamp: now},                        // 시간 OK, 개수 OK
	}

	// When
	candidates := policy.GetCleanupCandidates(context.Background(), states)

	// Then
	assert.Len(t, candidates, 1)
	assert.Equal(t, 1, candidates[0].Version)
}

// 헬퍼 함수들
func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func BenchmarkStateStore_Operations(b *testing.B) {
	ctx := context.Background()
	client, _ := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	defer client.Disconnect(ctx)

	collection := client.Database("bench_cqrsx").Collection("states")
	collection.Drop(ctx)

	store := NewMongoStateStore(collection, client)
	defer store.Close()

	aggregateID := uuid.New()

	b.Run("Save", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			state := createTestState(aggregateID, i+1)
			store.Save(ctx, state)
		}
	})

	b.Run("Load", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			store.Load(ctx, aggregateID)
		}
	})

	b.Run("LoadVersion", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			version := (i % 100) + 1
			store.LoadVersion(ctx, aggregateID, version)
		}
	})
}

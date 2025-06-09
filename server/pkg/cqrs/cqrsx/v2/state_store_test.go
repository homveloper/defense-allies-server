// state_store_test.go - 상태 저장소 TDD 테스트
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

// 테스트용 MongoDB 설정
func setupTestStateStore(t *testing.T) StateStore {
	// 실제 테스트에서는 testcontainers 사용 권장
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	collection := client.Database("test_cqrsx_state").Collection("states")

	// 테스트 시작 전 컬렉션 정리
	collection.Drop(ctx)

	t.Cleanup(func() {
		collection.Drop(ctx)
		client.Disconnect(ctx)
	})

	return NewMongoStateStore(collection, client, WithIndexing(), WithMetrics())
}

func createTestState(aggregateID uuid.UUID, version int) *AggregateState {
	data := []byte(`{"field1": "value1", "field2": 42}`)
	state := NewAggregateState(aggregateID, "TestAggregate", version, data)
	state.SetMetadata("test", "metadata")
	return state
}

func TestStateStore_Save(t *testing.T) {
	// Given
	store := setupTestStateStore(t)
	defer store.Close()

	aggregateID := uuid.New()
	state := createTestState(aggregateID, 1)

	// When
	err := store.Save(context.Background(), state)

	// Then
	assert.NoError(t, err)

	// 저장된 상태를 확인
	loaded, err := store.Load(context.Background(), aggregateID)
	assert.NoError(t, err)
	assert.Equal(t, state.string, loaded.string)
	assert.Equal(t, state.AggregateType, loaded.AggregateType)
	assert.Equal(t, state.Version, loaded.Version)
	assert.Equal(t, state.Data, loaded.Data)
}

func TestStateStore_Load(t *testing.T) {
	// Given
	store := setupTestStateStore(t)
	defer store.Close()

	aggregateID := uuid.New()
	state := createTestState(aggregateID, 1)

	err := store.Save(context.Background(), state)
	require.NoError(t, err)

	// When
	loaded, err := store.Load(context.Background(), aggregateID)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, state.string, loaded.string)
	assert.Equal(t, state.Version, loaded.Version)
	assert.Equal(t, string(state.Data), string(loaded.Data))
}

func TestStateStore_LoadNonExistent(t *testing.T) {
	// Given
	store := setupTestStateStore(t)
	defer store.Close()

	nonExistentID := uuid.New()

	// When
	_, err := store.Load(context.Background(), nonExistentID)

	// Then
	assert.ErrorIs(t, err, ErrStateNotFound)
}

func TestStateStore_LoadVersion(t *testing.T) {
	// Given
	store := setupTestStateStore(t)
	defer store.Close()

	aggregateID := uuid.New()

	// 여러 버전 저장
	state1 := createTestState(aggregateID, 1)
	state2 := createTestState(aggregateID, 2)
	state3 := createTestState(aggregateID, 3)

	err := store.Save(context.Background(), state1)
	require.NoError(t, err)
	err = store.Save(context.Background(), state2)
	require.NoError(t, err)
	err = store.Save(context.Background(), state3)
	require.NoError(t, err)

	// When
	loaded2, err := store.LoadVersion(context.Background(), aggregateID, 2)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, 2, loaded2.Version)

	// 최신 버전은 3이어야 함
	latest, err := store.Load(context.Background(), aggregateID)
	assert.NoError(t, err)
	assert.Equal(t, 3, latest.Version)
}

func TestStateStore_Delete(t *testing.T) {
	// Given
	store := setupTestStateStore(t)
	defer store.Close()

	aggregateID := uuid.New()
	state := createTestState(aggregateID, 1)

	err := store.Save(context.Background(), state)
	require.NoError(t, err)

	// When
	err = store.Delete(context.Background(), aggregateID, 1)

	// Then
	assert.NoError(t, err)

	// 삭제된 상태는 조회되지 않아야 함
	_, err = store.LoadVersion(context.Background(), aggregateID, 1)
	assert.ErrorIs(t, err, ErrStateNotFound)
}

func TestStateStore_List(t *testing.T) {
	// Given
	store := setupTestStateStore(t)
	defer store.Close()

	aggregateID := uuid.New()

	// 여러 버전 저장
	for i := 1; i <= 5; i++ {
		state := createTestState(aggregateID, i)
		err := store.Save(context.Background(), state)
		require.NoError(t, err)
	}

	// When
	states, err := store.List(context.Background(), aggregateID)

	// Then
	assert.NoError(t, err)
	assert.Len(t, states, 5)

	// 버전 순으로 정렬되어 있는지 확인
	for i, state := range states {
		assert.Equal(t, i+1, state.Version)
	}
}

func TestStateStore_Count(t *testing.T) {
	// Given
	store := setupTestStateStore(t)
	defer store.Close()

	aggregateID := uuid.New()

	// 3개 버전 저장
	for i := 1; i <= 3; i++ {
		state := createTestState(aggregateID, i)
		err := store.Save(context.Background(), state)
		require.NoError(t, err)
	}

	// When
	count, err := store.Count(context.Background(), aggregateID)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestStateStore_Exists(t *testing.T) {
	// Given
	store := setupTestStateStore(t)
	defer store.Close()

	aggregateID := uuid.New()
	state := createTestState(aggregateID, 1)

	err := store.Save(context.Background(), state)
	require.NoError(t, err)

	// When & Then
	exists, err := store.Exists(context.Background(), aggregateID, 1)
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = store.Exists(context.Background(), aggregateID, 999)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestStateStore_WithCompression(t *testing.T) {
	// Given
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)
	defer client.Disconnect(ctx)

	collection := client.Database("test_cqrsx_state").Collection("states_compression")
	collection.Drop(ctx)

	store := NewMongoStateStore(collection, client, WithCompression(CompressionGzip))
	defer store.Close()

	aggregateID := uuid.New()

	// 큰 데이터 생성 (압축 효과를 위해)
	largeData := make([]byte, 5000)
	for i := range largeData {
		largeData[i] = byte('A' + (i % 26))
	}

	state := NewAggregateState(aggregateID, "TestAggregate", 1, largeData)

	// When
	err = store.Save(ctx, state)
	require.NoError(t, err)

	// Then
	loaded, err := store.Load(ctx, aggregateID)
	assert.NoError(t, err)
	assert.Equal(t, largeData, loaded.Data)

	// 압축으로 인한 저장 공간 절약 확인
	mongoStore := store
	internalMetrics := mongoStore.GetInternalMetrics()
	assert.Greater(t, internalMetrics.compressionSaved, int64(0))
}

func TestStateStore_WithEncryption(t *testing.T) {
	// Given
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)
	defer client.Disconnect(ctx)

	collection := client.Database("test_cqrsx_state").Collection("states_encryption")
	collection.Drop(ctx)

	encryptor := NewAESEncryptor("test-secret-key")
	store := NewMongoStateStore(collection, client, WithEncryption(encryptor))
	defer store.Close()

	aggregateID := uuid.New()
	sensitiveData := []byte(`{"secret": "very sensitive information"}`)
	state := NewAggregateState(aggregateID, "TestAggregate", 1, sensitiveData)

	// When
	err = store.Save(ctx, state)
	require.NoError(t, err)

	// Then
	loaded, err := store.Load(ctx, aggregateID)
	assert.NoError(t, err)
	assert.Equal(t, sensitiveData, loaded.Data)

	// 암호화 적용 확인
	mongoStore := store
	internalMetrics := mongoStore.GetInternalMetrics()
	assert.Greater(t, internalMetrics.encryptionApplied, int64(0))
}

func TestStateStore_WithRetentionPolicy(t *testing.T) {
	// Given
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)
	defer client.Disconnect(ctx)

	collection := client.Database("test_cqrsx_state").Collection("states_retention")
	collection.Drop(ctx)

	policy := KeepLast(3) // 최신 3개만 보존
	store := NewMongoStateStore(collection, client, WithRetentionPolicy(policy))
	defer store.Close()

	aggregateID := uuid.New()

	// When - 5개 버전 저장
	for i := 1; i <= 5; i++ {
		state := createTestState(aggregateID, i)
		err := store.Save(ctx, state)
		require.NoError(t, err)

		// 보존 정책이 비동기로 실행되므로 잠시 대기
		time.Sleep(100 * time.Millisecond)
	}

	// Then - 보존 정책 수동 적용 후 확인
	mongoStore := store
	mongoStore.ApplyRetentionPolicy(ctx, aggregateID)

	count, err := store.Count(ctx, aggregateID)
	assert.NoError(t, err)
	assert.LessOrEqual(t, count, int64(3)) // 최대 3개만 남아야 함
}

func TestAggregateState_Validation(t *testing.T) {
	tests := []struct {
		name      string
		state     *AggregateState
		expectErr bool
	}{
		{
			name: "Valid state",
			state: &AggregateState{
				string:        uuid.New(),
				AggregateType: "TestAggregate",
				Version:       1,
				Data:          []byte("test"),
				Timestamp:     time.Now(),
				Metadata:      make(map[string]any),
			},
			expectErr: false,
		},
		{
			name: "Invalid aggregate ID",
			state: &AggregateState{
				string:        uuid.Nil,
				AggregateType: "TestAggregate",
				Version:       1,
				Data:          []byte("test"),
				Timestamp:     time.Now(),
			},
			expectErr: true,
		},
		{
			name: "Empty aggregate type",
			state: &AggregateState{
				string:        uuid.New(),
				AggregateType: "",
				Version:       1,
				Data:          []byte("test"),
				Timestamp:     time.Now(),
			},
			expectErr: true,
		},
		{
			name: "Negative version",
			state: &AggregateState{
				string:        uuid.New(),
				AggregateType: "TestAggregate",
				Version:       -1,
				Data:          []byte("test"),
				Timestamp:     time.Now(),
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.state.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStateStore_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Given
	store := setupTestStateStore(t)
	defer store.Close()

	aggregateID := uuid.New()

	// When - 1000개 상태 저장
	start := time.Now()
	for i := 1; i <= 1000; i++ {
		state := createTestState(aggregateID, i)
		err := store.Save(context.Background(), state)
		require.NoError(t, err)
	}
	saveDuration := time.Since(start)

	// When - 1000개 상태 조회
	start = time.Now()
	states, err := store.List(context.Background(), aggregateID)
	loadDuration := time.Since(start)

	// Then
	require.NoError(t, err)
	assert.Len(t, states, 1000)

	t.Logf("Save 1000 states took: %v", saveDuration)
	t.Logf("Load 1000 states took: %v", loadDuration)

	// 성능 임계값 검증
	assert.Less(t, saveDuration, 10*time.Second, "Save should be reasonably fast")
	assert.Less(t, loadDuration, 2*time.Second, "Load should be fast")
}

func TestStateStore_Concurrency(t *testing.T) {
	// Given
	store := setupTestStateStore(t)
	defer store.Close()

	aggregateID := uuid.New()

	// When - 동시에 여러 버전 저장
	const numGoroutines = 10
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(version int) {
			state := createTestState(aggregateID, version)
			err := store.Save(context.Background(), state)
			errors <- err
		}(i + 1)
	}

	// Then - 모든 고루틴 완료 대기
	for i := 0; i < numGoroutines; i++ {
		err := <-errors
		assert.NoError(t, err)
	}

	// 저장된 상태 개수 확인
	count, err := store.Count(context.Background(), aggregateID)
	assert.NoError(t, err)
	assert.Equal(t, int64(numGoroutines), count)
}

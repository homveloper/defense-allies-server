// mongo_state_store_integration_test.go - 새로운 구조체를 사용하는 통합 테스트
package cqrsx

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 테스트용 MongoDB 설정
func setupIntegrationTestStore(t *testing.T, storeOptions ...StateStoreOption) StateStore {
	ctx := context.Background()
	// 연결 과정에 디버그 로그 추가
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017").
		SetDirect(true). // 직접 연결 시도
		SetTimeout(10 * time.Second).
		SetServerSelectionTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		// 에러 처리
		t.Fatal(fmt.Sprintf("MongoDB 연결 실패: %v", err))
	} else {
		err = client.Ping(ctx, nil)
		if err != nil {
			t.Fatal(fmt.Sprintf("MongoDB Ping 실패: %v", err))
		} else {
		}
	}

	collection := client.Database("test_cqrsx_integration").Collection("states")
	collection.Drop(ctx)

	t.Cleanup(func() {
		collection.Drop(ctx)
		client.Disconnect(ctx)
	})

	defaultOptions := []StateStoreOption{WithIndexing(), WithMetrics()}
	allOptions := append(defaultOptions, storeOptions...)

	return NewMongoStateStore(collection, client, allOptions...)
}

func createLargeTestState(aggregateID uuid.UUID, version int) *AggregateState {
	// 큰 데이터 생성 (압축 효과를 보기 위해)
	largeData := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		largeData[fmt.Sprintf("field_%d", i)] = "This is a repeated value for compression testing that should compress well with gzip or lz4 algorithms"
	}

	serializedData, _ := json.Marshal(largeData)
	state := NewAggregateState(aggregateID, "LargeGuild", version, serializedData)
	state.SetMetadata("testType", "large")
	state.SetMetadata("fieldCount", 1000)
	return state
}

// 테스트용 상태 생성 헬퍼 (integration test용)
type DummyData struct {
	Name        string   `json:"name"`
	Version     int      `json:"version"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

func createDummyData() DummyData {
	return DummyData{
		Name:        "Test Entity",
		Version:     1,
		Description: "Test data for integration operations",
		Tags:        []string{"test", "integration"},
	}
}

func createIntegrationTestState(aggregateID uuid.UUID, version int, data DummyData) *AggregateState {
	serializedData, _ := json.Marshal(data)

	state := NewAggregateState(aggregateID, "TestEntity", version, serializedData)
	state.SetMetadata("testCase", "integration")
	state.SetMetadata("created", time.Now().Format(time.RFC3339))

	return state
}

func aggregateStateToDeserializedData(state *AggregateState) DummyData {
	var deserializedData DummyData
	err := json.Unmarshal(state.Data, &deserializedData)
	if err != nil {
		panic(err)
	}
	return deserializedData
}

func TestIntegration_BasicOperations(t *testing.T) {
	// Given
	store := setupIntegrationTestStore(t)
	defer store.Close()

	ctx := context.Background()
	aggregateID := uuid.New()

	// When & Then - 저장
	data := createDummyData()
	state := createIntegrationTestState(aggregateID, 1, data)
	err := store.Save(ctx, state)
	assert.NoError(t, err)

	// When & Then - 로드
	loadedState, err := store.Load(ctx, aggregateID)
	assert.NoError(t, err)
	assert.Equal(t, state.AggregateID, loadedState.AggregateID)
	assert.Equal(t, state.Version, loadedState.Version)

	deserializedData := aggregateStateToDeserializedData(loadedState)
	assert.Equal(t, data, deserializedData)

	// When & Then - 존재 확인
	exists, err := store.Exists(ctx, aggregateID, 1)
	assert.NoError(t, err)
	assert.True(t, exists)

	// When & Then - 카운트
	count, err := store.Count(ctx, aggregateID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestIntegration_CompressionOnly(t *testing.T) {
	// Given - GZIP 압축만 적용
	store := setupIntegrationTestStore(t, WithCompression(CompressionGzip))
	defer store.Close()

	ctx := context.Background()
	aggregateID := uuid.New()

	// When
	state := createLargeTestState(aggregateID, 1)
	originalSize := len(state.Data)

	err := store.Save(ctx, state)
	require.NoError(t, err)

	// Then
	loadedState, err := store.Load(ctx, aggregateID)
	require.NoError(t, err)
	assert.Equal(t, state.Data, loadedState.Data)

	// 압축 효과 확인
	mongoStore := store.(*MongoStateStore)
	internalMetrics := mongoStore.GetInternalMetrics()
	assert.Greater(t, internalMetrics.compressionSaved, int64(0))

	t.Logf("Original size: %d bytes, Compression saved: %d bytes",
		originalSize, internalMetrics.compressionSaved)
}

func TestIntegration_EncryptionOnly(t *testing.T) {
	// Given - AES 암호화만 적용
	encryptor := NewAESEncryptor("integration-test-key")
	store := setupIntegrationTestStore(t, WithEncryption(encryptor))
	defer store.Close()

	ctx := context.Background()
	aggregateID := uuid.New()

	// When
	sensitiveData := map[string]interface{}{
		"password":     "secret123",
		"apiKey":       "sk-1234567890abcdef",
		"personalInfo": "very sensitive information",
	}
	serializedData, _ := json.Marshal(sensitiveData)
	state := NewAggregateState(aggregateID, "SensitiveUser", 1, serializedData)

	err := store.Save(ctx, state)
	require.NoError(t, err)

	// Then
	loadedState, err := store.Load(ctx, aggregateID)
	require.NoError(t, err)
	assert.Equal(t, state.Data, loadedState.Data)

	// 암호화 적용 확인
	mongoStore := store.(*MongoStateStore)
	internalMetrics := mongoStore.GetInternalMetrics()
	assert.Greater(t, internalMetrics.encryptionApplied, int64(0))
}

func TestIntegration_CompressionAndEncryption(t *testing.T) {
	// Given - 압축 + 암호화 모두 적용
	encryptor := NewAESEncryptor("compression-encryption-key")
	store := setupIntegrationTestStore(t,
		WithCompression(CompressionLZ4),
		WithEncryption(encryptor),
	)
	defer store.Close()

	ctx := context.Background()
	aggregateID := uuid.New()

	// When
	state := createLargeTestState(aggregateID, 1)
	originalSize := len(state.Data)

	err := store.Save(ctx, state)
	require.NoError(t, err)

	// Then
	loadedState, err := store.Load(ctx, aggregateID)
	require.NoError(t, err)
	assert.Equal(t, state.Data, loadedState.Data)

	// 압축과 암호화가 모두 적용되었는지 확인
	mongoStore := store.(*MongoStateStore)
	internalMetrics := mongoStore.GetInternalMetrics()
	assert.Greater(t, internalMetrics.compressionSaved, int64(0))
	assert.Greater(t, internalMetrics.encryptionApplied, int64(0))

	t.Logf("Original size: %d bytes, Compression saved: %d bytes, Encryption applied: %d times",
		originalSize, internalMetrics.compressionSaved, internalMetrics.encryptionApplied)
}

func TestIntegration_RetentionPolicy(t *testing.T) {
	// Given - 최신 3개만 보존하는 정책
	policy := KeepLast(3)
	store := setupIntegrationTestStore(t, WithRetentionPolicy(policy))
	defer store.Close()

	ctx := context.Background()
	aggregateID := uuid.New()

	// When - 5개 버전 저장
	for i := 1; i <= 5; i++ {
		data := createDummyData()
		state := createIntegrationTestState(aggregateID, i, data)
		err := store.Save(ctx, state)
		require.NoError(t, err)

		// 보존 정책이 비동기로 실행되므로 잠시 대기
		time.Sleep(10 * time.Millisecond)
	}

	// 보존 정책 수동 적용 (테스트에서 확실히 하기 위해)
	mongoStore := store.(*MongoStateStore)
	mongoStore.ApplyRetentionPolicy(ctx, aggregateID)

	// Then
	count, err := store.Count(ctx, aggregateID)
	assert.NoError(t, err)
	assert.LessOrEqual(t, count, int64(3)) // 최대 3개만 남아야 함

	// 최신 버전들이 남아있는지 확인
	exists5, err := store.Exists(ctx, aggregateID, 5)
	assert.NoError(t, err)
	assert.True(t, exists5)

	exists4, err := store.Exists(ctx, aggregateID, 4)
	assert.NoError(t, err)
	assert.True(t, exists4)

	exists3, err := store.Exists(ctx, aggregateID, 3)
	assert.NoError(t, err)
	assert.True(t, exists3)
}

func TestIntegration_MultipleVersions(t *testing.T) {
	// Given
	store := setupIntegrationTestStore(t)
	defer store.Close()

	ctx := context.Background()
	aggregateID := uuid.New()

	// When - 여러 버전 저장
	versions := []int{1, 3, 5, 7, 10}
	for _, version := range versions {
		data := createDummyData()
		state := createIntegrationTestState(aggregateID, version, data)
		state.SetMetadata("versionInfo", fmt.Sprintf("Version %d data", version))
		err := store.Save(ctx, state)
		require.NoError(t, err)
	}

	// Then - 모든 버전 조회
	states, err := store.List(ctx, aggregateID)
	assert.NoError(t, err)
	assert.Len(t, states, len(versions))

	// 버전 순서대로 정렬되어 있는지 확인
	for i, state := range states {
		assert.Equal(t, versions[i], state.Version)
	}

	// 특정 버전 로드
	state5, err := store.LoadVersion(ctx, aggregateID, 5)
	assert.NoError(t, err)
	assert.Equal(t, 5, state5.Version)

	versionInfo, exists := state5.GetMetadata("versionInfo")
	assert.True(t, exists)
	assert.Equal(t, "Version 5 data", versionInfo)

	// 최신 버전 로드 (10이어야 함)
	latest, err := store.Load(ctx, aggregateID)
	assert.NoError(t, err)
	assert.Equal(t, 10, latest.Version)
}

func TestIntegration_QueryableStore(t *testing.T) {
	// Given
	store := setupIntegrationTestStore(t)
	defer store.Close()

	queryStore, ok := store.(QueryableStateStore)
	require.True(t, ok)

	ctx := context.Background()

	// 다양한 집합체 타입과 버전 생성
	guildID1 := uuid.New()
	guildID2 := uuid.New()
	userID1 := uuid.New()

	baseTime := time.Now().Add(-2 * time.Hour)

	testCases := []struct {
		aggregateID   uuid.UUID
		aggregateType string
		version       int
		timestamp     time.Time
	}{
		{guildID1, "Guild", 1, baseTime},
		{guildID1, "Guild", 2, baseTime.Add(30 * time.Minute)},
		{guildID2, "Guild", 1, baseTime.Add(1 * time.Hour)},
		{userID1, "User", 1, baseTime.Add(90 * time.Minute)},
	}

	for _, tc := range testCases {
		state := NewAggregateState(tc.aggregateID, tc.aggregateType, tc.version, []byte("test data"))
		state.Timestamp = tc.timestamp
		err := store.Save(ctx, state)
		require.NoError(t, err)
	}

	// When & Then - 집합체 타입별 쿼리
	guildStates, err := queryStore.Query(ctx, StateQuery{
		AggregateType: "Guild",
	})
	assert.NoError(t, err)
	assert.Len(t, guildStates, 3)

	// When & Then - 특정 집합체 쿼리
	guild1States, err := queryStore.Query(ctx, StateQuery{
		AggregateIDs: []uuid.UUID{guildID1},
	})
	assert.NoError(t, err)
	assert.Len(t, guild1States, 2)

	// When & Then - 시간 범위 쿼리
	startTime := baseTime.Add(45 * time.Minute)
	endTime := baseTime.Add(2 * time.Hour)
	timeRangeStates, err := queryStore.Query(ctx, StateQuery{
		StartTime: &startTime,
		EndTime:   &endTime,
	})
	assert.NoError(t, err)
	assert.Len(t, timeRangeStates, 2) // Guild2와 User1

	// When & Then - 집합체 타입 목록 조회
	types, err := queryStore.GetAggregateTypes(ctx)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"Guild", "User"}, types)

	// When & Then - 버전 목록 조회
	versions, err := queryStore.GetVersions(ctx, guildID1)
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2}, versions)
}

func TestIntegration_MetricsStore(t *testing.T) {
	// Given
	store := setupIntegrationTestStore(t, WithCompression(CompressionGzip))
	defer store.Close()

	metricsStore, ok := store.(MetricsStateStore)
	require.True(t, ok)

	ctx := context.Background()

	// 다양한 크기의 데이터 저장
	for i := 1; i <= 10; i++ {
		aggregateID := uuid.New()
		state := createLargeTestState(aggregateID, 1)
		err := store.Save(ctx, state)
		require.NoError(t, err)
	}

	// When
	metrics, err := metricsStore.GetMetrics(ctx)

	// Then
	require.NoError(t, err)
	assert.Equal(t, int64(10), metrics.TotalStates)
	assert.Greater(t, metrics.TotalStorageBytes, int64(0))
	assert.Greater(t, metrics.AverageSize, int64(0))
	assert.Contains(t, metrics.AggregateTypes, "LargeGuild")

	// When - 특정 집합체 메트릭
	aggregateID := uuid.New()
	state := createLargeTestState(aggregateID, 1)
	err = store.Save(ctx, state)
	require.NoError(t, err)

	aggregateMetrics, err := metricsStore.GetAggregateMetrics(ctx, aggregateID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), aggregateMetrics.TotalStates)
	assert.Equal(t, []string{"LargeGuild"}, aggregateMetrics.AggregateTypes)
}

func TestIntegration_ErrorRecovery(t *testing.T) {
	// Given
	store := setupIntegrationTestStore(t)
	defer store.Close()

	ctx := context.Background()

	// When & Then - 존재하지 않는 상태 로드
	nonExistentID := uuid.New()
	_, err := store.Load(ctx, nonExistentID)
	assert.ErrorIs(t, err, ErrStateNotFound)

	// When & Then - 존재하지 않는 버전 로드
	aggregateID := uuid.New()
	state := createIntegrationTestState(aggregateID, 1, createDummyData())
	err = store.Save(ctx, state)
	require.NoError(t, err)

	_, err = store.LoadVersion(ctx, aggregateID, 999)
	assert.ErrorIs(t, err, ErrStateNotFound)

	// When & Then - 존재하지 않는 상태 삭제
	err = store.Delete(ctx, nonExistentID, 1)
	assert.ErrorIs(t, err, ErrStateNotFound)
}

func TestIntegration_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Given
	store := setupIntegrationTestStore(t, WithCompression(CompressionLZ4))
	defer store.Close()

	ctx := context.Background()
	aggregateID := uuid.New()

	// When - 대량 저장 성능 테스트
	const numStates = 100
	start := time.Now()
	for i := 1; i <= numStates; i++ {
		state := createLargeTestState(aggregateID, i)
		err := store.Save(ctx, state)
		require.NoError(t, err)
	}
	saveDuration := time.Since(start)

	// When - 대량 조회 성능 테스트
	start = time.Now()
	states, err := store.List(ctx, aggregateID)
	loadDuration := time.Since(start)

	// Then
	require.NoError(t, err)
	assert.Len(t, states, numStates)

	saveOpsPerSec := float64(numStates) / saveDuration.Seconds()

	t.Logf("Performance results:")
	t.Logf("- Save %d large states: %v (%.2f ops/sec)", numStates, saveDuration, saveOpsPerSec)
	t.Logf("- Load %d states: %v", numStates, loadDuration)

	// 성능 임계값 확인
	assert.Greater(t, saveOpsPerSec, 10.0, "Should handle at least 10 saves/sec")
	assert.Less(t, loadDuration, 2*time.Second, "Should load states quickly")

	// 메트릭 확인
	mongoStore := store.(*MongoStateStore)
	internalMetrics := mongoStore.GetInternalMetrics()

	t.Logf("Store metrics:")
	t.Logf("- Save operations: %d", internalMetrics.saveOps)
	t.Logf("- Average save time: %v", time.Duration(internalMetrics.totalSaveTime.Nanoseconds()/internalMetrics.saveOps))
	t.Logf("- Compression saved: %d bytes", internalMetrics.compressionSaved)

	assert.Greater(t, internalMetrics.saveOps, int64(0))
	assert.Greater(t, internalMetrics.compressionSaved, int64(0))
}

func TestIntegration_DocumentVersionCompatibility(t *testing.T) {
	// Given
	store := setupIntegrationTestStore(t)
	defer store.Close()

	ctx := context.Background()
	aggregateID := uuid.New()

	// When - 현재 버전으로 저장
	state := createIntegrationTestState(aggregateID, 1, createDummyData())
	err := store.Save(ctx, state)
	require.NoError(t, err)

	// Then - 문서 버전 확인 (직접 MongoDB에서)
	mongoStore := store.(*MongoStateStore)

	var doc mongoStateDocument
	filter := bson.M{"aggregateId": aggregateID.String()}
	err = mongoStore.collection.FindOne(ctx, filter).Decode(&doc)
	require.NoError(t, err)

	assert.Equal(t, DocumentVersionCurrent, doc.DocumentVersion)

	// 문서 구조 검증
	assert.Equal(t, "json", doc.DataFormat)
	assert.Equal(t, "raw", doc.DataEncoding)
	assert.False(t, doc.CreatedAt.IsZero())
	assert.False(t, doc.UpdatedAt.IsZero())
}

func TestIntegration_ConcurrentAccess(t *testing.T) {
	// Given
	store := setupIntegrationTestStore(t)
	defer store.Close()

	ctx := context.Background()
	aggregateID := uuid.New()

	// When - 동시에 여러 버전 저장
	const numGoroutines = 10
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(version int) {
			state := createIntegrationTestState(aggregateID, version, createDummyData())
			state.SetMetadata("goroutine", version)
			err := store.Save(ctx, state)
			errors <- err
		}(i + 1)
	}

	// Then - 모든 고루틴 완료 대기
	for i := 0; i < numGoroutines; i++ {
		err := <-errors
		assert.NoError(t, err)
	}

	// 저장된 상태 개수 확인
	count, err := store.Count(ctx, aggregateID)
	assert.NoError(t, err)
	assert.Equal(t, int64(numGoroutines), count)

	// 모든 버전이 저장되었는지 확인
	states, err := store.List(ctx, aggregateID)
	assert.NoError(t, err)
	assert.Len(t, states, numGoroutines)
}

// mongo_state_store_test.go - MongoDB 상태 저장소 기본 CRUD 테스트
package cqrsx

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 테스트용 MongoDB 설정
func setupTestStore(t *testing.T, storeOptions ...StateStoreOption) StateStore {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	// 테스트별 고유 컬렉션 사용
	collectionName := "test_states_" + uuid.New().String()
	collection := client.Database("test_cqrsx").Collection(collectionName)

	t.Cleanup(func() {
		collection.Drop(ctx)
		client.Disconnect(ctx)
	})

	// 기본 옵션으로 인덱싱과 메트릭 활성화
	defaultOptions := []StateStoreOption{WithIndexing(), WithMetrics()}
	allOptions := append(defaultOptions, storeOptions...)

	return NewMongoStateStore(collection, client, allOptions...)
}

// 테스트용 상태 생성 헬퍼
func createTestMongoState(aggregateID uuid.UUID, version int) *AggregateState {
	data := map[string]interface{}{
		"name":        "Test Entity",
		"version":     version,
		"description": "Test data for CRUD operations",
		"tags":        []string{"test", "crud"},
	}
	serializedData, _ := json.Marshal(data)

	state := NewAggregateState(aggregateID, "TestEntity", version, serializedData)
	state.SetMetadata("testCase", "crud")
	state.SetMetadata("created", time.Now().Format(time.RFC3339))

	return state
}

func TestNewMongoStateStore(t *testing.T) {
	tests := []struct {
		name    string
		options []StateStoreOption
	}{
		{
			name:    "Default configuration",
			options: []StateStoreOption{},
		},
		{
			name:    "With compression",
			options: []StateStoreOption{WithCompression(CompressionGzip)},
		},
		{
			name:    "With encryption",
			options: []StateStoreOption{WithEncryption(NewAESEncryptor("test-key"))},
		},
		{
			name: "With full options",
			options: []StateStoreOption{
				WithCompression(CompressionLZ4),
				WithEncryption(NewAESEncryptor("test-key")),
				WithBatchSize(50),
				WithMetrics(),
				WithIndexing(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			ctx := context.Background()
			client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
			require.NoError(t, err)
			defer client.Disconnect(ctx)

			collection := client.Database("test_cqrsx").Collection("test_creation")
			defer collection.Drop(ctx)

			// When
			store := NewMongoStateStore(collection, client, tt.options...)

			// Then
			assert.NotNil(t, store)
			assert.NotNil(t, store.collection)
			assert.NotNil(t, store.client)
			assert.NotNil(t, store.config)
			assert.NotNil(t, store.metrics)

			// Close 테스트
			err = store.Close()
			assert.NoError(t, err)
		})
	}
}

func TestMongoStateStore_Save(t *testing.T) {
	tests := []struct {
		name          string
		setupState    func() *AggregateState
		expectError   bool
		expectedError string
	}{
		{
			name: "Valid state save",
			setupState: func() *AggregateState {
				return createTestMongoState(uuid.New(), 1)
			},
			expectError: false,
		},
		{
			name: "Large data save",
			setupState: func() *AggregateState {
				largeData := make(map[string]string, 1000)
				for i := 0; i < 1000; i++ {
					largeData[string(rune('a'+i%26))] = "Large data content that should be compressed well"
				}
				serialized, _ := json.Marshal(largeData)
				return NewAggregateState(uuid.New(), "LargeEntity", 1, serialized)
			},
			expectError: false,
		},
		{
			name: "Empty data save",
			setupState: func() *AggregateState {
				return NewAggregateState(uuid.New(), "EmptyEntity", 1, []byte{})
			},
			expectError: false,
		},
		{
			name: "Multiple versions of same aggregate",
			setupState: func() *AggregateState {
				id := uuid.New()
				return createTestMongoState(id, 2) // version 2
			},
			expectError: false,
		},
		{
			name: "Invalid state - nil UUID",
			setupState: func() *AggregateState {
				state := createTestMongoState(uuid.New(), 1)
				state.string = uuid.Nil
				return state
			},
			expectError:   true,
			expectedError: "aggregate ID cannot be nil",
		},
		{
			name: "Invalid state - empty type",
			setupState: func() *AggregateState {
				state := createTestMongoState(uuid.New(), 1)
				state.AggregateType = ""
				return state
			},
			expectError:   true,
			expectedError: "aggregate type cannot be empty",
		},
		{
			name: "Invalid state - negative version",
			setupState: func() *AggregateState {
				state := createTestMongoState(uuid.New(), 1)
				state.Version = -1
				return state
			},
			expectError:   true,
			expectedError: "version cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			store := setupTestStore(t)
			defer store.Close()

			ctx := context.Background()
			state := tt.setupState()

			// When
			err := store.Save(ctx, state)

			// Then
			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			} else {
				assert.NoError(t, err)

				// 저장 성공 시 실제로 저장되었는지 확인
				exists, err := store.Exists(ctx, state.string, state.Version)
				assert.NoError(t, err)
				assert.True(t, exists)
			}
		})
	}
}

func TestMongoStateStore_Load(t *testing.T) {
	// Given
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	aggregateID := uuid.New()

	// 여러 버전 저장
	versions := []int{1, 3, 5, 7}
	for _, version := range versions {
		state := createTestMongoState(aggregateID, version)
		state.SetMetadata("version_info", version)
		err := store.Save(ctx, state)
		require.NoError(t, err)
	}

	tests := []struct {
		name            string
		aggregateID     uuid.UUID
		expectError     bool
		expectedError   string
		expectedVersion int
	}{
		{
			name:            "Load latest version",
			aggregateID:     aggregateID,
			expectError:     false,
			expectedVersion: 7, // 최신 버전
		},
		{
			name:          "Load non-existent aggregate",
			aggregateID:   uuid.New(),
			expectError:   true,
			expectedError: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			loadedState, err := store.Load(ctx, tt.aggregateID)

			// Then
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, loadedState)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, loadedState)
				assert.Equal(t, tt.aggregateID, loadedState.string)
				assert.Equal(t, tt.expectedVersion, loadedState.Version)
				assert.Equal(t, "TestEntity", loadedState.AggregateType)

				// 메타데이터 확인
				versionInfo, exists := loadedState.GetMetadata("version_info")
				assert.True(t, exists)
				assert.EqualValues(t, tt.expectedVersion, versionInfo)
			}
		})
	}
}

func TestMongoStateStore_LoadVersion(t *testing.T) {
	// Given
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	aggregateID := uuid.New()

	// 여러 버전 저장
	versions := []int{1, 3, 5, 7}
	for _, version := range versions {
		state := createTestMongoState(aggregateID, version)
		state.SetMetadata("version_info", version)
		err := store.Save(ctx, state)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		aggregateID   uuid.UUID
		version       int
		expectError   bool
		expectedError string
	}{
		{
			name:        "Load existing version 1",
			aggregateID: aggregateID,
			version:     1,
			expectError: false,
		},
		{
			name:        "Load existing version 5",
			aggregateID: aggregateID,
			version:     5,
			expectError: false,
		},
		{
			name:          "Load non-existent version",
			aggregateID:   aggregateID,
			version:       999,
			expectError:   true,
			expectedError: "not found",
		},
		{
			name:          "Load from non-existent aggregate",
			aggregateID:   uuid.New(),
			version:       1,
			expectError:   true,
			expectedError: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			loadedState, err := store.LoadVersion(ctx, tt.aggregateID, tt.version)

			// Then
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, loadedState)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, loadedState)
				assert.Equal(t, tt.aggregateID, loadedState.string)
				assert.Equal(t, tt.version, loadedState.Version)
				assert.Equal(t, "TestEntity", loadedState.AggregateType)

				// 메타데이터 확인
				versionInfo, exists := loadedState.GetMetadata("version_info")
				assert.True(t, exists)
				assert.EqualValues(t, tt.version, versionInfo)
			}
		})
	}
}

func TestMongoStateStore_Delete(t *testing.T) {
	// Given
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	aggregateID := uuid.New()

	// 여러 버전 저장
	versions := []int{1, 2, 3, 4, 5}
	for _, version := range versions {
		state := createTestMongoState(aggregateID, version)
		err := store.Save(ctx, state)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		aggregateID   uuid.UUID
		version       int
		expectError   bool
		expectedError string
	}{
		{
			name:        "Delete existing version",
			aggregateID: aggregateID,
			version:     3,
			expectError: false,
		},
		{
			name:          "Delete non-existent version",
			aggregateID:   aggregateID,
			version:       999,
			expectError:   true,
			expectedError: "not found",
		},
		{
			name:          "Delete from non-existent aggregate",
			aggregateID:   uuid.New(),
			version:       1,
			expectError:   true,
			expectedError: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			err := store.Delete(ctx, tt.aggregateID, tt.version)

			// Then
			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			} else {
				assert.NoError(t, err)

				// 삭제 확인
				exists, err := store.Exists(ctx, tt.aggregateID, tt.version)
				assert.NoError(t, err)
				assert.False(t, exists)

				// 다른 버전들은 여전히 존재하는지 확인
				count, err := store.Count(ctx, tt.aggregateID)
				assert.NoError(t, err)
				assert.Equal(t, int64(4), count) // 5개 중 1개 삭제되어 4개
			}
		})
	}
}

func TestMongoStateStore_DeleteAll(t *testing.T) {
	// Given
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	aggregateID1 := uuid.New()
	aggregateID2 := uuid.New()

	// 두 집합체에 각각 여러 버전 저장
	for _, aggregateID := range []uuid.UUID{aggregateID1, aggregateID2} {
		for version := 1; version <= 3; version++ {
			state := createTestMongoState(aggregateID, version)
			err := store.Save(ctx, state)
			require.NoError(t, err)
		}
	}

	// When - aggregateID1의 모든 버전 삭제
	err := store.DeleteAll(ctx, aggregateID1)

	// Then
	assert.NoError(t, err)

	// aggregateID1의 모든 버전이 삭제되었는지 확인
	count1, err := store.Count(ctx, aggregateID1)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count1)

	// aggregateID2는 영향받지 않았는지 확인
	count2, err := store.Count(ctx, aggregateID2)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count2)

	// When - 존재하지 않는 집합체 삭제 시도
	err = store.DeleteAll(ctx, uuid.New())

	// Then - 에러 없이 성공해야 함
	assert.NoError(t, err)
}

func TestMongoStateStore_List(t *testing.T) {
	// Given
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	aggregateID := uuid.New()

	// 비순차적 버전들로 저장 (정렬 테스트를 위해)
	versions := []int{5, 1, 3, 7, 2}
	for _, version := range versions {
		state := createTestMongoState(aggregateID, version)
		state.SetMetadata("save_order", len(versions)-version) // 저장 순서와 다른 값
		err := store.Save(ctx, state)
		require.NoError(t, err)
	}

	tests := []struct {
		name        string
		aggregateID uuid.UUID
		expectedLen int
	}{
		{
			name:        "List all versions",
			aggregateID: aggregateID,
			expectedLen: 5,
		},
		{
			name:        "List non-existent aggregate",
			aggregateID: uuid.New(),
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			states, err := store.List(ctx, tt.aggregateID)

			// Then
			assert.NoError(t, err)
			assert.Len(t, states, tt.expectedLen)

			if tt.expectedLen > 0 {
				// 버전순으로 정렬되어 있는지 확인
				expectedVersions := []int{1, 2, 3, 5, 7}
				for i, state := range states {
					assert.Equal(t, expectedVersions[i], state.Version)
					assert.Equal(t, tt.aggregateID, state.string)
					assert.Equal(t, "TestEntity", state.AggregateType)
				}
			}
		})
	}
}

func TestMongoStateStore_Count(t *testing.T) {
	// Given
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	aggregateID := uuid.New()

	tests := []struct {
		name          string
		setupVersions []int
		aggregateID   uuid.UUID
		expectedCount int64
	}{
		{
			name:          "Count zero states",
			setupVersions: []int{},
			aggregateID:   aggregateID,
			expectedCount: 0,
		},
		{
			name:          "Count single state",
			setupVersions: []int{1},
			aggregateID:   aggregateID,
			expectedCount: 1,
		},
		{
			name:          "Count multiple states",
			setupVersions: []int{1, 2, 3, 4, 5},
			aggregateID:   aggregateID,
			expectedCount: 5,
		},
		{
			name:          "Count non-existent aggregate",
			setupVersions: []int{},
			aggregateID:   uuid.New(),
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given - 테스트별 새로운 집합체 ID 사용
			testAggregateID := uuid.New()
			if tt.name == "Count non-existent aggregate" {
				testAggregateID = tt.aggregateID // 기존 집합체 ID 사용
			}

			// 상태들 저장
			for _, version := range tt.setupVersions {
				state := createTestMongoState(testAggregateID, version)
				err := store.Save(ctx, state)
				require.NoError(t, err)
			}

			// When
			count, err := store.Count(ctx, testAggregateID)

			// Then
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}

func TestMongoStateStore_Exists(t *testing.T) {
	// Given
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	aggregateID := uuid.New()

	// 몇 개 버전 저장
	versions := []int{1, 3, 5}
	for _, version := range versions {
		state := createTestMongoState(aggregateID, version)
		err := store.Save(ctx, state)
		require.NoError(t, err)
	}

	tests := []struct {
		name        string
		aggregateID uuid.UUID
		version     int
		expected    bool
	}{
		{
			name:        "Exists - version 1",
			aggregateID: aggregateID,
			version:     1,
			expected:    true,
		},
		{
			name:        "Exists - version 3",
			aggregateID: aggregateID,
			version:     3,
			expected:    true,
		},
		{
			name:        "Exists - version 5",
			aggregateID: aggregateID,
			version:     5,
			expected:    true,
		},
		{
			name:        "Not exists - version 2",
			aggregateID: aggregateID,
			version:     2,
			expected:    false,
		},
		{
			name:        "Not exists - version 999",
			aggregateID: aggregateID,
			version:     999,
			expected:    false,
		},
		{
			name:        "Not exists - different aggregate",
			aggregateID: uuid.New(),
			version:     1,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			exists, err := store.Exists(ctx, tt.aggregateID, tt.version)

			// Then
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, exists)
		})
	}
}

func TestMongoStateStore_CRUD_Integration(t *testing.T) {
	// Given
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	aggregateID := uuid.New()

	// When & Then - 전체 CRUD 플로우 테스트

	// 1. 초기 상태 확인 (존재하지 않음)
	count, err := store.Count(ctx, aggregateID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	exists, err := store.Exists(ctx, aggregateID, 1)
	assert.NoError(t, err)
	assert.False(t, exists)

	// 2. 첫 번째 버전 저장
	state1 := createTestMongoState(aggregateID, 1)
	err = store.Save(ctx, state1)
	assert.NoError(t, err)

	// 3. 저장 확인
	exists, err = store.Exists(ctx, aggregateID, 1)
	assert.NoError(t, err)
	assert.True(t, exists)

	count, err = store.Count(ctx, aggregateID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// 4. 로드 테스트
	loadedState, err := store.Load(ctx, aggregateID)
	assert.NoError(t, err)
	assert.Equal(t, state1.string, loadedState.string)
	assert.Equal(t, state1.Version, loadedState.Version)
	assert.Equal(t, state1.Data, loadedState.Data)

	// 5. 추가 버전들 저장
	for version := 2; version <= 5; version++ {
		state := createTestMongoState(aggregateID, version)
		err = store.Save(ctx, state)
		assert.NoError(t, err)
	}

	// 6. 리스트 테스트
	states, err := store.List(ctx, aggregateID)
	assert.NoError(t, err)
	assert.Len(t, states, 5)

	// 7. 특정 버전 로드
	state3, err := store.LoadVersion(ctx, aggregateID, 3)
	assert.NoError(t, err)
	assert.Equal(t, 3, state3.Version)

	// 8. 특정 버전 삭제
	err = store.Delete(ctx, aggregateID, 3)
	assert.NoError(t, err)

	exists, err = store.Exists(ctx, aggregateID, 3)
	assert.NoError(t, err)
	assert.False(t, exists)

	count, err = store.Count(ctx, aggregateID)
	assert.NoError(t, err)
	assert.Equal(t, int64(4), count)

	// 9. 최신 버전이 여전히 5인지 확인
	latest, err := store.Load(ctx, aggregateID)
	assert.NoError(t, err)
	assert.Equal(t, 5, latest.Version)

	// 10. 모든 버전 삭제
	err = store.DeleteAll(ctx, aggregateID)
	assert.NoError(t, err)

	count, err = store.Count(ctx, aggregateID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// 11. 삭제 후 로드 시도 (에러 발생해야 함)
	_, err = store.Load(ctx, aggregateID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// eventstore_test.go - 종합 테스트 suite
package cqrsx

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EventStoreTestSuite는 모든 이벤트 저장소 구현체에 대한 공통 테스트입니다
type EventStoreTestSuite struct {
	suite.Suite
	store       EventStore
	client      *mongo.Client
	database    *mongo.Database
	collection  *mongo.Collection
	ctx         context.Context
	testGuildID uuid.UUID
	testUserID  uuid.UUID
}

// SetupSuite는 테스트 suite 설정을 수행합니다
func (s *EventStoreTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.testGuildID = uuid.New()
	s.testUserID = uuid.New()

	// 테스트용 MongoDB 연결
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(s.ctx, clientOptions)
	require.NoError(s.T(), err)

	s.client = client
	s.database = client.Database("test_gameevents")
	s.collection = s.database.Collection("test_events")
}

// TearDownSuite는 테스트 suite 정리를 수행합니다
func (s *EventStoreTestSuite) TearDownSuite() {
	if s.client != nil {
		s.database.Drop(s.ctx)
		s.client.Disconnect(s.ctx)
	}
}

// SetupTest는 각 테스트 전 초기화를 수행합니다
func (s *EventStoreTestSuite) SetupTest() {
	s.collection.Drop(s.ctx)
}

// TestSaveAndLoadEvents는 기본적인 저장/로드 기능을 테스트합니다
func (s *EventStoreTestSuite) TestSaveAndLoadEvents() {
	// Given: 새로운 길드 이벤트들
	events := s.createTestGuildEvents()

	// When: 이벤트들을 저장
	err := s.store.Save(s.ctx, events, 0)

	// Then: 저장이 성공해야 함
	require.NoError(s.T(), err)

	// When: 이벤트들을 로드
	loadedEvents, err := s.store.Load(s.ctx, s.testGuildID)

	// Then: 로드가 성공하고 모든 이벤트가 일치해야 함
	require.NoError(s.T(), err)
	assert.Len(s.T(), loadedEvents, len(events))

	for i, event := range events {
		assert.Equal(s.T(), event.AggregateID(), loadedEvents[i].AggregateID())
		assert.Equal(s.T(), event.EventType(), loadedEvents[i].EventType())
		assert.Equal(s.T(), event.Version(), loadedEvents[i].Version())
	}
}

// TestConcurrencyControl은 동시성 제어를 테스트합니다
func (s *EventStoreTestSuite) TestConcurrencyControl() {
	// Given: 초기 이벤트
	initialEvents := s.createTestGuildEvents()[:1] // 첫 번째 이벤트만
	err := s.store.Save(s.ctx, initialEvents, 0)
	require.NoError(s.T(), err)

	// When: 잘못된 버전으로 저장 시도 (동시성 충돌 시뮬레이션)
	newEvents := []Event{
		s.createMemberJoinedEvent(2), // 버전 2
	}
	err = s.store.Save(s.ctx, newEvents, 0) // 잘못된 예상 버전

	// Then: 동시성 충돌 에러가 발생해야 함
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrConcurrencyConflict, err)
}

// TestLoadFromVersion은 특정 버전부터 로드하는 기능을 테스트합니다
func (s *EventStoreTestSuite) TestLoadFromVersion() {
	// Given: 여러 이벤트 저장
	events := s.createTestGuildEvents()
	err := s.store.Save(s.ctx, events, 0)
	require.NoError(s.T(), err)

	// When: 버전 3부터 로드
	fromVersion := 3
	loadedEvents, err := s.store.LoadFrom(s.ctx, s.testGuildID, fromVersion)

	// Then: 버전 3 이상의 이벤트만 로드되어야 함
	require.NoError(s.T(), err)
	for _, event := range loadedEvents {
		assert.GreaterOrEqual(s.T(), event.Version(), fromVersion)
	}
}

// TestEmptyStream은 빈 스트림 처리를 테스트합니다
func (s *EventStoreTestSuite) TestEmptyStream() {
	// Given: 존재하지 않는 집합체 ID

	// When: 빈 스트림 로드 시도
	events, err := s.store.Load(s.ctx, uuid.New())

	// Then: 에러 없이 빈 슬라이스를 반환해야 함
	require.NoError(s.T(), err)
	assert.Empty(s.T(), events)
}

// TestLargeEventStream은 대용량 이벤트 스트림을 테스트합니다
func (s *EventStoreTestSuite) TestLargeEventStream() {
	// Given: 대량의 이벤트 생성 (1000개)
	events := make([]Event, 1000)
	for i := 0; i < 1000; i++ {
		events[i] = s.createResourceSharedEvent(i+1, "gold", 100)
	}

	// When: 배치로 저장
	batchSize := 100
	for i := 0; i < len(events); i += batchSize {
		end := i + batchSize
		if end > len(events) {
			end = len(events)
		}
		batch := events[i:end]
		err := s.store.Save(s.ctx, batch, i)
		require.NoError(s.T(), err)
	}

	// Then: 모든 이벤트가 정확히 로드되어야 함
	loadedEvents, err := s.store.Load(s.ctx, s.testGuildID)
	require.NoError(s.T(), err)
	assert.Len(s.T(), loadedEvents, 1000)
}

// TestEventMetadata는 이벤트 메타데이터 처리를 테스트합니다
func (s *EventStoreTestSuite) TestEventMetadata() {
	// Given: 메타데이터가 포함된 이벤트
	event := s.createGuildCreatedEvent()
	event.Metadata()["userId"] = s.testUserID.String()
	event.Metadata()["correlationId"] = uuid.New().String()
	event.Metadata()["source"] = "api"

	// When: 저장 후 로드
	err := s.store.Save(s.ctx, []Event{event}, 0)
	require.NoError(s.T(), err)

	loadedEvents, err := s.store.Load(s.ctx, s.testGuildID)
	require.NoError(s.T(), err)

	// Then: 메타데이터가 보존되어야 함
	require.Len(s.T(), loadedEvents, 1)
	metadata := loadedEvents[0].Metadata()
	assert.Equal(s.T(), s.testUserID.String(), metadata["userId"])
	assert.Equal(s.T(), "api", metadata["source"])
}

// TestPerformanceMetrics는 성능 메트릭 수집을 테스트합니다
func (s *EventStoreTestSuite) TestPerformanceMetrics() {
	// Given: 여러 작업 수행
	events := s.createTestGuildEvents()

	// When: 저장과 로드 작업
	s.store.Save(s.ctx, events, 0)
	s.store.Load(s.ctx, s.testGuildID)
	s.store.LoadFrom(s.ctx, s.testGuildID, 1)

	// Then: 메트릭이 수집되어야 함
	metrics := s.store.GetMetrics()
	assert.Greater(s.T(), metrics.SaveOperations, int64(0))
	assert.Greater(s.T(), metrics.LoadOperations, int64(0))
	assert.NotZero(s.T(), metrics.LastOperation)
}

// Helper methods

func (s *EventStoreTestSuite) createTestGuildEvents() []Event {
	return []Event{
		s.createGuildCreatedEvent(),
		s.createMemberJoinedEvent(2),
		s.createResourceSharedEvent(3, "gold", 1000),
		s.createMemberJoinedEvent(4),
		s.createGuildLevelUpEvent(5),
	}
}

func (s *EventStoreTestSuite) createGuildCreatedEvent() *BaseEvent {
	data := &GuildCreatedEvent{
		GuildID:    s.testGuildID,
		FounderID:  s.testUserID,
		Name:       "Test Guild",
		MaxMembers: 50,
		GameType:   "RPG",
		CreatedAt:  time.Now(),
	}
	return NewBaseEvent(s.testGuildID, "GuildCreated", data, 1)
}

func (s *EventStoreTestSuite) createMemberJoinedEvent(version int) *BaseEvent {
	data := &MemberJoinedEvent{
		GuildID:          s.testGuildID,
		MemberID:         uuid.New(),
		Role:             "member",
		InvitedBy:        s.testUserID,
		JoinedAt:         time.Now(),
		InvitationMethod: "direct",
	}
	return NewBaseEvent(s.testGuildID, "MemberJoined", data, version)
}

func (s *EventStoreTestSuite) createResourceSharedEvent(version int, resourceType string, amount int) *BaseEvent {
	data := &ResourceSharedEvent{
		GuildID:       s.testGuildID,
		ContributorID: s.testUserID,
		ResourceType:  resourceType,
		Amount:        amount,
		TotalBalance:  amount * version,
		SharedAt:      time.Now(),
	}
	return NewBaseEvent(s.testGuildID, "ResourceShared", data, version)
}

func (s *EventStoreTestSuite) createGuildLevelUpEvent(version int) *BaseEvent {
	data := &GuildLevelUpEvent{
		GuildID:    s.testGuildID,
		OldLevel:   version - 1,
		NewLevel:   version,
		Experience: version * 1000,
		LevelUpAt:  time.Now(),
		Rewards:    []string{"gold_bonus", "exp_bonus"},
	}
	return NewBaseEvent(s.testGuildID, "GuildLevelUp", data, version)
}

// 구체적인 구현체별 테스트 suites

// StreamEventStoreTestSuite는 스트림 이벤트 저장소 전용 테스트입니다
type StreamEventStoreTestSuite struct {
	EventStoreTestSuite
}

func (s *StreamEventStoreTestSuite) SetupTest() {
	s.EventStoreTestSuite.SetupTest()
	serializer := NewJSONEventSerializer()
	s.store = NewStreamEventStore(s.collection, serializer)
}

func TestStreamEventStore(t *testing.T) {
	suite.Run(t, new(StreamEventStoreTestSuite))
}

// DocumentEventStoreTestSuite는 문서 이벤트 저장소 전용 테스트입니다
type DocumentEventStoreTestSuite struct {
	EventStoreTestSuite
}

func (s *DocumentEventStoreTestSuite) SetupTest() {
	s.EventStoreTestSuite.SetupTest()
	serializer := NewJSONEventSerializer()
	s.store = NewDocumentEventStore(s.collection, s.client, serializer)
}

// TestComplexQueries는 복잡한 쿼리 기능을 테스트합니다 (DocumentEventStore 전용)
func (s *DocumentEventStoreTestSuite) TestComplexQueries() {
	// Given: 다양한 이벤트들 저장
	events := s.createTestGuildEvents()
	err := s.store.Save(s.ctx, events, 0)
	require.NoError(s.T(), err)

	queryableStore := s.store.(*DocumentEventStore)

	// When: 특정 이벤트 타입으로 쿼리
	query := EventQuery{
		EventTypes: []EventType{"MemberJoined"},
	}
	foundEvents, err := queryableStore.FindEvents(s.ctx, query)

	// Then: 해당 타입의 이벤트만 반환되어야 함
	require.NoError(s.T(), err)
	for _, event := range foundEvents {
		assert.Equal(s.T(), EventType("MemberJoined"), event.EventType())
	}
}

func TestDocumentEventStore(t *testing.T) {
	suite.Run(t, new(DocumentEventStoreTestSuite))
}

// HybridEventStoreTestSuite는 하이브리드 이벤트 저장소 전용 테스트입니다
type HybridEventStoreTestSuite struct {
	EventStoreTestSuite
	hotStore  EventStore
	coldStore QueryableEventStore
}

func (s *HybridEventStoreTestSuite) SetupTest() {
	s.EventStoreTestSuite.SetupTest()

	serializer := NewJSONEventSerializer()
	s.hotStore = NewStreamEventStore(s.collection, serializer)

	coldCollection := s.database.Collection("test_events_archive")
	s.coldStore = NewDocumentEventStore(coldCollection, s.client, serializer)

	config := HybridConfig{
		HotDataThreshold:  24 * time.Hour, // 24시간
		ArchiveInterval:   time.Hour,      // 1시간마다
		MaxHotEvents:      100,
		EnableAutoArchive: false, // 테스트에서는 수동 제어
	}

	s.store = NewHybridEventStore(s.hotStore, s.coldStore, config)
}

// TestHotColdDataSeparation은 Hot/Cold 데이터 분리를 테스트합니다
func (s *HybridEventStoreTestSuite) TestHotColdDataSeparation() {
	// Given: 최근 이벤트와 오래된 이벤트
	recentEvents := s.createTestGuildEvents()

	// When: 최근 이벤트 저장 (Hot 스토어에 저장됨)
	err := s.store.Save(s.ctx, recentEvents, 0)
	require.NoError(s.T(), err)

	// Then: Hot 스토어에서 직접 조회 가능해야 함
	hotEvents, err := s.hotStore.Load(s.ctx, s.testGuildID)
	require.NoError(s.T(), err)
	assert.Len(s.T(), hotEvents, len(recentEvents))

	// And: 하이브리드 스토어에서도 모든 데이터 조회 가능해야 함
	allEvents, err := s.store.Load(s.ctx, s.testGuildID)
	require.NoError(s.T(), err)
	assert.Len(s.T(), allEvents, len(recentEvents))
}

func TestHybridEventStore(t *testing.T) {
	suite.Run(t, new(HybridEventStoreTestSuite))
}

// EventStoreMigratorTestSuite는 마이그레이션 도구 테스트입니다
type EventStoreMigratorTestSuite struct {
	suite.Suite
	ctx         context.Context
	client      *mongo.Client
	database    *mongo.Database
	sourceStore EventStore
	targetStore EventStore
	testGuildID uuid.UUID
}

func (s *EventStoreMigratorTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.testGuildID = uuid.New()

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(s.ctx, clientOptions)
	require.NoError(s.T(), err)

	s.client = client
	s.database = client.Database("test_migration")
}

func (s *EventStoreMigratorTestSuite) SetupTest() {
	s.database.Drop(s.ctx)

	serializer := NewJSONEventSerializer()

	sourceCollection := s.database.Collection("source_events")
	s.sourceStore = NewStreamEventStore(sourceCollection, serializer)

	targetCollection := s.database.Collection("target_events")
	s.targetStore = NewDocumentEventStore(targetCollection, s.client, serializer)
}

func (s *EventStoreMigratorTestSuite) TearDownSuite() {
	if s.client != nil {
		s.database.Drop(s.ctx)
		s.client.Disconnect(s.ctx)
	}
}

// TestStreamToDocumentMigration은 스트림에서 문서 방식으로 마이그레이션을 테스트합니다
func (s *EventStoreMigratorTestSuite) TestStreamToDocumentMigration() {
	// Given: 소스 스토어에 이벤트들 저장
	events := s.createTestEvents()
	err := s.sourceStore.Save(s.ctx, events, 0)
	require.NoError(s.T(), err)

	// When: 마이그레이션 실행
	migrator := NewEventStoreMigrator(s.sourceStore, s.targetStore)
	err = migrator.MigrateStream(s.ctx, s.testGuildID)
	require.NoError(s.T(), err)

	// Then: 타겟 스토어에서 모든 이벤트가 조회되어야 함
	targetEvents, err := s.targetStore.Load(s.ctx, s.testGuildID)
	require.NoError(s.T(), err)
	assert.Len(s.T(), targetEvents, len(events))

	// And: 이벤트 내용이 일치해야 함
	for i, originalEvent := range events {
		assert.Equal(s.T(), originalEvent.EventType(), targetEvents[i].EventType())
		assert.Equal(s.T(), originalEvent.Version(), targetEvents[i].Version())
	}
}

func (s *EventStoreMigratorTestSuite) createTestEvents() []Event {
	events := make([]Event, 5)
	for i := 0; i < 5; i++ {
		data := &ResourceSharedEvent{
			GuildID:       s.testGuildID,
			ContributorID: uuid.New(),
			ResourceType:  "gold",
			Amount:        100 * (i + 1),
			TotalBalance:  100 * (i + 1) * (i + 1),
			SharedAt:      time.Now(),
		}
		events[i] = NewBaseEvent(s.testGuildID, "ResourceShared", data, i+1)
	}
	return events
}

func TestEventStoreMigrator(t *testing.T) {
	suite.Run(t, new(EventStoreMigratorTestSuite))
}

// BenchmarkEventStores는 성능 벤치마크 테스트입니다
func BenchmarkEventStores(b *testing.B) {
	ctx := context.Background()
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	require.NoError(b, err)
	defer client.Disconnect(ctx)

	database := client.Database("benchmark_test")
	defer database.Drop(ctx)

	serializer := NewJSONEventSerializer()
	testGuildID := uuid.New()

	// 테스트 이벤트 생성
	createEvents := func(count int) []Event {
		events := make([]Event, count)
		for i := 0; i < count; i++ {
			data := &ResourceSharedEvent{
				GuildID:       testGuildID,
				ContributorID: uuid.New(),
				ResourceType:  "gold",
				Amount:        100,
				TotalBalance:  100 * (i + 1),
				SharedAt:      time.Now(),
			}
			events[i] = NewBaseEvent(testGuildID, "ResourceShared", data, i+1)
		}
		return events
	}

	b.Run("StreamEventStore_Save", func(b *testing.B) {
		collection := database.Collection("stream_bench")
		store := NewStreamEventStore(collection, serializer)
		events := createEvents(100)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			collection.Drop(ctx)
			store.Save(ctx, events, 0)
		}
	})

	b.Run("DocumentEventStore_Save", func(b *testing.B) {
		collection := database.Collection("doc_bench")
		store := NewDocumentEventStore(collection, client, serializer)
		events := createEvents(100)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			collection.Drop(ctx)
			store.Save(ctx, events, 0)
		}
	})

	b.Run("StreamEventStore_Load", func(b *testing.B) {
		collection := database.Collection("stream_load_bench")
		store := NewStreamEventStore(collection, serializer)
		events := createEvents(100)
		store.Save(ctx, events, 0)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			store.Load(ctx, testGuildID)
		}
	})

	b.Run("DocumentEventStore_Load", func(b *testing.B) {
		collection := database.Collection("doc_load_bench")
		store := NewDocumentEventStore(collection, client, serializer)
		events := createEvents(100)
		store.Save(ctx, events, 0)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			store.Load(ctx, testGuildID)
		}
	})
}

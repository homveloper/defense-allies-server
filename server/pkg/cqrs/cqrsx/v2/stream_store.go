// stream_store.go - 이벤트 스트림 방식 구현
package cqrsx

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// StreamEventStore는 MongoDB 스트림 방식 이벤트 저장소입니다
type StreamEventStore struct {
	collection *mongo.Collection
	serializer EventSerializer
	metrics    *streamMetrics
	maxEvents  int // 스트림당 최대 이벤트 수 (16MB 제한 대응)
}

type streamMetrics struct {
	mu            sync.RWMutex
	saveOps       int64
	loadOps       int64
	totalSaveTime time.Duration
	totalLoadTime time.Duration
	errors        int64
	lastOperation time.Time
}

// EventStreamDocument는 MongoDB에 저장되는 스트림 문서입니다
type EventStreamDocument struct {
	ID         string          `bson:"_id"`
	StreamName string          `bson:"streamName"`
	StreamType string          `bson:"streamType"`
	Version    int             `bson:"version"`
	Events     []EventDocument `bson:"events"`
	UpdatedAt  time.Time       `bson:"updatedAt"`
	CreatedAt  time.Time       `bson:"createdAt"`
}

// EventDocument는 스트림 내의 개별 이벤트입니다
type EventDocument struct {
	EventID   string      `bson:"eventId"`
	EventType string      `bson:"eventType"`
	Version   int         `bson:"version"`
	Timestamp time.Time   `bson:"timestamp"`
	Data      interface{} `bson:"data"`
	Metadata  interface{} `bson:"metadata"`
}

// NewStreamEventStore는 새로운 스트림 이벤트 저장소를 생성합니다
func NewStreamEventStore(collection *mongo.Collection, serializer EventSerializer) *StreamEventStore {
	return &StreamEventStore{
		collection: collection,
		serializer: serializer,
		metrics:    &streamMetrics{},
		maxEvents:  1000, // 기본값: 스트림당 1000개 이벤트
	}
}

// Save는 이벤트들을 스트림에 저장합니다
func (s *StreamEventStore) Save(ctx context.Context, events []Event, expectedVersion int) error {
	if len(events) == 0 {
		return nil
	}

	start := time.Now()
	defer func() {
		s.updateMetrics(true, time.Since(start), nil)
	}()

	streamName := s.getStreamName(events[0].string())
	eventDocs := make([]EventDocument, len(events))

	// 이벤트들을 문서로 변환
	for i, event := range events {
		eventDocs[i] = EventDocument{
			EventID:   uuid.New().String(),
			EventType: string(event.EventType()),
			Version:   expectedVersion + i + 1,
			Timestamp: event.Timestamp(),
			Data:      event.Data(),
			Metadata:  event.Metadata(),
		}
	}

	// 낙관적 동시성 제어로 저장
	filter := bson.M{
		"streamName": streamName,
		"version":    expectedVersion,
	}

	update := bson.M{
		"$push": bson.M{
			"events": bson.M{"$each": eventDocs},
		},
		"$inc": bson.M{
			"version": len(events),
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
		"$setOnInsert": bson.M{
			"_id":        streamName,
			"streamName": streamName,
			"streamType": s.getStreamType(events[0]),
			"createdAt":  time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)
	result, err := s.collection.UpdateOne(ctx, filter, update, opts)

	if err != nil {
		s.updateMetrics(true, time.Since(start), err)
		return fmt.Errorf("failed to save events: %w", err)
	}

	// 동시성 충돌 확인
	if result.MatchedCount == 0 && result.UpsertedCount == 0 {
		s.updateMetrics(true, time.Since(start), ErrConcurrencyConflict)
		return ErrConcurrencyConflict
	}

	// 스트림 크기 확인 및 분할 (16MB 제한 대응)
	if err := s.checkAndSplitStream(ctx, streamName); err != nil {
		// 로그만 남기고 계속 진행 (비동기로 처리)
		fmt.Printf("Warning: failed to check stream size: %v\n", err)
	}

	return nil
}

// Load는 집합체의 모든 이벤트를 로드합니다
func (s *StreamEventStore) Load(ctx context.Context, aggregateID uuid.UUID) ([]Event, error) {
	start := time.Now()
	defer func() {
		s.updateMetrics(false, time.Since(start), nil)
	}()

	streamName := s.getStreamName(aggregateID)

	var doc EventStreamDocument
	err := s.collection.FindOne(ctx, bson.M{"streamName": streamName}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []Event{}, nil
		}
		s.updateMetrics(false, time.Since(start), err)
		return nil, fmt.Errorf("failed to load stream: %w", err)
	}

	// 이벤트 문서들을 Event 인터페이스로 변환
	events := make([]Event, len(doc.Events))
	for i, eventDoc := range doc.Events {
		event, err := s.deserializeEvent(eventDoc, aggregateID)
		if err != nil {
			s.updateMetrics(false, time.Since(start), err)
			return nil, fmt.Errorf("failed to deserialize event %d: %w", i, err)
		}
		events[i] = event
	}

	return events, nil
}

// LoadFrom은 특정 버전부터 이벤트를 로드합니다
func (s *StreamEventStore) LoadFrom(ctx context.Context, aggregateID uuid.UUID, fromVersion int) ([]Event, error) {
	allEvents, err := s.Load(ctx, aggregateID)
	if err != nil {
		return nil, err
	}

	// 버전 필터링
	var filteredEvents []Event
	for _, event := range allEvents {
		if event.Version() >= fromVersion {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return filteredEvents, nil
}

// GetMetrics는 성능 메트릭을 반환합니다
func (s *StreamEventStore) GetMetrics() StoreMetrics {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	var avgSave, avgLoad time.Duration
	if s.metrics.saveOps > 0 {
		avgSave = s.metrics.totalSaveTime / time.Duration(s.metrics.saveOps)
	}
	if s.metrics.loadOps > 0 {
		avgLoad = s.metrics.totalLoadTime / time.Duration(s.metrics.loadOps)
	}

	return StoreMetrics{
		SaveOperations:  s.metrics.saveOps,
		LoadOperations:  s.metrics.loadOps,
		AverageSaveTime: avgSave,
		AverageLoadTime: avgLoad,
		ErrorCount:      s.metrics.errors,
		LastOperation:   s.metrics.lastOperation,
		StorageStrategy: StrategyStream,
	}
}

// Close는 연결을 정리합니다
func (s *StreamEventStore) Close() error {
	// MongoDB 컬렉션은 별도로 닫을 필요가 없음
	return nil
}

// CreateIndexes는 성능 최적화를 위한 인덱스를 생성합니다
func (s *StreamEventStore) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"streamName", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{"streamType", 1},
				{"updatedAt", -1},
			},
		},
		{
			Keys: bson.D{{"events.eventType", 1}},
		},
		{
			Keys: bson.D{{"events.timestamp", -1}},
		},
	}

	_, err := s.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// Private helper methods

func (s *StreamEventStore) getStreamName(aggregateID uuid.UUID) string {
	return fmt.Sprintf("guild-%s", aggregateID.String())
}

func (s *StreamEventStore) getStreamType(event Event) string {
	// 이벤트 타입에서 스트림 타입 추출
	return "Guild" // 게임 길드 시스템용
}

func (s *StreamEventStore) deserializeEvent(eventDoc EventDocument, aggregateID uuid.UUID) (Event, error) {
	return &BaseEvent{
		aggregateID: aggregateID,
		eventType:   EventType(eventDoc.EventType),
		data:        eventDoc.Data,
		version:     eventDoc.Version,
		timestamp:   eventDoc.Timestamp,
		metadata:    eventDoc.Metadata.(map[string]interface{}),
	}, nil
}

func (s *StreamEventStore) updateMetrics(isSave bool, duration time.Duration, err error) {
	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()

	if isSave {
		s.metrics.saveOps++
		s.metrics.totalSaveTime += duration
	} else {
		s.metrics.loadOps++
		s.metrics.totalLoadTime += duration
	}

	if err != nil {
		s.metrics.errors++
	}

	s.metrics.lastOperation = time.Now()
}

func (s *StreamEventStore) checkAndSplitStream(ctx context.Context, streamName string) error {
	// 스트림 크기 확인 (16MB 제한 대응)
	var doc struct {
		EventCount int `bson:"eventCount"`
	}

	pipeline := []bson.M{
		{"$match": bson.M{"streamName": streamName}},
		{"$project": bson.M{"eventCount": bson.M{"$size": "$events"}}},
	}

	cursor, err := s.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		if doc.EventCount > s.maxEvents {
			// 비동기로 스트림 분할 처리
			go s.splitStream(context.Background(), streamName)
		}
	}

	return nil
}

func (s *StreamEventStore) splitStream(ctx context.Context, streamName string) {
	// 스트림 분할 로직 구현
	// 실제 구현에서는 더 정교한 분할 전략이 필요
	fmt.Printf("Stream %s needs splitting (> %d events)\n", streamName, s.maxEvents)
}

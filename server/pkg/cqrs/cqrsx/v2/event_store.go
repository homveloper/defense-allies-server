// document_store.go - 단일 이벤트 방식 구현
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

// DocumentEventStore는 MongoDB 단일 이벤트 방식 저장소입니다
type DocumentEventStore struct {
	collection *mongo.Collection
	client     *mongo.Client
	serializer EventSerializer
	metrics    *documentMetrics
}

type documentMetrics struct {
	mu             sync.RWMutex
	saveOps        int64
	loadOps        int64
	queryOps       int64
	totalSaveTime  time.Duration
	totalLoadTime  time.Duration
	totalQueryTime time.Duration
	errors         int64
	lastOperation  time.Time
}

// SingleEventDocument는 MongoDB에 저장되는 개별 이벤트 문서입니다
type SingleEventDocument struct {
	ID         string      `bson:"_id"`
	StreamName string      `bson:"streamName"`
	StreamType string      `bson:"streamType"`
	string     string      `bson:"aggregateId"`
	EventType  string      `bson:"eventType"`
	Version    int         `bson:"version"`
	Timestamp  time.Time   `bson:"timestamp"`
	Data       interface{} `bson:"data"`
	Metadata   interface{} `bson:"metadata"`
	CreatedAt  time.Time   `bson:"createdAt"`
}

// NewDocumentEventStore는 새로운 문서 이벤트 저장소를 생성합니다
func NewDocumentEventStore(collection *mongo.Collection, client *mongo.Client, serializer EventSerializer) *DocumentEventStore {
	return &DocumentEventStore{
		collection: collection,
		client:     client,
		serializer: serializer,
		metrics:    &documentMetrics{},
	}
}

// Save는 이벤트들을 개별 문서로 저장합니다
func (d *DocumentEventStore) Save(ctx context.Context, events []Event, expectedVersion int) error {
	if len(events) == 0 {
		return nil
	}

	start := time.Now()
	defer func() {
		d.updateMetrics("save", time.Since(start), nil)
	}()

	aggregateID := events[0].string()

	// 트랜잭션 세션 시작
	session, err := d.client.StartSession()
	if err != nil {
		d.updateMetrics("save", time.Since(start), err)
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	// 트랜잭션 실행
	err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return err
		}

		// 현재 버전 확인 (낙관적 동시성 제어)
		currentVersion, err := d.getCurrentVersion(sc, aggregateID)
		if err != nil {
			session.AbortTransaction(sc)
			return err
		}

		if currentVersion != expectedVersion {
			session.AbortTransaction(sc)
			return ErrConcurrencyConflict
		}

		// 이벤트들을 개별 문서로 저장
		docs := make([]interface{}, len(events))
		for i, event := range events {
			docs[i] = SingleEventDocument{
				ID:         fmt.Sprintf("%s-%d", aggregateID.String(), expectedVersion+i+1),
				StreamName: d.getStreamName(aggregateID),
				StreamType: d.getStreamType(event),
				string:     aggregateID.String(),
				EventType:  string(event.EventType()),
				Version:    expectedVersion + i + 1,
				Timestamp:  event.Timestamp(),
				Data:       event.Data(),
				Metadata:   event.Metadata(),
				CreatedAt:  time.Now(),
			}
		}

		_, err = d.collection.InsertMany(sc, docs)
		if err != nil {
			session.AbortTransaction(sc)
			return fmt.Errorf("failed to insert events: %w", err)
		}

		return session.CommitTransaction(sc)
	})

	if err != nil {
		d.updateMetrics("save", time.Since(start), err)
		return err
	}

	return nil
}

// Load는 집합체의 모든 이벤트를 로드합니다
func (d *DocumentEventStore) Load(ctx context.Context, aggregateID uuid.UUID) ([]Event, error) {
	start := time.Now()
	defer func() {
		d.updateMetrics("load", time.Since(start), nil)
	}()

	filter := bson.M{"aggregateId": aggregateID.String()}
	opts := options.Find().SetSort(bson.M{"version": 1})

	cursor, err := d.collection.Find(ctx, filter, opts)
	if err != nil {
		d.updateMetrics("load", time.Since(start), err)
		return nil, fmt.Errorf("failed to find events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []Event
	for cursor.Next(ctx) {
		var doc SingleEventDocument
		if err := cursor.Decode(&doc); err != nil {
			d.updateMetrics("load", time.Since(start), err)
			return nil, fmt.Errorf("failed to decode event: %w", err)
		}

		event := d.documentToEvent(doc, aggregateID)
		events = append(events, event)
	}

	if err := cursor.Err(); err != nil {
		d.updateMetrics("load", time.Since(start), err)
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return events, nil
}

// LoadFrom은 특정 버전부터 이벤트를 로드합니다
func (d *DocumentEventStore) LoadFrom(ctx context.Context, aggregateID uuid.UUID, fromVersion int) ([]Event, error) {
	start := time.Now()
	defer func() {
		d.updateMetrics("load", time.Since(start), nil)
	}()

	filter := bson.M{
		"aggregateId": aggregateID.String(),
		"version":     bson.M{"$gte": fromVersion},
	}
	opts := options.Find().SetSort(bson.M{"version": 1})

	cursor, err := d.collection.Find(ctx, filter, opts)
	if err != nil {
		d.updateMetrics("load", time.Since(start), err)
		return nil, fmt.Errorf("failed to find events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []Event
	for cursor.Next(ctx) {
		var doc SingleEventDocument
		if err := cursor.Decode(&doc); err != nil {
			d.updateMetrics("load", time.Since(start), err)
			return nil, fmt.Errorf("failed to decode event: %w", err)
		}

		event := d.documentToEvent(doc, aggregateID)
		events = append(events, event)
	}

	return events, nil
}

// FindEvents는 복잡한 쿼리로 이벤트를 찾습니다 (DocumentEventStore의 장점)
func (d *DocumentEventStore) FindEvents(ctx context.Context, query EventQuery) ([]Event, error) {
	start := time.Now()
	defer func() {
		d.updateMetrics("query", time.Since(start), nil)
	}()

	filter := d.buildFilter(query)
	opts := d.buildOptions(query)

	cursor, err := d.collection.Find(ctx, filter, opts)
	if err != nil {
		d.updateMetrics("query", time.Since(start), err)
		return nil, fmt.Errorf("failed to find events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []Event
	for cursor.Next(ctx) {
		var doc SingleEventDocument
		if err := cursor.Decode(&doc); err != nil {
			d.updateMetrics("query", time.Since(start), err)
			return nil, fmt.Errorf("failed to decode event: %w", err)
		}

		aggregateID, _ := uuid.Parse(doc.string)
		event := d.documentToEvent(doc, aggregateID)
		events = append(events, event)
	}

	return events, nil
}

// CountEvents는 쿼리 조건에 맞는 이벤트 수를 반환합니다
func (d *DocumentEventStore) CountEvents(ctx context.Context, query EventQuery) (int64, error) {
	start := time.Now()
	defer func() {
		d.updateMetrics("query", time.Since(start), nil)
	}()

	filter := d.buildFilter(query)
	count, err := d.collection.CountDocuments(ctx, filter)
	if err != nil {
		d.updateMetrics("query", time.Since(start), err)
		return 0, fmt.Errorf("failed to count events: %w", err)
	}

	return count, nil
}

// GetMetrics는 성능 메트릭을 반환합니다
func (d *DocumentEventStore) GetMetrics() StoreMetrics {
	d.metrics.mu.RLock()
	defer d.metrics.mu.RUnlock()

	var avgSave, avgLoad time.Duration
	if d.metrics.saveOps > 0 {
		avgSave = d.metrics.totalSaveTime / time.Duration(d.metrics.saveOps)
	}
	if d.metrics.loadOps > 0 {
		avgLoad = d.metrics.totalLoadTime / time.Duration(d.metrics.loadOps)
	}

	return StoreMetrics{
		SaveOperations:  d.metrics.saveOps,
		LoadOperations:  d.metrics.loadOps,
		AverageSaveTime: avgSave,
		AverageLoadTime: avgLoad,
		ErrorCount:      d.metrics.errors,
		LastOperation:   d.metrics.lastOperation,
		StorageStrategy: StrategyDocument,
	}
}

// Close는 연결을 정리합니다
func (d *DocumentEventStore) Close() error {
	return nil
}

// CreateIndexes는 성능 최적화를 위한 인덱스를 생성합니다
func (d *DocumentEventStore) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{"aggregateId", 1},
				{"version", 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{"eventType", 1},
				{"timestamp", -1},
			},
		},
		{
			Keys: bson.D{
				{"streamType", 1},
				{"timestamp", -1},
			},
		},
		{
			Keys: bson.D{
				{"aggregateId", 1},
				{"timestamp", 1},
			},
		},
		{
			Keys: bson.D{{"timestamp", -1}},
		},
	}

	_, err := d.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// Private helper methods

func (d *DocumentEventStore) getStreamName(aggregateID uuid.UUID) string {
	return fmt.Sprintf("guild-%s", aggregateID.String())
}

func (d *DocumentEventStore) getStreamType(event Event) string {
	return "Guild"
}

func (d *DocumentEventStore) getCurrentVersion(ctx context.Context, aggregateID uuid.UUID) (int, error) {
	filter := bson.M{"aggregateId": aggregateID.String()}
	opts := options.FindOne().SetSort(bson.M{"version": -1})

	var doc struct {
		Version int `bson:"version"`
	}

	err := d.collection.FindOne(ctx, filter, opts).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, nil // 새로운 집합체
		}
		return 0, err
	}

	return doc.Version, nil
}

func (d *DocumentEventStore) documentToEvent(doc SingleEventDocument, aggregateID uuid.UUID) Event {
	metadata := make(Metadata)
	if doc.Metadata != nil {
		if m, ok := doc.Metadata.(map[string]interface{}); ok {
			metadata = m
		}
	}

	return &BaseEvent{
		aggregateID: aggregateID,
		eventType:   EventType(doc.EventType),
		data:        doc.Data,
		version:     doc.Version,
		timestamp:   doc.Timestamp,
		metadata:    metadata,
	}
}

func (d *DocumentEventStore) buildFilter(query EventQuery) bson.M {
	filter := bson.M{}

	if len(query.EventTypes) > 0 {
		eventTypes := make([]string, len(query.EventTypes))
		for i, et := range query.EventTypes {
			eventTypes[i] = string(et)
		}
		filter["eventType"] = bson.M{"$in": eventTypes}
	}

	if query.StartTime != nil || query.EndTime != nil {
		timeFilter := bson.M{}
		if query.StartTime != nil {
			timeFilter["$gte"] = *query.StartTime
		}
		if query.EndTime != nil {
			timeFilter["$lte"] = *query.EndTime
		}
		filter["timestamp"] = timeFilter
	}

	if len(query.AggregateIDs) > 0 {
		aggregateIDStrs := make([]string, len(query.AggregateIDs))
		for i, id := range query.AggregateIDs {
			aggregateIDStrs[i] = id.String()
		}
		filter["aggregateId"] = bson.M{"$in": aggregateIDStrs}
	}

	return filter
}

func (d *DocumentEventStore) buildOptions(query EventQuery) *options.FindOptions {
	opts := options.Find()

	// 기본 정렬: 시간순
	opts.SetSort(bson.M{"timestamp": 1})

	if query.Limit > 0 {
		opts.SetLimit(int64(query.Limit))
	}

	if query.Offset > 0 {
		opts.SetSkip(int64(query.Offset))
	}

	return opts
}

func (d *DocumentEventStore) updateMetrics(operation string, duration time.Duration, err error) {
	d.metrics.mu.Lock()
	defer d.metrics.mu.Unlock()

	switch operation {
	case "save":
		d.metrics.saveOps++
		d.metrics.totalSaveTime += duration
	case "load":
		d.metrics.loadOps++
		d.metrics.totalLoadTime += duration
	case "query":
		d.metrics.queryOps++
		d.metrics.totalQueryTime += duration
	}

	if err != nil {
		d.metrics.errors++
	}

	d.metrics.lastOperation = time.Now()
}

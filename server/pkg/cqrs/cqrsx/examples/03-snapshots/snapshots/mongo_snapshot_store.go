package snapshots

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"defense-allies-server/pkg/cqrs/cqrsx"
)

// MongoSnapshotDocument MongoDB 스냅샷 문서 구조
type MongoSnapshotDocument struct {
	ID            string                 `bson:"_id"`
	AggregateID   string                 `bson:"aggregate_id"`
	AggregateType string                 `bson:"aggregate_type"`
	Version       int                    `bson:"version"`
	Data          []byte                 `bson:"data"`
	ContentType   string                 `bson:"content_type"`
	Compression   string                 `bson:"compression"`
	Size          int64                  `bson:"size"`
	Timestamp     time.Time              `bson:"timestamp"`
	Metadata      map[string]interface{} `bson:"metadata"`
}

// MongoSnapshot MongoDB 스냅샷 구현
type MongoSnapshot struct {
	aggregateID   string
	aggregateType string
	version       int
	data          []byte
	contentType   string
	compression   string
	timestamp     time.Time
	metadata      map[string]interface{}
}

func (s *MongoSnapshot) AggregateID() string              { return s.aggregateID }
func (s *MongoSnapshot) AggregateType() string            { return s.aggregateType }
func (s *MongoSnapshot) Version() int                     { return s.version }
func (s *MongoSnapshot) Data() []byte                     { return s.data }
func (s *MongoSnapshot) Timestamp() time.Time             { return s.timestamp }
func (s *MongoSnapshot) Metadata() map[string]interface{} { return s.metadata }

// MongoSnapshotStore MongoDB 기반 스냅샷 저장소
type MongoSnapshotStore struct {
	client         *cqrsx.MongoClientManager
	collectionName string
}

// NewMongoSnapshotStore MongoDB 스냅샷 저장소 생성
func NewMongoSnapshotStore(client *cqrsx.MongoClientManager, collectionName string) *MongoSnapshotStore {
	if collectionName == "" {
		collectionName = "snapshots"
	}

	return &MongoSnapshotStore{
		client:         client,
		collectionName: collectionName,
	}
}

// SaveSnapshot 스냅샷 저장
func (s *MongoSnapshotStore) SaveSnapshot(ctx context.Context, snapshot Snapshot) error {
	collection := s.client.GetCollection(s.collectionName)

	return s.client.ExecuteCommand(ctx, func() error {
		// 문서 ID 생성 (aggregate_id + version)
		docID := fmt.Sprintf("%s_%d", snapshot.AggregateID(), snapshot.Version())

		doc := MongoSnapshotDocument{
			ID:            docID,
			AggregateID:   snapshot.AggregateID(),
			AggregateType: snapshot.AggregateType(),
			Version:       snapshot.Version(),
			Data:          snapshot.Data(),
			Size:          int64(len(snapshot.Data())),
			Timestamp:     snapshot.Timestamp(),
			Metadata:      snapshot.Metadata(),
		}

		// ContentType과 Compression 정보 추출
		if metadata := snapshot.Metadata(); metadata != nil {
			if ct, ok := metadata["content_type"].(string); ok {
				doc.ContentType = ct
			}
			if comp, ok := metadata["compression"].(string); ok {
				doc.Compression = comp
			}
		}

		// Upsert 옵션으로 저장 (같은 버전이면 덮어쓰기)
		opts := options.Replace().SetUpsert(true)
		filter := bson.M{"_id": docID}

		_, err := collection.ReplaceOne(ctx, filter, doc, opts)
		if err != nil {
			return fmt.Errorf("failed to save snapshot: %w", err)
		}

		return nil
	})
}

// GetSnapshot 최신 스냅샷 조회 (maxVersion 이하의 가장 최신)
func (s *MongoSnapshotStore) GetSnapshot(ctx context.Context, aggregateID string, maxVersion int) (Snapshot, error) {
	collection := s.client.GetCollection(s.collectionName)
	var snapshot Snapshot

	err := s.client.ExecuteCommand(ctx, func() error {
		// maxVersion 이하의 가장 최신 스냅샷 조회
		filter := bson.M{
			"aggregate_id": aggregateID,
			"version":      bson.M{"$lte": maxVersion},
		}

		opts := options.FindOne().SetSort(bson.D{{Key: "version", Value: -1}})

		var doc MongoSnapshotDocument
		err := collection.FindOne(ctx, filter, opts).Decode(&doc)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return &SnapshotError{
					Code:      ErrCodeSnapshotNotFound,
					Message:   "no snapshot found",
					Operation: "GetSnapshot",
				}
			}
			return fmt.Errorf("failed to get snapshot: %w", err)
		}

		// 메타데이터에 ContentType과 Compression 추가
		metadata := doc.Metadata
		if metadata == nil {
			metadata = make(map[string]interface{})
		}
		metadata["content_type"] = doc.ContentType
		metadata["compression"] = doc.Compression
		metadata["size"] = doc.Size

		snapshot = &MongoSnapshot{
			aggregateID:   doc.AggregateID,
			aggregateType: doc.AggregateType,
			version:       doc.Version,
			data:          doc.Data,
			contentType:   doc.ContentType,
			compression:   doc.Compression,
			timestamp:     doc.Timestamp,
			metadata:      metadata,
		}

		return nil
	})

	return snapshot, err
}

// GetSnapshotByVersion 특정 버전의 스냅샷 조회
func (s *MongoSnapshotStore) GetSnapshotByVersion(ctx context.Context, aggregateID string, version int) (Snapshot, error) {
	collection := s.client.GetCollection(s.collectionName)
	var snapshot Snapshot

	err := s.client.ExecuteCommand(ctx, func() error {
		docID := fmt.Sprintf("%s_%d", aggregateID, version)
		filter := bson.M{"_id": docID}

		var doc MongoSnapshotDocument
		err := collection.FindOne(ctx, filter).Decode(&doc)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return &SnapshotError{
					Code:      ErrCodeSnapshotNotFound,
					Message:   fmt.Sprintf("snapshot not found for version %d", version),
					Operation: "GetSnapshotByVersion",
				}
			}
			return fmt.Errorf("failed to get snapshot by version: %w", err)
		}

		// 메타데이터에 ContentType과 Compression 추가
		metadata := doc.Metadata
		if metadata == nil {
			metadata = make(map[string]interface{})
		}
		metadata["content_type"] = doc.ContentType
		metadata["compression"] = doc.Compression
		metadata["size"] = doc.Size

		snapshot = &MongoSnapshot{
			aggregateID:   doc.AggregateID,
			aggregateType: doc.AggregateType,
			version:       doc.Version,
			data:          doc.Data,
			contentType:   doc.ContentType,
			compression:   doc.Compression,
			timestamp:     doc.Timestamp,
			metadata:      metadata,
		}

		return nil
	})

	return snapshot, err
}

// DeleteSnapshot 특정 버전의 스냅샷 삭제
func (s *MongoSnapshotStore) DeleteSnapshot(ctx context.Context, aggregateID string, version int) error {
	collection := s.client.GetCollection(s.collectionName)

	return s.client.ExecuteCommand(ctx, func() error {
		docID := fmt.Sprintf("%s_%d", aggregateID, version)
		filter := bson.M{"_id": docID}

		result, err := collection.DeleteOne(ctx, filter)
		if err != nil {
			return fmt.Errorf("failed to delete snapshot: %w", err)
		}

		if result.DeletedCount == 0 {
			return &SnapshotError{
				Code:      ErrCodeSnapshotNotFound,
				Message:   fmt.Sprintf("snapshot not found for version %d", version),
				Operation: "DeleteSnapshot",
			}
		}

		return nil
	})
}

// DeleteOldSnapshots 오래된 스냅샷들 삭제 (keepCount개만 유지)
func (s *MongoSnapshotStore) DeleteOldSnapshots(ctx context.Context, aggregateID string, keepCount int) error {
	if keepCount <= 0 {
		return fmt.Errorf("keepCount must be positive")
	}

	collection := s.client.GetCollection(s.collectionName)

	return s.client.ExecuteCommand(ctx, func() error {
		// 최신 keepCount개를 제외한 나머지 조회
		filter := bson.M{"aggregate_id": aggregateID}
		opts := options.Find().
			SetSort(bson.D{{Key: "version", Value: -1}}).
			SetSkip(int64(keepCount))

		cursor, err := collection.Find(ctx, filter, opts)
		if err != nil {
			return fmt.Errorf("failed to find old snapshots: %w", err)
		}
		defer cursor.Close(ctx)

		var idsToDelete []string
		for cursor.Next(ctx) {
			var doc MongoSnapshotDocument
			if err := cursor.Decode(&doc); err != nil {
				continue
			}
			idsToDelete = append(idsToDelete, doc.ID)
		}

		if len(idsToDelete) > 0 {
			deleteFilter := bson.M{"_id": bson.M{"$in": idsToDelete}}
			_, err := collection.DeleteMany(ctx, deleteFilter)
			if err != nil {
				return fmt.Errorf("failed to delete old snapshots: %w", err)
			}
		}

		return nil
	})
}

// ListSnapshots 스냅샷 목록 조회
func (s *MongoSnapshotStore) ListSnapshots(ctx context.Context, aggregateID string) ([]Snapshot, error) {
	collection := s.client.GetCollection(s.collectionName)
	var snapshots []Snapshot

	err := s.client.ExecuteCommand(ctx, func() error {
		filter := bson.M{"aggregate_id": aggregateID}
		opts := options.Find().SetSort(bson.D{{Key: "version", Value: 1}})

		cursor, err := collection.Find(ctx, filter, opts)
		if err != nil {
			return fmt.Errorf("failed to list snapshots: %w", err)
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var doc MongoSnapshotDocument
			if err := cursor.Decode(&doc); err != nil {
				continue
			}

			// 메타데이터에 ContentType과 Compression 추가
			metadata := doc.Metadata
			if metadata == nil {
				metadata = make(map[string]interface{})
			}
			metadata["content_type"] = doc.ContentType
			metadata["compression"] = doc.Compression
			metadata["size"] = doc.Size

			snapshot := &MongoSnapshot{
				aggregateID:   doc.AggregateID,
				aggregateType: doc.AggregateType,
				version:       doc.Version,
				data:          doc.Data,
				contentType:   doc.ContentType,
				compression:   doc.Compression,
				timestamp:     doc.Timestamp,
				metadata:      metadata,
			}

			snapshots = append(snapshots, snapshot)
		}

		return cursor.Err()
	})

	return snapshots, err
}

// GetSnapshotStats 스냅샷 통계 조회
func (s *MongoSnapshotStore) GetSnapshotStats(ctx context.Context) (map[string]interface{}, error) {
	collection := s.client.GetCollection(s.collectionName)
	stats := make(map[string]interface{})

	err := s.client.ExecuteCommand(ctx, func() error {
		// 전체 스냅샷 개수
		totalCount, err := collection.CountDocuments(ctx, bson.M{})
		if err != nil {
			return fmt.Errorf("failed to count snapshots: %w", err)
		}
		stats["total_snapshots"] = totalCount

		// 집계 파이프라인으로 통계 계산
		pipeline := []bson.M{
			{
				"$group": bson.M{
					"_id":        nil,
					"total_size": bson.M{"$sum": "$size"},
					"avg_size":   bson.M{"$avg": "$size"},
					"min_size":   bson.M{"$min": "$size"},
					"max_size":   bson.M{"$max": "$size"},
					"oldest":     bson.M{"$min": "$timestamp"},
					"newest":     bson.M{"$max": "$timestamp"},
				},
			},
		}

		cursor, err := collection.Aggregate(ctx, pipeline)
		if err != nil {
			return fmt.Errorf("failed to aggregate snapshot stats: %w", err)
		}
		defer cursor.Close(ctx)

		if cursor.Next(ctx) {
			var result bson.M
			if err := cursor.Decode(&result); err == nil {
				stats["total_size"] = result["total_size"]
				stats["average_size"] = result["avg_size"]
				stats["min_size"] = result["min_size"]
				stats["max_size"] = result["max_size"]
				stats["oldest_snapshot"] = result["oldest"]
				stats["newest_snapshot"] = result["newest"]
			}
		}

		return nil
	})

	return stats, err
}

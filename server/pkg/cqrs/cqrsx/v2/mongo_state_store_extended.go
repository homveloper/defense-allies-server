// mongo_state_store_extended_updated.go - 새로운 구조체를 사용하는 확장 기능
package cqrsx

// import (
// 	"context"
// 	"fmt"
// 	"time"

// 	"github.com/google/uuid"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// // Query는 조건에 맞는 상태들을 조회합니다 (QueryableStateStore 구현)
// func (m *MongoStateStore) Query(ctx context.Context, query StateQuery) ([]*AggregateState, error) {
// 	start := time.Now()
// 	defer func() {
// 		m.updateMetrics("query", time.Since(start), nil)
// 	}()

// 	filter := m.buildQueryFilter(query)
// 	opts := m.buildQueryOptions(query)

// 	cursor, err := m.collection.Find(ctx, filter, opts)
// 	if err != nil {
// 		m.updateMetrics("query", time.Since(start), err)
// 		return nil, fmt.Errorf("failed to execute query: %w", err)
// 	}
// 	defer cursor.Close(ctx)

// 	var states []*AggregateState
// 	for cursor.Next(ctx) {
// 		var doc mongoStateDocument
// 		if err := cursor.Decode(&doc); err != nil {
// 			m.updateMetrics("query", time.Since(start), err)
// 			return nil, fmt.Errorf("failed to decode query result: %w", err)
// 		}

// 		state, err := m.documentToState(&doc)
// 		if err != nil {
// 			m.updateMetrics("query", time.Since(start), err)
// 			return nil, fmt.Errorf("failed to convert document to state: %w", err)
// 		}

// 		states = append(states, state)
// 	}

// 	if err := cursor.Err(); err != nil {
// 		m.updateMetrics("query", time.Since(start), err)
// 		return nil, fmt.Errorf("cursor error: %w", err)
// 	}

// 	return states, nil
// }

// // CountByQuery는 쿼리 조건에 맞는 상태 개수를 반환합니다
// func (m *MongoStateStore) CountByQuery(ctx context.Context, query StateQuery) (int64, error) {
// 	filter := m.buildQueryFilter(query)
// 	count, err := m.collection.CountDocuments(ctx, filter)
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to count by query: %w", err)
// 	}
// 	return count, nil
// }

// // GetAggregateTypes는 저장된 모든 집합체 타입을 반환합니다
// func (m *MongoStateStore) GetAggregateTypes(ctx context.Context) ([]string, error) {
// 	pipeline := []bson.M{
// 		{"$group": bson.M{"_id": "$aggregateType"}},
// 		{"$sort": bson.M{"_id": 1}},
// 	}

// 	cursor, err := m.collection.Aggregate(ctx, pipeline)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get aggregate types: %w", err)
// 	}
// 	defer cursor.Close(ctx)

// 	var types []string
// 	for cursor.Next(ctx) {
// 		var result struct {
// 			ID string `bson:"_id"`
// 		}
// 		if err := cursor.Decode(&result); err != nil {
// 			return nil, fmt.Errorf("failed to decode aggregate type: %w", err)
// 		}
// 		types = append(types, result.ID)
// 	}

// 	return types, nil
// }

// // GetVersions는 집합체의 모든 버전을 반환합니다
// func (m *MongoStateStore) GetVersions(ctx context.Context, aggregateID uuid.UUID) ([]int, error) {
// 	filter := bson.M{"aggregateId": aggregateID.String()}
// 	opts := options.Find().SetProjection(bson.M{"version": 1}).SetSort(bson.M{"version": 1})

// 	cursor, err := m.collection.Find(ctx, filter, opts)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get versions: %w", err)
// 	}
// 	defer cursor.Close(ctx)

// 	var versions []int
// 	for cursor.Next(ctx) {
// 		var doc struct {
// 			Version int `bson:"version"`
// 		}
// 		if err := cursor.Decode(&doc); err != nil {
// 			return nil, fmt.Errorf("failed to decode version: %w", err)
// 		}
// 		versions = append(versions, doc.Version)
// 	}

// 	return versions, nil
// }

// // GetMetrics는 저장소 메트릭을 반환합니다 (MetricsStateStore 구현)
// func (m *MongoStateStore) GetMetrics(ctx context.Context) (*StateMetrics, error) {
// 	pipeline := []bson.M{
// 		{
// 			"$group": bson.M{
// 				"_id":             nil,
// 				"totalStates":     bson.M{"$sum": 1},
// 				"totalStorage":    bson.M{"$sum": "$size.stored"},
// 				"avgSize":         bson.M{"$avg": "$size.stored"},
// 				"maxSize":         bson.M{"$max": "$size.stored"},
// 				"minSize":         bson.M{"$min": "$size.stored"},
// 				"oldestTimestamp": bson.M{"$min": "$stateTimestamp"},
// 				"newestTimestamp": bson.M{"$max": "$stateTimestamp"},
// 			},
// 		},
// 	}

// 	cursor, err := m.collection.Aggregate(ctx, pipeline)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get metrics: %w", err)
// 	}
// 	defer cursor.Close(ctx)

// 	var result struct {
// 		TotalStates     int64     `bson:"totalStates"`
// 		TotalStorage    int64     `bson:"totalStorage"`
// 		AvgSize         int64     `bson:"avgSize"`
// 		MaxSize         int64     `bson:"maxSize"`
// 		MinSize         int64     `bson:"minSize"`
// 		OldestTimestamp time.Time `bson:"oldestTimestamp"`
// 		NewestTimestamp time.Time `bson:"newestTimestamp"`
// 	}

// 	if cursor.Next(ctx) {
// 		if err := cursor.Decode(&result); err != nil {
// 			return nil, fmt.Errorf("failed to decode metrics: %w", err)
// 		}
// 	}

// 	// 집합체 타입들 조회
// 	aggregateTypes, err := m.GetAggregateTypes(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get aggregate types for metrics: %w", err)
// 	}

// 	return &StateMetrics{
// 		TotalStates:       result.TotalStates,
// 		TotalStorageBytes: result.TotalStorage,
// 		AverageSize:       result.AvgSize,
// 		MaxSize:           result.MaxSize,
// 		MinSize:           result.MinSize,
// 		AggregateTypes:    aggregateTypes,
// 		OldestState:       result.OldestTimestamp,
// 		NewestState:       result.NewestTimestamp,
// 	}, nil
// }

// // GetAggregateMetrics는 특정 집합체의 메트릭을 반환합니다
// func (m *MongoStateStore) GetAggregateMetrics(ctx context.Context, aggregateID uuid.UUID) (*StateMetrics, error) {
// 	pipeline := []bson.M{
// 		{"$match": bson.M{"aggregateId": aggregateID.String()}},
// 		{
// 			"$group": bson.M{
// 				"_id":             nil,
// 				"totalStates":     bson.M{"$sum": 1},
// 				"totalStorage":    bson.M{"$sum": "$size.stored"},
// 				"avgSize":         bson.M{"$avg": "$size.stored"},
// 				"maxSize":         bson.M{"$max": "$size.stored"},
// 				"minSize":         bson.M{"$min": "$size.stored"},
// 				"oldestTimestamp": bson.M{"$min": "$stateTimestamp"},
// 				"newestTimestamp": bson.M{"$max": "$stateTimestamp"},
// 				"aggregateType":   bson.M{"$first": "$aggregateType"},
// 			},
// 		},
// 	}

// 	cursor, err := m.collection.Aggregate(ctx, pipeline)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get aggregate metrics: %w", err)
// 	}
// 	defer cursor.Close(ctx)

// 	var result struct {
// 		TotalStates     int64     `bson:"totalStates"`
// 		TotalStorage    int64     `bson:"totalStorage"`
// 		AvgSize         int64     `bson:"avgSize"`
// 		MaxSize         int64     `bson:"maxSize"`
// 		MinSize         int64     `bson:"minSize"`
// 		OldestTimestamp time.Time `bson:"oldestTimestamp"`
// 		NewestTimestamp time.Time `bson:"newestTimestamp"`
// 		AggregateType   string    `bson:"aggregateType"`
// 	}

// 	if cursor.Next(ctx) {
// 		if err := cursor.Decode(&result); err != nil {
// 			return nil, fmt.Errorf("failed to decode aggregate metrics: %w", err)
// 		}
// 	} else {
// 		return nil, ErrStateNotFound
// 	}

// 	return &StateMetrics{
// 		TotalStates:       result.TotalStates,
// 		TotalStorageBytes: result.TotalStorage,
// 		AverageSize:       result.AvgSize,
// 		MaxSize:           result.MaxSize,
// 		MinSize:           result.MinSize,
// 		AggregateTypes:    []string{result.AggregateType},
// 		OldestState:       result.OldestTimestamp,
// 		NewestState:       result.NewestTimestamp,
// 	}, nil
// }

// // CreateIndexes는 성능 최적화를 위한 인덱스를 생성합니다
// func (m *MongoStateStore) CreateIndexes(ctx context.Context) error {
// 	indexes := []mongo.IndexModel{
// 		{
// 			Keys: bson.D{
// 				{"aggregateId", 1},
// 				{"version", -1},
// 			},
// 			Options: options.Index().SetUnique(true).SetName("aggregateId_version_unique"),
// 		},
// 		{
// 			Keys: bson.D{
// 				{"aggregateType", 1},
// 				{"stateTimestamp", -1},
// 			},
// 			Options: options.Index().SetName("aggregateType_stateTimestamp"),
// 		},
// 		{
// 			Keys: bson.D{
// 				{"stateTimestamp", -1},
// 			},
// 			Options: options.Index().SetName("stateTimestamp_desc"),
// 		},
// 		{
// 			Keys: bson.D{
// 				{"size.stored", -1},
// 			},
// 			Options: options.Index().SetName("size_stored_desc"),
// 		},
// 		{
// 			Keys: bson.D{
// 				{"aggregateId", 1},
// 				{"stateTimestamp", 1},
// 			},
// 			Options: options.Index().SetName("aggregateId_stateTimestamp"),
// 		},
// 		{
// 			Keys: bson.D{
// 				{"compression.type", 1},
// 				{"encryption.type", 1},
// 			},
// 			Options: options.Index().SetName("compression_encryption_types"),
// 		},
// 		{
// 			Keys: bson.D{
// 				{"tags", 1},
// 			},
// 			Options: options.Index().SetName("tags_multikey"),
// 		},
// 		{
// 			Keys: bson.D{
// 				{"createdAt", -1},
// 			},
// 			Options: options.Index().SetName("createdAt_desc"),
// 		},
// 		{
// 			Keys: bson.D{
// 				{"updatedAt", -1},
// 			},
// 			Options: options.Index().SetName("updatedAt_desc"),
// 		},
// 		{
// 			Keys: bson.D{
// 				{"ttl", 1},
// 			},
// 			Options: options.Index().SetExpireAfterSeconds(0).SetName("ttl_auto_expire"),
// 		},
// 		{
// 			Keys: bson.D{
// 				{"lastAccessedAt", -1},
// 			},
// 			Options: options.Index().SetName("lastAccessedAt_desc"),
// 		},
// 	}

// 	_, err := m.collection.Indexes().CreateMany(ctx, indexes)
// 	if err != nil {
// 		return fmt.Errorf("failed to create state store indexes: %w", err)
// 	}

// 	return nil
// }

// // GetStoreMetrics는 MongoDB 저장소 자체의 메트릭을 반환합니다
// func (m *MongoStateStore) GetStoreMetrics() *MongoStoreMetrics {
// 	m.metrics.mu.RLock()
// 	defer m.metrics.mu.RUnlock()

// 	var avgSave, avgLoad, avgDelete, avgQuery time.Duration
// 	if m.metrics.saveOps > 0 {
// 		avgSave = m.metrics.totalSaveTime / time.Duration(m.metrics.saveOps)
// 	}
// 	if m.metrics.loadOps > 0 {
// 		avgLoad = m.metrics.totalLoadTime / time.Duration(m.metrics.loadOps)
// 	}
// 	if m.metrics.deleteOps > 0 {
// 		avgDelete = m.metrics.totalDeleteTime / time.Duration(m.metrics.deleteOps)
// 	}
// 	if m.metrics.queryOps > 0 {
// 		avgQuery = m.metrics.totalQueryTime / time.Duration(m.metrics.queryOps)
// 	}

// 	return &MongoStoreMetrics{
// 		SaveOperations:    m.metrics.saveOps,
// 		LoadOperations:    m.metrics.loadOps,
// 		DeleteOperations:  m.metrics.deleteOps,
// 		QueryOperations:   m.metrics.queryOps,
// 		AverageSaveTime:   avgSave,
// 		AverageLoadTime:   avgLoad,
// 		AverageDeleteTime: avgDelete,
// 		AverageQueryTime:  avgQuery,
// 		ErrorCount:        m.metrics.errors,
// 		LastOperation:     m.metrics.lastOperation,
// 		CompressionSaved:  m.metrics.compressionSaved,
// 		EncryptionApplied: m.metrics.encryptionApplied,
// 	}
// }

// // MongoStoreMetrics는 MongoDB 저장소의 성능 메트릭입니다
// type MongoStoreMetrics struct {
// 	SaveOperations    int64         `json:"saveOperations"`
// 	LoadOperations    int64         `json:"loadOperations"`
// 	DeleteOperations  int64         `json:"deleteOperations"`
// 	QueryOperations   int64         `json:"queryOperations"`
// 	AverageSaveTime   time.Duration `json:"averageSaveTime"`
// 	AverageLoadTime   time.Duration `json:"averageLoadTime"`
// 	AverageDeleteTime time.Duration `json:"averageDeleteTime"`
// 	AverageQueryTime  time.Duration `json:"averageQueryTime"`
// 	ErrorCount        int64         `json:"errorCount"`
// 	LastOperation     time.Time     `json:"lastOperation"`
// 	CompressionSaved  int64         `json:"compressionSaved"`
// 	EncryptionApplied int64         `json:"encryptionApplied"`
// }

// // buildQueryFilter는 StateQuery를 MongoDB 필터로 변환합니다
// func (m *MongoStateStore) buildQueryFilter(query StateQuery) bson.M {
// 	filter := bson.M{}

// 	if len(query.AggregateIDs) > 0 {
// 		aggregateIDStrs := make([]string, len(query.AggregateIDs))
// 		for i, id := range query.AggregateIDs {
// 			aggregateIDStrs[i] = id.String()
// 		}
// 		filter["aggregateId"] = bson.M{"$in": aggregateIDStrs}
// 	}

// 	if query.AggregateType != "" {
// 		filter["aggregateType"] = query.AggregateType
// 	}

// 	if query.MinVersion != nil || query.MaxVersion != nil {
// 		versionFilter := bson.M{}
// 		if query.MinVersion != nil {
// 			versionFilter["$gte"] = *query.MinVersion
// 		}
// 		if query.MaxVersion != nil {
// 			versionFilter["$lte"] = *query.MaxVersion
// 		}
// 		filter["version"] = versionFilter
// 	}

// 	if query.StartTime != nil || query.EndTime != nil {
// 		timeFilter := bson.M{}
// 		if query.StartTime != nil {
// 			timeFilter["$gte"] = *query.StartTime
// 		}
// 		if query.EndTime != nil {
// 			timeFilter["$lte"] = *query.EndTime
// 		}
// 		filter["stateTimestamp"] = timeFilter
// 	}

// 	return filter
// }

// // buildQueryOptions는 StateQuery를 MongoDB 옵션으로 변환합니다
// func (m *MongoStateStore) buildQueryOptions(query StateQuery) *options.FindOptions {
// 	opts := options.Find()

// 	// 기본 정렬: 시간순 (최신순)
// 	opts.SetSort(bson.M{"stateTimestamp": -1})

// 	if query.Limit > 0 {
// 		opts.SetLimit(int64(query.Limit))
// 	}

// 	if query.Offset > 0 {
// 		opts.SetSkip(int64(query.Offset))
// 	}

// 	return opts
// }

// // GetCompressionStats는 압축 통계를 반환합니다
// func (m *MongoStateStore) GetCompressionStats(ctx context.Context) (map[string]interface{}, error) {
// 	pipeline := []bson.M{
// 		{
// 			"$group": bson.M{
// 				"_id":                 "$compression.type",
// 				"count":               bson.M{"$sum": 1},
// 				"totalOriginal":       bson.M{"$sum": "$size.original"},
// 				"totalStored":         bson.M{"$sum": "$size.stored"},
// 				"avgCompressionRatio": bson.M{"$avg": "$size.compressionRatio"},
// 			},
// 		},
// 		{
// 			"$sort": bson.M{"count": -1},
// 		},
// 	}

// 	cursor, err := m.collection.Aggregate(ctx, pipeline)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get compression stats: %w", err)
// 	}
// 	defer cursor.Close(ctx)

// 	var stats []map[string]interface{}
// 	for cursor.Next(ctx) {
// 		var result map[string]interface{}
// 		if err := cursor.Decode(&result); err != nil {
// 			return nil, fmt.Errorf("failed to decode compression stats: %w", err)
// 		}
// 		stats = append(stats, result)
// 	}

// 	return map[string]interface{}{
// 		"compressionStats": stats,
// 	}, nil
// }

// // GetEncryptionStats는 암호화 통계를 반환합니다
// func (m *MongoStateStore) GetEncryptionStats(ctx context.Context) (map[string]interface{}, error) {
// 	pipeline := []bson.M{
// 		{
// 			"$group": bson.M{
// 				"_id":        "$encryption.type",
// 				"count":      bson.M{"$sum": 1},
// 				"avgKeySize": bson.M{"$avg": "$encryption.keySize"},
// 			},
// 		},
// 		{
// 			"$sort": bson.M{"count": -1},
// 		},
// 	}

// 	cursor, err := m.collection.Aggregate(ctx, pipeline)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get encryption stats: %w", err)
// 	}
// 	defer cursor.Close(ctx)

// 	var stats []map[string]interface{}
// 	for cursor.Next(ctx) {
// 		var result map[string]interface{}
// 		if err := cursor.Decode(&result); err != nil {
// 			return nil, fmt.Errorf("failed to decode encryption stats: %w", err)
// 		}
// 		stats = append(stats, result)
// 	}

// 	return map[string]interface{}{
// 		"encryptionStats": stats,
// 	}, nil
// }

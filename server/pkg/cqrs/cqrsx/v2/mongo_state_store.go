// mongo_state_store_updated.go - 새로운 구조체를 사용하는 MongoDB 상태 저장소
package cqrsx

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoStateStore는 MongoDB 기반 상태 저장소입니다
type MongoStateStore struct {
	collection *mongo.Collection
	client     *mongo.Client
	config     *StateStoreConfig
	metrics    *mongoStateMetrics
	mu         sync.RWMutex
}

// mongoStateMetrics는 MongoDB 저장소의 메트릭을 추적합니다
type mongoStateMetrics struct {
	mu                sync.RWMutex
	saveOps           int64
	loadOps           int64
	deleteOps         int64
	queryOps          int64
	totalSaveTime     time.Duration
	totalLoadTime     time.Duration
	totalDeleteTime   time.Duration
	totalQueryTime    time.Duration
	errors            int64
	lastOperation     time.Time
	compressionSaved  int64 // 압축으로 절약된 바이트
	encryptionApplied int64 // 암호화 적용된 상태 수
}

// NewMongoStateStore는 새로운 MongoDB 상태 저장소를 생성합니다
func NewMongoStateStore(collection *mongo.Collection, client *mongo.Client, options ...StateStoreOption) *MongoStateStore {
	config := NewDefaultStateStoreConfig()
	for _, option := range options {
		option(config)
	}

	store := &MongoStateStore{
		collection: collection,
		client:     client,
		config:     config,
		metrics:    &mongoStateMetrics{},
	}

	// 인덱스 생성 (백그라운드에서)
	if config.IndexingEnabled {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := store.CreateIndexes(ctx); err != nil {
				// 로그 남기기 (실제 로거 사용 권장)
				fmt.Printf("Failed to create indexes: %v\n", err)
			}
		}()
	}

	return store
}

// Save는 집합체 상태를 저장합니다
func (m *MongoStateStore) Save(ctx context.Context, state *AggregateState) error {
	start := time.Now()
	defer func() {
		m.updateMetrics("save", time.Since(start), nil)
	}()

	if err := state.Validate(); err != nil {
		m.updateMetrics("save", time.Since(start), err)
		return fmt.Errorf("invalid state: %w", err)
	}

	// MongoDB 문서 생성
	doc := FromAggregateState(state)

	// 데이터 처리 (압축/암호화)
	processedData, err := m.processData(state.Data, doc)
	if err != nil {
		m.updateMetrics("save", time.Since(start), err)
		return fmt.Errorf("failed to process data: %w", err)
	}

	// 처리된 데이터와 크기 정보 업데이트
	doc.Data = processedData
	doc.RecordProcessingTime(time.Since(start))

	// 중복 키 에러 방지를 위한 upsert 사용
	opts := options.Replace().SetUpsert(true)
	filter := bson.M{"_id": doc.ID}

	_, err = m.collection.ReplaceOne(ctx, filter, doc, opts)
	if err != nil {
		doc.SetLastError(err)
		m.updateMetrics("save", time.Since(start), err)
		return fmt.Errorf("failed to save state: %w", err)
	}

	// 메트릭 업데이트
	if doc.IsEncryptionEnabled() {
		m.metrics.mu.Lock()
		m.metrics.encryptionApplied++
		m.metrics.mu.Unlock()
	}

	// 보존 정책 적용 (백그라운드에서)
	if m.config.RetentionPolicy != nil {
		go m.applyRetentionPolicy(context.Background(), state.string)
	}

	return nil
}

// Load는 집합체의 최신 상태를 로드합니다
func (m *MongoStateStore) Load(ctx context.Context, aggregateID uuid.UUID) (*AggregateState, error) {
	start := time.Now()
	defer func() {
		m.updateMetrics("load", time.Since(start), nil)
	}()

	filter := bson.M{"aggregateId": aggregateID}
	opts := options.FindOne().SetSort(bson.M{"version": -1})

	var doc mongoStateDocument
	err := m.collection.FindOne(ctx, filter, opts).Decode(&doc)
	if err != nil {
		m.updateMetrics("load", time.Since(start), err)
		if err == mongo.ErrNoDocuments {
			return nil, ErrStateNotFound
		}
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	// 액세스 시간 업데이트
	doc.TouchAccessTime()
	m.collection.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"lastAccessedAt": doc.LastAccessedAt}})

	// 문서를 AggregateState로 변환
	state, err := m.documentToState(&doc)
	if err != nil {
		m.updateMetrics("load", time.Since(start), err)
		return nil, fmt.Errorf("failed to convert document to state: %w", err)
	}

	return state, nil
}

// LoadVersion은 특정 버전의 상태를 로드합니다
func (m *MongoStateStore) LoadVersion(ctx context.Context, aggregateID uuid.UUID, version int) (*AggregateState, error) {
	start := time.Now()
	defer func() {
		m.updateMetrics("load", time.Since(start), nil)
	}()

	filter := bson.M{
		"aggregateId": aggregateID,
		"version":     version,
	}

	var doc mongoStateDocument
	err := m.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		m.updateMetrics("load", time.Since(start), err)
		if err == mongo.ErrNoDocuments {
			return nil, ErrStateNotFound
		}
		return nil, fmt.Errorf("failed to load state version %d: %w", version, err)
	}

	// 액세스 시간 업데이트
	doc.TouchAccessTime()
	m.collection.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"lastAccessedAt": doc.LastAccessedAt}})

	// 문서를 AggregateState로 변환
	state, err := m.documentToState(&doc)
	if err != nil {
		m.updateMetrics("load", time.Since(start), err)
		return nil, fmt.Errorf("failed to convert document to state: %w", err)
	}

	return state, nil
}

// processData는 데이터를 압축/암호화 처리합니다
func (m *MongoStateStore) processData(data []byte, doc *mongoStateDocument) (interface{}, error) {
	originalSize := int64(len(data))
	processed := data

	// 압축 적용
	if m.config.CompressionEnabled && len(data) >= m.config.CompressionMinSize {
		var err error
		processed, err = m.compress(processed)
		if err != nil {
			return nil, fmt.Errorf("compression failed: %w", err)
		}

		// 압축 정보 설정
		compressionMetadata := map[string]interface{}{
			"originalSize": originalSize,
			"algorithm":    string(m.config.CompressionType),
		}

		doc.SetCompressionInfo(
			string(m.config.CompressionType),
			"1.0",
			6, // 기본 압축 레벨
			compressionMetadata,
		)
		doc.DataEncoding = "base64"
	}

	// 암호화 적용
	if m.config.EncryptionEnabled && m.config.Encryptor != nil {
		encryptedData, err := m.config.Encryptor.Encrypt(string(processed))
		if err != nil {
			return nil, fmt.Errorf("encryption failed: %w", err)
		}
		processed = encryptedData

		// 암호화 정보 설정
		encryptionMetadata := map[string]interface{}{
			"algorithm":     "AES-GCM",
			"keyDerivation": "PBKDF2",
		}

		doc.SetEncryptionInfo(
			"aes-gcm",
			"1.0",
			256, // 256비트 키
			encryptionMetadata,
		)
	}

	return primitive.Binary{Subtype: 0x00, Data: processed}, nil
}

// documentToState는 MongoDB 문서를 AggregateState로 변환합니다
func (m *MongoStateStore) documentToState(doc *mongoStateDocument) (*AggregateState, error) {
	// 문서 검증
	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document: %w", err)
	}

	// 데이터 처리 (복호화/압축 해제)
	var data []byte
	switch d := doc.Data.(type) {
	case primitive.Binary:
		data = d.Data
	case []byte:
		data = d
	case string:
		data = []byte(d)
	default:
		return nil, fmt.Errorf("unsupported data type: %T", doc.Data)
	}

	// 복호화
	if doc.IsEncryptionEnabled() && m.config.EncryptionEnabled && m.config.Encryptor != nil {
		decrypted, err := m.config.Encryptor.Decrypt(string(data))
		if err != nil {
			return nil, fmt.Errorf("decryption failed: %w", err)
		}
		data = []byte(decrypted)
	}

	// 압축 해제
	if doc.IsCompressionEnabled() && m.config.CompressionEnabled {
		decompressed, err := m.decompress(data, doc.GetCompressionType())
		if err != nil {
			return nil, fmt.Errorf("decompression failed: %w", err)
		}
		data = decompressed
	}

	// AggregateState로 변환
	state, err := doc.ToAggregateState()
	if err != nil {
		return nil, fmt.Errorf("failed to convert document to state: %w", err)
	}

	// 데이터 설정
	state.Data = data

	return state, nil
}

// compress는 데이터를 압축합니다
func (m *MongoStateStore) compress(data []byte) ([]byte, error) {
	switch m.config.CompressionType {
	case CompressionGzip:
		return compressGzip(data)
	case CompressionLZ4:
		return compressLZ4(data)
	default:
		return data, nil
	}
}

// decompress는 데이터 압축을 해제합니다
func (m *MongoStateStore) decompress(data []byte, compressionType string) ([]byte, error) {
	switch CompressionType(compressionType) {
	case CompressionGzip:
		return decompressGzip(data)
	case CompressionLZ4:
		return decompressLZ4(data)
	default:
		return data, nil
	}
}

// Delete는 특정 버전의 상태를 삭제합니다
func (m *MongoStateStore) Delete(ctx context.Context, aggregateID uuid.UUID, version int) error {
	start := time.Now()
	defer func() {
		m.updateMetrics("delete", time.Since(start), nil)
	}()

	filter := bson.M{
		"aggregateId": aggregateID,
		"version":     version,
	}

	result, err := m.collection.DeleteOne(ctx, filter)
	if err != nil {
		m.updateMetrics("delete", time.Since(start), err)
		return fmt.Errorf("failed to delete state: %w", err)
	}

	if result.DeletedCount == 0 {
		return ErrStateNotFound
	}

	return nil
}

// DeleteAll은 집합체의 모든 상태를 삭제합니다
func (m *MongoStateStore) DeleteAll(ctx context.Context, aggregateID uuid.UUID) error {
	start := time.Now()
	defer func() {
		m.updateMetrics("delete", time.Since(start), nil)
	}()

	filter := bson.M{"aggregateId": aggregateID}

	_, err := m.collection.DeleteMany(ctx, filter)
	if err != nil {
		m.updateMetrics("delete", time.Since(start), err)
		return fmt.Errorf("failed to delete all states: %w", err)
	}

	return nil
}

// List는 집합체의 모든 상태 버전을 조회합니다
func (m *MongoStateStore) List(ctx context.Context, aggregateID uuid.UUID) ([]*AggregateState, error) {
	start := time.Now()
	defer func() {
		m.updateMetrics("query", time.Since(start), nil)
	}()

	filter := bson.M{"aggregateId": aggregateID}
	opts := options.Find().SetSort(bson.M{"version": 1})

	cursor, err := m.collection.Find(ctx, filter, opts)
	if err != nil {
		m.updateMetrics("query", time.Since(start), err)
		return nil, fmt.Errorf("failed to list states: %w", err)
	}
	defer cursor.Close(ctx)

	var states []*AggregateState
	for cursor.Next(ctx) {
		var doc mongoStateDocument
		if err := cursor.Decode(&doc); err != nil {
			m.updateMetrics("query", time.Since(start), err)
			return nil, fmt.Errorf("failed to decode state document: %w", err)
		}

		state, err := m.documentToState(&doc)
		if err != nil {
			m.updateMetrics("query", time.Since(start), err)
			return nil, fmt.Errorf("failed to convert document to state: %w", err)
		}

		states = append(states, state)
	}

	if err := cursor.Err(); err != nil {
		m.updateMetrics("query", time.Since(start), err)
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return states, nil
}

// Count는 저장된 상태 개수를 반환합니다
func (m *MongoStateStore) Count(ctx context.Context, aggregateID uuid.UUID) (int64, error) {
	filter := bson.M{"aggregateId": aggregateID}
	count, err := m.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count states: %w", err)
	}
	return count, nil
}

// Exists는 특정 상태가 존재하는지 확인합니다
func (m *MongoStateStore) Exists(ctx context.Context, aggregateID uuid.UUID, version int) (bool, error) {
	filter := bson.M{
		"aggregateId": aggregateID,
		"version":     version,
	}

	count, err := m.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check state existence: %w", err)
	}

	return count > 0, nil
}

// Close는 저장소 연결을 정리합니다
func (m *MongoStateStore) Close() error {
	// MongoDB 클라이언트는 애플리케이션 레벨에서 관리되므로 여기서는 특별히 할 일 없음
	return nil
}

// CreateIndexes는 MongoDB 컬렉션에 필요한 인덱스를 생성합니다
func (m *MongoStateStore) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"aggregateId", 1}},
			Options: options.Index().SetName("idx_aggregateId"),
		},
		{
			Keys:    bson.D{{"aggregateId", 1}, {"version", -1}},
			Options: options.Index().SetName("idx_aggregateId_version"),
		},
		{
			Keys:    bson.D{{"aggregateType", 1}},
			Options: options.Index().SetName("idx_aggregateType"),
		},
		{
			Keys:    bson.D{{"createdAt", 1}},
			Options: options.Index().SetName("idx_createdAt"),
		},
		{
			Keys:    bson.D{{"stateTimestamp", 1}},
			Options: options.Index().SetName("idx_stateTimestamp"),
		},
		{
			Keys:    bson.D{{"tags", 1}},
			Options: options.Index().SetName("idx_tags"),
		},
		{
			Keys: bson.D{{"ttl", 1}},
			Options: options.Index().
				SetName("idx_ttl").
				SetExpireAfterSeconds(0), // TTL 인덱스
		},
	}

	_, err := m.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// GetInternalMetrics는 저장소의 현재 내부 메트릭을 반환합니다
func (m *MongoStateStore) GetInternalMetrics() *mongoStateMetrics {
	m.metrics.mu.RLock()
	defer m.metrics.mu.RUnlock()

	// 메트릭 복사본 반환 (동시성 안전)
	return &mongoStateMetrics{
		saveOps:           m.metrics.saveOps,
		loadOps:           m.metrics.loadOps,
		deleteOps:         m.metrics.deleteOps,
		queryOps:          m.metrics.queryOps,
		totalSaveTime:     m.metrics.totalSaveTime,
		totalLoadTime:     m.metrics.totalLoadTime,
		totalDeleteTime:   m.metrics.totalDeleteTime,
		totalQueryTime:    m.metrics.totalQueryTime,
		errors:            m.metrics.errors,
		lastOperation:     m.metrics.lastOperation,
		compressionSaved:  m.metrics.compressionSaved,
		encryptionApplied: m.metrics.encryptionApplied,
	}
}

// GetMetrics는 StateMetrics 형태로 메트릭을 반환합니다 (MetricsStateStore 인터페이스 구현)
func (m *MongoStateStore) GetMetrics(ctx context.Context) (*StateMetrics, error) {
	// 전체 문서 수 조회
	totalStates, err := m.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to count total states: %w", err)
	}

	// 집합체 타입별 통계
	pipeline := []bson.M{
		{"$group": bson.M{
			"_id":       "$aggregateType",
			"count":     bson.M{"$sum": 1},
			"totalSize": bson.M{"$sum": "$size.original"},
			"avgSize":   bson.M{"$avg": "$size.original"},
			"maxSize":   bson.M{"$max": "$size.original"},
			"minSize":   bson.M{"$min": "$size.original"},
		}},
	}

	cursor, err := m.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate metrics: %w", err)
	}
	defer cursor.Close(ctx)

	var totalStorageBytes int64
	var avgSize, maxSize, minSize int64
	var aggregateTypes []string

	for cursor.Next(ctx) {
		var result struct {
			ID        string  `bson:"_id"`
			Count     int64   `bson:"count"`
			TotalSize int64   `bson:"totalSize"`
			AvgSize   float64 `bson:"avgSize"`
			MaxSize   int64   `bson:"maxSize"`
			MinSize   int64   `bson:"minSize"`
		}

		if err := cursor.Decode(&result); err != nil {
			continue
		}

		aggregateTypes = append(aggregateTypes, result.ID)
		totalStorageBytes += result.TotalSize

		if result.MaxSize > maxSize {
			maxSize = result.MaxSize
		}
		if minSize == 0 || result.MinSize < minSize {
			minSize = result.MinSize
		}
		avgSize = int64(result.AvgSize)
	}

	// 시간 범위 조회
	var oldestState, newestState time.Time
	if totalStates > 0 {
		var oldest, newest mongoStateDocument

		// 가장 오래된 상태
		err = m.collection.FindOne(ctx, bson.M{},
			options.FindOne().SetSort(bson.M{"stateTimestamp": 1})).Decode(&oldest)
		if err == nil {
			oldestState = oldest.StateTimestamp
		}

		// 가장 최신 상태
		err = m.collection.FindOne(ctx, bson.M{},
			options.FindOne().SetSort(bson.M{"stateTimestamp": -1})).Decode(&newest)
		if err == nil {
			newestState = newest.StateTimestamp
		}
	}

	return &StateMetrics{
		TotalStates:       totalStates,
		TotalStorageBytes: totalStorageBytes,
		AverageSize:       avgSize,
		MaxSize:           maxSize,
		MinSize:           minSize,
		AggregateTypes:    aggregateTypes,
		OldestState:       oldestState,
		NewestState:       newestState,
	}, nil
}

// GetAggregateMetrics는 특정 집합체의 메트릭을 반환합니다
func (m *MongoStateStore) GetAggregateMetrics(ctx context.Context, aggregateID uuid.UUID) (*StateMetrics, error) {
	filter := bson.M{"aggregateId": aggregateID}

	// 해당 집합체의 상태 수 조회
	totalStates, err := m.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count aggregate states: %w", err)
	}

	if totalStates == 0 {
		return &StateMetrics{}, nil
	}

	// 집합체별 상세 통계
	pipeline := []bson.M{
		{"$match": filter},
		{"$group": bson.M{
			"_id":        "$aggregateType",
			"count":      bson.M{"$sum": 1},
			"totalSize":  bson.M{"$sum": "$size.original"},
			"avgSize":    bson.M{"$avg": "$size.original"},
			"maxSize":    bson.M{"$max": "$size.original"},
			"minSize":    bson.M{"$min": "$size.original"},
			"oldestDate": bson.M{"$min": "$stateTimestamp"},
			"newestDate": bson.M{"$max": "$stateTimestamp"},
		}},
	}

	cursor, err := m.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate aggregate metrics: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct {
		ID         string    `bson:"_id"`
		Count      int64     `bson:"count"`
		TotalSize  int64     `bson:"totalSize"`
		AvgSize    float64   `bson:"avgSize"`
		MaxSize    int64     `bson:"maxSize"`
		MinSize    int64     `bson:"minSize"`
		OldestDate time.Time `bson:"oldestDate"`
		NewestDate time.Time `bson:"newestDate"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode aggregate metrics: %w", err)
		}
	}

	return &StateMetrics{
		TotalStates:       result.Count,
		TotalStorageBytes: result.TotalSize,
		AverageSize:       int64(result.AvgSize),
		MaxSize:           result.MaxSize,
		MinSize:           result.MinSize,
		AggregateTypes:    []string{result.ID},
		OldestState:       result.OldestDate,
		NewestState:       result.NewestDate,
	}, nil
}

// Query는 조건에 맞는 상태들을 조회합니다 (QueryableStateStore 인터페이스 구현)
func (m *MongoStateStore) Query(ctx context.Context, query StateQuery) ([]*AggregateState, error) {
	start := time.Now()
	defer func() {
		m.updateMetrics("query", time.Since(start), nil)
	}()

	filter := m.buildQueryFilter(query)
	opts := m.buildQueryOptions(query)

	cursor, err := m.collection.Find(ctx, filter, opts)
	if err != nil {
		m.updateMetrics("query", time.Since(start), err)
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer cursor.Close(ctx)

	var states []*AggregateState
	for cursor.Next(ctx) {
		var doc mongoStateDocument
		if err := cursor.Decode(&doc); err != nil {
			m.updateMetrics("query", time.Since(start), err)
			return nil, fmt.Errorf("failed to decode state document: %w", err)
		}

		state, err := m.documentToState(&doc)
		if err != nil {
			m.updateMetrics("query", time.Since(start), err)
			return nil, fmt.Errorf("failed to convert document to state: %w", err)
		}

		states = append(states, state)
	}

	if err := cursor.Err(); err != nil {
		m.updateMetrics("query", time.Since(start), err)
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return states, nil
}

// CountByQuery는 쿼리 조건에 맞는 상태 개수를 반환합니다
func (m *MongoStateStore) CountByQuery(ctx context.Context, query StateQuery) (int64, error) {
	filter := m.buildQueryFilter(query)
	count, err := m.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count by query: %w", err)
	}
	return count, nil
}

// GetAggregateTypes는 저장된 모든 집합체 타입을 반환합니다
func (m *MongoStateStore) GetAggregateTypes(ctx context.Context) ([]string, error) {
	types, err := m.collection.Distinct(ctx, "aggregateType", bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to get aggregate types: %w", err)
	}

	result := make([]string, len(types))
	for i, t := range types {
		if str, ok := t.(string); ok {
			result[i] = str
		}
	}

	return result, nil
}

// GetVersions는 집합체의 모든 버전을 반환합니다
func (m *MongoStateStore) GetVersions(ctx context.Context, aggregateID uuid.UUID) ([]int, error) {
	filter := bson.M{"aggregateId": aggregateID}
	opts := options.Find().SetProjection(bson.M{"version": 1}).SetSort(bson.M{"version": 1})

	cursor, err := m.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}
	defer cursor.Close(ctx)

	var versions []int
	for cursor.Next(ctx) {
		var doc struct {
			Version int `bson:"version"`
		}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		versions = append(versions, doc.Version)
	}

	return versions, nil
}

// buildQueryFilter는 StateQuery로부터 MongoDB 필터를 구성합니다
func (m *MongoStateStore) buildQueryFilter(query StateQuery) bson.M {
	filter := bson.M{}

	if len(query.AggregateIDs) > 0 {
		ids := make([]string, len(query.AggregateIDs))
		for i, id := range query.AggregateIDs {
			ids[i] = id.String()
		}
		filter["aggregateId"] = bson.M{"$in": ids}
	}

	if query.AggregateType != "" {
		filter["aggregateType"] = query.AggregateType
	}

	if query.MinVersion != nil || query.MaxVersion != nil {
		versionFilter := bson.M{}
		if query.MinVersion != nil {
			versionFilter["$gte"] = *query.MinVersion
		}
		if query.MaxVersion != nil {
			versionFilter["$lte"] = *query.MaxVersion
		}
		filter["version"] = versionFilter
	}

	if query.StartTime != nil || query.EndTime != nil {
		timeFilter := bson.M{}
		if query.StartTime != nil {
			timeFilter["$gte"] = *query.StartTime
		}
		if query.EndTime != nil {
			timeFilter["$lte"] = *query.EndTime
		}
		filter["stateTimestamp"] = timeFilter
	}

	return filter
}

// buildQueryOptions는 StateQuery로부터 MongoDB 옵션을 구성합니다
func (m *MongoStateStore) buildQueryOptions(query StateQuery) *options.FindOptions {
	opts := options.Find()

	if query.Limit > 0 {
		opts.SetLimit(int64(query.Limit))
	}

	if query.Offset > 0 {
		opts.SetSkip(int64(query.Offset))
	}

	// 기본적으로 버전 순 정렬
	opts.SetSort(bson.M{"version": 1})

	return opts
}

// updateMetrics는 메트릭을 업데이트합니다
func (m *MongoStateStore) updateMetrics(operation string, duration time.Duration, err error) {
	if !m.config.MetricsEnabled {
		return
	}

	m.metrics.mu.Lock()
	defer m.metrics.mu.Unlock()

	switch operation {
	case "save":
		m.metrics.saveOps++
		m.metrics.totalSaveTime += duration
	case "load":
		m.metrics.loadOps++
		m.metrics.totalLoadTime += duration
	case "delete":
		m.metrics.deleteOps++
		m.metrics.totalDeleteTime += duration
	case "query":
		m.metrics.queryOps++
		m.metrics.totalQueryTime += duration
	}

	if err != nil {
		m.metrics.errors++
	}

	m.metrics.lastOperation = time.Now()
}

// ApplyRetentionPolicy는 보존 정책을 적용합니다 (테스트용 public 메서드)
func (m *MongoStateStore) ApplyRetentionPolicy(ctx context.Context, aggregateID uuid.UUID) {
	states, err := m.List(ctx, aggregateID)
	if err != nil {
		return
	}

	candidates := m.config.RetentionPolicy.GetCleanupCandidates(ctx, states)
	for _, candidate := range candidates {
		m.Delete(ctx, candidate.string, candidate.Version)
	}
}

// applyRetentionPolicy는 내부용 보존 정책 적용 메서드입니다
func (m *MongoStateStore) applyRetentionPolicy(ctx context.Context, aggregateID uuid.UUID) {
	m.ApplyRetentionPolicy(ctx, aggregateID)
}

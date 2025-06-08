// mongo_state_store.go - MongoDB 상태 저장소 구현
package cqrsx

// import (
// 	"context"
// 	"fmt"
// 	"sync"
// 	"time"

// 	"github.com/google/uuid"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// // mongoStateDocument는 MongoDB에 저장되는 상태 문서입니다
// type mongoStateDocument struct {
// 	// 기본 필드
// 	ID string `bson:"_id"`

// 	// 집합체 상태 필드
// 	AggregateID string `bson:"aggregateId"`

// 	// 집합체 타입 필드
// 	AggregateType string `bson:"aggregateType"`

// 	// 상태 버전 필드
// 	Version int `bson:"version"`

// 	// 상태 데이터 필드
// 	Data interface{} `bson:"data"`

// 	// 메타데이터 필드
// 	Metadata interface{} `bson:"metadata"`

// 	// 타임스탬프 필드
// 	Timestamp time.Time `bson:"timestamp"`

// 	// Size는 압축/암호화 적용 전의 원본 데이터 크기입니다
// 	Size int64 `bson:"size"`

// 	// Comppresed 는 데이터가 압축되었는지 여부를 나타냅니다
// 	Compressed bool `bson:"compressed"`

// 	// Encrypted는 데이터가 암호화되었는지 여부를 나타냅니다
// 	Encrypted bool `bson:"encrypted"`

// 	// CreatedAt은 상태가 생성된 타임스탬프입니다
// 	CreatedAt time.Time `bson:"createdAt"`
// }

// // NewMongoStateStore는 새로운 MongoDB 상태 저장소를 생성합니다
// func NewMongoStateStore(collection *mongo.Collection, client *mongo.Client, options ...StateStoreOption) *MongoStateStore {
// 	config := NewDefaultStateStoreConfig()
// 	for _, option := range options {
// 		option(config)
// 	}

// 	store := &MongoStateStore{
// 		collection: collection,
// 		client:     client,
// 		config:     config,
// 		metrics:    &mongoStateMetrics{},
// 	}

// 	// 인덱스 생성 (백그라운드에서)
// 	if config.IndexingEnabled {
// 		go func() {
// 			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 			defer cancel()
// 			if err := store.CreateIndexes(ctx); err != nil {
// 				// 로그 남기기 (실제 로거 사용 권장)
// 				fmt.Printf("Failed to create indexes: %v\n", err)
// 			}
// 		}()
// 	}

// 	return store
// }

// // Save는 집합체 상태를 저장합니다
// func (m *MongoStateStore) Save(ctx context.Context, state *AggregateState) error {
// 	start := time.Now()
// 	defer func() {
// 		m.updateMetrics("save", time.Since(start), nil)
// 	}()

// 	if err := state.Validate(); err != nil {
// 		m.updateMetrics("save", time.Since(start), err)
// 		return fmt.Errorf("invalid state: %w", err)
// 	}

// 	// 데이터 처리 (압축/암호화)
// 	processedData, compressedSize, compressed, encrypted, err := m.processData(state.Data)
// 	if err != nil {
// 		m.updateMetrics("save", time.Since(start), err)
// 		return fmt.Errorf("failed to process data: %w", err)
// 	}

// 	// MongoDB 문서 생성
// 	doc := mongoStateDocument{
// 		ID:            fmt.Sprintf("%s-%d", state.AggregateID.String(), state.Version),
// 		AggregateID:   state.AggregateID.String(),
// 		AggregateType: state.AggregateType,
// 		Version:       state.Version,
// 		Data:          processedData,
// 		Metadata:      state.Metadata,
// 		Timestamp:     state.Timestamp,
// 		Size:          int64(len(processedData)),
// 		Compressed:    compressed,
// 		Encrypted:     encrypted,
// 		CreatedAt:     time.Now(),
// 	}

// 	// 중복 키 에러 방지를 위한 upsert 사용
// 	opts := options.Replace().SetUpsert(true)
// 	filter := bson.M{"_id": doc.ID}

// 	_, err = m.collection.ReplaceOne(ctx, filter, doc, opts)
// 	if err != nil {
// 		m.updateMetrics("save", time.Since(start), err)
// 		return fmt.Errorf("failed to save state: %w", err)
// 	}

// 	// 메트릭 업데이트
// 	if compressed {
// 		originalSize := int64(len(state.Data))
// 		m.metrics.mu.Lock()
// 		m.metrics.compressionSaved += originalSize - doc.Size
// 		m.metrics.mu.Unlock()
// 	}
// 	if encrypted {
// 		m.metrics.mu.Lock()
// 		m.metrics.encryptionApplied++
// 		m.metrics.mu.Unlock()
// 	}

// 	// 보존 정책 적용 (백그라운드에서)
// 	if m.config.RetentionPolicy != nil {
// 		go m.applyRetentionPolicy(context.Background(), state.AggregateID)
// 	}

// 	return nil
// }

// // Load는 집합체의 최신 상태를 로드합니다
// func (m *MongoStateStore) Load(ctx context.Context, aggregateID uuid.UUID) (*AggregateState, error) {
// 	start := time.Now()
// 	defer func() {
// 		m.updateMetrics("load", time.Since(start), nil)
// 	}()

// 	filter := bson.M{"aggregateId": aggregateID.String()}
// 	opts := options.FindOne().SetSort(bson.M{"version": -1})

// 	var doc mongoStateDocument
// 	err := m.collection.FindOne(ctx, filter, opts).Decode(&doc)
// 	if err != nil {
// 		m.updateMetrics("load", time.Since(start), err)
// 		if err == mongo.ErrNoDocuments {
// 			return nil, ErrStateNotFound
// 		}
// 		return nil, fmt.Errorf("failed to load state: %w", err)
// 	}

// 	state, err := m.documentToState(doc)
// 	if err != nil {
// 		m.updateMetrics("load", time.Since(start), err)
// 		return nil, fmt.Errorf("failed to convert document to state: %w", err)
// 	}

// 	return state, nil
// }

// // LoadVersion은 특정 버전의 상태를 로드합니다
// func (m *MongoStateStore) LoadVersion(ctx context.Context, aggregateID uuid.UUID, version int) (*AggregateState, error) {
// 	start := time.Now()
// 	defer func() {
// 		m.updateMetrics("load", time.Since(start), nil)
// 	}()

// 	filter := bson.M{
// 		"aggregateId": aggregateID.String(),
// 		"version":     version,
// 	}

// 	var doc mongoStateDocument
// 	err := m.collection.FindOne(ctx, filter).Decode(&doc)
// 	if err != nil {
// 		m.updateMetrics("load", time.Since(start), err)
// 		if err == mongo.ErrNoDocuments {
// 			return nil, ErrStateNotFound
// 		}
// 		return nil, fmt.Errorf("failed to load state version %d: %w", version, err)
// 	}

// 	state, err := m.documentToState(doc)
// 	if err != nil {
// 		m.updateMetrics("load", time.Since(start), err)
// 		return nil, fmt.Errorf("failed to convert document to state: %w", err)
// 	}

// 	return state, nil
// }

// // Delete는 특정 버전의 상태를 삭제합니다
// func (m *MongoStateStore) Delete(ctx context.Context, aggregateID uuid.UUID, version int) error {
// 	start := time.Now()
// 	defer func() {
// 		m.updateMetrics("delete", time.Since(start), nil)
// 	}()

// 	filter := bson.M{
// 		"aggregateId": aggregateID.String(),
// 		"version":     version,
// 	}

// 	result, err := m.collection.DeleteOne(ctx, filter)
// 	if err != nil {
// 		m.updateMetrics("delete", time.Since(start), err)
// 		return fmt.Errorf("failed to delete state: %w", err)
// 	}

// 	if result.DeletedCount == 0 {
// 		return ErrStateNotFound
// 	}

// 	return nil
// }

// // DeleteAll은 집합체의 모든 상태를 삭제합니다
// func (m *MongoStateStore) DeleteAll(ctx context.Context, aggregateID uuid.UUID) error {
// 	start := time.Now()
// 	defer func() {
// 		m.updateMetrics("delete", time.Since(start), nil)
// 	}()

// 	filter := bson.M{"aggregateId": aggregateID.String()}

// 	_, err := m.collection.DeleteMany(ctx, filter)
// 	if err != nil {
// 		m.updateMetrics("delete", time.Since(start), err)
// 		return fmt.Errorf("failed to delete all states: %w", err)
// 	}

// 	return nil
// }

// // List는 집합체의 모든 상태 버전을 조회합니다
// func (m *MongoStateStore) List(ctx context.Context, aggregateID uuid.UUID) ([]*AggregateState, error) {
// 	start := time.Now()
// 	defer func() {
// 		m.updateMetrics("query", time.Since(start), nil)
// 	}()

// 	filter := bson.M{"aggregateId": aggregateID.String()}
// 	opts := options.Find().SetSort(bson.M{"version": 1})

// 	cursor, err := m.collection.Find(ctx, filter, opts)
// 	if err != nil {
// 		m.updateMetrics("query", time.Since(start), err)
// 		return nil, fmt.Errorf("failed to list states: %w", err)
// 	}
// 	defer cursor.Close(ctx)

// 	var states []*AggregateState
// 	for cursor.Next(ctx) {
// 		var doc mongoStateDocument
// 		if err := cursor.Decode(&doc); err != nil {
// 			m.updateMetrics("query", time.Since(start), err)
// 			return nil, fmt.Errorf("failed to decode state document: %w", err)
// 		}

// 		state, err := m.documentToState(doc)
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

// // Count는 저장된 상태 개수를 반환합니다
// func (m *MongoStateStore) Count(ctx context.Context, aggregateID uuid.UUID) (int64, error) {
// 	filter := bson.M{"aggregateId": aggregateID.String()}
// 	count, err := m.collection.CountDocuments(ctx, filter)
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to count states: %w", err)
// 	}
// 	return count, nil
// }

// // Exists는 특정 상태가 존재하는지 확인합니다
// func (m *MongoStateStore) Exists(ctx context.Context, aggregateID uuid.UUID, version int) (bool, error) {
// 	filter := bson.M{
// 		"aggregateId": aggregateID.String(),
// 		"version":     version,
// 	}

// 	count, err := m.collection.CountDocuments(ctx, filter)
// 	if err != nil {
// 		return false, fmt.Errorf("failed to check state existence: %w", err)
// 	}

// 	return count > 0, nil
// }

// // Close는 저장소 연결을 정리합니다
// func (m *MongoStateStore) Close() error {
// 	// MongoDB 클라이언트는 애플리케이션 레벨에서 관리되므로 여기서는 특별히 할 일 없음
// 	return nil
// }

// // processData는 데이터를 압축/암호화 처리합니다
// //
// // Returns:
// //   - processed: 처리된 데이터 (압축/암호화 적용)
// //   - compressedSize: 압축된 데이터의 크기 (압축 적용 시), 단위 바이트
// //   - compressed: 압축이 적용되었는지 여부
// //   - encrypted: 암호화가 적용되었는지 여부
// //   - err: 오류
// func (m *MongoStateStore) processData(data []byte) (interface{}, int64, bool, bool, error) {
// 	processed := data
// 	compressedSize := int64(len(data))
// 	compressed := false
// 	encrypted := false

// 	// 압축 적용
// 	if m.config.CompressionEnabled && len(data) >= m.config.CompressionMinSize {
// 		var err error
// 		processed, err = m.compress(processed)
// 		compressedSize = int64(len(processed))
// 		if err != nil {
// 			return nil, 0, false, false, fmt.Errorf("compression failed: %w", err)
// 		}
// 		compressed = true
// 	}

// 	// 암호화 적용
// 	if m.config.EncryptionEnabled && m.config.Encryptor != nil {
// 		encryptedData, err := m.config.Encryptor.Encrypt(string(processed))
// 		if err != nil {
// 			return nil, 0, compressed, false, fmt.Errorf("encryption failed: %w", err)
// 		}
// 		processed = encryptedData
// 		encrypted = true
// 	}

// 	return processed, compressedSize, compressed, encrypted, nil
// }

// // documentToState는 MongoDB 문서를 AggregateState로 변환합니다
// func (m *MongoStateStore) documentToState(doc mongoStateDocument) (*AggregateState, error) {
// 	aggregateID, err := uuid.Parse(doc.AggregateID)
// 	if err != nil {
// 		return nil, fmt.Errorf("invalid aggregate ID: %w", err)
// 	}

// 	// 데이터 처리 (복호화/압축 해제)
// 	var data []byte
// 	switch d := doc.Data.(type) {
// 	case []byte:
// 		data = d
// 	case string:
// 		data = []byte(d)
// 	default:
// 		return nil, fmt.Errorf("unsupported data type: %T", doc.Data)
// 	}

// 	// 복호화
// 	if doc.Encrypted && m.config.EncryptionEnabled && m.config.Encryptor != nil {
// 		decrypted, err := m.config.Encryptor.Decrypt(string(data))
// 		if err != nil {
// 			return nil, fmt.Errorf("decryption failed: %w", err)
// 		}
// 		data = []byte(decrypted)
// 	}

// 	// 압축 해제
// 	if doc.Compressed && m.config.CompressionEnabled {
// 		decompressed, err := m.decompress(data)
// 		if err != nil {
// 			return nil, fmt.Errorf("decompression failed: %w", err)
// 		}
// 		data = decompressed
// 	}

// 	// 메타데이터 변환
// 	metadata := make(map[string]any)
// 	if doc.Metadata != nil {
// 		if m, ok := doc.Metadata.(map[string]interface{}); ok {
// 			metadata = m
// 		}
// 	}

// 	return &AggregateState{
// 		AggregateID:   aggregateID,
// 		AggregateType: doc.AggregateType,
// 		Version:       doc.Version,
// 		Data:          data,
// 		Metadata:      metadata,
// 		Timestamp:     doc.Timestamp,
// 	}, nil
// }

// // compress는 데이터를 압축합니다
// func (m *MongoStateStore) compress(data []byte) ([]byte, error) {
// 	switch m.config.CompressionType {
// 	case CompressionGzip:
// 		return compressGzip(data)
// 	case CompressionLZ4:
// 		return compressLZ4(data)
// 	default:
// 		return data, nil
// 	}
// }

// // decompress는 데이터 압축을 해제합니다
// func (m *MongoStateStore) decompress(data []byte) ([]byte, error) {
// 	switch m.config.CompressionType {
// 	case CompressionGzip:
// 		return decompressGzip(data)
// 	case CompressionLZ4:
// 		return decompressLZ4(data)
// 	default:
// 		return data, nil
// 	}
// }

// // updateMetrics는 메트릭을 업데이트합니다
// func (m *MongoStateStore) updateMetrics(operation string, duration time.Duration, err error) {
// 	if !m.config.MetricsEnabled {
// 		return
// 	}

// 	m.metrics.mu.Lock()
// 	defer m.metrics.mu.Unlock()

// 	switch operation {
// 	case "save":
// 		m.metrics.saveOps++
// 		m.metrics.totalSaveTime += duration
// 	case "load":
// 		m.metrics.loadOps++
// 		m.metrics.totalLoadTime += duration
// 	case "delete":
// 		m.metrics.deleteOps++
// 		m.metrics.totalDeleteTime += duration
// 	case "query":
// 		m.metrics.queryOps++
// 		m.metrics.totalQueryTime += duration
// 	}

// 	if err != nil {
// 		m.metrics.errors++
// 	}

// 	m.metrics.lastOperation = time.Now()
// }

// // applyRetentionPolicy는 보존 정책을 적용합니다
// func (m *MongoStateStore) applyRetentionPolicy(ctx context.Context, aggregateID uuid.UUID) {
// 	states, err := m.List(ctx, aggregateID)
// 	if err != nil {
// 		return
// 	}

// 	candidates := m.config.RetentionPolicy.GetCleanupCandidates(ctx, states)
// 	for _, candidate := range candidates {
// 		m.Delete(ctx, candidate.AggregateID, candidate.Version)
// 	}
// }

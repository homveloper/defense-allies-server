package cqrsx

// import (
// 	"bytes"
// 	"compress/gzip"
// 	"context"
// 	"encoding/base64"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"maps"
// 	"time"

// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// // =============================================================================
// // BucketDataRegistry - 이벤트 타입별 강타입 저장소
// // =============================================================================

// // =============================================================================
// // 순수한 이벤트 데이터 구조 (애그리게이트와 무관)
// // =============================================================================

// func NewEventBucketMessage(eventID string, eventType string, aggregateID string,
// 	data interface{}, metadata map[string]interface{}, occurredAt time.Time, version int,
// ) EventBucketMessage {

// 	clonemeta := maps.Clone(metadata)

// 	return EventBucketMessage{
// 		EventID:     eventID,
// 		EventType:   eventType,
// 		AggregateID: aggregateID,
// 		Data:        data,
// 		Metadata:    clonemeta,
// 		OccurredAt:  occurredAt,
// 		Version:     version,
// 	}
// }

// type EventBucketData interface{}

// // EventBucketMessage는 이벤트 저장소에 저장되는 개별 이벤트 데이터 구조체입니다.
// // JSON 직렬화가 가능하며 이벤트 버킷 내에서 압축된 형태로 저장됩니다.
// type EventBucketMessage struct {
// 	// EventID는 이벤트의 고유 식별자입니다.
// 	EventID string `json:"event_id" bson:"event_id"`
// 	// EventType은 이벤트 타입을 나타냅니다 (예: "UserCreated", "OrderPlaced").
// 	EventType string `json:"event_type" bson:"event_type"`
// 	// AggregateID는 이벤트가 속한 애그리게이트의 식별자입니다.
// 	AggregateID string `json:"aggregate_id" bson:"aggregate_id"`
// 	// Data는 이벤트의 실제 데이터입니다 (강타입).
// 	Data EventBucketData `json:"data" bson:"data"`
// 	// Metadata는 이벤트 관련 메타데이터로 추적, 디버깅 등에 사용됩니다.
// 	Metadata map[string]interface{} `json:"metadata" bson:"metadata"`
// 	// OccurredAt은 이벤트 발생 시간입니다.
// 	OccurredAt time.Time `json:"occurred_at" bson:"occurred_at"`
// 	// Version은 이벤트 버전으로 낙관적 동시성 제어에 사용됩니다.
// 	Version int `json:"version" bson:"version"`
// }

// type EventBucket struct {
// 	// ID는 버킷의 고유 식별자입니다.
// 	ID string `bson:"_id,omitempty"`
// 	// AggregateID는 이벤트가 속한 애그리게이트의 식별자입니다.
// 	AggregateID string `bson:"aggregate_id"`
// 	// Events는 압축 해제된 이벤트 데이터 배열입니다.
// 	Events []EventBucketMessage `bson:"events"`
// 	// EventCount는 버킷에 포함된 이벤트 수입니다.
// 	EventCount int `bson:"event_count"`
// 	// Version은 버킷 버전으로 낙관적 잠금에 사용됩니다.
// 	Version int `bson:"version"`
// 	// CreatedAt은 버킷 생성 시간입니다.
// 	CreatedAt time.Time `bson:"created_at"`
// 	// UpdatedAt은 버킷 마지막 업데이트 시간입니다.
// 	UpdatedAt time.Time `bson:"updated_at"`
// 	// FirstEventTime은 버킷 내 첫 번째 이벤트 발생 시간입니다.
// 	FirstEventTime *time.Time `bson:"first_event_time,omitempty"`
// 	// LastEventTime은 버킷 내 마지막 이벤트 발생 시간입니다.
// 	LastEventTime *time.Time `bson:"last_event_time,omitempty"`
// }

// // =============================================================================
// // 요청/응답 구조체들
// // =============================================================================

// // 버킷 원시 조회 결과
// type BucketFindOneResult struct {
// 	Bucket *EventBucket `json:"bucket"`
// }

// // 버킷 생성/업데이트 요청
// type BucketWriteRequest struct {
// 	AggregateID     string               `json:"aggregate_id"`
// 	Events          []EventBucketMessage `json:"events"`
// 	ExpectedVersion int                  `json:"expected_version"` // 낙관적 잠금용
// }

// // Exists 결과
// type BucketExistsResult struct {
// 	Exists  bool      `json:"exists"`
// 	Version int       `json:"version"`
// 	Updated time.Time `json:"updated_at"`
// }

// // 버킷 뷰 조회 결과
// type BucketViewReadResult struct {
// 	View *EventBucket `json:"view"` // 전체 버킷 뷰 (강타입)
// }

// // 버킷 생성 결과
// type BucketInsertResult struct {
// 	Bucket   *EventBucket         `json:"view"`     // 전체 버킷 뷰 (강타입)
// 	Inserted []EventBucketMessage `json:"inserted"` // 추가된 이벤트들 (강타입)
// }

// // 버킷 업데이트 결과
// type BucketUpdateResult struct {
// 	Bucket   *EventBucket         `json:"view"`     // 전체 버킷 뷰 (강타입)
// 	Inserted []EventBucketMessage `json:"inserted"` // 추가된 이벤트들 (강타입)
// }

// // =============================================================================
// // BucketEventStore 인터페이스 (애그리게이트와 무관한 순수 이벤트 스토어)
// // =============================================================================

// type BucketEventStore interface {
// 	// === 읽기 작업 ===
// 	FindOne(ctx context.Context, aggregateID string) (*BucketFindOneResult, error)

// 	// === 쓰기 작업 (원자성 보장) ===

// 	// 새 버킷 생성
// 	FindOneAndInsert(ctx context.Context, request BucketWriteRequest) (*BucketInsertResult, error)

// 	// 기존 버킷 업데이트 (낙관적 잠금)
// 	FindOneAndUpdate(ctx context.Context, request BucketWriteRequest) (*BucketUpdateResult, error)

// 	// 유틸리티
// 	Exists(ctx context.Context, aggregateID string) (*BucketExistsResult, error)
// }

// // 이벤트 압축/해제 서비스 인터페이스
// type BucketCompressor interface {
// 	Compress(events []EventBucketMessage) (compressed string, uncompressedSize, compressedSize int, err error)
// 	Decompress(compressed string) ([]EventBucketMessage, error)
// }

// // =============================================================================
// // MongoDB 구현체
// // =============================================================================

// type MongoBucketEventStore struct {
// 	collection *mongo.Collection
// 	// compressor BucketCompressor
// 	serializer BucketSerializer
// }

// func NewMongoBucketEventStore(
// 	collection *mongo.Collection,
// 	// compressor BucketCompressor,
// 	serializer BucketSerializer,
// ) *MongoBucketEventStore {
// 	return &MongoBucketEventStore{
// 		collection: collection,
// 		// compressor: compressor,
// 		serializer: serializer,
// 	}
// }

// // === 읽기 작업 ===

// // 버킷 뷰 조회 (EventBucket, 강타입 선택 가능)
// func (s *MongoBucketEventStore) FindOne(ctx context.Context, aggregateID string) (*BucketFindOneResult, error) {
// 	filter := bson.M{"aggregate_id": aggregateID}

// 	var bucket EventBucket
// 	err := s.collection.FindOne(ctx, filter).Decode(&bucket)
// 	if err != nil {
// 		if errors.Is(err, mongo.ErrNoDocuments) {
// 			return nil, ErrBucketNotFound{AggregateID: aggregateID}
// 		}
// 		return nil, fmt.Errorf("failed to find bucket: %w", err)
// 	}

// 	return &BucketFindOneResult{
// 		Bucket: &bucket,
// 	}, nil
// }

// // 새 버킷 생성 (원자성 보장)
// func (s *MongoBucketEventStore) FindOneAndInsert(ctx context.Context, request BucketWriteRequest) (*BucketInsertResult, error) {
// 	// 입력 검증
// 	if err := s.validateWriteRequest(request); err != nil {
// 		return nil, fmt.Errorf("invalid write request: %w", err)
// 	}

// 	// 시간 정보 계산
// 	var firstEventTime, lastEventTime *time.Time
// 	if len(request.Events) > 0 {
// 		firstEventTime = &request.Events[0].OccurredAt
// 		lastEventTime = &request.Events[len(request.Events)-1].OccurredAt
// 	}

// 	// 이벤트 복사
// 	events := make([]EventBucketMessage, len(request.Events))
// 	copy(events, request.Events)

// 	// 새 버킷 생성
// 	now := time.Now()
// 	newBucket := EventBucket{
// 		AggregateID:    request.AggregateID,
// 		Events:         events,
// 		EventCount:     len(request.Events),
// 		Version:        calculateNewVersion(request.Events),
// 		CreatedAt:      now,
// 		UpdatedAt:      now,
// 		FirstEventTime: firstEventTime,
// 		LastEventTime:  lastEventTime,
// 	}

// 	// MongoDB 삽입 (중복 시 실패)
// 	filter := bson.M{"aggregate_id": request.AggregateID}
// 	update := bson.M{"$setOnInsert": newBucket}
// 	opts := options.FindOneAndUpdate().
// 		SetUpsert(true).
// 		SetReturnDocument(options.After)

// 	var result EventBucket
// 	err := s.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
// 	if err != nil {
// 		// 중복 키 에러 확인
// 		if mongo.IsDuplicateKeyError(err) {
// 			return nil, ErrBucketAlreadyExists{AggregateID: request.AggregateID}
// 		}
// 		return nil, fmt.Errorf("failed to insert bucket: %w", err)
// 	}

// 	// 결과가 새로 생성된 것인지 확인
// 	if result.Version != newBucket.Version {
// 		return nil, ErrBucketAlreadyExists{AggregateID: request.AggregateID}
// 	}

// 	return &BucketInsertResult{
// 		Bucket:   &result,
// 		Inserted: request.Events,
// 	}, nil
// }

// // 기존 버킷 업데이트 (원자성 + 낙관적 잠금)
// func (s *MongoBucketEventStore) FindOneAndUpdate(ctx context.Context, request BucketWriteRequest) (*BucketUpdateResult, error) {
// 	// 입력 검증
// 	if err := s.validateWriteRequest(request); err != nil {
// 		return nil, fmt.Errorf("invalid write request: %w", err)
// 	}

// 	// 이벤트 데이터 직렬화
// 	serializeds := make([]EventBucketMessage, len(request.Events))
// 	for i, event := range request.Events {
// 		serializedData, err := s.serializer.SerializeEventData(event.EventType, event.Data)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to serialize event data: %w", err)
// 		}
// 		serializeds[i] = event
// 		serializeds[i].Data = serializedData
// 	}

// 	// 1. 기존 버킷 조회
// 	existingResult, err := s.FindOne(ctx, request.AggregateID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to find existing bucket: %w", err)
// 	}

// 	// 2. 버전 확인 (낙관적 잠금)
// 	if existingResult.Bucket.Version != request.ExpectedVersion {
// 		return nil, ErrConcurrencyConflict{
// 			AggregateID:     request.AggregateID,
// 			ExpectedVersion: request.ExpectedVersion,
// 			ActualVersion:   existingResult.Bucket.Version,
// 		}
// 	}

// 	// 3. 기존 이벤트와 새 이벤트 병합
// 	allEvents := append(existingResult.Events, request.Events...)

// 	// 4. 전체 이벤트 압축
// 	compressedData, uncompressedSize, compressedSize, err := s.compressor.Compress(allEvents)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to compress merged events: %w", err)
// 	}

// 	// 5. 시간 정보 업데이트
// 	var firstEventTime, lastEventTime *time.Time
// 	if len(allEvents) > 0 {
// 		firstEventTime = &allEvents[0].OccurredAt
// 		lastEventTime = &allEvents[len(allEvents)-1].OccurredAt
// 	}

// 	// 6. 원자적 업데이트
// 	filter := bson.M{
// 		"aggregate_id": request.AggregateID,
// 		"version":      request.ExpectedVersion, // 낙관적 잠금
// 	}

// 	update := bson.M{
// 		"$set": bson.M{
// 			"compressed_events": compressedData,
// 			"event_count":       len(allEvents),
// 			"version":           calculateNewVersion(request.Events) + request.ExpectedVersion,
// 			"updated_at":        time.Now(),
// 			"uncompressed_size": uncompressedSize,
// 			"compressed_size":   compressedSize,
// 			"first_event_time":  firstEventTime,
// 			"last_event_time":   lastEventTime,
// 		},
// 	}

// 	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

// 	var updatedBucket compressedEventBucket
// 	err = s.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedBucket)
// 	if err != nil {
// 		if errors.Is(err, mongo.ErrNoDocuments) {
// 			// 다른 프로세스가 이미 업데이트함
// 			currentResult, fetchErr := s.FindOne(ctx, request.AggregateID)
// 			if fetchErr == nil {
// 				return nil, ErrConcurrencyConflict{
// 					AggregateID:     request.AggregateID,
// 					ExpectedVersion: request.ExpectedVersion,
// 					ActualVersion:   currentResult.Bucket.Version,
// 				}
// 			}
// 			return nil, ErrConcurrencyConflict{
// 				AggregateID:     request.AggregateID,
// 				ExpectedVersion: request.ExpectedVersion,
// 				ActualVersion:   -1, // 알 수 없음
// 			}
// 		}
// 		return nil, fmt.Errorf("failed to update bucket: %w", err)
// 	}

// 	return &BucketReadResult{
// 		Bucket: &updatedBucket,
// 		Events: allEvents,
// 	}, nil
// }

// // 버킷 존재 여부 확인
// func (s *MongoBucketEventStore) Exists(ctx context.Context, aggregateID string) (bool, error) {
// 	filter := bson.M{"aggregate_id": aggregateID}
// 	count, err := s.collection.CountDocuments(ctx, filter, options.Count().SetLimit(1))
// 	if err != nil {
// 		return false, fmt.Errorf("failed to check bucket existence: %w", err)
// 	}
// 	return count > 0, nil
// }

// // 현재 버전 조회
// func (s *MongoBucketEventStore) GetVersion(ctx context.Context, aggregateID string) (int, error) {
// 	filter := bson.M{"aggregate_id": aggregateID}
// 	projection := bson.M{"version": 1}
// 	opts := options.FindOne().SetProjection(projection)

// 	var result struct {
// 		Version int `bson:"version"`
// 	}

// 	err := s.collection.FindOne(ctx, filter, opts).Decode(&result)
// 	if err != nil {
// 		if errors.Is(err, mongo.ErrNoDocuments) {
// 			return 0, ErrBucketNotFound{AggregateID: aggregateID}
// 		}
// 		return 0, fmt.Errorf("failed to get bucket version: %w", err)
// 	}

// 	return result.Version, nil
// }

// // =============================================================================
// // 검증 및 유틸리티 함수
// // =============================================================================

// func (s *MongoBucketEventStore) validateWriteRequest(request BucketWriteRequest) error {
// 	if request.AggregateID == "" {
// 		return errors.New("aggregate_id is required")
// 	}

// 	if len(request.Events) == 0 {
// 		return errors.New("events cannot be empty")
// 	}

// 	if request.ExpectedVersion < 0 {
// 		return errors.New("expected_version must be non-negative")
// 	}

// 	// 이벤트 데이터 검증
// 	for i, event := range request.Events {
// 		if event.EventID == "" {
// 			return fmt.Errorf("event[%d].event_id is required", i)
// 		}
// 		if event.EventType == "" {
// 			return fmt.Errorf("event[%d].event_type is required", i)
// 		}
// 		if event.AggregateID != request.AggregateID {
// 			return fmt.Errorf("event[%d].aggregate_id mismatch", i)
// 		}
// 		if event.Version <= 0 {
// 			return fmt.Errorf("event[%d].version must be positive", i)
// 		}
// 	}

// 	return nil
// }

// func calculateNewVersion(events []EventBucketMessage) int {
// 	if len(events) == 0 {
// 		return 0
// 	}
// 	// 마지막 이벤트의 버전을 새 버전으로 사용
// 	maxVersion := 0
// 	for _, event := range events {
// 		if event.Version > maxVersion {
// 			maxVersion = event.Version
// 		}
// 	}
// 	return maxVersion
// }

// // =============================================================================
// // 에러 타입들
// // =============================================================================

// type ErrBucketNotFound struct {
// 	AggregateID string
// }

// func (e ErrBucketNotFound) Error() string {
// 	return fmt.Sprintf("bucket not found for aggregate: %s", e.AggregateID)
// }

// func (e ErrBucketNotFound) IsNotFound() bool {
// 	return true
// }

// type ErrBucketAlreadyExists struct {
// 	AggregateID string
// }

// func (e ErrBucketAlreadyExists) Error() string {
// 	return fmt.Sprintf("bucket already exists for aggregate: %s", e.AggregateID)
// }

// func (e ErrBucketAlreadyExists) IsConflict() bool {
// 	return true
// }

// type ErrConcurrencyConflict struct {
// 	AggregateID     string
// 	ExpectedVersion int
// 	ActualVersion   int
// }

// func (e ErrConcurrencyConflict) Error() string {
// 	return fmt.Sprintf("concurrency conflict for aggregate %s: expected version %d, actual version %d",
// 		e.AggregateID, e.ExpectedVersion, e.ActualVersion)
// }

// func (e ErrConcurrencyConflict) IsConcurrencyError() bool {
// 	return true
// }

// type GzipEventCompressionService struct{}

// func NewGzipEventCompressionService() *GzipEventCompressionService {
// 	return &GzipEventCompressionService{}
// }

// func (s *GzipEventCompressionService) Compress(events []EventBucketMessage) (string, int, int, error) {
// 	// 1. JSON 마샬링
// 	jsonData, err := json.Marshal(events)
// 	if err != nil {
// 		return "", 0, 0, fmt.Errorf("json marshaling failed: %w", err)
// 	}

// 	uncompressedSize := len(jsonData)

// 	// 2. gzip 압축
// 	var buf bytes.Buffer
// 	gzipWriter := gzip.NewWriter(&buf)

// 	if _, err := gzipWriter.Write(jsonData); err != nil {
// 		gzipWriter.Close()
// 		return "", 0, 0, fmt.Errorf("gzip compression failed: %w", err)
// 	}

// 	if err := gzipWriter.Close(); err != nil {
// 		return "", 0, 0, fmt.Errorf("gzip writer close failed: %w", err)
// 	}

// 	compressedData := buf.Bytes()
// 	compressedSize := len(compressedData)

// 	// 3. base64 인코딩
// 	encoded := base64.StdEncoding.EncodeToString(compressedData)

// 	return encoded, uncompressedSize, compressedSize, nil
// }

// func (s *GzipEventCompressionService) Decompress(compressed string) ([]EventBucketMessage, error) {
// 	// 1. base64 디코딩
// 	decoded, err := base64.StdEncoding.DecodeString(compressed)
// 	if err != nil {
// 		return nil, fmt.Errorf("base64 decoding failed: %w", err)
// 	}

// 	// 2. gzip 압축 해제
// 	reader, err := gzip.NewReader(bytes.NewReader(decoded))
// 	if err != nil {
// 		return nil, fmt.Errorf("gzip reader creation failed: %w", err)
// 	}
// 	defer reader.Close()

// 	jsonData, err := io.ReadAll(reader)
// 	if err != nil {
// 		return nil, fmt.Errorf("gzip decompression failed: %w", err)
// 	}

// 	// 3. JSON 언마샬링
// 	var events []EventBucketMessage
// 	if err := json.Unmarshal(jsonData, &events); err != nil {
// 		return nil, fmt.Errorf("json unmarshaling failed: %w", err)
// 	}

// 	return events, nil
// }

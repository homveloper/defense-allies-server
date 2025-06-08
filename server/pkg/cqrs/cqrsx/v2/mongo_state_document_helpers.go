// mongo_state_document_helpers.go - MongoDB 상태 문서 헬퍼 메서드들
package cqrsx

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ================================
// 🗜️ 압축 관련 헬퍼 메서드들
// ================================

// IsCompressionEnabled는 압축이 적용되었는지 확인합니다
func (doc *mongoStateDocument) IsCompressionEnabled() bool {
	return doc.Compression != nil && doc.Compression.Type != "" && doc.Compression.Type != "none"
}

// GetCompressionType은 압축 타입을 안전하게 반환합니다
func (doc *mongoStateDocument) GetCompressionType() string {
	if doc.IsCompressionEnabled() {
		return doc.Compression.Type
	}
	return "none"
}

// GetCompressionInfo는 압축 정보를 안전하게 반환합니다
func (doc *mongoStateDocument) GetCompressionInfo() *CompressionInfo {
	if doc.IsCompressionEnabled() {
		return doc.Compression
	}
	return nil
}

// SetCompressionInfo는 압축 정보를 설정합니다
func (doc *mongoStateDocument) SetCompressionInfo(compressionType string, version string, level int, metadata map[string]interface{}) {
	if compressionType == "" || compressionType == "none" {
		doc.Compression = nil
		return
	}

	doc.Compression = &CompressionInfo{
		Type:     compressionType,
		Version:  version,
		Level:    level,
		Metadata: metadata,
	}
}

// ================================
// 🔐 암호화 관련 헬퍼 메서드들
// ================================

// IsEncryptionEnabled는 암호화가 적용되었는지 확인합니다
func (doc *mongoStateDocument) IsEncryptionEnabled() bool {
	return doc.Encryption != nil && doc.Encryption.Type != "" && doc.Encryption.Type != "none"
}

// GetEncryptionType은 암호화 타입을 안전하게 반환합니다
func (doc *mongoStateDocument) GetEncryptionType() string {
	if doc.IsEncryptionEnabled() {
		return doc.Encryption.Type
	}
	return "none"
}

// GetEncryptionInfo는 암호화 정보를 안전하게 반환합니다
func (doc *mongoStateDocument) GetEncryptionInfo() *EncryptionInfo {
	if doc.IsEncryptionEnabled() {
		return doc.Encryption
	}
	return nil
}

// SetEncryptionInfo는 암호화 정보를 설정합니다
func (doc *mongoStateDocument) SetEncryptionInfo(encryptionType string, version string, keySize int, metadata map[string]interface{}) {
	if encryptionType == "" || encryptionType == "none" {
		doc.Encryption = nil
		return
	}

	doc.Encryption = &EncryptionInfo{
		Type:     encryptionType,
		Version:  version,
		KeySize:  keySize,
		Metadata: metadata,
	}
}

// ================================
// 🕐 시간 관련 헬퍼 메서드들
// ================================

// UpdateTimestamps는 생성/수정 시간을 업데이트합니다
func (doc *mongoStateDocument) UpdateTimestamps() {
	now := time.Now()
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = now
	}
	doc.UpdatedAt = now
}

// TouchAccessTime은 마지막 액세스 시간을 현재 시간으로 업데이트합니다
func (doc *mongoStateDocument) TouchAccessTime() {
	now := time.Now()
	doc.LastAccessedAt = &now
}

// SetTTL은 TTL(Time To Live)을 설정합니다
func (doc *mongoStateDocument) SetTTL(duration time.Duration) {
	ttl := time.Now().Add(duration)
	doc.TTL = &ttl
}

// IsExpired는 TTL이 만료되었는지 확인합니다
func (doc *mongoStateDocument) IsExpired() bool {
	return doc.TTL != nil && time.Now().After(*doc.TTL)
}

// GetAge는 문서가 생성된 후 경과 시간을 반환합니다
func (doc *mongoStateDocument) GetAge() time.Duration {
	if doc.CreatedAt.IsZero() {
		return 0
	}
	return time.Since(doc.CreatedAt)
}

// ================================
// 📊 메타데이터 관련 헬퍼 메서드들
// ================================

// SetSystemMetadata는 시스템 메타데이터를 설정합니다
func (doc *mongoStateDocument) SetSystemMetadata(key string, value interface{}) {
	if doc.SystemMetadata == nil {
		doc.SystemMetadata = make(map[string]interface{})
	}
	doc.SystemMetadata[key] = value
}

// GetSystemMetadata는 시스템 메타데이터를 조회합니다
func (doc *mongoStateDocument) GetSystemMetadata(key string) (interface{}, bool) {
	if doc.SystemMetadata == nil {
		return nil, false
	}
	value, exists := doc.SystemMetadata[key]
	return value, exists
}

// AddTag는 태그를 추가합니다
func (doc *mongoStateDocument) AddTag(tag string) {
	// 중복 체크
	for _, existingTag := range doc.Tags {
		if existingTag == tag {
			return
		}
	}
	doc.Tags = append(doc.Tags, tag)
}

// RemoveTag는 태그를 제거합니다
func (doc *mongoStateDocument) RemoveTag(tag string) {
	for i, existingTag := range doc.Tags {
		if existingTag == tag {
			doc.Tags = append(doc.Tags[:i], doc.Tags[i+1:]...)
			return
		}
	}
}

// HasTag는 특정 태그가 있는지 확인합니다
func (doc *mongoStateDocument) HasTag(tag string) bool {
	for _, existingTag := range doc.Tags {
		if existingTag == tag {
			return true
		}
	}
	return false
}

// ================================
// ⚡ 성능 관련 헬퍼 메서드들
// ================================

// RecordProcessingTime은 처리 시간을 기록합니다
func (doc *mongoStateDocument) RecordProcessingTime(duration time.Duration) {
	doc.ProcessingTimeMs = duration.Milliseconds()
}

// IncrementRetryCount는 재시도 횟수를 증가시킵니다
func (doc *mongoStateDocument) IncrementRetryCount() {
	doc.RetryCount++
}

// IncrementErrorCount는 에러 횟수를 증가시킵니다
func (doc *mongoStateDocument) IncrementErrorCount() {
	doc.ErrorCount++
}

// SetLastError는 마지막 에러 메시지를 설정합니다
func (doc *mongoStateDocument) SetLastError(err error) {
	if err != nil {
		doc.LastError = err.Error()
		doc.IncrementErrorCount()
	} else {
		doc.LastError = ""
	}
}

// ================================
// 🏗️ 문서 생성 헬퍼 메서드들
// ================================

// NewMongoStateDocument는 새로운 MongoDB 상태 문서를 생성합니다
func NewMongoStateDocument(aggregateID uuid.UUID, aggregateType string, version int, stateTimestamp time.Time) *mongoStateDocument {
	doc := &mongoStateDocument{
		ID:              fmt.Sprintf("%s-%s-v%d", aggregateType, aggregateID.String(), version),
		AggregateID:     aggregateID.String(),
		AggregateType:   aggregateType,
		Version:         version,
		DocumentVersion: DocumentVersionCurrent,
		StateTimestamp:  stateTimestamp,
		DataFormat:      "json",
		DataEncoding:    "raw",
		SystemMetadata:  make(map[string]interface{}),
		Tags:            make([]string, 0),
	}

	doc.UpdateTimestamps()
	return doc
}

// FromAggregateState는 AggregateState로부터 MongoDB 문서를 생성합니다
func FromAggregateState(state *AggregateState) *mongoStateDocument {
	doc := NewMongoStateDocument(state.AggregateID, state.AggregateType, state.Version, state.Timestamp)

	// 원본 데이터 설정
	doc.Data = state.Data

	// 메타데이터 복사
	if state.Metadata != nil {
		for k, v := range state.Metadata {
			doc.SetSystemMetadata(k, v)
		}
	}

	return doc
}

// ToAggregateState는 MongoDB 문서를 AggregateState로 변환합니다
func (doc *mongoStateDocument) ToAggregateState() (*AggregateState, error) {
	aggregateID, err := uuid.Parse(doc.AggregateID)
	if err != nil {
		return nil, fmt.Errorf("invalid aggregate ID: %w", err)
	}

	// 데이터 변환
	var data []byte
	switch d := doc.Data.(type) {
	case []byte:
		data = d
	case string:
		data = []byte(d)
	default:
		// JSON으로 직렬화
		jsonData, err := json.Marshal(d)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data: %w", err)
		}
		data = jsonData
	}

	// AggregateState 생성
	state := &AggregateState{
		AggregateID:   aggregateID,
		AggregateType: doc.AggregateType,
		Version:       doc.Version,
		Data:          data,
		Metadata:      make(map[string]interface{}),
		Timestamp:     doc.StateTimestamp,
	}

	// 시스템 메타데이터를 AggregateState 메타데이터로 복사
	if doc.SystemMetadata != nil {
		for k, v := range doc.SystemMetadata {
			state.Metadata[k] = v
		}
	}

	return state, nil
}

// ================================
// 🔍 유틸리티 메서드들
// ================================

// Clone은 문서의 깊은 복사본을 생성합니다
func (doc *mongoStateDocument) Clone() *mongoStateDocument {
	// JSON 직렬화/역직렬화를 통한 깊은 복사
	data, _ := json.Marshal(doc)
	cloned := &mongoStateDocument{}
	json.Unmarshal(data, cloned)
	return cloned
}

// GetSummary는 문서의 요약 정보를 반환합니다
func (doc *mongoStateDocument) GetSummary() map[string]interface{} {
	return map[string]interface{}{
		"id":             doc.ID,
		"aggregateType":  doc.AggregateType,
		"version":        doc.Version,
		"compressed":     doc.IsCompressionEnabled(),
		"encrypted":      doc.IsEncryptionEnabled(),
		"age":            doc.GetAge().String(),
		"processingTime": fmt.Sprintf("%dms", doc.ProcessingTimeMs),
		"retryCount":     doc.RetryCount,
		"errorCount":     doc.ErrorCount,
	}
}

// Validate는 문서의 유효성을 검증합니다
func (doc *mongoStateDocument) Validate() error {
	if doc.ID == "" {
		return fmt.Errorf("document ID cannot be empty")
	}

	if doc.AggregateID == "" {
		return fmt.Errorf("aggregate ID cannot be empty")
	}

	if doc.AggregateType == "" {
		return fmt.Errorf("aggregate type cannot be empty")
	}

	if doc.Version < 0 {
		return fmt.Errorf("version cannot be negative")
	}

	// 압축 정보 검증
	if doc.Compression != nil {
		if doc.Compression.Type == "" {
			return fmt.Errorf("compression type cannot be empty when compression info is set")
		}
		if doc.Compression.Level < 0 || doc.Compression.Level > 9 {
			return fmt.Errorf("compression level must be between 0 and 9")
		}
	}

	// 암호화 정보 검증
	if doc.Encryption != nil {
		if doc.Encryption.Type == "" {
			return fmt.Errorf("encryption type cannot be empty when encryption info is set")
		}
		if doc.Encryption.KeySize != 128 && doc.Encryption.KeySize != 192 && doc.Encryption.KeySize != 256 {
			return fmt.Errorf("encryption key size must be 128, 192, or 256")
		}
	}

	return nil
}

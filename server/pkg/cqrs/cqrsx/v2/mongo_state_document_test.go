// mongo_state_document_test.go - MongoDB 상태 문서 테스트
package cqrsx

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMongoStateDocument(t *testing.T) {
	// Given
	aggregateID := uuid.New()
	aggregateType := "Guild"
	version := 5
	stateTimestamp := time.Now()

	// When
	doc := NewMongoStateDocument(aggregateID, aggregateType, version, stateTimestamp)

	// Then
	assert.Equal(t, fmt.Sprintf("%s-%s-v%d", aggregateType, aggregateID, version), doc.ID)
	assert.Equal(t, aggregateID, doc.string)
	assert.Equal(t, aggregateType, doc.AggregateType)
	assert.Equal(t, version, doc.Version)
	assert.Equal(t, DocumentVersionCurrent, doc.DocumentVersion)
	assert.Equal(t, stateTimestamp, doc.StateTimestamp)
	assert.Equal(t, "json", doc.DataFormat)
	assert.Equal(t, "raw", doc.DataEncoding)
	assert.NotNil(t, doc.SystemMetadata)
	assert.NotNil(t, doc.Tags)
	assert.False(t, doc.CreatedAt.IsZero())
	assert.False(t, doc.UpdatedAt.IsZero())
}

func TestFromAggregateState(t *testing.T) {
	// Given
	aggregateID := uuid.New()
	data := []byte(`{"name": "Test Guild", "level": 5}`)
	state := NewAggregateState(aggregateID, "Guild", 3, data)
	state.SetMetadata("source", "test")
	state.SetMetadata("environment", "development")

	// When
	doc := FromAggregateState(state)

	// Then
	assert.Equal(t, state.string.String(), doc.string)
	assert.Equal(t, state.AggregateType, doc.AggregateType)
	assert.Equal(t, state.Version, doc.Version)
	assert.Equal(t, state.Data, doc.Data)

	// 메타데이터 확인
	source, exists := doc.GetSystemMetadata("source")
	assert.True(t, exists)
	assert.Equal(t, "test", source)

	env, exists := doc.GetSystemMetadata("environment")
	assert.True(t, exists)
	assert.Equal(t, "development", env)
}

func TestToAggregateState(t *testing.T) {
	// Given
	aggregateID := uuid.New()
	doc := NewMongoStateDocument(aggregateID, "Guild", 3, time.Now())
	doc.Data = []byte(`{"name": "Test Guild", "level": 5}`)
	doc.SetSystemMetadata("source", "test")
	doc.SetSystemMetadata("environment", "development")

	// When
	state, err := doc.ToAggregateState()

	// Then
	require.NoError(t, err)
	assert.Equal(t, aggregateID, state.string)
	assert.Equal(t, "Guild", state.AggregateType)
	assert.Equal(t, 3, state.Version)
	assert.Equal(t, doc.Data, state.Data)

	// 메타데이터 확인
	source, exists := state.GetMetadata("source")
	assert.True(t, exists)
	assert.Equal(t, "test", source)

	env, exists := state.GetMetadata("environment")
	assert.True(t, exists)
	assert.Equal(t, "development", env)
}

func TestCompressionInfo(t *testing.T) {
	// Given
	doc := NewMongoStateDocument(uuid.New(), "Guild", 1, time.Now())

	// When - 압축 정보가 없는 경우
	assert.False(t, doc.IsCompressionEnabled())
	assert.Equal(t, "none", doc.GetCompressionType())
	assert.Nil(t, doc.GetCompressionInfo())

	// When - 압축 정보 설정
	metadata := map[string]interface{}{
		"windowSize": 15,
		"memLevel":   8,
	}
	doc.SetCompressionInfo("gzip", "1.0", 6, metadata)

	// Then
	assert.True(t, doc.IsCompressionEnabled())
	assert.Equal(t, "gzip", doc.GetCompressionType())
	assert.NotNil(t, doc.GetCompressionInfo())
	assert.Equal(t, "gzip", doc.Compression.Type)
	assert.Equal(t, "1.0", doc.Compression.Version)
	assert.Equal(t, 6, doc.Compression.Level)
	assert.Equal(t, metadata, doc.Compression.Metadata)

	// When - 압축 해제
	doc.SetCompressionInfo("none", "", 0, nil)
	assert.False(t, doc.IsCompressionEnabled())
	assert.Nil(t, doc.Compression)
}

func TestEncryptionInfo(t *testing.T) {
	// Given
	doc := NewMongoStateDocument(uuid.New(), "Guild", 1, time.Now())

	// When - 암호화 정보가 없는 경우
	assert.False(t, doc.IsEncryptionEnabled())
	assert.Equal(t, "none", doc.GetEncryptionType())
	assert.Nil(t, doc.GetEncryptionInfo())

	// When - 암호화 정보 설정
	metadata := map[string]interface{}{
		"algorithm":     "AES-256-GCM",
		"keyDerivation": "PBKDF2",
	}
	doc.SetEncryptionInfo("aes-gcm", "1.0", 256, metadata)

	// Then
	assert.True(t, doc.IsEncryptionEnabled())
	assert.Equal(t, "aes-gcm", doc.GetEncryptionType())
	assert.NotNil(t, doc.GetEncryptionInfo())
	assert.Equal(t, "aes-gcm", doc.Encryption.Type)
	assert.Equal(t, "1.0", doc.Encryption.Version)
	assert.Equal(t, 256, doc.Encryption.KeySize)
	assert.Equal(t, metadata, doc.Encryption.Metadata)

	// When - 암호화 해제
	doc.SetEncryptionInfo("none", "", 0, nil)
	assert.False(t, doc.IsEncryptionEnabled())
	assert.Nil(t, doc.Encryption)
}

func TestTimestamps(t *testing.T) {
	// Given
	doc := NewMongoStateDocument(uuid.New(), "Guild", 1, time.Now())
	originalCreatedAt := doc.CreatedAt

	// When - 타임스탬프 업데이트 (이미 생성된 경우)
	time.Sleep(1 * time.Millisecond) // 미세한 시간 차이를 위해
	doc.UpdateTimestamps()

	// Then
	assert.Equal(t, originalCreatedAt, doc.CreatedAt)      // CreatedAt은 변경되지 않음
	assert.True(t, doc.UpdatedAt.After(originalCreatedAt)) // UpdatedAt은 변경됨

	// When - 액세스 시간 터치
	assert.Nil(t, doc.LastAccessedAt)
	doc.TouchAccessTime()
	assert.NotNil(t, doc.LastAccessedAt)
	assert.True(t, doc.LastAccessedAt.After(originalCreatedAt))

	// When - TTL 설정
	assert.Nil(t, doc.TTL)
	doc.SetTTL(24 * time.Hour)
	assert.NotNil(t, doc.TTL)
	assert.True(t, doc.TTL.After(time.Now()))

	// When - 만료 여부 확인
	assert.False(t, doc.IsExpired())

	// When - 나이 확인
	age := doc.GetAge()
	assert.Greater(t, age, time.Duration(0))
}

func TestMetadata(t *testing.T) {
	// Given
	doc := NewMongoStateDocument(uuid.New(), "Guild", 1, time.Now())

	// When - 시스템 메타데이터 설정
	doc.SetSystemMetadata("source", "GuildService")
	doc.SetSystemMetadata("environment", "production")
	doc.SetSystemMetadata("version", 1.5)

	// Then
	source, exists := doc.GetSystemMetadata("source")
	assert.True(t, exists)
	assert.Equal(t, "GuildService", source)

	env, exists := doc.GetSystemMetadata("environment")
	assert.True(t, exists)
	assert.Equal(t, "production", env)

	version, exists := doc.GetSystemMetadata("version")
	assert.True(t, exists)
	assert.Equal(t, 1.5, version)

	nonExistent, exists := doc.GetSystemMetadata("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, nonExistent)
}

func TestTags(t *testing.T) {
	// Given
	doc := NewMongoStateDocument(uuid.New(), "Guild", 1, time.Now())

	// When - 태그 추가
	doc.AddTag("active")
	doc.AddTag("premium")
	doc.AddTag("large-guild")

	// Then
	assert.True(t, doc.HasTag("active"))
	assert.True(t, doc.HasTag("premium"))
	assert.True(t, doc.HasTag("large-guild"))
	assert.False(t, doc.HasTag("inactive"))
	assert.Len(t, doc.Tags, 3)

	// When - 중복 태그 추가 (무시됨)
	doc.AddTag("active")
	assert.Len(t, doc.Tags, 3)

	// When - 태그 제거
	doc.RemoveTag("premium")
	assert.False(t, doc.HasTag("premium"))
	assert.Len(t, doc.Tags, 2)

	// When - 존재하지 않는 태그 제거 (무시됨)
	doc.RemoveTag("nonexistent")
	assert.Len(t, doc.Tags, 2)
}

func TestPerformanceMetrics(t *testing.T) {
	// Given
	doc := NewMongoStateDocument(uuid.New(), "Guild", 1, time.Now())

	// When - 처리 시간 기록
	doc.RecordProcessingTime(125 * time.Millisecond)
	assert.Equal(t, int64(125), doc.ProcessingTimeMs)

	// When - 재시도 횟수 증가
	assert.Equal(t, 0, doc.RetryCount)
	doc.IncrementRetryCount()
	doc.IncrementRetryCount()
	assert.Equal(t, 2, doc.RetryCount)

	// When - 에러 횟수 증가
	assert.Equal(t, 0, doc.ErrorCount)
	assert.Equal(t, "", doc.LastError)

	doc.IncrementErrorCount()
	assert.Equal(t, 1, doc.ErrorCount)

	// When - 에러 메시지 설정
	testErr := fmt.Errorf("test error message")
	doc.SetLastError(testErr)
	assert.Equal(t, "test error message", doc.LastError)
	assert.Equal(t, 2, doc.ErrorCount) // SetLastError가 IncrementErrorCount 호출

	// When - 에러 초기화
	doc.SetLastError(nil)
	assert.Equal(t, "", doc.LastError)
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name      string
		setupDoc  func() *mongoStateDocument
		expectErr bool
		errMsg    string
	}{
		{
			name: "Valid document",
			setupDoc: func() *mongoStateDocument {
				return NewMongoStateDocument(uuid.New(), "Guild", 1, time.Now())
			},
			expectErr: false,
		},
		{
			name: "Empty ID",
			setupDoc: func() *mongoStateDocument {
				doc := NewMongoStateDocument(uuid.New(), "Guild", 1, time.Now())
				doc.ID = ""
				return doc
			},
			expectErr: true,
			errMsg:    "document ID cannot be empty",
		},
		{
			name: "Empty aggregate ID",
			setupDoc: func() *mongoStateDocument {
				doc := NewMongoStateDocument(uuid.New(), "Guild", 1, time.Now())
				doc.string = ""
				return doc
			},
			expectErr: true,
			errMsg:    "aggregate ID cannot be empty",
		},
		{
			name: "Empty aggregate type",
			setupDoc: func() *mongoStateDocument {
				doc := NewMongoStateDocument(uuid.New(), "Guild", 1, time.Now())
				doc.AggregateType = ""
				return doc
			},
			expectErr: true,
			errMsg:    "aggregate type cannot be empty",
		},
		{
			name: "Negative version",
			setupDoc: func() *mongoStateDocument {
				doc := NewMongoStateDocument(uuid.New(), "Guild", 1, time.Now())
				doc.Version = -1
				return doc
			},
			expectErr: true,
			errMsg:    "version cannot be negative",
		},
		{
			name: "Negative size",
			setupDoc: func() *mongoStateDocument {
				doc := NewMongoStateDocument(uuid.New(), "Guild", 1, time.Now())
				return doc
			},
			expectErr: true,
			errMsg:    "size values cannot be negative",
		},
		{
			name: "Invalid compression ratio",
			setupDoc: func() *mongoStateDocument {
				doc := NewMongoStateDocument(uuid.New(), "Guild", 1, time.Now())
				return doc
			},
			expectErr: true,
			errMsg:    "compression ratio must be between 0 and 1",
		},
		{
			name: "Empty compression type when compression info set",
			setupDoc: func() *mongoStateDocument {
				doc := NewMongoStateDocument(uuid.New(), "Guild", 1, time.Now())
				doc.Compression = &CompressionInfo{Type: ""}
				return doc
			},
			expectErr: true,
			errMsg:    "compression type cannot be empty when compression info is set",
		},
		{
			name: "Invalid compression level",
			setupDoc: func() *mongoStateDocument {
				doc := NewMongoStateDocument(uuid.New(), "Guild", 1, time.Now())
				doc.SetCompressionInfo("gzip", "1.0", 15, nil)
				return doc
			},
			expectErr: true,
			errMsg:    "compression level must be between 0 and 9",
		},
		{
			name: "Invalid encryption key size",
			setupDoc: func() *mongoStateDocument {
				doc := NewMongoStateDocument(uuid.New(), "Guild", 1, time.Now())
				doc.SetEncryptionInfo("aes-gcm", "1.0", 64, nil)
				return doc
			},
			expectErr: true,
			errMsg:    "encryption key size must be 128, 192, or 256",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			doc := tt.setupDoc()

			// When
			err := doc.Validate()

			// Then
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClone(t *testing.T) {
	// Given
	doc := NewMongoStateDocument(uuid.New(), "Guild", 5, time.Now())
	doc.Data = []byte("test data")
	doc.SetCompressionInfo("gzip", "1.0", 6, map[string]interface{}{"test": "value"})
	doc.SetEncryptionInfo("aes-gcm", "1.0", 256, map[string]interface{}{"key": "value"})
	doc.SetSystemMetadata("source", "test")
	doc.AddTag("test-tag")

	// When
	cloned := doc.Clone()

	// Then
	assert.Equal(t, doc.ID, cloned.ID)
	assert.Equal(t, doc.string, cloned.string)
	assert.Equal(t, doc.AggregateType, cloned.AggregateType)
	assert.Equal(t, doc.Version, cloned.Version)
	assert.Equal(t, doc.IsCompressionEnabled(), cloned.IsCompressionEnabled())
	assert.Equal(t, doc.IsEncryptionEnabled(), cloned.IsEncryptionEnabled())

	// 독립성 확인 (원본 수정이 복사본에 영향 주지 않음)
	doc.AddTag("new-tag")
	assert.True(t, doc.HasTag("new-tag"))
	assert.False(t, cloned.HasTag("new-tag"))
}

func TestGetSummary(t *testing.T) {
	// Given
	doc := NewMongoStateDocument(uuid.New(), "Guild", 5, time.Now())
	doc.SetCompressionInfo("gzip", "1.0", 6, nil)
	doc.SetEncryptionInfo("aes-gcm", "1.0", 256, nil)
	doc.RecordProcessingTime(125 * time.Millisecond)
	doc.IncrementRetryCount()
	doc.IncrementErrorCount()

	// When
	summary := doc.GetSummary()

	// Then
	assert.Equal(t, doc.ID, summary["id"])
	assert.Equal(t, doc.AggregateType, summary["aggregateType"])
	assert.Equal(t, doc.Version, summary["version"])
	assert.Equal(t, true, summary["compressed"])
	assert.Equal(t, true, summary["encrypted"])
	assert.Equal(t, "125ms", summary["processingTime"])
	assert.Equal(t, 1, summary["retryCount"])
	assert.Equal(t, 1, summary["errorCount"])
	assert.Contains(t, summary["age"], "") // age는 문자열로 변환됨
}

func TestJSONSerialization(t *testing.T) {
	// Given
	doc := NewMongoStateDocument(uuid.New(), "Guild", 5, time.Now())
	doc.Data = map[string]interface{}{
		"name":  "Test Guild",
		"level": 10,
	}
	doc.SetCompressionInfo("gzip", "1.0", 6, map[string]interface{}{"test": "value"})
	doc.SetSystemMetadata("source", "test")
	doc.AddTag("test-tag")

	// When - JSON 직렬화
	jsonData, err := json.Marshal(doc)
	require.NoError(t, err)

	// When - JSON 역직렬화
	var newDoc mongoStateDocument
	err = json.Unmarshal(jsonData, &newDoc)
	require.NoError(t, err)

	// Then
	assert.Equal(t, doc.ID, newDoc.ID)
	assert.Equal(t, doc.string, newDoc.string)
	assert.Equal(t, doc.AggregateType, newDoc.AggregateType)
	assert.Equal(t, doc.Version, newDoc.Version)
	assert.Equal(t, doc.IsCompressionEnabled(), newDoc.IsCompressionEnabled())
	assert.Equal(t, doc.GetCompressionType(), newDoc.GetCompressionType())
}

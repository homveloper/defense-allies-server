// mongo_state_document.go - MongoDB 상태 문서 구조체
package cqrsx

import (
	"time"
)

// CompressionInfo는 압축 관련 정보를 담습니다
type CompressionInfo struct {
	// 압축 알고리즘 타입 ("gzip", "lz4", "none")
	// "none"이면 압축 미적용으로 간주
	// 예시: "gzip", "lz4", "none"
	Type string `bson:"type" json:"type"`

	// 압축 알고리즘 버전 ("1.0", "2.1")
	// 예시: "1.0"
	Version string `bson:"version" json:"version"`

	// 압축 레벨 (1-9, 알고리즘별로 다름)
	// 예시: gzip의 경우 1(빠름) ~ 9(최고압축)
	Level int `bson:"level" json:"level"`

	// 압축 관련 메타데이터
	// 예시: {"windowSize": 15, "memLevel": 8, "strategy": "default"}
	Metadata map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// EncryptionInfo는 암호화 관련 정보를 담습니다
type EncryptionInfo struct {
	// 암호화 알고리즘 타입 ("aes-gcm", "aes-cbc", "chacha20-poly1305", "none")
	// "none"이면 암호화 미적용으로 간주
	// 예시: "aes-gcm", "none"
	Type string `bson:"type" json:"type"`

	// 암호화 알고리즘 버전 ("1.0", "2.0")
	// 예시: "1.0"
	Version string `bson:"version" json:"version"`

	// 키 크기 (128, 192, 256 비트)
	// 예시: 256
	KeySize int `bson:"keySize" json:"keySize"`

	// 암호화 관련 메타데이터 (nonce, salt, iterations 등)
	// 예시: {"algorithm": "AES-256-GCM", "keyDerivation": "PBKDF2", "iterations": 100000}
	Metadata map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// mongoStateDocument는 MongoDB에 저장되는 집합체 상태 문서입니다
type mongoStateDocument struct {
	// ================================
	// 🏷️ 식별자 그룹
	// ================================

	// MongoDB 문서의 고유 식별자
	// 형식: "{aggregateType}-{aggregateId}-v{version}"
	// 예시: "Guild-123e4567-e89b-12d3-a456-426614174000-v5"
	ID string `bson:"_id" json:"id"`

	// 집합체의 고유 식별자 (UUID 문자열)
	// 예시: "123e4567-e89b-12d3-a456-426614174000"
	AggregateID string `bson:"aggregateId" json:"aggregateId"`

	// 집합체의 타입 (도메인 엔티티 이름)
	// 예시: "Guild", "User", "Order"
	AggregateType string `bson:"aggregateType" json:"aggregateType"`

	// 집합체 상태의 버전 (이벤트 기반 버전)
	// 예시: 5 (5번째 이벤트까지 적용된 상태)
	Version int `bson:"version" json:"version"`

	// 문서 스키마의 버전 (하위 호환성 관리용)
	// 예시: 2 (현재 스키마 버전)
	DocumentVersion int `bson:"documentVersion" json:"documentVersion"`

	// ================================
	// 📦 데이터 그룹
	// ================================

	// 실제 집합체 상태 데이터 (압축/암호화 적용 후)
	// 타입: []byte (base64 인코딩) 또는 interface{} (원본)
	// 예시: "H4sIAAAAAAAAA..." (base64 인코딩된 압축 데이터)
	Data interface{} `bson:"data" json:"data"`

	// 원본 데이터의 형식
	// 예시: "json", "binary", "protobuf"
	DataFormat string `bson:"dataFormat" json:"dataFormat"`

	// 저장된 데이터의 인코딩 방식
	// 예시: "base64", "hex", "raw"
	DataEncoding string `bson:"dataEncoding" json:"dataEncoding"`

	// ================================
	// 🗜️ 압축 정보 (nullable)
	// ================================

	// 압축 관련 정보
	// null이거나 Type="none"이면 압축 미적용
	Compression *CompressionInfo `bson:"compression,omitempty" json:"compression,omitempty"`

	// ================================
	// 🔐 암호화 정보 (nullable)
	// ================================

	// 암호화 관련 정보
	// null이거나 Type="none"이면 암호화 미적용
	Encryption *EncryptionInfo `bson:"encryption,omitempty" json:"encryption,omitempty"`

	// ================================
	// 🕐 시간 정보
	// ================================

	// 집합체 상태의 비즈니스 타임스탬프 (도메인 이벤트 발생 시각)
	// 예시: "2025-06-08T10:30:00Z"
	StateTimestamp time.Time `bson:"stateTimestamp" json:"stateTimestamp"`

	// 문서가 MongoDB에 생성된 시각
	// 예시: "2025-06-08T10:30:01Z"
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`

	// 문서가 마지막으로 수정된 시각
	// 예시: "2025-06-08T10:30:01Z"
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`

	// 마지막으로 이 상태가 조회된 시각 (캐시/TTL 관리용)
	// 예시: "2025-06-08T11:45:30Z"
	LastAccessedAt *time.Time `bson:"lastAccessedAt,omitempty" json:"lastAccessedAt,omitempty"`

	// 자동 삭제 시각 (TTL 인덱스용)
	// 예시: "2025-07-08T10:30:00Z" (30일 후 자동 삭제)
	TTL *time.Time `bson:"ttl,omitempty" json:"ttl,omitempty"`

	// ================================
	// 📊 메타데이터 그룹
	// ================================

	// 시스템 관련 메타데이터 (운영 정보)
	// 예시: {"source": "GuildService", "environment": "production", "host": "app-server-01"}
	SystemMetadata map[string]interface{} `bson:"systemMetadata,omitempty" json:"systemMetadata,omitempty"`

	// 검색 및 분류용 태그
	// 예시: ["active", "large-guild", "premium"]
	Tags []string `bson:"tags,omitempty" json:"tags,omitempty"`

	// ================================
	// ⚡ 성능 및 운영 정보
	// ================================

	// 상태 저장 처리 시간 (밀리초)
	// 예시: 125 (125ms 소요)
	ProcessingTimeMs int64 `bson:"processingTimeMs" json:"processingTimeMs"`

	// 저장 재시도 횟수
	// 예시: 0 (첫 번째 시도에서 성공)
	RetryCount int `bson:"retryCount" json:"retryCount"`

	// 에러 발생 횟수
	// 예시: 0 (에러 없음)
	ErrorCount int `bson:"errorCount" json:"errorCount"`

	// 마지막 에러 메시지 (디버깅용)
	// 예시: "compression failed: invalid data format"
	LastError string `bson:"lastError,omitempty" json:"lastError,omitempty"`
}

// DocumentVersionCurrent는 현재 문서 스키마 버전입니다
const DocumentVersionCurrent = 2

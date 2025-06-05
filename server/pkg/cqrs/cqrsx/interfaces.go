package cqrsx

import (
	"context"
	"defense-allies-server/pkg/cqrs"
)

// EventStore 이벤트 저장소 인터페이스
type EventStore interface {
	// SaveEvents 이벤트들을 저장
	SaveEvents(ctx context.Context, aggregateID string, events []cqrs.EventMessage, expectedVersion int) error
	
	// LoadEvents 이벤트들을 로드 (fromVersion부터 toVersion까지, toVersion이 0이면 끝까지)
	LoadEvents(ctx context.Context, aggregateID, aggregateType string, fromVersion, toVersion int) ([]cqrs.EventMessage, error)
	
	// GetEventHistory 이벤트 히스토리 조회
	GetEventHistory(ctx context.Context, aggregateID, aggregateType string, fromVersion int) ([]cqrs.EventMessage, error)
	
	// GetLastEventVersion 마지막 이벤트 버전 조회
	GetLastEventVersion(ctx context.Context, aggregateID, aggregateType string) (int, error)
}

// ReadStore 읽기 저장소 인터페이스
type ReadStore interface {
	// Save 읽기 모델 저장
	Save(ctx context.Context, readModel interface{}) error
	
	// GetByID ID로 읽기 모델 조회
	GetByID(ctx context.Context, id string, result interface{}) error
	
	// Query 쿼리로 읽기 모델 조회
	Query(ctx context.Context, filter interface{}, result interface{}) error
	
	// Delete 읽기 모델 삭제
	Delete(ctx context.Context, id string) error
}

// StateStore 상태 저장소 인터페이스
type StateStore interface {
	// SaveState 상태 저장
	SaveState(ctx context.Context, key string, state interface{}) error
	
	// GetState 상태 조회
	GetState(ctx context.Context, key string, result interface{}) error
	
	// DeleteState 상태 삭제
	DeleteState(ctx context.Context, key string) error
	
	// Exists 상태 존재 여부 확인
	Exists(ctx context.Context, key string) (bool, error)
}

// Repository 리포지토리 인터페이스
type Repository interface {
	// Save Aggregate 저장
	Save(ctx context.Context, aggregate interface{}) error
	
	// GetByID ID로 Aggregate 조회
	GetByID(ctx context.Context, id string) (interface{}, error)
	
	// GetEventHistory 이벤트 히스토리 조회
	GetEventHistory(ctx context.Context, id string) ([]cqrs.EventMessage, error)
}

// ClientManager 클라이언트 매니저 인터페이스
type ClientManager interface {
	// Close 연결 종료
	Close(ctx context.Context) error
	
	// Ping 연결 확인
	Ping(ctx context.Context) error
	
	// GetStats 통계 조회
	GetStats(ctx context.Context) (map[string]interface{}, error)
}

// state_store_factory.go - 상태 저장소 팩토리
package cqrsx

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

// StateStoreFactory는 상태 저장소 생성을 담당합니다
type StateStoreFactory struct {
	client *mongo.Client
	dbName string
}

// NewStateStoreFactory는 새로운 상태 저장소 팩토리를 생성합니다
func NewStateStoreFactory(client *mongo.Client, dbName string) *StateStoreFactory {
	return &StateStoreFactory{
		client: client,
		dbName: dbName,
	}
}

// CreateMongoStateStore는 MongoDB 상태 저장소를 생성합니다
func (f *StateStoreFactory) CreateMongoStateStore(collectionName string, options ...StateStoreOption) StateStore {
	collection := f.client.Database(f.dbName).Collection(collectionName)
	return NewMongoStateStore(collection, f.client, options...)
}

// CreateBasicStateStore는 기본 설정의 상태 저장소를 생성합니다
func (f *StateStoreFactory) CreateBasicStateStore(collectionName string) StateStore {
	return f.CreateMongoStateStore(collectionName, WithIndexing())
}

// CreateProductionStateStore는 프로덕션 환경용 상태 저장소를 생성합니다
func (f *StateStoreFactory) CreateProductionStateStore(collectionName, encryptionKey string) StateStore {
	return f.CreateMongoStateStore(
		collectionName,
		WithCompression(CompressionGzip),
		WithEncryption(NewAESEncryptor(encryptionKey)),
		WithRetentionPolicy(KeepLast(10)), // 최신 10개 버전 보존
		WithBatchSize(100),
		WithMetrics(),
		WithIndexing(),
	)
}

// CreateHighPerformanceStateStore는 고성능 상태 저장소를 생성합니다
func (f *StateStoreFactory) CreateHighPerformanceStateStore(collectionName string) StateStore {
	return f.CreateMongoStateStore(
		collectionName,
		WithCompression(CompressionLZ4), // 빠른 압축
		WithBatchSize(200),
		WithMetrics(),
		WithIndexing(),
	)
}

// CreateDevelopmentStateStore는 개발 환경용 상태 저장소를 생성합니다
func (f *StateStoreFactory) CreateDevelopmentStateStore(collectionName string) StateStore {
	return f.CreateMongoStateStore(
		collectionName,
		WithMetrics(),
		WithIndexing(),
		// 개발 환경에서는 압축/암호화 없이 빠른 개발을 위해
	)
}

// StateStoreBuilder는 빌더 패턴으로 상태 저장소를 구성합니다
type StateStoreBuilder struct {
	factory     *StateStoreFactory
	collection  string
	options     []StateStoreOption
}

// NewStateStoreBuilder는 새로운 빌더를 생성합니다
func NewStateStoreBuilder(factory *StateStoreFactory, collectionName string) *StateStoreBuilder {
	return &StateStoreBuilder{
		factory:    factory,
		collection: collectionName,
		options:    make([]StateStoreOption, 0),
	}
}

// WithGzipCompression은 GZIP 압축을 추가합니다
func (b *StateStoreBuilder) WithGzipCompression() *StateStoreBuilder {
	b.options = append(b.options, WithCompression(CompressionGzip))
	return b
}

// WithLZ4Compression은 LZ4 압축을 추가합니다
func (b *StateStoreBuilder) WithLZ4Compression() *StateStoreBuilder {
	b.options = append(b.options, WithCompression(CompressionLZ4))
	return b
}

// WithAESEncryption은 AES 암호화를 추가합니다
func (b *StateStoreBuilder) WithAESEncryption(passphrase string) *StateStoreBuilder {
	b.options = append(b.options, WithEncryption(NewAESEncryptor(passphrase)))
	return b
}

// WithKeepLastPolicy는 최신 N개 보존 정책을 추가합니다
func (b *StateStoreBuilder) WithKeepLastPolicy(count int) *StateStoreBuilder {
	b.options = append(b.options, WithRetentionPolicy(KeepLast(count)))
	return b
}

// WithTimePolicyDays는 N일 보존 정책을 추가합니다
func (b *StateStoreBuilder) WithTimePolicyDays(days int) *StateStoreBuilder {
	b.options = append(b.options, WithRetentionPolicy(KeepForDuration(24*time.Hour*time.Duration(days))))
	return b
}

// WithSizePolicyMB는 크기 기반 보존 정책을 추가합니다 (MB 단위)
func (b *StateStoreBuilder) WithSizePolicyMB(sizeMB int64) *StateStoreBuilder {
	b.options = append(b.options, WithRetentionPolicy(KeepWithinSize(sizeMB*1024*1024)))
	return b
}

// WithPerformanceOptimizations는 성능 최적화 옵션을 추가합니다
func (b *StateStoreBuilder) WithPerformanceOptimizations() *StateStoreBuilder {
	b.options = append(b.options, WithBatchSize(150))
	b.options = append(b.options, WithMetrics())
	b.options = append(b.options, WithIndexing())
	return b
}

// Build는 설정된 옵션으로 상태 저장소를 생성합니다
func (b *StateStoreBuilder) Build() StateStore {
	return b.factory.CreateMongoStateStore(b.collection, b.options...)
}

// StateStoreManager는 여러 상태 저장소를 관리합니다
type StateStoreManager struct {
	factory *StateStoreFactory
	stores  map[string]StateStore
}

// NewStateStoreManager는 새로운 관리자를 생성합니다
func NewStateStoreManager(factory *StateStoreFactory) *StateStoreManager {
	return &StateStoreManager{
		factory: factory,
		stores:  make(map[string]StateStore),
	}
}

// GetOrCreate는 이름으로 저장소를 가져오거나 생성합니다
func (m *StateStoreManager) GetOrCreate(name string, creator func(*StateStoreFactory) StateStore) StateStore {
	if store, exists := m.stores[name]; exists {
		return store
	}
	
	store := creator(m.factory)
	m.stores[name] = store
	return store
}

// GetGuildStore는 길드용 저장소를 반환합니다
func (m *StateStoreManager) GetGuildStore() StateStore {
	return m.GetOrCreate("guilds", func(f *StateStoreFactory) StateStore {
		return f.CreateProductionStateStore("guild_states", "guild-encryption-key")
	})
}

// GetUserStore는 사용자용 저장소를 반환합니다
func (m *StateStoreManager) GetUserStore() StateStore {
	return m.GetOrCreate("users", func(f *StateStoreFactory) StateStore {
		return f.CreateProductionStateStore("user_states", "user-encryption-key")
	})
}

// GetOrderStore는 주문용 저장소를 반환합니다
func (m *StateStoreManager) GetOrderStore() StateStore {
	return m.GetOrCreate("orders", func(f *StateStoreFactory) StateStore {
		return f.CreateHighPerformanceStateStore("order_states")
	})
}

// CloseAll은 모든 저장소를 정리합니다
func (m *StateStoreManager) CloseAll() error {
	var lastErr error
	for name, store := range m.stores {
		if err := store.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close store %s: %w", name, err)
		}
	}
	return lastErr
}

// StateStoreHealthChecker는 저장소 상태를 확인합니다
type StateStoreHealthChecker struct {
	stores map[string]StateStore
}

// NewStateStoreHealthChecker는 새로운 헬스 체커를 생성합니다
func NewStateStoreHealthChecker() *StateStoreHealthChecker {
	return &StateStoreHealthChecker{
		stores: make(map[string]StateStore),
	}
}

// AddStore는 체크할 저장소를 추가합니다
func (h *StateStoreHealthChecker) AddStore(name string, store StateStore) {
	h.stores[name] = store
}

// HealthCheck는 모든 저장소의 상태를 확인합니다
func (h *StateStoreHealthChecker) HealthCheck(ctx context.Context) map[string]bool {
	results := make(map[string]bool)
	
	for name, store := range h.stores {
		// 간단한 카운트 쿼리로 연결 상태 확인
		if queryStore, ok := store.(QueryableStateStore); ok {
			_, err := queryStore.CountByQuery(ctx, StateQuery{Limit: 1})
			results[name] = err == nil
		} else {
			// 기본 저장소는 더미 UUID로 존재 여부 확인
			dummyID := uuid.New()
			_, err := store.Exists(ctx, dummyID, 1)
			results[name] = err == nil
		}
	}
	
	return results
}

// IsAllHealthy는 모든 저장소가 정상인지 확인합니다
func (h *StateStoreHealthChecker) IsAllHealthy(ctx context.Context) bool {
	results := h.HealthCheck(ctx)
	for _, healthy := range results {
		if !healthy {
			return false
		}
	}
	return true
}

// 편의 함수들

// QuickMongo는 빠른 MongoDB 저장소 생성을 위한 헬퍼입니다
func QuickMongo(client *mongo.Client, dbName, collectionName string) StateStore {
	factory := NewStateStoreFactory(client, dbName)
	return factory.CreateBasicStateStore(collectionName)
}

// QuickMongoWithEncryption은 암호화가 포함된 MongoDB 저장소를 빠르게 생성합니다
func QuickMongoWithEncryption(client *mongo.Client, dbName, collectionName, encryptionKey string) StateStore {
	factory := NewStateStoreFactory(client, dbName)
	return factory.CreateProductionStateStore(collectionName, encryptionKey)
}

// QuickBuilder는 빠른 빌더 생성을 위한 헬퍼입니다
func QuickBuilder(client *mongo.Client, dbName, collectionName string) *StateStoreBuilder {
	factory := NewStateStoreFactory(client, dbName)
	return NewStateStoreBuilder(factory, collectionName)
}

// factory.go - 이벤트 저장소 팩토리와 설정
package cqrsx

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EventStoreFactory는 이벤트 저장소를 생성하고 관리합니다
type EventStoreFactory struct {
	config EventStoreConfig
	client *mongo.Client
	stores map[StorageStrategy]EventStore
}

// EventStoreConfig는 이벤트 저장소 설정입니다
type EventStoreConfig struct {
	// 기본 설정
	Strategy     StorageStrategy `json:"strategy" yaml:"strategy"`
	DatabaseName string          `json:"databaseName" yaml:"databaseName"`

	// MongoDB 설정
	MongoDB MongoConfig `json:"mongodb" yaml:"mongodb"`

	// Redis 설정 (캐싱용)
	Redis RedisConfig `json:"redis" yaml:"redis"`

	// 하이브리드 설정
	Hybrid HybridConfig `json:"hybrid" yaml:"hybrid"`

	// 성능 설정
	Performance PerformanceConfig `json:"performance" yaml:"performance"`

	// 모니터링 설정
	Monitoring MonitoringConfig `json:"monitoring" yaml:"monitoring"`
}

// MongoConfig는 MongoDB 연결 설정입니다
type MongoConfig struct {
	URI                string        `json:"uri" yaml:"uri"`
	ConnectTimeout     time.Duration `json:"connectTimeout" yaml:"connectTimeout"`
	MaxPoolSize        int           `json:"maxPoolSize" yaml:"maxPoolSize"`
	MinPoolSize        int           `json:"minPoolSize" yaml:"minPoolSize"`
	MaxConnIdleTime    time.Duration `json:"maxConnIdleTime" yaml:"maxConnIdleTime"`
	EventCollection    string        `json:"eventCollection" yaml:"eventCollection"`
	SnapshotCollection string        `json:"snapshotCollection" yaml:"snapshotCollection"`
}

// RedisConfig는 Redis 연결 설정입니다
type RedisConfig struct {
	Address     string        `json:"address" yaml:"address"`
	Password    string        `json:"password" yaml:"password"`
	DB          int           `json:"db" yaml:"db"`
	PoolSize    int           `json:"poolSize" yaml:"poolSize"`
	IdleTimeout time.Duration `json:"idleTimeout" yaml:"idleTimeout"`
	EnableCache bool          `json:"enableCache" yaml:"enableCache"`
}

// PerformanceConfig는 성능 관련 설정입니다
type PerformanceConfig struct {
	BatchSize         int           `json:"batchSize" yaml:"batchSize"`
	MaxEventSize      int           `json:"maxEventSize" yaml:"maxEventSize"`
	CacheSize         int           `json:"cacheSize" yaml:"cacheSize"`
	CacheTTL          time.Duration `json:"cacheTTL" yaml:"cacheTTL"`
	EnableCompression bool          `json:"enableCompression" yaml:"enableCompression"`
	EnableMetrics     bool          `json:"enableMetrics" yaml:"enableMetrics"`
}

// MonitoringConfig는 모니터링 설정입니다
type MonitoringConfig struct {
	EnableLogging   bool          `json:"enableLogging" yaml:"enableLogging"`
	EnableTracing   bool          `json:"enableTracing" yaml:"enableTracing"`
	MetricsInterval time.Duration `json:"metricsInterval" yaml:"metricsInterval"`
	HealthCheckPort int           `json:"healthCheckPort" yaml:"healthCheckPort"`
}

// NewEventStoreFactory는 새로운 이벤트 저장소 팩토리를 생성합니다
func NewEventStoreFactory(config EventStoreConfig) (*EventStoreFactory, error) {
	// 설정 검증
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// MongoDB 클라이언트 생성
	client, err := createMongoClient(config.MongoDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create mongo client: %w", err)
	}

	factory := &EventStoreFactory{
		config: config,
		client: client,
		stores: make(map[StorageStrategy]EventStore),
	}

	return factory, nil
}

// Create는 지정된 전략의 이벤트 저장소를 생성합니다
func (f *EventStoreFactory) Create(strategy StorageStrategy) (EventStore, error) {
	// 캐시된 인스턴스 확인
	if store, exists := f.stores[strategy]; exists {
		return store, nil
	}

	// 새 인스턴스 생성
	store, err := f.createStore(strategy)
	if err != nil {
		return nil, err
	}

	// 캐시에 저장
	f.stores[strategy] = store

	return store, nil
}

// GetDefault는 기본 설정의 이벤트 저장소를 반환합니다
func (f *EventStoreFactory) GetDefault() (EventStore, error) {
	return f.Create(f.config.Strategy)
}

// GetQueryableStore는 복잡한 쿼리를 지원하는 저장소를 반환합니다
func (f *EventStoreFactory) GetQueryableStore() (QueryableEventStore, error) {
	switch f.config.Strategy {
	case StrategyDocument:
		store, err := f.Create(StrategyDocument)
		if err != nil {
			return nil, err
		}
		return store.(*DocumentEventStore), nil

	case StrategyHybrid:
		store, err := f.Create(StrategyHybrid)
		if err != nil {
			return nil, err
		}
		return store.(*HybridEventStore), nil

	default:
		store, err := f.Create(StrategyDocument)
		if err != nil {
			return nil, err
		}
		return store.(*DocumentEventStore), nil
	}
}

// Migrate는 한 저장소에서 다른 저장소로 데이터를 마이그레이션합니다
func (f *EventStoreFactory) Migrate(ctx context.Context, from, to StorageStrategy) error {
	sourceStore, err := f.Create(from)
	if err != nil {
		return fmt.Errorf("failed to create source store: %w", err)
	}

	targetStore, err := f.Create(to)
	if err != nil {
		return fmt.Errorf("failed to create target store: %w", err)
	}

	// 마이그레이션 실행
	migrator := NewEventStoreMigrator(sourceStore, targetStore)
	return migrator.Migrate(ctx)
}

// Close는 모든 연결을 정리합니다
func (f *EventStoreFactory) Close() error {
	// 모든 저장소 닫기
	for _, store := range f.stores {
		if err := store.Close(); err != nil {
			return err
		}
	}

	// MongoDB 클라이언트 닫기
	if f.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return f.client.Disconnect(ctx)
	}

	return nil
}

// Private methods

func (f *EventStoreFactory) createStore(strategy StorageStrategy) (EventStore, error) {
	database := f.client.Database(f.config.DatabaseName)
	eventCollection := database.Collection(f.config.MongoDB.EventCollection)
	serializer := NewJSONEventSerializer()

	switch strategy {
	case StrategyStream:
		store := NewStreamEventStore(eventCollection, serializer)

		// 인덱스 생성
		if err := store.CreateIndexes(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to create indexes: %w", err)
		}

		return store, nil

	case StrategyDocument:
		store := NewDocumentEventStore(eventCollection, f.client, serializer)

		// 인덱스 생성
		if err := store.CreateIndexes(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to create indexes: %w", err)
		}

		return store, nil

	case StrategyHybrid:
		// Hot 스토어 (Stream 방식)
		hotStore := NewStreamEventStore(eventCollection, serializer)

		// Cold 스토어 (Document 방식)
		coldCollection := database.Collection(f.config.MongoDB.EventCollection + "_archive")
		coldStore := NewDocumentEventStore(coldCollection, f.client, serializer)

		// 인덱스 생성
		if err := hotStore.CreateIndexes(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to create hot store indexes: %w", err)
		}
		if err := coldStore.CreateIndexes(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to create cold store indexes: %w", err)
		}

		return NewHybridEventStore(hotStore, coldStore, f.config.Hybrid), nil

	default:
		return nil, fmt.Errorf("unsupported storage strategy: %s", strategy)
	}
}

// Validate는 설정의 유효성을 검증합니다
func (c *EventStoreConfig) Validate() error {
	// 기본 설정 검증
	if c.DatabaseName == "" {
		return fmt.Errorf("database name is required")
	}

	// 전략별 검증
	switch c.Strategy {
	case StrategyStream, StrategyDocument, StrategyHybrid:
		// 유효한 전략
	default:
		return fmt.Errorf("invalid storage strategy: %s", c.Strategy)
	}

	// MongoDB 설정 검증
	if err := c.MongoDB.Validate(); err != nil {
		return fmt.Errorf("mongodb config: %w", err)
	}

	// 하이브리드 설정 검증 (하이브리드 전략인 경우)
	if c.Strategy == StrategyHybrid {
		if err := c.Hybrid.Validate(); err != nil {
			return fmt.Errorf("hybrid config: %w", err)
		}
	}

	return nil
}

// Validate는 MongoDB 설정의 유효성을 검증합니다
func (m *MongoConfig) Validate() error {
	if m.URI == "" {
		return fmt.Errorf("mongodb URI is required")
	}

	if m.EventCollection == "" {
		m.EventCollection = "events" // 기본값 설정
	}

	if m.SnapshotCollection == "" {
		m.SnapshotCollection = "snapshots" // 기본값 설정
	}

	return nil
}

// Validate는 하이브리드 설정의 유효성을 검증합니다
func (h *HybridConfig) Validate() error {
	if h.HotDataThreshold <= 0 {
		return fmt.Errorf("hot data threshold must be positive")
	}

	if h.ArchiveInterval <= 0 {
		return fmt.Errorf("archive interval must be positive")
	}

	if h.MaxHotEvents <= 0 {
		h.MaxHotEvents = 1000 // 기본값
	}

	return nil
}

// 기본 설정 생성 함수들

// DefaultEventStoreConfig는 기본 설정을 반환합니다
func DefaultEventStoreConfig() EventStoreConfig {
	return EventStoreConfig{
		Strategy:     StrategyStream,
		DatabaseName: "gameevents",
		MongoDB: MongoConfig{
			URI:                "mongodb://localhost:27017",
			ConnectTimeout:     10 * time.Second,
			MaxPoolSize:        100,
			MinPoolSize:        10,
			MaxConnIdleTime:    10 * time.Minute,
			EventCollection:    "events",
			SnapshotCollection: "snapshots",
		},
		Redis: RedisConfig{
			Address:     "localhost:6379",
			DB:          0,
			PoolSize:    10,
			IdleTimeout: 5 * time.Minute,
			EnableCache: true,
		},
		Hybrid: HybridConfig{
			HotDataThreshold:  30 * 24 * time.Hour, // 30일
			ArchiveInterval:   6 * time.Hour,       // 6시간마다
			MaxHotEvents:      10000,
			EnableAutoArchive: true,
		},
		Performance: PerformanceConfig{
			BatchSize:         100,
			MaxEventSize:      1024 * 1024, // 1MB
			CacheSize:         1000,
			CacheTTL:          5 * time.Minute,
			EnableCompression: false,
			EnableMetrics:     true,
		},
		Monitoring: MonitoringConfig{
			EnableLogging:   true,
			EnableTracing:   false,
			MetricsInterval: 30 * time.Second,
			HealthCheckPort: 8080,
		},
	}
}

// Helper functions

func createMongoClient(config MongoConfig) (*mongo.Client, error) {
	clientOptions := options.Client().
		ApplyURI(config.URI).
		SetConnectTimeout(config.ConnectTimeout).
		SetMaxPoolSize(uint64(config.MaxPoolSize)).
		SetMinPoolSize(uint64(config.MinPoolSize)).
		SetMaxConnIdleTime(config.MaxConnIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// 연결 테스트
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	return client, nil
}

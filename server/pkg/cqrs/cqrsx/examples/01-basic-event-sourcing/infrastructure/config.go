package infrastructure

import (
	"context"
	"cqrs/cqrsx"
	"fmt"
	"log"
	"time"
)

// Config 애플리케이션 설정
type Config struct {
	MongoDB *MongoDBConfig `json:"mongodb"`
	App     *AppConfig     `json:"app"`
}

// MongoDBConfig MongoDB 연결 설정
type MongoDBConfig struct {
	URI                    string        `json:"uri"`
	Database               string        `json:"database"`
	ConnectTimeout         time.Duration `json:"connect_timeout"`
	ServerSelectionTimeout time.Duration `json:"server_selection_timeout"`
	MaxPoolSize            int           `json:"max_pool_size"`
}

// AppConfig 애플리케이션 설정
type AppConfig struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	LogLevel    string `json:"log_level"`
}

// GetDefaultConfig 기본 설정 반환
func GetDefaultConfig() *Config {
	return &Config{
		MongoDB: &MongoDBConfig{
			URI:                    "mongodb://localhost:27017",
			Database:               "cqrs_basic_example",
			ConnectTimeout:         10 * time.Second,
			ServerSelectionTimeout: 5 * time.Second,
			MaxPoolSize:            100,
		},
		App: &AppConfig{
			Name:        "Basic Event Sourcing Example",
			Version:     "1.0.0",
			Environment: "development",
			LogLevel:    "info",
		},
	}
}

// GetTestConfig 테스트용 설정 반환
func GetTestConfig() *Config {
	config := GetDefaultConfig()
	config.MongoDB.Database = "cqrs_basic_example_test"
	config.App.Environment = "test"
	return config
}

// Infrastructure 인프라스트럭처 컴포넌트들
type Infrastructure struct {
	Config      *Config
	MongoClient *cqrsx.MongoClientManager
	EventStore  *cqrsx.MongoEventStore
	UserRepo    *UserRepository
}

// NewInfrastructure 인프라스트럭처 초기화
func NewInfrastructure(config *Config) (*Infrastructure, error) {
	if config == nil {
		config = GetDefaultConfig()
	}

	// MongoDB 클라이언트 생성
	mongoConfig := &cqrsx.MongoConfig{
		URI:                    config.MongoDB.URI,
		Database:               config.MongoDB.Database,
		ConnectTimeout:         config.MongoDB.ConnectTimeout,
		ServerSelectionTimeout: config.MongoDB.ServerSelectionTimeout,
		MaxPoolSize:            config.MongoDB.MaxPoolSize,
	}

	mongoClient, err := cqrsx.NewMongoClientManager(mongoConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create MongoDB client: %w", err)
	}

	// Event Store 생성
	eventStore := cqrsx.NewMongoEventStore(mongoClient, mongoClient.GetCollectionName("events"))

	// User Repository 생성
	userRepo := NewUserRepository(eventStore)

	return &Infrastructure{
		Config:      config,
		MongoClient: mongoClient,
		EventStore:  eventStore,
		UserRepo:    userRepo,
	}, nil
}

// Initialize 인프라스트럭처 초기화 (스키마 생성 등)
func (infra *Infrastructure) Initialize(ctx context.Context) error {
	log.Printf("Initializing infrastructure for %s...", infra.Config.App.Name)

	// MongoDB Event Sourcing 스키마 초기화
	err := infra.MongoClient.InitializeEventSourcingSchema(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize event sourcing schema: %w", err)
	}

	log.Printf("Infrastructure initialized successfully")
	return nil
}

// Close 인프라스트럭처 정리
func (infra *Infrastructure) Close(ctx context.Context) error {
	log.Printf("Closing infrastructure...")

	if infra.MongoClient != nil {
		err := infra.MongoClient.Close(ctx)
		if err != nil {
			log.Printf("Error closing MongoDB client: %v", err)
			return err
		}
	}

	log.Printf("Infrastructure closed successfully")
	return nil
}

// HealthCheck 인프라스트럭처 상태 확인
func (infra *Infrastructure) HealthCheck(ctx context.Context) error {
	// MongoDB 연결 상태 확인 (MongoClientManager에 Ping 메서드가 없으므로 간단한 확인)
	if infra.MongoClient == nil {
		return fmt.Errorf("MongoDB client is not initialized")
	}

	// 간단한 연결 테스트 (실제로는 MongoDB 드라이버의 Ping을 사용해야 함)
	log.Printf("MongoDB health check: client is initialized")
	return nil
}

// GetMetrics 인프라스트럭처 메트릭 반환
func (infra *Infrastructure) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// MongoDB 메트릭
	if infra.MongoClient != nil {
		mongoMetrics := infra.MongoClient.GetMetrics()
		metrics["mongodb"] = mongoMetrics
	}

	// 애플리케이션 메트릭
	metrics["app"] = map[string]interface{}{
		"name":        infra.Config.App.Name,
		"version":     infra.Config.App.Version,
		"environment": infra.Config.App.Environment,
	}

	return metrics
}

// ClearData 모든 데이터 삭제 (개발/테스트용)
func (infra *Infrastructure) ClearData(ctx context.Context) error {
	log.Printf("Clearing all data...")

	// Events 컬렉션 삭제
	eventsCollection := infra.MongoClient.GetCollection(infra.MongoClient.GetCollectionName("events"))
	err := eventsCollection.Drop(ctx)
	if err != nil {
		log.Printf("Warning: Failed to drop events collection: %v", err)
	}

	// Snapshots 컬렉션 삭제
	snapshotsCollection := infra.MongoClient.GetCollection(infra.MongoClient.GetCollectionName("snapshots"))
	err = snapshotsCollection.Drop(ctx)
	if err != nil {
		log.Printf("Warning: Failed to drop snapshots collection: %v", err)
	}

	// 스키마 재초기화
	err = infra.Initialize(ctx)
	if err != nil {
		return fmt.Errorf("failed to reinitialize schema: %w", err)
	}

	log.Printf("All data cleared successfully")
	return nil
}

// ValidateConfig 설정 검증
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.MongoDB == nil {
		return fmt.Errorf("mongodb config cannot be nil")
	}

	if config.MongoDB.URI == "" {
		return fmt.Errorf("mongodb URI cannot be empty")
	}

	if config.MongoDB.Database == "" {
		return fmt.Errorf("mongodb database cannot be empty")
	}

	if config.App == nil {
		return fmt.Errorf("app config cannot be nil")
	}

	if config.App.Name == "" {
		return fmt.Errorf("app name cannot be empty")
	}

	return nil
}

// SetupTestInfrastructure 테스트용 인프라스트럭처 설정
func SetupTestInfrastructure() (*Infrastructure, func(), error) {
	config := GetTestConfig()

	infra, err := NewInfrastructure(config)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()
	err = infra.Initialize(ctx)
	if err != nil {
		infra.Close(ctx)
		return nil, nil, err
	}

	cleanup := func() {
		ctx := context.Background()
		infra.ClearData(ctx)
		infra.Close(ctx)
	}

	return infra, cleanup, nil
}

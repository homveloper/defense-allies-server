package infrastructure

import (
	"cqrs"
	"cqrs/cqrsx"
	"fmt"
)

// ReadStoreType represents the type of read store
type ReadStoreType string

const (
	ReadStoreTypeMongoDB ReadStoreType = "mongodb"
	ReadStoreTypeRedis   ReadStoreType = "redis"
	ReadStoreTypeMemory  ReadStoreType = "memory"
)

// ReadStoreConfig contains configuration for read stores
type ReadStoreConfig struct {
	Type    ReadStoreType `json:"type"`
	MongoDB *MongoConfig  `json:"mongodb,omitempty"`
	Redis   *RedisConfig  `json:"redis,omitempty"`
	Memory  *MemoryConfig `json:"memory,omitempty"`
}

// MongoConfig contains MongoDB-specific configuration
type MongoConfig struct {
	ConnectionString string `json:"connection_string"`
	DatabaseName     string `json:"database_name"`
	CollectionName   string `json:"collection_name"`
}

// RedisConfig contains Redis-specific configuration
type RedisConfig struct {
	Address  string `json:"address"`
	Password string `json:"password"`
	Database int    `json:"database"`
}

// MemoryConfig contains in-memory store configuration
type MemoryConfig struct {
	MaxSize int `json:"max_size"`
}

// ReadStoreFactory creates read stores for different read models
type ReadStoreFactory struct {
	config ReadStoreConfig
}

// NewReadStoreFactory creates a new ReadStoreFactory
func NewReadStoreFactory(config ReadStoreConfig) *ReadStoreFactory {
	return &ReadStoreFactory{
		config: config,
	}
}

// CreateUserViewStore creates a read store for UserView
func (f *ReadStoreFactory) CreateUserViewStore() (interface{}, error) {
	switch f.config.Type {
	case ReadStoreTypeMongoDB:
		return f.createMongoReadStore("user_views")
	case ReadStoreTypeRedis:
		return f.createRedisReadStore("user_views")
	case ReadStoreTypeMemory:
		return f.createMemoryReadStore()
	default:
		return nil, fmt.Errorf("unsupported read store type: %s", f.config.Type)
	}
}

// CreateOrderSummaryViewStore creates a read store for OrderSummaryView
func (f *ReadStoreFactory) CreateOrderSummaryViewStore() (interface{}, error) {
	switch f.config.Type {
	case ReadStoreTypeMongoDB:
		return f.createMongoReadStore("order_summary_views")
	case ReadStoreTypeRedis:
		return f.createRedisReadStore("order_summary_views")
	case ReadStoreTypeMemory:
		return f.createMemoryReadStore()
	default:
		return nil, fmt.Errorf("unsupported read store type: %s", f.config.Type)
	}
}

// CreateCustomerHistoryViewStore creates a read store for CustomerOrderHistoryView
func (f *ReadStoreFactory) CreateCustomerHistoryViewStore() (interface{}, error) {
	switch f.config.Type {
	case ReadStoreTypeMongoDB:
		return f.createMongoReadStore("customer_history_views")
	case ReadStoreTypeRedis:
		return f.createRedisReadStore("customer_history_views")
	case ReadStoreTypeMemory:
		return f.createMemoryReadStore()
	default:
		return nil, fmt.Errorf("unsupported read store type: %s", f.config.Type)
	}
}

// CreateProductPopularityViewStore creates a read store for ProductPopularityView
func (f *ReadStoreFactory) CreateProductPopularityViewStore() (interface{}, error) {
	switch f.config.Type {
	case ReadStoreTypeMongoDB:
		return f.createMongoReadStore("product_popularity_views")
	case ReadStoreTypeRedis:
		return f.createRedisReadStore("product_popularity_views")
	case ReadStoreTypeMemory:
		return f.createMemoryReadStore()
	default:
		return nil, fmt.Errorf("unsupported read store type: %s", f.config.Type)
	}
}

// CreateDashboardViewStore creates a read store for DashboardView with TTL support
func (f *ReadStoreFactory) CreateDashboardViewStore() (interface{}, error) {
	switch f.config.Type {
	case ReadStoreTypeMongoDB:
		// MongoDB with TTL index for dashboard views
		return f.createMongoReadStoreWithTTL("dashboard_views")
	case ReadStoreTypeRedis:
		// Redis with TTL support
		return f.createRedisReadStoreWithTTL("dashboard_views")
	case ReadStoreTypeMemory:
		return f.createMemoryReadStore()
	default:
		return nil, fmt.Errorf("unsupported read store type: %s", f.config.Type)
	}
}

// CreateGenericReadStore creates a generic read store
func (f *ReadStoreFactory) CreateGenericReadStore(collectionName string) (interface{}, error) {
	switch f.config.Type {
	case ReadStoreTypeMongoDB:
		return f.createMongoReadStore(collectionName)
	case ReadStoreTypeRedis:
		return f.createRedisReadStore(collectionName)
	case ReadStoreTypeMemory:
		return f.createMemoryReadStore()
	default:
		return nil, fmt.Errorf("unsupported read store type: %s", f.config.Type)
	}
}

// createMongoReadStore creates a MongoDB read store
func (f *ReadStoreFactory) createMongoReadStore(collectionName string) (interface{}, error) {
	if f.config.MongoDB == nil {
		return nil, fmt.Errorf("MongoDB configuration is required")
	}

	// Create MongoDB client manager
	mongoConfig := &cqrsx.MongoConfig{
		URI:      f.config.MongoDB.ConnectionString,
		Database: f.config.MongoDB.DatabaseName,
	}
	clientManager, err := cqrsx.NewMongoClientManager(mongoConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create MongoDB client manager: %w", err)
	}

	// Create MongoDB read store
	readStore := cqrsx.NewMongoReadStore(clientManager, collectionName)

	return readStore, nil
}

// createMongoReadStoreWithTTL creates a MongoDB read store with TTL support
func (f *ReadStoreFactory) createMongoReadStoreWithTTL(collectionName string) (interface{}, error) {
	readStore, err := f.createMongoReadStore(collectionName)
	if err != nil {
		return nil, err
	}

	// TODO: Configure TTL index for dashboard views
	// This would typically involve creating a TTL index on the expires_at field

	return readStore, nil
}

// createRedisReadStore creates a Redis read store
func (f *ReadStoreFactory) createRedisReadStore(keyPrefix string) (interface{}, error) {
	if f.config.Redis == nil {
		return nil, fmt.Errorf("Redis configuration is required")
	}

	// Create Redis client manager
	redisConfig := &cqrsx.RedisConfig{
		Host:     f.config.Redis.Address,
		Port:     6379, // Default Redis port
		Password: f.config.Redis.Password,
		Database: f.config.Redis.Database,
	}
	clientManager, err := cqrsx.NewRedisClientManager(redisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client manager: %w", err)
	}

	// Create Redis read store with JSON serializer
	serializer := &cqrsx.JSONReadModelSerializer{}
	readStore := cqrsx.NewRedisReadStore(clientManager, keyPrefix, serializer)

	return readStore, nil
}

// createRedisReadStoreWithTTL creates a Redis read store with TTL support
func (f *ReadStoreFactory) createRedisReadStoreWithTTL(keyPrefix string) (interface{}, error) {
	readStore, err := f.createRedisReadStore(keyPrefix)
	if err != nil {
		return nil, err
	}

	// Redis naturally supports TTL, so no additional configuration needed

	return readStore, nil
}

// createMemoryReadStore creates an in-memory read store
func (f *ReadStoreFactory) createMemoryReadStore() (interface{}, error) {
	// Create in-memory read store
	readStore := cqrs.NewInMemoryReadStore()

	return readStore, nil
}

// Note: Read model type registration is handled by the specific read store implementations

// GetConfig returns the current configuration
func (f *ReadStoreFactory) GetConfig() ReadStoreConfig {
	return f.config
}

// UpdateConfig updates the factory configuration
func (f *ReadStoreFactory) UpdateConfig(config ReadStoreConfig) {
	f.config = config
}

// ValidateConfig validates the configuration
func (f *ReadStoreFactory) ValidateConfig() error {
	switch f.config.Type {
	case ReadStoreTypeMongoDB:
		if f.config.MongoDB == nil {
			return fmt.Errorf("MongoDB configuration is required")
		}
		if f.config.MongoDB.ConnectionString == "" {
			return fmt.Errorf("MongoDB connection string is required")
		}
		if f.config.MongoDB.DatabaseName == "" {
			return fmt.Errorf("MongoDB database name is required")
		}
	case ReadStoreTypeRedis:
		if f.config.Redis == nil {
			return fmt.Errorf("Redis configuration is required")
		}
		if f.config.Redis.Address == "" {
			return fmt.Errorf("Redis address is required")
		}
	case ReadStoreTypeMemory:
		// Memory store doesn't require specific configuration
	default:
		return fmt.Errorf("unsupported read store type: %s", f.config.Type)
	}

	return nil
}

// CreateAllStores creates all read stores for the application
func (f *ReadStoreFactory) CreateAllStores() (map[string]interface{}, error) {
	if err := f.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	stores := make(map[string]interface{})

	// Create UserView store
	userStore, err := f.CreateUserViewStore()
	if err != nil {
		return nil, fmt.Errorf("failed to create UserView store: %w", err)
	}
	stores["UserView"] = userStore

	// Create OrderSummaryView store
	orderStore, err := f.CreateOrderSummaryViewStore()
	if err != nil {
		return nil, fmt.Errorf("failed to create OrderSummaryView store: %w", err)
	}
	stores["OrderSummaryView"] = orderStore

	// Create CustomerOrderHistoryView store
	historyStore, err := f.CreateCustomerHistoryViewStore()
	if err != nil {
		return nil, fmt.Errorf("failed to create CustomerOrderHistoryView store: %w", err)
	}
	stores["CustomerOrderHistoryView"] = historyStore

	// Create ProductPopularityView store
	popularityStore, err := f.CreateProductPopularityViewStore()
	if err != nil {
		return nil, fmt.Errorf("failed to create ProductPopularityView store: %w", err)
	}
	stores["ProductPopularityView"] = popularityStore

	// Create DashboardView store
	dashboardStore, err := f.CreateDashboardViewStore()
	if err != nil {
		return nil, fmt.Errorf("failed to create DashboardView store: %w", err)
	}
	stores["DashboardView"] = dashboardStore

	return stores, nil
}

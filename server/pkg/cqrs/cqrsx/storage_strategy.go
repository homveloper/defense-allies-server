package cqrsx

import (
	"cqrs"
	"time"
)

// StorageConfiguration represents the comprehensive configuration for all storage mechanisms.
// This structure centralizes all storage-related settings, enabling consistent configuration
// management across different storage strategies and implementations.
//
// Configuration sections:
//   - Redis: Connection and operational settings for Redis storage
//   - MongoDB: Connection and operational settings for MongoDB storage
//   - EventSourcing: Event sourcing specific settings and policies
//   - Performance: Performance tuning and optimization settings
type StorageConfiguration struct {
	Redis         *RedisConfig         `json:"redis"`          // Redis connection and operational configuration
	MongoDB       *MongoConfig         `json:"mongodb"`        // MongoDB connection and operational configuration
	EventSourcing *EventSourcingConfig `json:"event_sourcing"` // Event sourcing specific configuration
	Performance   *PerformanceConfig   `json:"performance"`    // Performance tuning configuration
}

// RedisConfig represents Redis connection and operational configuration.
// This structure contains all settings needed to establish and maintain
// Redis connections, including connection pooling, timeouts, and retry policies.
//
// Connection settings:
//   - Host, Port, Database: Basic connection parameters
//   - Password: Authentication credential
//   - PoolSize: Connection pool configuration
//   - Timeouts: Various timeout settings for different operations
type RedisConfig struct {
	Host         string        `json:"host"`          // Redis server hostname or IP address
	Port         int           `json:"port"`          // Redis server port number
	Database     int           `json:"database"`      // Redis database number (0-15)
	Password     string        `json:"password"`      // Redis authentication password (empty if no auth)
	PoolSize     int           `json:"pool_size"`     // Maximum number of connections in the pool
	MaxRetries   int           `json:"max_retries"`   // Maximum number of retry attempts for failed operations
	DialTimeout  time.Duration `json:"dial_timeout"`  // Timeout for establishing new connections
	ReadTimeout  time.Duration `json:"read_timeout"`  // Timeout for read operations
	WriteTimeout time.Duration `json:"write_timeout"` // Timeout for write operations
}

// MongoConfig represents MongoDB connection and operational configuration.
// This structure contains all settings needed to establish and maintain
// MongoDB connections, including connection pooling, timeouts, and authentication.
//
// Connection settings:
//   - URI: MongoDB connection string
//   - Database: Target database name
//   - Username, Password: Authentication credentials
//   - Connection pooling and timeout settings
type MongoConfig struct {
	URI                    string        `json:"uri"`                      // MongoDB connection URI
	Database               string        `json:"database"`                 // MongoDB database name
	Username               string        `json:"username"`                 // MongoDB username (optional)
	Password               string        `json:"password"`                 // MongoDB password (optional)
	MaxPoolSize            int           `json:"max_pool_size"`            // Maximum number of connections in the pool
	ConnectTimeout         time.Duration `json:"connect_timeout"`          // Timeout for establishing new connections
	SocketTimeout          time.Duration `json:"socket_timeout"`           // Timeout for socket operations
	ServerSelectionTimeout time.Duration `json:"server_selection_timeout"` // Timeout for server selection
}

// EventSourcingConfig represents event sourcing specific configuration.
// This structure contains settings that control event sourcing behavior,
// including snapshot management, compression, compaction, and retention policies.
//
// Key features:
//   - Snapshot frequency control for performance optimization
//   - Compression settings for storage efficiency
//   - Compaction policies for storage management
//   - Retention policies for compliance and cleanup
type EventSourcingConfig struct {
	SnapshotFrequency int               `json:"snapshot_frequency"` // Number of events between snapshots (0 = disabled)
	EnableCompression bool              `json:"enable_compression"` // Whether to compress event data
	CompactionPolicy  *CompactionPolicy `json:"compaction_policy"`  // Event compaction configuration
	RetentionPolicy   *RetentionPolicy  `json:"retention_policy"`   // Event retention configuration
}

// CompactionPolicy represents event compaction policy configuration.
// Event compaction removes old events that are no longer needed for aggregate
// reconstruction, typically after snapshots have been created. This helps
// manage storage growth and improve performance.
//
// Compaction strategy:
//   - Events before the latest snapshot can be safely removed
//   - Compaction runs periodically based on configured intervals
//   - Minimum event thresholds prevent unnecessary compaction
type CompactionPolicy struct {
	Enabled            bool          `json:"enabled"`               // Whether event compaction is enabled
	MinEventsToCompact int           `json:"min_events_to_compact"` // Minimum number of events before compaction
	CompactionInterval time.Duration `json:"compaction_interval"`   // How often to run compaction process
}

// RetentionPolicy represents event retention policy configuration.
// Retention policies define how long events are kept in the system before
// being archived or deleted. This is important for compliance, storage
// management, and performance optimization.
//
// Retention strategy:
//   - Events older than retention period are candidates for archival/deletion
//   - Archive storage can be used for long-term retention
//   - Policies can be customized based on business requirements
type RetentionPolicy struct {
	Enabled        bool   `json:"enabled"`         // Whether retention policy is enabled
	RetentionDays  int    `json:"retention_days"`  // Number of days to retain events
	ArchiveEnabled bool   `json:"archive_enabled"` // Whether to archive old events instead of deleting
	ArchiveStorage string `json:"archive_storage"` // Archive storage location/configuration
}

// PerformanceConfig represents performance tuning configuration.
// This structure contains settings that affect system performance,
// including batching, caching, connection pooling, and concurrency limits.
//
// Performance areas:
//   - Batch processing for improved throughput
//   - Caching for reduced latency
//   - Connection pooling for resource efficiency
//   - Concurrency limits for system stability
type PerformanceConfig struct {
	BatchSize          int           `json:"batch_size"`           // Number of operations to batch together
	CacheSize          int           `json:"cache_size"`           // Maximum number of items in cache
	CacheTTL           time.Duration `json:"cache_ttl"`            // Time-to-live for cached items
	ConnectionPoolSize int           `json:"connection_pool_size"` // Size of connection pools
	MaxConcurrentOps   int           `json:"max_concurrent_ops"`   // Maximum concurrent operations
}

// ConfigurableStorageStrategy provides a flexible, configuration-driven storage strategy implementation.
// This implementation allows different storage strategies to be configured per aggregate type,
// with fallback to a default strategy. It supports dependency injection through factory functions
// and provides comprehensive configuration management.
//
// Key features:
//   - Per-aggregate-type storage strategy configuration
//   - Default strategy fallback for unconfigured aggregates
//   - Factory function injection for repository creation
//   - Comprehensive configuration validation
//   - Runtime strategy modification support
//
// Usage patterns:
//   - Configure default strategy for most aggregates
//   - Override strategy for specific aggregates with special requirements
//   - Inject factory functions for different repository implementations
//   - Validate configuration before system startup
type ConfigurableStorageStrategy struct {
	DefaultType         cqrs.RepositoryType            `json:"default_type"`         // Default repository type for unconfigured aggregates
	AggregateStrategies map[string]cqrs.RepositoryType `json:"aggregate_strategies"` // Per-aggregate-type strategy overrides
	Configuration       *StorageConfiguration          `json:"configuration"`        // Storage configuration settings

	// Factory functions for creating repositories (not serialized)
	EventSourcedFactory func(string) (cqrs.EventSourcedRepository, error) `json:"-"` // Factory for event sourced repositories
	StateBasedFactory   func(string) (cqrs.StateBasedRepository, error)   `json:"-"` // Factory for state based repositories
	HybridFactory       func(string) (cqrs.HybridRepository, error)       `json:"-"` // Factory for hybrid repositories
}

// NewConfigurableStorageStrategy creates and initializes a new configurable storage strategy.
// This constructor sets up the strategy with a default repository type and configuration,
// preparing it for aggregate-specific strategy registration and factory injection.
//
// Parameters:
//   - defaultType: The default repository type to use for unconfigured aggregates
//   - config: Storage configuration containing connection and operational settings
//
// Returns:
//   - *ConfigurableStorageStrategy: A new strategy instance ready for configuration
//
// Usage:
//
//	strategy := NewConfigurableStorageStrategy(StateBased, config)
//	strategy.SetAggregateStrategy("User", EventSourced)
//	strategy.SetEventSourcedFactory(createEventSourcedRepo)
func NewConfigurableStorageStrategy(defaultType cqrs.RepositoryType, config *StorageConfiguration) *ConfigurableStorageStrategy {
	return &ConfigurableStorageStrategy{
		DefaultType:         defaultType,
		AggregateStrategies: make(map[string]cqrs.RepositoryType),
		Configuration:       config,
	}
}

// StorageStrategy interface implementation

func (css *ConfigurableStorageStrategy) GetRepositoryType(aggregateType string) cqrs.RepositoryType {
	if repoType, exists := css.AggregateStrategies[aggregateType]; exists {
		return repoType
	}
	return css.DefaultType
}

func (css *ConfigurableStorageStrategy) CreateRepository(aggregateType string) (cqrs.Repository, error) {
	repoType := css.GetRepositoryType(aggregateType)

	switch repoType {
	case cqrs.EventSourced:
		if css.EventSourcedFactory != nil {
			return css.EventSourcedFactory(aggregateType)
		}
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "event sourced factory not configured", nil)

	case cqrs.StateBased:
		if css.StateBasedFactory != nil {
			return css.StateBasedFactory(aggregateType)
		}
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "state based factory not configured", nil)

	case cqrs.Hybrid:
		if css.HybridFactory != nil {
			return css.HybridFactory(aggregateType)
		}
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "hybrid factory not configured", nil)

	default:
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "unknown repository type", nil)
	}
}

func (css *ConfigurableStorageStrategy) GetConfiguration() cqrs.StorageConfiguration {
	return css.Configuration
}

func (css *ConfigurableStorageStrategy) ValidateConfiguration() error {
	if css.Configuration == nil {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "storage configuration is nil", nil)
	}

	return css.Configuration.ValidateConfiguration()
}

// Helper methods

// StorageConfiguration interface implementation
func (sc *StorageConfiguration) ValidateConfiguration() error {
	// At least one storage configuration must be provided
	if sc.Redis == nil && sc.MongoDB == nil {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "at least one storage configuration (Redis or MongoDB) is required", nil)
	}

	// Validate Redis configuration if provided
	if sc.Redis != nil {
		if sc.Redis.Host == "" {
			return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "Redis host is required", nil)
		}

		if sc.Redis.Port <= 0 {
			return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "Redis port must be positive", nil)
		}
	}

	// Validate MongoDB configuration if provided
	if sc.MongoDB != nil {
		if sc.MongoDB.URI == "" {
			return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "MongoDB URI is required", nil)
		}

		if sc.MongoDB.Database == "" {
			return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "MongoDB database name is required", nil)
		}
	}

	return nil
}

func (sc *StorageConfiguration) GetConfigType() string {
	if sc.Redis != nil && sc.MongoDB != nil {
		return "hybrid"
	} else if sc.MongoDB != nil {
		return "mongodb"
	}
	return "redis"
}

// SetAggregateStrategy sets storage strategy for specific aggregate type
func (css *ConfigurableStorageStrategy) SetAggregateStrategy(aggregateType string, repoType cqrs.RepositoryType) {
	css.AggregateStrategies[aggregateType] = repoType
}

// SetEventSourcedFactory sets the factory for event sourced repositories
func (css *ConfigurableStorageStrategy) SetEventSourcedFactory(factory func(string) (cqrs.EventSourcedRepository, error)) {
	css.EventSourcedFactory = factory
}

// SetStateBasedFactory sets the factory for state based repositories
func (css *ConfigurableStorageStrategy) SetStateBasedFactory(factory func(string) (cqrs.StateBasedRepository, error)) {
	css.StateBasedFactory = factory
}

// SetHybridFactory sets the factory for hybrid repositories
func (css *ConfigurableStorageStrategy) SetHybridFactory(factory func(string) (cqrs.HybridRepository, error)) {
	css.HybridFactory = factory
}

// GetAggregateStrategies returns all configured aggregate strategies
func (css *ConfigurableStorageStrategy) GetAggregateStrategies() map[string]cqrs.RepositoryType {
	strategies := make(map[string]cqrs.RepositoryType)
	for k, v := range css.AggregateStrategies {
		strategies[k] = v
	}
	return strategies
}

package cqrs

import "time"

// RepositoryType represents the different storage strategies available for aggregates.
// This enumeration defines the fundamental approaches to aggregate persistence,
// each with distinct characteristics, performance profiles, and use cases.
// The choice of repository type affects how aggregates are stored, retrieved,
// and managed throughout their lifecycle.
type RepositoryType int

const (
	// EventSourced represents event sourcing storage strategy.
	// Aggregates are persisted as a sequence of events, providing complete audit trails
	// and enabling temporal queries. Best for complex business logic and audit requirements.
	EventSourced RepositoryType = iota

	// StateBased represents traditional state-based storage strategy.
	// Aggregates are persisted as current state snapshots, providing high performance
	// for read operations. Best for simple CRUD operations and performance-critical scenarios.
	StateBased

	// Hybrid represents a combination of event sourcing and state storage.
	// Provides both audit trails through events and performance through state snapshots.
	// Best for complex systems requiring both audit capabilities and high performance.
	Hybrid
)

// String returns the string representation of the repository type.
// This method enables readable logging, configuration, and debugging output.
//
// Returns:
//   - string: Human-readable repository type name
func (rt RepositoryType) String() string {
	switch rt {
	case EventSourced:
		return "event_sourced"
	case StateBased:
		return "state_based"
	case Hybrid:
		return "hybrid"
	default:
		return "unknown"
	}
}

// StorageStrategy interface defines the contract for storage strategy selection and management.
// This interface enables flexible, configurable storage approaches that can vary by aggregate type.
// It supports the Strategy pattern, allowing runtime selection of storage mechanisms based on
// business requirements, performance needs, and operational constraints.
//
// Key responsibilities:
//   - Determine appropriate storage strategy for each aggregate type
//   - Create repositories with proper configuration
//   - Manage storage configuration and validation
//   - Support dynamic strategy selection
//
// Implementation guidelines:
//   - Support multiple storage strategies within the same application
//   - Provide clear configuration validation
//   - Enable runtime strategy changes where appropriate
//   - Ensure consistent behavior across different strategies
type StorageStrategy interface {
	// GetRepositoryType determines the storage strategy for a specific aggregate type.
	// This method enables per-aggregate-type storage strategy selection, allowing
	// fine-grained control over storage approaches within the same application.
	//
	// Parameters:
	//   - aggregateType: The type of aggregate to get storage strategy for
	//
	// Returns:
	//   - RepositoryType: The storage strategy to use for this aggregate type
	//
	// Usage: Called by repository factories to determine which repository implementation to create
	GetRepositoryType(aggregateType string) RepositoryType

	// CreateRepository creates a repository instance for the specified aggregate type.
	// This method acts as a factory, creating the appropriate repository implementation
	// based on the configured storage strategy for the aggregate type.
	//
	// Parameters:
	//   - aggregateType: The type of aggregate to create a repository for
	//
	// Returns:
	//   - Repository: A repository instance configured for the aggregate type
	//   - error: nil on success, error if repository creation fails
	//
	// Error conditions:
	//   - aggregateType is empty: Returns validation error
	//   - no factory configured for strategy: Returns configuration error
	//   - repository creation fails: Returns creation error with underlying cause
	CreateRepository(aggregateType string) (Repository, error)

	// GetConfiguration returns the current storage configuration.
	// This method provides access to the underlying storage configuration,
	// enabling inspection and validation of storage settings.
	//
	// Returns:
	//   - *StorageConfiguration: The current storage configuration
	GetConfiguration() *StorageConfiguration

	// ValidateConfiguration checks if the current configuration is valid and complete.
	// This method performs comprehensive validation of all configuration settings
	// to ensure the storage strategy can operate correctly.
	//
	// Returns:
	//   - error: nil if configuration is valid, descriptive error if validation fails
	//
	// Validation checks:
	//   - Required configuration sections are present
	//   - Connection parameters are valid
	//   - Performance settings are within acceptable ranges
	//   - Factory functions are configured for enabled strategies
	ValidateConfiguration() error
}

// StorageConfiguration represents the comprehensive configuration for all storage mechanisms.
// This structure centralizes all storage-related settings, enabling consistent configuration
// management across different storage strategies and implementations.
//
// Configuration sections:
//   - Redis: Connection and operational settings for Redis storage
//   - EventSourcing: Event sourcing specific settings and policies
//   - Performance: Performance tuning and optimization settings
type StorageConfiguration struct {
	Redis         *RedisConfig         `json:"redis"`          // Redis connection and operational configuration
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
	DefaultType         RepositoryType            `json:"default_type"`         // Default repository type for unconfigured aggregates
	AggregateStrategies map[string]RepositoryType `json:"aggregate_strategies"` // Per-aggregate-type strategy overrides
	Configuration       *StorageConfiguration     `json:"configuration"`        // Storage configuration settings

	// Factory functions for creating repositories (not serialized)
	EventSourcedFactory func(string) (EventSourcedRepository, error) `json:"-"` // Factory for event sourced repositories
	StateBasedFactory   func(string) (StateBasedRepository, error)   `json:"-"` // Factory for state based repositories
	HybridFactory       func(string) (HybridRepository, error)       `json:"-"` // Factory for hybrid repositories
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
func NewConfigurableStorageStrategy(defaultType RepositoryType, config *StorageConfiguration) *ConfigurableStorageStrategy {
	return &ConfigurableStorageStrategy{
		DefaultType:         defaultType,
		AggregateStrategies: make(map[string]RepositoryType),
		Configuration:       config,
	}
}

// StorageStrategy interface implementation

func (css *ConfigurableStorageStrategy) GetRepositoryType(aggregateType string) RepositoryType {
	if repoType, exists := css.AggregateStrategies[aggregateType]; exists {
		return repoType
	}
	return css.DefaultType
}

func (css *ConfigurableStorageStrategy) CreateRepository(aggregateType string) (Repository, error) {
	repoType := css.GetRepositoryType(aggregateType)

	switch repoType {
	case EventSourced:
		if css.EventSourcedFactory != nil {
			return css.EventSourcedFactory(aggregateType)
		}
		return nil, NewCQRSError(ErrCodeRepositoryError.String(), "event sourced factory not configured", nil)

	case StateBased:
		if css.StateBasedFactory != nil {
			return css.StateBasedFactory(aggregateType)
		}
		return nil, NewCQRSError(ErrCodeRepositoryError.String(), "state based factory not configured", nil)

	case Hybrid:
		if css.HybridFactory != nil {
			return css.HybridFactory(aggregateType)
		}
		return nil, NewCQRSError(ErrCodeRepositoryError.String(), "hybrid factory not configured", nil)

	default:
		return nil, NewCQRSError(ErrCodeRepositoryError.String(), "unknown repository type", nil)
	}
}

func (css *ConfigurableStorageStrategy) GetConfiguration() *StorageConfiguration {
	return css.Configuration
}

func (css *ConfigurableStorageStrategy) ValidateConfiguration() error {
	if css.Configuration == nil {
		return NewCQRSError(ErrCodeRepositoryError.String(), "storage configuration is nil", nil)
	}

	if css.Configuration.Redis == nil {
		return NewCQRSError(ErrCodeRepositoryError.String(), "Redis configuration is required", nil)
	}

	if css.Configuration.Redis.Host == "" {
		return NewCQRSError(ErrCodeRepositoryError.String(), "Redis host is required", nil)
	}

	if css.Configuration.Redis.Port <= 0 {
		return NewCQRSError(ErrCodeRepositoryError.String(), "Redis port must be positive", nil)
	}

	return nil
}

// Helper methods

// SetAggregateStrategy sets storage strategy for specific aggregate type
func (css *ConfigurableStorageStrategy) SetAggregateStrategy(aggregateType string, repoType RepositoryType) {
	css.AggregateStrategies[aggregateType] = repoType
}

// SetEventSourcedFactory sets the factory for event sourced repositories
func (css *ConfigurableStorageStrategy) SetEventSourcedFactory(factory func(string) (EventSourcedRepository, error)) {
	css.EventSourcedFactory = factory
}

// SetStateBasedFactory sets the factory for state based repositories
func (css *ConfigurableStorageStrategy) SetStateBasedFactory(factory func(string) (StateBasedRepository, error)) {
	css.StateBasedFactory = factory
}

// SetHybridFactory sets the factory for hybrid repositories
func (css *ConfigurableStorageStrategy) SetHybridFactory(factory func(string) (HybridRepository, error)) {
	css.HybridFactory = factory
}

// GetAggregateStrategies returns all configured aggregate strategies
func (css *ConfigurableStorageStrategy) GetAggregateStrategies() map[string]RepositoryType {
	strategies := make(map[string]RepositoryType)
	for k, v := range css.AggregateStrategies {
		strategies[k] = v
	}
	return strategies
}

// RepositoryFactory interface defines the contract for repository creation.
// This interface provides a standardized way to create different types of repositories
// based on aggregate types and storage requirements. It supports the Abstract Factory
// pattern, enabling flexible repository creation strategies.
//
// Key responsibilities:
//   - Create repositories of different types (EventSourced, StateBased, Hybrid)
//   - Support aggregate-type-specific repository configuration
//   - Provide introspection capabilities for supported repository types
//   - Enable dependency injection and testing through interface abstraction
//
// Implementation guidelines:
//   - Ensure consistent repository configuration across all created instances
//   - Provide meaningful error messages for unsupported aggregate types
//   - Support lazy initialization and resource pooling where appropriate
//   - Enable runtime repository type discovery and validation
type RepositoryFactory interface {
	// CreateEventSourcedRepository creates an event sourced repository for the specified aggregate type.
	// This method creates repositories that store aggregates as sequences of events,
	// providing complete audit trails and enabling temporal queries.
	//
	// Parameters:
	//   - aggregateType: The type of aggregate this repository will manage
	//
	// Returns:
	//   - EventSourcedRepository: A repository instance configured for event sourcing
	//   - error: nil on success, error if repository creation fails
	//
	// Error conditions:
	//   - aggregateType is empty: Returns validation error
	//   - event sourcing not supported: Returns configuration error
	//   - repository creation fails: Returns creation error with underlying cause
	CreateEventSourcedRepository(aggregateType string) (EventSourcedRepository, error)

	// CreateStateBasedRepository creates a state-based repository for the specified aggregate type.
	// This method creates repositories that store aggregates as current state snapshots,
	// providing high performance for read operations and simple CRUD scenarios.
	//
	// Parameters:
	//   - aggregateType: The type of aggregate this repository will manage
	//
	// Returns:
	//   - StateBasedRepository: A repository instance configured for state-based storage
	//   - error: nil on success, error if repository creation fails
	//
	// Error conditions:
	//   - aggregateType is empty: Returns validation error
	//   - state-based storage not supported: Returns configuration error
	//   - repository creation fails: Returns creation error with underlying cause
	CreateStateBasedRepository(aggregateType string) (StateBasedRepository, error)

	// CreateHybridRepository creates a hybrid repository for the specified aggregate type.
	// This method creates repositories that combine event sourcing and state storage,
	// providing both audit capabilities and high performance.
	//
	// Parameters:
	//   - aggregateType: The type of aggregate this repository will manage
	//
	// Returns:
	//   - HybridRepository: A repository instance configured for hybrid storage
	//   - error: nil on success, error if repository creation fails
	//
	// Error conditions:
	//   - aggregateType is empty: Returns validation error
	//   - hybrid storage not supported: Returns configuration error
	//   - repository creation fails: Returns creation error with underlying cause
	CreateHybridRepository(aggregateType string) (HybridRepository, error)

	// GetSupportedTypes returns the repository types supported by this factory.
	// This method enables runtime discovery of factory capabilities and validation
	// of repository type requests before attempting creation.
	//
	// Returns:
	//   - []RepositoryType: Slice of supported repository types
	//
	// Usage: Check supported types before calling creation methods to avoid errors
	GetSupportedTypes() []RepositoryType
}

package cqrs

import "time"

// RepositoryType represents different repository storage types
type RepositoryType int

const (
	EventSourced RepositoryType = iota
	StateBased
	Hybrid // Event sourcing + state storage combination
)

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

// StorageStrategy interface for selecting storage approach
type StorageStrategy interface {
	GetRepositoryType(aggregateType string) RepositoryType
	CreateRepository(aggregateType string) (Repository, error)
	GetConfiguration() *StorageConfiguration
	ValidateConfiguration() error
}

// StorageConfiguration represents storage configuration
type StorageConfiguration struct {
	// Redis configuration
	Redis *RedisConfig

	// Event sourcing configuration
	EventSourcing *EventSourcingConfig

	// Performance configuration
	Performance *PerformanceConfig
}

// RedisConfig represents Redis connection configuration
type RedisConfig struct {
	Host         string
	Port         int
	Database     int
	Password     string
	PoolSize     int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// EventSourcingConfig represents event sourcing configuration
type EventSourcingConfig struct {
	SnapshotFrequency int
	EnableCompression bool
	CompactionPolicy  *CompactionPolicy
	RetentionPolicy   *RetentionPolicy
}

// CompactionPolicy represents event compaction policy
type CompactionPolicy struct {
	Enabled            bool
	MinEventsToCompact int
	CompactionInterval time.Duration
}

// RetentionPolicy represents event retention policy
type RetentionPolicy struct {
	Enabled        bool
	RetentionDays  int
	ArchiveEnabled bool
	ArchiveStorage string
}

// PerformanceConfig represents performance configuration
type PerformanceConfig struct {
	BatchSize          int
	CacheSize          int
	CacheTTL           time.Duration
	ConnectionPoolSize int
	MaxConcurrentOps   int
}

// ConfigurableStorageStrategy provides configuration-based storage strategy
type ConfigurableStorageStrategy struct {
	DefaultType         RepositoryType
	AggregateStrategies map[string]RepositoryType
	Configuration       *StorageConfiguration

	// Factory functions
	EventSourcedFactory func(string) (EventSourcedRepository, error)
	StateBasedFactory   func(string) (StateBasedRepository, error)
	HybridFactory       func(string) (HybridRepository, error)
}

// NewConfigurableStorageStrategy creates a new configurable storage strategy
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

// RepositoryFactory interface for creating repositories
type RepositoryFactory interface {
	CreateEventSourcedRepository(aggregateType string) (EventSourcedRepository, error)
	CreateStateBasedRepository(aggregateType string) (StateBasedRepository, error)
	CreateHybridRepository(aggregateType string) (HybridRepository, error)
	GetSupportedTypes() []RepositoryType
}

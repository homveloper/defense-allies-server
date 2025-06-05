package cqrs

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
	//   - StorageConfiguration: The current storage configuration
	GetConfiguration() StorageConfiguration

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
// This interface allows different storage implementations to define their own configuration
// structures while maintaining a common contract for validation and management.
//
// Implementation note: Concrete configuration structures should be defined in the
// infrastructure layer (e.g., cqrsx package) to maintain dependency inversion.
type StorageConfiguration interface {
	// ValidateConfiguration checks if the configuration is valid and complete
	ValidateConfiguration() error

	// GetConfigType returns the type of configuration (e.g., "redis", "postgres")
	GetConfigType() string
}

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

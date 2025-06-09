package cqrs

import (
	"context"
	"fmt"
	"sync"
)

// InMemoryQueryDispatcher provides an in-memory implementation of QueryDispatcher interface.
// It manages query handlers and routes incoming queries to appropriate handlers based on query type.
// This implementation is thread-safe and suitable for single-instance applications.
//
// Fields:
//   - handlers: A map storing query handlers indexed by query type string
//   - mutex: Read-write mutex for thread-safe access to handlers map
type InMemoryQueryDispatcher struct {
	handlers map[string]QueryHandler // Map of query type -> handler
	mutex    sync.RWMutex            // Protects concurrent access to handlers map
}

// NewInMemoryQueryDispatcher creates and initializes a new in-memory query dispatcher.
//
// Returns:
//   - *InMemoryQueryDispatcher: A new dispatcher instance with empty handlers map
//
// Usage:
//
//	dispatcher := NewInMemoryQueryDispatcher()
func NewInMemoryQueryDispatcher() *InMemoryQueryDispatcher {
	return &InMemoryQueryDispatcher{
		handlers: make(map[string]QueryHandler),
	}
}

// QueryDispatcher interface implementation

// Dispatch routes a query to the appropriate handler and executes it.
// This method performs validation, handler lookup, and query execution in a thread-safe manner.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - query: The query to be executed (must implement Query interface)
//
// Returns:
//   - *QueryResult: Result containing success status, data, error, and execution metadata
//   - error: Always returns nil (errors are wrapped in QueryResult.Error)
//
// Error conditions:
//   - Query is nil: Returns QueryResult with validation error
//   - Query validation fails: Returns QueryResult with validation error
//   - No handler found: Returns QueryResult with handler not found error
//   - Handler execution fails: Returns QueryResult from handler (may contain error)
//
// Thread safety: This method is safe for concurrent use
func (d *InMemoryQueryDispatcher) Dispatch(ctx context.Context, query Query) (*QueryResult, error) {
	// Validate input parameters
	if query == nil {
		return &QueryResult{
			Success: false,
			Error:   NewCQRSError(ErrCodeQueryValidation.String(), "query cannot be nil", nil),
		}, nil
	}

	// Validate query structure and business rules
	if err := query.Validate(); err != nil {
		return &QueryResult{
			Success: false,
			Error:   NewCQRSError(ErrCodeQueryValidation.String(), "query validation failed", err),
		}, nil
	}

	// Find appropriate handler using read lock for thread safety
	d.mutex.RLock()
	handler, exists := d.handlers[query.QueryType()]
	d.mutex.RUnlock()

	// Check if handler exists for this query type
	if !exists {
		return &QueryResult{
			Success: false,
			Error:   NewCQRSError(ErrCodeQueryValidation.String(), fmt.Sprintf("no handler found for query type: %s", query.QueryType()), ErrQueryHandlerNotFound),
		}, nil
	}

	// Execute query using the found handler
	return handler.Handle(ctx, query)
}

// RegisterHandler registers a query handler for a specific query type.
// This method ensures that only one handler can be registered per query type to avoid conflicts.
//
// Parameters:
//   - queryType: The type of query this handler will process (must be non-empty)
//   - handler: The handler implementation (must be non-nil)
//
// Returns:
//   - error: nil on success, CQRSError on validation failure or duplicate registration
//
// Error conditions:
//   - queryType is empty: Returns validation error
//   - handler is nil: Returns validation error
//   - Handler already exists for queryType: Returns validation error
//
// Thread safety: This method is safe for concurrent use
func (d *InMemoryQueryDispatcher) RegisterHandler(queryType string, handler QueryHandler) error {
	// Validate input parameters
	if queryType == "" {
		return NewCQRSError(ErrCodeQueryValidation.String(), "query type cannot be empty", nil)
	}
	if handler == nil {
		return NewCQRSError(ErrCodeQueryValidation.String(), "handler cannot be nil", nil)
	}

	// Use write lock to ensure thread safety during registration
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Check for duplicate registration
	if _, exists := d.handlers[queryType]; exists {
		return NewCQRSError(ErrCodeQueryValidation.String(), fmt.Sprintf("handler already registered for query type: %s", queryType), nil)
	}

	// Register the handler
	d.handlers[queryType] = handler
	return nil
}

// UnregisterHandler removes a query handler for a specific query type.
// After unregistration, queries of this type will fail with "handler not found" error.
//
// Parameters:
//   - queryType: The type of query handler to remove (must be non-empty)
//
// Returns:
//   - error: nil on success, CQRSError on validation failure or handler not found
//
// Error conditions:
//   - queryType is empty: Returns validation error
//   - No handler registered for queryType: Returns handler not found error
//
// Thread safety: This method is safe for concurrent use
func (d *InMemoryQueryDispatcher) UnregisterHandler(queryType string) error {
	// Validate input parameters
	if queryType == "" {
		return NewCQRSError(ErrCodeQueryValidation.String(), "query type cannot be empty", nil)
	}

	// Use write lock to ensure thread safety during unregistration
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Check if handler exists
	if _, exists := d.handlers[queryType]; !exists {
		return NewCQRSError(ErrCodeQueryValidation.String(), fmt.Sprintf("no handler registered for query type: %s", queryType), ErrQueryHandlerNotFound)
	}

	// Remove the handler
	delete(d.handlers, queryType)
	return nil
}

// Helper methods

// GetRegisteredHandlers returns a list of all currently registered query types.
// This method is useful for debugging, monitoring, and introspection purposes.
//
// Returns:
//   - []string: A slice containing all registered query type names
//
// Thread safety: This method is safe for concurrent use
//
// Note: The returned slice is a copy, so modifications won't affect the internal state
func (d *InMemoryQueryDispatcher) GetRegisteredHandlers() []string {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	// Pre-allocate slice with known capacity for efficiency
	types := make([]string, 0, len(d.handlers))
	for queryType := range d.handlers {
		types = append(types, queryType)
	}
	return types
}

// HasHandler checks if a handler is registered for the given query type.
// This method is useful for conditional logic and validation before dispatching queries.
//
// Parameters:
//   - queryType: The query type to check for handler existence
//
// Returns:
//   - bool: true if a handler is registered, false otherwise
//
// Thread safety: This method is safe for concurrent use
func (d *InMemoryQueryDispatcher) HasHandler(queryType string) bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	_, exists := d.handlers[queryType]
	return exists
}

// GetHandlerCount returns the total number of registered query handlers.
// This method is useful for monitoring and capacity planning.
//
// Returns:
//   - int: The number of currently registered handlers
//
// Thread safety: This method is safe for concurrent use
func (d *InMemoryQueryDispatcher) GetHandlerCount() int {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return len(d.handlers)
}

// Clear removes all registered handlers from the dispatcher.
// This method is primarily used for testing and cleanup scenarios.
// After calling this method, all query dispatches will fail until handlers are re-registered.
//
// Thread safety: This method is safe for concurrent use
//
// Warning: This operation cannot be undone. All handler registrations will be lost.
func (d *InMemoryQueryDispatcher) Clear() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Create new map to ensure all references are cleared
	d.handlers = make(map[string]QueryHandler)
}

// BaseQueryHandler provides a base implementation of QueryHandler interface.
// This struct can be embedded in concrete query handlers to provide common functionality
// such as handler naming and query type management.
//
// Fields:
//   - name: Human-readable name for this handler (used for logging and debugging)
//   - queryTypes: Map of query types this handler can process (key=queryType, value=true)
type BaseQueryHandler struct {
	name       string          // Handler name for identification
	queryTypes map[string]bool // Set of supported query types
}

// NewBaseQueryHandler creates and initializes a new base query handler.
// This constructor sets up the handler with a name and list of supported query types.
//
// Parameters:
//   - name: Human-readable name for the handler (used for logging and debugging)
//   - queryTypes: Slice of query type strings this handler can process
//
// Returns:
//   - *BaseQueryHandler: A new base handler instance ready for use
//
// Usage:
//
//	handler := NewBaseQueryHandler("UserQueryHandler", []string{"GetUser", "ListUsers"})
func NewBaseQueryHandler(name string, queryTypes []string) *BaseQueryHandler {
	// Convert slice to map for O(1) lookup performance
	typeMap := make(map[string]bool)
	for _, queryType := range queryTypes {
		typeMap[queryType] = true
	}

	return &BaseQueryHandler{
		name:       name,
		queryTypes: typeMap,
	}
}

// QueryHandler interface implementation

// GetHandlerName returns the human-readable name of this handler.
// This name is used for logging, debugging, and monitoring purposes.
//
// Returns:
//   - string: The handler name set during construction
func (h *BaseQueryHandler) GetHandlerName() string {
	return h.name
}

// CanHandle checks if this handler can process the given query type.
// This method provides O(1) lookup performance using the internal map.
//
// Parameters:
//   - queryType: The query type string to check
//
// Returns:
//   - bool: true if this handler can process the query type, false otherwise
func (h *BaseQueryHandler) CanHandle(queryType string) bool {
	return h.queryTypes[queryType]
}

// Handle provides a default implementation that should be overridden by concrete handlers.
// This base implementation always returns an error indicating the method is not implemented.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - query: The query to be processed
//
// Returns:
//   - *QueryResult: Always returns a failed result with "not implemented" error
//   - error: Always returns nil (errors are wrapped in QueryResult.Error)
//
// Note: Concrete handlers MUST override this method to provide actual query processing logic
func (h *BaseQueryHandler) Handle(ctx context.Context, query Query) (*QueryResult, error) {
	// Base implementation - should be overridden by concrete handlers
	return &QueryResult{
		Success: false,
		Error:   NewCQRSError(ErrCodeQueryValidation.String(), "handle method not implemented", nil),
	}, nil
}

// Helper methods

// AddQueryType adds a new query type that this handler can process.
// This method allows dynamic expansion of handler capabilities at runtime.
//
// Parameters:
//   - queryType: The query type string to add to supported types
//
// Note: Adding a query type that already exists has no effect
func (h *BaseQueryHandler) AddQueryType(queryType string) {
	h.queryTypes[queryType] = true
}

// RemoveQueryType removes a query type from this handler's supported types.
// After removal, the handler will no longer be able to process queries of this type.
//
// Parameters:
//   - queryType: The query type string to remove from supported types
//
// Note: Removing a query type that doesn't exist has no effect
func (h *BaseQueryHandler) RemoveQueryType(queryType string) {
	delete(h.queryTypes, queryType)
}

// GetSupportedQueryTypes returns a list of all query types this handler can process.
// This method is useful for introspection, debugging, and dynamic handler discovery.
//
// Returns:
//   - []string: A slice containing all supported query type names
//
// Note: The returned slice is a copy, so modifications won't affect the internal state
func (h *BaseQueryHandler) GetSupportedQueryTypes() []string {
	// Pre-allocate slice with known capacity for efficiency
	types := make([]string, 0, len(h.queryTypes))
	for queryType := range h.queryTypes {
		types = append(types, queryType)
	}
	return types
}

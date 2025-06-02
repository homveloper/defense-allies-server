package cqrs

import (
	"context"
	"fmt"
	"sync"
)

// InMemoryQueryDispatcher provides an in-memory implementation of QueryDispatcher
type InMemoryQueryDispatcher struct {
	handlers map[string]QueryHandler
	mutex    sync.RWMutex
}

// NewInMemoryQueryDispatcher creates a new in-memory query dispatcher
func NewInMemoryQueryDispatcher() *InMemoryQueryDispatcher {
	return &InMemoryQueryDispatcher{
		handlers: make(map[string]QueryHandler),
	}
}

// QueryDispatcher interface implementation

func (d *InMemoryQueryDispatcher) Dispatch(ctx context.Context, query Query) (*QueryResult, error) {
	if query == nil {
		return &QueryResult{
			Success: false,
			Error:   NewCQRSError(ErrCodeQueryValidation.String(), "query cannot be nil", nil),
		}, nil
	}

	// Validate query
	if err := query.Validate(); err != nil {
		return &QueryResult{
			Success: false,
			Error:   NewCQRSError(ErrCodeQueryValidation.String(), "query validation failed", err),
		}, nil
	}

	// Find handler
	d.mutex.RLock()
	handler, exists := d.handlers[query.QueryType()]
	d.mutex.RUnlock()

	if !exists {
		return &QueryResult{
			Success: false,
			Error:   NewCQRSError(ErrCodeQueryValidation.String(), fmt.Sprintf("no handler found for query type: %s", query.QueryType()), ErrQueryHandlerNotFound),
		}, nil
	}

	// Execute query
	return handler.Handle(ctx, query)
}

func (d *InMemoryQueryDispatcher) RegisterHandler(queryType string, handler QueryHandler) error {
	if queryType == "" {
		return NewCQRSError(ErrCodeQueryValidation.String(), "query type cannot be empty", nil)
	}
	if handler == nil {
		return NewCQRSError(ErrCodeQueryValidation.String(), "handler cannot be nil", nil)
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if _, exists := d.handlers[queryType]; exists {
		return NewCQRSError(ErrCodeQueryValidation.String(), fmt.Sprintf("handler already registered for query type: %s", queryType), nil)
	}

	d.handlers[queryType] = handler
	return nil
}

func (d *InMemoryQueryDispatcher) UnregisterHandler(queryType string) error {
	if queryType == "" {
		return NewCQRSError(ErrCodeQueryValidation.String(), "query type cannot be empty", nil)
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if _, exists := d.handlers[queryType]; !exists {
		return NewCQRSError(ErrCodeQueryValidation.String(), fmt.Sprintf("no handler registered for query type: %s", queryType), ErrQueryHandlerNotFound)
	}

	delete(d.handlers, queryType)
	return nil
}

// Helper methods

// GetRegisteredHandlers returns all registered query types
func (d *InMemoryQueryDispatcher) GetRegisteredHandlers() []string {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	types := make([]string, 0, len(d.handlers))
	for queryType := range d.handlers {
		types = append(types, queryType)
	}
	return types
}

// HasHandler checks if a handler is registered for the given query type
func (d *InMemoryQueryDispatcher) HasHandler(queryType string) bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	_, exists := d.handlers[queryType]
	return exists
}

// GetHandlerCount returns the number of registered handlers
func (d *InMemoryQueryDispatcher) GetHandlerCount() int {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return len(d.handlers)
}

// Clear removes all registered handlers
func (d *InMemoryQueryDispatcher) Clear() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.handlers = make(map[string]QueryHandler)
}

// BaseQueryHandler provides a base implementation of QueryHandler
type BaseQueryHandler struct {
	name       string
	queryTypes map[string]bool
}

// NewBaseQueryHandler creates a new base query handler
func NewBaseQueryHandler(name string, queryTypes []string) *BaseQueryHandler {
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

func (h *BaseQueryHandler) GetHandlerName() string {
	return h.name
}

func (h *BaseQueryHandler) CanHandle(queryType string) bool {
	return h.queryTypes[queryType]
}

// Handle method should be implemented by concrete handlers
func (h *BaseQueryHandler) Handle(ctx context.Context, query Query) (*QueryResult, error) {
	// Base implementation - should be overridden
	return &QueryResult{
		Success: false,
		Error:   NewCQRSError(ErrCodeQueryValidation.String(), "handle method not implemented", nil),
	}, nil
}

// Helper methods

// AddQueryType adds a query type that this handler can process
func (h *BaseQueryHandler) AddQueryType(queryType string) {
	h.queryTypes[queryType] = true
}

// RemoveQueryType removes a query type from this handler
func (h *BaseQueryHandler) RemoveQueryType(queryType string) {
	delete(h.queryTypes, queryType)
}

// GetSupportedQueryTypes returns all supported query types
func (h *BaseQueryHandler) GetSupportedQueryTypes() []string {
	types := make([]string, 0, len(h.queryTypes))
	for queryType := range h.queryTypes {
		types = append(types, queryType)
	}
	return types
}

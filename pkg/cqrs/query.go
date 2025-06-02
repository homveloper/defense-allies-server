package cqrs

import (
	"context"
	"time"
)

// Query interface
type Query interface {
	// Basic identification information
	QueryID() string      // Unique query ID
	QueryType() string    // Query type

	// Metadata
	Timestamp() time.Time  // Query creation time
	UserID() string        // Query executing user
	CorrelationID() string // Correlation tracking ID

	// Query conditions
	GetCriteria() interface{} // Query criteria
	GetPagination() *Pagination // Pagination information
	GetSorting() *Sorting      // Sorting information

	// Validation
	Validate() error // Query validation
}

// Pagination information
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Offset   int `json:"offset"`
	Limit    int `json:"limit"`
}

// SortOrder represents sorting direction
type SortOrder int

const (
	Ascending SortOrder = iota
	Descending
)

func (so SortOrder) String() string {
	switch so {
	case Ascending:
		return "asc"
	case Descending:
		return "desc"
	default:
		return "asc"
	}
}

// Sorting information
type Sorting struct {
	Field string    `json:"field"`
	Order SortOrder `json:"order"`
}

// QueryResult represents query execution result
type QueryResult struct {
	Success       bool            `json:"success"`
	Data          interface{}     `json:"data"`
	Error         error           `json:"error,omitempty"`
	TotalCount    int64           `json:"total_count,omitempty"`
	Page          int             `json:"page,omitempty"`
	PageSize      int             `json:"page_size,omitempty"`
	ExecutionTime time.Duration   `json:"execution_time"`
}

// QueryHandler interface for handling queries
type QueryHandler interface {
	Handle(ctx context.Context, query Query) (*QueryResult, error)
	CanHandle(queryType string) bool
	GetHandlerName() string
}

// QueryDispatcher interface for dispatching queries
type QueryDispatcher interface {
	Dispatch(ctx context.Context, query Query) (*QueryResult, error)
	RegisterHandler(queryType string, handler QueryHandler) error
	UnregisterHandler(queryType string) error
}

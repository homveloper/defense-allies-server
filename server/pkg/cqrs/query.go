package cqrs

import (
	"context"
	"fmt"
	"time"
)

// Query interface
type Query interface {
	// Basic identification information
	QueryID() string   // Unique query ID
	QueryType() string // Query type

	// Metadata
	Timestamp() time.Time  // Query creation time
	CorrelationID() string // Correlation tracking ID

	// Query conditions
	GetCriteria() interface{}           // Query criteria
	GetPagination() *Pagination         // Pagination information
	GetSorting() *Sorting               // Sorting information (backward compatibility)
	GetSortingOptions() *SortingOptions // Advanced sorting options

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

// SortField represents a single field sorting specification
type SortField struct {
	Field     string                 `json:"field"`                // Field name to sort by
	Order     SortOrder              `json:"order"`                // Sort direction (asc/desc)
	Priority  int                    `json:"priority,omitempty"`   // Sort priority (1=highest, 2=second, etc.)
	NullsLast bool                   `json:"nulls_last,omitempty"` // Whether to put null values last
	Transform string                 `json:"transform,omitempty"`  // Optional transformation (e.g., "lower", "abs")
	Metadata  map[string]interface{} `json:"metadata,omitempty"`   // Additional sorting metadata
}

// SortingOptions represents comprehensive sorting configuration
type SortingOptions struct {
	Fields       []SortField            `json:"fields"`                  // Multiple field sorting
	DefaultField string                 `json:"default_field,omitempty"` // Default field if no sorting specified
	DefaultOrder SortOrder              `json:"default_order,omitempty"` // Default order if no sorting specified
	MaxFields    int                    `json:"max_fields,omitempty"`    // Maximum number of sort fields allowed
	Metadata     map[string]interface{} `json:"metadata,omitempty"`      // Additional sorting configuration
}

// Sorting information (backward compatibility)
type Sorting struct {
	Field string    `json:"field"`
	Order SortOrder `json:"order"`
}

// SortingOptions methods for building and validation

// NewSortingOptions creates a new SortingOptions instance
func NewSortingOptions() *SortingOptions {
	return &SortingOptions{
		Fields:       make([]SortField, 0),
		DefaultOrder: Ascending,
		MaxFields:    10, // Default maximum
		Metadata:     make(map[string]interface{}),
	}
}

// AddField adds a sorting field with the specified parameters
func (so *SortingOptions) AddField(field string, order SortOrder, priority int) *SortingOptions {
	so.Fields = append(so.Fields, SortField{
		Field:    field,
		Order:    order,
		Priority: priority,
	})
	return so
}

// AddFieldWithTransform adds a sorting field with transformation
func (so *SortingOptions) AddFieldWithTransform(field string, order SortOrder, priority int, transform string) *SortingOptions {
	so.Fields = append(so.Fields, SortField{
		Field:     field,
		Order:     order,
		Priority:  priority,
		Transform: transform,
	})
	return so
}

// SetDefaultField sets the default sorting field and order
func (so *SortingOptions) SetDefaultField(field string, order SortOrder) *SortingOptions {
	so.DefaultField = field
	so.DefaultOrder = order
	return so
}

// SetMaxFields sets the maximum number of sorting fields allowed
func (so *SortingOptions) SetMaxFields(maxFields int) *SortingOptions {
	so.MaxFields = maxFields
	return so
}

// Validate validates the sorting options
func (so *SortingOptions) Validate() error {
	if so == nil {
		return nil // nil sorting options is valid
	}

	// Check maximum fields limit
	if so.MaxFields > 0 && len(so.Fields) > so.MaxFields {
		return fmt.Errorf("too many sort fields: %d (max: %d)", len(so.Fields), so.MaxFields)
	}

	// Validate each field
	for i, field := range so.Fields {
		if field.Field == "" {
			return fmt.Errorf("sort field %d: field name cannot be empty", i)
		}
		if field.Priority < 0 {
			return fmt.Errorf("sort field %d: priority cannot be negative", i)
		}
	}

	return nil
}

// GetSortedFields returns fields sorted by priority
func (so *SortingOptions) GetSortedFields() []SortField {
	if so == nil || len(so.Fields) == 0 {
		return nil
	}

	// Create a copy and sort by priority
	fields := make([]SortField, len(so.Fields))
	copy(fields, so.Fields)

	// Sort by priority (lower number = higher priority)
	for i := 0; i < len(fields)-1; i++ {
		for j := i + 1; j < len(fields); j++ {
			if fields[i].Priority > fields[j].Priority {
				fields[i], fields[j] = fields[j], fields[i]
			}
		}
	}

	return fields
}

// ToLegacySorting converts to legacy Sorting format (first field only)
func (so *SortingOptions) ToLegacySorting() *Sorting {
	if so == nil || len(so.Fields) == 0 {
		if so != nil && so.DefaultField != "" {
			return &Sorting{
				Field: so.DefaultField,
				Order: so.DefaultOrder,
			}
		}
		return nil
	}

	sortedFields := so.GetSortedFields()
	if len(sortedFields) > 0 {
		return &Sorting{
			Field: sortedFields[0].Field,
			Order: sortedFields[0].Order,
		}
	}

	return nil
}

// QueryResult represents query execution result
type QueryResult struct {
	Success       bool          `json:"success"`
	Data          interface{}   `json:"data"`
	Error         error         `json:"error,omitempty"`
	TotalCount    int64         `json:"total_count,omitempty"`
	Page          int           `json:"page,omitempty"`
	PageSize      int           `json:"page_size,omitempty"`
	ExecutionTime time.Duration `json:"execution_time"`
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

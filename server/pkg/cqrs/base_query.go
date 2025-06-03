package cqrs

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// BaseQuery provides a base implementation of Query interface
type BaseQuery struct {
	queryID        string
	queryType      string
	timestamp      time.Time
	correlationID  string
	criteria       interface{}
	pagination     *Pagination
	sorting        *Sorting        // Legacy sorting (backward compatibility)
	sortingOptions *SortingOptions // Advanced sorting options
}

// NewBaseQuery creates a new BaseQuery
func NewBaseQuery(queryType string, criteria interface{}) *BaseQuery {
	return &BaseQuery{
		queryID:   uuid.New().String(),
		queryType: queryType,
		timestamp: time.Now(),
		criteria:  criteria,
	}
}

// Query interface implementation

func (q *BaseQuery) QueryID() string {
	return q.queryID
}

func (q *BaseQuery) QueryType() string {
	return q.queryType
}

func (q *BaseQuery) Timestamp() time.Time {
	return q.timestamp
}

func (q *BaseQuery) CorrelationID() string {
	return q.correlationID
}

func (q *BaseQuery) GetCriteria() interface{} {
	return q.criteria
}

func (q *BaseQuery) GetPagination() *Pagination {
	return q.pagination
}

func (q *BaseQuery) GetSorting() *Sorting {
	return q.sorting
}

func (q *BaseQuery) GetSortingOptions() *SortingOptions {
	return q.sortingOptions
}

func (q *BaseQuery) Validate() error {
	if q.queryID == "" {
		return fmt.Errorf("query ID cannot be empty")
	}
	if q.queryType == "" {
		return fmt.Errorf("query type cannot be empty")
	}

	// Note: UserID is not part of the base Query interface
	// Domain-specific query implementations should add UserID fields and validation
	// if required for their specific use case (e.g., user-specific queries)

	// Validate pagination if provided
	if q.pagination != nil {
		if q.pagination.PageSize <= 0 {
			return fmt.Errorf("page size must be greater than 0")
		}
		if q.pagination.Page < 0 {
			return fmt.Errorf("page must be non-negative")
		}
	}

	return nil
}

// Helper methods

// SetQueryID sets the query ID (used when loading from storage)
func (q *BaseQuery) SetQueryID(queryID string) {
	q.queryID = queryID
}

// SetTimestamp sets the timestamp (used when loading from storage)
func (q *BaseQuery) SetTimestamp(timestamp time.Time) {
	q.timestamp = timestamp
}

// SetCorrelationID sets the correlation ID
func (q *BaseQuery) SetCorrelationID(correlationID string) {
	q.correlationID = correlationID
}

// SetCriteria sets the query criteria
func (q *BaseQuery) SetCriteria(criteria interface{}) {
	q.criteria = criteria
}

// SetPagination sets the pagination information
func (q *BaseQuery) SetPagination(pagination *Pagination) {
	q.pagination = pagination
}

// SetSorting sets the sorting information
func (q *BaseQuery) SetSorting(sorting *Sorting) {
	q.sorting = sorting
}

// SetSortingOptions sets the advanced sorting options
func (q *BaseQuery) SetSortingOptions(sortingOptions *SortingOptions) {
	q.sortingOptions = sortingOptions
}

// WithPagination is a fluent method to set pagination
func (q *BaseQuery) WithPagination(page, pageSize int) *BaseQuery {
	q.pagination = &Pagination{
		Page:     page,
		PageSize: pageSize,
		Offset:   page * pageSize,
		Limit:    pageSize,
	}
	return q
}

// WithSorting is a fluent method to set sorting
func (q *BaseQuery) WithSorting(field string, order SortOrder) *BaseQuery {
	q.sorting = &Sorting{
		Field: field,
		Order: order,
	}
	return q
}

// WithAdvancedSorting is a fluent method to set advanced sorting options
func (q *BaseQuery) WithAdvancedSorting(sortingOptions *SortingOptions) *BaseQuery {
	q.sortingOptions = sortingOptions
	return q
}

// WithMultiFieldSorting is a fluent method to create multi-field sorting
func (q *BaseQuery) WithMultiFieldSorting() *MultiFieldSortingBuilder {
	return &MultiFieldSortingBuilder{
		query:   q,
		options: NewSortingOptions(),
	}
}

// MultiFieldSortingBuilder provides a fluent API for building multi-field sorting
type MultiFieldSortingBuilder struct {
	query   *BaseQuery
	options *SortingOptions
}

// AddField adds a sorting field
func (msb *MultiFieldSortingBuilder) AddField(field string, order SortOrder, priority int) *MultiFieldSortingBuilder {
	msb.options.AddField(field, order, priority)
	return msb
}

// AddFieldWithTransform adds a sorting field with transformation
func (msb *MultiFieldSortingBuilder) AddFieldWithTransform(field string, order SortOrder, priority int, transform string) *MultiFieldSortingBuilder {
	msb.options.AddFieldWithTransform(field, order, priority, transform)
	return msb
}

// SetDefault sets the default sorting field and order
func (msb *MultiFieldSortingBuilder) SetDefault(field string, order SortOrder) *MultiFieldSortingBuilder {
	msb.options.SetDefaultField(field, order)
	return msb
}

// SetMaxFields sets the maximum number of sorting fields
func (msb *MultiFieldSortingBuilder) SetMaxFields(maxFields int) *MultiFieldSortingBuilder {
	msb.options.SetMaxFields(maxFields)
	return msb
}

// Build finalizes the sorting configuration and returns the query
func (msb *MultiFieldSortingBuilder) Build() *BaseQuery {
	msb.query.sortingOptions = msb.options
	return msb.query
}

// GetQueryInfo returns basic query information as a map
func (q *BaseQuery) GetQueryInfo() map[string]interface{} {
	info := map[string]interface{}{
		"query_id":       q.queryID,
		"query_type":     q.queryType,
		"timestamp":      q.timestamp,
		"correlation_id": q.correlationID,
	}

	if q.pagination != nil {
		info["pagination"] = q.pagination
	}

	if q.sorting != nil {
		info["sorting"] = q.sorting
	}

	if q.sortingOptions != nil {
		info["sorting_options"] = q.sortingOptions
	}

	return info
}

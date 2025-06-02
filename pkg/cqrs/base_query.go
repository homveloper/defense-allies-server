package cqrs

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// BaseQuery provides a base implementation of Query interface
type BaseQuery struct {
	queryID       string
	queryType     string
	timestamp     time.Time
	userID        string
	correlationID string
	criteria      interface{}
	pagination    *Pagination
	sorting       *Sorting
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

func (q *BaseQuery) UserID() string {
	return q.userID
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

func (q *BaseQuery) Validate() error {
	if q.queryID == "" {
		return fmt.Errorf("query ID cannot be empty")
	}
	if q.queryType == "" {
		return fmt.Errorf("query type cannot be empty")
	}
	
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

// SetUserID sets the user ID
func (q *BaseQuery) SetUserID(userID string) {
	q.userID = userID
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

// GetQueryInfo returns basic query information as a map
func (q *BaseQuery) GetQueryInfo() map[string]interface{} {
	info := map[string]interface{}{
		"query_id":       q.queryID,
		"query_type":     q.queryType,
		"timestamp":      q.timestamp,
		"user_id":        q.userID,
		"correlation_id": q.correlationID,
	}
	
	if q.pagination != nil {
		info["pagination"] = q.pagination
	}
	
	if q.sorting != nil {
		info["sorting"] = q.sorting
	}
	
	return info
}

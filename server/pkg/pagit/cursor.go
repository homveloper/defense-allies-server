package pagit

import (
	"context"
	"errors"
)

// CursorRequest represents cursor-based pagination parameters
type CursorRequest struct {
	PageSize int    `json:"page_size" query:"page_size" form:"page_size"`
	Cursor   string `json:"cursor,omitempty" query:"cursor" form:"cursor"`
}

// CursorResponse represents cursor-based paginated results
type CursorResponse[T any] struct {
	Items      []T    `json:"items"`
	PageSize   int    `json:"page_size"`
	HasNext    bool   `json:"has_next"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
}

// CursorOptions for cursor-based pagination behavior
type CursorOptions struct {
	DefaultPageSize int
	MaxPageSize     int
	MinPageSize     int
}

var DefaultCursorOptions = CursorOptions{
	DefaultPageSize: 20,
	MaxPageSize:     100,
	MinPageSize:     1,
}

// CursorAdapter interface for cursor-based pagination
type CursorAdapter[T any] interface {
	// FetchWithCursor retrieves items using cursor-based pagination
	FetchWithCursor(ctx context.Context, cursor string, limit int) ([]T, string, error)
}

// PaginateCursor performs cursor-based pagination
func PaginateCursor[T any](ctx context.Context, req CursorRequest, adapter CursorAdapter[T], opts ...CursorOptions) (*CursorResponse[T], error) {
	options := DefaultCursorOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	// Validate page size
	if req.PageSize < options.MinPageSize {
		req.PageSize = options.DefaultPageSize
	}
	if req.PageSize > options.MaxPageSize {
		req.PageSize = options.MaxPageSize
	}

	// Fetch items with cursor (fetch one extra to check if there's next page)
	items, nextCursor, err := adapter.FetchWithCursor(ctx, req.Cursor, req.PageSize+1)
	if err != nil {
		return nil, err
	}

	// Check if there are more items
	hasNext := len(items) > req.PageSize
	if hasNext {
		items = items[:req.PageSize]
	}

	return &CursorResponse[T]{
		Items:      items,
		PageSize:   req.PageSize,
		HasNext:    hasNext,
		NextCursor: nextCursor,
		PrevCursor: req.Cursor,
	}, nil
}

// Common cursor errors
var (
	ErrInvalidCursor = errors.New("invalid cursor")
)

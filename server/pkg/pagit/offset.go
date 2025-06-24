package pagit

import (
	"context"
	"errors"
)

// OffsetRequest represents offset-based pagination parameters
type OffsetRequest struct {
	Page     int `json:"page" query:"page" form:"page"`
	PageSize int `json:"page_size" query:"page_size" form:"page_size"`
}

// OffsetResponse represents offset-based paginated results
type OffsetResponse[T any] struct {
	Items      []T   `json:"items"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// OffsetOptions for offset-based pagination behavior
type OffsetOptions struct {
	DefaultPageSize int
	MaxPageSize     int
	MinPageSize     int
}

var DefaultOffsetOptions = OffsetOptions{
	DefaultPageSize: 20,
	MaxPageSize:     100,
	MinPageSize:     1,
}

// OffsetAdapter interface for offset-based pagination
type OffsetAdapter[T any] interface {
	// Count returns total number of items
	Count(ctx context.Context) (int64, error)
	// Fetch retrieves items for the given page
	Fetch(ctx context.Context, offset, limit int) ([]T, error)
}

// Paginate performs offset-based pagination
func Paginate[T any](ctx context.Context, req OffsetRequest, adapter OffsetAdapter[T], opts ...OffsetOptions) (*OffsetResponse[T], error) {
	options := DefaultOffsetOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	// Validate and normalize request
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < options.MinPageSize {
		req.PageSize = options.DefaultPageSize
	}
	if req.PageSize > options.MaxPageSize {
		req.PageSize = options.MaxPageSize
	}

	// Get total count
	total, err := adapter.Count(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	offset := (req.Page - 1) * req.PageSize

	// Fetch items
	items, err := adapter.Fetch(ctx, offset, req.PageSize)
	if err != nil {
		return nil, err
	}

	return &OffsetResponse[T]{
		Items:      items,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1,
	}, nil
}

// Common offset errors
var (
	ErrInvalidPage     = errors.New("invalid page number")
	ErrInvalidPageSize = errors.New("invalid page size")
)

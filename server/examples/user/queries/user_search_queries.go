package queries

import (
	"fmt"
	"strings"
	"time"

	"defense-allies-server/pkg/cqrs"
)

// SearchUsersQuery represents a query to search users with various filters
type SearchUsersQuery struct {
	*cqrs.BaseQuery

	// Text search
	SearchText string `json:"search_text,omitempty"`

	// Basic filters
	Status    string   `json:"status,omitempty"`    // active, inactive, deactivated
	Roles     []string `json:"roles,omitempty"`     // Filter by roles
	Cities    []string `json:"cities,omitempty"`    // Filter by cities
	Countries []string `json:"countries,omitempty"` // Filter by countries

	// Date filters
	CreatedAfter   *time.Time `json:"created_after,omitempty"`
	CreatedBefore  *time.Time `json:"created_before,omitempty"`
	LastLoginAfter *time.Time `json:"last_login_after,omitempty"`

	// Pagination
	Offset int `json:"offset"`
	Limit  int `json:"limit"`

	// Sorting
	SortBy    string `json:"sort_by"`    // name, email, created_at, last_login_at
	SortOrder string `json:"sort_order"` // asc, desc
}

// NewSearchUsersQuery creates a new SearchUsersQuery
func NewSearchUsersQuery() *SearchUsersQuery {
	query := &SearchUsersQuery{
		BaseQuery: cqrs.NewBaseQuery(
			"SearchUsers",
			map[string]interface{}{},
		),
		Limit:     50, // Default limit
		SortBy:    "created_at",
		SortOrder: "desc",
	}
	return query
}

// WithSearchText sets the search text
func (q *SearchUsersQuery) WithSearchText(text string) *SearchUsersQuery {
	q.SearchText = strings.TrimSpace(text)
	return q
}

// WithStatus sets the status filter
func (q *SearchUsersQuery) WithStatus(status string) *SearchUsersQuery {
	q.Status = status
	return q
}

// WithRoles sets the roles filter
func (q *SearchUsersQuery) WithRoles(roles ...string) *SearchUsersQuery {
	q.Roles = roles
	return q
}

// WithCities sets the cities filter
func (q *SearchUsersQuery) WithCities(cities ...string) *SearchUsersQuery {
	q.Cities = cities
	return q
}

// WithCountries sets the countries filter
func (q *SearchUsersQuery) WithCountries(countries ...string) *SearchUsersQuery {
	q.Countries = countries
	return q
}

// WithCreatedDateRange sets the created date range filter
func (q *SearchUsersQuery) WithCreatedDateRange(after, before *time.Time) *SearchUsersQuery {
	q.CreatedAfter = after
	q.CreatedBefore = before
	return q
}

// WithLastLoginAfter sets the last login after filter
func (q *SearchUsersQuery) WithLastLoginAfter(after *time.Time) *SearchUsersQuery {
	q.LastLoginAfter = after
	return q
}

// WithPagination sets pagination parameters
func (q *SearchUsersQuery) WithPagination(offset, limit int) *SearchUsersQuery {
	q.Offset = offset
	q.Limit = limit
	return q
}

// WithSorting sets sorting parameters
func (q *SearchUsersQuery) WithSorting(sortBy, sortOrder string) *SearchUsersQuery {
	q.SortBy = sortBy
	q.SortOrder = sortOrder
	return q
}

// Validate validates the query parameters
func (q *SearchUsersQuery) Validate() error {
	if err := q.BaseQuery.Validate(); err != nil {
		return err
	}

	if q.Limit <= 0 || q.Limit > 1000 {
		return fmt.Errorf("limit must be between 1 and 1000")
	}

	if q.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}

	validSortFields := map[string]bool{
		"name":          true,
		"email":         true,
		"created_at":    true,
		"last_login_at": true,
		"status":        true,
		"display_name":  true,
	}

	if !validSortFields[q.SortBy] {
		return fmt.Errorf("invalid sort field: %s", q.SortBy)
	}

	if q.SortOrder != "asc" && q.SortOrder != "desc" {
		return fmt.Errorf("sort order must be 'asc' or 'desc'")
	}

	return nil
}

// GetUserByIDQuery represents a query to get a user by ID
type GetUserByIDQuery struct {
	*cqrs.BaseQuery
	TargetUserID string `json:"target_user_id"` // Renamed to avoid conflict with UserID() method
}

// NewGetUserByIDQuery creates a new GetUserByIDQuery
func NewGetUserByIDQuery(userID string) *GetUserByIDQuery {
	query := &GetUserByIDQuery{
		BaseQuery: cqrs.NewBaseQuery(
			"GetUserByID",
			map[string]interface{}{
				"target_user_id": userID,
			},
		),
		TargetUserID: userID,
	}

	return query
}

// GetTargetUserID returns the target user ID to query
func (q *GetUserByIDQuery) GetTargetUserID() string {
	return q.TargetUserID
}

// Validate validates the query
func (q *GetUserByIDQuery) Validate() error {
	if err := q.BaseQuery.Validate(); err != nil {
		return err
	}

	if strings.TrimSpace(q.TargetUserID) == "" {
		return fmt.Errorf("target user ID cannot be empty")
	}

	return nil
}

// GetUsersByRoleQuery represents a query to get users by role
type GetUsersByRoleQuery struct {
	*cqrs.BaseQuery
	Role      string `json:"role"`
	Status    string `json:"status,omitempty"` // Optional status filter
	Offset    int    `json:"offset"`
	Limit     int    `json:"limit"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
}

// NewGetUsersByRoleQuery creates a new GetUsersByRoleQuery
func NewGetUsersByRoleQuery(role string) *GetUsersByRoleQuery {
	query := &GetUsersByRoleQuery{
		BaseQuery: cqrs.NewBaseQuery(
			"GetUsersByRole",
			map[string]interface{}{
				"role": role,
			},
		),
		Role:      role,
		Limit:     50,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
	return query
}

// WithStatus sets the status filter
func (q *GetUsersByRoleQuery) WithStatus(status string) *GetUsersByRoleQuery {
	q.Status = status
	return q
}

// WithPagination sets pagination parameters
func (q *GetUsersByRoleQuery) WithPagination(offset, limit int) *GetUsersByRoleQuery {
	q.Offset = offset
	q.Limit = limit
	return q
}

// WithSorting sets sorting parameters
func (q *GetUsersByRoleQuery) WithSorting(sortBy, sortOrder string) *GetUsersByRoleQuery {
	q.SortBy = sortBy
	q.SortOrder = sortOrder
	return q
}

// Validate validates the query
func (q *GetUsersByRoleQuery) Validate() error {
	if err := q.BaseQuery.Validate(); err != nil {
		return err
	}

	if strings.TrimSpace(q.Role) == "" {
		return fmt.Errorf("role cannot be empty")
	}

	if q.Limit <= 0 || q.Limit > 1000 {
		return fmt.Errorf("limit must be between 1 and 1000")
	}

	if q.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}

	return nil
}

// UserSearchResult represents the result of a user search
type UserSearchResult struct {
	Users      []UserSearchItem `json:"users"`
	TotalCount int64            `json:"total_count"`
	Offset     int              `json:"offset"`
	Limit      int              `json:"limit"`
	HasMore    bool             `json:"has_more"`
}

// UserSearchItem represents a user item in search results
type UserSearchItem struct {
	UserID      string     `json:"user_id"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	DisplayName string     `json:"display_name"`
	Status      string     `json:"status"`
	Roles       []string   `json:"roles"`
	Avatar      string     `json:"avatar,omitempty"`
	City        string     `json:"city,omitempty"`
	Country     string     `json:"country,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

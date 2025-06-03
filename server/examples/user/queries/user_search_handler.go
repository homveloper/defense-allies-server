package queries

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"defense-allies-server/examples/user/projections"
	"defense-allies-server/pkg/cqrs"
)

// UserSearchHandler handles user search queries
type UserSearchHandler struct {
	readStore cqrs.ReadStore
}

// NewUserSearchHandler creates a new UserSearchHandler
func NewUserSearchHandler(readStore cqrs.ReadStore) *UserSearchHandler {
	return &UserSearchHandler{
		readStore: readStore,
	}
}

// Handle handles the search users query
func (h *UserSearchHandler) Handle(ctx context.Context, query cqrs.Query) (*cqrs.QueryResult, error) {
	var result interface{}
	var err error

	switch q := query.(type) {
	case *SearchUsersQuery:
		result, err = h.handleSearchUsers(ctx, q)
	case *GetUserByIDQuery:
		result, err = h.handleGetUserByID(ctx, q)
	case *GetUsersByRoleQuery:
		result, err = h.handleGetUsersByRole(ctx, q)
	default:
		err = fmt.Errorf("unsupported query type: %T", query)
	}

	if err != nil {
		return &cqrs.QueryResult{
			Success: false,
			Error:   err,
		}, err
	}

	return &cqrs.QueryResult{
		Success: true,
		Data:    result,
	}, nil
}

// CanHandle checks if this handler can handle the given query type
func (h *UserSearchHandler) CanHandle(queryType string) bool {
	switch queryType {
	case "SearchUsers", "GetUserByID", "GetUsersByRole":
		return true
	default:
		return false
	}
}

// GetHandlerName returns the name of this handler
func (h *UserSearchHandler) GetHandlerName() string {
	return "UserSearchHandler"
}

// handleSearchUsers handles the search users query
func (h *UserSearchHandler) handleSearchUsers(ctx context.Context, query *SearchUsersQuery) (*UserSearchResult, error) {
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	// Get all users from read store
	allUsers, err := h.getAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	// Apply filters
	filteredUsers := h.applyFilters(allUsers, query)

	// Apply sorting
	h.applySorting(filteredUsers, query.SortBy, query.SortOrder)

	// Calculate pagination
	totalCount := int64(len(filteredUsers))
	start := query.Offset
	end := start + query.Limit

	if start > len(filteredUsers) {
		start = len(filteredUsers)
	}
	if end > len(filteredUsers) {
		end = len(filteredUsers)
	}

	paginatedUsers := filteredUsers[start:end]

	// Convert to search items
	searchItems := make([]UserSearchItem, len(paginatedUsers))
	for i, user := range paginatedUsers {
		searchItems[i] = h.convertToSearchItem(user)
	}

	return &UserSearchResult{
		Users:      searchItems,
		TotalCount: totalCount,
		Offset:     query.Offset,
		Limit:      query.Limit,
		HasMore:    end < len(filteredUsers),
	}, nil
}

// handleGetUserByID handles the get user by ID query
func (h *UserSearchHandler) handleGetUserByID(ctx context.Context, query *GetUserByIDQuery) (*projections.UserView, error) {
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	readModel, err := h.readStore.GetByID(ctx, query.GetTargetUserID(), "UserView")
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	userView, ok := readModel.(*projections.UserView)
	if !ok {
		return nil, fmt.Errorf("invalid read model type: %T", readModel)
	}

	return userView, nil
}

// handleGetUsersByRole handles the get users by role query
func (h *UserSearchHandler) handleGetUsersByRole(ctx context.Context, query *GetUsersByRoleQuery) (*UserSearchResult, error) {
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	// Get all users from read store
	allUsers, err := h.getAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	// Filter by role
	var filteredUsers []*projections.UserView
	for _, user := range allUsers {
		if user.HasRole(query.Role) {
			// Apply status filter if specified
			if query.Status == "" || user.Status == query.Status {
				filteredUsers = append(filteredUsers, user)
			}
		}
	}

	// Apply sorting
	h.applySorting(filteredUsers, query.SortBy, query.SortOrder)

	// Calculate pagination
	totalCount := int64(len(filteredUsers))
	start := query.Offset
	end := start + query.Limit

	if start > len(filteredUsers) {
		start = len(filteredUsers)
	}
	if end > len(filteredUsers) {
		end = len(filteredUsers)
	}

	paginatedUsers := filteredUsers[start:end]

	// Convert to search items
	searchItems := make([]UserSearchItem, len(paginatedUsers))
	for i, user := range paginatedUsers {
		searchItems[i] = h.convertToSearchItem(user)
	}

	return &UserSearchResult{
		Users:      searchItems,
		TotalCount: totalCount,
		Offset:     query.Offset,
		Limit:      query.Limit,
		HasMore:    end < len(filteredUsers),
	}, nil
}

// getAllUsers gets all users from the read store
func (h *UserSearchHandler) getAllUsers(ctx context.Context) ([]*projections.UserView, error) {
	// Use the InMemoryReadStore's GetModelsByType method for simplicity
	// In a real system, you might want to implement a more efficient way to get all users
	// or use database-specific query capabilities

	// Type assertion to access InMemoryReadStore specific methods
	if inMemoryStore, ok := h.readStore.(*cqrs.InMemoryReadStore); ok {
		readModels := inMemoryStore.GetModelsByType("UserView")

		var users []*projections.UserView
		for _, readModel := range readModels {
			if userView, ok := readModel.(*projections.UserView); ok {
				users = append(users, userView)
			}
		}

		return users, nil
	}

	// Fallback to Query method for other store types
	criteria := cqrs.QueryCriteria{
		Filters: map[string]interface{}{},
		Limit:   1000, // Set a reasonable limit
	}

	readModels, err := h.readStore.Query(ctx, criteria)
	if err != nil {
		return nil, err
	}

	var users []*projections.UserView
	for _, readModel := range readModels {
		if userView, ok := readModel.(*projections.UserView); ok {
			users = append(users, userView)
		}
	}

	return users, nil
}

// applyFilters applies search filters to the user list
func (h *UserSearchHandler) applyFilters(users []*projections.UserView, query *SearchUsersQuery) []*projections.UserView {
	var filtered []*projections.UserView

	for _, user := range users {
		if h.matchesFilters(user, query) {
			filtered = append(filtered, user)
		}
	}

	return filtered
}

// matchesFilters checks if a user matches the search filters
func (h *UserSearchHandler) matchesFilters(user *projections.UserView, query *SearchUsersQuery) bool {
	// Text search
	if query.SearchText != "" {
		searchText := strings.ToLower(query.SearchText)
		userText := strings.ToLower(user.SearchableText)
		if !strings.Contains(userText, searchText) {
			return false
		}
	}

	// Status filter
	if query.Status != "" && user.Status != query.Status {
		return false
	}

	// Roles filter
	if len(query.Roles) > 0 {
		hasRole := false
		for _, role := range query.Roles {
			if user.HasRole(role) {
				hasRole = true
				break
			}
		}
		if !hasRole {
			return false
		}
	}

	// Cities filter
	if len(query.Cities) > 0 {
		hasCity := false
		for _, city := range query.Cities {
			if strings.EqualFold(user.City, city) {
				hasCity = true
				break
			}
		}
		if !hasCity {
			return false
		}
	}

	// Countries filter
	if len(query.Countries) > 0 {
		hasCountry := false
		for _, country := range query.Countries {
			if strings.EqualFold(user.Country, country) {
				hasCountry = true
				break
			}
		}
		if !hasCountry {
			return false
		}
	}

	// Date filters
	if query.CreatedAfter != nil && user.CreatedAt.Before(*query.CreatedAfter) {
		return false
	}

	if query.CreatedBefore != nil && user.CreatedAt.After(*query.CreatedBefore) {
		return false
	}

	if query.LastLoginAfter != nil {
		if user.LastLoginAt == nil || user.LastLoginAt.Before(*query.LastLoginAfter) {
			return false
		}
	}

	return true
}

// applySorting applies sorting to the user list
func (h *UserSearchHandler) applySorting(users []*projections.UserView, sortBy, sortOrder string) {
	sort.Slice(users, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "name":
			less = users[i].Name < users[j].Name
		case "email":
			less = users[i].Email < users[j].Email
		case "display_name":
			less = users[i].DisplayName < users[j].DisplayName
		case "status":
			less = users[i].Status < users[j].Status
		case "last_login_at":
			if users[i].LastLoginAt == nil && users[j].LastLoginAt == nil {
				less = false
			} else if users[i].LastLoginAt == nil {
				less = true
			} else if users[j].LastLoginAt == nil {
				less = false
			} else {
				less = users[i].LastLoginAt.Before(*users[j].LastLoginAt)
			}
		default: // created_at
			less = users[i].CreatedAt.Before(users[j].CreatedAt)
		}

		if sortOrder == "desc" {
			return !less
		}
		return less
	})
}

// convertToSearchItem converts a UserView to a UserSearchItem
func (h *UserSearchHandler) convertToSearchItem(user *projections.UserView) UserSearchItem {
	return UserSearchItem{
		UserID:      user.UserID,
		Email:       user.Email,
		Name:        user.Name,
		DisplayName: user.DisplayName,
		Status:      user.Status,
		Roles:       user.Roles,
		Avatar:      user.Avatar,
		City:        user.City,
		Country:     user.Country,
		CreatedAt:   user.CreatedAt,
		LastLoginAt: user.LastLoginAt,
	}
}

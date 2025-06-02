package handlers

import (
	"context"
	"fmt"
	"time"

	"defense-allies-server/examples/user/projections"
	"defense-allies-server/pkg/cqrs"
)

// UserQueryHandler handles user-related queries
type UserQueryHandler struct {
	readStore cqrs.ReadStore
}

// NewUserQueryHandler creates a new UserQueryHandler
func NewUserQueryHandler(readStore cqrs.ReadStore) *UserQueryHandler {
	return &UserQueryHandler{
		readStore: readStore,
	}
}

// Handle handles the query
func (h *UserQueryHandler) Handle(ctx context.Context, query cqrs.Query) (*cqrs.QueryResult, error) {
	startTime := time.Now()

	// Validate query
	if err := query.Validate(); err != nil {
		return &cqrs.QueryResult{
			Success:       false,
			Error:         fmt.Errorf("query validation failed: %w", err),
			ExecutionTime: time.Since(startTime),
		}, nil
	}

	var result *cqrs.QueryResult
	var err error

	// Handle based on query type string instead of type assertion
	switch query.QueryType() {
	case "GetUser":
		if userQuery, ok := query.(*projections.UserViewQuery); ok {
			result, err = h.handleGetUser(ctx, userQuery)
		} else {
			return &cqrs.QueryResult{
				Success:       false,
				Error:         fmt.Errorf("invalid query type for GetUser: %T", query),
				ExecutionTime: time.Since(startTime),
			}, nil
		}
	case "ListUsers":
		if userQuery, ok := query.(*projections.UserViewQuery); ok {
			result, err = h.handleListUsers(ctx, userQuery)
		} else {
			return &cqrs.QueryResult{
				Success:       false,
				Error:         fmt.Errorf("invalid query type for ListUsers: %T", query),
				ExecutionTime: time.Since(startTime),
			}, nil
		}
	default:
		return &cqrs.QueryResult{
			Success:       false,
			Error:         fmt.Errorf("unsupported query type: %s", query.QueryType()),
			ExecutionTime: time.Since(startTime),
		}, nil
	}

	if err != nil {
		return &cqrs.QueryResult{
			Success:       false,
			Error:         err,
			ExecutionTime: time.Since(startTime),
		}, nil
	}

	result.ExecutionTime = time.Since(startTime)
	return result, nil
}

// CanHandle returns true if the handler can handle the query type
func (h *UserQueryHandler) CanHandle(queryType string) bool {
	switch queryType {
	case "GetUser", "ListUsers":
		return true
	default:
		return false
	}
}

// GetHandlerName returns the handler name
func (h *UserQueryHandler) GetHandlerName() string {
	return "UserQueryHandler"
}

// handleGetUser handles GetUser query
func (h *UserQueryHandler) handleGetUser(ctx context.Context, query *projections.UserViewQuery) (*cqrs.QueryResult, error) {
	// Get user view from read store
	readModel, err := h.readStore.GetByID(ctx, query.TargetUserID, "UserView")
	if err != nil {
		return &cqrs.QueryResult{
			Success: false,
			Error:   fmt.Errorf("failed to get user: %w", err),
		}, nil
	}

	userView, ok := readModel.(*projections.UserView)
	if !ok {
		return &cqrs.QueryResult{
			Success: false,
			Error:   fmt.Errorf("invalid read model type: expected *projections.UserView, got %T", readModel),
		}, nil
	}

	return &cqrs.QueryResult{
		Success: true,
		Data:    userView,
	}, nil
}

// handleListUsers handles ListUsers query
func (h *UserQueryHandler) handleListUsers(ctx context.Context, query *projections.UserViewQuery) (*cqrs.QueryResult, error) {
	// Build query criteria
	criteria := cqrs.QueryCriteria{
		Filters: make(map[string]interface{}),
	}

	// Add status filter if specified
	if query.Status != "" {
		criteria.Filters["status"] = query.Status
	}

	// Add pagination if specified
	if pagination := query.GetPagination(); pagination != nil {
		criteria.Limit = pagination.Limit
		criteria.Offset = pagination.Offset
	} else {
		// Default pagination
		criteria.Limit = 50
		criteria.Offset = 0
	}

	// Add sorting
	criteria.SortBy = "created_at"
	criteria.SortOrder = cqrs.Descending

	// Query read models
	readModels, err := h.readStore.Query(ctx, criteria)
	if err != nil {
		return &cqrs.QueryResult{
			Success: false,
			Error:   fmt.Errorf("failed to query users: %w", err),
		}, nil
	}

	// Convert to UserView slice
	userViews := make([]*projections.UserView, 0, len(readModels))
	for _, readModel := range readModels {
		if userView, ok := readModel.(*projections.UserView); ok {
			userViews = append(userViews, userView)
		} else {
			return &cqrs.QueryResult{
				Success: false,
				Error:   fmt.Errorf("invalid read model type: expected *projections.UserView, got %T", readModel),
			}, nil
		}
	}

	// Get total count
	totalCount, err := h.readStore.Count(ctx, criteria)
	if err != nil {
		return &cqrs.QueryResult{
			Success: false,
			Error:   fmt.Errorf("failed to count users: %w", err),
		}, nil
	}

	// Calculate pagination info
	var page, pageSize int
	if pagination := query.GetPagination(); pagination != nil {
		page = pagination.Page
		pageSize = pagination.PageSize
	} else {
		page = 1
		pageSize = 50
	}

	return &cqrs.QueryResult{
		Success:    true,
		Data:       userViews,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// UserQueryDispatcher wraps the standard query dispatcher with user-specific functionality
type UserQueryDispatcher struct {
	dispatcher cqrs.QueryDispatcher
}

// NewUserQueryDispatcher creates a new UserQueryDispatcher
func NewUserQueryDispatcher(readStore cqrs.ReadStore) *UserQueryDispatcher {
	dispatcher := cqrs.NewInMemoryQueryDispatcher()

	// Register user query handler
	userHandler := NewUserQueryHandler(readStore)
	dispatcher.RegisterHandler("GetUser", userHandler)
	dispatcher.RegisterHandler("ListUsers", userHandler)

	return &UserQueryDispatcher{
		dispatcher: dispatcher,
	}
}

// Dispatch dispatches the query
func (d *UserQueryDispatcher) Dispatch(ctx context.Context, query cqrs.Query) (*cqrs.QueryResult, error) {
	return d.dispatcher.Dispatch(ctx, query)
}

// RegisterHandler registers a query handler
func (d *UserQueryDispatcher) RegisterHandler(queryType string, handler cqrs.QueryHandler) error {
	return d.dispatcher.RegisterHandler(queryType, handler)
}

// UnregisterHandler unregisters a query handler
func (d *UserQueryDispatcher) UnregisterHandler(queryType string) error {
	return d.dispatcher.UnregisterHandler(queryType)
}

// Helper functions for creating common queries

// CreateGetUserQuery creates a GetUser query
func CreateGetUserQuery(userID string) *projections.UserViewQuery {
	return projections.NewGetUserQuery(userID)
}

// CreateListUsersQuery creates a ListUsers query
func CreateListUsersQuery(status string, page, pageSize int) *projections.UserViewQuery {
	var pagination *cqrs.Pagination
	if page > 0 && pageSize > 0 {
		pagination = &cqrs.Pagination{
			Page:     page,
			PageSize: pageSize,
			Offset:   (page - 1) * pageSize,
			Limit:    pageSize,
		}
	}

	return projections.NewListUsersQuery(status, pagination)
}

// CreateListActiveUsersQuery creates a query to list active users
func CreateListActiveUsersQuery(page, pageSize int) *projections.UserViewQuery {
	return CreateListUsersQuery("active", page, pageSize)
}

// CreateListDeactivatedUsersQuery creates a query to list deactivated users
func CreateListDeactivatedUsersQuery(page, pageSize int) *projections.UserViewQuery {
	return CreateListUsersQuery("deactivated", page, pageSize)
}

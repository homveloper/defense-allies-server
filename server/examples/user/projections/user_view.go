package projections

import (
	"context"
	"fmt"
	"strings"
	"time"

	"defense-allies-server/examples/user/domain"
	"defense-allies-server/pkg/cqrs"
)

// UserView represents a read model for user data
type UserView struct {
	*cqrs.BaseReadModel
	UserID             string     `json:"user_id"`
	Email              string     `json:"email"`
	Name               string     `json:"name"`
	Status             string     `json:"status"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	LastLoginAt        *time.Time `json:"last_login_at,omitempty"`
	DeactivatedAt      *time.Time `json:"deactivated_at,omitempty"`
	DeactivationReason string     `json:"deactivation_reason,omitempty"`

	// Profile information
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio"`
	Avatar      string `json:"avatar"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Address     string `json:"address,omitempty"`
	City        string `json:"city,omitempty"`
	Country     string `json:"country,omitempty"`
	PostalCode  string `json:"postal_code,omitempty"`

	// Role information
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`

	// Searchable fields
	SearchableText string `json:"searchable_text"` // Combined text for full-text search
}

// NewUserView creates a new UserView
func NewUserView(userID string) *UserView {
	userView := &UserView{
		BaseReadModel: cqrs.NewBaseReadModel(userID, "UserView", map[string]interface{}{}), // Initialize with empty map
		UserID:        userID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	return userView
}

// GetData returns the UserView data as a map for serialization
func (uv *UserView) GetData() interface{} {
	return map[string]interface{}{
		"user_id":             uv.UserID,
		"email":               uv.Email,
		"name":                uv.Name,
		"status":              uv.Status,
		"created_at":          uv.CreatedAt,
		"updated_at":          uv.UpdatedAt,
		"last_login_at":       uv.LastLoginAt,
		"deactivated_at":      uv.DeactivatedAt,
		"deactivation_reason": uv.DeactivationReason,
		"first_name":          uv.FirstName,
		"last_name":           uv.LastName,
		"display_name":        uv.DisplayName,
		"bio":                 uv.Bio,
		"avatar":              uv.Avatar,
		"phone_number":        uv.PhoneNumber,
		"address":             uv.Address,
		"city":                uv.City,
		"country":             uv.Country,
		"postal_code":         uv.PostalCode,
		"roles":               uv.Roles,
		"permissions":         uv.Permissions,
		"searchable_text":     uv.SearchableText,
	}
}

// UpdateSearchableText updates the searchable text field for full-text search
func (uv *UserView) UpdateSearchableText() {
	var parts []string

	// Add basic information
	if uv.Email != "" {
		parts = append(parts, uv.Email)
	}
	if uv.Name != "" {
		parts = append(parts, uv.Name)
	}
	if uv.FirstName != "" {
		parts = append(parts, uv.FirstName)
	}
	if uv.LastName != "" {
		parts = append(parts, uv.LastName)
	}
	if uv.DisplayName != "" {
		parts = append(parts, uv.DisplayName)
	}
	if uv.Bio != "" {
		parts = append(parts, uv.Bio)
	}

	// Add location information
	if uv.City != "" {
		parts = append(parts, uv.City)
	}
	if uv.Country != "" {
		parts = append(parts, uv.Country)
	}

	// Add roles
	for _, role := range uv.Roles {
		parts = append(parts, role)
	}

	// Combine all parts with spaces
	uv.SearchableText = strings.Join(parts, " ")
}

// HasRole checks if the user has a specific role
func (uv *UserView) HasRole(role string) bool {
	for _, r := range uv.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasPermission checks if the user has a specific permission
func (uv *UserView) HasPermission(permission string) bool {
	for _, p := range uv.Permissions {
		if p == permission || p == "*" {
			return true
		}
	}
	return false
}

// GetFullName returns the full name
func (uv *UserView) GetFullName() string {
	if uv.FirstName != "" && uv.LastName != "" {
		return uv.FirstName + " " + uv.LastName
	}
	if uv.FirstName != "" {
		return uv.FirstName
	}
	if uv.LastName != "" {
		return uv.LastName
	}
	return uv.Name
}

// UserViewProjection handles user events and updates the UserView read model
type UserViewProjection struct {
	readStore cqrs.ReadStore
}

// NewUserViewProjection creates a new UserViewProjection
func NewUserViewProjection(readStore cqrs.ReadStore) *UserViewProjection {
	return &UserViewProjection{
		readStore: readStore,
	}
}

// GetProjectionName returns the projection name
func (p *UserViewProjection) GetProjectionName() string {
	return "UserViewProjection"
}

// GetVersion returns the projection version
func (p *UserViewProjection) GetVersion() string {
	return "1.0.0"
}

// GetLastProcessedEvent returns the last processed event ID
func (p *UserViewProjection) GetLastProcessedEvent() string {
	// In a real implementation, this would be persisted
	return ""
}

// CanHandle returns true if the projection can handle the event type
func (p *UserViewProjection) CanHandle(eventType string) bool {
	switch eventType {
	case domain.UserCreatedEventType,
		domain.EmailChangedEventType,
		domain.UserDeactivatedEventType,
		domain.UserActivatedEventType:
		return true
	default:
		return false
	}
}

// Project processes the event and updates the read model
func (p *UserViewProjection) Project(ctx context.Context, event cqrs.EventMessage) error {
	switch e := event.(type) {
	case *domain.UserCreatedEvent:
		return p.handleUserCreated(ctx, e)
	case *domain.EmailChangedEvent:
		return p.handleEmailChanged(ctx, e)
	case *domain.UserDeactivatedEvent:
		return p.handleUserDeactivated(ctx, e)
	case *domain.UserActivatedEvent:
		return p.handleUserActivated(ctx, e)
	default:
		return fmt.Errorf("unsupported event type: %T", event)
	}
}

// GetState returns the current projection state
func (p *UserViewProjection) GetState() cqrs.ProjectionState {
	return cqrs.ProjectionRunning
}

// Reset resets the projection
func (p *UserViewProjection) Reset(ctx context.Context) error {
	// In a real implementation, this would clear all read models
	return nil
}

// Rebuild rebuilds the projection from events
func (p *UserViewProjection) Rebuild(ctx context.Context) error {
	// In a real implementation, this would replay all events
	return nil
}

// Event handlers

// handleUserCreated handles UserCreatedEvent
func (p *UserViewProjection) handleUserCreated(ctx context.Context, event *domain.UserCreatedEvent) error {
	userView := NewUserView(event.UserID)
	userView.Email = event.Email
	userView.Name = event.Name
	userView.Status = "active"
	userView.CreatedAt = event.CreatedAt
	userView.UpdatedAt = event.CreatedAt
	userView.SetVersion(event.Version())

	return p.readStore.Save(ctx, userView)
}

// handleEmailChanged handles EmailChangedEvent
func (p *UserViewProjection) handleEmailChanged(ctx context.Context, event *domain.EmailChangedEvent) error {
	// Load existing user view
	readModel, err := p.readStore.GetByID(ctx, event.UserID, "UserView")
	if err != nil {
		return fmt.Errorf("failed to load user view: %w", err)
	}

	userView, ok := readModel.(*UserView)
	if !ok {
		return fmt.Errorf("invalid read model type: expected *UserView, got %T", readModel)
	}

	// Update email
	userView.Email = event.NewEmail
	userView.UpdatedAt = event.Timestamp()
	userView.SetVersion(event.Version())

	return p.readStore.Save(ctx, userView)
}

// handleUserDeactivated handles UserDeactivatedEvent
func (p *UserViewProjection) handleUserDeactivated(ctx context.Context, event *domain.UserDeactivatedEvent) error {
	// Load existing user view
	readModel, err := p.readStore.GetByID(ctx, event.UserID, "UserView")
	if err != nil {
		return fmt.Errorf("failed to load user view: %w", err)
	}

	userView, ok := readModel.(*UserView)
	if !ok {
		return fmt.Errorf("invalid read model type: expected *UserView, got %T", readModel)
	}

	// Update status
	userView.Status = "deactivated"
	userView.DeactivatedAt = &event.DeactivatedAt
	userView.DeactivationReason = event.Reason
	userView.UpdatedAt = event.Timestamp()
	userView.SetVersion(event.Version())

	return p.readStore.Save(ctx, userView)
}

// handleUserActivated handles UserActivatedEvent
func (p *UserViewProjection) handleUserActivated(ctx context.Context, event *domain.UserActivatedEvent) error {
	// Load existing user view
	readModel, err := p.readStore.GetByID(ctx, event.UserID, "UserView")
	if err != nil {
		return fmt.Errorf("failed to load user view: %w", err)
	}

	userView, ok := readModel.(*UserView)
	if !ok {
		return fmt.Errorf("invalid read model type: expected *UserView, got %T", readModel)
	}

	// Update status
	userView.Status = "active"
	userView.DeactivatedAt = nil
	userView.DeactivationReason = ""
	userView.UpdatedAt = event.Timestamp()
	userView.SetVersion(event.Version())

	return p.readStore.Save(ctx, userView)
}

// UserViewQuery represents a query for user views
type UserViewQuery struct {
	*cqrs.BaseQuery
	TargetUserID string `json:"target_user_id,omitempty"` // Renamed to avoid conflict with UserID() method
	Email        string `json:"email,omitempty"`
	Status       string `json:"status,omitempty"`
}

// NewGetUserQuery creates a query to get a specific user
func NewGetUserQuery(userID string) *UserViewQuery {
	return &UserViewQuery{
		BaseQuery: cqrs.NewBaseQuery(
			"GetUser",
			map[string]interface{}{
				"user_id": userID,
			},
		),
		TargetUserID: userID,
	}
}

// NewListUsersQuery creates a query to list users
func NewListUsersQuery(status string, pagination *cqrs.Pagination) *UserViewQuery {
	criteria := map[string]interface{}{}
	if status != "" {
		criteria["status"] = status
	}

	query := &UserViewQuery{
		BaseQuery: cqrs.NewBaseQuery(
			"ListUsers",
			criteria,
		),
		Status: status,
	}

	if pagination != nil {
		query.SetPagination(pagination)
	}

	return query
}

// Validate validates the query
func (q *UserViewQuery) Validate() error {
	if err := q.BaseQuery.Validate(); err != nil {
		return err
	}

	if q.QueryType() == "GetUser" && q.TargetUserID == "" {
		return fmt.Errorf("user_id is required for GetUser query")
	}

	return nil
}

// UserReadModelFactory creates UserView read models
type UserReadModelFactory struct{}

// NewUserReadModelFactory creates a new UserReadModelFactory
func NewUserReadModelFactory() *UserReadModelFactory {
	return &UserReadModelFactory{}
}

// CreateReadModel creates a read model based on type
func (f *UserReadModelFactory) CreateReadModel(modelType string, id string, data interface{}) (cqrs.ReadModel, error) {
	switch modelType {
	case "UserView":
		return f.createUserView(id, data)
	default:
		return nil, fmt.Errorf("unsupported read model type: %s", modelType)
	}
}

// createUserView creates a UserView from data
func (f *UserReadModelFactory) createUserView(id string, data interface{}) (*UserView, error) {
	// Try to convert data to UserView directly if it's already the right type
	if userView, ok := data.(*UserView); ok {
		return userView, nil
	}

	// Create a new UserView and populate it from data
	userView := NewUserView(id)

	// If data is a map, extract the fields
	if dataMap, ok := data.(map[string]interface{}); ok {
		// Extract fields from the map
		if email, exists := dataMap["email"]; exists {
			if emailStr, ok := email.(string); ok {
				userView.Email = emailStr
			}
		}
		if name, exists := dataMap["name"]; exists {
			if nameStr, ok := name.(string); ok {
				userView.Name = nameStr
			}
		}
		if status, exists := dataMap["status"]; exists {
			if statusStr, ok := status.(string); ok {
				userView.Status = statusStr
			}
		}

		// Handle time fields
		if createdAt, exists := dataMap["created_at"]; exists {
			if createdAtStr, ok := createdAt.(string); ok {
				if parsedTime, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
					userView.CreatedAt = parsedTime
				}
			}
		}
		if updatedAt, exists := dataMap["updated_at"]; exists {
			if updatedAtStr, ok := updatedAt.(string); ok {
				if parsedTime, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
					userView.UpdatedAt = parsedTime
				}
			}
		}

		// Handle optional time fields
		if lastLoginAt, exists := dataMap["last_login_at"]; exists && lastLoginAt != nil {
			if lastLoginAtStr, ok := lastLoginAt.(string); ok {
				if parsedTime, err := time.Parse(time.RFC3339, lastLoginAtStr); err == nil {
					userView.LastLoginAt = &parsedTime
				}
			}
		}
		if deactivatedAt, exists := dataMap["deactivated_at"]; exists && deactivatedAt != nil {
			if deactivatedAtStr, ok := deactivatedAt.(string); ok {
				if parsedTime, err := time.Parse(time.RFC3339, deactivatedAtStr); err == nil {
					userView.DeactivatedAt = &parsedTime
				}
			}
		}
		if deactivationReason, exists := dataMap["deactivation_reason"]; exists {
			if reasonStr, ok := deactivationReason.(string); ok {
				userView.DeactivationReason = reasonStr
			}
		}
	}
	return userView, nil
}

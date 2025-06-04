package queries

import (
	"context"
	"fmt"
	"strings"

	"defense-allies-server/examples/guild/infrastructure/projections"
	"defense-allies-server/pkg/cqrs"
)

// Query type constants
const (
	GetGuildQueryType        = "GetGuild"
	GetGuildMembersQueryType = "GetGuildMembers"
	SearchGuildsQueryType    = "SearchGuilds"
	GetGuildRankingQueryType = "GetGuildRanking"
)

// GetGuildQuery represents a query to get a specific guild
type GetGuildQuery struct {
	*cqrs.BaseQuery
	GuildID string `json:"guild_id"`
}

// NewGetGuildQuery creates a new GetGuildQuery
func NewGetGuildQuery(guildID string) *GetGuildQuery {
	return &GetGuildQuery{
		BaseQuery: cqrs.NewBaseQuery(
			GetGuildQueryType,
			map[string]interface{}{
				"guild_id": guildID,
			},
		),
		GuildID: guildID,
	}
}

// Validate validates the get guild query
func (q *GetGuildQuery) Validate() error {
	if q.GuildID == "" {
		return fmt.Errorf("guild ID cannot be empty")
	}
	return nil
}

// GetGuildMembersQuery represents a query to get guild members
type GetGuildMembersQuery struct {
	*cqrs.BaseQuery
	GuildID   string `json:"guild_id"`
	Status    string `json:"status,omitempty"`     // Filter by status (Active, Pending, Kicked)
	Role      string `json:"role,omitempty"`       // Filter by role
	Limit     int    `json:"limit,omitempty"`      // Limit number of results
	Offset    int    `json:"offset,omitempty"`     // Offset for pagination
	SortBy    string `json:"sort_by,omitempty"`    // Sort field (joined_at, contribution, role)
	SortOrder string `json:"sort_order,omitempty"` // Sort order (asc, desc)
}

// NewGetGuildMembersQuery creates a new GetGuildMembersQuery
func NewGetGuildMembersQuery(guildID string) *GetGuildMembersQuery {
	return &GetGuildMembersQuery{
		BaseQuery: cqrs.NewBaseQuery(
			GetGuildMembersQueryType,
			map[string]interface{}{
				"guild_id": guildID,
			},
		),
		GuildID:   guildID,
		Limit:     50, // Default limit
		Offset:    0,  // Default offset
		SortBy:    "joined_at",
		SortOrder: "asc",
	}
}

// WithStatus adds status filter
func (q *GetGuildMembersQuery) WithStatus(status string) *GetGuildMembersQuery {
	q.Status = status
	return q
}

// WithRole adds role filter
func (q *GetGuildMembersQuery) WithRole(role string) *GetGuildMembersQuery {
	q.Role = role
	return q
}

// WithPagination adds pagination
func (q *GetGuildMembersQuery) WithPagination(limit, offset int) *GetGuildMembersQuery {
	q.Limit = limit
	q.Offset = offset
	return q
}

// WithSorting adds sorting
func (q *GetGuildMembersQuery) WithSorting(sortBy, sortOrder string) *GetGuildMembersQuery {
	q.SortBy = sortBy
	q.SortOrder = sortOrder
	return q
}

// Validate validates the get guild members query
func (q *GetGuildMembersQuery) Validate() error {
	if q.GuildID == "" {
		return fmt.Errorf("guild ID cannot be empty")
	}
	if q.Limit < 0 || q.Limit > 1000 {
		return fmt.Errorf("limit must be between 0 and 1000")
	}
	if q.Offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}
	return nil
}

// SearchGuildsQuery represents a query to search guilds
type SearchGuildsQuery struct {
	*cqrs.BaseQuery
	SearchText string `json:"search_text,omitempty"` // Search in name, tag, description
	Status     string `json:"status,omitempty"`      // Filter by status
	IsPublic   *bool  `json:"is_public,omitempty"`   // Filter by public/private
	MinLevel   int    `json:"min_level,omitempty"`   // Filter by minimum level
	MaxLevel   int    `json:"max_level,omitempty"`   // Filter by maximum level
	Limit      int    `json:"limit,omitempty"`       // Limit number of results
	Offset     int    `json:"offset,omitempty"`      // Offset for pagination
	SortBy     string `json:"sort_by,omitempty"`     // Sort field (name, level, member_count, founded_at)
	SortOrder  string `json:"sort_order,omitempty"`  // Sort order (asc, desc)
}

// NewSearchGuildsQuery creates a new SearchGuildsQuery
func NewSearchGuildsQuery() *SearchGuildsQuery {
	return &SearchGuildsQuery{
		BaseQuery: cqrs.NewBaseQuery(
			SearchGuildsQueryType,
			map[string]interface{}{},
		),
		Limit:     20, // Default limit
		Offset:    0,  // Default offset
		SortBy:    "name",
		SortOrder: "asc",
	}
}

// WithSearchText adds search text
func (q *SearchGuildsQuery) WithSearchText(searchText string) *SearchGuildsQuery {
	q.SearchText = searchText
	return q
}

// WithStatus adds status filter
func (q *SearchGuildsQuery) WithStatus(status string) *SearchGuildsQuery {
	q.Status = status
	return q
}

// WithPublicFilter adds public/private filter
func (q *SearchGuildsQuery) WithPublicFilter(isPublic bool) *SearchGuildsQuery {
	q.IsPublic = &isPublic
	return q
}

// WithLevelRange adds level range filter
func (q *SearchGuildsQuery) WithLevelRange(minLevel, maxLevel int) *SearchGuildsQuery {
	q.MinLevel = minLevel
	q.MaxLevel = maxLevel
	return q
}

// WithPagination adds pagination
func (q *SearchGuildsQuery) WithPagination(limit, offset int) *SearchGuildsQuery {
	q.Limit = limit
	q.Offset = offset
	return q
}

// WithSorting adds sorting
func (q *SearchGuildsQuery) WithSorting(sortBy, sortOrder string) *SearchGuildsQuery {
	q.SortBy = sortBy
	q.SortOrder = sortOrder
	return q
}

// Validate validates the search guilds query
func (q *SearchGuildsQuery) Validate() error {
	if q.Limit < 0 || q.Limit > 1000 {
		return fmt.Errorf("limit must be between 0 and 1000")
	}
	if q.Offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}
	if q.MinLevel < 0 {
		return fmt.Errorf("min level cannot be negative")
	}
	if q.MaxLevel > 0 && q.MaxLevel < q.MinLevel {
		return fmt.Errorf("max level cannot be less than min level")
	}
	return nil
}

// GuildQueryResult represents the result of a guild query
type GuildQueryResult struct {
	Guild   *projections.GuildView    `json:"guild,omitempty"`
	Guilds  []*projections.GuildView  `json:"guilds,omitempty"`
	Members []*projections.MemberView `json:"members,omitempty"`
	Total   int                       `json:"total,omitempty"`
	Limit   int                       `json:"limit,omitempty"`
	Offset  int                       `json:"offset,omitempty"`
}

// GuildQueryHandler handles guild-related queries
type GuildQueryHandler struct {
	*cqrs.BaseQueryHandler
	readStore cqrs.ReadStore
}

// NewGuildQueryHandler creates a new GuildQueryHandler
func NewGuildQueryHandler(readStore cqrs.ReadStore) *GuildQueryHandler {
	supportedQueries := []string{
		GetGuildQueryType,
		GetGuildMembersQueryType,
		SearchGuildsQueryType,
	}

	return &GuildQueryHandler{
		BaseQueryHandler: cqrs.NewBaseQueryHandler("GuildQueryHandler", supportedQueries),
		readStore:        readStore,
	}
}

// Handle handles the incoming query
func (h *GuildQueryHandler) Handle(ctx context.Context, query cqrs.Query) (*cqrs.QueryResult, error) {
	// Validate query
	if err := query.Validate(); err != nil {
		return &cqrs.QueryResult{
			Success: false,
			Error:   fmt.Errorf("query validation failed: %w", err),
		}, nil
	}

	var result interface{}
	var err error

	switch q := query.(type) {
	case *GetGuildQuery:
		result, err = h.handleGetGuild(ctx, q)
	case *GetGuildMembersQuery:
		result, err = h.handleGetGuildMembers(ctx, q)
	case *SearchGuildsQuery:
		result, err = h.handleSearchGuilds(ctx, q)
	default:
		return &cqrs.QueryResult{
			Success: false,
			Error:   fmt.Errorf("unsupported query type: %T", query),
		}, nil
	}

	if err != nil {
		return &cqrs.QueryResult{
			Success: false,
			Error:   err,
		}, nil
	}

	return &cqrs.QueryResult{
		Success: true,
		Data:    result,
	}, nil
}

// handleGetGuild handles GetGuildQuery
func (h *GuildQueryHandler) handleGetGuild(ctx context.Context, query *GetGuildQuery) (*GuildQueryResult, error) {
	// Load guild view
	readModel, err := h.readStore.GetByID(ctx, query.GuildID, "GuildView")
	if err != nil {
		return nil, fmt.Errorf("failed to load guild view: %w", err)
	}

	guildView, ok := readModel.(*projections.GuildView)
	if !ok {
		return nil, fmt.Errorf("invalid read model type: expected *GuildView, got %T", readModel)
	}

	return &GuildQueryResult{
		Guild: guildView,
	}, nil
}

// handleGetGuildMembers handles GetGuildMembersQuery
func (h *GuildQueryHandler) handleGetGuildMembers(ctx context.Context, query *GetGuildMembersQuery) (*GuildQueryResult, error) {
	// Get all member views for this guild
	allMembers, err := h.getAllMembersForGuild(ctx, query.GuildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guild members: %w", err)
	}

	// Apply filters
	filteredMembers := h.filterMembers(allMembers, query)

	// Apply sorting
	sortedMembers := h.sortMembers(filteredMembers, query.SortBy, query.SortOrder)

	// Apply pagination
	total := len(sortedMembers)
	start := query.Offset
	end := start + query.Limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedMembers := make([]*projections.MemberView, 0)
	if start < end {
		paginatedMembers = sortedMembers[start:end]
	}

	return &GuildQueryResult{
		Members: paginatedMembers,
		Total:   total,
		Limit:   query.Limit,
		Offset:  query.Offset,
	}, nil
}

// handleSearchGuilds handles SearchGuildsQuery
func (h *GuildQueryHandler) handleSearchGuilds(ctx context.Context, query *SearchGuildsQuery) (*GuildQueryResult, error) {
	// Get all guild views
	allGuilds, err := h.getAllGuilds(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all guilds: %w", err)
	}

	// Apply filters
	filteredGuilds := h.filterGuilds(allGuilds, query)

	// Apply sorting
	sortedGuilds := h.sortGuilds(filteredGuilds, query.SortBy, query.SortOrder)

	// Apply pagination
	total := len(sortedGuilds)
	start := query.Offset
	end := start + query.Limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedGuilds := make([]*projections.GuildView, 0)
	if start < end {
		// Create a slice to hold guild views instead of the result struct
		for _, guild := range sortedGuilds[start:end] {
			paginatedGuilds = append(paginatedGuilds, guild)
		}
	}

	return &GuildQueryResult{
		Guilds: paginatedGuilds,
		Total:  total,
		Limit:  query.Limit,
		Offset: query.Offset,
	}, nil
}

// Helper methods for data retrieval and filtering

// getAllMembersForGuild retrieves all member views for a specific guild
func (h *GuildQueryHandler) getAllMembersForGuild(ctx context.Context, guildID string) ([]*projections.MemberView, error) {
	// In a real implementation, this would use a proper query to get members by guild ID
	// For this example, we'll simulate by checking if we can get any member views
	// Since we don't have a proper query mechanism in the in-memory store, we'll return empty for now
	// In production, you'd implement proper indexing and querying

	members := make([]*projections.MemberView, 0)

	// Try to get some sample member IDs (this is a simplified approach)
	// In a real system, you'd have proper indexing by guild ID
	sampleUserIDs := []string{"founder123", "member001", "member002"}

	for _, userID := range sampleUserIDs {
		memberID := fmt.Sprintf("%s:%s", guildID, userID)
		readModel, err := h.readStore.GetByID(ctx, memberID, "MemberView")
		if err != nil {
			// Member doesn't exist, continue
			continue
		}

		if memberView, ok := readModel.(*projections.MemberView); ok {
			members = append(members, memberView)
		}
	}

	return members, nil
}

// getAllGuilds retrieves all guild views
func (h *GuildQueryHandler) getAllGuilds(ctx context.Context) ([]*projections.GuildView, error) {
	// In a real implementation, this would use a proper query to get all guilds
	// For this example, we'll return empty since we don't have a way to list all items in the in-memory store
	// In production, you'd implement proper indexing and querying

	guilds := make([]*projections.GuildView, 0)

	// This is a limitation of our simple in-memory implementation
	// In a real system, you'd have proper indexing and be able to list all guilds

	return guilds, nil
}

// filterMembers applies filters to member list
func (h *GuildQueryHandler) filterMembers(members []*projections.MemberView, query *GetGuildMembersQuery) []*projections.MemberView {
	filtered := make([]*projections.MemberView, 0)

	for _, member := range members {
		// Apply status filter
		if query.Status != "" && member.Status != query.Status {
			continue
		}

		// Apply role filter
		if query.Role != "" && member.Role != query.Role {
			continue
		}

		filtered = append(filtered, member)
	}

	return filtered
}

// sortMembers sorts members based on the specified criteria
func (h *GuildQueryHandler) sortMembers(members []*projections.MemberView, sortBy, sortOrder string) []*projections.MemberView {
	if len(members) == 0 {
		return members
	}

	// Create a copy to avoid modifying the original slice
	sorted := make([]*projections.MemberView, len(members))
	copy(sorted, members)

	// Simple sorting implementation (in production, you'd use a more sophisticated approach)
	// For this example, we'll implement basic sorting by joined_at
	if sortBy == "joined_at" {
		// Sort by joined date (newest first if desc, oldest first if asc)
		for i := 0; i < len(sorted)-1; i++ {
			for j := i + 1; j < len(sorted); j++ {
				var shouldSwap bool
				if sortOrder == "desc" {
					shouldSwap = sorted[i].JoinedAt.Before(sorted[j].JoinedAt)
				} else {
					shouldSwap = sorted[i].JoinedAt.After(sorted[j].JoinedAt)
				}

				if shouldSwap {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
	}

	return sorted
}

// filterGuilds applies filters to guild list
func (h *GuildQueryHandler) filterGuilds(guilds []*projections.GuildView, query *SearchGuildsQuery) []*projections.GuildView {
	filtered := make([]*projections.GuildView, 0)

	for _, guild := range guilds {
		// Apply search text filter
		if query.SearchText != "" {
			searchText := strings.ToLower(query.SearchText)
			searchableText := strings.ToLower(guild.SearchableText)
			if !strings.Contains(searchableText, searchText) {
				continue
			}
		}

		// Apply status filter
		if query.Status != "" && guild.Status != query.Status {
			continue
		}

		// Apply public/private filter
		if query.IsPublic != nil && guild.IsPublic != *query.IsPublic {
			continue
		}

		// Apply level range filter
		if query.MinLevel > 0 && guild.Level < query.MinLevel {
			continue
		}
		if query.MaxLevel > 0 && guild.Level > query.MaxLevel {
			continue
		}

		filtered = append(filtered, guild)
	}

	return filtered
}

// sortGuilds sorts guilds based on the specified criteria
func (h *GuildQueryHandler) sortGuilds(guilds []*projections.GuildView, sortBy, sortOrder string) []*projections.GuildView {
	if len(guilds) == 0 {
		return guilds
	}

	// Create a copy to avoid modifying the original slice
	sorted := make([]*projections.GuildView, len(guilds))
	copy(sorted, guilds)

	// Simple sorting implementation
	switch sortBy {
	case "name":
		for i := 0; i < len(sorted)-1; i++ {
			for j := i + 1; j < len(sorted); j++ {
				var shouldSwap bool
				if sortOrder == "desc" {
					shouldSwap = strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
				} else {
					shouldSwap = strings.ToLower(sorted[i].Name) > strings.ToLower(sorted[j].Name)
				}

				if shouldSwap {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
	case "level":
		for i := 0; i < len(sorted)-1; i++ {
			for j := i + 1; j < len(sorted); j++ {
				var shouldSwap bool
				if sortOrder == "desc" {
					shouldSwap = sorted[i].Level < sorted[j].Level
				} else {
					shouldSwap = sorted[i].Level > sorted[j].Level
				}

				if shouldSwap {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
	case "member_count":
		for i := 0; i < len(sorted)-1; i++ {
			for j := i + 1; j < len(sorted); j++ {
				var shouldSwap bool
				if sortOrder == "desc" {
					shouldSwap = sorted[i].ActiveMemberCount < sorted[j].ActiveMemberCount
				} else {
					shouldSwap = sorted[i].ActiveMemberCount > sorted[j].ActiveMemberCount
				}

				if shouldSwap {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
	case "founded_at":
		for i := 0; i < len(sorted)-1; i++ {
			for j := i + 1; j < len(sorted); j++ {
				var shouldSwap bool
				if sortOrder == "desc" {
					shouldSwap = sorted[i].FoundedAt.Before(sorted[j].FoundedAt)
				} else {
					shouldSwap = sorted[i].FoundedAt.After(sorted[j].FoundedAt)
				}

				if shouldSwap {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
	}

	return sorted
}

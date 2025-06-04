package domain

import (
	"defense-allies-server/pkg/cqrs"
)

// Event type constants
const (
	// Guild events
	GuildCreatedEventType         = "GuildCreated"
	GuildInfoUpdatedEventType     = "GuildInfoUpdated"
	GuildSettingsUpdatedEventType = "GuildSettingsUpdated"
	GuildDisbandedEventType       = "GuildDisbanded"

	// Member events
	MemberInvitedEventType  = "MemberInvited"
	MemberJoinedEventType   = "MemberJoined"
	MemberLeftEventType     = "MemberLeft"
	MemberKickedEventType   = "MemberKicked"
	MemberPromotedEventType = "MemberPromoted"
	MemberDemotedEventType  = "MemberDemoted"

	// Mining events
	MineDiscoveredEventType         = "MineDiscovered"
	MiningStartedEventType          = "MiningStarted"
	WorkerAssignedEventType         = "WorkerAssigned"
	WorkerRemovedEventType          = "WorkerRemoved"
	MineralsExtractedEventType      = "MineralsExtracted"
	MiningOperationStartedEventType = "MiningOperationStarted"
	MineralsHarvestedEventType      = "MineralsHarvested"
	MiningOperationStoppedEventType = "MiningOperationStopped"

	// Transport Recruitment events
	TransportRecruitmentCreatedEventType   = "TransportRecruitmentCreated"
	TransportRecruitmentJoinedEventType    = "TransportRecruitmentJoined"
	TransportRecruitmentLeftEventType      = "TransportRecruitmentLeft"
	TransportRecruitmentStartedEventType   = "TransportRecruitmentStarted"
	TransportRecruitmentCompletedEventType = "TransportRecruitmentCompleted"

	// Transport events
	TransportStartedEventType   = "TransportStarted"
	TransportAttackedEventType  = "TransportAttacked"
	TransportDefendedEventType  = "TransportDefended"
	TransportCompletedEventType = "TransportCompleted"
	TransportRaidedEventType    = "TransportRaided"
	TransportCancelledEventType = "TransportCancelled"
)

// Guild Events

// GuildCreatedEvent represents a guild creation event
type GuildCreatedEvent struct {
	*cqrs.BaseEventMessage
	GuildID         string `json:"guild_id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	FounderID       string `json:"founder_id"`
	FounderUsername string `json:"founder_username"`
}

// NewGuildCreatedEvent creates a new guild created event
func NewGuildCreatedEvent(guildID, name, description, founderID, founderUsername string) *GuildCreatedEvent {
	return &GuildCreatedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			GuildCreatedEventType,
			guildID,
			"Guild",
			1, // version
			map[string]interface{}{
				"guild_id":         guildID,
				"name":             name,
				"description":      description,
				"founder_id":       founderID,
				"founder_username": founderUsername,
			},
		),
		GuildID:         guildID,
		Name:            name,
		Description:     description,
		FounderID:       founderID,
		FounderUsername: founderUsername,
	}
}

// GuildInfoUpdatedEvent represents a guild info update event
type GuildInfoUpdatedEvent struct {
	*cqrs.BaseEventMessage
	GuildID     string `json:"guild_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Notice      string `json:"notice"`
	Tag         string `json:"tag"`
	UpdatedBy   string `json:"updated_by"`
}

// NewGuildInfoUpdatedEvent creates a new guild info updated event
func NewGuildInfoUpdatedEvent(guildID, name, description, notice, tag, updatedBy string) *GuildInfoUpdatedEvent {
	return &GuildInfoUpdatedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			GuildInfoUpdatedEventType,
			guildID,
			"Guild",
			1, // version
			map[string]interface{}{
				"guild_id":    guildID,
				"name":        name,
				"description": description,
				"notice":      notice,
				"tag":         tag,
				"updated_by":  updatedBy,
			},
		),
		GuildID:     guildID,
		Name:        name,
		Description: description,
		Notice:      notice,
		Tag:         tag,
		UpdatedBy:   updatedBy,
	}
}

// GuildSettingsUpdatedEvent represents a guild settings update event
type GuildSettingsUpdatedEvent struct {
	*cqrs.BaseEventMessage
	GuildID         string `json:"guild_id"`
	MaxMembers      int    `json:"max_members"`
	MinLevel        int    `json:"min_level"`
	IsPublic        bool   `json:"is_public"`
	RequireApproval bool   `json:"require_approval"`
	UpdatedBy       string `json:"updated_by"`
}

// NewGuildSettingsUpdatedEvent creates a new guild settings updated event
func NewGuildSettingsUpdatedEvent(guildID string, maxMembers, minLevel int, isPublic, requireApproval bool, updatedBy string) *GuildSettingsUpdatedEvent {
	return &GuildSettingsUpdatedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			GuildSettingsUpdatedEventType,
			guildID,
			"Guild",
			1, // version
			map[string]interface{}{
				"guild_id":         guildID,
				"max_members":      maxMembers,
				"min_level":        minLevel,
				"is_public":        isPublic,
				"require_approval": requireApproval,
				"updated_by":       updatedBy,
			},
		),
		GuildID:         guildID,
		MaxMembers:      maxMembers,
		MinLevel:        minLevel,
		IsPublic:        isPublic,
		RequireApproval: requireApproval,
		UpdatedBy:       updatedBy,
	}
}

// Member Events

// MemberInvitedEvent represents a member invitation event
type MemberInvitedEvent struct {
	*cqrs.BaseEventMessage
	GuildID   string `json:"guild_id"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	InvitedBy string `json:"invited_by"`
}

// NewMemberInvitedEvent creates a new member invited event
func NewMemberInvitedEvent(guildID, userID, username, invitedBy string) *MemberInvitedEvent {
	return &MemberInvitedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			MemberInvitedEventType,
			guildID,
			"Guild",
			1, // version
			map[string]interface{}{
				"guild_id":   guildID,
				"user_id":    userID,
				"username":   username,
				"invited_by": invitedBy,
			},
		),
		GuildID:   guildID,
		UserID:    userID,
		Username:  username,
		InvitedBy: invitedBy,
	}
}

// MemberJoinedEvent represents a member joining event
type MemberJoinedEvent struct {
	*cqrs.BaseEventMessage
	GuildID string `json:"guild_id"`
	UserID  string `json:"user_id"`
}

// NewMemberJoinedEvent creates a new member joined event
func NewMemberJoinedEvent(guildID, userID string) *MemberJoinedEvent {
	return &MemberJoinedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			MemberJoinedEventType,
			guildID,
			"Guild",
			1, // version
			map[string]interface{}{
				"guild_id": guildID,
				"user_id":  userID,
			},
		),
		GuildID: guildID,
		UserID:  userID,
	}
}

// MemberKickedEvent represents a member being kicked event
type MemberKickedEvent struct {
	*cqrs.BaseEventMessage
	GuildID  string `json:"guild_id"`
	UserID   string `json:"user_id"`
	KickedBy string `json:"kicked_by"`
	Reason   string `json:"reason"`
}

// NewMemberKickedEvent creates a new member kicked event
func NewMemberKickedEvent(guildID, userID, kickedBy, reason string) *MemberKickedEvent {
	return &MemberKickedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			MemberKickedEventType,
			guildID,
			"Guild",
			1, // version
			map[string]interface{}{
				"guild_id":  guildID,
				"user_id":   userID,
				"kicked_by": kickedBy,
				"reason":    reason,
			},
		),
		GuildID:  guildID,
		UserID:   userID,
		KickedBy: kickedBy,
		Reason:   reason,
	}
}

// MemberPromotedEvent represents a member promotion event
type MemberPromotedEvent struct {
	*cqrs.BaseEventMessage
	GuildID    string    `json:"guild_id"`
	UserID     string    `json:"user_id"`
	PromotedBy string    `json:"promoted_by"`
	OldRole    GuildRole `json:"old_role"`
	NewRole    GuildRole `json:"new_role"`
}

// NewMemberPromotedEvent creates a new member promoted event
func NewMemberPromotedEvent(guildID, userID, promotedBy string, oldRole, newRole GuildRole) *MemberPromotedEvent {
	return &MemberPromotedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			MemberPromotedEventType,
			guildID,
			"Guild",
			1, // version
			map[string]interface{}{
				"guild_id":    guildID,
				"user_id":     userID,
				"promoted_by": promotedBy,
				"old_role":    oldRole.String(),
				"new_role":    newRole.String(),
			},
		),
		GuildID:    guildID,
		UserID:     userID,
		PromotedBy: promotedBy,
		OldRole:    oldRole,
		NewRole:    newRole,
	}
}

// Mining Events

// MiningOperationStartedEvent represents a mining operation start event
type MiningOperationStartedEvent struct {
	*cqrs.BaseEventMessage
	GuildID     string   `json:"guild_id"`
	OperationID string   `json:"operation_id"`
	NodeID      string   `json:"node_id"`
	WorkerIDs   []string `json:"worker_ids"`
	StartedBy   string   `json:"started_by"`
}

// NewMiningOperationStartedEvent creates a new mining operation started event
func NewMiningOperationStartedEvent(guildID, operationID, nodeID string, workerIDs []string, startedBy string) *MiningOperationStartedEvent {
	return &MiningOperationStartedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			MiningOperationStartedEventType,
			guildID,
			"Guild",
			1, // version
			map[string]interface{}{
				"guild_id":     guildID,
				"operation_id": operationID,
				"node_id":      nodeID,
				"worker_ids":   workerIDs,
				"started_by":   startedBy,
			},
		),
		GuildID:     guildID,
		OperationID: operationID,
		NodeID:      nodeID,
		WorkerIDs:   workerIDs,
		StartedBy:   startedBy,
	}
}

// MineralsHarvestedEvent represents a minerals harvest event
type MineralsHarvestedEvent struct {
	*cqrs.BaseEventMessage
	GuildID          string                `json:"guild_id"`
	OperationID      string                `json:"operation_id"`
	Harvested        map[MineralType]int64 `json:"harvested"`
	TreasuryIncrease int64                 `json:"treasury_increase"`
	HarvestedBy      string                `json:"harvested_by"`
}

// NewMineralsHarvestedEvent creates a new minerals harvested event
func NewMineralsHarvestedEvent(guildID, operationID string, harvested map[MineralType]int64, treasuryIncrease int64, harvestedBy string) *MineralsHarvestedEvent {
	// Convert map for serialization
	harvestedData := make(map[string]interface{})
	for mineralType, amount := range harvested {
		harvestedData[mineralType.String()] = amount
	}

	return &MineralsHarvestedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			MineralsHarvestedEventType,
			guildID,
			"Guild",
			1, // version
			map[string]interface{}{
				"guild_id":          guildID,
				"operation_id":      operationID,
				"harvested":         harvestedData,
				"treasury_increase": treasuryIncrease,
				"harvested_by":      harvestedBy,
			},
		),
		GuildID:          guildID,
		OperationID:      operationID,
		Harvested:        harvested,
		TreasuryIncrease: treasuryIncrease,
		HarvestedBy:      harvestedBy,
	}
}

// MiningOperationStoppedEvent represents a mining operation stop event
type MiningOperationStoppedEvent struct {
	*cqrs.BaseEventMessage
	GuildID     string `json:"guild_id"`
	OperationID string `json:"operation_id"`
	StoppedBy   string `json:"stopped_by"`
}

// NewMiningOperationStoppedEvent creates a new mining operation stopped event
func NewMiningOperationStoppedEvent(guildID, operationID, stoppedBy string) *MiningOperationStoppedEvent {
	return &MiningOperationStoppedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			MiningOperationStoppedEventType,
			guildID,
			"Guild",
			1, // version
			map[string]interface{}{
				"guild_id":     guildID,
				"operation_id": operationID,
				"stopped_by":   stoppedBy,
			},
		),
		GuildID:     guildID,
		OperationID: operationID,
		StoppedBy:   stoppedBy,
	}
}

// Transport Recruitment Events

// TransportRecruitmentCreatedEvent represents a transport recruitment creation event
type TransportRecruitmentCreatedEvent struct {
	*cqrs.BaseEventMessage
	GuildID           string                `json:"guild_id"`
	RecruitmentID     string                `json:"recruitment_id"`
	Title             string                `json:"title"`
	Description       string                `json:"description"`
	MaxParticipants   int                   `json:"max_participants"`
	MinParticipants   int                   `json:"min_participants"`
	Duration          int64                 `json:"duration"`       // Duration in nanoseconds
	TransportTime     int64                 `json:"transport_time"` // Transport time in nanoseconds
	TotalCargo        map[MineralType]int64 `json:"total_cargo"`
	CreatedBy         string                `json:"created_by"`
	CreatedByUsername string                `json:"created_by_username"`
}

// NewTransportRecruitmentCreatedEvent creates a new transport recruitment created event
func NewTransportRecruitmentCreatedEvent(guildID, recruitmentID, title, description string,
	maxParticipants, minParticipants int, duration, transportTime int64,
	totalCargo map[MineralType]int64, createdBy, createdByUsername string) *TransportRecruitmentCreatedEvent {

	// Convert map for serialization
	cargoData := make(map[string]interface{})
	for mineralType, amount := range totalCargo {
		cargoData[mineralType.String()] = amount
	}

	return &TransportRecruitmentCreatedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			TransportRecruitmentCreatedEventType,
			guildID,
			"Guild",
			1, // version
			map[string]interface{}{
				"guild_id":            guildID,
				"recruitment_id":      recruitmentID,
				"title":               title,
				"description":         description,
				"max_participants":    maxParticipants,
				"min_participants":    minParticipants,
				"duration":            duration,
				"transport_time":      transportTime,
				"total_cargo":         cargoData,
				"created_by":          createdBy,
				"created_by_username": createdByUsername,
			},
		),
		GuildID:           guildID,
		RecruitmentID:     recruitmentID,
		Title:             title,
		Description:       description,
		MaxParticipants:   maxParticipants,
		MinParticipants:   minParticipants,
		Duration:          duration,
		TransportTime:     transportTime,
		TotalCargo:        totalCargo,
		CreatedBy:         createdBy,
		CreatedByUsername: createdByUsername,
	}
}

// TransportRecruitmentJoinedEvent represents a transport recruitment join event
type TransportRecruitmentJoinedEvent struct {
	*cqrs.BaseEventMessage
	GuildID       string `json:"guild_id"`
	RecruitmentID string `json:"recruitment_id"`
	UserID        string `json:"user_id"`
	Username      string `json:"username"`
}

// NewTransportRecruitmentJoinedEvent creates a new transport recruitment joined event
func NewTransportRecruitmentJoinedEvent(guildID, recruitmentID, userID, username string) *TransportRecruitmentJoinedEvent {
	return &TransportRecruitmentJoinedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			TransportRecruitmentJoinedEventType,
			guildID,
			"Guild",
			1, // version
			map[string]interface{}{
				"guild_id":       guildID,
				"recruitment_id": recruitmentID,
				"user_id":        userID,
				"username":       username,
			},
		),
		GuildID:       guildID,
		RecruitmentID: recruitmentID,
		UserID:        userID,
		Username:      username,
	}
}

// TransportRecruitmentLeftEvent represents a transport recruitment leave event
type TransportRecruitmentLeftEvent struct {
	*cqrs.BaseEventMessage
	GuildID       string `json:"guild_id"`
	RecruitmentID string `json:"recruitment_id"`
	UserID        string `json:"user_id"`
	Username      string `json:"username"`
}

// NewTransportRecruitmentLeftEvent creates a new transport recruitment left event
func NewTransportRecruitmentLeftEvent(guildID, recruitmentID, userID, username string) *TransportRecruitmentLeftEvent {
	return &TransportRecruitmentLeftEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			TransportRecruitmentLeftEventType,
			guildID,
			"Guild",
			1, // version
			map[string]interface{}{
				"guild_id":       guildID,
				"recruitment_id": recruitmentID,
				"user_id":        userID,
				"username":       username,
			},
		),
		GuildID:       guildID,
		RecruitmentID: recruitmentID,
		UserID:        userID,
		Username:      username,
	}
}

// TransportRecruitmentStartedEvent represents a transport recruitment start event
type TransportRecruitmentStartedEvent struct {
	*cqrs.BaseEventMessage
	GuildID       string `json:"guild_id"`
	RecruitmentID string `json:"recruitment_id"`
	TransportID   string `json:"transport_id"`
	StartedBy     string `json:"started_by"`
}

// NewTransportRecruitmentStartedEvent creates a new transport recruitment started event
func NewTransportRecruitmentStartedEvent(guildID, recruitmentID, transportID, startedBy string) *TransportRecruitmentStartedEvent {
	return &TransportRecruitmentStartedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			TransportRecruitmentStartedEventType,
			guildID,
			"Guild",
			1, // version
			map[string]interface{}{
				"guild_id":       guildID,
				"recruitment_id": recruitmentID,
				"transport_id":   transportID,
				"started_by":     startedBy,
			},
		),
		GuildID:       guildID,
		RecruitmentID: recruitmentID,
		TransportID:   transportID,
		StartedBy:     startedBy,
	}
}

// TransportRecruitmentCompletedEvent represents a transport recruitment completion event
type TransportRecruitmentCompletedEvent struct {
	*cqrs.BaseEventMessage
	GuildID       string                           `json:"guild_id"`
	RecruitmentID string                           `json:"recruitment_id"`
	Rewards       map[string]map[MineralType]int64 `json:"rewards"` // userID -> rewards
	CompletedBy   string                           `json:"completed_by"`
}

// NewTransportRecruitmentCompletedEvent creates a new transport recruitment completed event
func NewTransportRecruitmentCompletedEvent(guildID, recruitmentID string, rewards map[string]map[MineralType]int64, completedBy string) *TransportRecruitmentCompletedEvent {
	// Convert rewards for serialization
	rewardsData := make(map[string]interface{})
	for userID, userRewards := range rewards {
		userRewardsData := make(map[string]interface{})
		for mineralType, amount := range userRewards {
			userRewardsData[mineralType.String()] = amount
		}
		rewardsData[userID] = userRewardsData
	}

	return &TransportRecruitmentCompletedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			TransportRecruitmentCompletedEventType,
			guildID,
			"Guild",
			1, // version
			map[string]interface{}{
				"guild_id":       guildID,
				"recruitment_id": recruitmentID,
				"rewards":        rewardsData,
				"completed_by":   completedBy,
			},
		),
		GuildID:       guildID,
		RecruitmentID: recruitmentID,
		Rewards:       rewards,
		CompletedBy:   completedBy,
	}
}

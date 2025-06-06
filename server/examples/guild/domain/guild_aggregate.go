package domain

import (
	"fmt"
	"time"

	"defense-allies-server/pkg/cqrs"
)

// GuildStatus represents the status of a guild
type GuildStatus int

const (
	// GuildStatusActive represents an active guild
	GuildStatusActive GuildStatus = iota
	// GuildStatusInactive represents an inactive guild
	GuildStatusInactive
	// GuildStatusDisbanded represents a disbanded guild
	GuildStatusDisbanded
)

// String returns the string representation of the guild status
func (s GuildStatus) String() string {
	switch s {
	case GuildStatusActive:
		return "Active"
	case GuildStatusInactive:
		return "Inactive"
	case GuildStatusDisbanded:
		return "Disbanded"
	default:
		return "Unknown"
	}
}

// GuildAggregate represents the guild aggregate root
type GuildAggregate struct {
	*cqrs.BaseAggregate

	// Guild basic information
	name        string
	description string
	notice      string
	tag         string // Guild tag/abbreviation
	status      GuildStatus

	// Guild settings
	maxMembers      int
	isPublic        bool
	requireApproval bool
	minLevel        int

	// Guild members
	members map[string]*GuildMember // userID -> member

	// Guild resources
	treasury              int64                            // Guild treasury amount
	mines                 map[string]*Mine                 // mineID -> mine
	transports            map[string]*Transport            // transportID -> transport
	transportRecruitments map[string]*TransportRecruitment // recruitmentID -> recruitment

	// Guild statistics
	totalContribution int64
	level             int
	experience        int64
	ranking           int

	// Mining system
	mining *GuildMining

	// Timestamps
	foundedAt    time.Time
	lastActiveAt time.Time
}

// NewGuildAggregate creates a new guild aggregate
func NewGuildAggregate(id, name, description, founderID, founderUsername string) *GuildAggregate {
	now := time.Now()

	guild := &GuildAggregate{
		BaseAggregate:         cqrs.NewBaseAggregate(id, "Guild"),
		name:                  name,
		description:           description,
		notice:                "",
		tag:                   "",
		status:                GuildStatusActive,
		maxMembers:            50, // Default max members
		isPublic:              true,
		requireApproval:       false,
		minLevel:              1,
		members:               make(map[string]*GuildMember),
		treasury:              0,
		mines:                 make(map[string]*Mine),
		transports:            make(map[string]*Transport),
		transportRecruitments: make(map[string]*TransportRecruitment),
		totalContribution:     0,
		level:                 1,
		experience:            0,
		ranking:               0,
		foundedAt:             now,
		lastActiveAt:          now,
	}

	// Add founder as guild leader
	founder := NewGuildMember(founderID, founderUsername, "")
	founder.Role = RoleLeader
	founder.Status = StatusActive
	guild.members[founderID] = founder

	// Apply guild created event
	event := NewGuildCreatedEvent(id, name, description, founderID, founderUsername)
	guild.Apply(event, true) // Use Apply with isNew=true to track the event

	return guild
}

// LoadGuildAggregate loads a guild aggregate from events
func LoadGuildAggregate(id string, events []cqrs.EventMessage) (*GuildAggregate, error) {
	guild := &GuildAggregate{
		BaseAggregate:         cqrs.NewBaseAggregate(id, "Guild"),
		members:               make(map[string]*GuildMember),
		mines:                 make(map[string]*Mine),
		transports:            make(map[string]*Transport),
		transportRecruitments: make(map[string]*TransportRecruitment),
	}

	for _, event := range events {
		if err := guild.ApplyEvent(event); err != nil {
			return nil, fmt.Errorf("failed to apply event %s: %w", event.EventType(), err)
		}
	}

	guild.ClearChanges()
	return guild, nil
}

// Guild basic operations

// UpdateInfo updates guild basic information
func (g *GuildAggregate) UpdateInfo(name, description, notice, tag string, updatedBy string) error {
	member, exists := g.members[updatedBy]
	if !exists {
		return fmt.Errorf("user %s is not a member of the guild", updatedBy)
	}

	if !member.HasPermission(PermissionManageGuild) {
		return fmt.Errorf("user %s does not have permission to manage guild", updatedBy)
	}

	event := NewGuildInfoUpdatedEvent(g.ID(), name, description, notice, tag, updatedBy)
	g.Apply(event, true)
	return nil
}

// UpdateSettings updates guild settings
func (g *GuildAggregate) UpdateSettings(maxMembers, minLevel int, isPublic, requireApproval bool, updatedBy string) error {
	member, exists := g.members[updatedBy]
	if !exists {
		return fmt.Errorf("user %s is not a member of the guild", updatedBy)
	}

	if !member.HasPermission(PermissionManageGuild) {
		return fmt.Errorf("user %s does not have permission to manage guild", updatedBy)
	}

	if maxMembers < len(g.GetActiveMembers()) {
		return fmt.Errorf("max members cannot be less than current active members count")
	}

	event := NewGuildSettingsUpdatedEvent(g.ID(), maxMembers, minLevel, isPublic, requireApproval, updatedBy)
	g.Apply(event, true)
	return nil
}

// Member management operations

// InviteMember invites a new member to the guild
func (g *GuildAggregate) InviteMember(userID, username, invitedBy string) error {
	if g.status != GuildStatusActive {
		return fmt.Errorf("guild is not active")
	}

	inviter, exists := g.members[invitedBy]
	if !exists {
		return fmt.Errorf("inviter %s is not a member of the guild", invitedBy)
	}

	if !inviter.HasPermission(PermissionInviteMembers) {
		return fmt.Errorf("user %s does not have permission to invite members", invitedBy)
	}

	if _, exists := g.members[userID]; exists {
		return fmt.Errorf("user %s is already a member of the guild", userID)
	}

	if len(g.members) >= g.maxMembers {
		return fmt.Errorf("guild has reached maximum member capacity")
	}

	event := NewMemberInvitedEvent(g.ID(), userID, username, invitedBy)
	g.Apply(event, true)
	return nil
}

// AcceptInvitation accepts an invitation to join the guild
func (g *GuildAggregate) AcceptInvitation(userID string) error {
	member, exists := g.members[userID]
	if !exists {
		return fmt.Errorf("user %s was not invited to the guild", userID)
	}

	if member.Status != StatusPending {
		return fmt.Errorf("user %s does not have a pending invitation", userID)
	}

	event := NewMemberJoinedEvent(g.ID(), userID)
	g.Apply(event, true)
	return nil
}

// KickMember kicks a member from the guild
func (g *GuildAggregate) KickMember(userID, kickedBy, reason string) error {
	member, exists := g.members[userID]
	if !exists {
		return fmt.Errorf("user %s is not a member of the guild", userID)
	}

	kicker, exists := g.members[kickedBy]
	if !exists {
		return fmt.Errorf("kicker %s is not a member of the guild", kickedBy)
	}

	if !kicker.CanKick(member.Role) {
		return fmt.Errorf("user %s cannot kick member with role %s", kickedBy, member.Role.String())
	}

	if userID == kickedBy {
		return fmt.Errorf("cannot kick yourself")
	}

	event := NewMemberKickedEvent(g.ID(), userID, kickedBy, reason)
	g.Apply(event, true)
	return nil
}

// PromoteMember promotes a member to a higher role
func (g *GuildAggregate) PromoteMember(userID, promotedBy string, newRole GuildRole) error {
	member, exists := g.members[userID]
	if !exists {
		return fmt.Errorf("user %s is not a member of the guild", userID)
	}

	promoter, exists := g.members[promotedBy]
	if !exists {
		return fmt.Errorf("promoter %s is not a member of the guild", promotedBy)
	}

	if !promoter.CanPromote(newRole) {
		return fmt.Errorf("user %s cannot promote to role %s", promotedBy, newRole.String())
	}

	if newRole <= member.Role {
		return fmt.Errorf("new role must be higher than current role")
	}

	event := NewMemberPromotedEvent(g.ID(), userID, promotedBy, member.Role, newRole)
	g.Apply(event, true)
	return nil
}

// Getters

// GetName returns the guild name
func (g *GuildAggregate) GetName() string {
	return g.name
}

// GetDescription returns the guild description
func (g *GuildAggregate) GetDescription() string {
	return g.description
}

// GetStatus returns the guild status
func (g *GuildAggregate) GetStatus() GuildStatus {
	return g.status
}

// GetMember returns a guild member by user ID
func (g *GuildAggregate) GetMember(userID string) (*GuildMember, bool) {
	member, exists := g.members[userID]
	if exists {
		return member.Clone(), true
	}
	return nil, false
}

// GetActiveMembers returns all active members
func (g *GuildAggregate) GetActiveMembers() []*GuildMember {
	activeMembers := make([]*GuildMember, 0)
	for _, member := range g.members {
		if member.IsActive() {
			activeMembers = append(activeMembers, member.Clone())
		}
	}
	return activeMembers
}

// GetMemberCount returns the total number of members
func (g *GuildAggregate) GetMemberCount() int {
	return len(g.members)
}

// GetActiveMemberCount returns the number of active members
func (g *GuildAggregate) GetActiveMemberCount() int {
	count := 0
	for _, member := range g.members {
		if member.IsActive() {
			count++
		}
	}
	return count
}

// GetTreasury returns the guild treasury amount
func (g *GuildAggregate) GetTreasury() int64 {
	return g.treasury
}

// GetLevel returns the guild level
func (g *GuildAggregate) GetLevel() int {
	return g.level
}

// Mining operations

// GetMining returns the guild mining state
func (g *GuildAggregate) GetMining() *GuildMining {
	if g.mining == nil {
		g.mining = NewGuildMining(g.ID())
	}
	return g.mining
}

// StartMiningOperation starts a new mining operation
func (g *GuildAggregate) StartMiningOperation(operationID, nodeID string, workerUserIDs []string, startedBy string) error {
	member, exists := g.members[startedBy]
	if !exists {
		return fmt.Errorf("user %s is not a member of the guild", startedBy)
	}

	if !member.HasPermission(PermissionManageMining) {
		return fmt.Errorf("user %s does not have permission to manage mining", startedBy)
	}

	// Validate all workers are guild members
	workers := make([]*MiningWorker, 0, len(workerUserIDs))
	for _, userID := range workerUserIDs {
		if workerMember, exists := g.members[userID]; exists && workerMember.IsActive() {
			worker := &MiningWorker{
				UserID:     userID,
				Username:   workerMember.Username,
				AssignedAt: time.Now(),
				Efficiency: 1.0,
				Experience: 0,
				Level:      1,
			}
			workers = append(workers, worker)
		} else {
			return fmt.Errorf("user %s is not an active guild member", userID)
		}
	}

	mining := g.GetMining()
	if err := mining.StartMiningOperation(operationID, nodeID, workers); err != nil {
		return err
	}

	event := NewMiningOperationStartedEvent(g.ID(), operationID, nodeID, workerUserIDs, startedBy)
	g.Apply(event, true)
	return nil
}

// HarvestMinerals harvests minerals from a mining operation
func (g *GuildAggregate) HarvestMinerals(operationID string, harvestedBy string) (map[MineralType]int64, error) {
	member, exists := g.members[harvestedBy]
	if !exists {
		return nil, fmt.Errorf("user %s is not a member of the guild", harvestedBy)
	}

	if !member.HasPermission(PermissionManageMining) {
		return nil, fmt.Errorf("user %s does not have permission to manage mining", harvestedBy)
	}

	mining := g.GetMining()
	harvested, err := mining.HarvestMinerals(operationID)
	if err != nil {
		return nil, err
	}

	if len(harvested) > 0 {
		// Calculate treasury value from harvested minerals
		treasuryIncrease := int64(0)
		for mineralType, amount := range harvested {
			treasuryIncrease += amount * mineralType.GetValue()
		}
		g.treasury += treasuryIncrease

		event := NewMineralsHarvestedEvent(g.ID(), operationID, harvested, treasuryIncrease, harvestedBy)
		g.Apply(event, true)
	}

	return harvested, nil
}

// StopMiningOperation stops a mining operation
func (g *GuildAggregate) StopMiningOperation(operationID string, stoppedBy string) error {
	member, exists := g.members[stoppedBy]
	if !exists {
		return fmt.Errorf("user %s is not a member of the guild", stoppedBy)
	}

	if !member.HasPermission(PermissionManageMining) {
		return fmt.Errorf("user %s does not have permission to manage mining", stoppedBy)
	}

	mining := g.GetMining()
	if err := mining.StopMiningOperation(operationID); err != nil {
		return err
	}

	event := NewMiningOperationStoppedEvent(g.ID(), operationID, stoppedBy)
	g.Apply(event, true)
	return nil
}

// Event application methods

// Apply applies an event to the aggregate (overrides BaseAggregate.Apply)
func (g *GuildAggregate) Apply(event cqrs.EventMessage, isNew bool) {
	// Call base implementation for infrastructure concerns
	g.BaseAggregate.Apply(event, isNew)

	// Apply domain-specific logic
	if err := g.applyDomainEvent(event); err != nil {
		// In a real implementation, you might want to handle this differently
		panic(fmt.Sprintf("failed to apply event: %v", err))
	}
}

// ApplyEvent applies an event to the aggregate (for event replay)
func (g *GuildAggregate) ApplyEvent(event cqrs.EventMessage) error {
	return g.applyDomainEvent(event)
}

// applyDomainEvent applies domain-specific event logic
func (g *GuildAggregate) applyDomainEvent(event cqrs.EventMessage) error {
	switch e := event.(type) {
	case *GuildCreatedEvent:
		return g.applyGuildCreatedEvent(e)
	case *GuildInfoUpdatedEvent:
		return g.applyGuildInfoUpdatedEvent(e)
	case *GuildSettingsUpdatedEvent:
		return g.applyGuildSettingsUpdatedEvent(e)
	case *MemberInvitedEvent:
		return g.applyMemberInvitedEvent(e)
	case *MemberJoinedEvent:
		return g.applyMemberJoinedEvent(e)
	case *MemberKickedEvent:
		return g.applyMemberKickedEvent(e)
	case *MemberPromotedEvent:
		return g.applyMemberPromotedEvent(e)
	case *MiningOperationStartedEvent:
		return g.applyMiningOperationStartedEvent(e)
	case *MineralsHarvestedEvent:
		return g.applyMineralsHarvestedEvent(e)
	case *MiningOperationStoppedEvent:
		return g.applyMiningOperationStoppedEvent(e)
	case *TransportRecruitmentCreatedEvent:
		return g.applyTransportRecruitmentCreatedEvent(e)
	case *TransportRecruitmentJoinedEvent:
		return g.applyTransportRecruitmentJoinedEvent(e)
	case *TransportRecruitmentLeftEvent:
		return g.applyTransportRecruitmentLeftEvent(e)
	case *TransportRecruitmentStartedEvent:
		return g.applyTransportRecruitmentStartedEvent(e)
	case *TransportRecruitmentCompletedEvent:
		return g.applyTransportRecruitmentCompletedEvent(e)
	default:
		return fmt.Errorf("unknown event type: %s", event.EventType())
	}
}

func (g *GuildAggregate) applyGuildCreatedEvent(event *GuildCreatedEvent) error {
	g.name = event.Name
	g.description = event.Description
	g.foundedAt = event.Timestamp()
	g.lastActiveAt = event.Timestamp()

	// Add founder as leader
	founder := NewGuildMember(event.FounderID, event.FounderUsername, "")
	founder.Role = RoleLeader
	founder.Status = StatusActive
	g.members[event.FounderID] = founder

	return nil
}

func (g *GuildAggregate) applyGuildInfoUpdatedEvent(event *GuildInfoUpdatedEvent) error {
	g.name = event.Name
	g.description = event.Description
	g.notice = event.Notice
	g.tag = event.Tag
	g.lastActiveAt = event.Timestamp()

	return nil
}

func (g *GuildAggregate) applyGuildSettingsUpdatedEvent(event *GuildSettingsUpdatedEvent) error {
	g.maxMembers = event.MaxMembers
	g.minLevel = event.MinLevel
	g.isPublic = event.IsPublic
	g.requireApproval = event.RequireApproval
	g.lastActiveAt = event.Timestamp()

	return nil
}

func (g *GuildAggregate) applyMemberInvitedEvent(event *MemberInvitedEvent) error {
	member := NewGuildMember(event.UserID, event.Username, event.InvitedBy)
	g.members[event.UserID] = member
	g.lastActiveAt = event.Timestamp()

	return nil
}

func (g *GuildAggregate) applyMemberJoinedEvent(event *MemberJoinedEvent) error {
	if member, exists := g.members[event.UserID]; exists {
		member.Activate()
		g.lastActiveAt = event.Timestamp()
	}

	return nil
}

func (g *GuildAggregate) applyMemberKickedEvent(event *MemberKickedEvent) error {
	if member, exists := g.members[event.UserID]; exists {
		member.Kick(event.KickedBy, event.Reason)
		g.lastActiveAt = event.Timestamp()
	}

	return nil
}

func (g *GuildAggregate) applyMemberPromotedEvent(event *MemberPromotedEvent) error {
	if member, exists := g.members[event.UserID]; exists {
		member.Role = event.NewRole
		g.lastActiveAt = event.Timestamp()
	}

	return nil
}

// Validation

// Validate validates the guild aggregate
func (g *GuildAggregate) Validate() error {
	if g.name == "" {
		return fmt.Errorf("guild name cannot be empty")
	}
	if g.maxMembers <= 0 {
		return fmt.Errorf("max members must be positive")
	}
	if g.minLevel < 1 {
		return fmt.Errorf("min level must be at least 1")
	}

	// Validate that there is at least one leader
	hasLeader := false
	for _, member := range g.members {
		if member.Role == RoleLeader && member.IsActive() {
			hasLeader = true
			break
		}
	}
	if !hasLeader {
		return fmt.Errorf("guild must have at least one active leader")
	}

	return nil
}

// Transport Recruitment operations

// CreateTransportRecruitment creates a new transport recruitment posting
func (g *GuildAggregate) CreateTransportRecruitment(recruitmentID, title, description string,
	maxParticipants, minParticipants int, duration, transportTime time.Duration,
	totalCargo map[MineralType]int64, createdBy string) error {

	member, exists := g.members[createdBy]
	if !exists {
		return fmt.Errorf("user %s is not a member of the guild", createdBy)
	}

	if !member.HasPermission(PermissionManageTransport) {
		return fmt.Errorf("user %s does not have permission to manage transport", createdBy)
	}

	// Check if there's already an active recruitment
	for _, recruitment := range g.transportRecruitments {
		if recruitment.Status == RecruitmentStatusOpen || recruitment.Status == RecruitmentStatusFull {
			return fmt.Errorf("there is already an active transport recruitment")
		}
	}

	recruitment := NewTransportRecruitment(recruitmentID, g.ID(), createdBy, member.Username,
		title, description, maxParticipants, minParticipants, duration, transportTime, totalCargo)

	if err := recruitment.Validate(); err != nil {
		return fmt.Errorf("invalid recruitment data: %w", err)
	}

	g.transportRecruitments[recruitmentID] = recruitment

	// Apply event
	event := NewTransportRecruitmentCreatedEvent(g.ID(), recruitmentID, title, description,
		maxParticipants, minParticipants, int64(duration), int64(transportTime), totalCargo, createdBy, member.Username)
	g.Apply(event, true)

	return nil
}

// JoinTransportRecruitment allows a member to join a transport recruitment
func (g *GuildAggregate) JoinTransportRecruitment(recruitmentID, userID string) error {
	member, exists := g.members[userID]
	if !exists {
		return fmt.Errorf("user %s is not a member of the guild", userID)
	}

	if !member.IsActive() {
		return fmt.Errorf("user %s is not an active member", userID)
	}

	recruitment, exists := g.transportRecruitments[recruitmentID]
	if !exists {
		return fmt.Errorf("transport recruitment %s not found", recruitmentID)
	}

	if err := recruitment.JoinRecruitment(userID, member.Username); err != nil {
		return err
	}

	// Apply event
	event := NewTransportRecruitmentJoinedEvent(g.ID(), recruitmentID, userID, member.Username)
	g.Apply(event, true)

	return nil
}

// LeaveTransportRecruitment allows a member to leave a transport recruitment
func (g *GuildAggregate) LeaveTransportRecruitment(recruitmentID, userID string) error {
	member, exists := g.members[userID]
	if !exists {
		return fmt.Errorf("user %s is not a member of the guild", userID)
	}

	recruitment, exists := g.transportRecruitments[recruitmentID]
	if !exists {
		return fmt.Errorf("transport recruitment %s not found", recruitmentID)
	}

	if err := recruitment.LeaveRecruitment(userID); err != nil {
		return err
	}

	// Apply event
	event := NewTransportRecruitmentLeftEvent(g.ID(), recruitmentID, userID, member.Username)
	g.Apply(event, true)

	return nil
}

// StartTransportFromRecruitment starts transport from a recruitment
func (g *GuildAggregate) StartTransportFromRecruitment(recruitmentID, transportID string, startedBy string) error {
	member, exists := g.members[startedBy]
	if !exists {
		return fmt.Errorf("user %s is not a member of the guild", startedBy)
	}

	if !member.HasPermission(PermissionManageTransport) {
		return fmt.Errorf("user %s does not have permission to manage transport", startedBy)
	}

	recruitment, exists := g.transportRecruitments[recruitmentID]
	if !exists {
		return fmt.Errorf("transport recruitment %s not found", recruitmentID)
	}

	if !recruitment.CanStart() {
		return fmt.Errorf("recruitment cannot be started: status=%s, participants=%d, min=%d",
			recruitment.Status.String(), recruitment.GetParticipantCount(), recruitment.MinParticipants)
	}

	if err := recruitment.StartTransport(transportID); err != nil {
		return err
	}

	// Apply event
	event := NewTransportRecruitmentStartedEvent(g.ID(), recruitmentID, transportID, startedBy)
	g.Apply(event, true)

	return nil
}

// CompleteTransportRecruitment completes a transport recruitment and distributes rewards
func (g *GuildAggregate) CompleteTransportRecruitment(recruitmentID string, completedBy string) (map[string]map[MineralType]int64, error) {
	member, exists := g.members[completedBy]
	if !exists {
		return nil, fmt.Errorf("user %s is not a member of the guild", completedBy)
	}

	if !member.HasPermission(PermissionManageTransport) {
		return nil, fmt.Errorf("user %s does not have permission to manage transport", completedBy)
	}

	recruitment, exists := g.transportRecruitments[recruitmentID]
	if !exists {
		return nil, fmt.Errorf("transport recruitment %s not found", recruitmentID)
	}

	if err := recruitment.CompleteTransport(); err != nil {
		return nil, err
	}

	// Distribute rewards to participants
	rewards := make(map[string]map[MineralType]int64)
	for userID, participant := range recruitment.Participants {
		rewards[userID] = participant.ExpectedReward
	}

	// Apply event
	event := NewTransportRecruitmentCompletedEvent(g.ID(), recruitmentID, rewards, completedBy)
	g.Apply(event, true)

	return rewards, nil
}

// ForceCompleteTransportRecruitment forcefully completes a transport recruitment (for testing/demo purposes)
func (g *GuildAggregate) ForceCompleteTransportRecruitment(recruitmentID string, completedBy string) (map[string]map[MineralType]int64, error) {
	member, exists := g.members[completedBy]
	if !exists {
		return nil, fmt.Errorf("user %s is not a member of the guild", completedBy)
	}

	if !member.HasPermission(PermissionManageTransport) {
		return nil, fmt.Errorf("user %s does not have permission to manage transport", completedBy)
	}

	recruitment, exists := g.transportRecruitments[recruitmentID]
	if !exists {
		return nil, fmt.Errorf("transport recruitment %s not found", recruitmentID)
	}

	if err := recruitment.ForceCompleteTransport(); err != nil {
		return nil, err
	}

	// Distribute rewards to participants
	rewards := make(map[string]map[MineralType]int64)
	for userID, participant := range recruitment.Participants {
		rewards[userID] = participant.ExpectedReward
	}

	// Apply event
	event := NewTransportRecruitmentCompletedEvent(g.ID(), recruitmentID, rewards, completedBy)
	g.Apply(event, true)

	return rewards, nil
}

// GetActiveTransportRecruitments returns all active transport recruitments
func (g *GuildAggregate) GetActiveTransportRecruitments() []*TransportRecruitment {
	var active []*TransportRecruitment
	for _, recruitment := range g.transportRecruitments {
		if recruitment.Status == RecruitmentStatusOpen || recruitment.Status == RecruitmentStatusFull {
			active = append(active, recruitment)
		}
	}
	return active
}

// GetTransportRecruitment returns a specific transport recruitment
func (g *GuildAggregate) GetTransportRecruitment(recruitmentID string) (*TransportRecruitment, bool) {
	recruitment, exists := g.transportRecruitments[recruitmentID]
	return recruitment, exists
}

// Mining event handlers

func (g *GuildAggregate) applyMiningOperationStartedEvent(event *MiningOperationStartedEvent) error {
	// Initialize mining if not exists
	if g.mining == nil {
		g.mining = NewGuildMining(g.ID())
	}

	g.lastActiveAt = event.Timestamp()
	return nil
}

func (g *GuildAggregate) applyMineralsHarvestedEvent(event *MineralsHarvestedEvent) error {
	// Update treasury
	g.treasury += event.TreasuryIncrease
	g.lastActiveAt = event.Timestamp()
	return nil
}

func (g *GuildAggregate) applyMiningOperationStoppedEvent(event *MiningOperationStoppedEvent) error {
	g.lastActiveAt = event.Timestamp()
	return nil
}

// Transport Recruitment event handlers

func (g *GuildAggregate) applyTransportRecruitmentCreatedEvent(event *TransportRecruitmentCreatedEvent) error {
	duration := time.Duration(event.Duration)
	transportTime := time.Duration(event.TransportTime)

	// Create recruitment
	recruitment := NewTransportRecruitment(event.RecruitmentID, g.ID(), event.CreatedBy, event.CreatedByUsername,
		event.Title, event.Description, event.MaxParticipants, event.MinParticipants, duration, transportTime, event.TotalCargo)

	// Set created time from event
	recruitment.CreatedAt = event.Timestamp()
	recruitment.ExpiresAt = event.Timestamp().Add(duration)

	g.transportRecruitments[event.RecruitmentID] = recruitment
	g.lastActiveAt = event.Timestamp()
	return nil
}

func (g *GuildAggregate) applyTransportRecruitmentJoinedEvent(event *TransportRecruitmentJoinedEvent) error {
	if recruitment, exists := g.transportRecruitments[event.RecruitmentID]; exists {
		participant := &TransportParticipant{
			UserID:         event.UserID,
			Username:       event.Username,
			JoinedAt:       event.Timestamp(),
			ExpectedReward: recruitment.RewardPerPerson,
		}
		recruitment.Participants[event.UserID] = participant

		// Update status if full
		if len(recruitment.Participants) >= recruitment.MaxParticipants {
			recruitment.Status = RecruitmentStatusFull
		}
	}

	g.lastActiveAt = event.Timestamp()
	return nil
}

func (g *GuildAggregate) applyTransportRecruitmentLeftEvent(event *TransportRecruitmentLeftEvent) error {
	if recruitment, exists := g.transportRecruitments[event.RecruitmentID]; exists {
		delete(recruitment.Participants, event.UserID)

		// Update status if no longer full
		if recruitment.Status == RecruitmentStatusFull && len(recruitment.Participants) < recruitment.MaxParticipants {
			recruitment.Status = RecruitmentStatusOpen
		}
	}

	g.lastActiveAt = event.Timestamp()
	return nil
}

func (g *GuildAggregate) applyTransportRecruitmentStartedEvent(event *TransportRecruitmentStartedEvent) error {
	if recruitment, exists := g.transportRecruitments[event.RecruitmentID]; exists {
		recruitment.Status = RecruitmentStatusStarted
		recruitment.TransportID = event.TransportID
		now := event.Timestamp()
		recruitment.StartedAt = &now
	}

	g.lastActiveAt = event.Timestamp()
	return nil
}

func (g *GuildAggregate) applyTransportRecruitmentCompletedEvent(event *TransportRecruitmentCompletedEvent) error {
	if recruitment, exists := g.transportRecruitments[event.RecruitmentID]; exists {
		now := event.Timestamp()
		recruitment.CompletedAt = &now
	}

	g.lastActiveAt = event.Timestamp()
	return nil
}

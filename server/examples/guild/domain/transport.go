package domain

import (
	"fmt"
	"time"
)

// TransportRecruitmentStatus represents the status of a transport recruitment
type TransportRecruitmentStatus int

const (
	// RecruitmentStatusOpen represents an open recruitment
	RecruitmentStatusOpen TransportRecruitmentStatus = iota
	// RecruitmentStatusFull represents a full recruitment (max participants reached)
	RecruitmentStatusFull
	// RecruitmentStatusExpired represents an expired recruitment
	RecruitmentStatusExpired
	// RecruitmentStatusStarted represents a recruitment that has started transport
	RecruitmentStatusStarted
	// RecruitmentStatusCancelled represents a cancelled recruitment
	RecruitmentStatusCancelled
)

// String returns the string representation of the recruitment status
func (s TransportRecruitmentStatus) String() string {
	switch s {
	case RecruitmentStatusOpen:
		return "Open"
	case RecruitmentStatusFull:
		return "Full"
	case RecruitmentStatusExpired:
		return "Expired"
	case RecruitmentStatusStarted:
		return "Started"
	case RecruitmentStatusCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// TransportParticipant represents a participant in transport recruitment
type TransportParticipant struct {
	UserID         string                `json:"user_id"`
	Username       string                `json:"username"`
	JoinedAt       time.Time             `json:"joined_at"`
	ExpectedReward map[MineralType]int64 `json:"expected_reward"` // Expected mineral reward
}

// TransportRecruitment represents a transport recruitment posting
type TransportRecruitment struct {
	ID                string `json:"id"`
	GuildID           string `json:"guild_id"`
	CreatedBy         string `json:"created_by"`
	CreatedByUsername string `json:"created_by_username"`
	Title             string `json:"title"`
	Description       string `json:"description"`

	// Recruitment settings
	MaxParticipants int           `json:"max_participants"`
	MinParticipants int           `json:"min_participants"`
	Duration        time.Duration `json:"duration"`       // How long recruitment stays open
	TransportTime   time.Duration `json:"transport_time"` // How long transport takes

	// Cargo and rewards
	TotalCargo      map[MineralType]int64 `json:"total_cargo"`       // Total minerals to transport
	RewardPerPerson map[MineralType]int64 `json:"reward_per_person"` // Reward per participant

	// Status and timing
	Status      TransportRecruitmentStatus `json:"status"`
	CreatedAt   time.Time                  `json:"created_at"`
	ExpiresAt   time.Time                  `json:"expires_at"`
	StartedAt   *time.Time                 `json:"started_at,omitempty"`
	CompletedAt *time.Time                 `json:"completed_at,omitempty"`

	// Participants
	Participants map[string]*TransportParticipant `json:"participants"` // userID -> participant

	// Related transport
	TransportID string `json:"transport_id,omitempty"`
}

// NewTransportRecruitment creates a new transport recruitment
func NewTransportRecruitment(id, guildID, createdBy, createdByUsername, title, description string,
	maxParticipants, minParticipants int, duration, transportTime time.Duration,
	totalCargo map[MineralType]int64) *TransportRecruitment {

	now := time.Now()

	// Calculate reward per person
	rewardPerPerson := make(map[MineralType]int64)
	for mineralType, amount := range totalCargo {
		rewardPerPerson[mineralType] = amount / int64(maxParticipants)
	}

	return &TransportRecruitment{
		ID:                id,
		GuildID:           guildID,
		CreatedBy:         createdBy,
		CreatedByUsername: createdByUsername,
		Title:             title,
		Description:       description,
		MaxParticipants:   maxParticipants,
		MinParticipants:   minParticipants,
		Duration:          duration,
		TransportTime:     transportTime,
		TotalCargo:        totalCargo,
		RewardPerPerson:   rewardPerPerson,
		Status:            RecruitmentStatusOpen,
		CreatedAt:         now,
		ExpiresAt:         now.Add(duration),
		Participants:      make(map[string]*TransportParticipant),
	}
}

// JoinRecruitment allows a user to join the transport recruitment
func (tr *TransportRecruitment) JoinRecruitment(userID, username string) error {
	if tr.Status != RecruitmentStatusOpen {
		return fmt.Errorf("recruitment is not open for joining, current status: %s", tr.Status.String())
	}

	if time.Now().After(tr.ExpiresAt) {
		tr.Status = RecruitmentStatusExpired
		return fmt.Errorf("recruitment has expired")
	}

	if _, exists := tr.Participants[userID]; exists {
		return fmt.Errorf("user %s is already participating in this recruitment", userID)
	}

	if len(tr.Participants) >= tr.MaxParticipants {
		tr.Status = RecruitmentStatusFull
		return fmt.Errorf("recruitment is full")
	}

	participant := &TransportParticipant{
		UserID:         userID,
		Username:       username,
		JoinedAt:       time.Now(),
		ExpectedReward: tr.RewardPerPerson,
	}

	tr.Participants[userID] = participant

	// Check if recruitment is now full
	if len(tr.Participants) >= tr.MaxParticipants {
		tr.Status = RecruitmentStatusFull
	}

	return nil
}

// LeaveRecruitment allows a user to leave the transport recruitment
func (tr *TransportRecruitment) LeaveRecruitment(userID string) error {
	if tr.Status == RecruitmentStatusStarted || tr.Status == RecruitmentStatusCancelled {
		return fmt.Errorf("cannot leave recruitment with status: %s", tr.Status.String())
	}

	if _, exists := tr.Participants[userID]; !exists {
		return fmt.Errorf("user %s is not participating in this recruitment", userID)
	}

	delete(tr.Participants, userID)

	// If recruitment was full, change status back to open
	if tr.Status == RecruitmentStatusFull && len(tr.Participants) < tr.MaxParticipants {
		tr.Status = RecruitmentStatusOpen
	}

	return nil
}

// StartTransport starts the transport operation
func (tr *TransportRecruitment) StartTransport(transportID string) error {
	if tr.Status != RecruitmentStatusFull && tr.Status != RecruitmentStatusOpen {
		return fmt.Errorf("recruitment must be full or open to start transport, current status: %s", tr.Status.String())
	}

	if len(tr.Participants) < tr.MinParticipants {
		return fmt.Errorf("not enough participants: %d (minimum: %d)", len(tr.Participants), tr.MinParticipants)
	}

	tr.Status = RecruitmentStatusStarted
	tr.TransportID = transportID
	now := time.Now()
	tr.StartedAt = &now

	return nil
}

// CompleteTransport completes the transport operation
func (tr *TransportRecruitment) CompleteTransport() error {
	if tr.Status != RecruitmentStatusStarted {
		return fmt.Errorf("recruitment must be started to complete transport, current status: %s", tr.Status.String())
	}

	if tr.StartedAt == nil {
		return fmt.Errorf("transport has not been started")
	}

	// Check if enough time has passed
	if time.Now().Before(tr.StartedAt.Add(tr.TransportTime)) {
		return fmt.Errorf("transport is not yet complete")
	}

	now := time.Now()
	tr.CompletedAt = &now

	return nil
}

// ForceCompleteTransport forcefully completes the transport operation (for testing/demo purposes)
func (tr *TransportRecruitment) ForceCompleteTransport() error {
	if tr.Status != RecruitmentStatusStarted {
		return fmt.Errorf("recruitment must be started to complete transport, current status: %s", tr.Status.String())
	}

	if tr.StartedAt == nil {
		return fmt.Errorf("transport has not been started")
	}

	now := time.Now()
	tr.CompletedAt = &now

	return nil
}

// CancelRecruitment cancels the recruitment
func (tr *TransportRecruitment) CancelRecruitment() error {
	if tr.Status == RecruitmentStatusStarted {
		return fmt.Errorf("cannot cancel recruitment that has already started transport")
	}

	tr.Status = RecruitmentStatusCancelled
	return nil
}

// IsExpired checks if the recruitment has expired
func (tr *TransportRecruitment) IsExpired() bool {
	return time.Now().After(tr.ExpiresAt) && tr.Status == RecruitmentStatusOpen
}

// CanJoin checks if a user can join the recruitment
func (tr *TransportRecruitment) CanJoin() bool {
	return tr.Status == RecruitmentStatusOpen &&
		len(tr.Participants) < tr.MaxParticipants &&
		time.Now().Before(tr.ExpiresAt)
}

// CanStart checks if the transport can be started
func (tr *TransportRecruitment) CanStart() bool {
	return (tr.Status == RecruitmentStatusOpen || tr.Status == RecruitmentStatusFull) &&
		len(tr.Participants) >= tr.MinParticipants
}

// IsCompleted checks if the transport is completed
func (tr *TransportRecruitment) IsCompleted() bool {
	return tr.CompletedAt != nil
}

// GetParticipantCount returns the number of participants
func (tr *TransportRecruitment) GetParticipantCount() int {
	return len(tr.Participants)
}

// GetRemainingTime returns the remaining time until expiration
func (tr *TransportRecruitment) GetRemainingTime() time.Duration {
	if tr.Status != RecruitmentStatusOpen {
		return 0
	}
	remaining := time.Until(tr.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetTransportRemainingTime returns the remaining time until transport completion
func (tr *TransportRecruitment) GetTransportRemainingTime() time.Duration {
	if tr.Status != RecruitmentStatusStarted || tr.StartedAt == nil {
		return 0
	}

	completionTime := tr.StartedAt.Add(tr.TransportTime)
	remaining := time.Until(completionTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Validate validates the recruitment data
func (tr *TransportRecruitment) Validate() error {
	if tr.ID == "" {
		return fmt.Errorf("recruitment ID cannot be empty")
	}
	if tr.GuildID == "" {
		return fmt.Errorf("guild ID cannot be empty")
	}
	if tr.CreatedBy == "" {
		return fmt.Errorf("created by cannot be empty")
	}
	if tr.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}
	if tr.MaxParticipants <= 0 {
		return fmt.Errorf("max participants must be positive")
	}
	if tr.MinParticipants <= 0 {
		return fmt.Errorf("min participants must be positive")
	}
	if tr.MinParticipants > tr.MaxParticipants {
		return fmt.Errorf("min participants cannot be greater than max participants")
	}
	if tr.Duration <= 0 {
		return fmt.Errorf("duration must be positive")
	}
	if tr.TransportTime <= 0 {
		return fmt.Errorf("transport time must be positive")
	}
	if len(tr.TotalCargo) == 0 {
		return fmt.Errorf("total cargo cannot be empty")
	}
	return nil
}

// TransportStatus represents the status of a transport operation
type TransportStatus int

const (
	// StatusPreparing represents a transport being prepared
	StatusPreparing TransportStatus = iota
	// StatusInTransit represents a transport in progress
	StatusInTransit
	// StatusUnderAttack represents a transport under attack
	StatusUnderAttack
	// StatusDefended represents a successfully defended transport
	StatusDefended
	// StatusRaided represents a transport that was raided
	StatusRaided
	// StatusCompleted represents a completed transport
	StatusCompleted
	// StatusCancelled represents a cancelled transport
	StatusCancelled
)

// String returns the string representation of the transport status
func (s TransportStatus) String() string {
	switch s {
	case StatusPreparing:
		return "Preparing"
	case StatusInTransit:
		return "InTransit"
	case StatusUnderAttack:
		return "UnderAttack"
	case StatusDefended:
		return "Defended"
	case StatusRaided:
		return "Raided"
	case StatusCompleted:
		return "Completed"
	case StatusCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// TransportCargo represents the cargo being transported
type TransportCargo struct {
	MineralType MineralType `json:"mineral_type"`
	Amount      int64       `json:"amount"`
	Value       int64       `json:"value"`
}

// GetTotalValue returns the total value of the cargo
func (c *TransportCargo) GetTotalValue() int64 {
	return c.Value
}

// Attack represents an attack on a transport
type Attack struct {
	AttackerGuildID string    `json:"attacker_guild_id"`
	AttackerUserID  string    `json:"attacker_user_id"`
	AttackTime      time.Time `json:"attack_time"`
	AttackPower     int64     `json:"attack_power"`
	IsSuccessful    bool      `json:"is_successful"`
}

// Defense represents a defense action during an attack
type Defense struct {
	DefenderUserID string    `json:"defender_user_id"`
	DefenseTime    time.Time `json:"defense_time"`
	DefensePower   int64     `json:"defense_power"`
}

// Transport represents a transport operation from mine to guild
type Transport struct {
	ID               string            `json:"id"`
	GuildID          string            `json:"guild_id"`
	MineID           string            `json:"mine_id"`
	StartedBy        string            `json:"started_by"`
	Cargo            []*TransportCargo `json:"cargo"`
	Status           TransportStatus   `json:"status"`
	StartTime        time.Time         `json:"start_time"`
	EstimatedArrival time.Time         `json:"estimated_arrival"`
	ActualArrival    *time.Time        `json:"actual_arrival,omitempty"`
	Route            string            `json:"route"`
	Attacks          []*Attack         `json:"attacks"`
	Defenses         []*Defense        `json:"defenses"`
	DefenseDeadline  *time.Time        `json:"defense_deadline,omitempty"`
}

// NewTransport creates a new transport
func NewTransport(id, guildID, mineID, startedBy, route string, cargo []*TransportCargo, duration time.Duration) *Transport {
	now := time.Now()
	return &Transport{
		ID:               id,
		GuildID:          guildID,
		MineID:           mineID,
		StartedBy:        startedBy,
		Cargo:            cargo,
		Status:           StatusPreparing,
		StartTime:        now,
		EstimatedArrival: now.Add(duration),
		Route:            route,
		Attacks:          make([]*Attack, 0),
		Defenses:         make([]*Defense, 0),
	}
}

// Start starts the transport
func (t *Transport) Start() error {
	if t.Status != StatusPreparing {
		return fmt.Errorf("transport must be in preparing status to start, current status: %s", t.Status.String())
	}
	t.Status = StatusInTransit
	t.StartTime = time.Now()
	return nil
}

// Attack attacks the transport
func (t *Transport) Attack(attackerGuildID, attackerUserID string, attackPower int64) error {
	if t.Status != StatusInTransit {
		return fmt.Errorf("transport must be in transit to be attacked, current status: %s", t.Status.String())
	}

	// Cannot attack own guild's transport
	if attackerGuildID == t.GuildID {
		return fmt.Errorf("cannot attack own guild's transport")
	}

	attack := &Attack{
		AttackerGuildID: attackerGuildID,
		AttackerUserID:  attackerUserID,
		AttackTime:      time.Now(),
		AttackPower:     attackPower,
		IsSuccessful:    false,
	}

	t.Attacks = append(t.Attacks, attack)
	t.Status = StatusUnderAttack

	// Set defense deadline (e.g., 10 minutes to defend)
	deadline := time.Now().Add(10 * time.Minute)
	t.DefenseDeadline = &deadline

	return nil
}

// Defend defends the transport
func (t *Transport) Defend(defenderUserID string, defensePower int64) error {
	if t.Status != StatusUnderAttack {
		return fmt.Errorf("transport must be under attack to be defended, current status: %s", t.Status.String())
	}

	if t.DefenseDeadline != nil && time.Now().After(*t.DefenseDeadline) {
		return fmt.Errorf("defense deadline has passed")
	}

	defense := &Defense{
		DefenderUserID: defenderUserID,
		DefenseTime:    time.Now(),
		DefensePower:   defensePower,
	}

	t.Defenses = append(t.Defenses, defense)

	// Calculate if defense is successful
	totalAttackPower := t.GetTotalAttackPower()
	totalDefensePower := t.GetTotalDefensePower()

	if totalDefensePower >= totalAttackPower {
		t.Status = StatusDefended
		t.DefenseDeadline = nil
		// Mark the last attack as unsuccessful
		if len(t.Attacks) > 0 {
			t.Attacks[len(t.Attacks)-1].IsSuccessful = false
		}
	}

	return nil
}

// ProcessDefenseDeadline processes the defense deadline
func (t *Transport) ProcessDefenseDeadline() error {
	if t.Status != StatusUnderAttack {
		return nil
	}

	if t.DefenseDeadline == nil || time.Now().Before(*t.DefenseDeadline) {
		return nil
	}

	// Defense deadline passed, check if defense was successful
	totalAttackPower := t.GetTotalAttackPower()
	totalDefensePower := t.GetTotalDefensePower()

	if totalDefensePower < totalAttackPower {
		// Attack successful, transport is raided
		t.Status = StatusRaided
		if len(t.Attacks) > 0 {
			t.Attacks[len(t.Attacks)-1].IsSuccessful = true
		}
	} else {
		// Defense successful
		t.Status = StatusDefended
		if len(t.Attacks) > 0 {
			t.Attacks[len(t.Attacks)-1].IsSuccessful = false
		}
	}

	t.DefenseDeadline = nil
	return nil
}

// Complete completes the transport
func (t *Transport) Complete() error {
	if t.Status != StatusInTransit && t.Status != StatusDefended {
		return fmt.Errorf("transport must be in transit or defended to be completed, current status: %s", t.Status.String())
	}

	if time.Now().Before(t.EstimatedArrival) {
		return fmt.Errorf("transport has not reached its destination yet")
	}

	t.Status = StatusCompleted
	now := time.Now()
	t.ActualArrival = &now
	return nil
}

// Cancel cancels the transport
func (t *Transport) Cancel() error {
	if t.Status == StatusCompleted || t.Status == StatusRaided || t.Status == StatusCancelled {
		return fmt.Errorf("cannot cancel transport with status: %s", t.Status.String())
	}
	t.Status = StatusCancelled
	return nil
}

// GetTotalCargoValue returns the total value of all cargo
func (t *Transport) GetTotalCargoValue() int64 {
	totalValue := int64(0)
	for _, cargo := range t.Cargo {
		totalValue += cargo.GetTotalValue()
	}
	return totalValue
}

// GetTotalAttackPower returns the total attack power from all attacks
func (t *Transport) GetTotalAttackPower() int64 {
	totalPower := int64(0)
	for _, attack := range t.Attacks {
		totalPower += attack.AttackPower
	}
	return totalPower
}

// GetTotalDefensePower returns the total defense power from all defenses
func (t *Transport) GetTotalDefensePower() int64 {
	totalPower := int64(0)
	for _, defense := range t.Defenses {
		totalPower += defense.DefensePower
	}
	return totalPower
}

// GetRemainingTime returns the remaining time until arrival
func (t *Transport) GetRemainingTime() time.Duration {
	if t.Status == StatusCompleted || t.Status == StatusRaided || t.Status == StatusCancelled {
		return 0
	}
	remaining := time.Until(t.EstimatedArrival)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetDefenseTimeRemaining returns the remaining time to defend
func (t *Transport) GetDefenseTimeRemaining() time.Duration {
	if t.DefenseDeadline == nil {
		return 0
	}
	remaining := time.Until(*t.DefenseDeadline)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// IsUnderAttack returns true if the transport is under attack
func (t *Transport) IsUnderAttack() bool {
	return t.Status == StatusUnderAttack
}

// IsCompleted returns true if the transport is completed
func (t *Transport) IsCompleted() bool {
	return t.Status == StatusCompleted
}

// IsRaided returns true if the transport was raided
func (t *Transport) IsRaided() bool {
	return t.Status == StatusRaided
}

// CanBeAttacked returns true if the transport can be attacked
func (t *Transport) CanBeAttacked() bool {
	return t.Status == StatusInTransit
}

// CanBeDefended returns true if the transport can be defended
func (t *Transport) CanBeDefended() bool {
	return t.Status == StatusUnderAttack &&
		t.DefenseDeadline != nil &&
		time.Now().Before(*t.DefenseDeadline)
}

// Validate validates the transport data
func (t *Transport) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("transport ID cannot be empty")
	}
	if t.GuildID == "" {
		return fmt.Errorf("guild ID cannot be empty")
	}
	if t.MineID == "" {
		return fmt.Errorf("mine ID cannot be empty")
	}
	if t.StartedBy == "" {
		return fmt.Errorf("started by cannot be empty")
	}
	if len(t.Cargo) == 0 {
		return fmt.Errorf("transport must have cargo")
	}
	if t.Route == "" {
		return fmt.Errorf("route cannot be empty")
	}
	if t.EstimatedArrival.Before(t.StartTime) {
		return fmt.Errorf("estimated arrival cannot be before start time")
	}
	return nil
}

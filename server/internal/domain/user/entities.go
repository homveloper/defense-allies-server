package user

import (
	"time"
)

// UserInventory represents a user's inventory data
type UserInventory struct {
	UserID        string                `json:"user_id"`
	Items         map[string]*InventoryItem `json:"items"`
	Capacity      int                   `json:"capacity"`
	UsedSlots     int                   `json:"used_slots"`
	LastUpdated   time.Time             `json:"last_updated"`
	Version       int                   `json:"version"`
}

// InventoryItem represents an item in the user's inventory
type InventoryItem struct {
	ID          string                 `json:"id"`
	ItemType    string                 `json:"item_type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Quantity    int                    `json:"quantity"`
	Rarity      string                 `json:"rarity"`
	Properties  map[string]interface{} `json:"properties"`
	AcquiredAt  time.Time              `json:"acquired_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// NewUserInventory creates a new UserInventory
func NewUserInventory(userID string, capacity int) *UserInventory {
	return &UserInventory{
		UserID:      userID,
		Items:       make(map[string]*InventoryItem),
		Capacity:    capacity,
		UsedSlots:   0,
		LastUpdated: time.Now(),
		Version:     1,
	}
}

// AddItem adds an item to the inventory
func (inv *UserInventory) AddItem(item *InventoryItem) error {
	if inv.UsedSlots >= inv.Capacity {
		return ErrInventoryFull
	}
	
	inv.Items[item.ID] = item
	inv.UsedSlots++
	inv.LastUpdated = time.Now()
	inv.Version++
	
	return nil
}

// RemoveItem removes an item from the inventory
func (inv *UserInventory) RemoveItem(itemID string) error {
	if _, exists := inv.Items[itemID]; !exists {
		return ErrItemNotFound
	}
	
	delete(inv.Items, itemID)
	inv.UsedSlots--
	inv.LastUpdated = time.Now()
	inv.Version++
	
	return nil
}

// GetItem gets an item from the inventory
func (inv *UserInventory) GetItem(itemID string) (*InventoryItem, bool) {
	item, exists := inv.Items[itemID]
	return item, exists
}

// UserAchievements represents a user's achievements data
type UserAchievements struct {
	UserID           string                    `json:"user_id"`
	Achievements     map[string]*Achievement   `json:"achievements"`
	UnlockedCount    int                       `json:"unlocked_count"`
	TotalCount       int                       `json:"total_count"`
	AchievementPoints int                      `json:"achievement_points"`
	LastUpdated      time.Time                 `json:"last_updated"`
	Version          int                       `json:"version"`
}

// Achievement represents a single achievement
type Achievement struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Category      string                 `json:"category"`
	Points        int                    `json:"points"`
	Rarity        string                 `json:"rarity"`
	IsUnlocked    bool                   `json:"is_unlocked"`
	UnlockedAt    *time.Time             `json:"unlocked_at,omitempty"`
	Progress      map[string]int         `json:"progress"`
	Requirements  map[string]interface{} `json:"requirements"`
	Rewards       map[string]interface{} `json:"rewards"`
}

// NewUserAchievements creates a new UserAchievements
func NewUserAchievements(userID string) *UserAchievements {
	return &UserAchievements{
		UserID:           userID,
		Achievements:     make(map[string]*Achievement),
		UnlockedCount:    0,
		TotalCount:       0,
		AchievementPoints: 0,
		LastUpdated:      time.Now(),
		Version:          1,
	}
}

// UnlockAchievement unlocks an achievement
func (ach *UserAchievements) UnlockAchievement(achievementID string) error {
	achievement, exists := ach.Achievements[achievementID]
	if !exists {
		return ErrAchievementNotFound
	}
	
	if achievement.IsUnlocked {
		return ErrAchievementAlreadyUnlocked
	}
	
	now := time.Now()
	achievement.IsUnlocked = true
	achievement.UnlockedAt = &now
	
	ach.UnlockedCount++
	ach.AchievementPoints += achievement.Points
	ach.LastUpdated = time.Now()
	ach.Version++
	
	return nil
}

// GetAchievement gets an achievement
func (ach *UserAchievements) GetAchievement(achievementID string) (*Achievement, bool) {
	achievement, exists := ach.Achievements[achievementID]
	return achievement, exists
}

// UpdateProgress updates progress for an achievement
func (ach *UserAchievements) UpdateProgress(achievementID, progressKey string, value int) error {
	achievement, exists := ach.Achievements[achievementID]
	if !exists {
		return ErrAchievementNotFound
	}
	
	if achievement.Progress == nil {
		achievement.Progress = make(map[string]int)
	}
	
	achievement.Progress[progressKey] = value
	ach.LastUpdated = time.Now()
	ach.Version++
	
	return nil
}

// UserStats represents a user's statistics data
type UserStats struct {
	UserID      string                 `json:"user_id"`
	GameStats   map[string]*GameStat   `json:"game_stats"`
	GlobalStats map[string]interface{} `json:"global_stats"`
	LastUpdated time.Time              `json:"last_updated"`
	Version     int                    `json:"version"`
}

// GameStat represents statistics for a specific game or activity
type GameStat struct {
	GameType      string                 `json:"game_type"`
	GamesPlayed   int                    `json:"games_played"`
	GamesWon      int                    `json:"games_won"`
	GamesLost     int                    `json:"games_lost"`
	WinRate       float64                `json:"win_rate"`
	TotalScore    int64                  `json:"total_score"`
	HighestScore  int64                  `json:"highest_score"`
	AverageScore  float64                `json:"average_score"`
	TotalPlayTime time.Duration          `json:"total_play_time"`
	Rank          int                    `json:"rank"`
	Rating        int                    `json:"rating"`
	CustomStats   map[string]interface{} `json:"custom_stats"`
	LastPlayed    *time.Time             `json:"last_played,omitempty"`
}

// NewUserStats creates a new UserStats
func NewUserStats(userID string) *UserStats {
	return &UserStats{
		UserID:      userID,
		GameStats:   make(map[string]*GameStat),
		GlobalStats: make(map[string]interface{}),
		LastUpdated: time.Now(),
		Version:     1,
	}
}

// UpdateGameStat updates statistics for a specific game
func (stats *UserStats) UpdateGameStat(gameType string, update *GameStatUpdate) error {
	gameStat, exists := stats.GameStats[gameType]
	if !exists {
		gameStat = &GameStat{
			GameType:    gameType,
			CustomStats: make(map[string]interface{}),
		}
		stats.GameStats[gameType] = gameStat
	}
	
	// Apply updates
	if update.GamesPlayed != nil {
		gameStat.GamesPlayed += *update.GamesPlayed
	}
	if update.GamesWon != nil {
		gameStat.GamesWon += *update.GamesWon
	}
	if update.GamesLost != nil {
		gameStat.GamesLost += *update.GamesLost
	}
	if update.Score != nil {
		gameStat.TotalScore += *update.Score
		if *update.Score > gameStat.HighestScore {
			gameStat.HighestScore = *update.Score
		}
	}
	if update.PlayTime != nil {
		gameStat.TotalPlayTime += *update.PlayTime
	}
	if update.Rank != nil {
		gameStat.Rank = *update.Rank
	}
	if update.Rating != nil {
		gameStat.Rating = *update.Rating
	}
	
	// Update calculated fields
	if gameStat.GamesPlayed > 0 {
		gameStat.WinRate = float64(gameStat.GamesWon) / float64(gameStat.GamesPlayed)
		gameStat.AverageScore = float64(gameStat.TotalScore) / float64(gameStat.GamesPlayed)
	}
	
	// Update custom stats
	for key, value := range update.CustomStats {
		gameStat.CustomStats[key] = value
	}
	
	now := time.Now()
	gameStat.LastPlayed = &now
	stats.LastUpdated = time.Now()
	stats.Version++
	
	return nil
}

// GameStatUpdate represents an update to game statistics
type GameStatUpdate struct {
	GamesPlayed *int                   `json:"games_played,omitempty"`
	GamesWon    *int                   `json:"games_won,omitempty"`
	GamesLost   *int                   `json:"games_lost,omitempty"`
	Score       *int64                 `json:"score,omitempty"`
	PlayTime    *time.Duration         `json:"play_time,omitempty"`
	Rank        *int                   `json:"rank,omitempty"`
	Rating      *int                   `json:"rating,omitempty"`
	CustomStats map[string]interface{} `json:"custom_stats,omitempty"`
}

// GetGameStat gets statistics for a specific game
func (stats *UserStats) GetGameStat(gameType string) (*GameStat, bool) {
	gameStat, exists := stats.GameStats[gameType]
	return gameStat, exists
}

// UserPreferences represents a user's preferences data
type UserPreferences struct {
	UserID          string                 `json:"user_id"`
	GamePreferences map[string]interface{} `json:"game_preferences"`
	UIPreferences   map[string]interface{} `json:"ui_preferences"`
	Notifications   *NotificationSettings  `json:"notifications"`
	Privacy         *PrivacySettings       `json:"privacy"`
	Language        string                 `json:"language"`
	Timezone        string                 `json:"timezone"`
	Theme           string                 `json:"theme"`
	LastUpdated     time.Time              `json:"last_updated"`
	Version         int                    `json:"version"`
}

// NotificationSettings represents notification preferences
type NotificationSettings struct {
	EmailNotifications bool `json:"email_notifications"`
	PushNotifications  bool `json:"push_notifications"`
	GameNotifications  bool `json:"game_notifications"`
	SocialNotifications bool `json:"social_notifications"`
	SystemNotifications bool `json:"system_notifications"`
}

// PrivacySettings represents privacy preferences
type PrivacySettings struct {
	ProfileVisibility   string `json:"profile_visibility"`   // public, friends, private
	OnlineStatus        bool   `json:"online_status"`
	GameActivity        bool   `json:"game_activity"`
	ShowAchievements    bool   `json:"show_achievements"`
	ShowStats           bool   `json:"show_stats"`
	AllowFriendRequests bool   `json:"allow_friend_requests"`
}

// NewUserPreferences creates a new UserPreferences
func NewUserPreferences(userID string) *UserPreferences {
	return &UserPreferences{
		UserID:          userID,
		GamePreferences: make(map[string]interface{}),
		UIPreferences:   make(map[string]interface{}),
		Notifications: &NotificationSettings{
			EmailNotifications:  true,
			PushNotifications:   true,
			GameNotifications:   true,
			SocialNotifications: true,
			SystemNotifications: true,
		},
		Privacy: &PrivacySettings{
			ProfileVisibility:   "public",
			OnlineStatus:        true,
			GameActivity:        true,
			ShowAchievements:    true,
			ShowStats:           true,
			AllowFriendRequests: true,
		},
		Language:    "en",
		Timezone:    "UTC",
		Theme:       "default",
		LastUpdated: time.Now(),
		Version:     1,
	}
}

// UpdatePreference updates a specific preference
func (prefs *UserPreferences) UpdatePreference(category, key string, value interface{}) error {
	switch category {
	case "game":
		prefs.GamePreferences[key] = value
	case "ui":
		prefs.UIPreferences[key] = value
	default:
		return ErrInvalidPreferenceCategory
	}
	
	prefs.LastUpdated = time.Now()
	prefs.Version++
	
	return nil
}

// UserSocialData represents a user's social data
type UserSocialData struct {
	UserID      string               `json:"user_id"`
	Friends     map[string]*Friend   `json:"friends"`
	Blocked     map[string]*BlockedUser `json:"blocked"`
	Guilds      []string             `json:"guilds"`
	Groups      []string             `json:"groups"`
	SocialLinks map[string]string    `json:"social_links"`
	LastUpdated time.Time            `json:"last_updated"`
	Version     int                  `json:"version"`
}

// Friend represents a friend relationship
type Friend struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	Status      string    `json:"status"` // pending, accepted, blocked
	AddedAt     time.Time `json:"added_at"`
	AcceptedAt  *time.Time `json:"accepted_at,omitempty"`
	LastContact time.Time `json:"last_contact"`
}

// BlockedUser represents a blocked user
type BlockedUser struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Reason    string    `json:"reason"`
	BlockedAt time.Time `json:"blocked_at"`
}

// NewUserSocialData creates a new UserSocialData
func NewUserSocialData(userID string) *UserSocialData {
	return &UserSocialData{
		UserID:      userID,
		Friends:     make(map[string]*Friend),
		Blocked:     make(map[string]*BlockedUser),
		Guilds:      make([]string, 0),
		Groups:      make([]string, 0),
		SocialLinks: make(map[string]string),
		LastUpdated: time.Now(),
		Version:     1,
	}
}

// AddFriend adds a friend
func (social *UserSocialData) AddFriend(friendID, username string) error {
	// Check if already friends or blocked
	if _, exists := social.Friends[friendID]; exists {
		return ErrAlreadyFriends
	}
	if _, exists := social.Blocked[friendID]; exists {
		return ErrUserBlocked
	}
	
	friend := &Friend{
		UserID:      friendID,
		Username:    username,
		Status:      "pending",
		AddedAt:     time.Now(),
		LastContact: time.Now(),
	}
	
	social.Friends[friendID] = friend
	social.LastUpdated = time.Now()
	social.Version++
	
	return nil
}

// AcceptFriend accepts a friend request
func (social *UserSocialData) AcceptFriend(friendID string) error {
	friend, exists := social.Friends[friendID]
	if !exists {
		return ErrFriendNotFound
	}
	
	if friend.Status != "pending" {
		return ErrInvalidFriendStatus
	}
	
	now := time.Now()
	friend.Status = "accepted"
	friend.AcceptedAt = &now
	
	social.LastUpdated = time.Now()
	social.Version++
	
	return nil
}

// BlockUser blocks a user
func (social *UserSocialData) BlockUser(userID, username, reason string) error {
	// Remove from friends if exists
	delete(social.Friends, userID)
	
	blockedUser := &BlockedUser{
		UserID:    userID,
		Username:  username,
		Reason:    reason,
		BlockedAt: time.Now(),
	}
	
	social.Blocked[userID] = blockedUser
	social.LastUpdated = time.Now()
	social.Version++
	
	return nil
}

// Domain errors
var (
	ErrInventoryFull              = NewDomainError("INVENTORY_FULL", "Inventory is full")
	ErrItemNotFound               = NewDomainError("ITEM_NOT_FOUND", "Item not found")
	ErrAchievementNotFound        = NewDomainError("ACHIEVEMENT_NOT_FOUND", "Achievement not found")
	ErrAchievementAlreadyUnlocked = NewDomainError("ACHIEVEMENT_ALREADY_UNLOCKED", "Achievement already unlocked")
	ErrInvalidPreferenceCategory  = NewDomainError("INVALID_PREFERENCE_CATEGORY", "Invalid preference category")
	ErrAlreadyFriends             = NewDomainError("ALREADY_FRIENDS", "Already friends")
	ErrUserBlocked                = NewDomainError("USER_BLOCKED", "User is blocked")
	ErrFriendNotFound             = NewDomainError("FRIEND_NOT_FOUND", "Friend not found")
	ErrInvalidFriendStatus        = NewDomainError("INVALID_FRIEND_STATUS", "Invalid friend status")
)

// DomainError represents a domain-specific error
type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewDomainError creates a new domain error
func NewDomainError(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

// Error implements the error interface
func (e *DomainError) Error() string {
	return e.Message
}
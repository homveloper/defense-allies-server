package models

import (
	"time"
)

// Player 플레이어 모델 (Redis 저장용)
type Player struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // JSON에서 제외
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LastSeen  time.Time `json:"last_seen"`
	IsOnline  bool      `json:"is_online"`
}

// PlayerStats 플레이어 통계
type PlayerStats struct {
	PlayerID     string `json:"player_id"`
	GamesPlayed  int    `json:"games_played"`
	GamesWon     int    `json:"games_won"`
	GamesLost    int    `json:"games_lost"`
	TotalScore   int64  `json:"total_score"`
	HighestScore int64  `json:"highest_score"`
	Rank         int    `json:"rank"`
	Rating       int    `json:"rating"`
}

// PlayerSession 플레이어 세션 정보
type PlayerSession struct {
	PlayerID  string    `json:"player_id"`
	SessionID string    `json:"session_id"`
	GameID    string    `json:"game_id,omitempty"`
	Status    string    `json:"status"` // "idle", "matchmaking", "in_game"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreatePlayerRequest 플레이어 생성 요청 구조체
type CreatePlayerRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginRequest 로그인 요청 구조체
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// PlayerResponse 플레이어 응답 구조체 (비밀번호 제외)
type PlayerResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LastSeen  time.Time `json:"last_seen"`
	IsOnline  bool      `json:"is_online"`
}

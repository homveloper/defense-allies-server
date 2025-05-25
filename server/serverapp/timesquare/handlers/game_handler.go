package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"defense-allies-server/serverapp/timesquare/middleware"
	"defense-allies-server/serverapp/timesquare/service"
)

// GameHandler 게임 관련 핸들러
type GameHandler struct {
	userService *service.RedisUserService
}

// NewGameHandler 새로운 게임 핸들러 생성
func NewGameHandler(userService *service.RedisUserService) *GameHandler {
	return &GameHandler{
		userService: userService,
	}
}

// GetProfile 유저 프로필 조회
func (gh *GameHandler) GetProfile(ctx context.Context) (map[string]interface{}, error) {
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}

	return map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt,
		"last_login": user.LastLogin,
		"game_data":  user.GameData,
	}, nil
}

// UpdateGameData 게임 데이터 업데이트
func (gh *GameHandler) UpdateGameData(ctx context.Context, params UpdateGameDataParams) error {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("user ID not found in context")
	}

	// 현재 게임 데이터 조회
	currentData, err := gh.userService.GetUserGameData(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get current game data: %w", err)
	}

	// 데이터 업데이트
	if params.Level != nil {
		currentData.Level = *params.Level
	}
	if params.Score != nil {
		currentData.Score = *params.Score
	}
	if params.Resources != nil {
		for resource, amount := range params.Resources {
			currentData.Resources[resource] = amount
		}
	}
	if params.Settings != nil {
		for setting, value := range params.Settings {
			currentData.Settings[setting] = value
		}
	}

	// 데이터베이스 업데이트
	return gh.userService.UpdateUserGameData(ctx, userID, currentData)
}

// GetGameData 게임 데이터 조회
func (gh *GameHandler) GetGameData(ctx context.Context) (*middleware.GameData, error) {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("user ID not found in context")
	}

	return gh.userService.GetUserGameData(ctx, userID)
}

// GetLeaderboard 리더보드 조회
func (gh *GameHandler) GetLeaderboard(ctx context.Context, params LeaderboardParams) ([]LeaderboardEntry, error) {
	// 최근 활성 유저들 조회
	since := time.Now().AddDate(0, 0, -30) // 최근 30일
	limit := 100
	if params.Limit != nil && *params.Limit > 0 && *params.Limit <= 1000 {
		limit = *params.Limit
	}

	users, err := gh.userService.GetUsersByLastLogin(ctx, since, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	// 리더보드 엔트리 생성
	var entries []LeaderboardEntry
	for _, user := range users {
		if user.GameData != nil {
			entries = append(entries, LeaderboardEntry{
				UserID:   user.ID,
				Username: user.Username,
				Level:    user.GameData.Level,
				Score:    user.GameData.Score,
			})
		}
	}

	return entries, nil
}

// JoinGame 게임 참가
func (gh *GameHandler) JoinGame(ctx context.Context, params JoinGameParams) (*GameSession, error) {
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}

	// 게임 세션 생성
	session := &GameSession{
		ID:        fmt.Sprintf("game_%s_%d", user.ID, time.Now().Unix()),
		UserID:    user.ID,
		Username:  user.Username,
		GameType:  params.GameType,
		Status:    "waiting",
		CreatedAt: time.Now(),
	}

	// 실제로는 게임 매칭 서비스나 세션 관리자에 등록
	// 여기서는 간단한 예제로 바로 반환

	return session, nil
}

// 파라미터 구조체들

// UpdateGameDataParams 게임 데이터 업데이트 파라미터
type UpdateGameDataParams struct {
	Level     *int              `json:"level,omitempty"`
	Score     *int64            `json:"score,omitempty"`
	Resources map[string]int64  `json:"resources,omitempty"`
	Settings  map[string]string `json:"settings,omitempty"`
}

// LeaderboardParams 리더보드 조회 파라미터
type LeaderboardParams struct {
	Limit *int `json:"limit,omitempty"`
}

// JoinGameParams 게임 참가 파라미터
type JoinGameParams struct {
	GameType string `json:"game_type"`
}

// 응답 구조체들

// LeaderboardEntry 리더보드 엔트리
type LeaderboardEntry struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Level    int    `json:"level"`
	Score    int64  `json:"score"`
}

// GameSession 게임 세션
type GameSession struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	GameType  string    `json:"game_type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// HTTP 핸들러들 (RPC 핸들러와 별도)

// HTTPGetProfile HTTP GET 프로필 조회
func (gh *GameHandler) HTTPGetProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"created_at": user.CreatedAt,
			"last_login": user.LastLogin,
			"game_data":  user.GameData,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HTTPUpdateGameData HTTP POST 게임 데이터 업데이트
func (gh *GameHandler) HTTPUpdateGameData(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}

	var params UpdateGameDataParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 현재 게임 데이터 조회
	currentData, err := gh.userService.GetUserGameData(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get game data", http.StatusInternalServerError)
		return
	}

	// 데이터 업데이트
	if params.Level != nil {
		currentData.Level = *params.Level
	}
	if params.Score != nil {
		currentData.Score = *params.Score
	}
	if params.Resources != nil {
		for resource, amount := range params.Resources {
			currentData.Resources[resource] = amount
		}
	}
	if params.Settings != nil {
		for setting, value := range params.Settings {
			currentData.Settings[setting] = value
		}
	}

	// 데이터베이스 업데이트
	if err := gh.userService.UpdateUserGameData(r.Context(), userID, currentData); err != nil {
		http.Error(w, "Failed to update game data", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Game data updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

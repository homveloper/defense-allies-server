package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"defense-allies-server/serverapp/timesquare/middleware"
)

// GameHandler 게임 관련 핸들러
type GameHandler struct {
	userService middleware.UserService
}

// NewGameHandler 새로운 게임 핸들러 생성
func NewGameHandler(userService middleware.UserService) *GameHandler {
	return &GameHandler{
		userService: userService,
	}
}

// GetGameData 게임 데이터 조회
func (gh *GameHandler) GetGameData(ctx context.Context) (*middleware.GameData, error) {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("user ID not found in context")
	}

	return gh.userService.GetUserGameData(ctx, userID)
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

// GetProfile HTTP GET 프로필 조회
func (gh *GameHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
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

// UpdateGameData HTTP POST 게임 데이터 업데이트
func (gh *GameHandler) UpdateGameData(w http.ResponseWriter, r *http.Request) {
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

// GetLeaderboard HTTP GET 리더보드 조회
func (gh *GameHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	// 쿼리 파라미터에서 limit 가져오기
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if parsedLimit, err := fmt.Sscanf(limitStr, "%d", &limit); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	// 최근 30일 활성 유저 조회
	since := time.Now().AddDate(0, 0, -30)
	users, err := gh.userService.GetUsersByLastLogin(r.Context(), since, limit)
	if err != nil {
		http.Error(w, "Failed to get leaderboard", http.StatusInternalServerError)
		return
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

	response := map[string]interface{}{
		"success": true,
		"data":    entries,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StartSession HTTP POST 게임 세션 시작
func (gh *GameHandler) StartSession(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	var params JoinGameParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 게임 세션 생성
	session := &GameSession{
		ID:        fmt.Sprintf("game_%s_%d", user.ID, time.Now().Unix()),
		UserID:    user.ID,
		Username:  user.Username,
		GameType:  params.GameType,
		Status:    "active",
		CreatedAt: time.Now(),
	}

	// 실제로는 게임 세션 관리 서비스에 저장해야 함
	// 여기서는 간단히 응답만 반환

	response := map[string]interface{}{
		"success": true,
		"data":    session,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// EndSession HTTP POST 게임 세션 종료
func (gh *GameHandler) EndSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}

	var params struct {
		SessionID string `json:"session_id"`
		Score     *int64 `json:"score,omitempty"`
		Level     *int   `json:"level,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 게임 결과 업데이트
	if params.Score != nil || params.Level != nil {
		gameData, err := gh.userService.GetUserGameData(r.Context(), userID)
		if err != nil {
			http.Error(w, "Failed to get game data", http.StatusInternalServerError)
			return
		}

		if params.Score != nil && *params.Score > gameData.Score {
			gameData.Score = *params.Score
		}
		if params.Level != nil && *params.Level > gameData.Level {
			gameData.Level = *params.Level
		}

		if err := gh.userService.UpdateUserGameData(r.Context(), userID, gameData); err != nil {
			http.Error(w, "Failed to update game data", http.StatusInternalServerError)
			return
		}
	}

	// 실제로는 게임 세션 관리 서비스에서 세션을 종료해야 함

	response := map[string]interface{}{
		"success": true,
		"message": "Game session ended successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

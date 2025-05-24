package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"defense-allies-server/internal/models"
)

// PlayerHandler 플레이어 관련 핸들러
type PlayerHandler struct {
	// 추후 서비스 레이어 추가 예정
}

// NewPlayerHandler 새로운 플레이어 핸들러를 생성합니다
func NewPlayerHandler() *PlayerHandler {
	return &PlayerHandler{}
}

// CreatePlayer 플레이어 생성 핸들러
func (h *PlayerHandler) CreatePlayer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreatePlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// TODO: 실제 플레이어 생성 로직 구현
	response := models.PlayerResponse{
		ID:        "player_1",
		Username:  req.Username,
		Email:     req.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		LastSeen:  time.Now(),
		IsOnline:  true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetPlayer 플레이어 조회 핸들러
func (h *PlayerHandler) GetPlayer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: 실제 플레이어 조회 로직 구현
	response := models.PlayerResponse{
		ID:        "player_1",
		Username:  "testplayer",
		Email:     "test@example.com",
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
		LastSeen:  time.Now(),
		IsOnline:  true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

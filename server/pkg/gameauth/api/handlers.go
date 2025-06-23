package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"defense-allies-server/pkg/gameauth/application/auth"
	"defense-allies-server/pkg/gameauth/application/providers"
	"defense-allies-server/pkg/gameauth/domain/common"
)

type Handlers struct {
	authService *auth.Service
}

func NewHandlers(authService *auth.Service) *Handlers {
	return &Handlers{
		authService: authService,
	}
}

func (h *Handlers) LoginGuest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DeviceID   string                 `json:"device_id"`
		DeviceInfo common.DeviceInfo      `json:"device_info"`
		ClientInfo common.ClientInfo      `json:"client_info"`
		Metadata   map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.DeviceID == "" {
		http.Error(w, "device_id is required", http.StatusBadRequest)
		return
	}

	if req.DeviceInfo.DeviceID == "" {
		req.DeviceInfo.DeviceID = req.DeviceID
	}

	guestCredentials := &providers.GuestCredentials{
		DeviceID:   req.DeviceID,
		DeviceInfo: req.DeviceInfo,
		Metadata:   req.Metadata,
	}

	loginReq := &auth.LoginRequest{
		ProviderType: common.ProviderTypeGuest,
		Credentials:  guestCredentials,
		ClientInfo:   req.ClientInfo,
	}

	response, err := h.authService.Login(r.Context(), loginReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Login failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) RefreshSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req auth.RefreshSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, "refresh_token is required", http.StatusBadRequest)
		return
	}

	response, err := h.authService.RefreshSession(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Refresh failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionToken := r.Header.Get("Authorization")
	if sessionToken == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	if len(sessionToken) > 7 && sessionToken[:7] == "Bearer " {
		sessionToken = sessionToken[7:]
	}

	validateReq := &auth.ValidateSessionRequest{
		SessionToken: sessionToken,
	}

	response, err := h.authService.ValidateSession(r.Context(), validateReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Session validation failed: %v", err), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionToken := r.Header.Get("Authorization")
	if sessionToken == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	if len(sessionToken) > 7 && sessionToken[:7] == "Bearer " {
		sessionToken = sessionToken[7:]
	}

	logoutReq := &auth.LogoutRequest{
		SessionToken: sessionToken,
	}

	err := h.authService.Logout(r.Context(), logoutReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Logout failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Successfully logged out",
	})
}

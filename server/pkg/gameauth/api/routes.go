package api

import (
	"net/http"

	"defense-allies-server/pkg/gameauth/application/auth"
)

func SetupRoutes(mux *http.ServeMux, authService *auth.Service) {
	handlers := NewHandlers(authService)

	mux.HandleFunc("/api/v1/auth/login/guest", handlers.LoginGuest)
	mux.HandleFunc("/api/v1/auth/session/refresh", handlers.RefreshSession)
	mux.HandleFunc("/api/v1/account/profile", handlers.GetProfile)
	mux.HandleFunc("/api/v1/auth/logout", handlers.Logout)
}
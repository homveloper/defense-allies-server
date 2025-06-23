package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"defense-allies-server/pkg/gameauth/application/auth"
	"github.com/golang-jwt/jwt/v5"
)

// AuthClaims JWT 토큰의 클레임 구조
type AuthClaims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// AuthMiddleware 인증 미들웨어
type AuthMiddleware struct {
	publicKey   *rsa.PublicKey
	guardianURL string
	userService UserService
	authService *auth.Service // gameauth 서비스 추가
}

// UserService 유저 관련 서비스 인터페이스
type UserService interface {
	GetUser(ctx context.Context, userID string) (*User, error)
	CreateUser(ctx context.Context, userInfo *UserInfo) (*User, error)
	UpdateLastLogin(ctx context.Context, userID string) error
	GetUserGameData(ctx context.Context, userID string) (*GameData, error)
	UpdateUserGameData(ctx context.Context, userID string, gameData *GameData) error
	GetUsersByLastLogin(ctx context.Context, since time.Time, limit int) ([]*User, error)
}

// User 유저 정보
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LastLogin time.Time `json:"last_login"`
	GameData  *GameData `json:"game_data,omitempty"`
}

// UserInfo 유저 생성 정보
type UserInfo struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// GameData 게임 관련 데이터
type GameData struct {
	Level     int               `json:"level"`
	Score     int64             `json:"score"`
	Resources map[string]int64  `json:"resources"`
	Settings  map[string]string `json:"settings"`
}

// NewAuthMiddleware 새로운 인증 미들웨어 생성
func NewAuthMiddleware(publicKey *rsa.PublicKey, guardianURL string, userService UserService, authService *auth.Service) *AuthMiddleware {
	return &AuthMiddleware{
		publicKey:   publicKey,
		guardianURL: guardianURL,
		userService: userService,
		authService: authService,
	}
}

// Authenticate HTTP 요청 인증
func (am *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authorization 헤더에서 토큰 추출
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			am.sendError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		// Bearer 토큰 형식 확인
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			am.sendError(w, http.StatusUnauthorized, "Invalid authorization format")
			return
		}

		// 먼저 gameauth 세션 토큰 검증 시도
		gameAccountID, err := am.verifyGameAuthSession(r.Context(), tokenString)
		if err == nil {
			// gameauth 세션 토큰으로 성공 - 키 값만 context에 저장
			ctx := context.WithValue(r.Context(), "game_account_id", gameAccountID)
			ctx = context.WithValue(ctx, "auth_type", "gameauth")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// gameauth 실패 시 JWT 토큰 검증
		claims, err := am.verifyToken(tokenString)
		if err != nil {
			am.sendError(w, http.StatusUnauthorized, fmt.Sprintf("Invalid token: %v", err))
			return
		}

		// JWT 토큰 검증 성공 - 키 값만 context에 저장 (유저 생성은 핸들러에서 처리)
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "auth_type", "jwt")

		// 다음 핸들러로 전달
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// verifyToken JWT 토큰 검증
func (am *AuthMiddleware) verifyToken(tokenString string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 서명 방법 확인
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return am.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
		// 토큰 만료 확인
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return nil, fmt.Errorf("token expired")
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// verifyGameAuthSession gameauth 세션 토큰 검증 - game_account_id만 반환
func (am *AuthMiddleware) verifyGameAuthSession(ctx context.Context, sessionToken string) (string, error) {
	// gameauth 서비스를 통해 세션 검증만 수행
	validateReq := &auth.ValidateSessionRequest{
		SessionToken: sessionToken,
	}

	response, err := am.authService.ValidateSession(ctx, validateReq)
	if err != nil {
		return "", fmt.Errorf("gameauth session validation failed: %w", err)
	}

	// 인증 성공 - game_account_id만 반환 (유저 생성은 핸들러에서 처리)
	return response.GameAccountID, nil
}

// processUser 메서드 제거됨 - 유저 생성은 핸들러에서 처리

// sendError 에러 응답 전송
func (am *AuthMiddleware) sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":   message,
		"status":  statusCode,
		"success": false,
	}

	json.NewEncoder(w).Encode(response)
}

// GetGameAccountIDFromContext 컨텍스트에서 게임 계정 ID 추출 (gameauth 인증)
func GetGameAccountIDFromContext(ctx context.Context) (string, bool) {
	gameAccountID, ok := ctx.Value("game_account_id").(string)
	return gameAccountID, ok
}

// GetUserIDFromContext 컨텍스트에서 유저 ID 추출 (JWT 인증)
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value("user_id").(string)
	return userID, ok
}

// GetUsernameFromContext 컨텍스트에서 유저명 추출
func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value("username").(string)
	return username, ok
}

// GetEmailFromContext 컨텍스트에서 이메일 추출 (JWT 인증)
func GetEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value("email").(string)
	return email, ok
}

// GetAuthTypeFromContext 컨텍스트에서 인증 타입 추출
func GetAuthTypeFromContext(ctx context.Context) (string, bool) {
	authType, ok := ctx.Value("auth_type").(string)
	return authType, ok
}

// GetCurrentUserID 현재 인증된 유저의 ID 반환 (gameauth 또는 JWT)
func GetCurrentUserID(ctx context.Context) (string, bool) {
	// gameauth 인증 확인
	if gameAccountID, ok := GetGameAccountIDFromContext(ctx); ok {
		return gameAccountID, true
	}
	
	// JWT 인증 확인
	if userID, ok := GetUserIDFromContext(ctx); ok {
		return userID, true
	}
	
	return "", false
}

// RequireRole 특정 역할 요구 미들웨어
func (am *AuthMiddleware) RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 현재 유저 ID 확인
			userID, ok := GetCurrentUserID(r.Context())
			if !ok {
				am.sendError(w, http.StatusUnauthorized, "User not authenticated")
				return
			}

			// 유저 정보 조회 (역할 확인을 위해)
			user, err := am.userService.GetUser(r.Context(), userID)
			if err != nil {
				am.sendError(w, http.StatusUnauthorized, "User not found")
				return
			}

			_ = user // TODO: 역할 확인 로직 구현

			// 역할 확인 (실제 구현에서는 User 구조체에 Roles 필드 추가 필요)
			// hasRole := false
			// for _, role := range user.Roles {
			//     if role == requiredRole {
			//         hasRole = true
			//         break
			//     }
			// }
			//
			// if !hasRole {
			//     am.sendError(w, http.StatusForbidden, "Insufficient permissions")
			//     return
			// }

			next.ServeHTTP(w, r)
		})
	}
}

package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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
}

// UserService 유저 관련 서비스 인터페이스
type UserService interface {
	GetUser(ctx context.Context, userID string) (*User, error)
	CreateUser(ctx context.Context, userInfo *UserInfo) (*User, error)
	UpdateLastLogin(ctx context.Context, userID string) error
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
func NewAuthMiddleware(publicKey *rsa.PublicKey, guardianURL string, userService UserService) *AuthMiddleware {
	return &AuthMiddleware{
		publicKey:   publicKey,
		guardianURL: guardianURL,
		userService: userService,
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

		// JWT 토큰 검증
		claims, err := am.verifyToken(tokenString)
		if err != nil {
			am.sendError(w, http.StatusUnauthorized, fmt.Sprintf("Invalid token: %v", err))
			return
		}

		// 유저 정보 처리 (신규 유저 생성 포함)
		user, err := am.processUser(r.Context(), claims)
		if err != nil {
			am.sendError(w, http.StatusInternalServerError, fmt.Sprintf("User processing failed: %v", err))
			return
		}

		// 컨텍스트에 유저 정보 추가
		ctx := context.WithValue(r.Context(), "user", user)
		ctx = context.WithValue(ctx, "user_id", user.ID)
		ctx = context.WithValue(ctx, "username", user.Username)

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

// processUser 유저 정보 처리 (신규 유저 생성 포함)
func (am *AuthMiddleware) processUser(ctx context.Context, claims *AuthClaims) (*User, error) {
	// 기존 유저 조회
	user, err := am.userService.GetUser(ctx, claims.UserID)
	if err == nil {
		// 기존 유저 - 마지막 로그인 시간 업데이트
		if updateErr := am.userService.UpdateLastLogin(ctx, user.ID); updateErr != nil {
			// 로그인 시간 업데이트 실패는 치명적이지 않음
			fmt.Printf("Failed to update last login for user %s: %v\n", user.ID, updateErr)
		}
		return user, nil
	}

	// 유저가 없는 경우 신규 생성
	userInfo := &UserInfo{
		ID:       claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		Roles:    claims.Roles,
	}

	newUser, err := am.userService.CreateUser(ctx, userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	return newUser, nil
}

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

// GetUserFromContext 컨텍스트에서 유저 정보 추출
func GetUserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value("user").(*User)
	return user, ok
}

// GetUserIDFromContext 컨텍스트에서 유저 ID 추출
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value("user_id").(string)
	return userID, ok
}

// GetUsernameFromContext 컨텍스트에서 유저명 추출
func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value("username").(string)
	return username, ok
}

// RequireRole 특정 역할 요구 미들웨어
func (am *AuthMiddleware) RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := GetUserFromContext(r.Context())
			if !ok {
				am.sendError(w, http.StatusUnauthorized, "User not found in context")
				return
			}

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

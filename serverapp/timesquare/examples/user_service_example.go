package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	"defense-allies-server/serverapp/timesquare/middleware"
	"defense-allies-server/serverapp/timesquare/service"
	"github.com/redis/go-redis/v9"
)

// CustomUserCreationHandler 커스텀 유저 생성 핸들러 예제
func CustomUserCreationHandler(userInfo *middleware.UserInfo) (*middleware.User, error) {
	now := time.Now()
	
	// VIP 유저인지 확인
	isVIP := false
	for _, role := range userInfo.Roles {
		if role == "vip" || role == "premium" {
			isVIP = true
			break
		}
	}
	
	// VIP 유저는 더 많은 리소스로 시작
	var gameData *middleware.GameData
	if isVIP {
		gameData = &middleware.GameData{
			Level:     5, // VIP는 레벨 5부터 시작
			Score:     0,
			Resources: map[string]int64{
				"gold":   10000, // 10배 골드
				"gems":   500,   // 10배 젬
				"energy": 200,   // 2배 에너지
			},
			Settings: map[string]string{
				"language":      "en",
				"sound":         "on",
				"notifications": "on",
				"vip_status":    "active",
			},
		}
	} else {
		gameData = &middleware.GameData{
			Level:     1,
			Score:     0,
			Resources: map[string]int64{
				"gold":   1000,
				"gems":   50,
				"energy": 100,
			},
			Settings: map[string]string{
				"language":      "en",
				"sound":         "on",
				"notifications": "on",
			},
		}
	}
	
	return &middleware.User{
		ID:        userInfo.ID,
		Username:  userInfo.Username,
		Email:     userInfo.Email,
		CreatedAt: now,
		UpdatedAt: now,
		LastLogin: now,
		GameData:  gameData,
	}, nil
}

// ExampleUsage 사용 예제
func ExampleUsage() {
	// Redis 클라이언트 생성
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	
	// 커스텀 유저 생성 핸들러로 서비스 생성
	userService := service.NewRedisUserService(rdb, CustomUserCreationHandler)
	
	ctx := context.Background()
	
	// 1. 일반 유저 생성/조회
	regularUserInfo := &middleware.UserInfo{
		ID:       "user123",
		Username: "player123",
		Email:    "player@example.com",
		Roles:    []string{"player"},
	}
	
	user, err := userService.(*service.RedisUserService).FindAndUpsertUser(ctx, regularUserInfo)
	if err != nil {
		log.Printf("Failed to find/create regular user: %v", err)
		return
	}
	
	fmt.Printf("Regular User Created: %+v\n", user)
	fmt.Printf("Regular User Game Data: %+v\n", user.GameData)
	
	// 2. VIP 유저 생성/조회
	vipUserInfo := &middleware.UserInfo{
		ID:       "vip456",
		Username: "vipplayer456",
		Email:    "vip@example.com",
		Roles:    []string{"player", "vip"},
	}
	
	vipUser, err := userService.(*service.RedisUserService).FindAndUpsertUser(ctx, vipUserInfo)
	if err != nil {
		log.Printf("Failed to find/create VIP user: %v", err)
		return
	}
	
	fmt.Printf("VIP User Created: %+v\n", vipUser)
	fmt.Printf("VIP User Game Data: %+v\n", vipUser.GameData)
	
	// 3. 동일한 유저 다시 조회 (기존 유저 반환)
	existingUser, err := userService.(*service.RedisUserService).FindAndUpsertUser(ctx, regularUserInfo)
	if err != nil {
		log.Printf("Failed to find existing user: %v", err)
		return
	}
	
	fmt.Printf("Existing User Found: %+v\n", existingUser)
	
	// 4. 동시성 테스트 시뮬레이션
	fmt.Println("\n=== 동시성 테스트 시작 ===")
	testConcurrency(userService, ctx)
}

// testConcurrency 동시성 테스트
func testConcurrency(userService middleware.UserService, ctx context.Context) {
	const numGoroutines = 10
	const testUserID = "concurrent_test_user"
	
	// 테스트 유저 정보
	testUserInfo := &middleware.UserInfo{
		ID:       testUserID,
		Username: "concurrentuser",
		Email:    "concurrent@example.com",
		Roles:    []string{"player"},
	}
	
	// 결과 채널
	results := make(chan *middleware.User, numGoroutines)
	errors := make(chan error, numGoroutines)
	
	// 동시에 여러 고루틴에서 같은 유저 생성 시도
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			user, err := userService.(*service.RedisUserService).FindAndUpsertUser(ctx, testUserInfo)
			if err != nil {
				errors <- fmt.Errorf("goroutine %d failed: %w", id, err)
				return
			}
			results <- user
		}(i)
	}
	
	// 결과 수집
	var successCount int
	var errorCount int
	
	for i := 0; i < numGoroutines; i++ {
		select {
		case user := <-results:
			successCount++
			fmt.Printf("Goroutine success: User ID = %s\n", user.ID)
		case err := <-errors:
			errorCount++
			fmt.Printf("Goroutine error: %v\n", err)
		case <-time.After(5 * time.Second):
			fmt.Printf("Timeout waiting for goroutine %d\n", i)
			errorCount++
		}
	}
	
	fmt.Printf("\n동시성 테스트 결과:\n")
	fmt.Printf("- 성공: %d/%d\n", successCount, numGoroutines)
	fmt.Printf("- 실패: %d/%d\n", errorCount, numGoroutines)
	
	// 최종 유저 상태 확인
	finalUser, err := userService.GetUser(ctx, testUserID)
	if err != nil {
		fmt.Printf("최종 유저 조회 실패: %v\n", err)
	} else {
		fmt.Printf("최종 유저 상태: %+v\n", finalUser)
	}
}

// ExampleWithDefaultHandler 기본 핸들러 사용 예제
func ExampleWithDefaultHandler() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	
	// 기본 유저 생성 핸들러 사용
	userService := service.NewRedisUserService(rdb, service.DefaultUserCreationHandler)
	
	ctx := context.Background()
	
	userInfo := &middleware.UserInfo{
		ID:       "default_user",
		Username: "defaultplayer",
		Email:    "default@example.com",
		Roles:    []string{"player"},
	}
	
	user, err := userService.(*service.RedisUserService).FindAndUpsertUser(ctx, userInfo)
	if err != nil {
		log.Printf("Failed to create user with default handler: %v", err)
		return
	}
	
	fmt.Printf("User created with default handler: %+v\n", user)
}

// ExampleAuthMiddlewareIntegration 인증 미들웨어와 통합 예제
func ExampleAuthMiddlewareIntegration() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	
	// 유저 서비스 생성
	userService := service.NewRedisUserService(rdb, CustomUserCreationHandler)
	
	// 인증 미들웨어에서 사용하는 방식 시뮬레이션
	ctx := context.Background()
	
	// JWT에서 추출된 유저 정보 시뮬레이션
	jwtUserInfo := &middleware.UserInfo{
		ID:       "auth0|12345",
		Username: "jwtuser",
		Email:    "jwt@example.com",
		Roles:    []string{"player", "premium"},
	}
	
	// FindAndUpsertUser로 유저 조회/생성
	user, err := userService.(*service.RedisUserService).FindAndUpsertUser(ctx, jwtUserInfo)
	if err != nil {
		log.Printf("Auth middleware user processing failed: %v", err)
		return
	}
	
	fmt.Printf("Auth middleware processed user: %+v\n", user)
	
	// 두 번째 요청 시뮬레이션 (기존 유저 조회)
	existingUser, err := userService.(*service.RedisUserService).FindAndUpsertUser(ctx, jwtUserInfo)
	if err != nil {
		log.Printf("Second request failed: %v", err)
		return
	}
	
	fmt.Printf("Second request returned existing user: %+v\n", existingUser)
	fmt.Printf("Last login updated: %v\n", existingUser.LastLogin.After(user.LastLogin))
}

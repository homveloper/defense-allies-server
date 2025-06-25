package main

import (
	"context"
	"fmt"
	"time"

	"defense-allies-server/internal/domain/user"
)

func main() {
	fmt.Println("🎯 타입 세이프 지연 로딩 시스템 데모")
	fmt.Println("=====================================")

	// 1. 기본 타입 세이프 로딩
	fmt.Println("\n1️⃣ 기본 타입 세이프 로딩")
	demoBasicTypeSafeLoading()

	// 2. 멀티 타입 로딩
	fmt.Println("\n2️⃣ 멀티 타입 동시 로딩")
	demoMultiTypeLoading()

	// 3. 배치 로딩
	fmt.Println("\n3️⃣ 배치 로딩")
	demoBatchLoading()

	// 4. 캐시 관리
	fmt.Println("\n4️⃣ 캐시 관리")
	demoCacheManagement()

	// 5. 성능 비교
	fmt.Println("\n5️⃣ 성능 비교 (캐시 vs 로딩)")
	demoPerformanceComparison()

	fmt.Println("\n✅ 모든 데모 완료!")
}

func demoBasicTypeSafeLoading() {
	storage := user.NewInMemoryStorage()
	defer storage.Close()

	// 타입 세이프 인벤토리 로더 생성
	inventoryLoader := user.NewGenericLazyLoader(user.GenericLazyLoaderConfig[*user.UserInventory]{
		CacheStorage:   storage,
		DefaultTTL:     10 * time.Minute,
		CacheKeyPrefix: "demo_inventory",
		LoaderFunc: func(ctx context.Context, userID string) (*user.UserInventory, error) {
			fmt.Printf("  📀 데이터베이스에서 %s 인벤토리 로딩...\n", userID)
			
			inventory := user.NewUserInventory(userID, 100)
			
			// 사용자별 다른 아이템 추가
			switch userID {
			case "warrior":
				inventory.AddItem(&user.InventoryItem{
					ID: "steel_sword", Name: "강철 검", Quantity: 1,
					ItemType: "weapon", Rarity: "uncommon", AcquiredAt: time.Now(),
				})
				inventory.AddItem(&user.InventoryItem{
					ID: "health_potion", Name: "체력 물약", Quantity: 3,
					ItemType: "consumable", Rarity: "common", AcquiredAt: time.Now(),
				})
			case "mage":
				inventory.AddItem(&user.InventoryItem{
					ID: "magic_staff", Name: "마법 지팡이", Quantity: 1,
					ItemType: "weapon", Rarity: "rare", AcquiredAt: time.Now(),
				})
				inventory.AddItem(&user.InventoryItem{
					ID: "mana_potion", Name: "마나 물약", Quantity: 5,
					ItemType: "consumable", Rarity: "common", AcquiredAt: time.Now(),
				})
			}
			
			return inventory, nil
		},
	})

	ctx := context.Background()

	// 타입 세이프 로딩 - 캐스팅 불필요!
	fmt.Println("  🛡️ 전사 인벤토리 로딩...")
	warriorInventory, err := inventoryLoader.Load(ctx, "warrior")
	if err != nil {
		fmt.Printf("  ❌ 오류: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 전사 인벤토리: %d/%d 슬롯 사용\n", 
		warriorInventory.UsedSlots, warriorInventory.Capacity)
	for itemID, item := range warriorInventory.Items {
		fmt.Printf("    - %s: %s (%s 등급, 수량: %d)\n", 
			itemID, item.Name, item.Rarity, item.Quantity)
	}

	fmt.Println("  🧙 마법사 인벤토리 로딩...")
	mageInventory, err := inventoryLoader.Load(ctx, "mage")
	if err != nil {
		fmt.Printf("  ❌ 오류: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 마법사 인벤토리: %d/%d 슬롯 사용\n", 
		mageInventory.UsedSlots, mageInventory.Capacity)
	for itemID, item := range mageInventory.Items {
		fmt.Printf("    - %s: %s (%s 등급, 수량: %d)\n", 
			itemID, item.Name, item.Rarity, item.Quantity)
	}

	// 두 번째 로딩 (캐시에서)
	fmt.Println("  ⚡ 전사 인벤토리 재로딩 (캐시)...")
	warriorInventory2, err := inventoryLoader.Load(ctx, "warrior")
	if err != nil {
		fmt.Printf("  ❌ 오류: %v\n", err)
		return
	}
	fmt.Printf("  ✅ 캐시에서 빠르게 로딩됨: %d 아이템\n", warriorInventory2.UsedSlots)
}

func demoMultiTypeLoading() {
	storage := user.NewInMemoryStorage()
	defer storage.Close()

	multiLoader := user.NewMultiTypeLazyLoader(user.MultiTypeLazyLoaderConfig{
		CacheStorage:   storage,
		DefaultTTL:     15 * time.Minute,
		CacheKeyPrefix: "demo_multi",
	})

	ctx := context.Background()
	userID := "demo_player"

	fmt.Printf("  👤 사용자 %s의 모든 데이터 동시 로딩...\n", userID)
	
	start := time.Now()
	userData, err := multiLoader.LoadAll(ctx, userID)
	elapsed := time.Since(start)
	
	if err != nil {
		fmt.Printf("  ❌ 오류: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 모든 데이터 로딩 완료 (%v)\n", elapsed)
	fmt.Printf("    - 인벤토리: %d/%d 슬롯\n", 
		userData.Inventory.UsedSlots, userData.Inventory.Capacity)
	fmt.Printf("    - 업적: %d점 (%d개 달성)\n", 
		userData.Achievements.AchievementPoints, userData.Achievements.UnlockedCount)
	fmt.Printf("    - 스탯: %d개 게임 유형 추적\n", 
		len(userData.Stats.GameStats))
	fmt.Printf("    - 설정: %s 테마, %s 언어\n", 
		userData.Preferences.Theme, userData.Preferences.Language)
	fmt.Printf("    - 소셜: %d명 친구, %d개 길드\n", 
		len(userData.SocialData.Friends), len(userData.SocialData.Guilds))

	// 개별 로딩도 가능 (타입 세이프)
	fmt.Println("  📊 개별 스탯 로딩...")
	stats, err := multiLoader.LoadStats(ctx, userID)
	if err != nil {
		fmt.Printf("  ❌ 오류: %v\n", err)
		return
	}
	
	fmt.Printf("  ✅ 스탯 로딩: 레벨 %v\n", stats.GlobalStats["level"])
}

func demoBatchLoading() {
	storage := user.NewInMemoryStorage()
	defer storage.Close()

	achievementsLoader := user.NewGenericLazyLoader(user.GenericLazyLoaderConfig[*user.UserAchievements]{
		CacheStorage:   storage,
		DefaultTTL:     10 * time.Minute,
		CacheKeyPrefix: "demo_achievements",
		LoaderFunc: func(ctx context.Context, userID string) (*user.UserAchievements, error) {
			fmt.Printf("  📀 %s 업적 로딩...\n", userID)
			
			achievements := user.NewUserAchievements(userID)
			
			// 사용자별 다른 업적
			now := time.Now()
			switch userID {
			case "veteran_player":
				achievements.Achievements["veteran"] = &user.Achievement{
					ID: "veteran", Name: "베테랑", Category: "experience",
					Points: 100, Rarity: "epic", IsUnlocked: true, UnlockedAt: &now,
				}
				achievements.UnlockedCount = 1
				achievements.AchievementPoints = 100
			case "rookie_player":
				achievements.Achievements["first_steps"] = &user.Achievement{
					ID: "first_steps", Name: "첫 걸음", Category: "onboarding",
					Points: 10, Rarity: "common", IsUnlocked: true, UnlockedAt: &now,
				}
				achievements.UnlockedCount = 1
				achievements.AchievementPoints = 10
			case "pro_player":
				achievements.Achievements["master"] = &user.Achievement{
					ID: "master", Name: "마스터", Category: "skill",
					Points: 500, Rarity: "legendary", IsUnlocked: true, UnlockedAt: &now,
				}
				achievements.UnlockedCount = 1
				achievements.AchievementPoints = 500
			}
			
			return achievements, nil
		},
	})

	ctx := context.Background()
	userIDs := []string{"veteran_player", "rookie_player", "pro_player"}

	fmt.Printf("  👥 %d명의 업적 배치 로딩...\n", len(userIDs))
	
	start := time.Now()
	results, err := achievementsLoader.LoadMultiple(ctx, userIDs)
	elapsed := time.Since(start)
	
	if err != nil {
		fmt.Printf("  ❌ 오류: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 배치 로딩 완료 (%v)\n", elapsed)
	
	for userID, achievements := range results {
		fmt.Printf("    - %s: %d점 (%d개 업적)\n", 
			userID, achievements.AchievementPoints, achievements.UnlockedCount)
		
		for _, achievement := range achievements.Achievements {
			fmt.Printf("      * %s (%s 등급, %d점)\n", 
				achievement.Name, achievement.Rarity, achievement.Points)
		}
	}
}

func demoCacheManagement() {
	storage := user.NewInMemoryStorage()
	defer storage.Close()

	loadCount := 0
	
	preferencesLoader := user.NewGenericLazyLoader(user.GenericLazyLoaderConfig[*user.UserPreferences]{
		CacheStorage:   storage,
		DefaultTTL:     5 * time.Minute,
		CacheKeyPrefix: "demo_preferences",
		LoaderFunc: func(ctx context.Context, userID string) (*user.UserPreferences, error) {
			loadCount++
			fmt.Printf("  📀 데이터베이스 로딩 #%d for %s\n", loadCount, userID)
			
			preferences := user.NewUserPreferences(userID)
			preferences.Theme = "dynamic"
			preferences.Language = "ko"
			
			return preferences, nil
		},
	})

	ctx := context.Background()
	userID := "cache_demo_user"

	// 첫 번째 로딩
	fmt.Println("  🔄 첫 번째 로딩 (데이터베이스)")
	prefs1, err := preferencesLoader.Load(ctx, userID)
	if err != nil {
		fmt.Printf("  ❌ 오류: %v\n", err)
		return
	}
	fmt.Printf("  ✅ 로딩 완료: %s 테마, %s 언어\n", prefs1.Theme, prefs1.Language)

	// 캐시 정보 확인
	cacheInfo, err := preferencesLoader.GetCacheInfo(ctx, userID)
	if err != nil {
		fmt.Printf("  ❌ 캐시 정보 오류: %v\n", err)
		return
	}
	fmt.Printf("  📊 캐시 상태: 존재=%v, TTL=%v\n", cacheInfo.Exists, cacheInfo.TTL)

	// 두 번째 로딩 (캐시)
	fmt.Println("  ⚡ 두 번째 로딩 (캐시)")
	prefs2, err := preferencesLoader.Load(ctx, userID)
	if err != nil {
		fmt.Printf("  ❌ 오류: %v\n", err)
		return
	}
	fmt.Printf("  ✅ 캐시에서 로딩: %s 테마\n", prefs2.Theme)

	// 캐시 무효화
	fmt.Println("  🗑️ 캐시 무효화")
	err = preferencesLoader.InvalidateCache(ctx, userID)
	if err != nil {
		fmt.Printf("  ❌ 무효화 오류: %v\n", err)
		return
	}

	// 세 번째 로딩 (다시 데이터베이스)
	fmt.Println("  🔄 세 번째 로딩 (무효화 후)")
	prefs3, err := preferencesLoader.Load(ctx, userID)
	if err != nil {
		fmt.Printf("  ❌ 오류: %v\n", err)
		return
	}
	fmt.Printf("  ✅ 재로딩 완료: %s 테마 (총 %d회 DB 호출)\n", prefs3.Theme, loadCount)
}

func demoPerformanceComparison() {
	storage := user.NewInMemoryStorage()
	defer storage.Close()

	multiLoader := user.NewMultiTypeLazyLoader(user.MultiTypeLazyLoaderConfig{
		CacheStorage:   storage,
		DefaultTTL:     30 * time.Minute,
		CacheKeyPrefix: "perf_demo",
	})

	ctx := context.Background()
	userIDs := []string{"user1", "user2", "user3", "user4", "user5"}

	// 첫 번째: 데이터베이스에서 로딩
	fmt.Printf("  📊 %d명 사용자 데이터 로딩 성능 테스트\n", len(userIDs))
	
	fmt.Println("  🔄 첫 번째: 데이터베이스에서 로딩")
	start := time.Now()
	for _, userID := range userIDs {
		_, err := multiLoader.LoadAll(ctx, userID)
		if err != nil {
			fmt.Printf("  ❌ 오류: %v\n", err)
			return
		}
	}
	dbLoadTime := time.Since(start)
	fmt.Printf("  ⏱️ DB 로딩 시간: %v (평균: %v/사용자)\n", 
		dbLoadTime, dbLoadTime/time.Duration(len(userIDs)))

	// 두 번째: 캐시에서 로딩
	fmt.Println("  ⚡ 두 번째: 캐시에서 로딩")
	start = time.Now()
	for _, userID := range userIDs {
		_, err := multiLoader.LoadAll(ctx, userID)
		if err != nil {
			fmt.Printf("  ❌ 오류: %v\n", err)
			return
		}
	}
	cacheLoadTime := time.Since(start)
	fmt.Printf("  ⏱️ 캐시 로딩 시간: %v (평균: %v/사용자)\n", 
		cacheLoadTime, cacheLoadTime/time.Duration(len(userIDs)))

	// 성능 개선 비교
	speedup := float64(dbLoadTime) / float64(cacheLoadTime)
	fmt.Printf("  🚀 성능 개선: %.1fx 더 빠름!\n", speedup)
	
	if speedup > 5 {
		fmt.Println("  🎯 훌륭한 성능! 캐시가 매우 효과적입니다.")
	} else if speedup > 2 {
		fmt.Println("  👍 좋은 성능! 캐시가 효과적입니다.")
	} else {
		fmt.Println("  📝 성능 개선이 있지만 더 최적화할 여지가 있습니다.")
	}
}
package user

import (
	"context"
	"fmt"
	"time"
)

// Example usage of type-safe GenericLazyLoader

// ExampleTypeSafeInventoryLoader demonstrates type-safe inventory loading
func ExampleTypeSafeInventoryLoader() {
	// Create in-memory storage
	storage := NewInMemoryStorage()
	defer storage.Close()

	// Create type-safe inventory loader
	inventoryLoader := NewGenericLazyLoader(GenericLazyLoaderConfig[*UserInventory]{
		CacheStorage:   storage,
		DefaultTTL:     10 * time.Minute,
		CacheKeyPrefix: "inventory",
		LoaderFunc: func(ctx context.Context, userID string) (*UserInventory, error) {
			// Simulate database load
			inventory := NewUserInventory(userID, 100)
			inventory.AddItem(&InventoryItem{
				ID: "sword_001", Name: "Iron Sword", Quantity: 1,
				ItemType: "weapon", Rarity: "common", AcquiredAt: time.Now(),
			})
			return inventory, nil
		},
	})

	ctx := context.Background()
	userID := "user123"

	// Type-safe loading - no casting needed!
	inventory, err := inventoryLoader.Load(ctx, userID)
	if err != nil {
		fmt.Printf("Error loading inventory: %v\n", err)
		return
	}

	// Direct access to typed fields
	fmt.Printf("User %s has %d items in inventory (capacity: %d)\n", 
		inventory.UserID, inventory.UsedSlots, inventory.Capacity)
	
	// Type-safe access to items
	for itemID, item := range inventory.Items {
		fmt.Printf("  Item %s: %s (qty: %d, type: %s)\n", 
			itemID, item.Name, item.Quantity, item.ItemType)
	}
}

// ExampleTypeSafeAchievementsLoader demonstrates type-safe achievements loading
func ExampleTypeSafeAchievementsLoader() {
	storage := NewInMemoryStorage()
	defer storage.Close()

	// Create type-safe achievements loader
	achievementsLoader := NewGenericLazyLoader(GenericLazyLoaderConfig[*UserAchievements]{
		CacheStorage:   storage,
		DefaultTTL:     15 * time.Minute,
		CacheKeyPrefix: "achievements",
		LoaderFunc: func(ctx context.Context, userID string) (*UserAchievements, error) {
			achievements := NewUserAchievements(userID)
			
			// Add sample achievements
			now := time.Now()
			achievements.Achievements["first_login"] = &Achievement{
				ID: "first_login", Name: "First Login", Category: "onboarding",
				Points: 10, Rarity: "common", IsUnlocked: true, UnlockedAt: &now,
			}
			achievements.Achievements["level_5"] = &Achievement{
				ID: "level_5", Name: "Reach Level 5", Category: "progression",
				Points: 25, Rarity: "common", IsUnlocked: false,
			}
			
			achievements.UnlockedCount = 1
			achievements.TotalCount = 2
			achievements.AchievementPoints = 10
			
			return achievements, nil
		},
	})

	ctx := context.Background()
	userID := "user456"

	// Type-safe loading
	achievements, err := achievementsLoader.Load(ctx, userID)
	if err != nil {
		fmt.Printf("Error loading achievements: %v\n", err)
		return
	}

	// Direct typed access
	fmt.Printf("User %s has unlocked %d/%d achievements (%d points)\n",
		achievements.UserID, achievements.UnlockedCount, 
		achievements.TotalCount, achievements.AchievementPoints)

	// Type-safe iteration
	for achievementID, achievement := range achievements.Achievements {
		status := "🔒 Locked"
		if achievement.IsUnlocked {
			status = "✅ Unlocked"
		}
		fmt.Printf("  %s: %s %s (%d points)\n", 
			achievementID, achievement.Name, status, achievement.Points)
	}
}

// ExampleMultiTypeLazyLoader demonstrates the multi-type lazy loader
func ExampleMultiTypeLazyLoader() {
	storage := NewInMemoryStorage()
	defer storage.Close()

	// Create multi-type lazy loader
	multiLoader := NewMultiTypeLazyLoader(MultiTypeLazyLoaderConfig{
		CacheStorage:   storage,
		DefaultTTL:     20 * time.Minute,
		CacheKeyPrefix: "user_data",
	})

	ctx := context.Background()
	userID := "user789"

	fmt.Println("=== Loading individual data types ===")

	// Load inventory (type-safe)
	inventory, err := multiLoader.LoadInventory(ctx, userID)
	if err != nil {
		fmt.Printf("Error loading inventory: %v\n", err)
		return
	}
	fmt.Printf("✅ Inventory loaded: %d items\n", inventory.UsedSlots)

	// Load achievements (type-safe)
	achievements, err := multiLoader.LoadAchievements(ctx, userID)
	if err != nil {
		fmt.Printf("Error loading achievements: %v\n", err)
		return
	}
	fmt.Printf("✅ Achievements loaded: %d points\n", achievements.AchievementPoints)

	// Load stats (type-safe)
	stats, err := multiLoader.LoadStats(ctx, userID)
	if err != nil {
		fmt.Printf("Error loading stats: %v\n", err)
		return
	}
	fmt.Printf("✅ Stats loaded: level %v\n", stats.GlobalStats["level"])

	fmt.Println("\n=== Loading all data concurrently ===")

	// Load all data types concurrently
	userData, err := multiLoader.LoadAll(ctx, userID)
	if err != nil {
		fmt.Printf("Error loading all user data: %v\n", err)
		return
	}

	fmt.Printf("🎯 All data loaded for user %s at %v\n", 
		userData.UserID, userData.LoadedAt.Format("15:04:05"))
	fmt.Printf("  Inventory: %d/%d slots used\n", 
		userData.Inventory.UsedSlots, userData.Inventory.Capacity)
	fmt.Printf("  Achievements: %d points from %d unlocked\n", 
		userData.Achievements.AchievementPoints, userData.Achievements.UnlockedCount)
	fmt.Printf("  Stats: %d game types tracked\n", 
		len(userData.Stats.GameStats))
	fmt.Printf("  Preferences: %s theme, %s language\n", 
		userData.Preferences.Theme, userData.Preferences.Language)
	fmt.Printf("  Social: %d friends, %d guilds\n", 
		len(userData.SocialData.Friends), len(userData.SocialData.Guilds))
}

// ExampleBatchLoading demonstrates batch loading with type safety
func ExampleBatchLoading() {
	storage := NewInMemoryStorage()
	defer storage.Close()

	// Create type-safe inventory loader
	inventoryLoader := NewGenericLazyLoader(GenericLazyLoaderConfig[*UserInventory]{
		CacheStorage:   storage,
		DefaultTTL:     5 * time.Minute,
		CacheKeyPrefix: "batch_inventory",
		LoaderFunc: func(ctx context.Context, userID string) (*UserInventory, error) {
			inventory := NewUserInventory(userID, 50)
			
			// Add different items based on user ID
			switch userID {
			case "user1":
				inventory.AddItem(&InventoryItem{
					ID: "sword", Name: "Steel Sword", Quantity: 1,
					ItemType: "weapon", Rarity: "uncommon", AcquiredAt: time.Now(),
				})
			case "user2":
				inventory.AddItem(&InventoryItem{
					ID: "bow", Name: "Elven Bow", Quantity: 1,
					ItemType: "weapon", Rarity: "rare", AcquiredAt: time.Now(),
				})
			case "user3":
				inventory.AddItem(&InventoryItem{
					ID: "staff", Name: "Magic Staff", Quantity: 1,
					ItemType: "weapon", Rarity: "epic", AcquiredAt: time.Now(),
				})
			}
			
			return inventory, nil
		},
	})

	ctx := context.Background()
	userIDs := []string{"user1", "user2", "user3"}

	// Batch load with type safety
	inventories, err := inventoryLoader.LoadMultiple(ctx, userIDs)
	if err != nil {
		fmt.Printf("Error batch loading inventories: %v\n", err)
		return
	}

	fmt.Println("=== Batch loaded inventories ===")
	for userID, inventory := range inventories {
		fmt.Printf("User %s inventory:\n", userID)
		for itemID, item := range inventory.Items {
			fmt.Printf("  %s: %s (%s rarity)\n", itemID, item.Name, item.Rarity)
		}
	}
}

// ExampleCacheManagement demonstrates cache management with type safety
func ExampleCacheManagement() {
	storage := NewInMemoryStorage()
	defer storage.Close()

	// Create type-safe stats loader
	statsLoader := NewGenericLazyLoader(GenericLazyLoaderConfig[*UserStats]{
		CacheStorage:   storage,
		DefaultTTL:     1 * time.Minute, // Short TTL for demo
		CacheKeyPrefix: "cache_demo_stats",
		LoaderFunc: func(ctx context.Context, userID string) (*UserStats, error) {
			fmt.Printf("📀 Loading stats from 'database' for user %s\n", userID)
			
			stats := NewUserStats(userID)
			stats.GlobalStats["level"] = 10
			stats.GlobalStats["experience"] = 5000
			stats.GlobalStats["last_login"] = time.Now().Format(time.RFC3339)
			
			return stats, nil
		},
	})

	ctx := context.Background()
	userID := "cache_user"

	// First load - from database
	fmt.Println("=== First load (from database) ===")
	stats1, err := statsLoader.Load(ctx, userID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("✅ Loaded stats: level %v, exp %v\n", 
		stats1.GlobalStats["level"], stats1.GlobalStats["experience"])

	// Second load - from cache
	fmt.Println("\n=== Second load (from cache) ===")
	stats2, err := statsLoader.Load(ctx, userID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("⚡ Cached stats: level %v, exp %v\n", 
		stats2.GlobalStats["level"], stats2.GlobalStats["experience"])

	// Check cache info
	cacheInfo, err := statsLoader.GetCacheInfo(ctx, userID)
	if err != nil {
		fmt.Printf("Error getting cache info: %v\n", err)
		return
	}
	fmt.Printf("📊 Cache info: exists=%v, TTL=%v\n", 
		cacheInfo.Exists, cacheInfo.TTL)

	// Invalidate cache
	fmt.Println("\n=== Invalidating cache ===")
	if err := statsLoader.InvalidateCache(ctx, userID); err != nil {
		fmt.Printf("Error invalidating cache: %v\n", err)
		return
	}
	fmt.Println("🗑️ Cache invalidated")

	// Third load - from database again
	fmt.Println("\n=== Third load (from database after invalidation) ===")
	stats3, err := statsLoader.Load(ctx, userID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("🔄 Reloaded stats: level %v, exp %v\n", 
		stats3.GlobalStats["level"], stats3.GlobalStats["experience"])
}

// ExamplePreloading demonstrates preloading data
func ExamplePreloading() {
	storage := NewInMemoryStorage()
	defer storage.Close()

	multiLoader := NewMultiTypeLazyLoader(MultiTypeLazyLoaderConfig{
		CacheStorage:   storage,
		DefaultTTL:     30 * time.Minute,
		CacheKeyPrefix: "preload_demo",
	})

	ctx := context.Background()
	userIDs := []string{"preload_user1", "preload_user2", "preload_user3"}

	fmt.Println("=== Preloading user data ===")
	start := time.Now()

	// Preload inventory for multiple users
	fmt.Println("🔄 Preloading inventories...")
	for _, userID := range userIDs {
		if err := multiLoader.inventoryLoader.PreloadData(ctx, userID); err != nil {
			fmt.Printf("Error preloading inventory for %s: %v\n", userID, err)
			continue
		}
		fmt.Printf("  ✅ Preloaded inventory for %s\n", userID)
	}

	preloadTime := time.Since(start)
	fmt.Printf("⏱️ Preloading completed in %v\n", preloadTime)

	fmt.Println("\n=== Fast access to preloaded data ===")
	accessStart := time.Now()

	for _, userID := range userIDs {
		inventory, err := multiLoader.LoadInventory(ctx, userID)
		if err != nil {
			fmt.Printf("Error accessing inventory for %s: %v\n", userID, err)
			continue
		}
		fmt.Printf("⚡ Fast access to %s inventory: %d items\n", 
			userID, inventory.UsedSlots)
	}

	accessTime := time.Since(accessStart)
	fmt.Printf("⏱️ Fast access completed in %v\n", accessTime)
	fmt.Printf("🚀 Speedup: %.2fx faster\n", 
		float64(preloadTime)/float64(accessTime))
}

// ExampleErrorHandling demonstrates error handling with type safety
func ExampleErrorHandling() {
	storage := NewInMemoryStorage()
	defer storage.Close()

	// Create loader that simulates errors
	errorLoader := NewGenericLazyLoader(GenericLazyLoaderConfig[*UserInventory]{
		CacheStorage:   storage,
		DefaultTTL:     10 * time.Minute,
		CacheKeyPrefix: "error_demo",
		LoaderFunc: func(ctx context.Context, userID string) (*UserInventory, error) {
			if userID == "error_user" {
				return nil, fmt.Errorf("simulated database error for user %s", userID)
			}
			if userID == "timeout_user" {
				return nil, context.DeadlineExceeded
			}
			
			// Normal case
			return NewUserInventory(userID, 100), nil
		},
	})

	ctx := context.Background()

	fmt.Println("=== Error handling demonstration ===")

	// Test normal case
	fmt.Println("🟢 Normal case:")
	inventory, err := errorLoader.Load(ctx, "normal_user")
	if err != nil {
		fmt.Printf("  ❌ Unexpected error: %v\n", err)
	} else {
		fmt.Printf("  ✅ Successfully loaded inventory for user: %s\n", inventory.UserID)
	}

	// Test error case
	fmt.Println("\n🔴 Error case:")
	inventory, err = errorLoader.Load(ctx, "error_user")
	if err != nil {
		fmt.Printf("  ❌ Expected error: %v\n", err)
		// inventory is zero value (*UserInventory = nil)
		if inventory == nil {
			fmt.Printf("  ✅ Inventory is properly nil on error\n")
		}
	} else {
		fmt.Printf("  ❌ Unexpected success\n")
	}

	// Test timeout case
	fmt.Println("\n⏰ Timeout case:")
	inventory, err = errorLoader.Load(ctx, "timeout_user")
	if err != nil {
		fmt.Printf("  ❌ Timeout error: %v\n", err)
		if inventory == nil {
			fmt.Printf("  ✅ Inventory is properly nil on timeout\n")
		}
	} else {
		fmt.Printf("  ❌ Unexpected success\n")
	}

	// Test batch loading with mixed results
	fmt.Println("\n📦 Batch loading with mixed results:")
	inventories, err := errorLoader.LoadMultiple(ctx, []string{"normal_user", "error_user"})
	if err != nil {
		fmt.Printf("  ❌ Batch load failed: %v\n", err)
		fmt.Printf("  ✅ Proper error propagation in batch loading\n")
	} else {
		fmt.Printf("  ❌ Unexpected batch success: %v\n", inventories)
	}
}
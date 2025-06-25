package user

import (
	"context"
	"testing"
	"time"
)

func TestGenericLazyLoader_TypeSafety(t *testing.T) {
	storage := NewInMemoryStorage()
	defer storage.Close()

	// Create type-safe inventory loader
	inventoryLoader := NewGenericLazyLoader(GenericLazyLoaderConfig[*UserInventory]{
		CacheStorage:   storage,
		DefaultTTL:     5 * time.Minute,
		CacheKeyPrefix: "test_inventory",
		LoaderFunc: func(ctx context.Context, userID string) (*UserInventory, error) {
			inventory := NewUserInventory(userID, 100)
			inventory.AddItem(&InventoryItem{
				ID: "test_item", Name: "Test Item", Quantity: 1,
				ItemType: "test", Rarity: "common", AcquiredAt: time.Now(),
			})
			return inventory, nil
		},
	})

	ctx := context.Background()
	userID := "test_user"

	// Load inventory
	inventory, err := inventoryLoader.Load(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to load inventory: %v", err)
	}

	// Type-safe assertions
	if inventory.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, inventory.UserID)
	}

	if inventory.UsedSlots != 1 {
		t.Errorf("Expected 1 used slot, got %d", inventory.UsedSlots)
	}

	// Check item exists
	item, exists := inventory.GetItem("test_item")
	if !exists {
		t.Error("Expected test_item to exist in inventory")
	}

	if item.Name != "Test Item" {
		t.Errorf("Expected item name 'Test Item', got '%s'", item.Name)
	}
}

func TestGenericLazyLoader_Caching(t *testing.T) {
	storage := NewInMemoryStorage()
	defer storage.Close()

	loadCount := 0
	
	// Create loader that counts calls
	statsLoader := NewGenericLazyLoader(GenericLazyLoaderConfig[*UserStats]{
		CacheStorage:   storage,
		DefaultTTL:     10 * time.Minute,
		CacheKeyPrefix: "test_stats",
		LoaderFunc: func(ctx context.Context, userID string) (*UserStats, error) {
			loadCount++
			stats := NewUserStats(userID)
			stats.GlobalStats["level"] = 10
			return stats, nil
		},
	})

	ctx := context.Background()
	userID := "cache_test_user"

	// First load - should call loader
	stats1, err := statsLoader.Load(ctx, userID)
	if err != nil {
		t.Fatalf("First load failed: %v", err)
	}

	if loadCount != 1 {
		t.Errorf("Expected 1 loader call, got %d", loadCount)
	}

	// Second load - should use cache
	stats2, err := statsLoader.Load(ctx, userID)
	if err != nil {
		t.Fatalf("Second load failed: %v", err)
	}

	if loadCount != 1 {
		t.Errorf("Expected 1 loader call after cache hit, got %d", loadCount)
	}

	// Verify both loads return same data
	if stats1.UserID != stats2.UserID {
		t.Error("Cache returned different UserID")
	}

	// JSON serialization/deserialization converts int to float64
	level1, ok1 := stats1.GlobalStats["level"].(int)
	level2, ok2 := stats2.GlobalStats["level"].(float64)
	if ok1 && ok2 {
		if float64(level1) != level2 {
			t.Errorf("Cache returned different level: %v vs %v", level1, level2)
		}
	} else {
		t.Errorf("Unexpected level types: %T vs %T", stats1.GlobalStats["level"], stats2.GlobalStats["level"])
	}
}

func TestGenericLazyLoader_BatchLoading(t *testing.T) {
	storage := NewInMemoryStorage()
	defer storage.Close()

	achievementsLoader := NewGenericLazyLoader(GenericLazyLoaderConfig[*UserAchievements]{
		CacheStorage:   storage,
		DefaultTTL:     5 * time.Minute,
		CacheKeyPrefix: "test_achievements",
		LoaderFunc: func(ctx context.Context, userID string) (*UserAchievements, error) {
			achievements := NewUserAchievements(userID)
			achievements.AchievementPoints = len(userID) * 10 // Different points per user
			return achievements, nil
		},
	})

	ctx := context.Background()
	userIDs := []string{"user1", "user2", "user3"}

	// Batch load
	results, err := achievementsLoader.LoadMultiple(ctx, userIDs)
	if err != nil {
		t.Fatalf("Batch load failed: %v", err)
	}

	// Verify all users loaded
	if len(results) != len(userIDs) {
		t.Errorf("Expected %d results, got %d", len(userIDs), len(results))
	}

	// Verify type-safe access to results
	for _, userID := range userIDs {
		achievements, exists := results[userID]
		if !exists {
			t.Errorf("Missing results for user %s", userID)
			continue
		}

		if achievements.UserID != userID {
			t.Errorf("Expected UserID %s, got %s", userID, achievements.UserID)
		}

		expectedPoints := len(userID) * 10
		if achievements.AchievementPoints != expectedPoints {
			t.Errorf("Expected %d points for %s, got %d", 
				expectedPoints, userID, achievements.AchievementPoints)
		}
	}
}

func TestMultiTypeLazyLoader_ConcurrentLoading(t *testing.T) {
	storage := NewInMemoryStorage()
	defer storage.Close()

	multiLoader := NewMultiTypeLazyLoader(MultiTypeLazyLoaderConfig{
		CacheStorage:   storage,
		DefaultTTL:     10 * time.Minute,
		CacheKeyPrefix: "test_multi",
	})

	ctx := context.Background()
	userID := "concurrent_test_user"

	// Load all data concurrently
	userData, err := multiLoader.LoadAll(ctx, userID)
	if err != nil {
		t.Fatalf("Concurrent load failed: %v", err)
	}

	// Verify all data types are loaded
	if userData.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, userData.UserID)
	}

	if userData.Inventory == nil {
		t.Error("Inventory not loaded")
	}

	if userData.Achievements == nil {
		t.Error("Achievements not loaded")
	}

	if userData.Stats == nil {
		t.Error("Stats not loaded")
	}

	if userData.Preferences == nil {
		t.Error("Preferences not loaded")
	}

	if userData.SocialData == nil {
		t.Error("Social data not loaded")
	}

	// Verify LoadedAt timestamp
	if userData.LoadedAt.IsZero() {
		t.Error("LoadedAt timestamp not set")
	}

	// Verify type-safe access
	if userData.Inventory.UserID != userID {
		t.Error("Inventory UserID mismatch")
	}

	if userData.Achievements.UserID != userID {
		t.Error("Achievements UserID mismatch")
	}
}

func TestGenericLazyLoader_CacheInvalidation(t *testing.T) {
	storage := NewInMemoryStorage()
	defer storage.Close()

	loadCount := 0
	
	preferencesLoader := NewGenericLazyLoader(GenericLazyLoaderConfig[*UserPreferences]{
		CacheStorage:   storage,
		DefaultTTL:     10 * time.Minute,
		CacheKeyPrefix: "test_preferences",
		LoaderFunc: func(ctx context.Context, userID string) (*UserPreferences, error) {
			loadCount++
			prefs := NewUserPreferences(userID)
			prefs.Theme = "dynamic" // Different each load
			return prefs, nil
		},
	})

	ctx := context.Background()
	userID := "invalidation_test_user"

	// First load
	prefs1, err := preferencesLoader.Load(ctx, userID)
	if err != nil {
		t.Fatalf("First load failed: %v", err)
	}

	if loadCount != 1 {
		t.Errorf("Expected 1 load, got %d", loadCount)
	}

	// Verify cache exists
	cacheInfo, err := preferencesLoader.GetCacheInfo(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get cache info: %v", err)
	}

	if !cacheInfo.Exists {
		t.Error("Cache should exist after first load")
	}

	// Invalidate cache
	err = preferencesLoader.InvalidateCache(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to invalidate cache: %v", err)
	}

	// Verify cache no longer exists
	cacheInfo, err = preferencesLoader.GetCacheInfo(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get cache info after invalidation: %v", err)
	}

	if cacheInfo.Exists {
		t.Error("Cache should not exist after invalidation")
	}

	// Second load should call loader again
	prefs2, err := preferencesLoader.Load(ctx, userID)
	if err != nil {
		t.Fatalf("Second load failed: %v", err)
	}

	if loadCount != 2 {
		t.Errorf("Expected 2 loads after invalidation, got %d", loadCount)
	}

	// Verify both results are valid
	if prefs1.UserID != userID || prefs2.UserID != userID {
		t.Error("UserID mismatch in preference loads")
	}
}

func TestMultiTypeLazyLoader_CacheInvalidation(t *testing.T) {
	storage := NewInMemoryStorage()
	defer storage.Close()

	multiLoader := NewMultiTypeLazyLoader(MultiTypeLazyLoaderConfig{
		CacheStorage:   storage,
		DefaultTTL:     5 * time.Minute,
		CacheKeyPrefix: "test_multi_invalidate",
	})

	ctx := context.Background()
	userID := "multi_invalidate_user"

	// Load some data to populate cache
	_, err := multiLoader.LoadInventory(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to load inventory: %v", err)
	}

	_, err = multiLoader.LoadStats(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to load stats: %v", err)
	}

	// Verify cache exists for inventory
	invCacheInfo, err := multiLoader.inventoryLoader.GetCacheInfo(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get inventory cache info: %v", err)
	}

	if !invCacheInfo.Exists {
		t.Error("Inventory cache should exist")
	}

	// Invalidate all user cache
	err = multiLoader.InvalidateUserCache(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to invalidate user cache: %v", err)
	}

	// Verify cache no longer exists
	invCacheInfo, err = multiLoader.inventoryLoader.GetCacheInfo(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get inventory cache info after invalidation: %v", err)
	}

	if invCacheInfo.Exists {
		t.Error("Inventory cache should not exist after invalidation")
	}
}
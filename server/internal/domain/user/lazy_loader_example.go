package user

import (
	"context"
	"fmt"
	"time"
)

// Example usage of LazyLoader with different storage backends

// ExampleWithRedis demonstrates using LazyLoader with Redis
func ExampleWithRedis() {
	// Note: This example assumes Redis client is available
	// For actual usage, you would create Redis client like:
	// rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	// redisStorage := NewRedisStorage(rdb)
	
	// For this example, we'll use in-memory storage instead
	memStorage := NewInMemoryStorage()
	defer memStorage.Close()

	// Create lazy loader with storage
	config := LazyLoaderConfig{
		CacheStorage:   memStorage,
		DefaultTTL:     15 * time.Minute,
		CacheKeyPrefix: "user_data",
	}
	
	loader := NewLazyLoader(config)
	
	// Register loader functions
	registerUserLoaders(loader)
	
	// Usage example
	ctx := context.Background()
	userID := "user123"
	
	inventory, err := loader.Load(ctx, "inventory", userID)
	if err != nil {
		fmt.Printf("Failed to load inventory: %v\n", err)
		return
	}
	
	fmt.Printf("Loaded inventory: %+v\n", inventory)
}

// ExampleWithInMemory demonstrates using LazyLoader with in-memory storage
func ExampleWithInMemory() {
	// Create in-memory storage
	memStorage := NewInMemoryStorage()
	defer memStorage.Close()

	// Create lazy loader with in-memory storage
	config := LazyLoaderConfig{
		CacheStorage:   memStorage,
		DefaultTTL:     10 * time.Minute,
		CacheKeyPrefix: "user_data",
	}
	
	loader := NewLazyLoader(config)
	
	// Register loader functions
	registerUserLoaders(loader)
	
	// Usage example
	ctx := context.Background()
	userID := "user456"
	
	// Load multiple data types
	results, err := loader.LoadMultiple(ctx, "batch", []string{userID})
	if err != nil {
		fmt.Printf("Failed to load batch data: %v\n", err)
		return
	}
	
	for dataType, data := range results {
		fmt.Printf("Loaded %s: %+v\n", dataType, data)
	}
}

// ExampleWithComposite demonstrates using LazyLoader with composite storage
func ExampleWithComposite() {
	// Setup primary storage (in-memory for this example)
	primaryStorage := NewInMemoryStorage()
	defer primaryStorage.Close()
	
	// Setup fallback storage
	fallbackStorage := NewInMemoryStorage()
	defer fallbackStorage.Close()
	
	// Create composite storage (primary with fallback)
	compositeStorage := NewCompositeStorage(primaryStorage, fallbackStorage, true)

	// Create lazy loader with composite storage
	config := LazyLoaderConfig{
		CacheStorage:   compositeStorage,
		DefaultTTL:     20 * time.Minute,
		CacheKeyPrefix: "user_composite",
	}
	
	loader := NewLazyLoader(config)
	
	// Register loader functions
	registerUserLoaders(loader)
	
	// Usage example
	ctx := context.Background()
	userID := "user789"
	
	// This will try primary first, then fall back to secondary
	preferences, err := loader.Load(ctx, "preferences", userID)
	if err != nil {
		fmt.Printf("Failed to load preferences: %v\n", err)
		return
	}
	
	fmt.Printf("Loaded preferences from composite storage: %+v\n", preferences)
}

// ExampleWithNoCache demonstrates using LazyLoader without caching
func ExampleWithNoCache() {
	// Create no-op storage (no caching)
	noopStorage := NewNoOpStorage()

	// Create lazy loader without caching
	config := LazyLoaderConfig{
		CacheStorage:   noopStorage,
		DefaultTTL:     0, // Irrelevant for no-op
		CacheKeyPrefix: "user_nocache",
	}
	
	loader := NewLazyLoader(config)
	
	// Register loader functions
	registerUserLoaders(loader)
	
	// Usage example - will always load from source
	ctx := context.Background()
	userID := "user999"
	
	socialData, err := loader.Load(ctx, "social", userID)
	if err != nil {
		fmt.Printf("Failed to load social data: %v\n", err)
		return
	}
	
	fmt.Printf("Loaded social data (no cache): %+v\n", socialData)
}

// registerUserLoaders registers all the loader functions for user data
func registerUserLoaders(loader *LazyLoader) {
	// Register inventory loader
	loader.RegisterLoader("inventory", func(ctx context.Context, key string) (any, error) {
		// Simulate loading from database
		inventory := NewUserInventory(key, 100)
		
		// Add some sample items
		inventory.AddItem(&InventoryItem{
			ID: "sword1", Name: "Iron Sword", Quantity: 1,
			ItemType: "weapon", Rarity: "common",
			AcquiredAt: time.Now(),
		})
		inventory.AddItem(&InventoryItem{
			ID: "potion1", Name: "Health Potion", Quantity: 5,
			ItemType: "consumable", Rarity: "common",
			AcquiredAt: time.Now(),
		})
		
		return inventory, nil
	})
	
	// Register achievements loader
	loader.RegisterLoader("achievements", func(ctx context.Context, key string) (any, error) {
		// Simulate loading from database
		achievements := NewUserAchievements(key)
		
		// Add some sample achievements
		now := time.Now()
		achievements.Achievements["first_kill"] = &Achievement{
			ID: "first_kill", Name: "First Kill", Category: "combat",
			Points: 10, Rarity: "common", IsUnlocked: true,
			UnlockedAt: &now,
		}
		achievements.Achievements["level_10"] = &Achievement{
			ID: "level_10", Name: "Reach Level 10", Category: "progression",
			Points: 50, Rarity: "uncommon", IsUnlocked: true,
			UnlockedAt: &now,
		}
		achievements.UnlockedCount = 2
		achievements.TotalCount = 10
		achievements.AchievementPoints = 60
		
		return achievements, nil
	})
	
	// Register stats loader
	loader.RegisterLoader("stats", func(ctx context.Context, key string) (any, error) {
		// Simulate loading from database
		stats := NewUserStats(key)
		
		// Add sample game stats
		towerDefenseStats := &GameStat{
			GameType:     "tower_defense",
			GamesPlayed:  42,
			GamesWon:     25,
			GamesLost:    17,
			WinRate:      0.595,
			TotalScore:   156789,
			HighestScore: 25000,
			AverageScore: 3733.07,
			Rank:         1200,
			Rating:       1850,
			CustomStats:  make(map[string]any),
		}
		stats.GameStats["tower_defense"] = towerDefenseStats
		
		// Add global stats
		stats.GlobalStats["total_playtime"] = "120h 30m"
		stats.GlobalStats["level"] = 15
		stats.GlobalStats["experience"] = 12500
		
		return stats, nil
	})
	
	// Register preferences loader
	loader.RegisterLoader("preferences", func(ctx context.Context, key string) (any, error) {
		// Simulate loading from database
		preferences := NewUserPreferences(key)
		
		// Customize some preferences
		preferences.Theme = "dark"
		preferences.Language = "en"
		preferences.GamePreferences["sound_volume"] = 0.8
		preferences.GamePreferences["music_volume"] = 0.6
		preferences.UIPreferences["auto_save"] = true
		preferences.UIPreferences["show_tooltips"] = true
		
		return preferences, nil
	})
	
	// Register social data loader
	loader.RegisterLoader("social", func(ctx context.Context, key string) (any, error) {
		// Simulate loading from database
		socialData := NewUserSocialData(key)
		
		// Add some friends
		socialData.AddFriend("friend1", "Alice")
		socialData.AddFriend("friend2", "Bob")
		socialData.AddFriend("friend3", "Charlie")
		
		// Accept some friend requests
		socialData.AcceptFriend("friend1")
		socialData.AcceptFriend("friend2")
		
		// Join some guilds
		socialData.Guilds = append(socialData.Guilds, "guild123", "guild456")
		
		// Add social links
		socialData.SocialLinks["discord"] = "user#1234"
		socialData.SocialLinks["steam"] = "steamuser123"
		
		return socialData, nil
	})
	
	// Register batch loader for multiple data types
	loader.RegisterLoader("batch", func(ctx context.Context, key string) (any, error) {
		// This could load multiple data types efficiently in one query
		batchInventory := NewUserInventory(key, 50)
		batchInventory.AddItem(&InventoryItem{
			ID: "batch_item", Name: "Batch Item", Quantity: 1,
			ItemType: "misc", Rarity: "common", AcquiredAt: time.Now(),
		})
		
		batchAchievements := NewUserAchievements(key)
		now := time.Now()
		batchAchievements.Achievements["batch_achievement"] = &Achievement{
			ID: "batch_achievement", Name: "Batch Achievement",
			Category: "misc", Points: 5, Rarity: "common",
			IsUnlocked: true, UnlockedAt: &now,
		}
		
		batchStats := NewUserStats(key)
		batchStats.GlobalStats["level"] = 1
		batchStats.GlobalStats["experience"] = 0
		
		return map[string]any{
			"inventory":    batchInventory,
			"achievements": batchAchievements,
			"stats":        batchStats,
		}, nil
	})
}

// Performance monitoring example
func ExampleWithPerformanceMonitoring() {
	// Setup storage with metrics
	memStorage := NewInMemoryStorage()
	defer memStorage.Close()
	
	// Wrap storage with performance monitoring
	monitoredStorage := &PerformanceMonitoredStorage{
		storage: memStorage,
		metrics: make(map[string]*StorageMetrics),
	}

	config := LazyLoaderConfig{
		CacheStorage:   monitoredStorage,
		DefaultTTL:     5 * time.Minute,
		CacheKeyPrefix: "user_monitored",
	}
	
	loader := NewLazyLoader(config)
	registerUserLoaders(loader)
	
	ctx := context.Background()
	userID := "monitored_user"
	
	// Load data and measure performance
	start := time.Now()
	
	for i := range 100 {
		_, err := loader.Load(ctx, "inventory", fmt.Sprintf("%s_%d", userID, i))
		if err != nil {
			fmt.Printf("Error loading inventory: %v\n", err)
		}
	}
	
	elapsed := time.Since(start)
	fmt.Printf("Loaded 100 inventories in %v\n", elapsed)
	
	// Print metrics
	monitoredStorage.PrintMetrics()
}

// PerformanceMonitoredStorage wraps another storage with performance monitoring
type PerformanceMonitoredStorage struct {
	storage CacheStorage
	metrics map[string]*StorageMetrics
}

type StorageMetrics struct {
	Operations    int64
	TotalDuration time.Duration
	Errors        int64
}

func (p *PerformanceMonitoredStorage) Get(ctx context.Context, key string) ([]byte, error) {
	start := time.Now()
	result, err := p.storage.Get(ctx, key)
	p.recordMetric("Get", time.Since(start), err)
	return result, err
}

func (p *PerformanceMonitoredStorage) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	start := time.Now()
	err := p.storage.Set(ctx, key, value, ttl)
	p.recordMetric("Set", time.Since(start), err)
	return err
}

func (p *PerformanceMonitoredStorage) Delete(ctx context.Context, key string) error {
	start := time.Now()
	err := p.storage.Delete(ctx, key)
	p.recordMetric("Delete", time.Since(start), err)
	return err
}

func (p *PerformanceMonitoredStorage) DeletePattern(ctx context.Context, pattern string) error {
	start := time.Now()
	err := p.storage.DeletePattern(ctx, pattern)
	p.recordMetric("DeletePattern", time.Since(start), err)
	return err
}

func (p *PerformanceMonitoredStorage) Exists(ctx context.Context, key string) (bool, error) {
	start := time.Now()
	result, err := p.storage.Exists(ctx, key)
	p.recordMetric("Exists", time.Since(start), err)
	return result, err
}

func (p *PerformanceMonitoredStorage) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	start := time.Now()
	result, err := p.storage.GetTTL(ctx, key)
	p.recordMetric("GetTTL", time.Since(start), err)
	return result, err
}

func (p *PerformanceMonitoredStorage) recordMetric(operation string, duration time.Duration, err error) {
	if p.metrics[operation] == nil {
		p.metrics[operation] = &StorageMetrics{}
	}
	
	metric := p.metrics[operation]
	metric.Operations++
	metric.TotalDuration += duration
	
	if err != nil {
		metric.Errors++
	}
}

func (p *PerformanceMonitoredStorage) PrintMetrics() {
	fmt.Println("Storage Performance Metrics:")
	for operation, metric := range p.metrics {
		avgDuration := metric.TotalDuration / time.Duration(metric.Operations)
		errorRate := float64(metric.Errors) / float64(metric.Operations) * 100
		
		fmt.Printf("  %s: %d ops, avg: %v, errors: %.2f%%\n", 
			operation, metric.Operations, avgDuration, errorRate)
	}
}
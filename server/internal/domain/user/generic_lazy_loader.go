package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// GenericLazyLoader is a type-safe lazy loader using generics
type GenericLazyLoader[T any] struct {
	cacheStorage   CacheStorage
	defaultTTL     time.Duration
	cacheKeyPrefix string
	loaderFunc     GenericLoaderFunc[T]
	mu             sync.RWMutex
}

// GenericLoaderFunc is a type-safe function type that loads data for a specific key
type GenericLoaderFunc[T any] func(ctx context.Context, key string) (T, error)

// GenericLazyLoaderConfig contains configuration for GenericLazyLoader
type GenericLazyLoaderConfig[T any] struct {
	CacheStorage   CacheStorage
	DefaultTTL     time.Duration
	CacheKeyPrefix string
	LoaderFunc     GenericLoaderFunc[T]
}

// NewGenericLazyLoader creates a new type-safe GenericLazyLoader
func NewGenericLazyLoader[T any](config GenericLazyLoaderConfig[T]) *GenericLazyLoader[T] {
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 15 * time.Minute
	}
	if config.CacheKeyPrefix == "" {
		config.CacheKeyPrefix = "generic_lazy_load"
	}

	return &GenericLazyLoader[T]{
		cacheStorage:   config.CacheStorage,
		defaultTTL:     config.DefaultTTL,
		cacheKeyPrefix: config.CacheKeyPrefix,
		loaderFunc:     config.LoaderFunc,
	}
}

// Load loads data for a specific key with type safety
func (l *GenericLazyLoader[T]) Load(ctx context.Context, key string) (T, error) {
	var zero T
	
	// Try to get from cache first
	if cached, err := l.getFromCache(ctx, key); err == nil {
		return cached, nil
	}

	// Load data using the registered loader
	if l.loaderFunc == nil {
		return zero, errors.New("no loader function registered")
	}

	data, err := l.loaderFunc(ctx, key)
	if err != nil {
		return zero, fmt.Errorf("failed to load data: %w", err)
	}

	// Cache the loaded data
	if err := l.setToCache(ctx, key, data); err != nil {
		// Log the error but don't fail the operation
		fmt.Printf("Failed to cache data for key %s: %v\n", key, err)
	}

	return data, nil
}

// LoadMultiple loads multiple keys with type safety
func (l *GenericLazyLoader[T]) LoadMultiple(ctx context.Context, keys []string) (map[string]T, error) {
	results := make(map[string]T)
	uncachedKeys := make([]string, 0)

	// First, try to get all keys from cache
	for _, key := range keys {
		if cached, err := l.getFromCache(ctx, key); err == nil {
			results[key] = cached
		} else {
			uncachedKeys = append(uncachedKeys, key)
		}
	}

	// If all keys were cached, return early
	if len(uncachedKeys) == 0 {
		return results, nil
	}

	// Load uncached data
	if l.loaderFunc == nil {
		return nil, errors.New("no loader function registered")
	}

	for _, key := range uncachedKeys {
		data, err := l.loaderFunc(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("failed to load data for key %s: %w", key, err)
		}

		results[key] = data

		// Cache the loaded data
		if err := l.setToCache(ctx, key, data); err != nil {
			fmt.Printf("Failed to cache data for key %s: %v\n", key, err)
		}
	}

	return results, nil
}

// InvalidateCache removes cached data for a specific key
func (l *GenericLazyLoader[T]) InvalidateCache(ctx context.Context, key string) error {
	cacheKey := l.buildCacheKey(key)
	return l.cacheStorage.Delete(ctx, cacheKey)
}

// InvalidateCachePattern removes cached data matching a pattern
func (l *GenericLazyLoader[T]) InvalidateCachePattern(ctx context.Context, pattern string) error {
	cacheKeyPattern := l.buildCacheKey(pattern)
	return l.cacheStorage.DeletePattern(ctx, cacheKeyPattern)
}

// PreloadData preloads data for a specific key
func (l *GenericLazyLoader[T]) PreloadData(ctx context.Context, key string) error {
	_, err := l.Load(ctx, key)
	return err
}

// PreloadMultipleData preloads data for multiple keys
func (l *GenericLazyLoader[T]) PreloadMultipleData(ctx context.Context, keys []string) error {
	_, err := l.LoadMultiple(ctx, keys)
	return err
}

// GetCacheInfo returns information about cached data
func (l *GenericLazyLoader[T]) GetCacheInfo(ctx context.Context, key string) (*CacheInfo, error) {
	cacheKey := l.buildCacheKey(key)
	
	// Check if key exists
	exists, err := l.cacheStorage.Exists(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	if !exists {
		return &CacheInfo{
			Key:    key,
			Exists: false,
		}, nil
	}

	// Get TTL
	ttl, err := l.cacheStorage.GetTTL(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	return &CacheInfo{
		Key:    key,
		Exists: true,
		TTL:    ttl,
	}, nil
}

// buildCacheKey builds a cache key for the given key
func (l *GenericLazyLoader[T]) buildCacheKey(key string) string {
	return fmt.Sprintf("%s:%s", l.cacheKeyPrefix, key)
}

// getFromCache retrieves data from cache storage with type safety
func (l *GenericLazyLoader[T]) getFromCache(ctx context.Context, key string) (T, error) {
	var zero T
	cacheKey := l.buildCacheKey(key)
	
	data, err := l.cacheStorage.Get(ctx, cacheKey)
	if err != nil {
		return zero, err
	}

	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return zero, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	return result, nil
}

// setToCache stores data in cache storage with type safety
func (l *GenericLazyLoader[T]) setToCache(ctx context.Context, key string, data T) error {
	cacheKey := l.buildCacheKey(key)
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data for caching: %w", err)
	}

	return l.cacheStorage.Set(ctx, cacheKey, jsonData, l.defaultTTL)
}

// Multi-type lazy loader for handling different data types
type MultiTypeLazyLoader struct {
	inventoryLoader    *GenericLazyLoader[*UserInventory]
	achievementsLoader *GenericLazyLoader[*UserAchievements]
	statsLoader        *GenericLazyLoader[*UserStats]
	preferencesLoader  *GenericLazyLoader[*UserPreferences]
	socialLoader       *GenericLazyLoader[*UserSocialData]
}

// MultiTypeLazyLoaderConfig contains configuration for MultiTypeLazyLoader
type MultiTypeLazyLoaderConfig struct {
	CacheStorage   CacheStorage
	DefaultTTL     time.Duration
	CacheKeyPrefix string
}

// NewMultiTypeLazyLoader creates a new multi-type lazy loader
func NewMultiTypeLazyLoader(config MultiTypeLazyLoaderConfig) *MultiTypeLazyLoader {
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 15 * time.Minute
	}
	
	return &MultiTypeLazyLoader{
		inventoryLoader: NewGenericLazyLoader(GenericLazyLoaderConfig[*UserInventory]{
			CacheStorage:   config.CacheStorage,
			DefaultTTL:     config.DefaultTTL,
			CacheKeyPrefix: config.CacheKeyPrefix + ":inventory",
			LoaderFunc:     loadUserInventory,
		}),
		achievementsLoader: NewGenericLazyLoader(GenericLazyLoaderConfig[*UserAchievements]{
			CacheStorage:   config.CacheStorage,
			DefaultTTL:     config.DefaultTTL,
			CacheKeyPrefix: config.CacheKeyPrefix + ":achievements",
			LoaderFunc:     loadUserAchievements,
		}),
		statsLoader: NewGenericLazyLoader(GenericLazyLoaderConfig[*UserStats]{
			CacheStorage:   config.CacheStorage,
			DefaultTTL:     config.DefaultTTL,
			CacheKeyPrefix: config.CacheKeyPrefix + ":stats",
			LoaderFunc:     loadUserStats,
		}),
		preferencesLoader: NewGenericLazyLoader(GenericLazyLoaderConfig[*UserPreferences]{
			CacheStorage:   config.CacheStorage,
			DefaultTTL:     config.DefaultTTL,
			CacheKeyPrefix: config.CacheKeyPrefix + ":preferences",
			LoaderFunc:     loadUserPreferences,
		}),
		socialLoader: NewGenericLazyLoader(GenericLazyLoaderConfig[*UserSocialData]{
			CacheStorage:   config.CacheStorage,
			DefaultTTL:     config.DefaultTTL,
			CacheKeyPrefix: config.CacheKeyPrefix + ":social",
			LoaderFunc:     loadUserSocialData,
		}),
	}
}

// LoadInventory loads user inventory with type safety
func (m *MultiTypeLazyLoader) LoadInventory(ctx context.Context, userID string) (*UserInventory, error) {
	return m.inventoryLoader.Load(ctx, userID)
}

// LoadAchievements loads user achievements with type safety
func (m *MultiTypeLazyLoader) LoadAchievements(ctx context.Context, userID string) (*UserAchievements, error) {
	return m.achievementsLoader.Load(ctx, userID)
}

// LoadStats loads user stats with type safety
func (m *MultiTypeLazyLoader) LoadStats(ctx context.Context, userID string) (*UserStats, error) {
	return m.statsLoader.Load(ctx, userID)
}

// LoadPreferences loads user preferences with type safety
func (m *MultiTypeLazyLoader) LoadPreferences(ctx context.Context, userID string) (*UserPreferences, error) {
	return m.preferencesLoader.Load(ctx, userID)
}

// LoadSocialData loads user social data with type safety
func (m *MultiTypeLazyLoader) LoadSocialData(ctx context.Context, userID string) (*UserSocialData, error) {
	return m.socialLoader.Load(ctx, userID)
}

// LoadAll loads all user data types
func (m *MultiTypeLazyLoader) LoadAll(ctx context.Context, userID string) (*UserData, error) {
	// Load all data concurrently using goroutines
	type result[T any] struct {
		data T
		err  error
	}

	inventoryChan := make(chan result[*UserInventory], 1)
	achievementsChan := make(chan result[*UserAchievements], 1)
	statsChan := make(chan result[*UserStats], 1)
	preferencesChan := make(chan result[*UserPreferences], 1)
	socialChan := make(chan result[*UserSocialData], 1)

	// Launch concurrent loads
	go func() {
		data, err := m.inventoryLoader.Load(ctx, userID)
		inventoryChan <- result[*UserInventory]{data, err}
	}()

	go func() {
		data, err := m.achievementsLoader.Load(ctx, userID)
		achievementsChan <- result[*UserAchievements]{data, err}
	}()

	go func() {
		data, err := m.statsLoader.Load(ctx, userID)
		statsChan <- result[*UserStats]{data, err}
	}()

	go func() {
		data, err := m.preferencesLoader.Load(ctx, userID)
		preferencesChan <- result[*UserPreferences]{data, err}
	}()

	go func() {
		data, err := m.socialLoader.Load(ctx, userID)
		socialChan <- result[*UserSocialData]{data, err}
	}()

	// Collect results
	inventoryResult := <-inventoryChan
	achievementsResult := <-achievementsChan
	statsResult := <-statsChan
	preferencesResult := <-preferencesChan
	socialResult := <-socialChan

	// Check for errors
	if inventoryResult.err != nil {
		return nil, fmt.Errorf("failed to load inventory: %w", inventoryResult.err)
	}
	if achievementsResult.err != nil {
		return nil, fmt.Errorf("failed to load achievements: %w", achievementsResult.err)
	}
	if statsResult.err != nil {
		return nil, fmt.Errorf("failed to load stats: %w", statsResult.err)
	}
	if preferencesResult.err != nil {
		return nil, fmt.Errorf("failed to load preferences: %w", preferencesResult.err)
	}
	if socialResult.err != nil {
		return nil, fmt.Errorf("failed to load social data: %w", socialResult.err)
	}

	return &UserData{
		UserID:       userID,
		Inventory:    inventoryResult.data,
		Achievements: achievementsResult.data,
		Stats:        statsResult.data,
		Preferences:  preferencesResult.data,
		SocialData:   socialResult.data,
		LoadedAt:     time.Now(),
	}, nil
}

// InvalidateUserCache invalidates all cache for a user
func (m *MultiTypeLazyLoader) InvalidateUserCache(ctx context.Context, userID string) error {
	var errs []error

	if err := m.inventoryLoader.InvalidateCache(ctx, userID); err != nil {
		errs = append(errs, fmt.Errorf("inventory cache: %w", err))
	}
	if err := m.achievementsLoader.InvalidateCache(ctx, userID); err != nil {
		errs = append(errs, fmt.Errorf("achievements cache: %w", err))
	}
	if err := m.statsLoader.InvalidateCache(ctx, userID); err != nil {
		errs = append(errs, fmt.Errorf("stats cache: %w", err))
	}
	if err := m.preferencesLoader.InvalidateCache(ctx, userID); err != nil {
		errs = append(errs, fmt.Errorf("preferences cache: %w", err))
	}
	if err := m.socialLoader.InvalidateCache(ctx, userID); err != nil {
		errs = append(errs, fmt.Errorf("social cache: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("cache invalidation errors: %v", errs)
	}

	return nil
}

// UserData represents all user data loaded together
type UserData struct {
	UserID       string              `json:"user_id"`
	Inventory    *UserInventory      `json:"inventory"`
	Achievements *UserAchievements   `json:"achievements"`
	Stats        *UserStats          `json:"stats"`
	Preferences  *UserPreferences    `json:"preferences"`
	SocialData   *UserSocialData     `json:"social_data"`
	LoadedAt     time.Time           `json:"loaded_at"`
}

// Loader functions for each data type
func loadUserInventory(ctx context.Context, userID string) (*UserInventory, error) {
	// Simulate loading from database
	inventory := NewUserInventory(userID, 100)
	
	// Add some sample items
	inventory.AddItem(&InventoryItem{
		ID: "default_sword", Name: "Starter Sword", Quantity: 1,
		ItemType: "weapon", Rarity: "common", AcquiredAt: time.Now(),
	})
	
	return inventory, nil
}

func loadUserAchievements(ctx context.Context, userID string) (*UserAchievements, error) {
	// Simulate loading from database
	achievements := NewUserAchievements(userID)
	
	// Add some default achievements
	now := time.Now()
	achievements.Achievements["welcome"] = &Achievement{
		ID: "welcome", Name: "Welcome!", Category: "onboarding",
		Points: 10, Rarity: "common", IsUnlocked: true,
		UnlockedAt: &now,
	}
	achievements.UnlockedCount = 1
	achievements.TotalCount = 100
	achievements.AchievementPoints = 10
	
	return achievements, nil
}

func loadUserStats(ctx context.Context, userID string) (*UserStats, error) {
	// Simulate loading from database
	stats := NewUserStats(userID)
	
	// Add default stats
	stats.GlobalStats["level"] = 1
	stats.GlobalStats["experience"] = 0
	stats.GlobalStats["total_playtime"] = "0h 0m"
	
	return stats, nil
}

func loadUserPreferences(ctx context.Context, userID string) (*UserPreferences, error) {
	// Simulate loading from database - return default preferences
	return NewUserPreferences(userID), nil
}

func loadUserSocialData(ctx context.Context, userID string) (*UserSocialData, error) {
	// Simulate loading from database - return empty social data
	return NewUserSocialData(userID), nil
}
package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// CacheStorage defines the interface for cache storage implementations
type CacheStorage interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)
	GetTTL(ctx context.Context, key string) (time.Duration, error)
}

// LazyLoader manages lazy loading of user data with flexible storage backend
type LazyLoader struct {
	cacheStorage   CacheStorage
	defaultTTL     time.Duration
	cacheKeyPrefix string
	loaderFuncs    map[string]LoaderFunc
	mu             sync.RWMutex
}

// LazyLoaderConfig contains configuration for LazyLoader
type LazyLoaderConfig struct {
	CacheStorage   CacheStorage
	DefaultTTL     time.Duration
	CacheKeyPrefix string
}

// NewLazyLoader creates a new LazyLoader instance
func NewLazyLoader(config LazyLoaderConfig) *LazyLoader {
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 15 * time.Minute
	}
	if config.CacheKeyPrefix == "" {
		config.CacheKeyPrefix = "lazy_load"
	}

	return &LazyLoader{
		cacheStorage:   config.CacheStorage,
		defaultTTL:     config.DefaultTTL,
		cacheKeyPrefix: config.CacheKeyPrefix,
		loaderFuncs:    make(map[string]LoaderFunc),
	}
}

// RegisterLoader registers a loader function for a specific data type
func (l *LazyLoader) RegisterLoader(dataType string, loader LoaderFunc) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.loaderFuncs[dataType] = loader
}

// Load loads data for a specific key and caches it in Redis
func (l *LazyLoader) Load(ctx context.Context, dataType, key string) (any, error) {
	// Try to get from cache first
	if cached, err := l.getFromCache(ctx, dataType, key); err == nil {
		return cached, nil
	}

	// Get the loader function
	l.mu.RLock()
	loader, exists := l.loaderFuncs[dataType]
	l.mu.RUnlock()

	if !exists {
		return nil, errors.New("no loader registered for data type: " + dataType)
	}

	// Load data using the registered loader
	data, err := loader(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to load data: %w", err)
	}

	// Cache the loaded data
	if err := l.setToCache(ctx, dataType, key, data); err != nil {
		// Log the error but don't fail the operation
		// In a production system, you might want to use a proper logger
		fmt.Printf("Failed to cache data for %s:%s: %v\n", dataType, key, err)
	}

	return data, nil
}

// LoadMultiple loads multiple keys of the same data type
func (l *LazyLoader) LoadMultiple(ctx context.Context, dataType string, keys []string) (map[string]any, error) {
	results := make(map[string]any)
	uncachedKeys := make([]string, 0)

	// First, try to get all keys from cache
	for _, key := range keys {
		if cached, err := l.getFromCache(ctx, dataType, key); err == nil {
			results[key] = cached
		} else {
			uncachedKeys = append(uncachedKeys, key)
		}
	}

	// If all keys were cached, return early
	if len(uncachedKeys) == 0 {
		return results, nil
	}

	// Get the loader function
	l.mu.RLock()
	loader, exists := l.loaderFuncs[dataType]
	l.mu.RUnlock()

	if !exists {
		return nil, errors.New("no loader registered for data type: " + dataType)
	}

	// Load uncached data
	for _, key := range uncachedKeys {
		data, err := loader(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("failed to load data for key %s: %w", key, err)
		}

		results[key] = data

		// Cache the loaded data
		if err := l.setToCache(ctx, dataType, key, data); err != nil {
			fmt.Printf("Failed to cache data for %s:%s: %v\n", dataType, key, err)
		}
	}

	return results, nil
}

// InvalidateCache removes cached data for a specific key
func (l *LazyLoader) InvalidateCache(ctx context.Context, dataType, key string) error {
	cacheKey := l.buildCacheKey(dataType, key)
	return l.cacheStorage.Delete(ctx, cacheKey)
}

// InvalidateCachePattern removes cached data matching a pattern
func (l *LazyLoader) InvalidateCachePattern(ctx context.Context, dataType, pattern string) error {
	cacheKeyPattern := l.buildCacheKey(dataType, pattern)
	return l.cacheStorage.DeletePattern(ctx, cacheKeyPattern)
}

// PreloadData preloads data for a specific key
func (l *LazyLoader) PreloadData(ctx context.Context, dataType, key string) error {
	_, err := l.Load(ctx, dataType, key)
	return err
}

// PreloadMultipleData preloads data for multiple keys
func (l *LazyLoader) PreloadMultipleData(ctx context.Context, dataType string, keys []string) error {
	_, err := l.LoadMultiple(ctx, dataType, keys)
	return err
}

// GetCacheInfo returns information about cached data
func (l *LazyLoader) GetCacheInfo(ctx context.Context, dataType, key string) (*CacheInfo, error) {
	cacheKey := l.buildCacheKey(dataType, key)
	
	// Check if key exists
	exists, err := l.cacheStorage.Exists(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	if !exists {
		return &CacheInfo{
			Key:      key,
			DataType: dataType,
			Exists:   false,
		}, nil
	}

	// Get TTL
	ttl, err := l.cacheStorage.GetTTL(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	return &CacheInfo{
		Key:      key,
		DataType: dataType,
		Exists:   true,
		TTL:      ttl,
	}, nil
}

// CacheInfo contains information about cached data
type CacheInfo struct {
	Key      string
	DataType string
	Exists   bool
	TTL      time.Duration
}

// buildCacheKey builds a cache key for the given data type and key
func (l *LazyLoader) buildCacheKey(dataType, key string) string {
	return fmt.Sprintf("%s:%s:%s", l.cacheKeyPrefix, dataType, key)
}

// getFromCache retrieves data from cache storage
func (l *LazyLoader) getFromCache(ctx context.Context, dataType, key string) (any, error) {
	cacheKey := l.buildCacheKey(dataType, key)
	
	data, err := l.cacheStorage.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	return result, nil
}

// setToCache stores data in cache storage
func (l *LazyLoader) setToCache(ctx context.Context, dataType, key string, data any) error {
	cacheKey := l.buildCacheKey(dataType, key)
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data for caching: %w", err)
	}

	return l.cacheStorage.Set(ctx, cacheKey, jsonData, l.defaultTTL)
}

// UserLazyLoader is a specialized lazy loader for user-related data
type UserLazyLoader struct {
	*LazyLoader
}

// NewUserLazyLoader creates a new UserLazyLoader
func NewUserLazyLoader(config LazyLoaderConfig) *UserLazyLoader {
	config.CacheKeyPrefix = "user_lazy_load"
	baseLoader := NewLazyLoader(config)
	
	return &UserLazyLoader{
		LazyLoader: baseLoader,
	}
}

// LoadUserInventory loads user inventory data
func (ul *UserLazyLoader) LoadUserInventory(ctx context.Context, userID string) (*UserInventory, error) {
	data, err := ul.Load(ctx, "inventory", userID)
	if err != nil {
		return nil, err
	}

	inventory, ok := data.(*UserInventory)
	if !ok {
		return nil, errors.New("invalid inventory data type")
	}

	return inventory, nil
}

// LoadUserAchievements loads user achievements data
func (ul *UserLazyLoader) LoadUserAchievements(ctx context.Context, userID string) (*UserAchievements, error) {
	data, err := ul.Load(ctx, "achievements", userID)
	if err != nil {
		return nil, err
	}

	achievements, ok := data.(*UserAchievements)
	if !ok {
		return nil, errors.New("invalid achievements data type")
	}

	return achievements, nil
}

// LoadUserStats loads user statistics data
func (ul *UserLazyLoader) LoadUserStats(ctx context.Context, userID string) (*UserStats, error) {
	data, err := ul.Load(ctx, "stats", userID)
	if err != nil {
		return nil, err
	}

	stats, ok := data.(*UserStats)
	if !ok {
		return nil, errors.New("invalid stats data type")
	}

	return stats, nil
}

// LoadUserPreferences loads user preferences data
func (ul *UserLazyLoader) LoadUserPreferences(ctx context.Context, userID string) (*UserPreferences, error) {
	data, err := ul.Load(ctx, "preferences", userID)
	if err != nil {
		return nil, err
	}

	preferences, ok := data.(*UserPreferences)
	if !ok {
		return nil, errors.New("invalid preferences data type")
	}

	return preferences, nil
}

// LoadUserSocialData loads user social data
func (ul *UserLazyLoader) LoadUserSocialData(ctx context.Context, userID string) (*UserSocialData, error) {
	data, err := ul.Load(ctx, "social", userID)
	if err != nil {
		return nil, err
	}

	socialData, ok := data.(*UserSocialData)
	if !ok {
		return nil, errors.New("invalid social data type")
	}

	return socialData, nil
}

// BatchLoadUserData loads multiple user data types for a user
func (ul *UserLazyLoader) BatchLoadUserData(ctx context.Context, userID string, dataTypes []string) (map[string]any, error) {
	results := make(map[string]any)
	
	for _, dataType := range dataTypes {
		data, err := ul.Load(ctx, dataType, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s data: %w", dataType, err)
		}
		results[dataType] = data
	}
	
	return results, nil
}
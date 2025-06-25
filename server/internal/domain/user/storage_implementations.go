package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisStorage implements CacheStorage interface using Redis
type RedisStorage struct {
	client *redis.Client
}

// NewRedisStorage creates a new Redis storage implementation
func NewRedisStorage(client *redis.Client) *RedisStorage {
	return &RedisStorage{client: client}
}

func (r *RedisStorage) Get(ctx context.Context, key string) ([]byte, error) {
	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return []byte(result), nil
}

func (r *RedisStorage) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisStorage) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisStorage) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}
	return nil
}

func (r *RedisStorage) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	return result > 0, err
}

func (r *RedisStorage) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// InMemoryStorage implements CacheStorage interface using in-memory storage
type InMemoryStorage struct {
	data     map[string]*cacheItem
	mu       sync.RWMutex
	stopChan chan struct{}
	started  bool
}

type cacheItem struct {
	value     []byte
	expiresAt time.Time
}

// NewInMemoryStorage creates a new in-memory storage implementation
func NewInMemoryStorage() *InMemoryStorage {
	storage := &InMemoryStorage{
		data:     make(map[string]*cacheItem),
		stopChan: make(chan struct{}),
	}
	storage.startCleanupRoutine()
	return storage
}

func (m *InMemoryStorage) Get(ctx context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	item, exists := m.data[key]
	if !exists {
		return nil, errors.New("key not found")
	}
	
	if time.Now().After(item.expiresAt) {
		delete(m.data, key)
		return nil, errors.New("key expired")
	}
	
	return item.value, nil
}

func (m *InMemoryStorage) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	expiresAt := time.Now().Add(ttl)
	if ttl <= 0 {
		expiresAt = time.Now().Add(24 * time.Hour) // Default 24 hours
	}
	
	m.data[key] = &cacheItem{
		value:     make([]byte, len(value)),
		expiresAt: expiresAt,
	}
	copy(m.data[key].value, value)
	
	return nil
}

func (m *InMemoryStorage) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	delete(m.data, key)
	return nil
}

func (m *InMemoryStorage) DeletePattern(ctx context.Context, pattern string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Simple pattern matching - only supports * at the end
	prefix := strings.TrimSuffix(pattern, "*")
	
	keysToDelete := make([]string, 0)
	for key := range m.data {
		if strings.HasPrefix(key, prefix) {
			keysToDelete = append(keysToDelete, key)
		}
	}
	
	for _, key := range keysToDelete {
		delete(m.data, key)
	}
	
	return nil
}

func (m *InMemoryStorage) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	item, exists := m.data[key]
	if !exists {
		return false, nil
	}
	
	if time.Now().After(item.expiresAt) {
		delete(m.data, key)
		return false, nil
	}
	
	return true, nil
}

func (m *InMemoryStorage) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	item, exists := m.data[key]
	if !exists {
		return 0, errors.New("key not found")
	}
	
	if time.Now().After(item.expiresAt) {
		delete(m.data, key)
		return 0, errors.New("key expired")
	}
	
	return time.Until(item.expiresAt), nil
}

func (m *InMemoryStorage) startCleanupRoutine() {
	if m.started {
		return
	}
	m.started = true
	
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				m.cleanup()
			case <-m.stopChan:
				return
			}
		}
	}()
}

func (m *InMemoryStorage) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	now := time.Now()
	for key, item := range m.data {
		if now.After(item.expiresAt) {
			delete(m.data, key)
		}
	}
}

func (m *InMemoryStorage) Close() {
	close(m.stopChan)
}

// NoOpStorage implements CacheStorage interface with no-op operations (no caching)
type NoOpStorage struct{}

// NewNoOpStorage creates a new no-op storage implementation
func NewNoOpStorage() *NoOpStorage {
	return &NoOpStorage{}
}

func (n *NoOpStorage) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, errors.New("no cache available")
}

func (n *NoOpStorage) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil // No-op, always succeeds
}

func (n *NoOpStorage) Delete(ctx context.Context, key string) error {
	return nil // No-op, always succeeds
}

func (n *NoOpStorage) DeletePattern(ctx context.Context, pattern string) error {
	return nil // No-op, always succeeds
}

func (n *NoOpStorage) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (n *NoOpStorage) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, errors.New("no cache available")
}

// CompositeStorage combines multiple storage implementations with fallback
type CompositeStorage struct {
	primary   CacheStorage
	fallback  CacheStorage
	writeBoth bool
}

// NewCompositeStorage creates a composite storage with primary and fallback
func NewCompositeStorage(primary, fallback CacheStorage, writeBoth bool) *CompositeStorage {
	return &CompositeStorage{
		primary:   primary,
		fallback:  fallback,
		writeBoth: writeBoth,
	}
}

func (c *CompositeStorage) Get(ctx context.Context, key string) ([]byte, error) {
	// Try primary first
	if data, err := c.primary.Get(ctx, key); err == nil {
		return data, nil
	}
	
	// Fallback to secondary
	if c.fallback != nil {
		return c.fallback.Get(ctx, key)
	}
	
	return nil, errors.New("data not found in any storage")
}

func (c *CompositeStorage) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	// Always write to primary
	err := c.primary.Set(ctx, key, value, ttl)
	
	// Write to fallback if configured
	if c.writeBoth && c.fallback != nil {
		if fallbackErr := c.fallback.Set(ctx, key, value, ttl); fallbackErr != nil {
			// Log but don't fail the operation
			fmt.Printf("Failed to write to fallback storage: %v\n", fallbackErr)
		}
	}
	
	return err
}

func (c *CompositeStorage) Delete(ctx context.Context, key string) error {
	var errs []error
	
	if err := c.primary.Delete(ctx, key); err != nil {
		errs = append(errs, err)
	}
	
	if c.fallback != nil {
		if err := c.fallback.Delete(ctx, key); err != nil {
			errs = append(errs, err)
		}
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("delete errors: %v", errs)
	}
	
	return nil
}

func (c *CompositeStorage) DeletePattern(ctx context.Context, pattern string) error {
	var errs []error
	
	if err := c.primary.DeletePattern(ctx, pattern); err != nil {
		errs = append(errs, err)
	}
	
	if c.fallback != nil {
		if err := c.fallback.DeletePattern(ctx, pattern); err != nil {
			errs = append(errs, err)
		}
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("delete pattern errors: %v", errs)
	}
	
	return nil
}

func (c *CompositeStorage) Exists(ctx context.Context, key string) (bool, error) {
	// Check primary first
	if exists, err := c.primary.Exists(ctx, key); err == nil && exists {
		return true, nil
	}
	
	// Check fallback
	if c.fallback != nil {
		return c.fallback.Exists(ctx, key)
	}
	
	return false, nil
}

func (c *CompositeStorage) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	// Try primary first
	if ttl, err := c.primary.GetTTL(ctx, key); err == nil {
		return ttl, nil
	}
	
	// Try fallback
	if c.fallback != nil {
		return c.fallback.GetTTL(ctx, key)
	}
	
	return 0, errors.New("key not found in any storage")
}
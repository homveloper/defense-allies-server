package cqrs

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClientManager manages Redis connections for CQRS infrastructure
type RedisClientManager struct {
	client  *redis.Client
	config  *RedisConfig
	metrics *RedisMetrics
}

// RedisMetrics represents Redis performance metrics
type RedisMetrics struct {
	ConnectionCount   int64
	CommandCount      int64
	ErrorCount        int64
	AverageLatency    time.Duration
	LastCommandTime   time.Time
	PoolStats         *redis.PoolStats
}

// NewRedisClientManager creates a new Redis client manager
func NewRedisClientManager(config *RedisConfig) (*RedisClientManager, error) {
	if config == nil {
		return nil, NewCQRSError(ErrCodeRepositoryError.String(), "Redis config cannot be nil", nil)
	}

	if err := validateRedisConfig(config); err != nil {
		return nil, err
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.Database,
		PoolSize:     config.PoolSize,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	manager := &RedisClientManager{
		client: client,
		config: config,
		metrics: &RedisMetrics{
			ConnectionCount: 0,
			CommandCount:    0,
			ErrorCount:      0,
			AverageLatency:  0,
			LastCommandTime: time.Time{},
		},
	}

	return manager, nil
}

// GetClient returns the Redis client
func (rm *RedisClientManager) GetClient() *redis.Client {
	return rm.client
}

// GetConfig returns the Redis configuration
func (rm *RedisClientManager) GetConfig() *RedisConfig {
	return rm.config
}

// Ping tests the Redis connection
func (rm *RedisClientManager) Ping(ctx context.Context) error {
	start := time.Now()
	
	err := rm.client.Ping(ctx).Err()
	
	rm.updateMetrics(time.Since(start), err)
	
	if err != nil {
		return NewCQRSError(ErrCodeRepositoryError.String(), "Redis ping failed", err)
	}
	
	return nil
}

// Close closes the Redis connection
func (rm *RedisClientManager) Close() error {
	if rm.client != nil {
		return rm.client.Close()
	}
	return nil
}

// GetMetrics returns current Redis metrics
func (rm *RedisClientManager) GetMetrics() *RedisMetrics {
	// Update pool stats
	if rm.client != nil {
		rm.metrics.PoolStats = rm.client.PoolStats()
	}

	// Return a copy of metrics
	return &RedisMetrics{
		ConnectionCount: rm.metrics.ConnectionCount,
		CommandCount:    rm.metrics.CommandCount,
		ErrorCount:      rm.metrics.ErrorCount,
		AverageLatency:  rm.metrics.AverageLatency,
		LastCommandTime: rm.metrics.LastCommandTime,
		PoolStats:       rm.metrics.PoolStats,
	}
}

// ExecuteCommand executes a Redis command with metrics tracking
func (rm *RedisClientManager) ExecuteCommand(ctx context.Context, cmd func() error) error {
	start := time.Now()
	
	err := cmd()
	
	rm.updateMetrics(time.Since(start), err)
	
	return err
}

// Helper methods

func (rm *RedisClientManager) updateMetrics(latency time.Duration, err error) {
	rm.metrics.CommandCount++
	rm.metrics.LastCommandTime = time.Now()
	
	if err != nil {
		rm.metrics.ErrorCount++
	}
	
	// Update average latency
	if rm.metrics.CommandCount == 1 {
		rm.metrics.AverageLatency = latency
	} else {
		rm.metrics.AverageLatency = (rm.metrics.AverageLatency + latency) / 2
	}
}

func validateRedisConfig(config *RedisConfig) error {
	if config.Host == "" {
		return NewCQRSError(ErrCodeRepositoryError.String(), "Redis host cannot be empty", nil)
	}
	
	if config.Port <= 0 || config.Port > 65535 {
		return NewCQRSError(ErrCodeRepositoryError.String(), "Redis port must be between 1 and 65535", nil)
	}
	
	if config.Database < 0 || config.Database > 15 {
		return NewCQRSError(ErrCodeRepositoryError.String(), "Redis database must be between 0 and 15", nil)
	}
	
	if config.PoolSize <= 0 {
		config.PoolSize = 10 // Default pool size
	}
	
	if config.MaxRetries < 0 {
		config.MaxRetries = 3 // Default max retries
	}
	
	if config.DialTimeout <= 0 {
		config.DialTimeout = 5 * time.Second // Default dial timeout
	}
	
	if config.ReadTimeout <= 0 {
		config.ReadTimeout = 3 * time.Second // Default read timeout
	}
	
	if config.WriteTimeout <= 0 {
		config.WriteTimeout = 3 * time.Second // Default write timeout
	}
	
	return nil
}

// RedisKeyBuilder helps build consistent Redis keys
type RedisKeyBuilder struct {
	prefix string
}

// NewRedisKeyBuilder creates a new Redis key builder
func NewRedisKeyBuilder(prefix string) *RedisKeyBuilder {
	return &RedisKeyBuilder{
		prefix: prefix,
	}
}

// AggregateKey builds a key for aggregate storage
func (kb *RedisKeyBuilder) AggregateKey(aggregateType, aggregateID string) string {
	return fmt.Sprintf("%s:aggregate:%s:%s", kb.prefix, aggregateType, aggregateID)
}

// EventKey builds a key for event storage
func (kb *RedisKeyBuilder) EventKey(aggregateType, aggregateID string) string {
	return fmt.Sprintf("%s:events:%s:%s", kb.prefix, aggregateType, aggregateID)
}

// SnapshotKey builds a key for snapshot storage
func (kb *RedisKeyBuilder) SnapshotKey(aggregateType, aggregateID string) string {
	return fmt.Sprintf("%s:snapshot:%s:%s", kb.prefix, aggregateType, aggregateID)
}

// ReadModelKey builds a key for read model storage
func (kb *RedisKeyBuilder) ReadModelKey(modelType, modelID string) string {
	return fmt.Sprintf("%s:readmodel:%s:%s", kb.prefix, modelType, modelID)
}

// IndexKey builds a key for index storage
func (kb *RedisKeyBuilder) IndexKey(modelType, field string) string {
	return fmt.Sprintf("%s:index:%s:%s", kb.prefix, modelType, field)
}

// MetadataKey builds a key for metadata storage
func (kb *RedisKeyBuilder) MetadataKey(aggregateType, aggregateID string) string {
	return fmt.Sprintf("%s:metadata:%s:%s", kb.prefix, aggregateType, aggregateID)
}

// LockKey builds a key for distributed locking
func (kb *RedisKeyBuilder) LockKey(aggregateType, aggregateID string) string {
	return fmt.Sprintf("%s:lock:%s:%s", kb.prefix, aggregateType, aggregateID)
}

// StreamKey builds a key for event streaming
func (kb *RedisKeyBuilder) StreamKey(streamName string) string {
	return fmt.Sprintf("%s:stream:%s", kb.prefix, streamName)
}

// GetPrefix returns the key prefix
func (kb *RedisKeyBuilder) GetPrefix() string {
	return kb.prefix
}

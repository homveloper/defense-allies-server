package cqrsx

import (
	"context"
	"defense-allies-server/pkg/cqrs"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRedisClientManager(t *testing.T) {
	// Arrange
	config := &cqrs.RedisConfig{
		Host:         "localhost",
		Port:         6379,
		Database:     0,
		Password:     "",
		PoolSize:     10,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	// Act
	manager, err := NewRedisClientManager(config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.GetClient())
	assert.Equal(t, config, manager.GetConfig())

	// Cleanup
	manager.Close()
}

func TestNewRedisClientManager_NilConfig(t *testing.T) {
	// Act
	manager, err := NewRedisClientManager(nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, manager)
	assert.Contains(t, err.Error(), "Redis config cannot be nil")
}

func TestNewRedisClientManager_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *cqrs.RedisConfig
		errMsg string
	}{
		{
			name: "Empty host",
			config: &cqrs.RedisConfig{
				Host: "",
				Port: 6379,
			},
			errMsg: "Redis host cannot be empty",
		},
		{
			name: "Invalid port - zero",
			config: &cqrs.RedisConfig{
				Host: "localhost",
				Port: 0,
			},
			errMsg: "Redis port must be between 1 and 65535",
		},
		{
			name: "Invalid port - too high",
			config: &cqrs.RedisConfig{
				Host: "localhost",
				Port: 70000,
			},
			errMsg: "Redis port must be between 1 and 65535",
		},
		{
			name: "Invalid database - negative",
			config: &cqrs.RedisConfig{
				Host:     "localhost",
				Port:     6379,
				Database: -1,
			},
			errMsg: "Redis database must be between 0 and 15",
		},
		{
			name: "Invalid database - too high",
			config: &cqrs.RedisConfig{
				Host:     "localhost",
				Port:     6379,
				Database: 16,
			},
			errMsg: "Redis database must be between 0 and 15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			manager, err := NewRedisClientManager(tt.config)

			// Assert
			assert.Error(t, err)
			assert.Nil(t, manager)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestRedisClientManager_ConfigDefaults(t *testing.T) {
	// Arrange
	config := &cqrs.RedisConfig{
		Host: "localhost",
		Port: 6379,
		// Other fields left empty to test defaults
		MaxRetries: -1, // Set to negative to trigger default
	}

	// Act
	manager, err := NewRedisClientManager(config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	// Check that defaults were applied
	assert.Equal(t, 10, config.PoolSize)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 5*time.Second, config.DialTimeout)
	assert.Equal(t, 3*time.Second, config.ReadTimeout)
	assert.Equal(t, 3*time.Second, config.WriteTimeout)

	// Cleanup
	manager.Close()
}

func TestRedisClientManager_GetMetrics(t *testing.T) {
	// Arrange
	config := &cqrs.RedisConfig{
		Host: "localhost",
		Port: 6379,
	}
	manager, err := NewRedisClientManager(config)
	assert.NoError(t, err)
	defer manager.Close()

	// Act
	metrics := manager.GetMetrics()

	// Assert
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.CommandCount)
	assert.Equal(t, int64(0), metrics.ErrorCount)
	assert.Equal(t, time.Duration(0), metrics.AverageLatency)
}

func TestRedisClientManager_ExecuteCommand(t *testing.T) {
	// Arrange
	config := &cqrs.RedisConfig{
		Host: "localhost",
		Port: 6379,
	}
	manager, err := NewRedisClientManager(config)
	assert.NoError(t, err)
	defer manager.Close()

	// Act
	err = manager.ExecuteCommand(context.Background(), func() error {
		return nil // Successful command
	})

	// Assert
	assert.NoError(t, err)

	metrics := manager.GetMetrics()
	assert.Equal(t, int64(1), metrics.CommandCount)
	assert.Equal(t, int64(0), metrics.ErrorCount)
}

func TestRedisClientManager_ExecuteCommand_WithError(t *testing.T) {
	// Arrange
	config := &cqrs.RedisConfig{
		Host: "localhost",
		Port: 6379,
	}
	manager, err := NewRedisClientManager(config)
	assert.NoError(t, err)
	defer manager.Close()

	expectedError := cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "test error", nil)

	// Act
	err = manager.ExecuteCommand(context.Background(), func() error {
		return expectedError
	})

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)

	metrics := manager.GetMetrics()
	assert.Equal(t, int64(1), metrics.CommandCount)
	assert.Equal(t, int64(1), metrics.ErrorCount)
}

func TestRedisKeyBuilder(t *testing.T) {
	// Arrange
	prefix := "test"
	kb := NewRedisKeyBuilder(prefix)

	// Test cases
	tests := []struct {
		name     string
		method   func() string
		expected string
	}{
		{
			name:     "AggregateKey",
			method:   func() string { return kb.AggregateKey("User", "123") },
			expected: "test:aggregate:User:123",
		},
		{
			name:     "EventKey",
			method:   func() string { return kb.EventKey("User", "123") },
			expected: "test:events:User:123",
		},
		{
			name:     "SnapshotKey",
			method:   func() string { return kb.SnapshotKey("User", "123") },
			expected: "test:snapshot:User:123",
		},
		{
			name:     "ReadModelKey",
			method:   func() string { return kb.ReadModelKey("UserView", "123") },
			expected: "test:readmodel:UserView:123",
		},
		{
			name:     "IndexKey",
			method:   func() string { return kb.IndexKey("UserView", "email") },
			expected: "test:index:UserView:email",
		},
		{
			name:     "MetadataKey",
			method:   func() string { return kb.MetadataKey("User", "123") },
			expected: "test:metadata:User:123",
		},
		{
			name:     "LockKey",
			method:   func() string { return kb.LockKey("User", "123") },
			expected: "test:lock:User:123",
		},
		{
			name:     "StreamKey",
			method:   func() string { return kb.StreamKey("events") },
			expected: "test:stream:events",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.method()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedisKeyBuilder_GetPrefix(t *testing.T) {
	// Arrange
	prefix := "myapp"
	kb := NewRedisKeyBuilder(prefix)

	// Act
	result := kb.GetPrefix()

	// Assert
	assert.Equal(t, prefix, result)
}

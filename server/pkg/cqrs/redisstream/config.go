package redisstream

import (
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStreamConfig defines configuration for Redis Stream EventBus
type RedisStreamConfig struct {
	// Redis connection settings
	Redis RedisConfig `yaml:"redis" json:"redis"`

	// Stream settings
	Stream StreamConfig `yaml:"stream" json:"stream"`

	// Retry settings
	Retry RetryConfig `yaml:"retry" json:"retry"`

	// Consumer settings
	Consumer ConsumerConfig `yaml:"consumer" json:"consumer"`

	// Monitoring settings
	Monitoring MonitoringConfig `yaml:"monitoring" json:"monitoring"`
}

// RedisConfig defines Redis connection configuration
type RedisConfig struct {
	// Redis server addresses for cluster mode
	Addrs []string `yaml:"addrs" json:"addrs"`

	// Single Redis server address (alternative to Addrs)
	Addr string `yaml:"addr" json:"addr"`

	// Authentication
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`

	// Database selection
	DB int `yaml:"db" json:"db"`

	// Connection settings
	MaxRetries   int           `yaml:"max_retries" json:"max_retries"`
	DialTimeout  time.Duration `yaml:"dial_timeout" json:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"`
	PoolSize     int           `yaml:"pool_size" json:"pool_size"`
	MinIdleConns int           `yaml:"min_idle_conns" json:"min_idle_conns"`
	PoolTimeout  time.Duration `yaml:"pool_timeout" json:"pool_timeout"`
	MaxIdleConns int           `yaml:"max_idle_conns" json:"max_idle_conns"`
}

// StreamConfig defines Redis Stream configuration
type StreamConfig struct {
	// Stream naming
	StreamPrefix   string `yaml:"stream_prefix" json:"stream_prefix"`
	NamespaceDelim string `yaml:"namespace_delim" json:"namespace_delim"`

	// Stream limits
	MaxLen       int64         `yaml:"max_len" json:"max_len"`
	MaxLenApprox int64         `yaml:"max_len_approx" json:"max_len_approx"`
	MinIDTrim    string        `yaml:"min_id_trim" json:"min_id_trim"`
	MessageTTL   time.Duration `yaml:"message_ttl" json:"message_ttl"`

	// Consumer Group settings
	ConsumerGroupPrefix string `yaml:"consumer_group_prefix" json:"consumer_group_prefix"`
	InstanceID          string `yaml:"instance_id" json:"instance_id"`

	// Reading settings
	BlockTime  time.Duration `yaml:"block_time" json:"block_time"`
	Count      int64         `yaml:"count" json:"count"`
	BufferSize int           `yaml:"buffer_size" json:"buffer_size"`

	// Priority settings
	EnablePriorityStreams bool `yaml:"enable_priority_streams" json:"enable_priority_streams"`

	// Dead letter queue
	DLQEnabled bool   `yaml:"dlq_enabled" json:"dlq_enabled"`
	DLQSuffix  string `yaml:"dlq_suffix" json:"dlq_suffix"`
}

// RetryConfig defines retry policy configuration
type RetryConfig struct {
	MaxAttempts   int           `yaml:"max_attempts" json:"max_attempts"`
	InitialDelay  time.Duration `yaml:"initial_delay" json:"initial_delay"`
	MaxDelay      time.Duration `yaml:"max_delay" json:"max_delay"`
	BackoffType   string        `yaml:"backoff_type" json:"backoff_type"` // "fixed", "exponential", "linear"
	BackoffFactor float64       `yaml:"backoff_factor" json:"backoff_factor"`

	// DLQ settings
	DLQEnabled     bool `yaml:"dlq_enabled" json:"dlq_enabled"`
	DLQMaxAttempts int  `yaml:"dlq_max_attempts" json:"dlq_max_attempts"`
}

// ConsumerConfig defines consumer configuration
type ConsumerConfig struct {
	// Consumer identification
	ServiceName string `yaml:"service_name" json:"service_name"`
	InstanceID  string `yaml:"instance_id" json:"instance_id"`

	// Processing settings
	MaxConcurrency    int           `yaml:"max_concurrency" json:"max_concurrency"`
	ProcessingTimeout time.Duration `yaml:"processing_timeout" json:"processing_timeout"`

	// Claiming settings
	ClaimInterval    time.Duration `yaml:"claim_interval" json:"claim_interval"`
	ClaimMinIdleTime time.Duration `yaml:"claim_min_idle_time" json:"claim_min_idle_time"`

	// Acknowledgment settings
	AutoAck    bool          `yaml:"auto_ack" json:"auto_ack"`
	AckTimeout time.Duration `yaml:"ack_timeout" json:"ack_timeout"`

	// Graceful shutdown
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" json:"shutdown_timeout"`
}

// MonitoringConfig defines monitoring configuration
type MonitoringConfig struct {
	Enabled             bool          `yaml:"enabled" json:"enabled"`
	MetricsInterval     time.Duration `yaml:"metrics_interval" json:"metrics_interval"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval" json:"health_check_interval"`

	// Circuit breaker settings
	CircuitBreakerEnabled bool          `yaml:"circuit_breaker_enabled" json:"circuit_breaker_enabled"`
	FailureThreshold      int           `yaml:"failure_threshold" json:"failure_threshold"`
	RecoveryTimeout       time.Duration `yaml:"recovery_timeout" json:"recovery_timeout"`

	// Logging settings
	LogLevel      string `yaml:"log_level" json:"log_level"`
	LogFormat     string `yaml:"log_format" json:"log_format"`
	EnableTracing bool   `yaml:"enable_tracing" json:"enable_tracing"`
}

// DefaultRedisStreamConfig returns default configuration
func DefaultRedisStreamConfig() *RedisStreamConfig {
	return &RedisStreamConfig{
		Redis: RedisConfig{
			Addr:         "localhost:6379",
			Password:     "",
			DB:           0,
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolSize:     10,
			MinIdleConns: 5,
			PoolTimeout:  4 * time.Second,
			MaxIdleConns: 10,
		},
		Stream: StreamConfig{
			StreamPrefix:          "events",
			NamespaceDelim:        ":",
			MaxLen:                10000,
			MaxLenApprox:          9000,
			MessageTTL:            24 * time.Hour,
			ConsumerGroupPrefix:   "defense_allies",
			InstanceID:            "default",
			BlockTime:             1 * time.Second,
			Count:                 10,
			BufferSize:            100,
			EnablePriorityStreams: true,
			DLQEnabled:            true,
			DLQSuffix:             "dlq",
		},
		Retry: RetryConfig{
			MaxAttempts:    3,
			InitialDelay:   100 * time.Millisecond,
			MaxDelay:       30 * time.Second,
			BackoffType:    "exponential",
			BackoffFactor:  2.0,
			DLQEnabled:     true,
			DLQMaxAttempts: 1,
		},
		Consumer: ConsumerConfig{
			ServiceName:       "default",
			InstanceID:        "default",
			MaxConcurrency:    10,
			ProcessingTimeout: 30 * time.Second,
			ClaimInterval:     1 * time.Minute,
			ClaimMinIdleTime:  5 * time.Minute,
			AutoAck:           false,
			AckTimeout:        10 * time.Second,
			ShutdownTimeout:   30 * time.Second,
		},
		Monitoring: MonitoringConfig{
			Enabled:               true,
			MetricsInterval:       10 * time.Second,
			HealthCheckInterval:   30 * time.Second,
			CircuitBreakerEnabled: true,
			FailureThreshold:      5,
			RecoveryTimeout:       1 * time.Minute,
			LogLevel:              "info",
			LogFormat:             "json",
			EnableTracing:         false,
		},
	}
}

// Validate validates the configuration
func (c *RedisStreamConfig) Validate() error {
	if c.Redis.Addr == "" && len(c.Redis.Addrs) == 0 {
		return ErrConfigInvalid("redis address must be specified")
	}

	if c.Stream.StreamPrefix == "" {
		return ErrConfigInvalid("stream prefix cannot be empty")
	}

	if c.Stream.ConsumerGroupPrefix == "" {
		return ErrConfigInvalid("consumer group prefix cannot be empty")
	}

	if c.Consumer.ServiceName == "" {
		return ErrConfigInvalid("service name cannot be empty")
	}

	if c.Retry.MaxAttempts < 1 {
		return ErrConfigInvalid("max retry attempts must be at least 1")
	}

	if c.Consumer.MaxConcurrency < 1 {
		return ErrConfigInvalid("max concurrency must be at least 1")
	}

	return nil
}

// CreateRedisClient creates a Redis client from configuration
func (c *RedisStreamConfig) CreateRedisClient() redis.UniversalClient {
	opts := &redis.UniversalOptions{
		Addrs:        c.Redis.Addrs,
		Username:     c.Redis.Username,
		Password:     c.Redis.Password,
		DB:           c.Redis.DB,
		MaxRetries:   c.Redis.MaxRetries,
		DialTimeout:  c.Redis.DialTimeout,
		ReadTimeout:  c.Redis.ReadTimeout,
		WriteTimeout: c.Redis.WriteTimeout,
		PoolSize:     c.Redis.PoolSize,
		MinIdleConns: c.Redis.MinIdleConns,
		PoolTimeout:  c.Redis.PoolTimeout,
		MaxIdleConns: c.Redis.MaxIdleConns,
	}

	// If single address is specified, use it instead of cluster mode
	if c.Redis.Addr != "" {
		opts.Addrs = []string{c.Redis.Addr}
	}

	return redis.NewUniversalClient(opts)
}

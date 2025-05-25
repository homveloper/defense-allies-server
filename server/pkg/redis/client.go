package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"defense-allies-server/configs"
)

// Client Redis 클라이언트 래퍼
type Client struct {
	rdb *redis.Client
	ctx context.Context
}

// NewClient 새로운 Redis 클라이언트를 생성합니다
func NewClient(config *configs.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	ctx := context.Background()

	// 연결 테스트
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{
		rdb: rdb,
		ctx: ctx,
	}, nil
}

// GetClient Redis 클라이언트 인스턴스를 반환합니다
func (c *Client) GetClient() *redis.Client {
	return c.rdb
}

// GetContext 컨텍스트를 반환합니다
func (c *Client) GetContext() context.Context {
	return c.ctx
}

// Close Redis 연결을 닫습니다
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Set 키-값을 설정합니다
func (c *Client) Set(key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(c.ctx, key, value, expiration).Err()
}

// Get 키로 값을 조회합니다
func (c *Client) Get(key string) (string, error) {
	return c.rdb.Get(c.ctx, key).Result()
}

// Del 키를 삭제합니다
func (c *Client) Del(keys ...string) error {
	return c.rdb.Del(c.ctx, keys...).Err()
}

// Exists 키가 존재하는지 확인합니다
func (c *Client) Exists(keys ...string) (int64, error) {
	return c.rdb.Exists(c.ctx, keys...).Result()
}

// HSet 해시 필드를 설정합니다
func (c *Client) HSet(key string, values ...interface{}) error {
	return c.rdb.HSet(c.ctx, key, values...).Err()
}

// HGet 해시 필드 값을 조회합니다
func (c *Client) HGet(key, field string) (string, error) {
	return c.rdb.HGet(c.ctx, key, field).Result()
}

// HGetAll 해시의 모든 필드를 조회합니다
func (c *Client) HGetAll(key string) (map[string]string, error) {
	return c.rdb.HGetAll(c.ctx, key).Result()
}

// HDel 해시 필드를 삭제합니다
func (c *Client) HDel(key string, fields ...string) error {
	return c.rdb.HDel(c.ctx, key, fields...).Err()
}

// LPush 리스트 앞쪽에 요소를 추가합니다
func (c *Client) LPush(key string, values ...interface{}) error {
	return c.rdb.LPush(c.ctx, key, values...).Err()
}

// RPush 리스트 뒤쪽에 요소를 추가합니다
func (c *Client) RPush(key string, values ...interface{}) error {
	return c.rdb.RPush(c.ctx, key, values...).Err()
}

// LPop 리스트 앞쪽에서 요소를 제거하고 반환합니다
func (c *Client) LPop(key string) (string, error) {
	return c.rdb.LPop(c.ctx, key).Result()
}

// RPop 리스트 뒤쪽에서 요소를 제거하고 반환합니다
func (c *Client) RPop(key string) (string, error) {
	return c.rdb.RPop(c.ctx, key).Result()
}

// LLen 리스트 길이를 반환합니다
func (c *Client) LLen(key string) (int64, error) {
	return c.rdb.LLen(c.ctx, key).Result()
}

// ZAdd Sorted Set에 요소를 추가합니다
func (c *Client) ZAdd(key string, members ...redis.Z) error {
	return c.rdb.ZAdd(c.ctx, key, members...).Err()
}

// ZRange Sorted Set에서 범위로 요소를 조회합니다
func (c *Client) ZRange(key string, start, stop int64) ([]string, error) {
	return c.rdb.ZRange(c.ctx, key, start, stop).Result()
}

// ZRangeWithScores 점수와 함께 Sorted Set 요소를 조회합니다
func (c *Client) ZRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	return c.rdb.ZRangeWithScores(c.ctx, key, start, stop).Result()
}

// Publish 메시지를 채널에 발행합니다
func (c *Client) Publish(channel string, message interface{}) error {
	return c.rdb.Publish(c.ctx, channel, message).Err()
}

// Subscribe 채널을 구독합니다
func (c *Client) Subscribe(channels ...string) *redis.PubSub {
	return c.rdb.Subscribe(c.ctx, channels...)
}

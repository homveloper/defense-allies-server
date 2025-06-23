package repositories

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"defense-allies-server/pkg/account/domain/gameaccount"
)

type RedisGameAccountRepository struct {
	client *redis.Client
}

func NewRedisGameAccountRepository(client *redis.Client) *RedisGameAccountRepository {
	return &RedisGameAccountRepository{
		client: client,
	}
}

const (
	gameAccountKeyPrefix        = "account:gameaccount:"
	gameAccountByUsernamePrefix = "account:gameaccount:username:"
)

func (r *RedisGameAccountRepository) Save(ctx context.Context, gameAccount *gameaccount.GameAccount) error {
	data, err := json.Marshal(gameAccount)
	if err != nil {
		return fmt.Errorf("failed to marshal game account: %w", err)
	}

	pipe := r.client.Pipeline()

	// 메인 키로 저장
	mainKey := gameAccountKeyPrefix + gameAccount.ID()
	pipe.Set(ctx, mainKey, data, 0)

	// Username으로 인덱싱
	usernameKey := gameAccountByUsernamePrefix + gameAccount.Username()
	pipe.Set(ctx, usernameKey, gameAccount.ID(), 0)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save game account: %w", err)
	}

	return nil
}

func (r *RedisGameAccountRepository) Load(ctx context.Context, id string) (*gameaccount.GameAccount, error) {
	key := gameAccountKeyPrefix + id
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load game account: %w", err)
	}

	var gameAccount gameaccount.GameAccount
	if err := json.Unmarshal([]byte(data), &gameAccount); err != nil {
		return nil, fmt.Errorf("failed to unmarshal game account: %w", err)
	}

	return &gameAccount, nil
}

func (r *RedisGameAccountRepository) FindByUsername(ctx context.Context, username string) (*gameaccount.GameAccount, error) {
	key := gameAccountByUsernamePrefix + username
	gameAccountID, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find game account by username: %w", err)
	}

	return r.Load(ctx, gameAccountID)
}

func (r *RedisGameAccountRepository) Delete(ctx context.Context, id string) error {
	// 먼저 GameAccount 로드하여 인덱스 키들을 가져옴
	gameAccount, err := r.Load(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to load game account for deletion: %w", err)
	}
	if gameAccount == nil {
		return nil // 이미 존재하지 않음
	}

	pipe := r.client.Pipeline()

	// 메인 키 삭제
	mainKey := gameAccountKeyPrefix + id
	pipe.Del(ctx, mainKey)

	// Username 인덱스 삭제
	usernameKey := gameAccountByUsernamePrefix + gameAccount.Username()
	pipe.Del(ctx, usernameKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete game account: %w", err)
	}

	return nil
}

func (r *RedisGameAccountRepository) Exists(ctx context.Context, id string) (bool, error) {
	key := gameAccountKeyPrefix + id
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check game account existence: %w", err)
	}
	return exists > 0, nil
}
package repositories

import (
	"context"
	"encoding/json"
	"fmt"

	"defense-allies-server/pkg/gameauth/domain/accountlink"
	"defense-allies-server/pkg/gameauth/domain/common"

	"github.com/redis/go-redis/v9"
)

type RedisAccountLinkRepository struct {
	client *redis.Client
}

func NewRedisAccountLinkRepository(client *redis.Client) *RedisAccountLinkRepository {
	return &RedisAccountLinkRepository{
		client: client,
	}
}

const (
	accountLinkKeyPrefix        = "gameauth:accountlink:"
	accountLinkByGameIDPrefix   = "gameauth:accountlink:gameid:"
	accountLinkByProviderPrefix = "gameauth:accountlink:provider:"
)

func (r *RedisAccountLinkRepository) Save(ctx context.Context, accountLink *accountlink.AccountLink) error {
	data, err := json.Marshal(accountLink)
	if err != nil {
		return fmt.Errorf("failed to marshal account link: %w", err)
	}

	pipe := r.client.Pipeline()

	// 메인 키로 저장
	mainKey := accountLinkKeyPrefix + accountLink.ID()
	pipe.Set(ctx, mainKey, data, 0)

	// GameAccountID로 인덱싱
	gameIDKey := accountLinkByGameIDPrefix + accountLink.GameAccountID()
	pipe.Set(ctx, gameIDKey, accountLink.ID(), 0)

	// Provider별 인덱싱
	for providerType, providerInfo := range accountLink.AuthProviders() {
		providerKey := fmt.Sprintf("%s%s:%s", accountLinkByProviderPrefix, providerType, providerInfo.ExternalID)
		pipe.Set(ctx, providerKey, accountLink.ID(), 0)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save account link: %w", err)
	}

	return nil
}

func (r *RedisAccountLinkRepository) Load(ctx context.Context, id string) (*accountlink.AccountLink, error) {
	key := accountLinkKeyPrefix + id
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load account link: %w", err)
	}

	var accountLink accountlink.AccountLink
	if err := json.Unmarshal([]byte(data), &accountLink); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account link: %w", err)
	}

	return &accountLink, nil
}

func (r *RedisAccountLinkRepository) FindByGameAccountID(ctx context.Context, gameAccountID string) (*accountlink.AccountLink, error) {
	key := accountLinkByGameIDPrefix + gameAccountID
	accountLinkID, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find account link by game account ID: %w", err)
	}

	return r.Load(ctx, accountLinkID)
}

func (r *RedisAccountLinkRepository) FindByProvider(ctx context.Context, providerType common.ProviderType, externalID string) (*accountlink.AccountLink, error) {
	key := fmt.Sprintf("%s%s:%s", accountLinkByProviderPrefix, providerType, externalID)
	accountLinkID, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find account link by provider: %w", err)
	}

	return r.Load(ctx, accountLinkID)
}

func (r *RedisAccountLinkRepository) Delete(ctx context.Context, id string) error {
	// 먼저 AccountLink 로드하여 인덱스 키들을 가져옴
	accountLink, err := r.Load(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to load account link for deletion: %w", err)
	}
	if accountLink == nil {
		return nil // 이미 존재하지 않음
	}

	pipe := r.client.Pipeline()

	// 메인 키 삭제
	mainKey := accountLinkKeyPrefix + id
	pipe.Del(ctx, mainKey)

	// GameAccountID 인덱스 삭제
	gameIDKey := accountLinkByGameIDPrefix + accountLink.GameAccountID()
	pipe.Del(ctx, gameIDKey)

	// Provider 인덱스들 삭제
	for providerType, providerInfo := range accountLink.AuthProviders() {
		providerKey := fmt.Sprintf("%s%s:%s", accountLinkByProviderPrefix, providerType, providerInfo.ExternalID)
		pipe.Del(ctx, providerKey)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete account link: %w", err)
	}

	return nil
}

func (r *RedisAccountLinkRepository) Exists(ctx context.Context, id string) (bool, error) {
	key := accountLinkKeyPrefix + id
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check account link existence: %w", err)
	}
	return exists > 0, nil
}

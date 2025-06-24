package pagitredis

import (
	"context"
	"encoding/json"
	
	"github.com/defense-allies/pagit"
	"github.com/redis/go-redis/v9"
)

// OffsetAdapter implements offset-based pagination for Redis sorted sets
type OffsetAdapter[T any] struct {
	client    *redis.Client
	key       string
	keyPrefix string // For multi-field sorting (optional)
	sort      pagit.SortConfig
	unmarshal func([]byte) (T, error)
}

// NewOffsetAdapter creates a new Redis offset pagination adapter
func NewOffsetAdapter[T any](client *redis.Client, key string, unmarshal func([]byte) (T, error)) *OffsetAdapter[T] {
	return &OffsetAdapter[T]{
		client:    client,
		key:       key,
		sort:      pagit.SortByIDDesc, // Default sort
		unmarshal: unmarshal,
	}
}

// NewOffsetAdapterWithSort creates a new Redis offset pagination adapter with custom sort
func NewOffsetAdapterWithSort[T any](client *redis.Client, key string, sort pagit.SortConfig, unmarshal func([]byte) (T, error)) *OffsetAdapter[T] {
	if sort.IsEmpty() {
		sort = pagit.SortByIDDesc
	}
	return &OffsetAdapter[T]{
		client:    client,
		key:       key,
		sort:      sort,
		unmarshal: unmarshal,
	}
}

// NewMultiFieldOffsetAdapter creates adapter for multiple sort fields
// keyPrefix should be like "users:sorted:" to access "users:sorted:name", "users:sorted:created_at" etc.
func NewMultiFieldOffsetAdapter[T any](client *redis.Client, keyPrefix string, sort pagit.SortConfig, unmarshal func([]byte) (T, error)) *OffsetAdapter[T] {
	if sort.IsEmpty() {
		sort = pagit.SortByIDDesc
	}
	return &OffsetAdapter[T]{
		client:    client,
		keyPrefix: keyPrefix,
		sort:      sort,
		unmarshal: unmarshal,
	}
}

// Count returns the total number of items in the sorted set
func (a *OffsetAdapter[T]) Count(ctx context.Context) (int64, error) {
	key := a.getKey()
	return a.client.ZCard(ctx, key).Result()
}

// Fetch retrieves items for offset-based pagination with sorting
func (a *OffsetAdapter[T]) Fetch(ctx context.Context, offset, limit int) ([]T, error) {
	key := a.getKey()
	primary := a.sort.Primary()
	
	var results []redis.Z
	var err error
	
	// Apply sort direction
	if primary.Direction == pagit.SortAsc {
		results, err = a.client.ZRangeWithScores(ctx, key, int64(offset), int64(offset+limit-1)).Result()
	} else {
		results, err = a.client.ZRevRangeWithScores(ctx, key, int64(offset), int64(offset+limit-1)).Result()
	}
	
	if err != nil {
		return nil, err
	}

	items := make([]T, 0, len(results))
	for _, result := range results {
		data := []byte(result.Member.(string))
		item, err := a.unmarshal(data)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// WithSort creates a new adapter with different sort configuration
func (a *OffsetAdapter[T]) WithSort(sort pagit.SortConfig) *OffsetAdapter[T] {
	return &OffsetAdapter[T]{
		client:    a.client,
		key:       a.key,
		keyPrefix: a.keyPrefix,
		sort:      sort,
		unmarshal: a.unmarshal,
	}
}

// getKey returns the appropriate Redis key based on sort configuration
func (a *OffsetAdapter[T]) getKey() string {
	if a.keyPrefix != "" {
		// Multi-field sorting: use keyPrefix + field
		primary := a.sort.Primary()
		return a.keyPrefix + primary.Field
	}
	// Single key: use the provided key
	return a.key
}

// HashOffsetAdapter implements offset-based pagination for Redis hashes
type HashOffsetAdapter[T any] struct {
	client    *redis.Client
	keyPrefix string
	listKey   string // Sorted set for ordering
	unmarshal func(map[string]string) (T, error)
}

// NewHashOffsetAdapter creates a new Redis hash offset pagination adapter
func NewHashOffsetAdapter[T any](client *redis.Client, keyPrefix, listKey string, unmarshal func(map[string]string) (T, error)) *HashOffsetAdapter[T] {
	return &HashOffsetAdapter[T]{
		client:    client,
		keyPrefix: keyPrefix,
		listKey:   listKey,
		unmarshal: unmarshal,
	}
}

// Count returns the total number of items
func (a *HashOffsetAdapter[T]) Count(ctx context.Context) (int64, error) {
	return a.client.ZCard(ctx, a.listKey).Result()
}

// Fetch retrieves items for offset-based pagination
func (a *HashOffsetAdapter[T]) Fetch(ctx context.Context, offset, limit int) ([]T, error) {
	// Get IDs from sorted set
	ids, err := a.client.ZRevRange(ctx, a.listKey, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, err
	}

	items := make([]T, 0, len(ids))
	for _, id := range ids {
		key := a.keyPrefix + ":" + id
		data, err := a.client.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, err
		}

		item, err := a.unmarshal(data)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// Helper function for JSON unmarshaling
func UnmarshalJSON[T any](data []byte) (T, error) {
	var item T
	err := json.Unmarshal(data, &item)
	return item, err
}
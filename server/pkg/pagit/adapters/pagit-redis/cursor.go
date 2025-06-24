package pagitredis

import (
	"context"
	"fmt"
	"strconv"

	"github.com/defense-allies/pagit"
	"github.com/redis/go-redis/v9"
)

// CursorAdapter implements cursor-based pagination for Redis sorted sets
type CursorAdapter[T any] struct {
	client    *redis.Client
	key       string
	keyPrefix string // For multi-field sorting (optional)
	sort      pagit.SortConfig
	unmarshal func([]byte) (T, error)
}

// NewCursorAdapter creates a new Redis cursor pagination adapter
func NewCursorAdapter[T any](client *redis.Client, key string, unmarshal func([]byte) (T, error)) *CursorAdapter[T] {
	return &CursorAdapter[T]{
		client:    client,
		key:       key,
		sort:      pagit.SortByIDDesc, // Default sort
		unmarshal: unmarshal,
	}
}

// NewCursorAdapterWithSort creates a new Redis cursor pagination adapter with custom sort
func NewCursorAdapterWithSort[T any](client *redis.Client, key string, sort pagit.SortConfig, unmarshal func([]byte) (T, error)) *CursorAdapter[T] {
	if sort.IsEmpty() {
		sort = pagit.SortByIDDesc
	}
	return &CursorAdapter[T]{
		client:    client,
		key:       key,
		sort:      sort,
		unmarshal: unmarshal,
	}
}

// NewMultiFieldCursorAdapter creates adapter for multiple sort fields
func NewMultiFieldCursorAdapter[T any](client *redis.Client, keyPrefix string, sort pagit.SortConfig, unmarshal func([]byte) (T, error)) *CursorAdapter[T] {
	if sort.IsEmpty() {
		sort = pagit.SortByIDDesc
	}
	return &CursorAdapter[T]{
		client:    client,
		keyPrefix: keyPrefix,
		sort:      sort,
		unmarshal: unmarshal,
	}
}

// FetchWithCursor retrieves items for cursor-based pagination with sorting
func (a *CursorAdapter[T]) FetchWithCursor(ctx context.Context, cursor string, limit int) ([]T, string, error) {
	key := a.getKey()
	primary := a.sort.Primary()
	
	var minScore, maxScore float64 = 0, 0
	var useMin, useMax bool
	
	if cursor != "" {
		score, err := strconv.ParseFloat(cursor, 64)
		if err != nil {
			return nil, "", fmt.Errorf("invalid cursor: %w", err)
		}
		
		if primary.Direction == pagit.SortAsc {
			minScore = score
			useMin = true
		} else {
			maxScore = score
			useMax = true
		}
	}

	var opt *redis.ZRangeBy
	var results []redis.Z
	var err error
	
	if primary.Direction == pagit.SortAsc {
		// Ascending order: get items with score > cursor
		min := "-inf"
		if useMin {
			min = fmt.Sprintf("(%f", minScore) // Exclusive
		}
		opt = &redis.ZRangeBy{
			Min:    min,
			Max:    "+inf",
			Offset: 0,
			Count:  int64(limit),
		}
		results, err = a.client.ZRangeByScoreWithScores(ctx, key, opt).Result()
	} else {
		// Descending order: get items with score < cursor
		max := "+inf"
		if useMax {
			max = fmt.Sprintf("(%f", maxScore) // Exclusive
		}
		opt = &redis.ZRangeBy{
			Min:    "-inf",
			Max:    max,
			Offset: 0,
			Count:  int64(limit),
		}
		results, err = a.client.ZRevRangeByScoreWithScores(ctx, key, opt).Result()
	}
	
	if err != nil {
		return nil, "", err
	}

	items := make([]T, 0, len(results))
	var nextCursor string

	for i, result := range results {
		if i == limit-1 {
			// This is the extra item for checking next page
			nextCursor = fmt.Sprintf("%f", result.Score)
			break
		}

		data := []byte(result.Member.(string))
		item, err := a.unmarshal(data)
		if err != nil {
			return nil, "", err
		}
		items = append(items, item)
	}

	return items, nextCursor, nil
}

// WithSort creates a new adapter with different sort configuration
func (a *CursorAdapter[T]) WithSort(sort pagit.SortConfig) *CursorAdapter[T] {
	return &CursorAdapter[T]{
		client:    a.client,
		key:       a.key,
		keyPrefix: a.keyPrefix,
		sort:      sort,
		unmarshal: a.unmarshal,
	}
}

// getKey returns the appropriate Redis key based on sort configuration
func (a *CursorAdapter[T]) getKey() string {
	if a.keyPrefix != "" {
		// Multi-field sorting: use keyPrefix + field
		primary := a.sort.Primary()
		return a.keyPrefix + primary.Field
	}
	// Single key: use the provided key
	return a.key
}

// HashCursorAdapter implements cursor-based pagination for Redis hashes
type HashCursorAdapter[T any] struct {
	client    *redis.Client
	keyPrefix string
	listKey   string // Sorted set for ordering
	unmarshal func(map[string]string) (T, error)
}

// NewHashCursorAdapter creates a new Redis hash cursor pagination adapter
func NewHashCursorAdapter[T any](client *redis.Client, keyPrefix, listKey string, unmarshal func(map[string]string) (T, error)) *HashCursorAdapter[T] {
	return &HashCursorAdapter[T]{
		client:    client,
		keyPrefix: keyPrefix,
		listKey:   listKey,
		unmarshal: unmarshal,
	}
}

// FetchWithCursor retrieves items for cursor-based pagination
func (a *HashCursorAdapter[T]) FetchWithCursor(ctx context.Context, cursor string, limit int) ([]T, string, error) {
	var minScore float64 = 0
	if cursor != "" {
		score, err := strconv.ParseFloat(cursor, 64)
		if err != nil {
			return nil, "", fmt.Errorf("invalid cursor: %w", err)
		}
		minScore = score
	}

	opt := &redis.ZRangeBy{
		Min:    fmt.Sprintf("(%f", minScore),
		Max:    "+inf",
		Offset: 0,
		Count:  int64(limit),
	}

	results, err := a.client.ZRangeByScoreWithScores(ctx, a.listKey, opt).Result()
	if err != nil {
		return nil, "", err
	}

	items := make([]T, 0, len(results))
	var nextCursor string

	for i, result := range results {
		if i == limit-1 {
			nextCursor = fmt.Sprintf("%f", result.Score)
			break
		}

		id := result.Member.(string)
		key := a.keyPrefix + ":" + id
		data, err := a.client.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, "", err
		}

		item, err := a.unmarshal(data)
		if err != nil {
			return nil, "", err
		}
		items = append(items, item)
	}

	return items, nextCursor, nil
}
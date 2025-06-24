# Pagit (Pagination Kit)

A simple yet powerful pagination toolkit for Go with support for multiple databases.

## Features

- **Simple API**: Just 2 main functions - `Paginate()` and `PaginateCursor()`
- **Database Agnostic**: Works with Redis, PostgreSQL, MongoDB, etc.
- **Type Safe**: Full generic support
- **Flexible**: Supports both offset-based and cursor-based pagination
- **Modular**: Use only what you need - offset or cursor independently
- **Zero Dependencies**: Only requires database drivers you're already using

## Installation

```bash
go get github.com/defense-allies/pagit
```

## Quick Start

### Offset-Based Pagination

```go
import (
    "github.com/defense-allies/pagit"
    pagitredis "github.com/defense-allies/pagit/adapters/pagit-redis"
)

// 1. Create an offset adapter
adapter := pagitredis.NewOffsetAdapter(redisClient, "users:list", pagitredis.UnmarshalJSON[User])

// 2. Create a pagination request
req := pagit.OffsetRequest{
    Page:     1,
    PageSize: 20,
}

// 3. Get paginated results
resp, err := pagit.Paginate(ctx, req, adapter)

// That's it! You now have:
// - resp.Items: Your paginated data
// - resp.Total: Total count
// - resp.HasNext/HasPrev: Navigation helpers
// - resp.TotalPages: Total number of pages
```

### Cursor-Based Pagination

For large datasets or real-time data:

```go
// 1. Create a cursor adapter
adapter := pagitredis.NewCursorAdapter(redisClient, "users:list", pagitredis.UnmarshalJSON[User])

// 2. First page
req := pagit.CursorRequest{PageSize: 20}
resp, _ := pagit.PaginateCursor(ctx, req, adapter)

// 3. Next page
req.Cursor = resp.NextCursor
resp, _ = pagit.PaginateCursor(ctx, req, adapter)
```

## Complete Independence

Each pagination type is completely independent:

- **Offset-only**: Import `pagit`, use `OffsetRequest`, `OffsetResponse`, `OffsetAdapter`
- **Cursor-only**: Import `pagit`, use `CursorRequest`, `CursorResponse`, `CursorAdapter`
- **Both**: Use both as needed

## Custom Adapters

Implement the simple adapter interfaces for any database:

### Offset Adapter
```go
type MyOffsetAdapter struct {
    db *MyDatabase
}

func (a *MyOffsetAdapter) Count(ctx context.Context) (int64, error) {
    return a.db.Count()
}

func (a *MyOffsetAdapter) Fetch(ctx context.Context, offset, limit int) ([]MyType, error) {
    return a.db.Query().Offset(offset).Limit(limit).Find()
}
```

### Cursor Adapter
```go
type MyCursorAdapter struct {
    db *MyDatabase
}

func (a *MyCursorAdapter) FetchWithCursor(ctx context.Context, cursor string, limit int) ([]MyType, string, error) {
    // Your cursor logic here
}
```

## Available Adapters

- **Redis** (`pagit-redis`): 
  - `pagitredis.NewOffsetAdapter()` for offset-based
  - `pagitredis.NewCursorAdapter()` for cursor-based
- **PostgreSQL** (`pagit-postgres`): Coming soon
- **MongoDB** (`pagit-mongo`): Coming soon
- **MySQL** (`pagit-mysql`): Coming soon

## Custom Options

```go
// Offset options
opts := pagit.OffsetOptions{
    DefaultPageSize: 50,
    MaxPageSize:     100,
    MinPageSize:     10,
}
resp, _ := pagit.Paginate(ctx, req, adapter, opts)

// Cursor options
opts := pagit.CursorOptions{
    DefaultPageSize: 50,
    MaxPageSize:     100,
    MinPageSize:     10,
}
resp, _ := pagit.PaginateCursor(ctx, req, adapter, opts)
```

## Why Pagit?

- **Minimal**: Only 2 functions to learn
- **Focused**: Each pagination type is independent
- **Extensible**: Easy to add new database adapters
- **Type-safe**: Full generic support
- **Production-ready**: Used in high-traffic applications

## License

MIT
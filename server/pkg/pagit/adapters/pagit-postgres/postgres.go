package pagitpostgres

import (
	"context"
	"database/sql"
	"fmt"
)

// Adapter implements pagination for PostgreSQL
type Adapter[T any] struct {
	db        *sql.DB
	table     string
	orderBy   string
	where     string
	args      []interface{}
	scanner   func(*sql.Rows) (T, error)
}

// NewAdapter creates a new PostgreSQL pagination adapter
func NewAdapter[T any](db *sql.DB, table, orderBy string, scanner func(*sql.Rows) (T, error)) *Adapter[T] {
	return &Adapter[T]{
		db:      db,
		table:   table,
		orderBy: orderBy,
		scanner: scanner,
		args:    []interface{}{},
	}
}

// WithWhere adds a WHERE clause to the query
func (a *Adapter[T]) WithWhere(where string, args ...interface{}) *Adapter[T] {
	a.where = where
	a.args = args
	return a
}

// Count returns the total number of items
func (a *Adapter[T]) Count(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", a.table)
	if a.where != "" {
		query += " WHERE " + a.where
	}

	var count int64
	err := a.db.QueryRowContext(ctx, query, a.args...).Scan(&count)
	return count, err
}

// Fetch retrieves items for offset-based pagination
func (a *Adapter[T]) Fetch(ctx context.Context, offset, limit int) ([]T, error) {
	query := fmt.Sprintf("SELECT * FROM %s", a.table)
	if a.where != "" {
		query += " WHERE " + a.where
	}
	query += fmt.Sprintf(" ORDER BY %s LIMIT %d OFFSET %d", a.orderBy, limit, offset)

	rows, err := a.db.QueryContext(ctx, query, a.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []T
	for rows.Next() {
		item, err := a.scanner(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// FetchWithCursor retrieves items for cursor-based pagination
func (a *Adapter[T]) FetchWithCursor(ctx context.Context, cursor string, limit int) ([]T, string, error) {
	query := fmt.Sprintf("SELECT * FROM %s", a.table)
	
	whereClause := a.where
	args := append([]interface{}{}, a.args...)
	
	if cursor != "" {
		cursorCondition := fmt.Sprintf("%s > $%d", a.orderBy, len(args)+1)
		if whereClause != "" {
			whereClause = fmt.Sprintf("(%s) AND %s", whereClause, cursorCondition)
		} else {
			whereClause = cursorCondition
		}
		args = append(args, cursor)
	}
	
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	query += fmt.Sprintf(" ORDER BY %s LIMIT %d", a.orderBy, limit)

	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var items []T
	var nextCursor string
	
	for rows.Next() {
		item, err := a.scanner(rows)
		if err != nil {
			return nil, "", err
		}
		items = append(items, item)
	}

	// If we got the full limit, there might be more items
	if len(items) == limit {
		// Get the cursor value from the last item
		// This would need to be customized based on the actual implementation
		nextCursor = "implement-based-on-your-needs"
	}

	return items, nextCursor, rows.Err()
}

// QueryBuilder provides a fluent interface for building complex queries
type QueryBuilder[T any] struct {
	adapter *Adapter[T]
	joins   []string
	selects string
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder[T any](db *sql.DB, table string, scanner func(*sql.Rows) (T, error)) *QueryBuilder[T] {
	return &QueryBuilder[T]{
		adapter: NewAdapter[T](db, table, "id", scanner),
		selects: "*",
	}
}

// Select specifies which columns to select
func (qb *QueryBuilder[T]) Select(columns string) *QueryBuilder[T] {
	qb.selects = columns
	return qb
}

// Join adds a JOIN clause
func (qb *QueryBuilder[T]) Join(join string) *QueryBuilder[T] {
	qb.joins = append(qb.joins, join)
	return qb
}

// Where adds a WHERE clause
func (qb *QueryBuilder[T]) Where(where string, args ...interface{}) *QueryBuilder[T] {
	qb.adapter.WithWhere(where, args...)
	return qb
}

// OrderBy sets the ORDER BY clause
func (qb *QueryBuilder[T]) OrderBy(orderBy string) *QueryBuilder[T] {
	qb.adapter.orderBy = orderBy
	return qb
}

// Build returns the configured adapter
func (qb *QueryBuilder[T]) Build() *Adapter[T] {
	return qb.adapter
}
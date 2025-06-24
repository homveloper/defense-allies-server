// Package pagit (Pagination Kit) provides simple yet powerful pagination toolkit
// with support for both offset-based and cursor-based pagination.
//
// This package is designed to be minimal and focused, providing only what you need:
//
// For offset-based pagination:
//   - Use OffsetRequest, OffsetResponse, OffsetAdapter, and Paginate()
//
// For cursor-based pagination:
//   - Use CursorRequest, CursorResponse, CursorAdapter, and PaginateCursor()
//
// Each approach is completely independent - you can use just offset-based,
// just cursor-based, or both depending on your needs.
package pagit
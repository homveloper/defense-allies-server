package user

import (
	"context"
	"sync"
	"time"
)

// LazyLoadable defines the interface for data that can be loaded lazily
type LazyLoadable interface {
	// IsLoaded returns true if the data has been loaded
	IsLoaded() bool
	
	// Load loads the data using the provided loader function
	Load(ctx context.Context, loader LoaderFunc) error
	
	// GetLoadedAt returns when the data was last loaded
	GetLoadedAt() *time.Time
	
	// ShouldReload determines if the data should be reloaded based on TTL
	ShouldReload(ttl time.Duration) bool
	
	// Reset marks the data as unloaded
	Reset()
}

// LoaderFunc is a function type that loads data for a specific key
type LoaderFunc func(ctx context.Context, key string) (any, error)

// LazyLoadableField represents a field that can be loaded lazily
type LazyLoadableField struct {
	key      string
	data     any
	loaded   bool
	loadedAt *time.Time
	mu       sync.RWMutex
}

// NewLazyLoadableField creates a new LazyLoadableField
func NewLazyLoadableField(key string) *LazyLoadableField {
	return &LazyLoadableField{
		key:    key,
		loaded: false,
	}
}

// IsLoaded returns true if the data has been loaded
func (f *LazyLoadableField) IsLoaded() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.loaded
}

// Load loads the data using the provided loader function
func (f *LazyLoadableField) Load(ctx context.Context, loader LoaderFunc) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	if f.loaded {
		return nil // Already loaded
	}
	
	data, err := loader(ctx, f.key)
	if err != nil {
		return err
	}
	
	now := time.Now()
	f.data = data
	f.loaded = true
	f.loadedAt = &now
	
	return nil
}

// GetData returns the loaded data (nil if not loaded)
func (f *LazyLoadableField) GetData() any {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	if !f.loaded {
		return nil
	}
	return f.data
}

// SetData sets the data directly (marks as loaded)
func (f *LazyLoadableField) SetData(data any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	now := time.Now()
	f.data = data
	f.loaded = true
	f.loadedAt = &now
}

// GetLoadedAt returns when the data was last loaded
func (f *LazyLoadableField) GetLoadedAt() *time.Time {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.loadedAt
}

// ShouldReload determines if the data should be reloaded based on TTL
func (f *LazyLoadableField) ShouldReload(ttl time.Duration) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	if !f.loaded || f.loadedAt == nil {
		return true
	}
	
	return time.Since(*f.loadedAt) > ttl
}

// Reset marks the data as unloaded
func (f *LazyLoadableField) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	f.data = nil
	f.loaded = false
	f.loadedAt = nil
}

// GetKey returns the key used for loading
func (f *LazyLoadableField) GetKey() string {
	return f.key
}

// LazyLoadableCollection represents a collection of lazy loadable fields
type LazyLoadableCollection struct {
	fields map[string]*LazyLoadableField
	mu     sync.RWMutex
}

// NewLazyLoadableCollection creates a new LazyLoadableCollection
func NewLazyLoadableCollection() *LazyLoadableCollection {
	return &LazyLoadableCollection{
		fields: make(map[string]*LazyLoadableField),
	}
}

// AddField adds a new lazy loadable field to the collection
func (c *LazyLoadableCollection) AddField(key string) *LazyLoadableField {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	field := NewLazyLoadableField(key)
	c.fields[key] = field
	return field
}

// GetField returns a lazy loadable field by key
func (c *LazyLoadableCollection) GetField(key string) (*LazyLoadableField, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	field, exists := c.fields[key]
	return field, exists
}

// GetOrCreateField gets an existing field or creates a new one
func (c *LazyLoadableCollection) GetOrCreateField(key string) *LazyLoadableField {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	field, exists := c.fields[key]
	if !exists {
		field = NewLazyLoadableField(key)
		c.fields[key] = field
	}
	return field
}

// LoadAll loads all fields in the collection
func (c *LazyLoadableCollection) LoadAll(ctx context.Context, loader LoaderFunc) error {
	c.mu.RLock()
	fields := make([]*LazyLoadableField, 0, len(c.fields))
	for _, field := range c.fields {
		fields = append(fields, field)
	}
	c.mu.RUnlock()
	
	for _, field := range fields {
		if err := field.Load(ctx, loader); err != nil {
			return err
		}
	}
	
	return nil
}

// ResetAll resets all fields in the collection
func (c *LazyLoadableCollection) ResetAll() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	for _, field := range c.fields {
		field.Reset()
	}
}

// GetLoadedFields returns all loaded fields
func (c *LazyLoadableCollection) GetLoadedFields() map[string]*LazyLoadableField {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	loaded := make(map[string]*LazyLoadableField)
	for key, field := range c.fields {
		if field.IsLoaded() {
			loaded[key] = field
		}
	}
	
	return loaded
}

// GetAllFields returns all fields (loaded and unloaded)
func (c *LazyLoadableCollection) GetAllFields() map[string]*LazyLoadableField {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	fields := make(map[string]*LazyLoadableField)
	for key, field := range c.fields {
		fields[key] = field
	}
	
	return fields
}
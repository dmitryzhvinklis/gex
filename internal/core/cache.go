package core

import (
	"sync"
	"time"
)

// CacheEntry represents a cached item
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// Cache provides high-performance caching for shell operations
type Cache struct {
	data  map[string]*CacheEntry
	mutex sync.RWMutex
	ttl   time.Duration
}

// NewCache creates a new cache with specified TTL
func NewCache(ttl time.Duration) *Cache {
	cache := &Cache{
		data: make(map[string]*CacheEntry),
		ttl:  ttl,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Set stores a value in the cache
func (c *Cache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Value, true
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, key)
}

// Clear removes all values from the cache
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data = make(map[string]*CacheEntry)
}

// Size returns the number of items in the cache
func (c *Cache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.data)
}

// cleanup removes expired entries
func (c *Cache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		for key, entry := range c.data {
			if now.After(entry.ExpiresAt) {
				delete(c.data, key)
			}
		}
		c.mutex.Unlock()
	}
}

// Global caches for different shell components
var (
	CommandCache    *Cache
	CompletionCache *Cache
	PathCache       *Cache
)

// InitializeCache initializes global caches
func InitializeCache() {
	CommandCache = NewCache(5 * time.Minute)     // Command results cache
	CompletionCache = NewCache(10 * time.Minute) // Tab completion cache
	PathCache = NewCache(30 * time.Minute)       // PATH lookup cache
}

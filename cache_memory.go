package qb

import (
	"encoding/json"
	"log"
	"reflect"
	"sync"
	"time"
)

// MemoryCache реализует кеш в памяти
type MemoryCache struct {
	data map[string]cacheItem
	mu   sync.RWMutex
	stop chan struct{} // Channel to signal cleanup goroutine to stop
}
type cacheItem struct {
	value      any
	expiration time.Time
}

func (q *QueryBuilder) SetMemoryCache() *MemoryCache {
	// If a cache already exists and it's a MemoryCache, stop its cleanup goroutine
	if oldCache, ok := q.cache.(*MemoryCache); ok && oldCache != nil {
		oldCache.StopCleanup()
	}

	cache := &MemoryCache{
		data: make(map[string]cacheItem),
		stop: make(chan struct{}),
	}
	go cache.startCleanup()
	q.cache = cache
	return cache
}

// StopCleanup signals the cleanup goroutine to stop.
func (c *MemoryCache) StopCleanup() {
	close(c.stop)
}

func (c *MemoryCache) Get(key string, dest any) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return false
	}

	if time.Now().After(item.expiration) {
		return false // Item expired, let cleanup goroutine handle deletion
	}

	// Directly copy the value if possible, otherwise use JSON serialization
	// as a fallback for deep copy.
	srcVal := reflect.ValueOf(item.value)
	destVal := reflect.ValueOf(dest)

	if destVal.Kind() != reflect.Ptr || destVal.IsNil() {
		log.Printf("Error: dest must be a non-nil pointer")
		return false
	}

	// If the types are directly assignable, we can do a much faster copy.
	if srcVal.Type().AssignableTo(destVal.Elem().Type()) {
		destVal.Elem().Set(srcVal)
		return true
	}

	// Fallback to JSON for deep copy if direct assignment is not possible.
	data, err := json.Marshal(item.value)
	if err != nil {
		log.Printf("Error marshalling cached value: %v", err)
		return false
	}
	if err := json.Unmarshal(data, dest); err != nil {
		log.Printf("Error unmarshalling cached value to dest: %v", err)
		return false
	}

	return true
}

func (c *MemoryCache) Set(key string, value any, expiration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(expiration),
	}
}

func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]cacheItem)
}

func (c *MemoryCache) startCleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for key, item := range c.data {
				if now.After(item.expiration) {
					delete(c.data, key)
				}
			}
			c.mu.Unlock()
		case <-c.stop:
			return // Stop the goroutine
		}
	}
}

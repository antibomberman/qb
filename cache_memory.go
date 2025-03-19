package qb

import (
	"sync"
	"time"
)

// MemoryCache реализует кеш в памяти
type MemoryCache struct {
	data map[string]cacheItem
	mu   sync.RWMutex
}
type cacheItem struct {
	value      any
	expiration time.Time
}

func (q *QueryBuilder) SetMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		data: make(map[string]cacheItem),
	}
	go cache.startCleanup()
	q.cache = cache
	return cache
}

func (c *MemoryCache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.expiration) {
		go c.Delete(key)
		return nil, false
	}

	return item.value, true
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

func (c *MemoryCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]cacheItem)
	return nil
}

func (c *MemoryCache) startCleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.data {
			if now.After(item.expiration) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

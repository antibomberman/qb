package dblayer

import "time"

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// Методы для работы с кешем
func (d *DBLayer) setCache(key string, value interface{}, duration time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()

	expiration := time.Time{}
	if duration > 0 {
		expiration = time.Now().Add(duration)
	}

	d.cache[key] = cacheItem{
		value:      value,
		expiration: expiration,
	}
}

func (d *DBLayer) getCache(key string) (interface{}, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	item, exists := d.cache[key]
	if !exists {
		return nil, false
	}

	if !item.expiration.IsZero() && item.expiration.Before(time.Now()) {
		delete(d.cache, key)
		return nil, false
	}

	return item.value, true
}

func (d *DBLayer) clearCache() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cache = make(map[string]cacheItem)
}

func (d *DBLayer) startCleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		d.cleanup()
	}
}

func (d *DBLayer) cleanup() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	for key, item := range d.cache {
		if !item.expiration.IsZero() && item.expiration.Before(now) {
			delete(d.cache, key)
		}
	}
}

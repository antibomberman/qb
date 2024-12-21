package dblayer

import (
	"encoding/json"
	"time"
)

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

// Remember включает кеширование для запроса
func (qb *QueryBuilder) Remember(duration time.Duration, key string) *QueryBuilder {
	qb.cacheKey = key
	qb.cacheDuration = duration
	return qb
}

// GetCached получает данные с учетом кеша
func (qb *QueryBuilder) GetCached(dest interface{}) (bool, error) {
	// Проверяем наличие ключа кеша
	if qb.cacheKey != "" {
		// Пытаемся получить из кеша
		if cached, ok := qb.dbl.getCache(qb.cacheKey); ok {
			// Копируем закешированные данные
			if data, ok := cached.([]byte); ok {
				return true, json.Unmarshal(data, dest)
			}
		}
	}

	// Если в кеше нет, получаем из БД
	found, err := qb.Get(dest)
	if err != nil {
		return false, err
	}

	// Сохраняем результат в кеш
	if found && qb.cacheKey != "" {
		// Сериализуем данные перед сохранением
		data, err := json.Marshal(dest)
		if err != nil {
			return false, err
		}
		qb.dbl.setCache(qb.cacheKey, data, qb.cacheDuration)
	}

	return found, nil
}

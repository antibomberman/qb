package dblayer

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

//add  optional driver // redis | memcached | map

type CacheDriver interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, expiration time.Duration)
	Delete(key string)
	Clear() error
}

// Remember включает кеширование для запроса
func (qb *QueryBuilder) Remember(key string, duration time.Duration) *QueryBuilder {
	qb.cacheKey = key
	qb.cacheDuration = duration
	return qb
}

// GetCached получает данные с учетом кеша
func (qb *QueryBuilder) GetCached(dest interface{}) (bool, error) {
	// Проверяем наличие ключа кеша
	if qb.cacheKey != "" {
		// Пытаемся получить из кеша
		if cached, ok := qb.dbl.cache.Get(qb.cacheKey); ok {
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
		qb.dbl.cache.Set(qb.cacheKey, data, qb.cacheDuration)
	}

	return found, nil
}

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// MemoryCache реализует кеш в памяти
type MemoryCache struct {
	data map[string]cacheItem
	mu   sync.RWMutex
}

func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		data: make(map[string]cacheItem),
	}
	go cache.startCleanup()
	return cache
}

func (c *MemoryCache) Get(key string) (interface{}, bool) {
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

func (c *MemoryCache) Set(key string, value interface{}, expiration time.Duration) {
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

// RedisCache реализует кеш в Redis

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisCache(addr string, password string, db int) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisCache{
		client: client,
		ctx:    context.Background(),
	}
}

func (c *RedisCache) Get(key string) (interface{}, bool) {
	val, err := c.client.Get(c.ctx, key).Result()
	if err != nil {
		return nil, false
	}

	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, false
	}

	return result, true
}

func (c *RedisCache) Set(key string, value interface{}, expiration time.Duration) {
	data, err := json.Marshal(value)
	if err != nil {
		return
	}

	c.client.Set(c.ctx, key, data, expiration)
}

func (c *RedisCache) Delete(key string) {
	c.client.Del(c.ctx, key)
}

func (c *RedisCache) Clear() error {
	return c.client.FlushDB(c.ctx).Err()
}

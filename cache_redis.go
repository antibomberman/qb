package qb

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

func (q *QueryBuilder) SetRedisCache(addr string, password string, db int) *RedisCache {
	ctx := context.TODO()
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	if client.Ping(ctx).Err() != nil {
		log.Fatal("redis ping failed")
	}
	rc := &RedisCache{
		client: client,
		ctx:    ctx,
	}
	q.cache = rc
	return rc
}

func (c *RedisCache) Get(key string) (any, bool) {
	val, err := c.client.Get(c.ctx, key).Result()
	if err != nil {
		return nil, false
	}

	var result any
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, false
	}

	return result, true
}

func (c *RedisCache) Set(key string, value any, expiration time.Duration) {
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

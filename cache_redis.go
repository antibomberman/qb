package qb

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

func (q *QueryBuilder) SetRedisCache(addr string, password string, db int) (*RedisCache, error) {
	ctx := context.TODO()
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	rc := &RedisCache{
		client: client,
		ctx:    ctx,
	}
	q.cache = rc
	return rc, nil
}

func (c *RedisCache) Get(key string, dest any) bool {
	val, err := c.client.Get(c.ctx, key).Result()
	if err != nil {
		return false
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		log.Printf("Error unmarshalling value from Redis cache: %v", err)
		return false
	}

	return true
}

func (c *RedisCache) Set(key string, value any, expiration time.Duration) {
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("Error marshalling value for Redis cache: %v", err)
		return
	}

	c.client.Set(c.ctx, key, data, expiration)
}

func (c *RedisCache) Delete(key string) {
	c.client.Del(c.ctx, key)
}

func (c *RedisCache) Clear() {
	err := c.client.FlushDB(c.ctx).Err()
	if err != nil {
		log.Printf("Error clearing Redis cache: %v", err)
	}
}

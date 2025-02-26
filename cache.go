package qb

import (
	"encoding/json"
	"time"
)

type CacheInterface interface {
	Get(key string) (any, bool)
	Set(key string, value any, expiration time.Duration)
	Delete(key string)
	Clear() error
}

func (q *QueryBuilder) SetCache(cache CacheInterface) {
	q.cache = cache
}
func (q *QueryBuilder) GetCache() CacheInterface {
	return q.cache
}

// Remember включает кеширование для запроса
func (qb *Builder) Remember(key string, duration time.Duration) *Builder {
	qb.cacheKey = key
	qb.cacheDuration = duration
	return qb
}

// GetCached получает данные с учетом кеша
func (qb *Builder) GetCached(dest any) (bool, error) {
	// Проверяем наличие ключа кеша
	if qb.cacheKey != "" {
		// Пытаемся получить из кеша
		if cached, ok := qb.queryBuilder.cache.Get(qb.cacheKey); ok {
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
		qb.queryBuilder.cache.Set(qb.cacheKey, data, qb.cacheDuration)
	}

	return found, nil
}

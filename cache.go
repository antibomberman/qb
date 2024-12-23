package dblayer

import (
	"encoding/json"
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

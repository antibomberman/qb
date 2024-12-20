package dblayer

import (
	"github.com/jmoiron/sqlx"
	"sync"
)

type DBLayer struct {
	db    *sqlx.DB
	cache map[string]cacheItem
	mu    sync.RWMutex
}

func NewDBLayer(db *sqlx.DB) *DBLayer {
	d := &DBLayer{
		db:    db,
		cache: make(map[string]cacheItem),
	}
	// Запускаем очистку кеша
	go d.startCleanup()
	return d
}

// Table теперь возвращает QueryBuilder с доступом к кешу
func (d *DBLayer) Table(name string) *QueryBuilder {
	return &QueryBuilder{
		table: name,
		db:    d.db,
		dbl:   d, // Добавляем ссылку на DBLayer
	}
}

// Table начинает построение запроса в транзакции
func (t *Transaction) Table(name string) *QueryBuilder {
	return &QueryBuilder{
		table: name,
		db:    t.tx,
	}
}

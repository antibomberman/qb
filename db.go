package dblayer

import (
	"context"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

type DBLayer struct {
	db    *sqlx.DB
	cache map[string]cacheItem
	mu    sync.RWMutex
}

type cacheItem struct {
	value      interface{}
	expiration time.Time
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

// Table теперь возвращает QueryBuilder с доступом к кешу
func (d *DBLayer) Table(name string) *QueryBuilder {
	return &QueryBuilder{
		table: name,
		db:    d.db,
		dbl:   d, // Добавляем ссылку на DBLayer
	}
}

// Transaction представляет транзакцию
type Transaction struct {
	tx *sqlx.Tx
}

// Begin начинает новую транзакцию
func (d *DBLayer) Begin() (*Transaction, error) {
	tx, err := d.db.Beginx()
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: tx}, nil
}

// BeginContext начинает новую транзакцию с контекстом
func (d *DBLayer) BeginContext(ctx context.Context) (*Transaction, error) {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: tx}, nil
}

// Transaction выполняет функцию в транзакции
func (d *DBLayer) Transaction(fn func(*Transaction) error) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// TransactionContext выполняет функцию в транзакции с контекстом
func (d *DBLayer) TransactionContext(ctx context.Context, fn func(*Transaction) error) error {
	tx, err := d.BeginContext(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Commit фиксирует транзакцию
func (t *Transaction) Commit() error {
	return t.tx.Commit()
}

// Rollback откатывает транзакцию
func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}

// Table начинает построение запроса в транзакции
func (t *Transaction) Table(name string) *QueryBuilder {
	return &QueryBuilder{
		table: name,
		db:    t.tx,
	}
}

package dblayer

import (
	"sync"

	"github.com/jmoiron/sqlx"
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

// Builder создает построитель таблиц
func (d *DBLayer) Builder(name string) *TableBuilder {
	return &TableBuilder{
		dbl:         d,
		name:        name,
		uniqueKeys:  make(map[string][]string),
		indexes:     make(map[string][]string),
		foreignKeys: make(map[string]*ForeignKey),
		engine:      "InnoDB",
		charset:     "utf8mb4",
		collate:     "utf8mb4_unicode_ci",
	}
}

// Alter создает построитель изменений таблиц
func (d *DBLayer) Alter(name string) *AlterTable {
	return &AlterTable{
		dbl:  d,
		name: name,
	}
}

// Truncate создает построитель очистки таблиц
func (d *DBLayer) Truncate(tables ...string) *TruncateTable {
	return &TruncateTable{
		dbl:    d,
		tables: tables,
	}
}

func (d *DBLayer) Drop(tables ...string) *DropTable {
	return &DropTable{
		dbl:    d,
		tables: tables,
	}
}

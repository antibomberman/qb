package dblayer

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

type DBLayer struct {
	db    *sqlx.DB
	cache map[string]cacheItem
	mu    sync.RWMutex
}

func New(driverName string, db *sql.DB) *DBLayer {
	x := sqlx.NewDb(db, driverName)
	d := &DBLayer{
		db:    x,
		cache: make(map[string]cacheItem),
	}
	go d.startCleanup()
	return d
}
func NewX(driverName string, dbx *sqlx.DB) *DBLayer {
	d := &DBLayer{
		db:    dbx,
		cache: make(map[string]cacheItem),
	}
	go d.startCleanup()
	return d
}

func Connect(driverName string, dataSourceName string) *DBLayer {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		panic(err)
	}
	return New(driverName, db)
}
func Connection(ctx context.Context, driverName string, dataSourceName string, maxAttempts int, connectionTimeout time.Duration) (*DBLayer, error) {

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		log.Printf("Attempting to connect to MySQL (attempt %d/%d)", attempt, maxAttempts)

		db, err := sql.Open(driverName, dataSourceName)
		if err != nil {
			log.Printf("Failed to open database connection: %v", err)
			time.Sleep(connectionTimeout)
			continue
		}
		ctx, cancel := context.WithTimeout(ctx, connectionTimeout)
		err = db.PingContext(ctx)
		cancel()

		if err == nil {
			log.Println("Connected to database")
			return New(driverName, db), nil
		}
		log.Printf("Failed to ping database: %v", err)
		db.Close()
		time.Sleep(connectionTimeout)
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts", maxAttempts)
}

func (d *DBLayer) Close() error {
	return d.db.Close()
}

// Table теперь возвращает QueryBuilder с доступом к кешу
func (d *DBLayer) Table(name string) *QueryBuilder {
	return &QueryBuilder{
		table: name,
		dbl:   d,
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

// Create создает новую таблицу
func (dbl *DBLayer) Create(name string, fn func(*Schema)) error {
	tb := &TableBuilder{
		dbl:         dbl,
		name:        name,
		uniqueKeys:  make(map[string][]string),
		indexes:     make(map[string][]string),
		foreignKeys: make(map[string]*ForeignKey),
		engine:      "InnoDB",
		charset:     "utf8mb4",
		collate:     "utf8mb4_unicode_ci",
	}

	schema := &Schema{table: tb}
	fn(schema)

	return dbl.Raw(tb.Build()).Exec()
}

// Update обновляет существующую таблицу
func (dbl *DBLayer) Update(name string, fn func(*Schema)) error {
	alter := &AlterTable{
		dbl:  dbl,
		name: name,
	}

	schema := &Schema{
		alter: alter,
		mode:  "update",
	}
	fn(schema)

	return alter.Execute()
}

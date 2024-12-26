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
	db           *sqlx.DB
	cache        CacheDriver
	mu           sync.RWMutex
	driverName   string
	dialect      BaseDialect
	queryDialect QueryDialect
	errorHandler ErrorHandler
}

type ErrorHandler interface {
	HandleError(err error) error
	WrapError(err error, msg string) error
}

func (d *DBLayer) SetDialect() {
	switch d.driverName {
	case "mysql":
		d.dialect = &MysqlDialect{}
	case "postgres":
		d.dialect = &MysqlDialect{}
	case "sqlite":
		d.dialect = &SqliteDialect{}
	}
}

func New(driverName string, db *sql.DB) *DBLayer {
	x := sqlx.NewDb(db, driverName)
	if x == nil {
		panic("Couldn't create database new method")
	}
	d := &DBLayer{
		db:         x,
		driverName: driverName,
	}

	return d
}
func NewX(driverName string, dbx *sqlx.DB) *DBLayer {
	if dbx == nil {
		panic("Couldn't create database newX method")
	}
	d := &DBLayer{
		db:         dbx,
		driverName: driverName,
	}
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
func (d *DBLayer) SeRedisCacheDriver(addr string, password string, db int) {
	d.cache = NewRedisCache(addr, password, db)
}
func (d *DBLayer) SeMemoryCacheDriver() {
	d.cache = NewMemoryCache()
}

// Table теперь возвращает QueryBuilder с доступом к кешу
func (d *DBLayer) Table(name string) *QueryBuilder {
	return &QueryBuilder{
		table: name,
		dbl:   d,
		db:    d.db,
	}
}

// Table начинает построение запроса в транзакции
func (t *Transaction) Table(name string) *QueryBuilder {
	return &QueryBuilder{
		table: name,
		db:    t.tx,
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
func (dbl *DBLayer) CreateTable(name string, fn func(*Schema)) error {
	schema := &Schema{
		dbl: dbl,
		definition: SchemaDefinition{
			name: name,
			mode: "create",
		},
	}

	fn(schema)

	return dbl.Raw(schema.BuildCreate()).Exec()
}

// Update обновляет существующую таблицу
func (dbl *DBLayer) UpdateTable(name string, fn func(*Schema)) error {
	schema := &Schema{
		dbl: dbl,
		definition: SchemaDefinition{
			name: name,
			mode: "update",
		},
	}

	fn(schema)

	return dbl.Raw(schema.BuildAlter()).Exec()
}

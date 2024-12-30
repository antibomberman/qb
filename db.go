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
	dialect      Dialect
	errorHandler ErrorHandler
}

type ErrorHandler interface {
	HandleError(err error) error
	WrapError(err error, msg string) error
}

func (d *DBLayer) setDialect() {
	switch d.driverName {
	case "mysql":
		d.dialect = &MysqlDialect{}
	case "postgres":
		d.dialect = &PostgresDialect{}
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
	d.setDialect()

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
	d.setDialect()

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
func (d *DBLayer) Ping() error {
	return d.db.Ping()
}
func (d *DBLayer) CacheRedisDriver(addr string, password string, db int) {
	d.cache = NewCacheRedis(addr, password, db)
}
func (d *DBLayer) CacheMemoryDriver() {
	d.cache = NewCacheMemory()
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
		dbl:   t.dbl,
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

func (dbl *DBLayer) CreateTable(name string, fn func(*Schema)) error {
	schema := &Schema{
		dbl: dbl,
		definition: SchemaDefinition{
			name: name,
			mode: "create",
			options: TableOptions{
				ifNotExists: false,
			},
			// Инициализируем все maps в constraints
			constraints: Constraints{
				primaryKey:  make([]string, 0),
				uniqueKeys:  make(map[string][]string),
				indexes:     make(map[string][]string),
				foreignKeys: make(map[string]*Foreign),
			},
		},
	}

	fn(schema)
	fmt.Println(schema.BuildCreate())
	return dbl.Raw(schema.BuildCreate()).Exec()
}

// Добавьте конструктор для Schema или измените CreateTable
func (dbl *DBLayer) CreateTableIfNotExists(name string, fn func(*Schema)) error {
	schema := &Schema{
		dbl: dbl,
		definition: SchemaDefinition{
			name: name,
			options: TableOptions{
				ifNotExists: true,
			},
			mode: "create",
			constraints: Constraints{
				primaryKey:  make([]string, 0),
				uniqueKeys:  make(map[string][]string),
				indexes:     make(map[string][]string),
				foreignKeys: make(map[string]*Foreign),
			},
		},
	}

	fn(schema)
	fmt.Println(schema.BuildCreate())
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
	fmt.Println(schema.BuildAlter())

	return dbl.Raw(schema.BuildAlter()).Exec()
}

//TODO  SHOW TABLES
//Mysql
//show table `SHOW TABLES LIKE 'users'`
//check table `DESCRIBE`
//

//postgresql
// show table `SELECT EXISTS ( SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users' )`
//check table `SELECT column_name, data_type, character_maximum_length FROM information_schema.columns WHERE table_name = 'users'``

// Create Aydit table
func (d *DBLayer) AuditTableCreate() error {
	err := d.CreateTableIfNotExists("audits", func(schema *Schema) {
		schema.ID()
		schema.String("table_name", 20)
		schema.BigInteger("record_id").Unsigned()
		schema.String("action", 10)
		schema.Json("old_data").Nullable()
		schema.Json("new_data").Nullable()
		schema.BigInteger("user_id").Unsigned()
		schema.Timestamps()
	})
	if err != nil {
		return err
	}

	return nil
}

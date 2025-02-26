package qb

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type DBInterface interface {
	sqlx.Ext
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	Beginx() (*sqlx.Tx, error)
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
}

type TxInterface interface {
	sqlx.Ext
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	Commit() error
	Rollback() error
}

type QueryBuilder struct {
	db         DBInterface
	driverName string
	cache      CacheInterface
}

func New(driverName string, db *sql.DB) *QueryBuilder {
	return NewX(driverName, sqlx.NewDb(db, driverName))
}
func NewX(driverName string, db *sqlx.DB) *QueryBuilder {
	return &QueryBuilder{
		db:         db,
		driverName: driverName,
		cache:      NewCacheMemory(),
	}
}

func (q *QueryBuilder) Query(table string) *Builder {
	return &Builder{
		tableName:    table,
		db:           q.db,
		queryBuilder: q,
		ctx:          context.TODO(),
	}
}

func (t *Transaction) Query(table string) *Builder {
	return &Builder{
		tableName:    table,
		db:           t.Tx,
		queryBuilder: t.QueryBuilder,
		ctx:          context.TODO(),
	}
}

func (t *Transaction) Raw(query string, args ...any) *RawQuery {
	return &RawQuery{
		query: query,
		args:  args,
		db:    t.Tx,
	}
}

func (q *QueryBuilder) GetDB() DBInterface {
	return q.db
}

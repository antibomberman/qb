package qb

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type QueryBuilder struct {
	db         DBInterface
	driverName string
	cache      CacheInterface
	logger     *slog.Logger
}

func New(driverName string, db *sql.DB) QueryBuilderInterface {
	return NewX(driverName, sqlx.NewDb(db, driverName))
}
func NewX(driverName string, db *sqlx.DB) QueryBuilderInterface {
	return &QueryBuilder{
		db:         db,
		driverName: driverName,
		cache:      NewCacheMemory(),
		logger:     NewDefaultLogger(),
	}
}

func (q *QueryBuilder) From(table string) BuilderInterface {
	return &Builder{
		tableName:    table,
		db:           q.db,
		queryBuilder: q,
		ctx:          context.TODO(),
	}
}

func (t *Transaction) From(table string) BuilderInterface {
	return &Builder{
		tableName:    table,
		db:           t.Tx,
		queryBuilder: t.QueryBuilder,
		ctx:          context.TODO(),
	}
}

func (t *Transaction) Raw(query string, args ...any) *RawQuery {
	return &RawQuery{
		query:        query,
		args:         args,
		db:           t.Tx,
		queryBuilder: t.QueryBuilder,
	}
}

func (q *QueryBuilder) GetDB() DBInterface {
	return q.db
}

func (q *QueryBuilder) SetLogger(logger *slog.Logger) {
	q.logger = logger
}

func (q *QueryBuilder) SetCache(cache CacheInterface) {
	q.cache = cache
}
func (q *QueryBuilder) GetCache() CacheInterface {
	return q.cache
}

package dblayer

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type Condition struct {
	operator string
	clause   string
	nested   []Condition
	args     []interface{}
}

type QueryBuilder struct {
	table         string
	conditions    []Condition
	db            interface{} // может быть *sqlx.DB или *sqlx.Tx
	columns       []string
	orderBy       []string
	groupBy       []string
	having        string
	limit         int
	offset        int
	joins         []Join
	alias         string
	dbl           *DBLayer
	cacheKey      string
	cacheDuration time.Duration
}

// Executor интерфейс для выполнения запросов
type Executor interface {
	sqlx.Ext
	sqlx.ExtContext
	DriverName() string
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	NamedExec(query string, arg interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

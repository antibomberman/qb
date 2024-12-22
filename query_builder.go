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
	db            interface{} // может быть *sqlx.DB или *sqlx.Tx
	dbl           *DBLayer
	table         string
	conditions    []Condition
	columns       []string
	orderBy       []string
	groupBy       []string
	having        string
	limit         int
	offset        int
	joins         []Join
	alias         string
	cacheKey      string
	cacheDuration time.Duration
	events        map[EventType][]EventHandler
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

// On регистрирует обработчик события
func (qb *QueryBuilder) On(event EventType, handler EventHandler) {
	if qb.events == nil {
		qb.events = make(map[EventType][]EventHandler)
	}
	qb.events[event] = append(qb.events[event], handler)
}

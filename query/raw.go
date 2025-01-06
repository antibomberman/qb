package query

import (
	"github.com/jmoiron/sqlx"
)

// Raw выполняет сырой SQL-запрос
func (q *QueryBuilder) Raw(query string, args ...interface{}) *RawQuery {
	return &RawQuery{
		query: query,
		args:  args,
		db:    q.db,
	}
}

// RawQuery представляет сырой SQL-запрос
type RawQuery struct {
	query string
	args  []interface{}
	db    sqlx.Ext
}

// Exec выполняет запрос без возврата результатов
func (r *RawQuery) Exec() error {
	_, err := r.db.Exec(r.query, r.args...)
	return err
}

// Query выполняет запрос и сканирует результаты в slice
func (r *RawQuery) Query(dest interface{}) error {
	return sqlx.Select(r.db, dest, r.query, r.args...)
}

// QueryRow выполняет запрос и сканирует один результат
func (r *RawQuery) QueryRow(dest interface{}) error {
	return sqlx.Get(r.db, dest, r.query, r.args...)
}

package qb

import (
	"context"
	"time"
)

// Raw выполняет сырой SQL-запрос
func (q *QueryBuilder) Raw(query string, args ...any) *RawQuery {
	return &RawQuery{
		query:        query,
		args:         args,
		db:           q.db,
		queryBuilder: q,
		ctx:          context.TODO(), // Default context
	}
}

// RawQuery представляет сырой SQL-запрос
type RawQuery struct {
	query        string
	args         []any
	db           Executor // Use Executor interface
	queryBuilder *QueryBuilder
	ctx          context.Context // Add context field
}

// Exec выполняет запрос без возврата результатов
func (r *RawQuery) Exec() error {
	start := time.Now()
	_, err := r.db.ExecContext(r.ctx, r.query, r.args...)
	r.queryBuilder.Debug("RawQuery.Exec", start, r.query, r.args)
	if err != nil {
		r.queryBuilder.Error(err.Error(), start, r.query, r.args)
	}
	return err
}

// Query выполняет запрос и сканирует результаты в slice
func (r *RawQuery) Query(dest any) error {
	start := time.Now()
	err := r.db.SelectContext(r.ctx, dest, r.query, r.args...)
	r.queryBuilder.Debug("RawQuery.Query", start, r.query, r.args)
	if err != nil {
		r.queryBuilder.Error(err.Error(), start, r.query, r.args)
	}
	return err
}

// QueryRow выполняет запрос и сканирует один результат
func (r *RawQuery) QueryRow(dest any) error {
	start := time.Now()
	err := r.db.GetContext(r.ctx, dest, r.query, r.args...)
	r.queryBuilder.Debug("RawQuery.QueryRow", start, r.query, r.args)
	if err != nil {
		r.queryBuilder.Error(err.Error(), start, r.query, r.args)
	}
	return err
}

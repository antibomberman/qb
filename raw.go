package qb

import "time"

// Raw выполняет сырой SQL-запрос
func (q *QueryBuilder) Raw(query string, args ...any) *RawQuery {
	return &RawQuery{
		query:        query,
		args:         args,
		db:           q.db,
		queryBuilder: q,
	}
}

// RawQuery представляет сырой SQL-запрос
type RawQuery struct {
	query        string
	args         []any
	db           Executor
	queryBuilder *QueryBuilder
}

// Exec выполняет запрос без возврата результатов
func (r *RawQuery) Exec() error {
	start := time.Now()
	r.queryBuilder.logger.Debug(time.Since(start).String(), r.query, r.args)
	_, err := r.db.Exec(r.query, r.args...)
	if err != nil {
		r.queryBuilder.Error(start, r.query, r.args)
	}
	return err
}

// Query выполняет запрос и сканирует результаты в slice
func (r *RawQuery) Query(dest any) error {
	start := time.Now()
	r.queryBuilder.logger.Debug(time.Since(start).String(), r.query, r.args)
	err := r.db.Select(dest, r.query, r.args...)
	if err != nil {
		r.queryBuilder.logger.Error(time.Since(start).String(), r.query, r.args)
	}
	return err
}

// QueryRow выполняет запрос и сканирует один результат
func (r *RawQuery) QueryRow(dest any) error {
	start := time.Now()
	r.queryBuilder.Debug(start, r.query, r.args)
	err := r.db.Get(dest, r.query, r.args...)
	if err != nil {
		r.queryBuilder.Error(start, r.query, r.args)
	}
	return err
}

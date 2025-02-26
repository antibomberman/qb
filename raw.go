package qb

// Raw выполняет сырой SQL-запрос
func (q *QueryBuilder) Raw(query string, args ...any) *RawQuery {
	return &RawQuery{
		query: query,
		args:  args,
		db:    q.db,
	}
}

// RawQuery представляет сырой SQL-запрос
type RawQuery struct {
	query string
	args  []any
	db    Executor
}

// Exec выполняет запрос без возврата результатов
func (r *RawQuery) Exec() error {
	r.db.(interface{ GetLogger() Logger }).GetLogger().Debug(r.query, r.args...)
	_, err := r.db.Exec(r.query, r.args...)
	if err != nil {
		r.db.(interface{ GetLogger() Logger }).GetLogger().Error(r.query, r.args...)
	}
	return err
}

// Query выполняет запрос и сканирует результаты в slice
func (r *RawQuery) Query(dest any) error {
	r.db.(interface{ GetLogger() Logger }).GetLogger().Debug(r.query, r.args...)
	err := r.db.Select(dest, r.query, r.args...)
	if err != nil {
		r.db.(interface{ GetLogger() Logger }).GetLogger().Error(r.query, r.args...)
	}
	return err
}

// QueryRow выполняет запрос и сканирует один результат
func (r *RawQuery) QueryRow(dest any) error {
	r.db.(interface{ GetLogger() Logger }).GetLogger().Debug(r.query, r.args...)
	err := r.db.Get(dest, r.query, r.args...)
	if err != nil {
		r.db.(interface{ GetLogger() Logger }).GetLogger().Error(r.query, r.args...)
	}
	return err
}

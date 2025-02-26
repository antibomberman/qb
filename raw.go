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
	_, err := r.db.Exec(r.query, r.args...)
	return err
}

// Query выполняет запрос и сканирует результаты в slice
func (r *RawQuery) Query(dest any) error {
	return r.db.Select(dest, r.query, r.args...)
}

// QueryRow выполняет запрос и сканирует один результат
func (r *RawQuery) QueryRow(dest any) error {
	return r.db.Get(dest, r.query, r.args...)
}

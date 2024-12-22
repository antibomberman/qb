package dblayer

// Versioning добавляет поддержку версионирования
type Versioning struct {
	Version int `db:"version"`
}

// WithVersion добавляет оптимистичную блокировку
func (qb *QueryBuilder) WithVersion(version int) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   "version = ?",
		args:     []interface{}{version},
	})
	return qb
}

// IncrementVersion увеличивает версию записи
func (qb *QueryBuilder) IncrementVersion() error {
	return qb.UpdateMap(map[string]interface{}{
		"version": qb.Raw("version + 1"),
	})
}

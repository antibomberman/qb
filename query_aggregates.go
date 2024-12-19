package dblayer

import (
	"context"
	"fmt"
	"strings"
)

// Avg вычисляет среднее значение колонки
func (qb *QueryBuilder) Avg(column string) (float64, error) {
	var result float64
	query := fmt.Sprintf("SELECT AVG(%s) FROM %s", column, qb.table)
	if len(qb.conditions) > 0 {
		whereSQL := buildConditions(qb.conditions)
		query += " WHERE " + whereSQL
		_, err := qb.execGet(&result, query)
		return result, err
	}
	_, err := qb.execGet(&result, query)
	return result, err
}

// Sum вычисляет сумму значений колонки
func (qb *QueryBuilder) Sum(column string) (float64, error) {
	var result float64
	query := fmt.Sprintf("SELECT SUM(%s) FROM %s", column, qb.table)
	if len(qb.conditions) > 0 {
		whereSQL := buildConditions(qb.conditions)
		query += " WHERE " + whereSQL
		_, err := qb.execGet(&result, query)
		return result, err
	}
	_, err := qb.execGet(&result, query)
	return result, err
}

// Min находит минимальное значение колонки
func (qb *QueryBuilder) Min(column string) (float64, error) {
	var result float64
	query := fmt.Sprintf("SELECT MIN(%s) FROM %s", column, qb.table)
	if len(qb.conditions) > 0 {
		whereSQL := buildConditions(qb.conditions)
		query += " WHERE " + whereSQL
		_, err := qb.execGet(&result, query)
		return result, err
	}
	_, err := qb.execGet(&result, query)
	return result, err
}

// Max находит максимальное значение колонки
func (qb *QueryBuilder) Max(column string) (float64, error) {
	var result float64
	query := fmt.Sprintf("SELECT MAX(%s) FROM %s", column, qb.table)
	if len(qb.conditions) > 0 {
		whereSQL := buildConditions(qb.conditions)
		query += " WHERE " + whereSQL
		_, err := qb.execGet(&result, query)
		return result, err
	}
	_, err := qb.execGet(&result, query)
	return result, err
}

// Pluck получает значения одной колонки
func (qb *QueryBuilder) Pluck(column string, dest interface{}) error {
	query := fmt.Sprintf("SELECT %s FROM %s", column, qb.table)
	if len(qb.conditions) > 0 {
		whereSQL := buildConditions(qb.conditions)
		query += " WHERE " + whereSQL
		_, err := qb.execSelect(dest, query)
		return err
	}
	_, err := qb.execSelect(dest, query)
	return err
}

// Chunk обрабатывает записи чанками
func (qb *QueryBuilder) Chunk(size int, fn func(items interface{}) error) error {
	offset := 0
	for {
		dest := make([]map[string]interface{}, 0, size)

		query, args := qb.buildQuery()
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", size, offset)

		found, err := qb.execSelect(&dest, query, args...)
		if err != nil {
			return err
		}

		if !found || len(dest) == 0 {
			break
		}

		if err := fn(dest); err != nil {
			return err
		}

		offset += size
	}
	return nil
}

// ChunkContext обрабатывает записи чанками с контекстом
func (qb *QueryBuilder) ChunkContext(ctx context.Context, size int, fn func(context.Context, interface{}) error) error {
	offset := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		dest := make([]map[string]interface{}, 0, size)

		query, args := qb.buildQuery()
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", size, offset)

		found, err := qb.execSelectContext(ctx, &dest, query, args...)
		if err != nil {
			return err
		}

		if !found || len(dest) == 0 {
			break
		}

		if err := fn(ctx, dest); err != nil {
			return err
		}

		offset += size
	}
	return nil
}

// WithinGroup выполняет оконную функцию
func (qb *QueryBuilder) WithinGroup(column string, window string) *QueryBuilder {
	qb.columns = append(qb.columns, fmt.Sprintf("%s WITHIN GROUP (%s)", column, window))
	return qb
}

// Distinct добавляет DISTINCT к запросу
func (qb *QueryBuilder) Distinct(columns ...string) *QueryBuilder {
	if len(columns) == 0 {
		qb.columns = append(qb.columns, "DISTINCT *")
	} else {
		qb.columns = append(qb.columns, "DISTINCT "+strings.Join(columns, ", "))
	}
	return qb
}

// Lock блокирует записи для обновления
func (qb *QueryBuilder) Lock(mode string) *QueryBuilder {
	qb.columns = append(qb.columns, mode)
	return qb
}

// Raw выполняет сырой SQL запрос
func (qb *QueryBuilder) Raw(query string, args ...interface{}) error {
	return qb.execExec(query, args...)
}

// RawQuery выполняет сырой SQL запрос с возвратом данных
func (qb *QueryBuilder) RawQuery(dest interface{}, query string, args ...interface{}) error {
	_, err := qb.execSelect(dest, query, args...)
	return err
}

// Value получает значение одного поля
func (qb *QueryBuilder) Value(column string) (interface{}, error) {
	var result interface{}
	query := fmt.Sprintf("SELECT %s FROM %s", column, qb.table)
	qb.Limit(1)
	if len(qb.conditions) > 0 {
		whereSQL := buildConditions(qb.conditions)
		query += " WHERE " + whereSQL
		query = qb.rebindQuery(query)
		_, err := qb.execGet(&result, query)
		return result, err
	}
	query = qb.rebindQuery(query)
	_, err := qb.execGet(&result, query)
	return result, err
}

// Values получает значения одного поля для всех записей
func (qb *QueryBuilder) Values(column string) ([]interface{}, error) {
	var result []interface{}
	query := fmt.Sprintf("SELECT %s FROM %s", column, qb.table)
	if len(qb.conditions) > 0 {
		whereSQL := buildConditions(qb.conditions)
		query += " WHERE " + whereSQL
		query = qb.rebindQuery(query)
		_, err := qb.execSelect(&result, query)
		return result, err
	}
	query = qb.rebindQuery(query)
	_, err := qb.execSelect(&result, query)
	return result, err
}

// WhereExists добавляет условие EXISTS
func (qb *QueryBuilder) WhereExists(subQuery *QueryBuilder) *QueryBuilder {
	sql, args := subQuery.buildQuery()
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("EXISTS (%s)", sql),
		args:     args,
	})
	return qb
}

// WhereNotExists добавляет условие NOT EXISTS
func (qb *QueryBuilder) WhereNotExists(subQuery *QueryBuilder) *QueryBuilder {
	sql, args := subQuery.buildQuery()
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("NOT EXISTS (%s)", sql),
		args:     args,
	})
	return qb
}

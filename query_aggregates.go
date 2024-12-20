package dblayer

import (
	"fmt"
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

// Count возвращает количество записей
func (qb *QueryBuilder) Count() (int64, error) {
	var count int64
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", qb.table)
	if len(qb.conditions) > 0 {
		whereSQL := buildConditions(qb.conditions)
		query += " WHERE " + whereSQL
		query = qb.rebindQuery(query)
		_, err := qb.execGet(&count, query)
		return count, err
	}
	_, err := qb.execGet(&count, query)
	return count, err
}

// Exists проверяет существование записей
func (qb *QueryBuilder) Exists() (bool, error) {
	count, err := qb.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

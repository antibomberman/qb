package dblayer

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"sort"
	"strings"
)

// Create создает новую запись из структуры и возвращает её id
func (qb *QueryBuilder) Create(data interface{}, fields ...string) (int64, error) {
	var insertFields, placeholders []string

	if len(fields) > 0 {
		// Используем только указанные поля
		insertFields = fields
		placeholders = make([]string, len(fields))
		for i, field := range fields {
			placeholders[i] = ":" + field
		}
	} else {
		// Используем все поля из структуры
		insertFields, placeholders = qb.getStructInfo(data)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		qb.table,
		strings.Join(insertFields, ", "),
		strings.Join(placeholders, ", "))

	if qb.getDriverName() == "postgres" {
		var id int64
		query += " RETURNING id"
		err := qb.getExecutor().QueryRowx(query, data).Scan(&id)
		return id, err
	}

	result, err := qb.getExecutor().NamedExec(query, data)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// CreateContext создает новую запись из структуры и возвращает её id
func (qb *QueryBuilder) CreateContext(ctx context.Context, data interface{}, fields ...string) (int64, error) {
	var insertFields, placeholders []string

	if len(fields) > 0 {
		// Используем только указанные поля
		insertFields = fields
		placeholders = make([]string, len(fields))
		for i, field := range fields {
			placeholders[i] = ":" + field
		}
	} else {
		// Используем все поля из структуры
		insertFields, placeholders = qb.getStructInfo(data)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		qb.table,
		strings.Join(insertFields, ", "),
		strings.Join(placeholders, ", "))

	if qb.getDriverName() == "postgres" {
		var id int64
		query += " RETURNING id"
		err := qb.getExecutor().QueryRowx(query, data).Scan(&id)
		return id, err
	}

	result, err := qb.getExecutor().NamedExecContext(ctx, query, data)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// CreateMap создает новую запись из map и возвращает её id
func (qb *QueryBuilder) CreateMap(data map[string]interface{}) (int64, error) {
	columns := make([]string, 0)
	placeholders := make([]string, 0)
	values := make([]interface{}, 0)

	// Используем все поля из map
	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		values = append(values, val)
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		qb.table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	if qb.getDriverName() == "postgres" {
		var id int64
		query = qb.rebindQuery(query + " RETURNING id")
		err := qb.getExecutor().QueryRowx(query, values...).Scan(&id)
		return id, err
	}

	result, err := qb.getExecutor().Exec(qb.rebindQuery(query), values...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// CreateMapContext создает новую запись из map с контекстом и возвращает её id
func (qb *QueryBuilder) CreateMapContext(ctx context.Context, data map[string]interface{}) (int64, error) {
	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		values = append(values, val)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		qb.table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	if qb.getDriverName() == "postgres" {
		var id int64
		query = qb.rebindQuery(query + " RETURNING id")
		err := qb.getExecutor().(sqlx.QueryerContext).QueryRowxContext(ctx, query, values...).Scan(&id)
		return id, err
	}

	result, err := qb.getExecutor().ExecContext(ctx, qb.rebindQuery(query), values...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// BatchInsert вставляет множество записей
func (qb *QueryBuilder) BatchInsert(records []map[string]interface{}) error {
	if len(records) == 0 {
		return nil
	}

	// Получаем все колонки из первой записи
	columns := make([]string, 0)
	for column := range records[0] {
		columns = append(columns, column)
	}
	sort.Strings(columns)

	// Создаем placeholders и значения
	var placeholders []string
	var values []interface{}
	for _, record := range records {
		placeholder := make([]string, len(columns))
		for i := range columns {
			placeholder[i] = "?"
			values = append(values, record[columns[i]])
		}
		placeholders = append(placeholders, "("+strings.Join(placeholder, ", ")+")")
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		qb.table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	return qb.execExec(query, values...)
}

// BatchInsertContext вставляет множество записей с контекстом
func (qb *QueryBuilder) BatchInsertContext(ctx context.Context, records []map[string]interface{}) error {
	if len(records) == 0 {
		return nil
	}

	// Получаем все колонки из первой записи
	columns := make([]string, 0)
	for column := range records[0] {
		columns = append(columns, column)
	}
	sort.Strings(columns)

	// Создаем placeholders и значения
	var placeholders []string
	var values []interface{}
	for _, record := range records {
		placeholder := make([]string, len(columns))
		for i := range columns {
			placeholder[i] = "?"
			values = append(values, record[columns[i]])
		}
		placeholders = append(placeholders, "("+strings.Join(placeholder, ", ")+")")
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		qb.table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	return qb.execExecContext(ctx, query, values...)
}

// BulkInsert выполняет массовую вставку записей с возвратом ID
func (qb *QueryBuilder) BulkInsert(records []map[string]interface{}) ([]int64, error) {
	if len(records) == 0 {
		return nil, nil
	}

	// Получаем все колонки из первой записи
	columns := make([]string, 0)
	for column := range records[0] {
		columns = append(columns, column)
	}
	sort.Strings(columns)

	// Создаем placeholders и значения
	var placeholders []string
	var values []interface{}
	for _, record := range records {
		placeholder := make([]string, len(columns))
		for i := range columns {
			placeholder[i] = "?"
			values = append(values, record[columns[i]])
		}
		placeholders = append(placeholders, "("+strings.Join(placeholder, ", ")+")")
	}

	var query string
	if qb.getDriverName() == "postgres" {
		query = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES %s RETURNING id",
			qb.table,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "),
		)
		var ids []int64
		err := qb.getExecutor().Select(&ids, qb.rebindQuery(query), values...)
		return ids, err
	}

	query = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		qb.table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	result, err := qb.getExecutor().Exec(qb.rebindQuery(query), values...)
	if err != nil {
		return nil, err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	ids := make([]int64, rowsAffected)
	for i := range ids {
		ids[i] = lastID + int64(i)
	}

	return ids, nil
}

// BulkInsertContext выполняет массовую вставку записей с возвратом ID и поддержкой контекста
func (qb *QueryBuilder) BulkInsertContext(ctx context.Context, records []map[string]interface{}) ([]int64, error) {
	if len(records) == 0 {
		return nil, nil
	}

	// Получаем все колонки из первой записи
	columns := make([]string, 0)
	for column := range records[0] {
		columns = append(columns, column)
	}
	sort.Strings(columns)

	// Создаем placeholders и значения
	var placeholders []string
	var values []interface{}
	for _, record := range records {
		placeholder := make([]string, len(columns))
		for i := range columns {
			placeholder[i] = "?"
			values = append(values, record[columns[i]])
		}
		placeholders = append(placeholders, "("+strings.Join(placeholder, ", ")+")")
	}

	var query string
	if qb.getDriverName() == "postgres" {
		query = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES %s RETURNING id",
			qb.table,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "),
		)
		var ids []int64
		err := qb.getExecutor().SelectContext(ctx, &ids, qb.rebindQuery(query), values...)
		return ids, err
	}

	query = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		qb.table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	result, err := qb.getExecutor().(sqlx.ExtContext).ExecContext(ctx, qb.rebindQuery(query), values...)
	if err != nil {
		return nil, err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	ids := make([]int64, rowsAffected)
	for i := range ids {
		ids[i] = lastID + int64(i)
	}

	return ids, nil
}

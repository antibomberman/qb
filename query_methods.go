package dblayer

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// DateFunctions содержит SQL функции для разных СУБД
type DateFunctions struct {
	DateDiff    string
	DateTrunc   string
	DateFormat  string
	TimeZone    string
	Extract     string
	DateAdd     string
	CurrentDate string
}

// Select указывает колонки для выборки
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.columns = columns
	return qb
}

// Where добавляет условие AND
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   condition,
		args:     args,
	})
	return qb
}

// WhereId добавляет условие WHERE id = ?
func (qb *QueryBuilder) WhereId(id interface{}) *QueryBuilder {
	qb.Where("id = ?", id)
	return qb
}

// OrWhere добавляет условие OR
func (qb *QueryBuilder) OrWhere(condition string, args ...interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "OR",
		clause:   condition,
		args:     args,
	})
	return qb
}

// WhereIn добавляет условие IN
func (qb *QueryBuilder) WhereIn(column string, values ...interface{}) *QueryBuilder {
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = "?"
	}
	condition := fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ","))
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   condition,
		args:     values,
	})
	return qb
}

// WhereGroup добавляет группу условий
func (qb *QueryBuilder) WhereGroup(fn func(*QueryBuilder)) *QueryBuilder {
	group := &QueryBuilder{}
	fn(group)
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		nested:   group.conditions,
	})
	return qb
}

// OrWhereGroup добавляет группу условий через OR
func (qb *QueryBuilder) OrWhereGroup(fn func(*QueryBuilder)) *QueryBuilder {
	group := &QueryBuilder{}
	fn(group)
	qb.conditions = append(qb.conditions, Condition{
		operator: "OR",
		nested:   group.conditions,
	})
	return qb
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

// OrderBy добавляет сортировку
func (qb *QueryBuilder) OrderBy(column string, direction string) *QueryBuilder {
	qb.orderBy = append(qb.orderBy, fmt.Sprintf("%s %s", column, direction))
	return qb
}

// GroupBy добавляет группировку
func (qb *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	qb.groupBy = columns
	return qb
}

// Having добавляет условие для группировки
func (qb *QueryBuilder) Having(condition string) *QueryBuilder {
	qb.having = condition
	return qb
}

// Limit устанавливает ограничение на количество записей
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset устанавливает смещение
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// As устанавливает алиас для таблицы
func (qb *QueryBuilder) As(alias string) *QueryBuilder {
	qb.alias = alias
	return qb
}

// Find ищет запись по id
func (qb *QueryBuilder) Find(id interface{}, dest interface{}) (bool, error) {
	qb.Where("id = ?", id)
	return qb.First(dest)
}

// FindContext ищет запись по id с контекстом
func (qb *QueryBuilder) FindContext(ctx context.Context, id interface{}, dest interface{}) (bool, error) {
	qb.Where("id = ?", id)
	return qb.FirstContext(ctx, dest)
}

// Increment увеличивает значение поля
func (qb *QueryBuilder) Increment(column string, value interface{}) error {
	var args []interface{}
	args = append(args, value)

	query := fmt.Sprintf("UPDATE %s SET %s = %s + ?", qb.table, column, column)

	if len(qb.conditions) > 0 {
		whereSQL := buildConditions(qb.conditions)
		query += " WHERE " + whereSQL
		// Добавляем аргументы из условий WHERE
		for _, cond := range qb.conditions {
			args = append(args, cond.args...)
		}
	}

	return qb.execExec(query, args...)
}

// Decrement уменьшает значение поля
func (qb *QueryBuilder) Decrement(column string, value interface{}) error {
	var args []interface{}
	args = append(args, value)

	query := fmt.Sprintf("UPDATE %s SET %s = %s - ?", qb.table, column, column)

	if len(qb.conditions) > 0 {
		whereSQL := buildConditions(qb.conditions)
		query += " WHERE " + whereSQL
		// Добавляем аргументы из условий WHERE
		for _, cond := range qb.conditions {
			args = append(args, cond.args...)
		}
	}

	return qb.execExec(query, args...)
}

// Get получает все записи
func (qb *QueryBuilder) Get(dest interface{}) (bool, error) {
	query, args := qb.buildQuery()
	return qb.execSelect(dest, query, args...)
}

// GetContext получает все записи с контекстом
func (qb *QueryBuilder) GetContext(ctx context.Context, dest interface{}) (bool, error) {
	query, args := qb.buildQuery()
	return qb.execSelectContext(ctx, dest, query, args...)
}

// First получает первую запись
func (qb *QueryBuilder) First(dest interface{}) (bool, error) {
	qb.Limit(1)
	query, args := qb.buildQuery()
	return qb.execGet(dest, query, args...)
}

// FirstContext получает первую запись с контекстом
func (qb *QueryBuilder) FirstContext(ctx context.Context, dest interface{}) (bool, error) {
	qb.Limit(1)
	query, args := qb.buildQuery()
	return qb.execGetContext(ctx, dest, query, args...)
}

// Delete удаляет записи
func (qb *QueryBuilder) Delete() error {
	if len(qb.conditions) == 0 {
		return errors.New("delete without conditions is not allowed")
	}

	whereSQL := buildConditions(qb.conditions)
	query := fmt.Sprintf("DELETE FROM %s WHERE %s",
		qb.table,
		whereSQL)

	return qb.execExec(query)
}

// DeleteContext удаляет записи с контекстом
func (qb *QueryBuilder) DeleteContext(ctx context.Context) error {
	if len(qb.conditions) == 0 {
		return errors.New("delete without conditions is not allowed")
	}

	whereSQL := buildConditions(qb.conditions)
	query := fmt.Sprintf("DELETE FROM %s WHERE %s",
		qb.table,
		whereSQL)

	return qb.execExecContext(ctx, query)
}

// SubQuery создает подзапрос
func (qb *QueryBuilder) SubQuery(alias string) *QueryBuilder {
	sql, args := qb.buildQuery()
	return &QueryBuilder{
		columns: []string{fmt.Sprintf("(%s) AS %s", sql, alias)},
		db:      qb.db,
		conditions: []Condition{{
			args: args,
		}},
	}
}

// WhereSubQuery добавляет условие с подзапросом
func (qb *QueryBuilder) WhereSubQuery(column string, operator string, subQuery *QueryBuilder) *QueryBuilder {
	sql, args := subQuery.buildQuery()
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("%s %s (%s)", column, operator, sql),
		args:     args,
	})
	return qb
}

// Union объединяет запросы через UNION
func (qb *QueryBuilder) Union(other *QueryBuilder) *QueryBuilder {
	sql1, args1 := qb.buildQuery()
	sql2, args2 := other.buildQuery()

	return &QueryBuilder{
		db:      qb.db,
		columns: []string{fmt.Sprintf("(%s) UNION (%s)", sql1, sql2)},
		conditions: []Condition{{
			args: append(args1, args2...),
		}},
	}
}

// UnionAll объединяет запросы через UNION ALL
func (qb *QueryBuilder) UnionAll(other *QueryBuilder) *QueryBuilder {
	sql1, args1 := qb.buildQuery()
	sql2, args2 := other.buildQuery()

	return &QueryBuilder{
		db:      qb.db,
		columns: []string{fmt.Sprintf("(%s) UNION ALL (%s)", sql1, sql2)},
		conditions: []Condition{{
			args: append(args1, args2...),
		}},
	}
}

// WhereNull добавляет проверку на NULL
func (qb *QueryBuilder) WhereNull(column string) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("%s IS NULL", column),
	})
	return qb
}

// WhereNotNull добавляет проверку на NOT NULL
func (qb *QueryBuilder) WhereNotNull(column string) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("%s IS NOT NULL", column),
	})
	return qb
}

// WhereBetween добавляет условие BETWEEN
func (qb *QueryBuilder) WhereBetween(column string, start, end interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("%s BETWEEN ? AND ?", column),
		args:     []interface{}{start, end},
	})
	return qb
}

// WhereNotBetween добавляет условие NOT BETWEEN
func (qb *QueryBuilder) WhereNotBetween(column string, start, end interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("%s NOT BETWEEN ? AND ?", column),
		args:     []interface{}{start, end},
	})
	return qb
}

// HavingRaw добавляет сырое условие HAVING
func (qb *QueryBuilder) HavingRaw(sql string, args ...interface{}) *QueryBuilder {
	if qb.having != "" {
		qb.having += " AND "
	}
	qb.having += sql
	qb.conditions = append(qb.conditions, Condition{args: args})
	return qb
}

// WithTransaction выполняет запрос в существующей транзакции
func (qb *QueryBuilder) WithTransaction(tx *Transaction) *QueryBuilder {
	qb.db = tx.tx
	return qb
}

// LockForUpdate блокирует записи для обновления
func (qb *QueryBuilder) LockForUpdate() *QueryBuilder {
	return qb.Lock("FOR UPDATE")
}

// LockForShare блокирует записи для чтения
func (qb *QueryBuilder) LockForShare() *QueryBuilder {
	return qb.Lock("FOR SHARE")
}

// SkipLocked пропускает заблокированные записи
func (qb *QueryBuilder) SkipLocked() *QueryBuilder {
	return qb.Lock("SKIP LOCKED")
}

// NoWait не ждет разблокировки записей
func (qb *QueryBuilder) NoWait() *QueryBuilder {
	return qb.Lock("NOWAIT")
}

// Lock блокирует записи для обновления
func (qb *QueryBuilder) Lock(mode string) *QueryBuilder {
	qb.columns = append(qb.columns, mode)
	return qb
}

// Window добавляет оконную функцию
func (qb *QueryBuilder) Window(column string, partition string, orderBy string) *QueryBuilder {
	windowFunc := fmt.Sprintf("%s OVER (PARTITION BY %s ORDER BY %s)",
		column, partition, orderBy)
	qb.columns = append(qb.columns, windowFunc)
	return qb
}

// RowNumber добавляет ROW_NUMBER()
func (qb *QueryBuilder) RowNumber(partition string, orderBy string, alias string) *QueryBuilder {
	windowFunc := fmt.Sprintf("ROW_NUMBER() OVER (PARTITION BY %s ORDER BY %s) AS %s",
		partition, orderBy, alias)
	qb.columns = append(qb.columns, windowFunc)
	return qb
}

// Rank добавляет RANK()
func (qb *QueryBuilder) Rank(partition string, orderBy string, alias string) *QueryBuilder {
	windowFunc := fmt.Sprintf("RANK() OVER (PARTITION BY %s ORDER BY %s) AS %s",
		partition, orderBy, alias)
	qb.columns = append(qb.columns, windowFunc)
	return qb
}

// DenseRank добавляет DENSE_RANK()
func (qb *QueryBuilder) DenseRank(partition string, orderBy string, alias string) *QueryBuilder {
	windowFunc := fmt.Sprintf("DENSE_RANK() OVER (PARTITION BY %s ORDER BY %s) AS %s",
		partition, orderBy, alias)
	qb.columns = append(qb.columns, windowFunc)
	return qb
}

// WhereRaw добавляет сырое условие WHERE
func (qb *QueryBuilder) WhereRaw(sql string, args ...interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   sql,
		args:     args,
	})
	return qb
}

// OrWhereRaw добавляет сырое условие через OR
func (qb *QueryBuilder) OrWhereRaw(sql string, args ...interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "OR",
		clause:   sql,
		args:     args,
	})
	return qb
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

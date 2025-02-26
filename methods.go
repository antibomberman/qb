package qb

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

func (qb *Builder) Context(ctx context.Context) *Builder {
	qb.ctx = ctx
	return qb
}

// Find ищет запись по id
func (qb *Builder) Find(id any, dest any) (bool, error) {
	qb.Where("id = ?", id)
	return qb.First(dest)
}
func (qb *Builder) FindAsync(id any, dest any) (chan bool, chan error) {
	foundCh := make(chan bool, 1)
	errorCh := make(chan error, 1)
	go func() {
		found, err := qb.Find(id, dest)
		foundCh <- found
		errorCh <- err
	}()
	return foundCh, errorCh
}

// Get получает все записи
func (qb *Builder) Get(dest any) (bool, error) {
	query, args := qb.buildSelectQuery()
	return qb.execSelectContext(qb.ctx, dest, query, args...)
}
func (qb *Builder) GetAsync(dest any) (chan bool, chan error) {
	foundCh := make(chan bool, 1)
	errorCh := make(chan error, 1)
	go func() {
		found, err := qb.Get(dest)
		errorCh <- err
		foundCh <- found
	}()
	return foundCh, errorCh
}

// First получает первую запись
func (qb *Builder) First(dest any) (bool, error) {
	qb.Limit(1)
	query, args := qb.buildSelectQuery()
	return qb.execGetContext(qb.ctx, dest, query, args...)
}
func (qb *Builder) FirstAsync(dest any) (chan bool, chan error) {
	foundCh := make(chan bool, 1)
	errorCh := make(chan error, 1)
	go func() {
		found, err := qb.First(dest)
		foundCh <- found
		errorCh <- err
	}()
	return foundCh, errorCh
}

func (qb *Builder) Create(data any, fields ...string) (any, error) {
	go qb.Trigger(BeforeCreate, data)
	defer func() {
		go qb.Trigger(AfterCreate, data)
	}()
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
		insertFields, placeholders, _ = qb.getStructInfo(data)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		qb.tableName,
		strings.Join(insertFields, ", "),
		strings.Join(placeholders, ", "))

	if qb.getDriverName() == "postgres" {
		var id any
		query += " RETURNING id"
		err := qb.getExecutor().QueryRowx(query, data).Scan(&id)
		return id, err
	}

	result, err := qb.getExecutor().NamedExecContext(qb.ctx, query, data)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
func (qb *Builder) CreateAsync(data any, fields ...string) (chan any, chan error) {
	idCh := make(chan any, 1)
	errorCh := make(chan error, 1)
	go func() {
		id, err := qb.Create(data, fields...)
		idCh <- id
		errorCh <- err
	}()
	return idCh, errorCh
}

// CreateMap создает новую запись из map и возвращает её id
func (qb *Builder) CreateMap(data map[string]any) (any, error) {
	go qb.Trigger(BeforeCreate, data)
	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]any, 0, len(data))

	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		values = append(values, val)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		qb.tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	if qb.getDriverName() == "postgres" {
		var id any
		query = qb.rebindQuery(query + " RETURNING id")
		err := qb.getExecutor().(sqlx.QueryerContext).QueryRowxContext(qb.ctx, query, values...).Scan(&id)
		return id, err
	}

	result, err := qb.getExecutor().ExecContext(qb.ctx, qb.rebindQuery(query), values...)
	if err != nil {
		return 0, err
	}
	go qb.Trigger(AfterCreate, data)
	return result.LastInsertId()
}
func (qb *Builder) CreateMapAsync(data map[string]any) (chan any, chan error) {
	idCh := make(chan any, 1)
	errorCh := make(chan error, 1)
	go func() {
		id, err := qb.CreateMap(data)
		idCh <- id
		errorCh <- err
	}()
	return idCh, errorCh
}

// BatchInsert вставляет множество записей
func (qb *Builder) BatchInsert(records []map[string]any) error {
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
	var values []any
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
		qb.tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	return qb.execExecContext(qb.ctx, query, values...)
}
func (qb *Builder) BatchInsertAsync(records []map[string]any) chan error {
	ch := make(chan error, 1)
	go func() {
		err := qb.BatchInsert(records)
		ch <- err
	}()
	return ch

}

// BulkInsert выполняет массовую вставку записей с возвратом ID
func (qb *Builder) BulkInsert(records []map[string]any) error {
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
	var values []any
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
			qb.tableName,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "),
		)
		var ids []any
		err := qb.getExecutor().SelectContext(qb.ctx, &ids, qb.rebindQuery(query), values...)
		return err
	}

	query = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		qb.tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	result, err := qb.getExecutor().(sqlx.ExtContext).ExecContext(qb.ctx, qb.rebindQuery(query), values...)
	if err != nil {
		return err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	ids := make([]any, rowsAffected)
	for i := range ids {
		ids[i] = lastID
	}

	return nil
}
func (qb *Builder) BulkInsertAsync(records []map[string]any) chan error {
	errorCh := make(chan error, 1)
	go func() {
		err := qb.BulkInsert(records)

		errorCh <- err
	}()
	return errorCh
}

// Update обновляет записи используя структуру
func (qb *Builder) Update(data any, fields ...string) error {
	go qb.Trigger(BeforeUpdate, data)
	defer func() {
		go qb.Trigger(AfterUpdate, data)
	}()
	query, args := qb.buildUpdateQuery(data, fields)
	return qb.execExecContext(qb.ctx, query, args...)
}
func (qb *Builder) UpdateAsync(data any, fields ...string) chan error {
	ch := make(chan error, 1)
	go func() {
		err := qb.Update(data, fields...)
		ch <- err
	}()
	return ch
}

// UpdateMap обновляет записи используя map
func (qb *Builder) UpdateMap(data map[string]any) error {
	go qb.Trigger(BeforeUpdate, data)
	defer func() {
		go qb.Trigger(AfterUpdate, data)
	}()
	query, args := qb.buildUpdateMapQuery(data)
	fmt.Println(query)
	return qb.execExecContext(qb.ctx, query, args...)
}
func (qb *Builder) UpdateMapAsync(data map[string]any) chan error {
	ch := make(chan error, 1)
	go func() {
		err := qb.UpdateMap(data)
		ch <- err
	}()
	return ch
}

// BulkUpdate выполняет массовое обновление записей
func (qb *Builder) BulkUpdate(records []map[string]any, keyColumn string) error {
	if len(records) == 0 {
		return nil
	}

	// Получаем все колонки из первой записи
	columns := make([]string, 0)
	for column := range records[0] {
		if column != keyColumn {
			columns = append(columns, column)
		}
	}
	sort.Strings(columns)

	// Формируем CASE выражения для каждой колонки
	cases := make([]string, len(columns))
	keyValues := make([]any, 0, len(records))
	valueArgs := make([]any, 0, len(records)*len(columns))

	for i, column := range columns {
		whenClauses := make([]string, 0, len(records))
		for _, record := range records {
			if i == 0 {
				keyValues = append(keyValues, record[keyColumn])
			}
			whenClauses = append(whenClauses, "WHEN ? THEN ?")
			valueArgs = append(valueArgs, record[keyColumn], record[column])
		}
		cases[i] = fmt.Sprintf("%s = CASE %s %s END",
			column,
			keyColumn,
			strings.Join(whenClauses, " "),
		)
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s IN (%s)",
		qb.tableName,
		strings.Join(cases, ", "),
		keyColumn,
		strings.Repeat("?,", len(records)-1)+"?",
	)

	// Объединяем все аргументы
	args := make([]any, 0, len(valueArgs)+len(keyValues))
	args = append(args, valueArgs...)
	args = append(args, keyValues...)

	return qb.execExecContext(qb.ctx, query, args...)
}
func (qb *Builder) BulkUpdateAsync(records []map[string]any, keyColumn string) chan error {
	ch := make(chan error, 1)
	go func() {
		err := qb.BulkUpdate(records, keyColumn)
		ch <- err
	}()
	return ch
}

// BatchUpdate обновляет записи пакетами указанного размера
func (qb *Builder) BatchUpdate(records []map[string]any, keyColumn string, batchSize int) error {
	if len(records) == 0 {
		return nil
	}

	// Разбиваем записи на пакеты
	for i := 0; i < len(records); i += batchSize {
		// Проверяем контекст
		if err := qb.ctx.Err(); err != nil {
			return err
		}

		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}
		batch := records[i:end]

		// Обновляем текущий пакет
		if err := qb.BulkUpdate(batch, keyColumn); err != nil {
			return err
		}
	}

	return nil
}
func (qb *Builder) BatchUpdateAsync(records []map[string]any, keyColumn string, batchSize int) chan error {
	ch := make(chan error, 1)
	go func() {
		err := qb.BatchUpdate(records, keyColumn, batchSize)
		ch <- err
	}()
	return ch
}

func (qb *Builder) Delete() error {
	if len(qb.conditions) == 0 {
		return errors.New("delete without conditions is not allowed")
	}

	head := fmt.Sprintf("DELETE FROM %s", qb.tableName)
	body, args := qb.buildBodyQuery()

	return qb.execExecContext(qb.ctx, head+body, args...)
}
func (qb *Builder) DeleteAsync() chan error {
	ch := make(chan error, 1)
	go func() {
		err := qb.Delete()
		ch <- err
	}()
	return ch
}

// Select указывает колонки для выборки
func (qb *Builder) Select(columns ...string) *Builder {
	qb.columns = columns
	return qb
}

// Where добавляет условие AND
func (qb *Builder) Where(condition string, args ...any) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   condition,
		args:     args,
	})
	return qb
}

// WhereId добавляет условие WHERE id = ?
func (qb *Builder) WhereId(id any) *Builder {
	qb.Where("id = ?", id)
	return qb
}

// OrWhere добавляет условие OR
func (qb *Builder) OrWhere(condition string, args ...any) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "OR",
		clause:   condition,
		args:     args,
	})
	return qb
}

// WhereIn добавляет условие IN
func (qb *Builder) WhereIn(column string, values ...any) *Builder {
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
func (qb *Builder) WhereGroup(fn func(*Builder)) *Builder {
	group := &Builder{}
	fn(group)

	var args []any
	for _, cond := range group.conditions {
		args = append(args, cond.args...)
	}

	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		nested:   group.conditions,
		args:     args, // Добавляем собранные аргументы
	})
	return qb
}

// OrWhereGroup добавляет группу условий через OR
func (qb *Builder) OrWhereGroup(fn func(*Builder)) *Builder {
	group := &Builder{}
	fn(group)
	var args []any
	for _, cond := range group.conditions {
		args = append(args, cond.args...)
	}
	qb.conditions = append(qb.conditions, Condition{
		operator: "OR",
		nested:   group.conditions,
		args:     args,
	})
	return qb
}

// WhereExists добавляет условие EXISTS
func (qb *Builder) WhereExists(subQuery *Builder) *Builder {
	sql, args := subQuery.buildSelectQuery()
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("EXISTS (%s)", sql),
		args:     args,
	})
	return qb
}

// WhereNotExists добавляет условие NOT EXISTS
func (qb *Builder) WhereNotExists(subQuery *Builder) *Builder {
	sql, args := subQuery.buildSelectQuery()
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("NOT EXISTS (%s)", sql),
		args:     args,
	})
	return qb
}

// OrderBy добавляет сортировку
func (qb *Builder) OrderBy(column string, direction string) *Builder {
	qb.orderBy = append(qb.orderBy, fmt.Sprintf("%s %s", column, direction))
	return qb
}

// GroupBy добавляет группировку
func (qb *Builder) GroupBy(columns ...string) *Builder {
	qb.groupBy = columns
	return qb
}

// Having добавляет условие для группировки
func (qb *Builder) Having(condition string) *Builder {
	qb.having = condition
	return qb
}

// Limit устанавливает ограничение на количество записей
func (qb *Builder) Limit(limit int) *Builder {
	qb.limit = limit
	return qb
}

// Offset устанавливает смещение
func (qb *Builder) Offset(offset int) *Builder {
	qb.offset = offset
	return qb
}

// As устанавливает алиас для таблицы
func (qb *Builder) As(alias string) *Builder {
	qb.alias = alias
	return qb
}

// Increment увеличивает значение поля
func (qb *Builder) Increment(column string, value any) error {
	head := fmt.Sprintf("UPDATE %s SET %s = %s + ?", qb.tableName, column, column)

	body, args := qb.buildBodyQuery()

	args = append([]any{value}, args...)

	return qb.execExecContext(qb.ctx, head+body, args...)
}

// Decrement уменьшает значение поля
func (qb *Builder) Decrement(column string, value any) error {
	head := fmt.Sprintf("UPDATE %s SET %s = %s - ?", qb.tableName, column, column)

	body, args := qb.buildBodyQuery()
	args = append([]any{value}, args...)

	return qb.execExecContext(qb.ctx, head+body, args...)
}

// SubQuery создает подзапрос
func (qb *Builder) SubQuery(alias string) *Builder {
	sql, args := qb.buildSelectQuery()
	return &Builder{
		columns: []string{fmt.Sprintf("(%s) AS %s", sql, alias)},
		db:      qb.db,
		conditions: []Condition{{
			args: args,
		}},
	}
}

// WhereSubQuery добавляет условие подзапросом
func (qb *Builder) WhereSubQuery(column string, operator string, subQuery *Builder) *Builder {
	sql, args := subQuery.buildSelectQuery()
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("%s %s (%s)", column, operator, sql),
		args:     args,
	})
	return qb
}

// Union объединяет запросы через UNION
func (qb *Builder) Union(other *Builder) *Builder {
	sql1, args1 := qb.buildSelectQuery()
	sql2, args2 := other.buildSelectQuery()

	return &Builder{
		db:      qb.db,
		columns: []string{fmt.Sprintf("(%s) UNION (%s)", sql1, sql2)},
		conditions: []Condition{{
			args: append(args1, args2...),
		}},
	}
}

// UnionAll объединяет запросы через UNION ALL
func (qb *Builder) UnionAll(other *Builder) *Builder {
	sql1, args1 := qb.buildSelectQuery()
	sql2, args2 := other.buildSelectQuery()

	return &Builder{
		db:      qb.db,
		columns: []string{fmt.Sprintf("(%s) UNION ALL (%s)", sql1, sql2)},
		conditions: []Condition{{
			args: append(args1, args2...),
		}},
	}
}

// WhereNull добавляет проверку на NULL
func (qb *Builder) WhereNull(column string) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("%s IS NULL", column),
	})
	return qb
}

// WhereNotNull добавляет проверку на NOT NULL
func (qb *Builder) WhereNotNull(column string) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("%s IS NOT NULL", column),
	})
	return qb
}

// WhereBetween добавляет условие BETWEEN
func (qb *Builder) WhereBetween(column string, start, end any) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("%s BETWEEN ? AND ?", column),
		args:     []any{start, end},
	})
	return qb
}

// WhereNotBetween добавляет условие NOT BETWEEN
func (qb *Builder) WhereNotBetween(column string, start, end any) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("%s NOT BETWEEN ? AND ?", column),
		args:     []any{start, end},
	})
	return qb
}

// HavingRaw добавляет сырое условие HAVING
func (qb *Builder) HavingRaw(sql string, args ...any) *Builder {
	if qb.having != "" {
		qb.having += " AND "
	}
	qb.having += sql
	qb.conditions = append(qb.conditions, Condition{args: args})
	return qb
}

// WithTransaction выполняет запрос в существующей транзакции
func (qb *Builder) WithTransaction(tx *Transaction) *Builder {
	qb.db = tx.Tx
	return qb
}

// LockForUpdate блокирует записи для обновления
func (qb *Builder) LockForUpdate() *Builder {
	return qb.Lock("FOR UPDATE")
}

// LockForShare блокирует записи для чтения
func (qb *Builder) LockForShare() *Builder {
	return qb.Lock("FOR SHARE")
}

// SkipLocked пропускает заблокированные записи
func (qb *Builder) SkipLocked() *Builder {
	return qb.Lock("SKIP LOCKED")
}

// NoWait не ждет разблокировки записей
func (qb *Builder) NoWait() *Builder {
	return qb.Lock("NOWAIT")
}

// Lock блокирует записи для обновления
func (qb *Builder) Lock(mode string) *Builder {
	qb.columns = append(qb.columns, mode)
	return qb
}

// Window добавляет оконную функцию
func (qb *Builder) Window(column string, partition string, orderBy string) *Builder {
	windowFunc := fmt.Sprintf("%s OVER (PARTITION BY %s ORDER BY %s)",
		column, partition, orderBy)
	qb.columns = append(qb.columns, windowFunc)
	return qb
}

// RowNumber добавляет ROW_NUMBER()
func (qb *Builder) RowNumber(partition string, orderBy string, alias string) *Builder {
	windowFunc := fmt.Sprintf("ROW_NUMBER() OVER (PARTITION BY %s ORDER BY %s) AS %s",
		partition, orderBy, alias)
	qb.columns = append(qb.columns, windowFunc)
	return qb
}

// Rank добавляет RANK()
func (qb *Builder) Rank(partition string, orderBy string, alias string) *Builder {
	windowFunc := fmt.Sprintf("RANK() OVER (PARTITION BY %s ORDER BY %s) AS %s",
		partition, orderBy, alias)
	qb.columns = append(qb.columns, windowFunc)
	return qb
}

// DenseRank добавляет DENSE_RANK()
func (qb *Builder) DenseRank(partition string, orderBy string, alias string) *Builder {
	windowFunc := fmt.Sprintf("DENSE_RANK() OVER (PARTITION BY %s ORDER BY %s) AS %s",
		partition, orderBy, alias)
	qb.columns = append(qb.columns, windowFunc)
	return qb
}

// WhereRaw добавляет сырое условие WHERE
func (qb *Builder) WhereRaw(sql string, args ...any) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   sql,
		args:     args,
	})
	return qb
}

// OrWhereRaw добавляет сырое условие через OR
func (qb *Builder) OrWhereRaw(sql string, args ...any) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "OR",
		clause:   sql,
		args:     args,
	})
	return qb
}

// Pluck получает значения одной колонки
func (qb *Builder) Pluck(column string, dest any) error {
	head := fmt.Sprintf("SELECT %s FROM %s", column, qb.tableName)

	body, args := qb.buildBodyQuery()
	_, err := qb.execSelect(dest, head+body, args...)
	return err
}

// Chunk обрабатывает записи чанками
func (qb *Builder) Chunk(size int, fn func(items any) error) error {
	return qb.ChunkContext(context.Background(), size, func(_ context.Context, items any) error {
		return fn(items)
	})
}

// ChunkContext обрабатывает записи чанками с контекстом
func (qb *Builder) ChunkContext(ctx context.Context, size int, fn func(context.Context, any) error) error {
	offset := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		dest := make([]map[string]any, 0, size)

		query, args := qb.buildSelectQuery()
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
func (qb *Builder) WithinGroup(column string, window string) *Builder {
	qb.columns = append(qb.columns, fmt.Sprintf("%s WITHIN GROUP (%s)", column, window))
	return qb
}

// Distinct добавляет DISTINCT к запросу
func (qb *Builder) Distinct(columns ...string) *Builder {
	if len(columns) == 0 {
		qb.columns = append(qb.columns, "DISTINCT *")
	} else {
		qb.columns = append(qb.columns, "DISTINCT "+strings.Join(columns, ", "))
	}
	return qb
}

// Value получает значение одного поля
func (qb *Builder) Value(column string) (any, error) {
	var result any
	head := fmt.Sprintf("SELECT %s FROM %s", column, qb.tableName)
	qb.Limit(1)

	body, args := qb.buildBodyQuery()

	_, err := qb.execGet(&result, head+body, args...)
	return result, err
}

// Values получает значения одного поля для всех записей
func (qb *Builder) Values(column string) ([]any, error) {
	var result []any
	head := fmt.Sprintf("SELECT %s FROM %s", column, qb.tableName)

	body, args := qb.buildBodyQuery()

	_, err := qb.execSelect(&result, head+body, args...)
	return result, err
}

// SoftDelete добавляет поддержку мягкого удаления
type SoftDelete struct {
	DeletedAt *time.Time `db:"deleted_at"`
}

// WithTrashed включает удаленные записи в выборку
func (qb *Builder) WithTrashed() *Builder {
	return qb
}

// OnlyTrashed выбирает только удаленные записи
func (qb *Builder) OnlyTrashed() *Builder {
	return qb.WhereNotNull("deleted_at")
}

// SoftDelete помечает записи как удаленные
func (qb *Builder) SoftDelete() error {
	return qb.UpdateMap(map[string]any{
		"deleted_at": time.Now(),
	})
}

// Restore восстанавливает удаленные записи
func (qb *Builder) Restore() error {
	return qb.UpdateMap(map[string]any{
		"deleted_at": nil,
	})
}

type JoinType string

const (
	InnerJoin JoinType = "INNER JOIN"
	LeftJoin  JoinType = "LEFT JOIN"
	RightJoin JoinType = "RIGHT JOIN"
	CrossJoin JoinType = "CROSS JOIN"
)

type Join struct {
	Type      JoinType
	tableName string
	Condition string
}

// Join добавляет INNER JOIN
func (qb *Builder) Join(table string, condition string) *Builder {
	qb.joins = append(qb.joins, Join{
		Type:      InnerJoin,
		tableName: table,
		Condition: condition,
	})
	return qb
}

// LeftJoin добавляет LEFT JOIN
func (qb *Builder) LeftJoin(table string, condition string) *Builder {
	qb.joins = append(qb.joins, Join{
		Type:      LeftJoin,
		tableName: table,
		Condition: condition,
	})
	return qb
}

// RightJoin добавляет RIGHT JOIN
func (qb *Builder) RightJoin(table string, condition string) *Builder {
	qb.joins = append(qb.joins, Join{
		Type:      RightJoin,
		tableName: table,
		Condition: condition,
	})
	return qb
}

// CrossJoin добавляет CROSS JOIN
func (qb *Builder) CrossJoin(table string) *Builder {
	qb.joins = append(qb.joins, Join{
		Type:      CrossJoin,
		tableName: table,
	})
	return qb
}

// Point представляет географическую точку
type Point struct {
	Lat float64
	Lng float64
}

// GeoSearch добавляет геопространственные запросы
func (qb *Builder) GeoSearch(column string, point Point, radius float64) *Builder {
	if qb.getDriverName() == "postgres" {
		// Для PostgreSQL с PostGIS
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause: fmt.Sprintf(
				"ST_DWithin(ST_SetSRID(ST_MakePoint(%s), 4326), ST_SetSRID(ST_MakePoint(?, ?), 4326), ?)",
				column),
			args: []any{point.Lng, point.Lat, radius},
		})
	} else {
		// Для MySQL
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause: fmt.Sprintf(
				"ST_Distance_Sphere(%s, POINT(?, ?)) <= ?",
				column),
			args: []any{point.Lng, point.Lat, radius},
		})
	}
	return qb
}

// FullTextSearch добавляет поддержку полнотекстового поиска
type FullTextSearch struct {
	SearchRank float64 `db:"search_rank"`
}

// Search выполняет полнотекстовый поиск
func (qb *Builder) Search(columns []string, query string) *Builder {
	if qb.getDriverName() == "postgres" {
		// Для PostgreSQL используем ts_vector и ts_query
		vectorExpr := make([]string, len(columns))
		for i, col := range columns {
			vectorExpr[i] = fmt.Sprintf("to_tsvector(%s)", col)
		}

		qb.columns = append(qb.columns,
			fmt.Sprintf("ts_rank_cd(to_tsvector(concat_ws(' ', %s)), plainto_tsquery(?)) as search_rank",
				strings.Join(columns, ", ")))

		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause: fmt.Sprintf("to_tsvector(concat_ws(' ', %s)) @@ plainto_tsquery(?)",
				strings.Join(columns, ", ")),
			args: []any{query},
		})

		qb.OrderBy("search_rank", "DESC")
	} else {
		// Для MySQL используем MATCH AGAINST
		qb.columns = append(qb.columns,
			fmt.Sprintf("MATCH(%s) AGAINST(? IN BOOLEAN MODE) as search_rank",
				strings.Join(columns, ",")))

		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause: fmt.Sprintf("MATCH(%s) AGAINST(? IN BOOLEAN MODE)",
				strings.Join(columns, ",")),
			args: []any{query},
		})

		qb.OrderBy("search_rank", "DESC")
	}

	return qb
}

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

// getDateFunctions возвращает функции для текущей СУБД
func (qb *Builder) getDateFunctions() DateFunctions {
	if qb.getDriverName() == "postgres" {
		return DateFunctions{
			DateDiff:    "DATE_PART('day', %s::timestamp - %s::timestamp)",
			DateTrunc:   "DATE_TRUNC",
			DateFormat:  "TO_CHAR",
			TimeZone:    "AT TIME ZONE",
			Extract:     "EXTRACT",
			DateAdd:     "% + INTERVAL '% %'",
			CurrentDate: "CURRENT_DATE",
		}
	}
	return DateFunctions{
		DateDiff:    "DATEDIFF(%s, %s)",
		DateTrunc:   "DATE_FORMAT", // MySQL не имеет прямого аналога DATE_TRUNC
		DateFormat:  "DATE_FORMAT",
		TimeZone:    "CONVERT_TZ",
		Extract:     "EXTRACT",
		DateAdd:     "DATE_ADD(%, INTERVAL % %)",
		CurrentDate: "CURDATE()",
	}
}

// convertToPostgresFormat преобразует формат даты из MySQL в PostgreSQL
func convertToPostgresFormat(mysqlFormat string) string {
	replacer := strings.NewReplacer(
		"%Y", "YYYY",
		"%m", "MM",
		"%d", "DD",
		"%H", "HH24",
		"%i", "MI",
		"%s", "SS",
	)
	return replacer.Replace(mysqlFormat)
}

// getMySQLDateFormat преобразует части даты в формат MySQL
func getMySQLDateFormat(part string) string {
	switch strings.ToLower(part) {
	case "year":
		return "%Y"
	case "month":
		return "%Y-%m"
	case "day":
		return "%Y-%m-%d"
	case "hour":
		return "%Y-%m-%d %H"
	case "minute":
		return "%Y-%m-%d %H:%i"
	default:
		return "%Y-%m-%d %H:%i:%s"
	}
}

// WhereDate добавляет условие по дате
func (qb *Builder) WhereDate(column string, operator string, value time.Time) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("DATE(%s) %s ?", column, operator),
		args:     []any{value.Format("2006-01-02")},
	})
	return qb
}

// WhereBetweenDates добавляет условие между датами
func (qb *Builder) WhereBetweenDates(column string, start time.Time, end time.Time) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("DATE(%s) BETWEEN ? AND ?", column),
		args:     []any{start.Format("2006-01-02"), end.Format("2006-01-02")},
	})
	return qb
}

// WhereDateTime добавляет условие по дате и времени
func (qb *Builder) WhereDateTime(column string, operator string, value time.Time) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("%s %s ?", column, operator),
		args:     []any{value.Format("2006-01-02 15:04:05")},
	})
	return qb
}

// WhereBetweenDateTime добавляет условие между датами и временем
func (qb *Builder) WhereBetweenDateTime(column string, start time.Time, end time.Time) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("%s BETWEEN ? AND ?", column),
		args: []any{
			start.Format("2006-01-02 15:04:05"),
			end.Format("2006-01-02 15:04:05"),
		},
	})
	return qb
}

// WhereYear добавляет условие по году
func (qb *Builder) WhereYear(column string, operator string, year int) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("EXTRACT(YEAR FROM %s) %s ?", column, operator),
		args:     []any{year},
	})
	return qb
}

// WhereMonth добавляет условие по месяцу
func (qb *Builder) WhereMonth(column string, operator string, month int) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("EXTRACT(MONTH FROM %s) %s ?", column, operator),
		args:     []any{month},
	})
	return qb
}

// WhereDay добавляет условие по дню
func (qb *Builder) WhereDay(column string, operator string, day int) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("EXTRACT(DAY FROM %s) %s ?", column, operator),
		args:     []any{day},
	})
	return qb
}

// WhereTime добавляет условие по времени (без учета даты)
func (qb *Builder) WhereTime(column string, operator string, value time.Time) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("TIME(%s) %s ?", column, operator),
		args:     []any{value.Format("15:04:05")},
	})
	return qb
}

// WhereDateIsNull проверяет является ли дата NULL
func (qb *Builder) WhereDateIsNull(column string) *Builder {
	return qb.WhereNull(column)
}

// WhereDateIsNotNull проверяет является ли дата NOT NULL
func (qb *Builder) WhereDateIsNotNull(column string) *Builder {
	return qb.WhereNotNull(column)
}

// WhereCurrentDate добавляет условие на текущую дату
func (qb *Builder) WhereCurrentDate(column string, operator string) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("DATE(%s) %s CURRENT_DATE", column, operator),
	})
	return qb
}

// WhereLastDays добавляет условие за последние n дней
func (qb *Builder) WhereLastDays(column string, days int) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("DATE(%s) >= CURRENT_DATE - INTERVAL ? DAY", column),
		args:     []any{days},
	})
	return qb
}

// WhereWeekday добавляет условие по дню недели (1 = Понедельник, 7 = Воскресенье)
func (qb *Builder) WhereWeekday(column string, operator string, weekday int) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("EXTRACT(DOW FROM %s) %s ?", column, operator),
		args:     []any{weekday},
	})
	return qb
}

// WhereQuarter добавляет условие по кварталу (1-4)
func (qb *Builder) WhereQuarter(column string, operator string, quarter int) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("EXTRACT(QUARTER FROM %s) %s ?", column, operator),
		args:     []any{quarter},
	})
	return qb
}

// WhereWeek добавляет условие по номеру недели в году (1-53)
func (qb *Builder) WhereWeek(column string, operator string, week int) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("EXTRACT(WEEK FROM %s) %s ?", column, operator),
		args:     []any{week},
	})
	return qb
}

// WhereDateRange добавляет условие по диапазону дат с включением/исключением границ
func (qb *Builder) WhereDateRange(column string, start time.Time, end time.Time, inclusive bool) *Builder {
	if inclusive {
		return qb.WhereBetweenDates(column, start, end)
	}

	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("DATE(%s) > ? AND DATE(%s) < ?", column, column),
		args:     []any{start.Format("2006-01-02"), end.Format("2006-01-02")},
	})
	return qb
}

// WhereNextDays добавляет условие на следующие n дней
func (qb *Builder) WhereNextDays(column string, days int) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("DATE(%s) <= CURRENT_DATE + INTERVAL ? DAY AND DATE(%s) >= CURRENT_DATE", column, column),
		args:     []any{days},
	})
	return qb
}

// WhereDateBetweenColumns проверяет, что дата находится между значениями двух других колонок
func (qb *Builder) WhereDateBetweenColumns(dateColumn string, startColumn string, endColumn string) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("DATE(%s) BETWEEN DATE(%s) AND DATE(%s)", dateColumn, startColumn, endColumn),
	})
	return qb
}

// WhereAge добавляет условие по возрасту (для дат рождения)
func (qb *Builder) WhereAge(column string, operator string, age int) *Builder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("EXTRACT(YEAR FROM AGE(%s)) %s ?", column, operator),
		args:     []any{age},
	})
	return qb
}

// WhereDateDiff добавляет условие по разнице между датами
func (qb *Builder) WhereDateDiff(column1 string, column2 string, operator string, days int) *Builder {
	df := qb.getDateFunctions()
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf(df.DateDiff+" %s ?", column1, column2, operator),
		args:     []any{days},
	})
	return qb
}

// WhereDateTrunc добавляет условие с усечением даты
func (qb *Builder) WhereDateTrunc(part string, column string, operator string, value time.Time) *Builder {
	df := qb.getDateFunctions()
	var clause string
	var args []any

	if qb.getDriverName() == "postgres" {
		clause = fmt.Sprintf("%s(?, %s) %s ?", df.DateTrunc, column, operator)
		args = []any{part, value}
	} else {
		// Преобразуем part в формат MySQL
		format := getMySQLDateFormat(part)
		clause = fmt.Sprintf("%s(%s, ?) %s ?", df.DateTrunc, column, operator)
		args = []any{format, value.Format(format)}
	}

	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   clause,
		args:     args,
	})
	return qb
}

// WhereTimeWindow добавляет условие попадания времени в окно
func (qb *Builder) WhereTimeWindow(column string, startTime, endTime time.Time) *Builder {
	if qb.getDriverName() == "postgres" {
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause:   fmt.Sprintf("EXTRACT(HOUR FROM %s) * 60 + EXTRACT(MINUTE FROM %s) BETWEEN ? AND ?", column, column),
			args: []any{
				startTime.Hour()*60 + startTime.Minute(),
				endTime.Hour()*60 + endTime.Minute(),
			},
		})
	} else {
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause:   fmt.Sprintf("TIME(%s) BETWEEN ? AND ?", column),
			args: []any{
				startTime.Format("15:04:05"),
				endTime.Format("15:04:05"),
			},
		})
	}
	return qb
}

// WhereBusinessDays добавляет условие только по рабочим дням
func (qb *Builder) WhereBusinessDays(column string) *Builder {
	if qb.getDriverName() == "postgres" {
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause:   fmt.Sprintf("EXTRACT(DOW FROM %s) BETWEEN 1 AND 5", column),
		})
	} else {
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause:   fmt.Sprintf("WEEKDAY(%s) < 5", column),
		})
	}
	return qb
}

// WhereDateFormat добавляет условие по отформатированной дате
func (qb *Builder) WhereDateFormat(column string, format string, operator string, value string) *Builder {
	df := qb.getDateFunctions()

	if qb.getDriverName() == "postgres" {
		// Преобразуем формат из MySQL в PostgreSQL
		pgFormat := convertToPostgresFormat(format)
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause:   fmt.Sprintf("%s(%s, ?) %s ?", df.DateFormat, column, operator),
			args:     []any{pgFormat, value},
		})
	} else {
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause:   fmt.Sprintf("%s(%s, ?) %s ?", df.DateFormat, column, operator),
			args:     []any{format, value},
		})
	}
	return qb
}

// WhereTimeZone добавляет условие с учетом временной зоны
func (qb *Builder) WhereTimeZone(column string, operator string, value time.Time, timezone string) *Builder {
	df := qb.getDateFunctions()

	if qb.getDriverName() == "postgres" {
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause:   fmt.Sprintf("%s %s ? %s ?", column, df.TimeZone, operator),
			args:     []any{timezone, value},
		})
	} else {
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause:   fmt.Sprintf("%s(%s, 'UTC', ?) %s ?", df.TimeZone, column, operator),
			args:     []any{timezone, value},
		})
	}
	return qb
}

type PaginationResult struct {
	Total       int64 `json:"total"`
	PerPage     int   `json:"per_page"`
	CurrentPage int   `json:"current_page"`
	LastPage    int   `json:"last_page"`
}
type PaginationTokenResult struct {
	NextToken string `json:"next_token"`
	HasMore   bool   `json:"has_more"`
}

// CursorPagination результат курсор-пагинации
type CursorPagination struct {
	Data       any    `json:"data"`
	NextCursor string `json:"next_cursor"`
	HasMore    bool   `json:"has_more"`
}

// Paginate выполняет пагинацию результатов
func (qb *Builder) Paginate(page int, perPage int, dest any) (*PaginationResult, error) {
	total, err := qb.Count()
	if err != nil {
		return nil, err
	}

	lastPage := int(math.Ceil(float64(total) / float64(perPage)))

	qb.Limit(perPage)
	qb.Offset((page - 1) * perPage)

	_, err = qb.Get(dest)
	if err != nil {
		return nil, err
	}

	return &PaginationResult{
		Total:       total,
		PerPage:     perPage,
		CurrentPage: page,
		LastPage:    lastPage,
	}, nil
}

// PaginateWithToken выполняет пагинацию с токеном
func (qb *Builder) PaginateWithToken(token string, limit int, dest any) (*PaginationTokenResult, error) {
	if token != "" {
		// Декодируем токен
		tokenData, err := base64.URLEncoding.DecodeString(token)
		if err != nil {
			return nil, err
		}

		var lastID any
		if err := json.Unmarshal(tokenData, &lastID); err != nil {
			return nil, err
		}

		qb.Where("id > ?", lastID)
	}

	qb.Limit(limit + 1) // Берем на 1 больше для проверки наличия следующей страницы

	if _, err := qb.Get(dest); err != nil {
		return nil, err
	}

	// Проверяем есть ли следующая страница
	val := reflect.ValueOf(dest).Elem()
	hasMore := val.Len() > limit

	if hasMore {
		// Удаляем последний элемент
		val.Set(val.Slice(0, limit))

		// Создаем токен из последнего ID
		lastItem := val.Index(limit - 1)
		lastID := lastItem.FieldByName("ID").Int()

		tokenData, err := json.Marshal(lastID)
		if err != nil {
			return nil, err
		}

		nextToken := base64.URLEncoding.EncodeToString(tokenData)

		return &PaginationTokenResult{
			NextToken: nextToken,
			HasMore:   true,
		}, nil
	}

	return &PaginationTokenResult{
		HasMore: false,
	}, nil
}

// PaginateWithCursor выполняет пагинацию с курсором
func (qb *Builder) PaginateWithCursor(cursor string, limit int, dest any) (*CursorPagination, error) {
	if cursor != "" {
		qb.Where("id > ?", cursor)
	}

	qb.Limit(limit + 1)

	if _, err := qb.Get(dest); err != nil {
		return nil, err
	}

	val := reflect.ValueOf(dest).Elem()
	hasMore := val.Len() > limit

	if hasMore {
		val.Set(val.Slice(0, limit))
		lastItem := val.Index(limit - 1)
		nextCursor := fmt.Sprint(lastItem.FieldByName("ID").Interface())

		return &CursorPagination{
			Data:       dest,
			NextCursor: nextCursor,
			HasMore:    true,
		}, nil
	}

	return &CursorPagination{
		Data:    dest,
		HasMore: false,
	}, nil
}

// Avg вычисляет среднее значение колонки
func (qb *Builder) Avg(column string) (float64, error) {
	var result float64
	head := fmt.Sprintf("SELECT AVG(%s) FROM %s", column, qb.tableName)

	body, args := qb.buildBodyQuery()
	_, err := qb.execGetContext(qb.ctx, &result, head+body, args...)
	return result, err
}

// Sum вычисляет сумму значений колонки
func (qb *Builder) Sum(column string) (float64, error) {
	var result float64
	head := fmt.Sprintf("SELECT SUM(%s) FROM %s", column, qb.tableName)
	body, args := qb.buildBodyQuery()
	_, err := qb.execGetContext(qb.ctx, &result, head+body, args...)
	return result, err
}

// Min находит минимальное значение колонки
func (qb *Builder) Min(column string) (float64, error) {
	var result float64
	head := fmt.Sprintf("SELECT MIN(%s) FROM %s", column, qb.tableName)
	body, args := qb.buildBodyQuery()
	_, err := qb.execGetContext(qb.ctx, &result, head+body, args...)
	return result, err
}

// Max находит максимальное значение колонки
func (qb *Builder) Max(column string) (float64, error) {
	var result float64
	head := fmt.Sprintf("SELECT MAX(%s) FROM %s", column, qb.tableName)
	body, args := qb.buildBodyQuery()
	_, err := qb.execGetContext(qb.ctx, &result, head+body, args...)
	return result, err
}

// Count возвращает количество записей
func (qb *Builder) Count() (int64, error) {
	var count int64
	head := fmt.Sprintf("SELECT COUNT(*) FROM %s", qb.tableName)

	body, args := qb.buildBodyQuery()
	fmt.Println(head+body, args)
	_, err := qb.execGetContext(qb.ctx, &count, head+body, args...)
	return count, err
}

// Exists проверяет существование записей
func (qb *Builder) Exists() (bool, error) {
	count, err := qb.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// AuditLog представляет запись аудита
type AuditLog struct {
	ID        int64     `db:"id"`
	tableName string    `db:"table_name"`
	RecordID  any       `db:"record_id"`
	Action    string    `db:"action"`
	OldData   []byte    `db:"old_data"`
	NewData   []byte    `db:"new_data"`
	UserID    any       `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

// WithAudit включает аудит для запроса
func (qb *Builder) WithAudit(userID any) *Builder {
	qb.On(BeforeUpdate, func(data any) error {
		var oldData []byte
		var recordID any
		var err error

		for _, cond := range qb.conditions {
			if cond.clause == "id = ?" {
				recordID = cond.args[0]
				break
			}
		}

		switch v := data.(type) {
		case map[string]any:
			oldData, err = json.Marshal(v)
		default:
			oldData, err = json.Marshal(data)
		}

		if err != nil {
			return err
		}

		// Создаем запись в таблице audits
		_, err = qb.queryBuilder.Query("audits").Create(&AuditLog{
			tableName: qb.tableName,
			RecordID:  recordID,
			Action:    "update",
			OldData:   oldData,
			UserID:    userID,
			CreatedAt: time.Now(),
		})
		return err
	})

	qb.On(AfterUpdate, func(data any) error {
		var newData []byte
		var recordID any
		var err error

		// Получаем ID из условий WHERE
		for _, cond := range qb.conditions {
			if cond.clause == "id = ?" {
				recordID = any(cond.args[0].(int))
				break
			}
		}

		switch v := data.(type) {
		case map[string]any:
			newData, err = json.Marshal(v)
		default:
			newData, err = json.Marshal(data)
		}

		if err != nil {
			return err
		}

		return qb.queryBuilder.Query("audits").
			Where("table_name = ?", qb.tableName).
			Where("record_id = ?", recordID).
			OrderBy("id", "DESC").
			Limit(1).
			UpdateMap(map[string]any{
				"new_data": newData,
			})
	})

	qb.On(AfterCreate, func(data any) error {
		var newData []byte
		var recordID any
		var err error

		switch v := data.(type) {
		case map[string]any:
			newData, err = json.Marshal(v)
			if id, ok := v["id"]; ok {
				recordID = id
			}
		default:
			val := reflect.ValueOf(data)
			if val.Kind() == reflect.Ptr {
				val = val.Elem()
			}
			if val.Kind() == reflect.Struct {
				recordID = val.FieldByName("ID").Int()
			}
			newData, err = json.Marshal(data)
		}

		if err != nil {
			return err
		}

		// Создаем запись в таблице audits
		_, err = qb.queryBuilder.Query("audits").Create(&AuditLog{
			tableName: qb.tableName,
			RecordID:  recordID,
			Action:    "create",
			NewData:   newData,
			UserID:    userID,
			CreatedAt: time.Now(),
		})
		return err
	})

	return qb
}

// QueuedOperation представляет отложенную операцию
type QueuedOperation struct {
	ID        int64     `db:"id"`
	Operation string    `db:"operation"`
	Data      []byte    `db:"data"`
	Status    string    `db:"status"`
	RunAt     time.Time `db:"run_at"`
}

// Queue добавляет операцию в очередь
func (qb *Builder) Queue(operation string, data any, runAt time.Time) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = qb.Create(&QueuedOperation{
		Operation: operation,
		Data:      jsonData,
		Status:    "pending",
		RunAt:     runAt,
	})
	return err
}

// ProcessQueue обрабатывает очередь
func (qb *Builder) ProcessQueue(handler func(QueuedOperation) error) error {
	var operations []QueuedOperation

	_, err := qb.Where("status = ? AND run_at <= ?", "pending", time.Now()).
		Get(&operations)
	if err != nil {
		return err
	}

	for _, op := range operations {
		if err := handler(op); err != nil {
			return err
		}

		err = qb.Where("id = ?", op.ID).
			UpdateMap(map[string]any{
				"status": "completed",
			})
		if err != nil {
			return err
		}
	}

	return nil
}

// MetricsCollector собирает метрики выполнения запросов
type MetricsCollector struct {
	mu      sync.RWMutex
	metrics map[string]*QueryMetrics
}

type QueryMetrics struct {
	Count        int64
	TotalTime    time.Duration
	AverageTime  time.Duration
	MaxTime      time.Duration
	ErrorCount   int64
	LastExecuted time.Time
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*QueryMetrics),
	}
}

func (mc *MetricsCollector) Track(query string, duration time.Duration, err error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	m, exists := mc.metrics[query]
	if !exists {
		m = &QueryMetrics{}
		mc.metrics[query] = m
	}

	m.Count++
	m.TotalTime += duration
	m.AverageTime = m.TotalTime / time.Duration(m.Count)
	if duration > m.MaxTime {
		m.MaxTime = duration
	}
	if err != nil {
		m.ErrorCount++
	}
	m.LastExecuted = time.Now()
}

// WithMetrics добавляет сбор метрик
func (qb *Builder) WithMetrics(collector *MetricsCollector) *Builder {
	qb.On(BeforeCreate, func(data any) error {
		start := time.Now()
		collector.Track("CREATE", time.Since(start), nil)
		return nil
	})

	qb.On(BeforeUpdate, func(data any) error {
		start := time.Now()
		collector.Track("UPDATE", time.Since(start), nil)
		return nil
	})

	return qb
}

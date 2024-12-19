package dblayer

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

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

// Join добавляет INNER JOIN
func (qb *QueryBuilder) Join(table string, condition string) *QueryBuilder {
	qb.joins = append(qb.joins, Join{
		Type:      InnerJoin,
		Table:     table,
		Condition: condition,
	})
	return qb
}

// LeftJoin добавляет LEFT JOIN
func (qb *QueryBuilder) LeftJoin(table string, condition string) *QueryBuilder {
	qb.joins = append(qb.joins, Join{
		Type:      LeftJoin,
		Table:     table,
		Condition: condition,
	})
	return qb
}

// RightJoin добавляет RIGHT JOIN
func (qb *QueryBuilder) RightJoin(table string, condition string) *QueryBuilder {
	qb.joins = append(qb.joins, Join{
		Type:      RightJoin,
		Table:     table,
		Condition: condition,
	})
	return qb
}

// CrossJoin добавляет CROSS JOIN
func (qb *QueryBuilder) CrossJoin(table string) *QueryBuilder {
	qb.joins = append(qb.joins, Join{
		Type:  CrossJoin,
		Table: table,
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

type PaginationResult struct {
	Data        interface{} `json:"data"`
	Total       int64       `json:"total"`
	PerPage     int         `json:"per_page"`
	CurrentPage int         `json:"current_page"`
	LastPage    int         `json:"last_page"`
}

// Paginate выполняет пагинацию результатов
func (qb *QueryBuilder) Paginate(page int, perPage int, dest interface{}) (*PaginationResult, error) {
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
		Data:        dest,
		Total:       total,
		PerPage:     perPage,
		CurrentPage: page,
		LastPage:    lastPage,
	}, nil
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

// Create создает новую запись из структуры
func (qb *QueryBuilder) Create(data interface{}) error {
	fields, placeholders := qb.getStructInfo(data)
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		qb.table,
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "))

	_, err := qb.getExecutor().NamedExec(query, data)
	return err
}

// CreateContext создает новую запись из структуры
func (qb *QueryBuilder) CreateContext(ctx context.Context, data interface{}) error {
	fields, placeholders := qb.getStructInfo(data)
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		qb.table,
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "))

	_, err := qb.getExecutor().NamedExecContext(ctx, query, data)
	return err
}

// CreateMap создает новую запись из map
func (qb *QueryBuilder) CreateMap(data map[string]interface{}) error {
	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	i := 1
	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		values = append(values, val)
		i++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		qb.table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return qb.execExec(query, values...)
}

// CreateMapContext создает новую запись из map с контекстом
func (qb *QueryBuilder) CreateMapContext(ctx context.Context, data map[string]interface{}) error {
	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	i := 1
	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		values = append(values, val)
		i++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		qb.table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return qb.execExecContext(ctx, query, values...)
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

// Update обновляет записи используя структуру
func (qb *QueryBuilder) Update(data interface{}, fields ...string) error {
	var sets []string
	if len(fields) > 0 {
		// Обновляем только ��казанные поля
		for _, field := range fields {
			sets = append(sets, fmt.Sprintf("%s = :%s", field, field))
		}
	} else {
		// Обновляем все поля
		fields, _ := qb.getStructInfo(data)
		for _, field := range fields {
			sets = append(sets, fmt.Sprintf("%s = :%s", field, field))
		}
	}

	whereSQL := buildConditions(qb.conditions)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		qb.table,
		strings.Join(sets, ", "),
		whereSQL)

	_, err := qb.getExecutor().NamedExec(query, data)
	return err
}

// UpdateContext обновляет записи с контекстом
func (qb *QueryBuilder) UpdateContext(ctx context.Context, data interface{}, fields ...string) error {
	var sets []string
	if len(fields) > 0 {
		// Обновляем только указанные поля
		for _, field := range fields {
			sets = append(sets, fmt.Sprintf("%s = :%s", field, field))
		}
	} else {
		// Обновляем все поля
		fields, _ := qb.getStructInfo(data)
		for _, field := range fields {
			sets = append(sets, fmt.Sprintf("%s = :%s", field, field))
		}
	}

	whereSQL := buildConditions(qb.conditions)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		qb.table,
		strings.Join(sets, ", "),
		whereSQL)

	_, err := qb.getExecutor().NamedExecContext(ctx, query, data)
	return err
}

// UpdateMap обновляет записи используя map
func (qb *QueryBuilder) UpdateMap(data map[string]interface{}) error {
	sets := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for col, val := range data {
		sets = append(sets, col+" = ?")
		values = append(values, val)
	}

	whereSQL := buildConditions(qb.conditions)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		qb.table,
		strings.Join(sets, ", "),
		whereSQL)

	return qb.execExec(query, values...)
}

func (qb *QueryBuilder) UpdateMapContext(ctx context.Context, data map[string]interface{}) error {
	sets := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for col, val := range data {
		sets = append(sets, col+" = ?")
		values = append(values, val)
	}

	whereSQL := buildConditions(qb.conditions)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		qb.table,
		strings.Join(sets, ", "),
		whereSQL)

	return qb.execExecContext(ctx, query, values...)
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

package qb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

type Condition struct {
	operator string
	clause   string
	nested   []Condition
	args     []any
}

type Builder struct {
	db           Executor
	tableName    string
	queryBuilder *QueryBuilder
	ctx          context.Context

	conditions      []Condition
	columns         []string
	orderBy         []string
	groupBy         []string
	having          string
	havingArgs      []any
	limit           int
	offset          int
	joins           []Join
	alias           string
	cacheKey        string
	cacheDuration   time.Duration
	events          map[EventType][]EventHandler
	distinctColumns []string // Новое поле для DISTINCT
	lockClause      string   // Для хранения предложения блокировки
	rawQuery        string   // Для хранения уже сгенерированного SQL (для подзапросов/объединений)
	rawArgs         []any    // Для хранения аргументов rawQuery
	isDistinct      bool     // Флаг, указывающий, был ли вызван Distinct

}

// buildConditions собирает условия WHERE в строку

// execGet выполняет запрос и получает одну запись
func (qb *Builder) execGet(dest any, query string, args ...any) (bool, error) {
	return qb.executeQuery(nil, "execGet", dest, query, args, true)
}

// execSelect выполняет запрос и получает множество записей
func (qb *Builder) execSelect(dest any, query string, args ...any) (bool, error) {
	return qb.executeQuery(nil, "execSelect", dest, query, args, false)
}

// execExec выполняет запрос без возврата данных
func (qb *Builder) execExec(query string, args ...any) error {
	return qb.executeExec(nil, "execExec", query, args)
}

// execGetContext выполняет запрос с контекстом и получает одну запись
func (qb *Builder) execGetContext(ctx context.Context, dest any, query string, args ...any) (bool, error) {
	return qb.executeQuery(ctx, "execGetContext", dest, query, args, true)
}

// execSelectContext выполняет запрос с контекстом и получает множество записей
func (qb *Builder) execSelectContext(ctx context.Context, dest any, query string, args ...any) (bool, error) {
	return qb.executeQuery(ctx, "execSelectContext", dest, query, args, false)
}

// execExecContext выполняет запрос с контекстом
func (qb *Builder) execExecContext(ctx context.Context, query string, args ...any) error {
	return qb.executeExec(ctx, "execExecContext", query, args)
}

// executeQuery performs a Get or Select operation, handles logging, and checks for sql.ErrNoRows.
func (qb *Builder) executeQuery(
	ctx context.Context,
	opName string,
	dest any,
	query string,
	args []any,
	isGet bool, // true for Get, false for Select
) (bool, error) {
	start := time.Now()
	var err error

	executor := qb.getExecutor()
	if ctx != nil {
		if isGet {
			err = executor.GetContext(ctx, dest, query, args...)
		} else {
			err = executor.SelectContext(ctx, dest, query, args...)
		}
	} else {
		if isGet {
			err = executor.Get(dest, query, args...)
		} else {
			err = executor.Select(dest, query, args...)
		}
	}

	qb.queryBuilder.Debug(opName, start, query, args)
	if err != nil {
		qb.queryBuilder.Error(err.Error(), start, query, args)
	}

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// executeExec performs an Exec operation and handles logging.
func (qb *Builder) executeExec(
	ctx context.Context,
	opName string,
	query string,
	args []any,
) error {
	start := time.Now()
	var err error

	executor := qb.getExecutor()
	if ctx != nil {
		_, err = executor.ExecContext(ctx, query, args...)
	} else {
		_, err = executor.Exec(query, args...)
	}

	qb.queryBuilder.Debug(opName, start, query, args)
	if err != nil {
		qb.queryBuilder.Error(err.Error(), start, query, args)
	}
	return err
}

// On регистрирует обработчик события
func (qb *Builder) On(event EventType, handler EventHandler) {
	if qb.events == nil {
		qb.events = make(map[EventType][]EventHandler)
	}
	qb.events[event] = append(qb.events[event], handler)
}

func (qb *Builder) Trigger(event EventType, data any) {
	if handlers, ok := qb.events[event]; ok {
		for _, handler := range handlers {
			if err := handler(data); err != nil {
				// Log the error instead of fataling
				qb.queryBuilder.Error(err.Error(), time.Now(), fmt.Sprintf("Event Trigger: %s", event), data)
			}
		}
	}
}

// getExecutor возвращает исполнитель запросов
func (qb *Builder) getExecutor() Executor {
	return qb.db
}

// rebindQuery преобразует плейсхолдеры под нужный диалект SQL
func (qb *Builder) rebindQuery(query string) string {
	return qb.getExecutor().Rebind(query)
}

func (qb *Builder) getStructInfo(data any) (fields []string, placeholders []string, values map[string]any) {
	values = make(map[string]any)
	v := reflect.ValueOf(data)
	if v.IsZero() {
		return fields, placeholders, values
	}

	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fields, placeholders, values
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		if tag := t.Field(i).Tag.Get("db"); tag != "" && tag != "-" {
			fields = append(fields, tag)
			placeholders = append(placeholders, ":"+tag)
			values[tag] = v.Field(i).Interface()
		}
	}
	return fields, placeholders, values
}
func getFieldNameByDBTag(structType reflect.Type) map[string]string {
	tagToField := make(map[string]string)

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		tag := field.Tag.Get("db")
		if tag != "" && tag != "-" && tag != "id" {
			tagToField[tag] = field.Name
		}
	}

	return tagToField
}

// getDriverName возвращает имя драйвера базы данных
func (qb *Builder) getDriverName() string {
	return qb.queryBuilder.driverName
}

// quoteIdentifier оборачивает идентификатор в кавычки, специфичные для драйвера
func (qb *Builder) quoteIdentifier(name string) string {
	driver := qb.getDriverName()
	parts := strings.Split(name, ".")
	quotedParts := make([]string, len(parts))
	for i, part := range parts {
		switch driver {
		case "mysql":
			if !strings.HasPrefix(part, "`") && !strings.HasSuffix(part, "`") {
				quotedParts[i] = "`" + part + "`"
			} else {
				quotedParts[i] = part
			}
		case "postgres":
			if !strings.HasPrefix(part, "\"") && !strings.HasSuffix(part, "\"") {
				quotedParts[i] = `"` + part + `"`
			} else {
				quotedParts[i] = part
			}
		default:
			if !strings.HasPrefix(part, "`") && !strings.HasSuffix(part, "`") {
				quotedParts[i] = "`" + part + "`"
			} else {
				quotedParts[i] = part
			}
		}
	}
	return strings.Join(quotedParts, ".")
}

// buildSetClause формирует SET-часть SQL-запроса для UPDATE
func (qb *Builder) buildSetClause(data any, fields []string) (string, []any, error) {
	var sets []string
	var args []any

	if len(fields) > 0 { // Обновление полей структуры
		v := reflect.ValueOf(data)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			return "", nil, errors.New("data must be a struct or a pointer to a struct when fields are specified")
		}

		t := v.Type()
		tagToField := getFieldNameByDBTag(t)

		for _, dbTag := range fields {
			fieldName, ok := tagToField[dbTag]
			if !ok {
				// Log this error, but don't stop the process
				qb.queryBuilder.Error(fmt.Sprintf("tag %s not found in struct for update", dbTag), time.Now(), "buildSetClause")
				continue
			}
			sets = append(sets, fmt.Sprintf("%s = ?", qb.quoteIdentifier(dbTag)))
			fieldVal := v.FieldByName(fieldName)
			if fieldVal.IsValid() {
				args = append(args, fieldVal.Interface())
			} else {
				// This case should ideally not happen if tagToField is correct
				qb.queryBuilder.Error(fmt.Sprintf("field %s not found in struct for update", fieldName), time.Now(), "buildSetClause")
				args = append(args, nil)
			}
		}
	} else { // Обновление из map[string]any или всех полей структуры
		val := reflect.ValueOf(data)
		if val.Kind() == reflect.Map && val.Type().Key().Kind() == reflect.String { // map[string]any
			for _, key := range val.MapKeys() {
				col := key.String()
				sets = append(sets, fmt.Sprintf("%s = ?", qb.quoteIdentifier(col)))
				args = append(args, val.MapIndex(key).Interface())
			}
		} else { // Все поля структуры
			fields, _, values := qb.getStructInfo(data)
			for _, field := range fields {
				sets = append(sets, fmt.Sprintf("%s = ?", qb.quoteIdentifier(field)))
				args = append(args, values[field])
			}
		}
	}

	if len(sets) == 0 {
		return "", nil, errors.New("no fields to update")
	}

	return strings.Join(sets, ", "), args, nil
}

func (qb *Builder) buildBodyQuery() (string, []any) {
	var args []any
	var sql strings.Builder

	for _, join := range qb.joins {
		// Упрощенная обработка quotedTableName
		quotedTableName := qb.quoteIdentifier(join.tableName)
		if strings.Contains(join.tableName, " ") { // Проверяем, есть ли алиас
			parts := strings.Fields(join.tableName)
			if len(parts) == 2 { // table alias
				// Оборачиваем только имя таблицы, алиас оставляем без кавычек
				quotedTableName = fmt.Sprintf("%s %s", qb.quoteIdentifier(parts[0]), parts[1])
			} else if len(parts) == 3 && strings.ToLower(parts[1]) == "as" { // table AS alias
				// Оборачиваем только имя таблицы, алиас оставляем без кавычек
				quotedTableName = fmt.Sprintf("%s AS %s", qb.quoteIdentifier(parts[0]), parts[2])
			}
		}

		if join.Type == CrossJoin {
			sql.WriteString(fmt.Sprintf(" %s %s", join.Type, quotedTableName))
		} else {
			sql.WriteString(fmt.Sprintf(" %s %s ON %s", join.Type, quotedTableName, join.Condition))
		}
	}

	if len(qb.conditions) > 0 {
		whereSQL, whereArgs := qb.buildConditions(qb.conditions)
		sql.WriteString(" WHERE " + whereSQL)
		args = append(args, whereArgs...)
	}

	if len(qb.groupBy) > 0 {
		sql.WriteString(" GROUP BY " + strings.Join(qb.groupBy, ", "))
	}

	if qb.having != "" {
		sql.WriteString(" HAVING " + qb.having)
		args = append(args, qb.havingArgs...)
	}

	if len(qb.orderBy) > 0 {
		sql.WriteString(" ORDER BY " + strings.Join(qb.orderBy, ", "))
	}

	if qb.limit > 0 {
		sql.WriteString(fmt.Sprintf(" LIMIT %d", qb.limit))
	}

	if qb.offset > 0 {
		sql.WriteString(fmt.Sprintf(" OFFSET %d", qb.offset))
	}

	return sql.String(), args
}

// buildQuery собирает полный SQL запрос
func (qb *Builder) buildSelectQuery(dest any) (string, []any) {
	var selectClause string
	var selectColumns []string

	if qb.isDistinct {
		if len(qb.distinctColumns) == 0 {
			selectClause = "DISTINCT *"
		} else {
			for _, col := range qb.distinctColumns {
				selectColumns = append(selectColumns, qb.quoteIdentifier(col))
			}
			selectClause = "DISTINCT " + strings.Join(selectColumns, ", ")
		}
	} else if len(qb.columns) > 0 {
		for _, col := range qb.columns {
			parts := strings.Fields(col)
			if len(parts) >= 3 && strings.ToLower(parts[len(parts)-2]) == "as" {
				// Это случай "column AS alias"
				columnPart := strings.Join(parts[:len(parts)-2], " ") // Восстанавливаем часть с колонкой
				aliasPart := parts[len(parts)-1]                      // Алиас

				// Улучшенная эвристика для выражений
				isExpression := strings.ContainsAny(columnPart, "()") ||
					strings.ContainsAny(columnPart, "+*/") ||
					strings.Contains(strings.ToUpper(columnPart), "CASE") ||
					strings.Contains(strings.ToUpper(columnPart), "WHEN") ||
					strings.Contains(strings.ToUpper(columnPart), "THEN") ||
					strings.Contains(strings.ToUpper(columnPart), "END")

				if isExpression {
					selectColumns = append(selectColumns, fmt.Sprintf("%s AS %s", columnPart, aliasPart))
				} else {
					selectColumns = append(selectColumns, fmt.Sprintf("%s AS %s", qb.quoteIdentifier(columnPart), aliasPart))
				}
			} else {
				// Улучшенная эвристика для выражений
				isExpression := strings.ContainsAny(col, "()") ||
					strings.ContainsAny(col, "+*/") ||
					strings.Contains(strings.ToUpper(col), "CASE") ||
					strings.Contains(strings.ToUpper(col), "WHEN") ||
					strings.Contains(strings.ToUpper(col), "THEN") ||
					strings.Contains(strings.ToUpper(col), "END")

				if isExpression {
					selectColumns = append(selectColumns, col)
				} else {
					selectColumns = append(selectColumns, qb.quoteIdentifier(col))
				}
			}
		}
		selectClause = strings.Join(selectColumns, ", ")
	} else if dest != nil {
		fields, _, _ := qb.getStructInfo(dest)
		if len(fields) > 0 {
			for _, field := range fields {
				selectColumns = append(selectColumns, qb.quoteIdentifier(field))
			}
		}
		selectClause = strings.Join(selectColumns, ", ")
	} else {
		selectClause = "*"
	}

	tableName := qb.quoteIdentifier(qb.tableName)
	if qb.alias != "" {
		tableName = fmt.Sprintf("%s AS %s", tableName, qb.quoteIdentifier(qb.alias))
	}

	head := fmt.Sprintf("SELECT %s FROM %s", selectClause, tableName)
	body, args := qb.buildBodyQuery()
	finalQuery := head + body

	if qb.lockClause != "" {
		finalQuery = fmt.Sprintf("%s %s", finalQuery, qb.lockClause)
	}

	return finalQuery, args
}

// buildUpdateQuery собирает SQL запрос для UPDATE
func (qb *Builder) buildUpdateQuery(data any, fields []string) (string, []any) {
	setClause, setArgs, err := qb.buildSetClause(data, fields)
	if err != nil {
		// В случае ошибки, возвращаем пустую строку и nil,
		// или можно логировать ошибку и возвращать что-то, что позволит запросу не упасть,
		// но это зависит от желаемого поведения.
		// Для простоты, пока вернем пустую строку и nil.
		qb.queryBuilder.Error(err.Error(), time.Now(), "buildUpdateQuery")
		return "", nil
	}

	tableName := qb.quoteIdentifier(qb.tableName)
	if qb.alias != "" {
		tableName = fmt.Sprintf("%s AS %s", tableName, qb.quoteIdentifier(qb.alias))
	}

	head := fmt.Sprintf("UPDATE %s SET %s", tableName, setClause)

	body, bodyArgs := qb.buildBodyQuery()
	args := append(setArgs, bodyArgs...)
	return head + body, args
}

// buildInsertQuery собирает SQL запрос для INSERT
func (qb *Builder) buildUpdateMapQuery(data map[string]any) (string, []any, error) {
	setClause, setArgs, err := qb.buildSetClause(data, nil) // Pass nil for fields to indicate map processing
	if err != nil {
		return "", nil, err
	}

	tableName := qb.quoteIdentifier(qb.tableName)
	if qb.alias != "" {
		tableName = fmt.Sprintf("%s AS %s", tableName, qb.quoteIdentifier(qb.alias))
	}
	head := fmt.Sprintf("UPDATE %s SET %s", tableName, setClause)

	body, bodyArgs := qb.buildBodyQuery()
	args := append(setArgs, bodyArgs...)

	return head + body, args, nil
}

// ToSql возвращает сгенерированный SQL-запрос и его аргументы.
func (qb *Builder) ToSql() (string, []any) {
	if qb.rawQuery != "" { // Если это уже сгенерированный rawQuery
		return qb.rebindQuery(qb.rawQuery), qb.rawArgs
	}
	query, args := qb.buildSelectQuery(nil) // dest is nil as we don't execute
	return qb.rebindQuery(query), args
}

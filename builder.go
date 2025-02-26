package qb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
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

	conditions    []Condition
	columns       []string
	orderBy       []string
	groupBy       []string
	having        string
	limit         int
	offset        int
	joins         []Join
	alias         string
	cacheKey      string
	cacheDuration time.Duration
	events        map[EventType][]EventHandler
}

// Executor интерфейс для выполнения запросов
type Executor interface {
	sqlx.Ext
	Get(dest any, query string, args ...any) error
	Select(dest any, query string, args ...any) error
	NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	GetContext(ctx context.Context, dest any, query string, args ...any) error
}

// execGet выполняет запрос и получает одну запись
func (qb *Builder) execGet(dest any, query string, args ...any) (bool, error) {
	start := time.Now()

	query = qb.rebindQuery(query)
	err := qb.getExecutor().Get(dest, query, args...)
	qb.queryBuilder.logger.Debug(start, query, args...)
	if err != nil {
		qb.queryBuilder.logger.Error(start, query, args...)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// execSelect выполняет запрос и получает множество записей
func (qb *Builder) execSelect(dest any, query string, args ...any) (bool, error) {
	start := time.Now()
	query = qb.rebindQuery(query)
	err := qb.getExecutor().Select(dest, query, args...)
	qb.queryBuilder.logger.Debug(start, query, args...)
	if err != nil {
		qb.queryBuilder.logger.Error(start, query, args...)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// execExec выполняет запрос без возврата данных
func (qb *Builder) execExec(query string, args ...any) error {
	start := time.Now()
	query = qb.rebindQuery(query)
	_, err := qb.getExecutor().Exec(query, args...)
	qb.queryBuilder.logger.Debug(start, query, args...)
	if err != nil {
		qb.queryBuilder.logger.Error(start, query, args...)
	}
	return err
}

// execGetContext выполняет запрос с контекстом и получает одну запись
func (qb *Builder) execGetContext(ctx context.Context, dest any, query string, args ...any) (bool, error) {
	start := time.Now()
	query = qb.rebindQuery(query)
	err := qb.getExecutor().GetContext(ctx, dest, query, args...)
	qb.queryBuilder.logger.Debug(start, query, args...)
	if err != nil {
		qb.queryBuilder.logger.Error(start, query, args...)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// execSelectContext выполняет запрос с контекстом и получает множество записей
func (qb *Builder) execSelectContext(ctx context.Context, dest any, query string, args ...any) (bool, error) {
	start := time.Now()
	query = qb.rebindQuery(query)
	err := qb.getExecutor().SelectContext(ctx, dest, query, args...)
	qb.queryBuilder.logger.Debug(start, query, args...)
	if err != nil {
		qb.queryBuilder.logger.Error(start, query, args...)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// execExecContext выполняет запрос с контекстом
func (qb *Builder) execExecContext(ctx context.Context, query string, args ...any) error {
	start := time.Now()
	query = qb.rebindQuery(query)
	_, err := qb.getExecutor().ExecContext(ctx, query, args...)
	qb.queryBuilder.logger.Debug(start, query, args...)
	if err != nil {
		qb.queryBuilder.logger.Error(start, query, args...)
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
				log.Fatalln(err)
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

// buildConditions собирает условия WHERE в строку
func buildConditions(conditions []Condition) string {
	var parts []string

	for i, cond := range conditions {
		var part string

		if len(cond.nested) > 0 {
			nestedSQL := buildConditions(cond.nested)
			part = "(" + nestedSQL + ")"
		} else {
			part = cond.clause
		}

		if i == 0 {
			parts = append(parts, part)
		} else {
			parts = append(parts, cond.operator+" "+part)
		}
	}

	return strings.Join(parts, " ")
}

func (qb *Builder) getStructInfo(data any) (fields []string, placeholders []string, values map[string]any) {
	values = make(map[string]any)
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		if tag := t.Field(i).Tag.Get("db"); tag != "" && tag != "-" && tag != "id" {
			fields = append(fields, tag)
			placeholders = append(placeholders, ":"+tag)
			values[tag] = v.Field(i).Interface()
		}
	}
	return
}

// getDriverName возвращает имя драйвера базы данных
func (qb *Builder) getDriverName() string {
	return qb.queryBuilder.driverName
}

func (qb *Builder) buildBodyQuery() (string, []any) {
	var args []any
	var sql strings.Builder

	for _, join := range qb.joins {
		if join.Type == CrossJoin {
			sql.WriteString(fmt.Sprintf(" %s %s", join.Type, join.tableName))
		} else {
			sql.WriteString(fmt.Sprintf(" %s %s ON %s", join.Type, join.tableName, join.Condition))
		}
	}

	if len(qb.conditions) > 0 {
		whereSQL := buildConditions(qb.conditions)
		sql.WriteString(" WHERE " + whereSQL)

		for _, cond := range qb.conditions {
			args = append(args, cond.args...)
		}
	}

	if len(qb.groupBy) > 0 {
		sql.WriteString(" GROUP BY " + strings.Join(qb.groupBy, ", "))
	}

	if qb.having != "" {
		sql.WriteString(" HAVING " + qb.having)
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
func (qb *Builder) buildSelectQuery() (string, []any) {
	selectClause := "*"
	if len(qb.columns) > 0 {
		selectClause = strings.Join(qb.columns, ", ")
	}
	tableName := qb.tableName
	if qb.alias != "" {
		tableName = fmt.Sprintf("%s AS %s", tableName, qb.alias)
	}

	head := fmt.Sprintf("SELECT %s FROM %s", selectClause, tableName)
	body, args := qb.buildBodyQuery()
	return head + body, args
}

// buildUpdateQuery собирает SQL запрос для UPDATE
func (qb *Builder) buildUpdateQuery(data any, fields []string) (string, []any) {

	var sets []string
	var args []any

	if len(fields) > 0 {
		for _, field := range fields {
			sets = append(sets, fmt.Sprintf("%s = ?", field))
			val := reflect.ValueOf(data)
			if val.Kind() == reflect.Ptr {
				val = val.Elem()
			}
			field_val := val.FieldByName(field)
			if field_val.IsValid() {
				args = append(args, field_val.Interface())
			}
		}
	} else {
		fields, _, values := qb.getStructInfo(data)
		for _, field := range fields {
			sets = append(sets, fmt.Sprintf("%s = ?", field))
			args = append(args, values[field])
		}
	}
	tableName := qb.tableName
	if qb.alias != "" {
		tableName = fmt.Sprintf("%s AS %s", tableName, qb.alias)
	}

	head := fmt.Sprintf("UPDATE %s SET %s", qb.tableName, strings.Join(sets, ", "))

	body, bodyArgs := qb.buildBodyQuery()
	args = append(args, bodyArgs...)
	return head + body, args

}

// buildInsertQuery собирает SQL запрос для INSERT
func (qb *Builder) buildUpdateMapQuery(data map[string]any) (string, []any) {

	var sets []string
	var args []any

	for col, val := range data {
		sets = append(sets, col+" = ?")
		args = append(args, val)
	}

	tableName := qb.tableName
	if qb.alias != "" {
		tableName = fmt.Sprintf("%s AS %s", tableName, qb.alias)
	}
	head := fmt.Sprintf("UPDATE %s SET %s", qb.tableName, strings.Join(sets, ", "))

	body, bodyArgs := qb.buildBodyQuery()
	args = append(args, bodyArgs...)
	return head + body, args

}

package dblayer

import (
	"context"
	"database/sql"
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
	args     []interface{}
}

type QueryBuilder struct {
	db            interface{} // может быть *sqlx.DB или *sqlx.Tx
	dbl           *DBLayer
	table         string
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
	sqlx.ExtContext
	DriverName() string
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	NamedExec(query string, arg interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// On регистрирует обработчик события
func (qb *QueryBuilder) On(event EventType, handler EventHandler) {
	if qb.events == nil {
		qb.events = make(map[EventType][]EventHandler)
	}
	qb.events[event] = append(qb.events[event], handler)
}
func (qb *QueryBuilder) Trigger(event EventType, data interface{}) {
	if handlers, ok := qb.events[event]; ok {
		for _, handler := range handlers {
			if err := handler(data); err != nil {
				log.Fatalln(err)
			}
		}
	}
}

// getExecutor возвращает исполнитель запросов
func (qb *QueryBuilder) getExecutor() Executor {
	switch db := qb.db.(type) {
	case *sqlx.Tx:
		return db
	case *sqlx.DB:
		return db
	default:
		panic("invalid database executor")
	}
}

// rebindQuery преобразует плейсхолдеры под нужный диалект SQL
func (qb *QueryBuilder) rebindQuery(query string) string {
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

// getStructInfo получает информацию о полях структуры
//
//	func (qb *QueryBuilder) getStructInfo(data interface{}) (fields []string, placeholders []string) {
//		v := reflect.ValueOf(data)
//		if v.Kind() == reflect.Ptr {
//			v = v.Elem()
//		}
//		t := v.Type()
//
//		for i := 0; i < t.NumField(); i++ {
//			if tag := t.Field(i).Tag.Get("db"); tag != "" && tag != "-" && tag != "id" {
//				fields = append(fields, tag)
//				placeholders = append(placeholders, ":"+tag)
//			}
//		}
//		return
//	}
func (qb *QueryBuilder) getStructInfo(data interface{}) (fields []string, placeholders []string, values map[string]interface{}) {
	values = make(map[string]interface{})
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
func (qb *QueryBuilder) getDriverName() string {
	return qb.dbl.db.DriverName()
	// if db, ok := qb.db.(*sqlx.DB); ok {
	// 	return db.DriverName()
	// }
	// if tx, ok := qb.db.(*sqlx.Tx); ok {
	// 	return tx.DriverName()
	// }
	// return ""
}

// buildQuery собирает полный SQL запрос
func (qb *QueryBuilder) buildSelectQuery() (string, []interface{}) {
	var args []interface{}

	selectClause := "*"
	if len(qb.columns) > 0 {
		selectClause = strings.Join(qb.columns, ", ")
	}

	tableName := qb.table
	if qb.alias != "" {
		tableName = fmt.Sprintf("%s AS %s", tableName, qb.alias)
	}

	sql := fmt.Sprintf("SELECT %s FROM %s", selectClause, tableName)

	for _, join := range qb.joins {
		if join.Type == CrossJoin {
			sql += fmt.Sprintf(" %s %s", join.Type, join.Table)
		} else {
			sql += fmt.Sprintf(" %s %s ON %s", join.Type, join.Table, join.Condition)
		}
	}

	if len(qb.conditions) > 0 {
		whereSQL := buildConditions(qb.conditions)
		sql += " WHERE " + whereSQL

		// Собираем все аргументы из условий
		for _, cond := range qb.conditions {
			args = append(args, cond.args...)
		}
	}

	if len(qb.groupBy) > 0 {
		sql += " GROUP BY " + strings.Join(qb.groupBy, ", ")
	}

	if qb.having != "" {
		sql += " HAVING " + qb.having
	}

	if len(qb.orderBy) > 0 {
		sql += " ORDER BY " + strings.Join(qb.orderBy, ", ")
	}

	if qb.limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", qb.limit)
	}

	if qb.offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", qb.offset)
	}
	return sql, args
}

// buildUpdateQuery собирает SQL запрос для UPDATE
func (qb *QueryBuilder) buildUpdateQuery(data interface{}, fields []string) (string, []interface{}) {

	var sets []string
	var args []interface{}

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

	query := fmt.Sprintf("UPDATE %s SET %s", qb.table, strings.Join(sets, ", "))

	tableName := qb.table
	if qb.alias != "" {
		tableName = fmt.Sprintf("%s AS %s", tableName, qb.alias)
	}
	for _, join := range qb.joins {
		if join.Type == CrossJoin {
			query += fmt.Sprintf(" %s %s", join.Type, join.Table)
		} else {
			query += fmt.Sprintf(" %s %s ON %s", join.Type, join.Table, join.Condition)
		}
	}

	if len(qb.conditions) > 0 {
		whereSQL := buildConditions(qb.conditions)
		query += " WHERE " + whereSQL
		for _, cond := range qb.conditions {
			args = append(args, cond.args...)
		}
	}

	if len(qb.groupBy) > 0 {
		query += " GROUP BY " + strings.Join(qb.groupBy, ", ")
	}

	if qb.having != "" {
		query += " HAVING " + qb.having
	}

	if len(qb.orderBy) > 0 {
		query += " ORDER BY " + strings.Join(qb.orderBy, ", ")
	}

	if qb.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", qb.limit)
	}

	if qb.offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", qb.offset)
	}
	return query, args

}
func (qb *QueryBuilder) buildUpdateMapQuery(data map[string]interface{}) (string, []interface{}) {

	var sets []string
	var args []interface{}

	for col, val := range data {
		sets = append(sets, col+" = ?")
		args = append(args, val)
	}

	query := fmt.Sprintf("UPDATE %s SET %s", qb.table, strings.Join(sets, ", "))

	tableName := qb.table
	if qb.alias != "" {
		tableName = fmt.Sprintf("%s AS %s", tableName, qb.alias)
	}
	for _, join := range qb.joins {
		if join.Type == CrossJoin {
			query += fmt.Sprintf(" %s %s", join.Type, join.Table)
		} else {
			query += fmt.Sprintf(" %s %s ON %s", join.Type, join.Table, join.Condition)
		}
	}

	if len(qb.conditions) > 0 {
		whereSQL := buildConditions(qb.conditions)
		query += " WHERE " + whereSQL
		for _, cond := range qb.conditions {
			args = append(args, cond.args...)
		}
	}

	if len(qb.groupBy) > 0 {
		query += " GROUP BY " + strings.Join(qb.groupBy, ", ")
	}

	if qb.having != "" {
		query += " HAVING " + qb.having
	}

	if len(qb.orderBy) > 0 {
		query += " ORDER BY " + strings.Join(qb.orderBy, ", ")
	}

	if qb.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", qb.limit)
	}

	if qb.offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", qb.offset)
	}
	return query, args

}

package dblayer

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"reflect"
	"strings"
)

func (r *DBLayer) BulkUpdate(ctx context.Context, tableName string, updates []interface{}, idColumn string) error {
	if len(updates) == 0 {
		return nil
	}

	var cases []string
	var params []interface{}
	columns := make(map[string]bool)

	for _, update := range updates {
		val := reflect.ValueOf(update)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		if val.Kind() != reflect.Struct {
			return fmt.Errorf("each item in updates must be a struct or a pointer to a struct")
		}

		updateMap, err := structToMap(val.Interface())
		if err != nil {
			return fmt.Errorf("failed to convert struct to map: %w", err)
		}

		id, ok := updateMap[idColumn]
		if !ok {
			return fmt.Errorf("id column %s not found in operations data", idColumn)
		}

		for column, value := range updateMap {
			if column != idColumn {
				cases = append(cases, fmt.Sprintf("WHEN ? THEN ?"))
				params = append(params, id, value)
				columns[column] = true
			}
		}
	}

	if len(columns) == 0 {
		return fmt.Errorf("no columns to operations")
	}

	query := fmt.Sprintf("UPDATE %s SET ", tableName)
	for column := range columns {
		query += fmt.Sprintf("%s = CASE %s %s ELSE %s END, ", column, idColumn, strings.Join(cases, " "), column)
	}
	query = strings.TrimSuffix(query, ", ")
	query += fmt.Sprintf(" WHERE %s IN (?)", idColumn)

	ids := make([]interface{}, len(updates))
	for i, update := range updates {
		val := reflect.ValueOf(update)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		updateMap, _ := structToMap(val.Interface())
		ids[i] = updateMap[idColumn]
	}

	query, args, err := sqlx.In(query, append(params, ids)...)
	if err != nil {
		return fmt.Errorf("failed to expand IN clause: %w", err)
	}

	query = r.db.Rebind(query)
	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute bulk operations: %w", err)
	}

	return nil
}
func (r *DBLayer) BulkUpdateRecord(ctx context.Context, tableName string, updates []map[string]interface{}, idColumn string) error {
	if len(updates) == 0 {
		return nil
	}

	var cases []string
	var params []interface{}
	columns := make(map[string]bool)

	for _, update := range updates {
		id, ok := update[idColumn]
		if !ok {
			return fmt.Errorf("id column %s not found in operations data", idColumn)
		}

		for column, value := range update {
			if column != idColumn {
				cases = append(cases, fmt.Sprintf("WHEN ? THEN ?"))
				params = append(params, id, value)
				columns[column] = true
			}
		}
	}

	if len(columns) == 0 {
		return fmt.Errorf("no columns to operations")
	}

	query := fmt.Sprintf("UPDATE %s SET ", tableName)
	for column := range columns {
		query += fmt.Sprintf("%s = CASE %s %s ELSE %s END, ", column, idColumn, strings.Join(cases, " "), column)
	}
	query = strings.TrimSuffix(query, ", ")
	query += fmt.Sprintf(" WHERE %s IN (?)", idColumn)

	ids := make([]interface{}, len(updates))
	for i, update := range updates {
		ids[i] = update[idColumn]
	}

	query, args, err := sqlx.In(query, append(params, ids)...)
	if err != nil {
		return fmt.Errorf("failed to expand IN clause: %w", err)
	}

	query = r.db.Rebind(query)
	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute bulk operations: %w", err)
	}

	return nil
}

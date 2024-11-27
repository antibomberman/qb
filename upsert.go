package dblayer

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

func (r *DBLayer) Upsert(ctx context.Context, tableName string, record interface{}, uniqueColumns []string) error {
	val := reflect.ValueOf(record)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("record must be a struct or a pointer to a struct")
	}

	columns := make([]string, 0)
	values := make([]interface{}, 0)
	updateClauses := make([]string, 0)

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		value := val.Field(i).Interface()
		tag := field.Tag.Get("db")
		if tag != "" && tag != "-" {
			columns = append(columns, tag)
			values = append(values, value)
			if !contains(uniqueColumns, tag) {
				updateClauses = append(updateClauses, fmt.Sprintf("%s = VALUES(%s)", tag, tag))
			}
		}
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON DUPLICATE KEY UPDATE %s",
		tableName,
		strings.Join(columns, ", "),
		strings.Repeat("?, ", len(columns)-1)+"?",
		strings.Join(updateClauses, ", "))

	_, err := r.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to upsert record in %s: %w", tableName, err)
	}

	return nil
}

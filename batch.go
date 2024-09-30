package dblayer

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

func (r *DBLayer) BatchInsert(ctx context.Context, tableName string, records []interface{}) error {
	if len(records) == 0 {
		return nil
	}

	val := reflect.ValueOf(records[0])
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("records must be structs or pointers to structs")
	}

	columns := make([]string, 0)
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		tag := field.Tag.Get("db")
		if tag != "" && tag != "-" && field.Name != "ID" && field.Name != "Id" && field.Name != "id" {
			columns = append(columns, tag)
		}
	}

	placeholders := make([]string, len(records))
	for i := range records {
		placeholders[i] = "(" + strings.TrimRight(strings.Repeat("?,", len(columns)), ",") + ")"
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		tableName,
		strings.Join(columns, ","),
		strings.Join(placeholders, ","))
	values := make([]interface{}, 0, len(records)*len(columns))
	for _, record := range records {
		val := reflect.ValueOf(record)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)
			tag := field.Tag.Get("db")
			if tag != "" && tag != "-" && field.Name != "ID" && field.Name != "Id" && field.Name != "id" {
				values = append(values, val.Field(i).Interface())
			}
		}
	}
	fmt.Printf("Executing batch insert query: %s\n", query)
	fmt.Printf("Values: %v\n", values)
	_, err := r.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to batch insert into %s: %w", tableName, err)
	}

	return nil
}

func (r *DBLayer) BatchInsertRecords(ctx context.Context, tableName string, records []map[string]interface{}) error {
	if len(records) == 0 {
		return nil
	}
	var columns []string
	for column := range records[0] {
		columns = append(columns, column)
	}

	placeholders := make([]string, len(records))
	for i := range records {
		placeholders[i] = "(" + strings.Repeat("?,", len(columns)-1) + "?)"
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		tableName,
		strings.Join(columns, ","),
		strings.Join(placeholders, ","))

	var values []interface{}
	for _, record := range records {
		for _, column := range columns {
			values = append(values, record[column])
		}
	}

	_, err := r.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to batch insert into %s: %w", tableName, err)
	}

	return nil
}

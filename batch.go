package dblayer

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

// testing
func (r *DBLayer) batchTypeInsert(ctx context.Context, tableName string, records interface{}) error {
	val := reflect.ValueOf(records)
	if val.Kind() != reflect.Slice {
		return fmt.Errorf("records must be a slice")
	}

	if val.Len() == 0 {
		return nil
	}

	var columns []string
	var placeholders []string
	var values []interface{}

	// Определяем тип первого элемента
	firstElem := val.Index(0)
	if firstElem.Kind() == reflect.Interface {
		firstElem = firstElem.Elem()
	}

	switch firstElem.Kind() {
	case reflect.Map:
		// Обработка для map[string]interface{}
		if firstElem.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("map key must be string")
		}
		for _, key := range firstElem.MapKeys() {
			column := key.String()
			if column != "ID" {
				columns = append(columns, column)
			}
		}
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i)
			if elem.Kind() == reflect.Interface {
				elem = elem.Elem()
			}
			if elem.Kind() != reflect.Map {
				return fmt.Errorf("all elements must be map[string]interface{}")
			}
			placeholder := make([]string, 0, len(columns))
			for _, column := range columns {
				if value := elem.MapIndex(reflect.ValueOf(column)); value.IsValid() {
					placeholder = append(placeholder, "?")
					values = append(values, value.Interface())
				} else {
					placeholder = append(placeholder, "NULL")
				}
			}
			placeholders = append(placeholders, "("+strings.Join(placeholder, ",")+")")
		}
	case reflect.Struct:
		// Обработка для структур
		t := firstElem.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			tag := field.Tag.Get("db")
			if tag != "" && tag != "-" && field.Name != "ID" {
				columns = append(columns, tag)
			}
		}
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i)
			if elem.Kind() == reflect.Interface {
				elem = elem.Elem()
			}
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			if elem.Kind() != reflect.Struct {
				return fmt.Errorf("all elements must be structs or pointers to structs")
			}
			placeholder := make([]string, 0, len(columns))
			for j := 0; j < t.NumField(); j++ {
				field := t.Field(j)
				tag := field.Tag.Get("db")
				if tag != "" && tag != "-" && field.Name != "ID" {
					placeholder = append(placeholder, "?")
					values = append(values, elem.Field(j).Interface())
				}
			}
			placeholders = append(placeholders, "("+strings.Join(placeholder, ",")+")")
		}
	default:
		return fmt.Errorf("unsupported type: %v", firstElem.Kind())
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		tableName,
		strings.Join(columns, ","),
		strings.Join(placeholders, ","))

	fmt.Printf("Executing batch insert query: %s\n", query)
	fmt.Printf("Values: %v\n", values)
	_, err := r.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to batch insert into %s: %w", tableName, err)
	}

	return nil
}

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

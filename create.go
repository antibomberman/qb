package dblayer

import (
	"context"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"reflect"
	"strings"
)

func (r *DBLayer) Create(ctx context.Context, tableName string, record interface{}) (int64, error) {
	val := reflect.ValueOf(record)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return 0, fmt.Errorf("record must be a struct or a pointer to a struct")
	}

	columns := make([]string, 0)
	values := make([]string, 0)
	args := make(map[string]interface{})

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		value := val.Field(i).Interface()

		tag := field.Tag.Get("db")
		if tag != "" && tag != "-" && field.Name != "ID" {
			columns = append(columns, tag)
			values = append(values, ":"+tag)
			args[tag] = value
		}
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName, strings.Join(columns, ", "), strings.Join(values, ", "))

	driverName := r.db.DriverName()
	if driverName == "postgres" {
		query += " RETURNING id"
		var id int64
		err := r.db.QueryRowxContext(ctx, query, args).Scan(&id)
		if err != nil {
			if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
				return 0, UniqueViolationError{Column: pgErr.Constraint}
			}
			return 0, fmt.Errorf("failed to create record in %s: %w", tableName, err)
		}
		return id, nil
	}

	result, err := r.db.NamedExecContext(ctx, query, args)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, UniqueViolationError{Column: pgErr.Constraint}
		}
		return 0, fmt.Errorf("failed to create record in %s: %w", tableName, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

func (r *DBLayer) CreateRecord(ctx context.Context, tableName string, record map[string]interface{}) (int64, error) {
	columns := make([]string, 0)
	values := make([]string, 0)
	args := make(map[string]interface{})

	for key, value := range record {
		columns = append(columns, key)
		values = append(values, ":"+key)
		args[key] = value
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(columns, ", "), strings.Join(values, ", "))

	driverName := r.db.DriverName()
	if driverName == "postgres" {
		query += " RETURNING id"
		var id int64
		err := r.db.QueryRowxContext(ctx, query, args).Scan(&id)
		if err != nil {
			if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
				return 0, UniqueViolationError{Column: pgErr.Constraint}
			}
			return 0, fmt.Errorf("failed to create record in %s: %w", tableName, err)
		}
		return id, nil
	}

	result, err := r.db.NamedExecContext(ctx, query, args)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, UniqueViolationError{Column: pgErr.Constraint}
		}
		return 0, fmt.Errorf("failed to create record in %s: %w", tableName, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

func (r *DBLayer) GetOrCreate(ctx context.Context, tableName string, conditions []Condition, defaultValues map[string]interface{}, result interface{}) (bool, error) {
	exists, err := r.Get(ctx, tableName, conditions, result)
	if !exists {
		for _, cond := range conditions {
			defaultValues[cond.Column] = cond.Value
		}
		_, err = r.CreateRecord(ctx, tableName, defaultValues)
		if err != nil {
			return false, fmt.Errorf("failed to create record in %s: %w", tableName, err)
		}
		exists, err = r.Get(ctx, tableName, conditions, result)
	}
	return exists, err
}

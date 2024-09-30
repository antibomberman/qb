package dblayer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"reflect"
	"strings"
)

type DBLayer struct {
	db *sqlx.DB
}

func NewDBLayer(db *sqlx.DB) *DBLayer {
	return &DBLayer{db: db}
}

type Condition struct {
	Column   string
	Operator string
	Value    interface{}
}

type UniqueViolationError struct {
	Column string
}

func (e UniqueViolationError) Error() string {
	return fmt.Sprintf("unique constraint violation on column: %s", e.Column)
}

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

func (r *DBLayer) Update(ctx context.Context, tableName string, updates interface{}, conditions []Condition) (int64, error) {
	updatesMap, err := structToMap(updates)
	if err != nil {
		return 0, fmt.Errorf("failed to convert struct to map: %w", err)
	}

	query, args := r.buildUpdateQuery(tableName, updatesMap, conditions)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to update record in %s: %w", tableName, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return affected, nil
}

func (r *DBLayer) UpdateRecord(ctx context.Context, tableName string, updates map[string]interface{}, conditions []Condition) (int64, error) {
	query, args := r.buildUpdateQuery(tableName, updates, conditions)
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to update record in %s: %w", tableName, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return affected, nil
}

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
			return fmt.Errorf("id column %s not found in update data", idColumn)
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
		return fmt.Errorf("no columns to update")
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
		return fmt.Errorf("failed to execute bulk update: %w", err)
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
			return fmt.Errorf("id column %s not found in update data", idColumn)
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
		return fmt.Errorf("no columns to update")
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
		return fmt.Errorf("failed to execute bulk update: %w", err)
	}

	return nil
}

func (r *DBLayer) UpsertRecord(ctx context.Context, tableName string, record interface{}, uniqueColumns []string) error {
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

func (r *DBLayer) Delete(ctx context.Context, tableName string, conditions []Condition) (int64, error) {
	query, args := r.buildDeleteQuery(tableName, conditions)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to delete record from %s: %w", tableName, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return affected, nil
}

func (r *DBLayer) GetOrCreate(ctx context.Context, tableName string, conditions []Condition, defaultValues map[string]interface{}, result interface{}) (bool, error) {
	created := false
	err := r.Get(ctx, tableName, conditions, result)
	if errors.Is(err, sql.ErrNoRows) {
		for _, cond := range conditions {
			defaultValues[cond.Column] = cond.Value
		}
		_, err = r.CreateRecord(ctx, tableName, defaultValues)
		if err != nil {
			return false, fmt.Errorf("failed to create record in %s: %w", tableName, err)
		}
		created = true
		err = r.Get(ctx, tableName, conditions, result)
	}
	return created, err
}

func (r *DBLayer) GroupBy(ctx context.Context, tableName string, groupColumns []string, aggregations map[string]string, conditions []Condition) ([]map[string]interface{}, error) {
	query, args := r.buildGroupByQuery(tableName, groupColumns, aggregations, conditions)
	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute group by query on %s: %w", tableName, err)
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		err := rows.MapScan(row)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group by result: %w", err)
		}
		result = append(result, row)
	}

	return result, nil
}

func (r *DBLayer) InTransaction(ctx context.Context, fn func(context.Context, *sqlx.Tx) error) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	err = fn(ctx, tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("error rolling back transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *DBLayer) Lock(ctx context.Context, tableName string, conditions []Condition) error {
	query, args := r.buildSelectQuery(tableName, conditions)
	query += " FOR UPDATE"

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to lock records in %s: %w", tableName, err)
	}
	return nil
}

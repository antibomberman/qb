package dblayer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
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

func (r *DBLayer) Exists(ctx context.Context, tableName string, conditions []Condition) (bool, error) {
	query, args := r.buildExistsQuery(tableName, conditions)

	var exists bool
	err := r.db.QueryRowxContext(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check existence in %s: %w", tableName, err)
	}

	return exists, nil
}

func (r *DBLayer) CreateRecord(ctx context.Context, tableName string, record interface{}) (int64, error) {
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
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
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

func (r *DBLayer) GetRecord(ctx context.Context, tableName string, conditions []Condition, result interface{}) error {
	query, args := r.buildSelectQuery(tableName, conditions)

	err := r.db.GetContext(ctx, result, query, args...)
	if err != nil {
		return fmt.Errorf("failed to get record from %s: %w", tableName, err)
	}

	return nil
}

func (r *DBLayer) UpdateRecord(ctx context.Context, tableName string, updates map[string]interface{}, conditions []Condition) (int64, error) {
	if _, ok := updates["updated_at"]; !ok {
		updates["updated_at"] = time.Now()
	}

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

func (r *DBLayer) DeleteRecord(ctx context.Context, tableName string, conditions []Condition) (int64, error) {
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

func (r *DBLayer) ListRecords(ctx context.Context, tableName string, conditions []Condition, orderBy string, limit, offset int, result interface{}) error {
	query, args := r.buildSelectQuery(tableName, conditions)

	if orderBy != "" {
		query += " ORDER BY " + orderBy
	}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
		if offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", offset)
		}
	}

	err := r.db.SelectContext(ctx, result, query, args...)
	if err != nil {
		return fmt.Errorf("failed to list records from %s: %w", tableName, err)
	}

	return nil
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
		if tag != "" && tag != "-" {
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
			if tag != "" && tag != "-" {
				values = append(values, val.Field(i).Interface())
			}
		}
	}

	_, err := r.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to batch insert into %s: %w", tableName, err)
	}

	return nil
}

func (r *DBLayer) SearchLike(ctx context.Context, tableName string, searchColumn string, searchTerm string, additionalConditions []Condition, result interface{}) error {
	conditions := append([]Condition{{Column: searchColumn, Operator: "LIKE", Value: "%" + searchTerm + "%"}}, additionalConditions...)
	query, args := r.buildSelectQuery(tableName, conditions)

	err := r.db.SelectContext(ctx, result, query, args...)
	if err != nil {
		return fmt.Errorf("failed to search in %s: %w", tableName, err)
	}

	return nil
}

func (r *DBLayer) PaginateWithCursor(ctx context.Context, tableName string, cursorColumn string, cursorValue interface{}, pageSize int, conditions []Condition, result interface{}) error {
	cursorCondition := Condition{Column: cursorColumn, Operator: ">", Value: cursorValue}
	allConditions := append(conditions, cursorCondition)

	query, args := r.buildSelectQuery(tableName, allConditions)
	query += fmt.Sprintf(" ORDER BY %s LIMIT ?", cursorColumn)
	args = append(args, pageSize)

	err := r.db.SelectContext(ctx, result, query, args...)
	if err != nil {
		return fmt.Errorf("failed to paginate %s: %w", tableName, err)
	}

	return nil
}

func (r *DBLayer) ExecuteRawQuery(ctx context.Context, query string, args []interface{}, result interface{}) error {
	err := r.db.SelectContext(ctx, result, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute raw query: %w", err)
	}

	return nil
}

func (r *DBLayer) SelectFields(ctx context.Context, tableName string, fields []string, conditions []Condition, result interface{}) error {
	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(fields, ", "), tableName)
	where, args := r.buildWhereClause(conditions)
	if where != "" {
		query += " WHERE " + where
	}

	err := r.db.SelectContext(ctx, result, query, args...)
	if err != nil {
		return fmt.Errorf("failed to select fields from %s: %w", tableName, err)
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

func (r *DBLayer) ExecuteInBatches(ctx context.Context, items []interface{}, batchSize int, fn func(context.Context, []interface{}) error) error {
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[i:end]
		if err := fn(ctx, batch); err != nil {
			return err
		}
	}
	return nil
}

func (r *DBLayer) GetOrCreate(ctx context.Context, tableName string, conditions []Condition, defaultValues map[string]interface{}, result interface{}) (bool, error) {
	created := false
	err := r.GetRecord(ctx, tableName, conditions, result)
	if err == sql.ErrNoRows {
		// Запись не найдена, создаем новую
		for _, cond := range conditions {
			defaultValues[cond.Column] = cond.Value
		}
		_, err = r.CreateRecord(ctx, tableName, defaultValues)
		if err != nil {
			return false, fmt.Errorf("failed to create record in %s: %w", tableName, err)
		}
		created = true
		// Получаем только что созданную запись
		err = r.GetRecord(ctx, tableName, conditions, result)
	}
	return created, err
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

func (r *DBLayer) WithinRadius(ctx context.Context, tableName string, latColumn, lonColumn string, lat, lon float64, radiusKm float64, result interface{}) error {
	query := fmt.Sprintf(`
		SELECT *, 
		       (6371 * acos(cos(radians(?)) * cos(radians(%s)) * cos(radians(%s) - radians(?)) + sin(radians(?)) * sin(radians(%s)))) AS distance 
		FROM %s
		HAVING distance < ?
		ORDER BY distance
	`, latColumn, lonColumn, latColumn, tableName)

	err := r.db.SelectContext(ctx, result, query, lat, lon, lat, radiusKm)
	if err != nil {
		return fmt.Errorf("failed to find records within radius in %s: %w", tableName, err)
	}
	return nil
}

func (r *DBLayer) FullTextSearch(ctx context.Context, tableName string, searchColumns []string, searchTerm string, result interface{}) error {
	query := fmt.Sprintf(`
		SELECT * FROM %s
		WHERE MATCH(%s) AGAINST(? IN BOOLEAN MODE)
	`, tableName, strings.Join(searchColumns, ","))

	err := r.db.SelectContext(ctx, result, query, searchTerm)
	if err != nil {
		return fmt.Errorf("failed to perform full-text search in %s: %w", tableName, err)
	}
	return nil
}

func (r *DBLayer) ExecuteWithRetry(ctx context.Context, maxAttempts int, operation func(context.Context) error) error {
	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err = operation(ctx)
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt*100) * time.Millisecond):
		}
	}
	return fmt.Errorf("operation failed after %d attempts: %w", maxAttempts, err)
}

func (r *DBLayer) SaveJSON(ctx context.Context, tableName, idColumn string, id interface{}, jsonColumn string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	query := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?", tableName, jsonColumn, idColumn)
	_, err = r.db.ExecContext(ctx, query, jsonData, id)
	if err != nil {
		return fmt.Errorf("failed to save JSON data: %w", err)
	}
	return nil
}

func (r *DBLayer) LoadJSON(ctx context.Context, tableName, idColumn string, id interface{}, jsonColumn string, result interface{}) error {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", jsonColumn, tableName, idColumn)
	var jsonData []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(&jsonData)
	if err != nil {
		return fmt.Errorf("failed to load JSON data: %w", err)
	}

	err = json.Unmarshal(jsonData, result)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}

func (r *DBLayer) ExecuteWithTimeout(timeout time.Duration, operation func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- operation(ctx)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *DBLayer) BulkUpdate(ctx context.Context, tableName string, updates []map[string]interface{}, idColumn string) error {
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

	query, args, err := sqlx.In(query, params...)
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

func (r *DBLayer) GetHistoryWithTemporalValidity(ctx context.Context, tableName, idColumn string, id interface{}, validFromColumn, validToColumn string) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`
		SELECT * FROM %s
		WHERE %s = ?
		ORDER BY %s DESC
	`, tableName, idColumn, validFromColumn)

	rows, err := r.db.QueryxContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		entry := make(map[string]interface{})
		err := rows.MapScan(entry)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history entry: %w", err)
		}
		history = append(history, entry)
	}

	return history, nil
}

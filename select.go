package dblayer

import (
	"context"
	"fmt"
	"strings"
)

func (r *DBLayer) Exists(ctx context.Context, tableName string, conditions []Condition) (bool, error) {
	query, args := r.buildExistsQuery(tableName, conditions)

	var exists bool
	err := r.db.QueryRowxContext(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check existence in %s: %w", tableName, err)
	}

	return exists, nil
}

func (r *DBLayer) Get(ctx context.Context, tableName string, conditions []Condition, result interface{}) error {
	query, args := r.buildSelectQuery(tableName, conditions)

	err := r.db.GetContext(ctx, result, query, args...)
	if err != nil {
		return fmt.Errorf("failed to get record from %s: %w", tableName, err)
	}

	return nil
}

func (r *DBLayer) List(ctx context.Context, tableName string, conditions []Condition, orderBy string, limit, offset int, result interface{}) error {
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
func (r *DBLayer) SearchLike(ctx context.Context, tableName string, searchColumn string, searchTerm string, additionalConditions []Condition, result interface{}) error {
	conditions := append([]Condition{{Column: searchColumn, Operator: "LIKE", Value: "%" + searchTerm + "%"}}, additionalConditions...)
	query, args := r.buildSelectQuery(tableName, conditions)

	err := r.db.SelectContext(ctx, result, query, args...)
	if err != nil {
		return fmt.Errorf("failed to search in %s: %w", tableName, err)
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

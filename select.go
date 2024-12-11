package dblayer

import (
	"context"
	"database/sql"
	"errors"
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

func (r *DBLayer) Get(ctx context.Context, tableName string, conditions []Condition, result interface{}) (bool, error) {
	query, args := r.buildSelectQuery(tableName, conditions)

	err := r.db.GetContext(ctx, result, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *DBLayer) Last(ctx context.Context, tableName string, conditions []Condition, result interface{}) (bool, error) {
	query, args := r.buildSelectQuery(tableName, conditions)

	query += " ORDER BY id DESC LIMIT 1"

	err := r.db.GetContext(ctx, result, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *DBLayer) First(ctx context.Context, tableName string, conditions []Condition, result interface{}) (bool, error) {
	query, args := r.buildSelectQuery(tableName, conditions)

	query += " ORDER BY id ASC LIMIT 1"

	err := r.db.GetContext(ctx, result, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *DBLayer) All(ctx context.Context, tableName string, conditions []Condition, orderBy string, result interface{}) (bool, error) {
	query, args := r.buildSelectQuery(tableName, conditions)

	if orderBy != "" {
		query += " ORDER BY " + orderBy
	}

	err := r.db.SelectContext(ctx, result, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
func (r *DBLayer) List(ctx context.Context, tableName string, conditions []Condition, orderBy string, limit, offset int, result interface{}) (bool, error) {
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
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *DBLayer) PaginateWithCursor(ctx context.Context, tableName string, cursorColumn string, cursorValue interface{}, pageSize int, conditions []Condition, result interface{}) (bool, error) {
	cursorCondition := Condition{Column: cursorColumn, Operator: ">", Value: cursorValue}
	allConditions := append(conditions, cursorCondition)

	query, args := r.buildSelectQuery(tableName, allConditions)
	query += fmt.Sprintf(" ORDER BY %s LIMIT ?", cursorColumn)
	args = append(args, pageSize)

	err := r.db.SelectContext(ctx, result, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *DBLayer) SelectFields(ctx context.Context, tableName string, fields []string, conditions []Condition, result interface{}) (bool, error) {
	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(fields, ", "), tableName)
	where, args := r.buildWhereClause(conditions)
	if where != "" {
		query += " WHERE " + where
	}

	err := r.db.SelectContext(ctx, result, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *DBLayer) SearchLike(ctx context.Context, tableName string, searchColumn string, searchTerm string, additionalConditions []Condition, result interface{}) (bool, error) {
	conditions := append([]Condition{{Column: searchColumn, Operator: "LIKE", Value: "%" + searchTerm + "%"}}, additionalConditions...)
	query, args := r.buildSelectQuery(tableName, conditions)

	err := r.db.SelectContext(ctx, result, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *DBLayer) WithinRadius(ctx context.Context, tableName string, latColumn, lonColumn string, lat, lon float64, radiusKm float64, result interface{}) (bool, error) {
	query := fmt.Sprintf(`
		SELECT *, 
		       (6371 * acos(cos(radians(?)) * cos(radians(%s)) * cos(radians(%s) - radians(?)) + sin(radians(?)) * sin(radians(%s)))) AS distance 
		FROM %s
		HAVING distance < ?
		ORDER BY distance
	`, latColumn, lonColumn, latColumn, tableName)

	err := r.db.SelectContext(ctx, result, query, lat, lon, lat, radiusKm)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
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

func (r *DBLayer) SelectRaw(ctx context.Context, query string, args []interface{}, result interface{}) (bool, error) {
	err := r.db.SelectContext(ctx, result, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *DBLayer) GetRaw(ctx context.Context, query string, args []interface{}, result interface{}) (bool, error) {
	err := r.db.GetContext(ctx, result, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

package dblayer

import (
	"context"
	"fmt"
)

func (r *DBLayer) Update(ctx context.Context, tableName string, conditions []Condition, updates interface{}) (int64, error) {
	updatesMap, err := structToMap(updates)
	if err != nil {
		return 0, fmt.Errorf("failed to convert struct to map: %w", err)
	}

	query, args := r.buildUpdateQuery(tableName, updatesMap, conditions)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to operations record in %s: %w", tableName, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return affected, nil
}

func (r *DBLayer) UpdateRecord(ctx context.Context, tableName string, conditions []Condition, updates map[string]interface{}) (int64, error) {
	query, args := r.buildUpdateQuery(tableName, updates, conditions)
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to operations record in %s: %w", tableName, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return affected, nil
}

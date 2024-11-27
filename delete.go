package dblayer

import (
	"context"
	"fmt"
)

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

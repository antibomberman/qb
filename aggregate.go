package dblayer

import (
	"context"
	"fmt"
)

func (r *DBLayer) Count(ctx context.Context, tableName string, conditions []Condition) (int64, error) {
	query, args := r.buildAggregateQuery("COUNT(*)", tableName, conditions)
	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count records in %s: %w", tableName, err)
	}
	return count, nil
}

func (r *DBLayer) Avg(ctx context.Context, tableName, column string, conditions []Condition) (float64, error) {
	query, args := r.buildAggregateQuery(fmt.Sprintf("AVG(%s)", column), tableName, conditions)
	var avg float64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&avg)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate average of %s in %s: %w", column, tableName, err)
	}
	return avg, nil
}

func (r *DBLayer) Min(ctx context.Context, tableName, column string, conditions []Condition) (interface{}, error) {
	return r.aggregateValue(ctx, "MIN", tableName, column, conditions)
}

func (r *DBLayer) Max(ctx context.Context, tableName, column string, conditions []Condition) (interface{}, error) {
	return r.aggregateValue(ctx, "MAX", tableName, column, conditions)
}

func (r *DBLayer) Sum(ctx context.Context, tableName, column string, conditions []Condition) (float64, error) {
	query, args := r.buildAggregateQuery(fmt.Sprintf("SUM(%s)", column), tableName, conditions)
	var sum float64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate sum of %s in %s: %w", column, tableName, err)
	}
	return sum, nil
}

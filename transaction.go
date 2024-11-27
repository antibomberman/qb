package dblayer

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
)

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

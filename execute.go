package dblayer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (r *DBLayer) ExecuteRawQuery(ctx context.Context, query string, args []interface{}, result interface{}) (bool, error) {
	err := r.db.SelectContext(ctx, result, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
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

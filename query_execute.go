package dblayer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// execGet выполняет запрос и получает одну запись
func (qb *QueryBuilder) execGet(dest interface{}, query string, args ...interface{}) (bool, error) {
	query = qb.rebindQuery(query)
	err := qb.getExecutor().Get(dest, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// execSelect выполняет запрос и получает множество записей
func (qb *QueryBuilder) execSelect(dest interface{}, query string, args ...interface{}) (bool, error) {
	query = qb.rebindQuery(query)
	err := qb.getExecutor().Select(dest, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// execExec выполняет запрос без возврата данных
func (qb *QueryBuilder) execExec(query string, args ...interface{}) error {
	query = qb.rebindQuery(query)
	_, err := qb.getExecutor().Exec(query, args...)
	return err
}

// execGetContext выполняет запрос с контекстом и получает одну запись
func (qb *QueryBuilder) execGetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) (bool, error) {
	query = qb.rebindQuery(query)
	if ex, ok := qb.getExecutor().(interface {
		GetContext(context.Context, interface{}, string, ...interface{}) error
	}); ok {
		err := ex.GetContext(ctx, dest, query, args...)
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return err == nil, err
	}
	return false, fmt.Errorf("executor doesn't support context")
}

// execSelectContext выполняет запрос с контекстом и получает множество записей
func (qb *QueryBuilder) execSelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) (bool, error) {
	query = qb.rebindQuery(query)
	if ex, ok := qb.getExecutor().(interface {
		SelectContext(context.Context, interface{}, string, ...interface{}) error
	}); ok {
		err := ex.SelectContext(ctx, dest, query, args...)
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return err == nil, err
	}
	return false, fmt.Errorf("executor doesn't support context")
}

// execExecContext выполняет запрос с контекстом
func (qb *QueryBuilder) execExecContext(ctx context.Context, query string, args ...interface{}) error {
	query = qb.rebindQuery(query)
	if ex, ok := qb.getExecutor().(interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	}); ok {
		_, err := ex.ExecContext(ctx, query, args...)
		return err
	}
	return fmt.Errorf("executor doesn't support context")
}

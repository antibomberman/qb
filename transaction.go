package qb

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// Transaction представляет транзакцию
type Transaction struct {
	Tx           *sqlx.Tx
	QueryBuilder *QueryBuilder
	ctx          context.Context
}

// Begin начинает новую транзакцию
func (q *QueryBuilder) Begin() (*Transaction, error) {
	tx, err := q.db.Beginx()
	if err != nil {
		return nil, err
	}
	return &Transaction{Tx: tx, QueryBuilder: q}, nil
}

// BeginContext начинает новую транзакцию с контекстом
func (q *QueryBuilder) BeginContext(ctx context.Context) (*Transaction, error) {
	tx, err := q.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Transaction{Tx: tx, QueryBuilder: q, ctx: ctx}, nil
}

// Transaction выполняет функцию в транзакции
func (q *QueryBuilder) Transaction(fn func(*Transaction) error) error {
	return q.runInTransaction(nil, fn)
}

// TransactionContext выполняет функцию в транзакции с контекстом
func (q *QueryBuilder) TransactionContext(ctx context.Context, fn func(*Transaction) error) error {
	return q.runInTransaction(ctx, fn)
}

// runInTransaction - вспомогательная функция для выполнения операций в транзакции
func (q *QueryBuilder) runInTransaction(ctx context.Context, fn func(*Transaction) error) error {
	var tx *Transaction
	var err error

	if ctx != nil {
		tx, err = q.BeginContext(ctx)
	} else {
		tx, err = q.Begin()
	}

	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Commit фиксирует транзакцию
func (t *Transaction) Commit() error {
	return t.Tx.Commit()
}

// Rollback откатывает транзакцию
func (t *Transaction) Rollback() error {
	return t.Tx.Rollback()
}

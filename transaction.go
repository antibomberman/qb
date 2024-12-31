package DBL

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// Transaction представляет транзакцию
type Transaction struct {
	tx  *sqlx.Tx
	dbl *DBL
}

// Begin начинает новую транзакцию
func (d *DBL) Begin() (*Transaction, error) {
	tx, err := d.db.Beginx()
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: tx, dbl: d}, nil
}

// BeginContext начинает новую транзакцию с контекстом
func (d *DBL) BeginContext(ctx context.Context) (*Transaction, error) {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: tx, dbl: d}, nil
}

// Transaction выполняет функцию в транзакции
func (d *DBL) Transaction(fn func(*Transaction) error) error {
	tx, err := d.Begin()
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

// TransactionContext выполняет функцию в транзакции с контекстом
func (d *DBL) TransactionContext(ctx context.Context, fn func(*Transaction) error) error {
	tx, err := d.BeginContext(ctx)
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
	return t.tx.Commit()
}

// Rollback откатывает транзакцию
func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}

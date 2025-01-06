package dblayer

import (
	"context"
	q "github.com/antibomberman/dblayer/query"
	t "github.com/antibomberman/dblayer/table"
)

// Query Builder
func (d *DBLayer) Query(table string) *q.Builder {
	qb := q.NewQueryBuilder(d.DB, d.DriverName)
	return qb.Query(table)
}

func (d *DBLayer) Begin() (*q.Transaction, error) {
	qb := q.NewQueryBuilder(d.DB, d.DriverName)
	return qb.Begin()
}
func (d *DBLayer) BeginContext(ctx context.Context) (*q.Transaction, error) {
	qb := q.NewQueryBuilder(d.DB, d.DriverName)
	return qb.BeginContext(ctx)
}

func (d *DBLayer) Transaction(fn func(*q.Transaction) error) error {
	qb := q.NewQueryBuilder(d.DB, d.DriverName)
	return qb.Transaction(fn)
}
func (d *DBLayer) TransactionContext(ctx context.Context, fn func(*q.Transaction) error) error {
	qb := q.NewQueryBuilder(d.DB, d.DriverName)
	return qb.TransactionContext(ctx, fn)
}
func (d *DBLayer) Raw(query string, args ...interface{}) *q.RawQuery {
	qb := q.NewQueryBuilder(d.DB, d.DriverName)
	return qb.Raw(query, args...)
}

// Table Builder

func (d *DBLayer) Table(name string) *t.Table {
	qb := t.New(d.DB, d.DriverName)

	return qb.Table(name)

}

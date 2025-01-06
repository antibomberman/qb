package dblayer

import (
	"context"

	m "github.com/antibomberman/dblayer/migrate"
	q "github.com/antibomberman/dblayer/query"
	t "github.com/antibomberman/dblayer/table"
)

// Query Builder
func (d *DBLayer) Query(table string) *q.Builder {
	qb := q.New(d.DriverName, d.DB)
	return qb.Query(table)
}

func (d *DBLayer) Begin() (*q.Transaction, error) {
	qb := q.New(d.DriverName, d.DB)
	return qb.Begin()
}
func (d *DBLayer) BeginContext(ctx context.Context) (*q.Transaction, error) {
	qb := q.New(d.DriverName, d.DB)
	return qb.BeginContext(ctx)
}

func (d *DBLayer) Transaction(fn func(*q.Transaction) error) error {
	qb := q.New(d.DriverName, d.DB)
	return qb.Transaction(fn)
}
func (d *DBLayer) TransactionContext(ctx context.Context, fn func(*q.Transaction) error) error {
	qb := q.New(d.DriverName, d.DB)
	return qb.TransactionContext(ctx, fn)
}
func (d *DBLayer) Raw(query string, args ...interface{}) *q.RawQuery {
	qb := q.New(d.DriverName, d.DB)
	return qb.Raw(query, args...)
}

// Table Builder
func (d *DBLayer) Table(name string) *t.Table {
	qb := t.New(d.DriverName, d.DB)
	return qb.Table(name)
}
func (d *DBLayer) Migrate() *m.MigrateBuilder {
	qb := m.New(d.DriverName, d.DB)
	return qb
}

// GetQueryBuilder
func (d *DBLayer) GetQueryBuilder() *q.QueryBuilder {
	return q.New(d.DriverName, d.DB)
}

// GetTableBuilder
func (d *DBLayer) GetTableBuilder() *t.TableBuilder {
	return t.New(d.DriverName, d.DB)
}

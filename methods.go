package dblayer

import (
	"context"
	q "github.com/antibomberman/dblayer/query"
	s "github.com/antibomberman/dblayer/schema"
)

func (d *DBLayer) Table(name string) *q.Builder {
	qb := q.NewQueryBuilder(d.DB, d.DriverName)

	return qb.Table(name)
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

func (d *DBLayer) CacheRedisDriver(addr string, password string, db int) {
	qb := q.NewQueryBuilder(d.DB, d.DriverName)
	qb.CacheRedisDriver(addr, password, db)
}
func (d *DBLayer) CacheMemoryDriver() {
	qb := q.NewQueryBuilder(d.DB, d.DriverName)
	qb.CacheMemoryDriver()
}
func (d *DBLayer) Cache() q.CacheDriver {
	qb := q.NewQueryBuilder(d.DB, d.DriverName)
	return qb.Cache
}
func (d *DBLayer) Raw(query string, args ...interface{}) *q.RawQuery {
	qb := q.NewQueryBuilder(d.DB, d.DriverName)
	return qb.Raw(query, args...)
}

//schema

func (d *DBLayer) CreateTable(name string, fn func(builder *s.Builder)) error {
	qb := s.NewSchemaBuilder(d.DB, d.DriverName)

	return qb.CreateTable(name, fn)

}
func (d *DBLayer) CreateTableIfNotExists(name string, fn func(builder *s.Builder)) error {
	qb := s.NewSchemaBuilder(d.DB, d.DriverName)

	return qb.CreateTableIfNotExists(name, fn)

}
func (d *DBLayer) UpdateTable(name string, fn func(builder *s.Builder)) error {
	qb := s.NewSchemaBuilder(d.DB, d.DriverName)
	return qb.UpdateTable(name, fn)
}

func (d *DBLayer) DropTable(name string) error {
	qb := s.NewSchemaBuilder(d.DB, d.DriverName)
	return qb.DropTable(name)
}

func (d *DBLayer) TruncateTable(name string) error {
	qb := s.NewSchemaBuilder(d.DB, d.DriverName)
	return qb.TruncateTable(name)
}

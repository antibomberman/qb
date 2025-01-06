package query

import (
	"context"
	"github.com/jmoiron/sqlx"
)

type QueryBuilder struct {
	DB         *sqlx.DB
	DriverName string
	Cache      CacheDriver
}

func (q *QueryBuilder) CacheRedisDriver(addr string, password string, db int) {
	q.Cache = NewCacheRedis(addr, password, db)
}
func (q *QueryBuilder) CacheMemoryDriver() {
	q.Cache = NewCacheMemory()
}

func New(db *sqlx.DB, driverName string) *QueryBuilder {
	return &QueryBuilder{
		DB:         db,
		DriverName: driverName,
		Cache:      NewCacheMemory(),
	}
}

func (q *QueryBuilder) Query(table string) *Builder {
	return &Builder{
		TableName:    table,
		DB:           q.DB,
		QueryBuilder: q,
		Ctx:          context.TODO(),
	}
}

// Table начинает построение запроса в транзакции
func (t *Transaction) Query(table string) *Builder {
	return &Builder{
		TableName:    table,
		DB:           t.Tx,
		QueryBuilder: t.QueryBuilder,
		Ctx:          context.TODO(),
	}
}

//func (d *DBL) AuditTableCreate() error {
//	err := d.CreateTableIfNotExists("audits", func(schema *schema2.Schema) {
//		schema.ID()
//		schema.String("table_name", 20)
//		schema.BigInteger("record_id").Unsigned()
//		schema.String("action", 10)
//		schema.Json("old_data").Nullable()
//		schema.Json("new_data").Nullable()
//		schema.BigInteger("user_id").Unsigned()
//		schema.Timestamps()
//	})
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

package query

import (
	"context"
	"github.com/jmoiron/sqlx"
)

type QueryBuilder struct {
	db         *sqlx.DB
	driverName string
	cache      CacheDriver
}

func (q *QueryBuilder) CacheRedisDriver(addr string, password string, db int) {
	q.cache = NewCacheRedis(addr, password, db)
}
func (q *QueryBuilder) CacheMemoryDriver() {
	q.cache = NewCacheMemory()
}

func New(db *sqlx.DB, driverName string) *QueryBuilder {
	return &QueryBuilder{
		db:         db,
		driverName: driverName,
		cache:      NewCacheMemory(),
	}
}

func (q *QueryBuilder) Query(table string) *Builder {
	return &Builder{
		tableName:    table,
		db:           q.db,
		queryBuilder: q,
		ctx:          context.TODO(),
	}
}

func (t *Transaction) Query(table string) *Builder {
	return &Builder{
		tableName:    table,
		db:           t.Tx,
		queryBuilder: t.QueryBuilder,
		ctx:          context.TODO(),
	}
}

func (t *Transaction) Raw(query string, args ...interface{}) *RawQuery {
	return &RawQuery{
		query: query,
		args:  args,
		db:    t.Tx,
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

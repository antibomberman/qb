package migrate

import (
	q "github.com/antibomberman/dblayer/query"
	t "github.com/antibomberman/dblayer/table"
	"github.com/jmoiron/sqlx"
)

type MigrateBuilder struct {
	queryBuilder *q.QueryBuilder
	tableBuilder *t.TableBuilder
}

func New(db *sqlx.DB, driverName string) *MigrateBuilder {
	qb := q.New(db, driverName)
	tb := t.New(db, driverName)
	return &MigrateBuilder{
		queryBuilder: qb,
		tableBuilder: tb,
	}

}

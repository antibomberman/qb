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

func New(driverName string, db *sqlx.DB) *MigrateBuilder {
	qb := q.New(driverName, db)
	tb := t.New(driverName, db)
	return &MigrateBuilder{
		queryBuilder: qb,
		tableBuilder: tb,
	}

}

package dblayer

import (
	"github.com/jmoiron/sqlx"
)

type DBLayer struct {
	DB         *sqlx.DB
	DriverName string
}

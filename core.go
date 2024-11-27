package dblayer

import "github.com/jmoiron/sqlx"

type DBLayer struct {
	db *sqlx.DB
}

func NewDBLayer(db *sqlx.DB) *DBLayer {
	return &DBLayer{db: db}
}

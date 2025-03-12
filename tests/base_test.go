package tests

import (
	"database/sql"
	"log"

	"github.com/antibomberman/qb"
)

func Connect() qb.QueryBuilderInterface {
	db, err := sql.Open("mysql", "root:test_password@tcp(localhost:3316)/test_db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	return qb.New("mysql", db)
}

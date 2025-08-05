package tests

import (
	"database/sql"
	"log"

	"github.com/antibomberman/qb"
	_ "github.com/go-sql-driver/mysql"
)

func Connect() qb.QueryBuilderInterface {
	db, err := sql.Open("mysql", "root:rootpassword@tcp(localhost:3316)/test_db?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}

	return qb.New("mysql", db)
}

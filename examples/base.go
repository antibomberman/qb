package main

import (
	"database/sql"
	"log"

	"github.com/antibomberman/qb"
	_ "github.com/go-sql-driver/mysql"
)

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func main() {

	db, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/test")
	if err != nil {
		panic(err)
	}
	queryBuilder := qb.New("mysql", db)

	products := Product{}
	found, err := queryBuilder.From("products").Where("active = ?", 1).Get(&products)
	if err != nil {
		log.Fatal(err)
	}
	if !found {
		log.Println("No products found")
		return
	}
	log.Printf("Found product: %+v", products)
}

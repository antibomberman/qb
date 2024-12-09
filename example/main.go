package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/antibomberman/dblayer"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/jmoiron/sqlx"
	"log"
	_ "modernc.org/sqlite"
	"strconv"
	"time"
)

type Product struct {
	ID          int            `db:"id"`
	Name        string         `db:"name"`
	Article     string         `db:"article"`
	Description sql.NullString `db:"description"`
	Price       int            `db:"price"`
}

var schema = `
	DROP TABLE IF EXISTS products;
    CREATE TABLE products (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        article TEXT NOT NULL,
        description TEXT,
        price INTEGER NOT NULL
    )
`

func main() {
	db, err := sqlx.Open("sqlite", "./example/example.db")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()
	_, err = db.Exec(schema)
	if err != nil {
		log.Fatalln(err)
	}

	dblayer := dblayer.NewDBLayer(db)
	ctx := context.Background()

	id, err := dblayer.Create(ctx, "products", Product{
		Name:        "Product 1",
		Article:     "P12345",
		Description: sql.NullString{String: "This is product 1", Valid: true},
		Price:       100,
	})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Inserted product with ID: %d\n", id)
	products := faker()
	err = dblayer.BatchInsert(ctx, "products", products)
	if err != nil {
		log.Fatalf("Error: %v", err)
		return
	}
	err = dblayer.BatchInsertRecords(ctx, "products", []map[string]interface{}{{
		"name":        gofakeit.Company(),
		"article":     1233,
		"description": 123,
		"price":       gofakeit.Number(100, 1000),
	}})
	if err != nil {
		log.Fatalf("Error: %v", err)
		return
	}

	count, err := dblayer.Count(ctx, "products", nil)
	if err != nil {
		log.Fatalf("Error: %v", err)
		return
	}
	fmt.Printf("Total products: %d\n", count)

	minPrice, err := dblayer.Min(ctx, "products", "price", nil)

	fmt.Printf("Minimum price: %v\n", minPrice)

	maxPrice, err := dblayer.Max(ctx, "products", "price", nil)

	fmt.Printf("Maximum price: %v\n", maxPrice)

	avgPrice, err := dblayer.Avg(ctx, "products", "price", nil)

	fmt.Printf("Average price: %v\n", avgPrice)

	totalRevenue, err := dblayer.Sum(ctx, "products", "price", nil)

	fmt.Printf("Total revenue: %v\n", totalRevenue)
	err = dblayer.InTransaction(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
		product := Product{
			Name:        "Smartphone",
			Article:     "SMART123",
			Description: sql.NullString{String: "A high-end smartphone", Valid: true},
			Price:       999,
		}
		id, err := dblayer.Create(ctx, "products", product)
		if err != nil {
			return fmt.Errorf("failed to create product: %w", err)
		}
		fmt.Printf("Created product with ID: %d\n", id)

		return nil
	})
}

func faker() []interface{} {
	products := make([]interface{}, 0)
	gofakeit.Seed(time.Now().UnixNano())
	randomCount := gofakeit.Number(1, 2)

	for i := 0; i < randomCount; i++ {
		article := gofakeit.Number(100000, 999999)
		description := gofakeit.LoremIpsumSentence(gofakeit.Number(3, 30))
		product := Product{
			Name:        gofakeit.ProductName(),
			Article:     strconv.Itoa(article),
			Description: sql.NullString{String: description, Valid: true},
			Price:       gofakeit.Number(1, 1000),
		}

		products = append(products, product)

	}

	return products
}

func convertToInterfaceSlice(products []Product) []interface{} {
	result := make([]interface{}, len(products))
	for i, v := range products {
		result[i] = v
	}
	return result
}

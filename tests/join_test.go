package tests

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestProduct defines the structure for the products table in tests.
type TestProduct struct {
	ID         int    `db:"id"`
	Name       string `db:"name"`
	CategoryID int    `db:"category_id"`
}

// TestCategory defines the structure for the categories table in tests.
type TestCategory struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

func TestJoinQuery(t *testing.T) {
	qb := Connect()

	// 1. Create tables
	err := qb.Raw(`
		CREATE TABLE IF NOT EXISTS test_categories (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL
		);
	`).Exec()
	if err != nil {
		log.Fatalf("Failed to create test_categories table: %v", err)
	}

	err = qb.Raw(`
		CREATE TABLE IF NOT EXISTS test_products (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			category_id INT,
			FOREIGN KEY (category_id) REFERENCES test_categories(id)
		);
	`).Exec()
	if err != nil {
		log.Fatalf("Failed to create test_products table: %v", err)
	}

	// 2. Truncate tables for a clean state
	err = qb.Raw("SET FOREIGN_KEY_CHECKS = 0;").Exec()
	assert.NoError(t, err)
	err = qb.Raw("TRUNCATE TABLE test_products").Exec()
	assert.NoError(t, err)
	err = qb.Raw("TRUNCATE TABLE test_categories").Exec()
	assert.NoError(t, err)
	err = qb.Raw("SET FOREIGN_KEY_CHECKS = 1;").Exec()
	assert.NoError(t, err)

	// 3. Insert test data
	categoryID, err := qb.From("test_categories").CreateMap(map[string]any{"name": "Electronics"})
	assert.NoError(t, err)

	productID, err := qb.From("test_products").CreateMap(map[string]any{"name": "Laptop", "category_id": categoryID})
	assert.NoError(t, err)

	// 4. Build and execute the JOIN query
	builder := qb.From("test_products").As("p").
		Select("p.name as product_name", "c.name as category_name").
		Join("test_categories as c", "p.category_id = c.id").
		Where("p.id = ?", productID)

	sql, args := builder.ToSql()
	fmt.Printf("Generated SQL: %s\n", sql)
	fmt.Printf("Generated Args: %v\n", args)

	rows, err := builder.Rows()

	// 5. Assert the results
	assert.NoError(t, err)
	assert.Len(t, rows, 1, "Expected to find one result")

	if len(rows) > 0 {
		assert.Equal(t, "Laptop", string(rows[0]["product_name"].([]byte)), "Product name should match")
		assert.Equal(t, "Electronics", string(rows[0]["category_name"].([]byte)), "Category name should match")
	}
}

func TestJoinQueryWithoutAlias(t *testing.T) {
	qb := Connect()

	// 1. Create tables if not exist (assuming they are created in the first test)
	// For isolated tests, you would repeat the table creation logic here.

	// 2. Truncate tables for a clean state
	err := qb.Raw("SET FOREIGN_KEY_CHECKS = 0;").Exec()
	assert.NoError(t, err)
	err = qb.Raw("TRUNCATE TABLE test_products").Exec()
	assert.NoError(t, err)
	err = qb.Raw("TRUNCATE TABLE test_categories").Exec()
	assert.NoError(t, err)
	err = qb.Raw("SET FOREIGN_KEY_CHECKS = 1;").Exec()
	assert.NoError(t, err)

	// 3. Insert test data
	categoryID, err := qb.From("test_categories").CreateMap(map[string]any{"name": "Books"})
	assert.NoError(t, err)

	productID, err := qb.From("test_products").CreateMap(map[string]any{"name": "Go Programming", "category_id": categoryID})
	assert.NoError(t, err)

	// 4. Build and execute the JOIN query without aliases
	builder := qb.From("test_products").
		Select("test_products.name", "test_categories.name as category_name").
		Join("test_categories", "test_products.category_id = test_categories.id").
		Where("test_products.id = ?", productID)

	sql, args := builder.ToSql()
	fmt.Printf("Generated SQL (without alias): %s\n", sql)
	fmt.Printf("Generated Args (without alias): %v\n", args)

	rows, err := builder.Rows()

	// 5. Assert the results
	assert.NoError(t, err)
	assert.Len(t, rows, 1, "Expected to find one result")

	if len(rows) > 0 {
		assert.Equal(t, "Go Programming", string(rows[0]["name"].([]byte)), "Product name should match")
		assert.Equal(t, "Books", string(rows[0]["category_name"].([]byte)), "Category name should match")
	}
}

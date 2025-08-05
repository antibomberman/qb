package tests

import (
	"database/sql"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Product struct {
	ID        int64        `db:"id"`
	Name      string       `db:"name"`
	Quantity  int          `db:"quantity"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

func TestMain(m *testing.M) {
	testingDb := Connect()
	var err error
	var count int
	err = testingDb.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'test_products' AND column_name = 'quantity'").QueryRow(&count)
	if err != nil {
		log.Fatalf("Ошибка при проверке существования столбца quantity: %v", err)
	}

	if count == 0 {
		err = testingDb.Raw("ALTER TABLE test_products ADD COLUMN quantity INT DEFAULT 0").Exec()
		if err != nil {
			log.Fatalf("Ошибка при добавлении столбца quantity: %v", err)
		}
	}
	m.Run()
}

func TestFind(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()

	// Create a product first
	product := Product{Name: "Test Product"}
	id, err := testingDb.From("test_products").Create(&product)
	assert.NoError(t, err)
	assert.NotNil(t, id)

	// Find the product
	var foundProduct Product
	found, err := testingDb.From("test_products").Find(id, &foundProduct)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, product.Name, foundProduct.Name)
}

func TestGet(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()

	// Create some products
	_, err := testingDb.From("test_products").Create(&Product{Name: "Product 1"})
	assert.NoError(t, err)
	_, err = testingDb.From("test_products").Create(&Product{Name: "Product 2"})
	assert.NoError(t, err)

	// Get all products
	var products []Product
	found, err := testingDb.From("test_products").Get(&products)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Len(t, products, 2)
}

func TestFirst(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()

	// Create some products
	_, err := testingDb.From("test_products").Create(&Product{Name: "Product 1"})
	assert.NoError(t, err)
	_, err = testingDb.From("test_products").Create(&Product{Name: "Product 2"})
	assert.NoError(t, err)

	// Get the first product
	var product Product
	found, err := testingDb.From("test_products").OrderByAsc("id").First(&product)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "Product 1", product.Name)
}

func TestCreate(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()

	product := Product{Name: "New Product"}
	id, err := testingDb.From("test_products").Create(&product)
	assert.NoError(t, err)
	assert.NotNil(t, id)

	var createdProduct Product
	found, err := testingDb.From("test_products").Find(id, &createdProduct)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, product.Name, createdProduct.Name)
}

func TestUpdate(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()

	product := Product{Name: "Old Name"}
	id, err := testingDb.From("test_products").Create(&product)
	assert.NoError(t, err)

	updatedProduct := Product{ID: id.(int64), Name: "New Name"}
	err = testingDb.From("test_products").WhereId(id).Update(&updatedProduct)
	assert.NoError(t, err)

	var foundProduct Product
	found, err := testingDb.From("test_products").Find(id, &foundProduct)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, updatedProduct.Name, foundProduct.Name)
}

func TestDelete(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()

	product := Product{Name: "Product to Delete"}
	id, err := testingDb.From("test_products").Create(&product)
	assert.NoError(t, err)

	err = testingDb.From("test_products").WhereId(id).Delete()
	assert.NoError(t, err)

	var foundProduct Product
	found, err := testingDb.From("test_products").Find(id, &foundProduct)
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestIncrement(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()

	product := Product{Name: "Product", Quantity: 10}
	id, err := testingDb.From("test_products").Create(&product)
	assert.NoError(t, err)

	err = testingDb.From("test_products").WhereId(id).Increment("quantity", 5)
	assert.NoError(t, err)

	var foundProduct Product
	found, err := testingDb.From("test_products").Find(id, &foundProduct)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, 15, foundProduct.Quantity)
}

func TestDecrement(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()

	product := Product{Name: "Product", Quantity: 10}
	id, err := testingDb.From("test_products").Create(&product)
	assert.NoError(t, err)

	err = testingDb.From("test_products").WhereId(id).Decrement("quantity", 3)
	assert.NoError(t, err)

	var foundProduct Product
	found, err := testingDb.From("test_products").Find(id, &foundProduct)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, 7, foundProduct.Quantity)
}

func TestBatchInsert(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()

	records := []map[string]any{
		{"name": "Product A"},
		{"name": "Product B"},
	}
	err := testingDb.From("test_products").BatchInsert(records)
	assert.NoError(t, err)

	var products []Product
	found, err := testingDb.From("test_products").Get(&products)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Len(t, products, 2)
}

func TestBulkInsert(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()

	records := []map[string]any{
		{"name": "Product C"},
		{"name": "Product D"},
	}
	err := testingDb.From("test_products").BulkInsert(records)
	assert.NoError(t, err)

	var products []Product
	found, err := testingDb.From("test_products").Get(&products)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Len(t, products, 2)
}

func TestBulkUpdate(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()

	// Create products
	_, err := testingDb.From("test_products").Create(&Product{Name: "Product 1", Quantity: 10})
	assert.NoError(t, err)
	_, err = testingDb.From("test_products").Create(&Product{Name: "Product 2", Quantity: 20})
	assert.NoError(t, err)

	records := []map[string]any{
		{"id": int64(1), "name": "Updated Product 1", "quantity": 15},
		{"id": int64(2), "name": "Updated Product 2", "quantity": 25},
	}
	err = testingDb.From("test_products").BulkUpdate(records, "id")
	assert.NoError(t, err)

	var products []Product
	found, err := testingDb.From("test_products").Get(&products)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Len(t, products, 2)
	assert.Equal(t, "Updated Product 1", products[0].Name)
	assert.Equal(t, 15, products[0].Quantity)
	assert.Equal(t, "Updated Product 2", products[1].Name)
	assert.Equal(t, 25, products[1].Quantity)
}

func TestBatchUpdate(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()

	// Create products
	_, err := testingDb.From("test_products").Create(&Product{Name: "Product 1", Quantity: 10})
	assert.NoError(t, err)
	_, err = testingDb.From("test_products").Create(&Product{Name: "Product 2", Quantity: 20})
	assert.NoError(t, err)
	_, err = testingDb.From("test_products").Create(&Product{Name: "Product 3", Quantity: 30})
	assert.NoError(t, err)

	records := []map[string]any{
		{"id": int64(1), "name": "Batch Updated Product 1", "quantity": 11},
		{"id": int64(2), "name": "Batch Updated Product 2", "quantity": 21},
		{"id": int64(3), "name": "Batch Updated Product 3", "quantity": 31},
	}
	err = testingDb.From("test_products").BatchUpdate(records, "id", 2)
	assert.NoError(t, err)

	var products []Product
	found, err := testingDb.From("test_products").Get(&products)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Len(t, products, 3)
	assert.Equal(t, "Batch Updated Product 1", products[0].Name)
	assert.Equal(t, 11, products[0].Quantity)
	assert.Equal(t, "Batch Updated Product 2", products[1].Name)
	assert.Equal(t, 21, products[1].Quantity)
	assert.Equal(t, "Batch Updated Product 3", products[2].Name)
	assert.Equal(t, 31, products[2].Quantity)
}

func TestUpdateMap(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()

	product := Product{Name: "Original Name", Quantity: 5}
	id, err := testingDb.From("test_products").Create(&product)
	assert.NoError(t, err)

	updateData := map[string]any{
		"name":     "Updated Name via Map",
		"quantity": 15,
	}
	err = testingDb.From("test_products").WhereId(id).UpdateMap(updateData)
	assert.NoError(t, err)

	var foundProduct Product
	found, err := testingDb.From("test_products").Find(id, &foundProduct)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "Updated Name via Map", foundProduct.Name)
	assert.Equal(t, 15, foundProduct.Quantity)
}

func TestRows(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()

	// Create some products
	_, err := testingDb.From("test_products").Create(&Product{Name: "Product A", Quantity: 10})
	assert.NoError(t, err)
	_, err = testingDb.From("test_products").Create(&Product{Name: "Product B", Quantity: 20})
	assert.NoError(t, err)

	rows, err := testingDb.From("test_products").Rows()
	assert.NoError(t, err)
	assert.Len(t, rows, 2)
	assert.Equal(t, "Product A", string(rows[0]["name"].([]uint8))) // Преобразование к string
	assert.Equal(t, int64(10), rows[0]["quantity"])
	assert.Equal(t, "Product B", string(rows[1]["name"].([]uint8))) // Преобразование к string
	assert.Equal(t, int64(20), rows[1]["quantity"])
}

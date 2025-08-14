package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpsert(t *testing.T) {
	// Connect to the database
	qb := Connect()

	// Test case 1: Insert a new record
	_, err := qb.From("products").Upsert(map[string]any{
		"name":     "Test Product 1",
		"quantity": 10,
	}, []string{"name"}, []string{"quantity"})
	assert.NoError(t, err)

	// Verify the record was inserted
	var product Product
	found, err := qb.From("products").Where("name = ?", "Test Product 1").First(&product)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "Test Product 1", product.Name)
	assert.Equal(t, 10, product.Quantity)

	// Test case 2: Update an existing record
	_, err = qb.From("products").Upsert(map[string]any{
		"name":     "Test Product 1",
		"quantity": 20,
	}, []string{"name"}, []string{"quantity"})
	assert.NoError(t, err)

	// Verify the record was updated
	var updatedProduct Product
	found, err = qb.From("products").Where("name = ?", "Test Product 1").First(&updatedProduct)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "Test Product 1", updatedProduct.Name)
	assert.Equal(t, 20, updatedProduct.Quantity)
}

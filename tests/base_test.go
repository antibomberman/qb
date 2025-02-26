package tests

import (
	"testing"

	"github.com/antibomberman/qb"
	"github.com/stretchr/testify/assert"

	"github.com/antibomberman/qb/mocks"
)

func TestQueryBuilder(t *testing.T) {
	mockDB := mocks.NewDBInterface(t)
	builder := mocks.NewQueryBuilderInterface(t)
	t.Run("Query", func(t *testing.T) {
		query := builder.Query("users")

		assert.NotNil(t, query)
		assert.IsType(t, &qb.Builder{}, query)
	})

	t.Run("ExecuteQuery", func(t *testing.T) {
		mockDB.On("Get",
			&struct{ ID int }{},
			"SELECT * FROM users WHERE id = $1",
			1,
		).Return(nil)

		// Выполняем запрос
		var result struct{ ID int }
		found, err := builder.Query("users").Where("id", 1).First(&result)

		assert.NoError(t, err)
		assert.True(t, found)
		mockDB.AssertExpectations(t)
	})

	t.Run("Transaction", func(t *testing.T) {
		mockTx := mocks.NewTxInterface(t)

		mockDB.On("Beginx").Return(mockTx, nil)
		mockTx.On("Commit").Return(nil)

		err := builder.Transaction(func(tx *qb.Transaction) error {
			return nil
		})

		assert.NoError(t, err)
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})
}

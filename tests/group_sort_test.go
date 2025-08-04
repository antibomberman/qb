package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderBy(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, _ := builder.OrderBy("name", "desc").ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` ORDER BY name desc", sql)
}

func TestOrderByAsc(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, _ := builder.OrderByAsc("name").ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` ORDER BY name ASC", sql)
}

func TestOrderByDesc(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, _ := builder.OrderByDesc("name").ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` ORDER BY name DESC", sql)
}

func TestGroupBy(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, _ := builder.GroupBy("category_id").ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` GROUP BY category_id", sql)
}

func TestHaving(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, _ := builder.GroupBy("category_id").Having("COUNT(*) > 1").ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` GROUP BY category_id HAVING COUNT(*) > 1", sql)
}

func TestHavingRaw(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, args := builder.GroupBy("category_id").HavingRaw("COUNT(*) > ?", 1).ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` GROUP BY category_id HAVING COUNT(*) > ?", sql)
	assert.Equal(t, []any{1}, args)
}

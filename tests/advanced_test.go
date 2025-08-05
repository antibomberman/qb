package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWindow(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, _ := builder.Window("SUM(price)", "category_id", "price DESC").ToSql()
	assert.Equal(t, "SELECT SUM(price) OVER (PARTITION BY category_id ORDER BY price DESC) FROM `test_products`", sql)
}

func TestRowNumber(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, _ := builder.RowNumber("category_id", "price DESC", "row_num").ToSql()
	assert.Equal(t, "SELECT ROW_NUMBER() OVER (PARTITION BY category_id ORDER BY price DESC) AS row_num FROM `test_products`", sql)
}

func TestRank(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, _ := builder.Rank("category_id", "price DESC", "rank_in_category").ToSql()
	assert.Equal(t, "SELECT RANK() OVER (PARTITION BY category_id ORDER BY price DESC) AS rank_in_category FROM `test_products`", sql)
}

func TestDenseRank(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, _ := builder.DenseRank("category_id", "price DESC", "dense_rank_in_category").ToSql()
	assert.Equal(t, "SELECT DENSE_RANK() OVER (PARTITION BY category_id ORDER BY price DESC) AS dense_rank_in_category FROM `test_products`", sql)
}

func TestAvg(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()
	testingDb.From("test_products").CreateMap(map[string]any{"name": "Product A", "price": 100})
	testingDb.From("test_products").CreateMap(map[string]any{"name": "Product B", "price": 200})

	avg, err := testingDb.From("test_products").Avg("price")
	assert.NoError(t, err)
	assert.Equal(t, 150.0, avg)
}

func TestSum(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()
	testingDb.From("test_products").CreateMap(map[string]any{"name": "Product A", "price": 100})
	testingDb.From("test_products").CreateMap(map[string]any{"name": "Product B", "price": 200})

	sum, err := testingDb.From("test_products").Sum("price")
	assert.NoError(t, err)
	assert.Equal(t, 300.0, sum)
}

func TestMin(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()
	testingDb.From("test_products").CreateMap(map[string]any{"name": "Product A", "price": 100})
	testingDb.From("test_products").CreateMap(map[string]any{"name": "Product B", "price": 200})

	min, err := testingDb.From("test_products").Min("price")
	assert.NoError(t, err)
	assert.Equal(t, 100.0, min)
}

func TestMax(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()
	testingDb.From("test_products").CreateMap(map[string]any{"name": "Product A", "price": 100})
	testingDb.From("test_products").CreateMap(map[string]any{"name": "Product B", "price": 200})

	max, err := testingDb.From("test_products").Max("price")
	assert.NoError(t, err)
	assert.Equal(t, 200.0, max)
}

func TestLockForUpdate(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, _ := builder.LockForUpdate().ToSql()
	assert.Equal(t, "SELECT * FROM `test_products` FOR UPDATE", sql)
}

func TestLockForShare(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, _ := builder.LockForShare().ToSql()
	assert.Equal(t, "SELECT * FROM `test_products` FOR SHARE", sql)
}

func TestSkipLocked(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, _ := builder.SkipLocked().ToSql()
	assert.Equal(t, "SELECT * FROM `test_products` SKIP LOCKED", sql)
}

func TestNoWait(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, _ := builder.NoWait().ToSql()
	assert.Equal(t, "SELECT * FROM `test_products` NOWAIT", sql)
}

func TestLock(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, _ := builder.Lock("IN SHARE MODE").ToSql()
	assert.Equal(t, "SELECT * FROM `test_products` IN SHARE MODE", sql)
}

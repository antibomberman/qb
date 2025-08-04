package tests

import (
	"testing"

	"github.com/antibomberman/qb"
	"github.com/stretchr/testify/assert"
)

func TestWhere(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, args := builder.Where("name = ?", "test").ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` WHERE name = ?", sql)
	assert.Equal(t, []any{"test"}, args)
}

func TestWhereId(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, args := builder.WhereId(1).ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` WHERE id = ?", sql)
	assert.Equal(t, []any{1}, args)
}

func TestOrWhere(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, args := builder.Where("name = ?", "test").OrWhere("name = ?", "test2").ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` WHERE name = ? OR name = ?", sql)
	assert.Equal(t, []any{"test", "test2"}, args)
}

func TestWhereIn(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, args := builder.WhereIn("id", 1, 2, 3).ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` WHERE id IN (?,?,?)", sql)
	assert.Equal(t, []any{1, 2, 3}, args)
}

func TestWhereGroup(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, args := builder.Where("name = ?", "test").WhereGroup(func(q *qb.Builder) {
		q.Where("id = ?", 1).OrWhere("id = ?", 2)
	}).ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` WHERE name = ? AND (id = ? OR id = ?)", sql)
	assert.Equal(t, []any{"test", 1, 2}, args)
}

func TestOrWhereGroup(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, args := builder.Where("name = ?", "test").OrWhereGroup(func(q *qb.Builder) {
		q.Where("id = ?", 1).OrWhere("id = ?", 2)
	}).ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` WHERE name = ? OR (id = ? OR id = ?)", sql)
	assert.Equal(t, []any{"test", 1, 2}, args)
}

func TestWhereExists(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	subQuery := testingDb.From("test_categories").Where("id = ?", 1)
	sql, args := builder.WhereExists(subQuery).ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` WHERE EXISTS (SELECT * FROM `test_categories` WHERE id = ?)", sql)
	assert.Equal(t, []any{1}, args)
}

func TestWhereNotExists(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	subQuery := testingDb.From("test_categories").Where("id = ?", 1)
	sql, args := builder.WhereNotExists(subQuery).ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` WHERE NOT EXISTS (SELECT * FROM `test_categories` WHERE id = ?)", sql)
	assert.Equal(t, []any{1}, args)
}

func TestWhereNull(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, _ := builder.WhereNull("deleted_at").ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` WHERE deleted_at IS NULL", sql)
}

func TestWhereNotNull(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, _ := builder.WhereNotNull("deleted_at").ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` WHERE deleted_at IS NOT NULL", sql)
}

func TestWhereBetween(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, args := builder.WhereBetween("price", 100, 200).ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` WHERE price BETWEEN ? AND ?", sql)
	assert.Equal(t, []any{100, 200}, args)
}

func TestWhereNotBetween(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, args := builder.WhereNotBetween("price", 100, 200).ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` WHERE price NOT BETWEEN ? AND ?", sql)
	assert.Equal(t, []any{100, 200}, args)
}

func TestWhereRaw(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, args := builder.WhereRaw("price > ? AND price < ?", 100, 200).ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` WHERE price > ? AND price < ?", sql)
	assert.Equal(t, []any{100, 200}, args)
}

func TestOrWhereRaw(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, args := builder.Where("name = ?", "test").OrWhereRaw("price > ?", 100).ToSql()

	assert.Equal(t, "SELECT * FROM `test_products` WHERE name = ? OR price > ?", sql)
	assert.Equal(t, []any{"test", 100}, args)
}

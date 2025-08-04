package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWhereDate(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()
	builder := testingDb.From("test_products")

	now := time.Now()
	sql, args := builder.WhereDate("created_at", "=", now).ToSql()

	assert.Equal(t, fmt.Sprintf("SELECT * FROM `test_products` WHERE DATE(`created_at`) = ?"), sql)
	assert.Equal(t, []any{now.Format("2006-01-02")}, args)
}

func TestWhereBetweenDates(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()
	builder := testingDb.From("test_products")

	start := time.Now()
	end := time.Now().Add(time.Hour * 24)
	sql, args := builder.WhereBetweenDates("created_at", start, end).ToSql()

	assert.Equal(t, fmt.Sprintf("SELECT * FROM `test_products` WHERE DATE(`created_at`) BETWEEN ? AND ?"), sql)
	assert.Equal(t, []any{start.Format("2006-01-02"), end.Format("2006-01-02")}, args)
}

func TestWhereDateTime(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()
	builder := testingDb.From("test_products")

	now := time.Now()
	sql, args := builder.WhereDateTime("created_at", "=", now).ToSql()

	assert.Equal(t, fmt.Sprintf("SELECT * FROM `test_products` WHERE `created_at` = ?"), sql)
	assert.Equal(t, []any{now.Format("2006-01-02 15:04:05")}, args)
}

func TestWhereBetweenDateTime(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()
	builder := testingDb.From("test_products")

	start := time.Now()
	end := time.Now().Add(time.Hour * 24)
	sql, args := builder.WhereBetweenDateTime("created_at", start, end).ToSql()

	assert.Equal(t, fmt.Sprintf("SELECT * FROM `test_products` WHERE `created_at` BETWEEN ? AND ?"), sql)
	assert.Equal(t, []any{start.Format("2006-01-02 15:04:05"), end.Format("2006-01-02 15:04:05")}, args)
}

func TestWhereYear(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()
	builder := testingDb.From("test_products")

	sql, args := builder.WhereYear("created_at", "=", 2023).ToSql()

	assert.Equal(t, fmt.Sprintf("SELECT * FROM `test_products` WHERE EXTRACT(YEAR FROM `created_at`) = ?"), sql)
	assert.Equal(t, []any{2023}, args)
}

func TestWhereMonth(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()
	builder := testingDb.From("test_products")

	sql, args := builder.WhereMonth("created_at", "=", 12).ToSql()

	assert.Equal(t, fmt.Sprintf("SELECT * FROM `test_products` WHERE EXTRACT(MONTH FROM `created_at`) = ?"), sql)
	assert.Equal(t, []any{12}, args)
}

func TestWhereDay(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()
	builder := testingDb.From("test_products")

	sql, args := builder.WhereDay("created_at", "=", 25).ToSql()

	assert.Equal(t, fmt.Sprintf("SELECT * FROM `test_products` WHERE EXTRACT(DAY FROM `created_at`) = ?"), sql)
	assert.Equal(t, []any{25}, args)
}

func TestWhereTime(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()
	builder := testingDb.From("test_products")

	now := time.Now()
	sql, args := builder.WhereTime("created_at", "=", now).ToSql()

	assert.Equal(t, fmt.Sprintf("SELECT * FROM `test_products` WHERE TIME(`created_at`) = ?"), sql)
	assert.Equal(t, []any{now.Format("15:04:05")}, args)
}

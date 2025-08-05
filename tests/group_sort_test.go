package tests

import (
	"fmt"
	testing "testing"

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

func TestSelectCount(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")
	sql, _ := builder.Select("COUNT(*)").ToSql()
	assert.Equal(t, "SELECT COUNT(*) FROM `test_products`", sql)
}

func TestComplexSelectWithCountAndAlias(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, args := builder.
		Select("COUNT(*) AS total_products", "category_id").
		GroupBy("category_id").
		ToSql()

	fmt.Printf("Generated SQL: %s\n", sql)
	fmt.Printf("Generated Args: %v\n", args)

	expectedSQL := "SELECT COUNT(*) AS total_products, `category_id` FROM `test_products` GROUP BY category_id"
	assert.Equal(t, expectedSQL, sql)
	assert.Empty(t, args)
}

func TestSelectMultipleFunctionsAndAliases(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, args := builder.
		Select("COUNT(*) AS total_products", "SUM(price) AS total_price", "AVG(quantity) AS avg_quantity", "category_id").
		GroupBy("category_id").
		Having("COUNT(*) > 5").
		ToSql()

	fmt.Printf("Generated SQL (Multiple Functions): %s\n", sql)
	fmt.Printf("Generated Args (Multiple Functions): %v\n", args)

	expectedSQL := "SELECT COUNT(*) AS total_products, SUM(price) AS total_price, AVG(quantity) AS avg_quantity, `category_id` FROM `test_products` GROUP BY category_id HAVING COUNT(*) > 5"
	assert.Equal(t, expectedSQL, sql)
	assert.Empty(t, args)
}

func TestSelectMixedColumnsAndExpressions(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, args := builder.
		Select("name", "price AS product_price", "price * quantity AS total_value", "LOWER(name)", "id").
		ToSql()

	fmt.Printf("Generated SQL (Mixed Columns): %s\n", sql)
	fmt.Printf("Generated Args (Mixed Columns): %v\n", args)

	expectedSQL := "SELECT `name`, `price` AS product_price, price * quantity AS total_value, LOWER(name), `id` FROM `test_products`"
	assert.Equal(t, expectedSQL, sql)
	assert.Empty(t, args)
}

func TestSelectComplexCaseExpression(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, args := builder.
		Select("CASE WHEN quantity > 10 THEN 'High' WHEN quantity <= 10 THEN 'Low' ELSE 'Unknown' END AS stock_level", "name").
		ToSql()

	fmt.Printf("Generated SQL (Complex Case): %s\n", sql)
	fmt.Printf("Generated Args (Complex Case): %v\n", args)

	expectedSQL := "SELECT CASE WHEN quantity > 10 THEN 'High' WHEN quantity <= 10 THEN 'Low' ELSE 'Unknown' END AS stock_level, `name` FROM `test_products`"
	assert.Equal(t, expectedSQL, sql)
	assert.Empty(t, args)
}

func TestLimitOffset(t *testing.T) {
	testingDb := Connect()
	builder := testingDb.From("test_products")

	sql, _ := builder.Limit(10).Offset(5).ToSql()
	assert.Equal(t, "SELECT * FROM `test_products` LIMIT 10 OFFSET 5", sql)
}

func TestPluck(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()

	// Create some products
	testingDb.From("test_products").Create(&Product{Name: "Apple"})
	testingDb.From("test_products").Create(&Product{Name: "Banana"})
	testingDb.From("test_products").Create(&Product{Name: "Orange"})

	var names []string
	err := testingDb.From("test_products").OrderByAsc("id").Pluck("name", &names)
	assert.NoError(t, err)
	assert.Equal(t, []string{"Apple", "Banana", "Orange"}, names)
}

func TestValue(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()

	// Create a product
	id, _ := testingDb.From("test_products").Create(&Product{Name: "Single Product", Quantity: 100})

	nameAny, err := testingDb.From("test_products").WhereId(id).Value("name")
	assert.NoError(t, err)
	name, ok := nameAny.([]uint8)
	assert.True(t, ok)
	assert.Equal(t, "Single Product", string(name))

	quantity, err := testingDb.From("test_products").WhereId(id).Value("quantity")
	assert.NoError(t, err)
	assert.Equal(t, int64(100), quantity)
}

func TestValues(t *testing.T) {
	testingDb := Connect()
	testingDb.Raw("truncate table test_products").Exec()

	// Create some products
	testingDb.From("test_products").Create(&Product{Name: "Item 1", Quantity: 10})
	testingDb.From("test_products").Create(&Product{Name: "Item 2", Quantity: 20})
	testingDb.From("test_products").Create(&Product{Name: "Item 3", Quantity: 30})

	quantities, err := testingDb.From("test_products").OrderByAsc("id").Values("quantity")
	assert.NoError(t, err)
	assert.Equal(t, []any{int64(10), int64(20), int64(30)}, quantities)
}

func TestDistinct(t *testing.T) {
	testingDb := Connect()

	// Test case 1: Distinct with specific columns
	builder1 := testingDb.From("test_products")
	sql1, _ := builder1.Distinct("category_id").ToSql()
	assert.Equal(t, "SELECT DISTINCT `category_id` FROM `test_products`", sql1)

	// Test case 2: Distinct without arguments (SELECT DISTINCT *)
	builder2 := testingDb.From("test_products") // Use a new builder
	sql2, _ := builder2.Distinct().ToSql()
	assert.Equal(t, "SELECT DISTINCT * FROM `test_products`", sql2)
}

func TestSubQuery(t *testing.T) {
	testingDb := Connect()
	// SubQuery for a simple select
	sub := testingDb.From("orders").Select("user_id").Where("amount > ?", 100)
	sql, args := sub.SubQuery("sub_users").ToSql()
	assert.Equal(t, "(SELECT `user_id` FROM `orders` WHERE amount > ?) AS sub_users", sql)
	assert.Equal(t, []any{100}, args)

	// SubQuery in a WHERE IN clause
	mainQuery := testingDb.From("users").WhereIn("`id`", sub.SubQuery("")) // Empty alias for direct use
	sql, args = mainQuery.ToSql()
	assert.Equal(t, "SELECT * FROM `users` WHERE `id` IN (SELECT `user_id` FROM `orders` WHERE amount > ?)", sql)
	assert.Equal(t, []any{100}, args)
}

func TestWhereSubQuery(t *testing.T) {
	testingDb := Connect()
	sub := testingDb.From("orders").Select("user_id").Where("status = ?", "completed")
	sql, args := testingDb.From("users").WhereSubQuery("`id`", "IN", sub).ToSql()
	assert.Equal(t, "SELECT * FROM `users` WHERE `id` IN (SELECT `user_id` FROM `orders` WHERE status = ?)", sql)
	assert.Equal(t, []any{"completed"}, args)
}

func TestUnion(t *testing.T) {
	testingDb := Connect()
	qb1 := testingDb.From("users").Select("name").Where("age > ?", 18)
	qb2 := testingDb.From("customers").Select("name").Where("city = ?", "New York")

	combined, args := qb1.Union(qb2).ToSql()
	assert.Equal(t, "(SELECT `name` FROM `users` WHERE age > ?) UNION (SELECT `name` FROM `customers` WHERE city = ?)", combined)
	assert.Equal(t, []any{18, "New York"}, args)
}

func TestUnionAll(t *testing.T) {
	testingDb := Connect()
	qb1 := testingDb.From("users").Select("name").Where("age > ?", 18)
	qb2 := testingDb.From("customers").Select("name").Where("city = ?", "New York")

	combined, args := qb1.UnionAll(qb2).ToSql()
	assert.Equal(t, "(SELECT `name` FROM `users` WHERE age > ?) UNION ALL (SELECT `name` FROM `customers` WHERE city = ?)", combined)
	assert.Equal(t, []any{18, "New York"}, args)
}

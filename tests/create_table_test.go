package tests

import (
	"testing"

	"github.com/antibomberman/dblayer"
)

func TestCreateTable(t *testing.T) {
	dbl, err := ConnectDB()
	if err != nil {
		t.Fatal(err)
	}
	dbl.Raw("DROP TABLE users").Exec()
	dbl.Raw("DROP TABLE posts").Exec()

	// Тест создания таблицы
	err = dbl.CreateTableIfNotExists("users", func(schema *dblayer.Schema) {
		schema.ID()
		schema.String("username", 50)
		schema.String("email", 100).NotNull().Unique()
		schema.Phone("phone")
		schema.Password("password")
		schema.Timestamps()
	})
	if err != nil {
		t.Fatal(err)
	}
	// Тест создания таблицы
	err = dbl.CreateTableIfNotExists("posts", func(schema *dblayer.Schema) {
		schema.ID()
		schema.BigInteger("user_id").Unsigned().Nullable().Foreign("users").References("users", "id").CascadeOnDelete()

		schema.BigInteger("_user_id").Unsigned()

		schema.String("title", 255)
		schema.Text("content")
		schema.Timestamps()
		schema.Foreign("_user_id").References("users", "id").CascadeOnDelete()

	})

	if err != nil {
		t.Fatal(err)
	}

}

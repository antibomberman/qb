package tests

import (
	"github.com/antibomberman/dblayer/table"
	"testing"
)

func TestCreateTable(t *testing.T) {
	dbl, err := ConnectDB()
	if err != nil {
		t.Fatal(err)
	}
	dbl.Table("users").Drop()
	dbl.Table("posts").Drop()

	// Тест создания таблицы
	err = dbl.Table("users").CreateIfNotExists(func(schema *table.Builder) {
		schema.ID()
		schema.String("username", 50)
		schema.String("email", 100)
		schema.Phone("phone")
		schema.Password("password")
		schema.Timestamps()
	})
	if err != nil {
		t.Fatal(err)
	}
	// Тест создания таблицы
	err = dbl.Table("posts").CreateIfNotExists(func(schema *table.Builder) {
		schema.ID()
		schema.BigInteger("user_id").Unsigned().Nullable().Foreign("users").References("users", "id").CascadeOnDelete()

		schema.BigInteger("_user_id").Unsigned()

		schema.String("def", 255)
		schema.String("title", 255).NotNull()
		schema.Text("content").Nullable()
		schema.Timestamps()
		schema.SoftDeletes()
		schema.Foreign("_user_id").References("users", "id").CascadeOnDelete()

	})

	if err != nil {
		t.Fatal(err)
	}

}

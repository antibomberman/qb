package tests

import (
	"fmt"
	"github.com/antibomberman/dblayer"
	"testing"
)

func TestCreateTable(t *testing.T) {
	dbl, err := ConnectDB()
	if err != nil {
		t.Fatal(err)
	}
	dbl.Raw("DROP TABLE users").Exec()
	// Тест создания таблицы
	err = dbl.CreateTableIfNotExists("users", func(schema *dblayer.Schema) {
		schema.BigInteger("id").Unsigned().Primary().AutoIncrement()
		schema.String("username", 50)
		schema.String("email", 100).NotNull().Unique()
		schema.Phone("phone")
		schema.Password("password")
		schema.Timestamps()
	})
	if err != nil {
		t.Errorf("Ошибка создания таблицы: %v", err)
	}
	user := User{
		Username: "test",
		Email:    "test@example.com",
		Phone:    "1234567890",
		Password: "password",
	}

	_, err = dbl.Table("users").Create(user)
	if err != nil {
		t.Errorf("Ошибка создания записи в таблице: %v", err)
	}
	fmt.Println("--------------------------------------")
	dbl.UpdateTable("users", func(schema *dblayer.Schema) {
		// Если колонки нет - ADD COLUMN
		schema.String("new_column", 255)

		// Если колонка уже есть - MODIFY COLUMN
		schema.String("username", 100)
	})
	dbl.UpdateTable("users", func(schema *dblayer.Schema) {
		schema.DropColumn("new_column")

	})

	count, err := dbl.Table("users").Count()
	if err != nil {
		t.Fatalf("Ошибка получения количества записей в таблице: %v", err)
	}

	fmt.Println(count)
}

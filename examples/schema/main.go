package main

import (
	"database/sql"
	"log"

	"github.com/antibomberman/dblayer"
)

var DBLayer *dblayer.DBLayer

func main() {

	db, err := sql.Open("sqlite", "./examples/example.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	DBLayer = dblayer.New("sqlite", db)

}

// CreateTable создает таблицу
func CreateTable() {
	DBLayer.CreateTable("users", func(table *dblayer.Schema) {
		table.BigIncrements("id")
		table.String("name", 255)
		table.String("email", 255)
		table.Timestamps()
	})
}
func UpdateTable() {
	// Обновление таблицы
	DBLayer.UpdateTable("users", func(table *dblayer.Schema) {
		// Добавление новых колонок - тот же API, что и при создании
		table.String("phone", 20)
		// Специфичные для обновления операции
		table.RenameColumn("name", "full_name")
		table.DropColumn("old_field")
		table.ModifyColumn(dblayer.Column{
			Name:   "email",
			Type:   "VARCHAR",
			Length: 255,
			Unique: true,
		})
	})
}

func BuildTable() {
	// Создание таблицы пользователей
	DBLayer.CreateTable("users", func(table *dblayer.Schema) {
		table.Column("id").Type("bigint").AutoIncrement().Primary().Add()
		table.Column("name").Type("varchar", 255).Comment("Имя пользователя").Add()
		table.Column("email").Type("varchar", 255).Unique().Add()
		table.Column("password").Type("varchar", 255).Add()
		table.Column("status").Type("enum", 20).Default("active").Add()
		table.Column("created_at").Type("timestamp").Default("CURRENT_TIMESTAMP").Add()
		table.Column("updated_at").Type("timestamp").Nullable().Add()

		table.UniqueKey("uk_email", "email")
		table.Index("idx_status", "status")
		table.Comment("Таблица пользователей")
	})
	// Создание таблицы заказов с внешними ключами
	DBLayer.CreateTable("orders", func(table *dblayer.Schema) {
		table.Column("id").Type("bigint").AutoIncrement().Primary().Add()
		table.Column("user_id").Type("bigint").Add()
		table.Column("total").Type("decimal", 10).Default(0).Add()
		table.Column("status").Type("varchar", 50).Default("pending").Add()
		table.Column("created_at").Type("timestamp").Default("CURRENT_TIMESTAMP").Add()
		table.ForeignKey("user_id", "users", "id").
			OnDelete("CASCADE").
			OnUpdate("CASCADE").
			Add()
		table.Index("idx_user", "user_id")
		table.Index("idx_status", "status")

	})

}

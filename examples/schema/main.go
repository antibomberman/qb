package main

import (
	"database/sql"
	"log"

	"github.com/antibomberman/DBL"
	_ "modernc.org/sqlite"
)

var DBL *DBL.DBL

func main() {

	db, err := sql.Open("sqlite", "./examples/example.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	DBL = DBL.New("sqlite", db)
	BuildTable()
}

// CreateTable создает таблицу
func CreateTable() {
	DBL.CreateTable("users", func(table *DBL.Schema) {
		table.ID()
		table.String("name", 255)
		table.String("email", 255)
		table.Timestamps()
	})
}
func UpdateTable() {
	// Обновление таблицы
	DBL.UpdateTable("users", func(table *DBL.Schema) {
		// Добавление новых колонок - тот же API, что и при создании
		table.String("phone", 20)
		// Специфичные для обновления операции
		table.RenameColumn("name", "full_name")
		table.DropColumn("old_field")

	})
}

func BuildTable() {
	err := DBL.CreateTable("users", func(table *DBL.Schema) {
		table.Column("id").Type("bigint").AutoIncrement().Primary()
		table.Column("name").Type("varchar", 255).Comment("Имя пользователя")
		table.Column("email").Type("varchar", 255).Unique()
		table.Column("password").Type("varchar", 255)
		table.Column("status").Type("enum", 20).Default("active")
		table.Column("created_at").Type("timestamp").Default("CURRENT_TIMESTAMP")
		table.Column("updated_at").Type("timestamp").Nullable()

		table.UniqueKey("uk_email", "email")
		table.Index("idx_status", "status")
		table.Comment("Таблица пользователей")
	})
	if err != nil {
		log.Fatal(err)
	}

}

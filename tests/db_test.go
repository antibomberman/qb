package tests

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/antibomberman/dblayer"
	_ "github.com/go-sql-driver/mysql"
)

func TestCreateTable(t *testing.T) {
	// Подготовка тестовой БД
	db, err := sql.Open("mysql", "user:password@/testdb")
	if err != nil {
		t.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	dbl := dblayer.New("mysql", db)

	// Тест создания таблицы
	err = dbl.CreateTable("users", func(schema *dblayer.Schema) {
		schema.Integer("id").Primary().AutoIncrement()
		schema.String("username", 50).NotNull()
		schema.String("email", 100).NotNull().Unique()
		schema.Timestamp("created_at").Default("CURRENT_TIMESTAMP")
		schema.Timestamp("updated_at").Default("CURRENT_TIMESTAMP").OnUpdate("CURRENT_TIMESTAMP")
	})

	if err != nil {
		t.Errorf("Ошибка создания таблицы: %v", err)
	}

	// Проверка существования таблицы
	var tableName string
	err = db.QueryRow("SHOW TABLES LIKE 'users'").Scan(&tableName)
	if err != nil {
		t.Errorf("Таблица не была создана: %v", err)
	}

	// Проверка структуры таблицы
	rows, err := db.Query("DESCRIBE users")
	if err != nil {
		t.Errorf("Ошибка получения структуры таблицы: %v", err)
	}
	defer rows.Close()

	expectedColumns := map[string]string{
		"id":         "int",
		"username":   "varchar(50)",
		"email":      "varchar(100)",
		"created_at": "timestamp",
		"updated_at": "timestamp",
	}

	for rows.Next() {
		var field, fieldType string
		var null, key, extra, defaultValue sql.NullString
		err := rows.Scan(&field, &fieldType, &null, &key, &defaultValue, &extra)
		if err != nil {
			t.Errorf("Ошибка сканирования строки: %v", err)
		}

		expectedType, exists := expectedColumns[field]
		if !exists {
			t.Errorf("Неожиданное поле: %s", field)
		} else if !strings.Contains(strings.ToLower(fieldType), expectedType) {
			t.Errorf("Неверный тип для поля %s: ожидался %s, получен %s", field, expectedType, fieldType)
		}
	}

	// Очистка после теста
	_, err = db.Exec("DROP TABLE users")
	if err != nil {
		t.Errorf("Ошибка удаления тестовой таблицы: %v", err)
	}
}

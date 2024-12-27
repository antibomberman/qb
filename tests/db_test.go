package tests

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/antibomberman/dblayer"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

const (
	maxAttempts = 3
	timeout     = time.Second * 3
)

func TestMysqlCreateTable(t *testing.T) {
	ctx := context.Background()
	dsn := "test_user:test_password@tcp(localhost:3307)/test_db"
	dbl, err := dblayer.Connection(ctx, "mysql", dsn, maxAttempts, timeout)
	if err != nil {
		t.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer dbl.Close()

	err = dbl.Ping()
	if err != nil {
		t.Fatalf("Ошибка подключения к БД: %v", err)
	}

	// Тест создания таблицы
	err = dbl.CreateTableIfNotExists("users", func(schema *dblayer.Schema) {
		schema.BigInteger("id").Unsigned().Primary().AutoIncrement()
		schema.String("username", 50)
		schema.String("email", 100).NotNull().Unique()
		schema.Timestamp("created_at").Default("CURRENT_TIMESTAMP")
		schema.Timestamp("updated_at").Default("CURRENT_TIMESTAMP").OnUpdate("CURRENT_TIMESTAMP")
	})

	if err != nil {
		t.Errorf("Ошибка создания таблицы: %v", err)
	}

	count, err := dbl.Table("users").Count()
	if err != nil {
		t.Fatalf("Ошибка получения количества записей в таблице: %v", err)
	}
	fmt.Println(count)
}
func TestPostgresCreateTable(t *testing.T) {
	// Подготовка тестовой БД
	db, err := sql.Open("postgres", "user=test_user password=test_password host=localhost port=5433 dbname=test_db sslmode=disable")
	if err != nil {
		t.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		t.Fatalf("Ошибка подключения к БД: %v", err)
	}

	dbl := dblayer.New("postgres", db)

	// Тест создания таблицы
	err = dbl.CreateTableIfNotExists("users", func(schema *dblayer.Schema) {
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
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'users'
		)`).Scan(&exists)
	if err != nil {
		t.Errorf("Ошибка проверки существования таблицы: %v", err)
	}
	if !exists {
		t.Error("Таблица не была создана")
	}

	// Проверка структуры таблицы
	rows, err := db.Query(`
		SELECT column_name, data_type, character_maximum_length 
		FROM information_schema.columns 
		WHERE table_name = 'users'`)
	if err != nil {
		t.Errorf("Ошибка получения структуры таблицы: %v", err)
	}
	defer rows.Close()

	expectedColumns := map[string]string{
		"id":         "integer",
		"username":   "character varying",
		"email":      "character varying",
		"created_at": "timestamp without time zone",
		"updated_at": "timestamp without time zone",
	}

	for rows.Next() {
		var field, fieldType string
		var maxLength sql.NullInt64
		err := rows.Scan(&field, &fieldType, &maxLength)
		if err != nil {
			t.Errorf("Ошибка сканирования строки: %v", err)
		}

		expectedType, exists := expectedColumns[field]
		if !exists {
			t.Errorf("Неожиданное поле: %s", field)
		} else if fieldType != expectedType {
			t.Errorf("Неверный тип для поля %s: ожидался %s, получен %s", field, expectedType, fieldType)
		}
	}

	// Очистка после теста
	_, err = db.Exec("DROP TABLE users")
	if err != nil {
		t.Errorf("Ошибка удаления тестовой таблицы: %v", err)
	}
}

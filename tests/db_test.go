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

type User struct {
	ID        int64      `db:"id"`
	Username  string     `db:"username"`
	Email     string     `db:"email"`
	Phone     string     `db:"phone"`
	Password  string     `db:"password"`
	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

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
	now := time.Now()
	user := User{
		Username:  "test",
		Email:     "test@example.com",
		Phone:     "1234567890",
		Password:  "password",
		CreatedAt: &now,
	}

	_, err = dbl.Table("users").Create(user)
	if err != nil {
		t.Errorf("Ошибка создания записи в таблице: %v", err)
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

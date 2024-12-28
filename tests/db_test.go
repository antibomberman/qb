package tests

import (
	"context"
	"testing"
	"time"

	"github.com/antibomberman/dblayer"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

const (
	maxAttempts = 3
	timeout     = time.Second * 3
	mysqlDSN    = "test_user:test_password@tcp(localhost:3307)/test_db"
	postgresDSN = "user=test_user password=test_password host=localhost port=5433 dbname=test_db sslmode=disable"
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

func TestAuditTable(t *testing.T) {
	ctx := context.Background()
	dbl, err := dblayer.Connection(ctx, "mysql", mysqlDSN, maxAttempts, timeout)
	if err != nil {
		t.Fatalf("Ошибка подключения к БД: %v", err)
	}
	//defer dbl.Close()

	err = dbl.Ping()
	if err != nil {
		t.Fatalf("Ошибка подключения к БД: %v", err)
	}

	// err = dbl.AuditTableCreate()
	// if err != nil {
	// 	t.Fatalf("Ошибка создания таблицы аудита: %v", err)
	// }
	user := User{
		Username: "tes3t",
		Email:    "t2est@example.com",
		Phone:    "1",
		Password: "password",
	}
	// _, err = dbl.Table("users").WithAudit(1).Create(user)
	// if err != nil {
	// 	t.Fatalf("Ошибка создания записи в таблице: %v", err)
	// }
	//err = dbl.Table("users").WhereId(1).UpdateMap(map[string]interface{}{
	//	"email": "test4@example.com",
	//})
	err = dbl.Table("users").WithAudit(1).WhereId(7).Update(user)
	if err != nil {
		t.Fatalf("Ошибка обновления записи в таблице: %v", err)
	}

}

package tests

import (
	"database/sql"
	"time"
)

const (
	maxAttempts = 3
	timeout     = time.Second * 3
	dsn         = "test_user:test_password@tcp(localhost:3307)/test_db?parseTime=true"
	//postgresDSN = "user=test_user password=test_password host=localhost port=5433 dbname=test_db sslmode=disable"
	driver = "mysql"
)

type User struct {
	ID        int64        `db:"id"`
	Username  string       `db:"username"`
	Email     string       `db:"email"`
	Phone     string       `db:"phone"`
	Password  string       `db:"password"`
	CreatedAt sql.NullTime `db:"created_at"`
	UpdatedAt sql.NullTime `db:"updated_at"`
}

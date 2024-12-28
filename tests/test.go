package tests

import (
	"time"
)

const (
	maxAttempts = 3
	timeout     = time.Second * 3
	dsn         = "test_user:test_password@tcp(localhost:3307)/test_db"
	//postgresDSN = "user=test_user password=test_password host=localhost port=5433 dbname=test_db sslmode=disable"
	driver = "mysql"
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

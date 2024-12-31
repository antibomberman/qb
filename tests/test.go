package tests

import (
	"context"
	"database/sql"
	"github.com/antibomberman/DBL"
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

func ConnectDB() (*DBL.DBL, error) {
	ctx := context.Background()
	dbl, err := DBL.Connection(ctx, driver, dsn, maxAttempts, timeout)
	if err != nil {
		return nil, err
	}
	err = dbl.Ping()
	if err != nil {
		return nil, err
	}

	return dbl, nil

}

package internal

import (
	"context"
	dblayer "github.com/antibomberman/dblayer"
	"time"
)

const (
	maxAttempts = 3
	timeout     = time.Second * 3
	dsn         = "test_user:test_password@tcp(localhost:3307)/test_db?parseTime=true"
	//postgresDSN = "user=test_user password=test_password host=localhost port=5433 dbname=test_db sslmode=disable"
	driver = "mysql"
)

func ConnectDB() (*dblayer.DBLayer, error) {
	ctx := context.Background()
	dbl, err := dblayer.Connection(ctx, driver, dsn, maxAttempts, timeout)
	if err != nil {
		return nil, err
	}
	err = dbl.Ping()
	if err != nil {
		return nil, err
	}

	return dbl, nil

}

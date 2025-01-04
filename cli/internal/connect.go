package internal

import (
	"context"
	"fmt"
	"github.com/antibomberman/dblayer"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

const (
	dsn = "test_user:test_password@tcp(localhost:3307)/test_db?parseTime=true"
	//postgresDSN = "user=test_user password=test_password host=localhost port=5433 dbname=test_db sslmode=disable"
	driver = "mysql"
)

func ConnectDB() (*dblayer.DBLayer, error) {
	ctx := context.Background()

	cfg := Load()
	timeDuration := time.Duration(cfg.Timeout) * time.Second
	fmt.Println(cfg.Driver, cfg.DSN)
	dbl, err := dblayer.Connection(ctx, cfg.Driver, cfg.DSN, cfg.MaxAttempts, timeDuration)
	if err != nil {
		return nil, err
	}
	err = dbl.Ping()
	if err != nil {
		return nil, err
	}

	return dbl, nil

}

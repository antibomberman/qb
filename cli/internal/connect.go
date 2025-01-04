package internal

import (
	"context"
	"fmt"
	"github.com/antibomberman/dblayer"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
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
	dbl := dblayer.Connect(driver, dsn)
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
func ConnectDB2() (*dblayer.DBLayer, error) {
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return dblayer.NewX(driver, db), nil
}

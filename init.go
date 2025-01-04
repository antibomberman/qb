package dblayer

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

func New(driverName string, db *sql.DB) *DBLayer {
	x := sqlx.NewDb(db, driverName)
	if x == nil {
		panic("Couldn't create dataBBLayer new method")
	}
	d := &DBLayer{
		DB:         x,
		DriverName: driverName,
	}

	return d
}
func NewX(driverName string, dbx *sqlx.DB) *DBLayer {
	if dbx == nil {
		panic("Couldn't create dataDBLayer newX method")
	}
	d := &DBLayer{
		DB:         dbx,
		DriverName: driverName,
	}

	return d
}
func Connect(driverName string, dataSourceName string) *DBLayer {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		panic(err)
	}
	return New(driverName, db)
}
func Connection(ctx context.Context, driverName string, dataSourceName string, maxAttempts int, connectionTimeout time.Duration) (*DBLayer, error) {

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		log.Printf("Attempting to connect (attempt %d/%d)", attempt, maxAttempts)

		db, err := sql.Open(driverName, dataSourceName)
		if err != nil {
			log.Printf("Failed to open dataBBLayer connection: %v", err)
			time.Sleep(connectionTimeout)
			continue
		}
		ctx, cancel := context.WithTimeout(ctx, connectionTimeout)
		err = db.PingContext(ctx)
		cancel()

		if err == nil {
			log.Println("Connected to dataDBLayer")
			return New(driverName, db), nil
		}
		log.Printf("Failed to ping dataDBLayer: %v", err)
		db.Close()
		time.Sleep(connectionTimeout)
	}

	return nil, fmt.Errorf("failed to connect to dataDBLayer after %d attempts", maxAttempts)
}

func (d *DBLayer) Close() error {
	return d.DB.Close()
}
func (d *DBLayer) Ping() error {
	return d.DB.Ping()
}

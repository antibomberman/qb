package main

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/antibomberman/qb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jaswdr/faker/v2"
	_ "github.com/lib/pq"
)

type Product struct {
	ID        int          `db:"id" json:"id"`
	Name      string       `db:"name" json:"name"`
	Price     float64      `db:"price" json:"price"`
	IsActive  bool         `db:"is_active" json:"is_active"`
	CreatedAt sql.NullTime `db:"created_at" json:"created_at"`
	UpdatedAt sql.NullTime `db:"updated_at" json:"updated_at"`
}

func connectMysql() (string, *sql.DB, error) {
	op := "examples.connectMysql"
	db, err := sql.Open("mysql", "test_user:test_password@tcp(localhost:3316)/test_db")
	if err != nil {
		log.Println(op, err)
		return "", nil, err
	}

	db.SetMaxOpenConns(100) // Было 25000
	db.SetMaxIdleConns(10)  // Было 50000
	db.SetConnMaxLifetime(5 * time.Minute)
	err = db.Ping()
	if err != nil {
		log.Println(op+"ping ", err)
		return "", nil, err
	}
	return "mysql", db, nil
}
func connectPostgres() (string, *sql.DB, error) {
	op := "examples.connectPostgres"
	db, err := sql.Open("postgres", "postgres://test_user:test_password@localhost:5442/test_db")
	if err != nil {
		log.Println(op, err)
	}

	db.SetMaxOpenConns(250)                // Максимум 25 открытых соединений
	db.SetMaxIdleConns(50)                 // Поддерживать 5 соединений в режиме ожидания
	db.SetConnMaxLifetime(5 * time.Minute) // Максимальное время жизни соединения

	err = db.Ping()
	if err != nil {
		log.Println(op+"ping ", err)
		return "", nil, err
	}

	return "postgres", db, nil
}

func main() {

	driver, db, err := connectMysql()

	if err != nil {
		log.Fatal("connect", err)
	}
	queryBuilder := qb.New(driver, db)

	fake := faker.New()

	err = queryBuilder.Raw(`
		CREATE TABLE IF NOT EXISTS products (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255),
			price DECIMAL(10, 2),
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		);
	`).Exec()
	if err != nil {
		log.Fatal("create table", err)
	}
	err = queryBuilder.Raw("truncate table products").Exec()
	if err != nil {
		log.Fatal("truncate", err)
	}

	//startTime := time.Now()
	//createFromMap(queryBuilder, fake)
	//elapsed := time.Since(startTime)
	//log.Println("createFromMap", elapsed)
	//
	startTime := time.Now()
	createFromStruct(queryBuilder, fake)
	elapsed := time.Since(startTime)
	log.Println("createFromStruct", elapsed)

	//startTime := time.Now()
	//createFromMapWg(queryBuilder, fake)
	//elapsed := time.Since(startTime)
	//log.Println("createFromStructAsync", elapsed)

}

func createFromStruct(queryBuilder qb.QueryBuilderInterface, fake faker.Faker) {
	product := Product{
		Name:     fake.Person().Name(),
		Price:    fake.Float64(100, 1000, 10000),
		IsActive: fake.Bool(),
	}
	for i := 0; i < 100; i++ {
		_, err := queryBuilder.From("products").Create(product)
		if err != nil {
			log.Fatal(err)
		}
	}
}
func createFromMap(queryBuilder qb.QueryBuilderInterface, fake faker.Faker) {

	for i := 0; i < 100; i++ {
		_, err := queryBuilder.From("products").CreateMap(map[string]any{
			"name":       fake.Person().Name(),
			"price":      fake.Float64(100, 1000, 10000),
			"is_active":  fake.Bool(),
			"created_at": fake.Time().Time(time.Now()),
			"updated_at": fake.Time().Time(time.Now()),
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}
func createFromStructAsync(queryBuilder qb.QueryBuilderInterface, fake faker.Faker) {
	product := Product{
		Name:     fake.Person().Name(),
		Price:    fake.Float64(100, 1000, 10000),
		IsActive: fake.Bool(),
	}
	for i := 0; i < 100; i++ {
		idChan, errChan := queryBuilder.From("products").CreateAsync(product)
		select {
		case id := <-idChan:
			log.Println(id)
		case err := <-errChan:
			log.Fatal(err)
		}
	}
}
func createFromMapWg(queryBuilder qb.QueryBuilderInterface, fake faker.Faker) {
	totalRecords := 1_000_000 // Всего записей
	maxConcurrent := 100      // Максимум одновременных горутин
	sem := make(chan struct{}, maxConcurrent)
	wg := sync.WaitGroup{}

	for i := 0; i < totalRecords; i++ {
		wg.Add(1)
		sem <- struct{}{} // Блокируем, если достигнут максимум горутин

		go func() {
			defer wg.Done()
			defer func() { <-sem }() // Освобождаем слот в семафоре

			_, err := queryBuilder.From("products").CreateMap(map[string]any{
				"name":       fake.Person().Name(),
				"price":      fake.Float64(100, 1000, 10000),
				"is_active":  fake.Bool(),
				"created_at": time.Now(),
				"updated_at": time.Now(),
			})
			if err != nil {
				log.Printf("Ошибка вставки: %v", err)
			}
		}()
	}
	wg.Wait()
}

package advanced

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/antibomberman/dblayer"
	QB "github.com/antibomberman/dblayer/query"
	_ "github.com/go-sql-driver/mysql"
)

func main() {

	// Подключение к базе данных
	// Connect to database
	db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/testdb?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	dbl := dblayer.New("mysql", db)
	_, err = contextExample(dbl)
	if err != nil {
		fmt.Printf("Error count table")
	}
	// Пример использования очередей
	// Queue usage example
	if err := queueExample(dbl); err != nil {
		log.Fatal(err)
	}
	// Пример использования событий
	// Events usage example
	if err := eventsExample(dbl); err != nil {
		log.Fatal(err)
	}
	// Пример геопространственных запросов
	// Geospatial queries example
	if err := geoExample(dbl); err != nil {
		log.Fatal(err)
	}
}

func queueExample(dbl *dblayer.DBLayer) error {
	// Создаем отложенную операцию
	// Create a delayed operation
	err := dbl.Query("queued_operations").Queue(
		"send_email",
		map[string]interface{}{
			"to":      "user@example.com",
			"subject": "Hello!",
			"body":    "This is a delayed message",
		},
		time.Now().Add(1*time.Hour),
	)
	if err != nil {
		return err
	}
	// Обработка очереди
	// Process the queue
	return dbl.Query("queued_operations").ProcessQueue(func(op QB.QueuedOperation) error {
		fmt.Printf("Processing operation: %s\n", op.Operation)
		return nil
	})
}

func eventsExample(dbl *dblayer.DBLayer) error {
	// Регистрируем обработчики событий
	// Register event handlers
	qb := dbl.Query("users")

	qb.On(QB.BeforeCreate, func(data interface{}) error {
		fmt.Println("Before creating user")
		return nil
	})
	qb.On(QB.AfterCreate, func(data interface{}) error {
		fmt.Println("After creating user")
		return nil
	})
	// Создаем пользователя (сработают события)
	// Create a user (events will trigger)
	_, err := qb.Create(map[string]interface{}{
		"name":  "New User",
		"email": "new@example.com",
	})
	return err
}

func contextExample(dbl *dblayer.DBLayer) (int64, error) {

	qb := dbl.Query("users")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	count, err := qb.Context(ctx).Count()
	if err != nil {
		return 0, err
	}

	return count, nil
}

func geoExample(dbl *dblayer.DBLayer) error {
	// Поиск мест в радиусе 5 км от точки
	// Search for places within 5km radius from point
	var places []struct {
		ID   int64   `db:"id"`
		Name string  `db:"name"`
		Lat  float64 `db:"lat"`
		Lng  float64 `db:"lng"`
	}
	point := QB.Point{
		Lat: 55.7558,
		Lng: 37.6173,
	}
	_, err := dbl.Query("places").
		GeoSearch("location", point, 5000). // радиус в метрах / radius in meters
		OrderBy("distance", "ASC").
		Limit(10).
		Get(&places)
	return err
}

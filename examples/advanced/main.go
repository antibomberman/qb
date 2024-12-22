package advanced

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/antibomberman/dblayer"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Подключение к базе данных
	db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/testdb?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	dbl := dblayer.NewDBLayer("mysql", db)
	// Пример использования очередей
	if err := queueExample(dbl); err != nil {
		log.Fatal(err)
	}
	// Пример использования событий
	if err := eventsExample(dbl); err != nil {
		log.Fatal(err)
	}
	// Пример геопространственных запросов
	if err := geoExample(dbl); err != nil {
		log.Fatal(err)
	}
}
func queueExample(dbl *dblayer.DBLayer) error {
	// Создаем отложенную операцию
	err := dbl.Table("queued_operations").Queue(
		"send_email",
		map[string]interface{}{
			"to":      "user@example.com",
			"subject": "Привет!",
			"body":    "Это отложенное сообщение",
		},
		time.Now().Add(1*time.Hour),
	)
	if err != nil {
		return err
	}
	// Обработка очереди
	return dbl.Table("queued_operations").ProcessQueue(func(op dblayer.QueuedOperation) error {
		fmt.Printf("Обработка операции: %s\n", op.Operation)
		return nil
	})
}
func eventsExample(dbl *dblayer.DBLayer) error {
	// Регистрируем обработчики событий
	qb := dbl.Table("users")

	qb.On(dblayer.BeforeCreate, func(data interface{}) error {
		fmt.Println("Перед созданием пользователя")
		return nil
	})
	qb.On(dblayer.AfterCreate, func(data interface{}) error {
		fmt.Println("После создания пользователя")
		return nil
	})
	// Создаем пользователя (сработают события)
	_, err := qb.Create(map[string]interface{}{
		"name":  "Новый пользователь",
		"email": "new@example.com",
	})
	return err
}
func geoExample(dbl *dblayer.DBLayer) error {
	// Поиск мест в радиусе 5 км от точки
	var places []struct {
		ID   int64   `db:"id"`
		Name string  `db:"name"`
		Lat  float64 `db:"lat"`
		Lng  float64 `db:"lng"`
	}
	point := dblayer.Point{
		Lat: 55.7558,
		Lng: 37.6173,
	}
	_, err := dbl.Table("places").
		GeoSearch("location", point, 5000). // радиус в метрах
		OrderBy("distance", "ASC").
		Limit(10).
		Get(&places)
	return err

}

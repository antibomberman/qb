# DBLayer
### DBLayer - это Go пакет, предоставляющий удобный интерфейс для работы с реляционными базами данных. Он обеспечивает абстракцию над database/sql и sqlx, упрощая выполнение общих операций с базой данных.

## Установка
### go get github.com/antibomberman/dblayer@v0.0.6


## Пример

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/yourusername/dblayer"
)

type User struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
	Age  int    `db:"age"`
}

func main() {
	// Подключение к базе данных
	db, err := sqlx.Connect("postgres", "user=postgres dbname=testdb sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Создание экземпляра DBLayer
	dbLayer := dblayer.NewDBLayer(db)

	// Создание нового пользователя
	user := User{Name: "John Doe", Age: 30}
	ctx := context.Background()
	id, err := dbLayer.Create(ctx, "users", user)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	fmt.Printf("Created user with ID: %d\n", id)

	// Получение пользователя
	var retrievedUser User
	err = dbLayer.Get(ctx, "users", []dblayer.Condition{{Column: "id", Operator: "=", Value: id}}, &retrievedUser)
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
	}
	fmt.Printf("Retrieved user: %+v\n", retrievedUser)

	// Обновление пользователя
	updates := map[string]interface{}{"age": 31}
	affected, err := dbLayer.UpdateRecord(ctx, "users", updates, []dblayer.Condition{{Column: "id", Operator: "=", Value: id}})
	if err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}
	fmt.Printf("Updated %d user(s)\n", affected)

	// Получение списка пользователей
	var users []User
	err = dbLayer.List(ctx, "users", nil, "name ASC", 10, 0, &users)
	if err != nil {
		log.Fatalf("Failed to list users: %v", err)
	}
	fmt.Printf("Users: %+v\n", users)

	// Удаление пользователя
	affected, err = dbLayer.Delete(ctx, "users", []dblayer.Condition{{Column: "id", Operator: "=", Value: id}})
	if err != nil {
		log.Fatalf("Failed to delete user: %v", err)
	}
	fmt.Printf("Deleted %d user(s)\n", affected)
}
```


### Основные операции
Exists: Проверяет существование записи в таблице.
```
Exists(ctx context.Context, tableName string, conditions []Condition) (bool, error)
```
Create: Создает новую запись в таблице.
```
Create(ctx context.Context, tableName string, record interface{}) (int64, error)
```
Get: Получает запись из таблицы.
```
Get(ctx context.Context, tableName string, conditions []Condition, result interface{}) error
```
Update: Обновляет запись в таблице.
```
Update(ctx context.Context, tableName string, updates map[string]interface{}, conditions []Condition) (int64, error)
```
Delete: Удаляет запись из таблицы.
```
Delete(ctx context.Context, tableName string, conditions []Condition) (int64, error)
```
List: Получает список записей из таблицы.
```
List(ctx context.Context, tableName string, conditions []Condition, orderBy string, limit, offset int, result interface{}) error
```
### Агрегатные функции

Count: Подсчитывает количество записей.
```
Count(ctx context.Context, tableName string, conditions []Condition) (int64, error)
```
Avg: Вычисляет среднее значение.
```
Avg(ctx context.Context, tableName, column string, conditions []Condition) (float64, error)
```
Min: Находит минимальное значение.
```
Min(ctx context.Context, tableName, column string, conditions []Condition) (interface{}, error)
```
Max: Находит максимальное значение.
```
Max(ctx context.Context, tableName, column string, conditions []Condition) (interface{}, error)
```
Sum: Вычисляет сумму.
```
Sum(ctx context.Context, tableName, column string, conditions []Condition) (float64, error)
```
### Дополнительные операции
InTransaction: Выполняет операции в транзакции.
```
InTransaction(ctx context.Context, fn func(context.Context, *sqlx.Tx) error) error
```
BatchInsert: Выполняет пакетную вставку записей.
```
BatchInsert(ctx context.Context, tableName string, records []interface{}) error
```
ExecuteRawQuery: Выполняет произвольный SQL-запрос.
```
ExecuteRawQuery(ctx context.Context, query string, args []interface{}, result interface{}) error
```

ExecuteWithRetry: Выполняет операцию с автоматическими повторными попытками в случае ошибки
```
ExecuteWithRetry(ctx context.Context, maxAttempts int, operation func(context.Context) error) error

```
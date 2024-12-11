# DBLayer
### DBLayer - это Go пакет, предоставляющий удобный интерфейс для работы с реляционными базами данных. Он обеспечивает абстракцию над database/sql и sqlx, упрощая выполнение общих операций с базой данных.

## Установка
### go get github.com/antibomberman/dblayer@v0.0.12


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
		log.Fatalf("Failed to operations user: %v", err)
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
```go
Exists(ctx context.Context, tableName string, conditions []Condition) (bool, error)
```
Create: Создает новую запись в таблице.
```go
Create(ctx context.Context, tableName string, record interface{}) (int64, error)
CreateRecord(ctx context.Context, tableName string, record map[string]interface{}) (int64, error)
GetOrCreate(ctx context.Context, tableName string, conditions []Condition, defaultValues map[string]interface{}, result interface{}) (bool, error) 
```
Get: Получает запись из таблицы.
```go
Get(ctx context.Context, tableName string, conditions []Condition, result interface{})  (bool,error)
First(ctx context.Context, tableName string, conditions []Condition, result interface{}) (bool,error)
Last(ctx context.Context, tableName string, conditions []Condition, result interface{}) (bool,error)
```
Update: Обновляет запись в таблице.
```go
Update(ctx context.Context, tableName string, conditions []Condition, updates map[string]interface{}) (int64, error)
UpdateRecord(ctx context.Context, tableName string,conditions []Condition, updates map[string]interface{}) (int64, error)
BulkUpdate(ctx context.Context, tableName string, updates []interface{}, idColumn string)
BulkUpdateRecord(ctx context.Context, tableName string, updates []map[string]interface{}, idColumn string) error
Upsert(ctx context.Context, tableName string, record interface{}, uniqueColumns []string) error
```
Delete: Удаляет запись из таблицы.
```go
Delete(ctx context.Context, tableName string, conditions []Condition) (int64, error)
```
List: Получает список записей из таблицы.
```go
All(ctx context.Context, tableName string, conditions []Condition, orderBy string, result interface{}) error
List(ctx context.Context, tableName string, conditions []Condition, orderBy string, limit, offset int, result interface{}) error
PaginateWithCursor(ctx context.Context, tableName string, cursorColumn string, cursorValue interface{}, pageSize int, conditions []Condition, result interface{}) error
SelectFields(ctx context.Context, tableName string, fields []string, conditions []Condition, result interface{}) error
SearchLike(ctx context.Context, tableName string, searchColumn string, searchTerm string, additionalConditions []Condition, result interface{}) error
GroupBy(ctx context.Context, tableName string, groupColumns []string, aggregations map[string]string, conditions []Condition) ([]map[string]interface{}, error)
WithinRadius(ctx context.Context, tableName string, latColumn, lonColumn string, lat, lon float64, radiusKm float64, result interface{}) error
```
### Агрегатные функции

```go
Count(ctx context.Context, tableName string, conditions []Condition) (int64, error)
Avg(ctx context.Context, tableName, column string, conditions []Condition) (float64, error)
Min(ctx context.Context, tableName, column string, conditions []Condition) (interface{}, error)
Max(ctx context.Context, tableName, column string, conditions []Condition) (interface{}, error)
Sum(ctx context.Context, tableName, column string, conditions []Condition) (float64, error)
```
### Дополнительные операции
InTransaction: Выполняет операции в транзакции.
```go
InTransaction(ctx context.Context, fn func(context.Context, *sqlx.Tx) error) error
```
BatchInsert: Выполняет пакетную вставку записей.
```go
BatchInsert(ctx context.Context, tableName string, records []interface{}) error
BatchInsertRecords(ctx context.Context, tableName string, records []map[string]interface{}) error
```
ExecuteRawQuery: Выполняет произвольный SQL-запрос.
```go
ExecuteRawQuery(ctx context.Context, query string, args []interface{}, result interface{}) error
```

ExecuteWithRetry: Выполняет операцию с автоматическими повторными попытками в случае ошибки
```go
ExecuteWithRetry(ctx context.Context, maxAttempts int, operation func(context.Context) error) error
```
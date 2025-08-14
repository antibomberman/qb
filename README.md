# Go Query Builder (QB)

**QueryBuilder** — это мощный и гибкий SQL-билдер для Go, предоставляющий удобный интерфейс для работы с базами данных MySQL

## 🚀 Основные возможности

- Построение SQL-запросов через **цепочку методов** (fluent interface).
- Полная поддержка **транзакций**.
- **Кеширование** результатов запросов (Memory | Redis).
- Пакетные операции `INSERT` и `UPDATE`.
- Поддержка **"сырых"** SQL-запросов.
- Различные типы `JOIN`-соединений.
- Условия `WHERE` с **вложенными группами**.
- Поддержка **подзапросов**.

## 📦 Установка

```bash
go get github.com/antibomberman/qb@v1.2.19
```

## 💡 Пример использования

Ниже приведен простой пример получения данных о продукте из базы данных.

```go
package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/antibomberman/qb"
	_ "github.com/go-sql-driver/mysql" // или другой драйвер
)

// Product определяет структуру для таблицы 'products'
type Product struct {
	ID        int        `db:"id"`
	Name      string     `db:"name"`
	Active    bool       `db:"active"`
	CreatedAt time.Time  `db:"created_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func main() {
	// 1. Инициализация подключения к БД
	db, err := sql.Open("mysql", "root:rootpassword@tcp(localhost:3316)/test_db")
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	// 2. Создание экземпляра Query Builder
	queryBuilder := qb.New("mysql", db)

	// Пример 1: Простое построение SQL-запроса и его аргументов
	sqlQuery, args, err := queryBuilder.From("products").Where("id = ?", 1).ToSql()
	if err != nil {
		log.Fatalf("Ошибка построения запроса: %v", err)
	}
	log.Printf("Сгенерированный SQL: %s\nАргументы: %v\n", sqlQuery, args)

	// Пример 2: Выполнение запроса и получение одной записи
	var product Product
	found, err := queryBuilder.
		From("products").
		Where("active = ?", true).
		WhereNull("deleted_at").
		OrderBy("created_at", "DESC").
		Get(&product) // Используем Get для получения одной записи

	if err != nil {
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	if !found {
		log.Println("Продукт не найден")
		return
	}
	log.Printf("Найденный продукт: %+v", product)
}
```

## 📚 Документация

Более подробную информацию и примеры использования вы найдете в нашей [документации](docs/ru.md).

## 🤝 Как внести вклад

Мы рады любому вкладу! Перед созданием коммита, пожалуйста, убедитесь, что ваш код отформатирован и все тесты проходят успешно:

```bash
gofmt -s -w . && go test -v ./...
```

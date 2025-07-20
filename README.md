# QueryBuilder
### QueryBuilder - это мощный и гибкий SQL-билдер для Go, предоставляющий удобный интерфейс для работы с базами данных MySQL, PostgreSQL и SQLite.

## Установка
### go get github.com/antibomberman/qb@v1.2.12

## Основные возможности

- Построение SQL-запросов через цепочку методов
- Поддержка транзакций
- Кеширование результатов запросов Memory | Redis (теперь с явной инициализацией)
- Пакетные операции вставки и обновления
- Работа с датой и временем
- Поддержка сырых SQL-запросов
- Различные типы JOIN-соединений
- Условия WHERE с вложенными группами
- Поддержка подзапросов
- Оконные функции
- Мягкое удаление записей

```go
    db, err := sql.Open("mysql", "root:rootpassword@tcp(localhost:3316)/test_db")
    if err != nil {
        panic(err)
    }
    queryBuilder := qb.New("mysql", db)
    
    sqlQuery, args := queryBuilder.From("products").Where("id = ?", 1).ToSql()
    fmt.Printf("Generated SQL: %s\n", sqlQuery)
    fmt.Printf("Generated Args: %v\n", args)

    products := Product{}
    found, err := queryBuilder.From("products").
		Where("active = ?", true).
		WhereNull("deleted_at").
		OrderBy("created_at", "DESC").
		Get(&products)
	
```
	
    if err != nil {
        log.Fatal(err)
    }
    if !found {
        log.Println("No products found")
        return
    }
    log.Printf("Found product: %+v", products)
```


## Документация

Подробная документация [документации по запросам](https://github.com/antibomberman/qb/blob/main/docs/query_ru.md).


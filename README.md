# QueryBuilder
### QueryBuilder - это мощный и гибкий SQL-билдер для Go, предоставляющий удобный интерфейс для работы с базами данных MySQL, PostgreSQL и SQLite.

## Установка
### go get github.com/antibomberman/QueryBuilder@v1.1.91

## Основные возможности

- Построение SQL-запросов через цепочку методов
- Поддержка транзакций
- Кеширование результатов запросов Memory | Redis 
- Пакетные операции вставки и обновления
- Работа с датой и временем
- Поддержка сырых SQL-запросов
- Различные типы JOIN-соединений
- Условия WHERE с вложенными группами
- Поддержка подзапросов
- Оконные функции
- Мягкое удаление записей

```go
    db, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/test")
    if err != nil {
        panic(err)
    }
    queryBuilder := qb.New("mysql", db)
    
    products := Product{}
    found, err := queryBuilder.From("products").
		Where("active = ?", true).
        WhereNull("orders.deleted_at").
		OrderBy("created_at", "DESC").
		Get(&products)
	
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


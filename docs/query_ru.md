# Документация по пакету qb

## Обзор
Пакет `qb` представляет собой мощный и гибкий построитель SQL-запросов для Go. Он предоставляет удобный интерфейс для работы с базами данных, поддерживая различные СУБД.


#### Основные методы:
- `New(driverName string, db *sql.DB)` - создает новый экземпляр QueryBuilder
- `NewX(driverName string, db *sqlx.DB)` - создает новый экземпляр с расширенной функциональностью
- `From(table string)` - указывает таблицу для запроса
- `SetLogger(logger *slog.Logger)` - устанавливает логгер
- `SetCache(cache CacheInterface)` - устанавливает систему кеширования

### Builder
Структура для построения SQL-запросов.

#### Основные методы выборки данных:
- `Find(id any, dest any)` - поиск записи по ID
- `First(dest any)` - получение первой записи
- `Get(dest any)` - получение всех записей
- `Pluck(column string, dest any)` - получение значений одной колонки
- `Value(column string)` - получение значения одного поля

#### Методы условий:
- `Where(condition string, args ...any)` - добавление условия AND
- `OrWhere(condition string, args ...any)` - добавление условия OR
- `WhereIn(column string, values ...any)` - условие IN
- `WhereNull(column string)` - проверка на NULL
- `WhereNotNull(column string)` - проверка на NOT NULL
- `WhereBetween(column string, start, end any)` - условие BETWEEN

#### Методы модификации данных:
- `Create(data any, fields ...string)` - создание новой записи
- `CreateMap(data map[string]any)` - создание записи из map
- `Update(data any, fields ...string)` - обновление записей
- `UpdateMap(data map[string]any)` - обновление записей используя map
- `Delete()` - удаление записей

#### Методы агрегации:
- `Count()` - подсчет количества записей
- `Avg(column string)` - среднее значение
- `Sum(column string)` - сумма значений
- `Min(column string)` - минимальное значение
- `Max(column string)` - максимальное значение

#### Методы группировки и сортировки:
- `GroupBy(columns ...string)` - группировка
- `Having(condition string)` - условие для группировки
- `OrderBy(column string, direction string)` - сортировка
- `Limit(limit int)` - ограничение количества записей
- `Offset(offset int)` - смещение

#### Методы объединения таблиц:
- `Join(table string, condition string)` - INNER JOIN
- `LeftJoin(table string, condition string)` - LEFT JOIN
- `RightJoin(table string, condition string)` - RIGHT JOIN
- `CrossJoin(table string)` - CROSS JOIN

#### Методы пагинации:
- `Paginate(page int, perPage int, dest any)` - стандартная пагинация
- `PaginateWithToken(token string, limit int, dest any)` - пагинация с токеном
- `PaginateWithCursor(cursor string, limit int, dest any)` - курсор-пагинация

#### Методы работы с датами:
- `WhereDate(column string, operator string, value time.Time)` - условие по дате
- `WhereDateTime(column string, operator string, value time.Time)` - условие по дате и времени
- `WhereYear(column string, operator string, year int)` - условие по году
- `WhereMonth(column string, operator string, month int)` - условие по месяцу
- `WhereDay(column string, operator string, day int)` - условие по дню

#### Асинхронные методы:
Большинство методов имеют асинхронные версии с суффиксом `Async`, например:
- `FindAsync(id any, dest any)`
- `FirstAsync(dest any)`
- `GetAsync(dest any)`
- `CreateAsync(data any, fields ...string)`
- `UpdateAsync(data any, fields ...string)`

#### Дополнительные возможности:
- `WithAudit(userID any)` - включение аудита изменений
- `WithMetrics(collector *MetricsCollector)` - сбор метрик
- `Remember(key string, duration time.Duration)` - кеширование
- `Queue(operation string, data any, runAt time.Time)` - отложенные операции

## Примеры использования

### Базовые операции
```go
// Создание построителя запросов
qb := New("postgres", db)

// Получение записи по ID
var user User
found, err := qb.From("users").Find(1, &user)

// Получение всех активных пользователей
var users []User
found, err := qb.From("users").
    Where("active = ?", true).
    OrderBy("created_at", "DESC").
    Get(&users)

// Создание новой записи
id, err := qb.From("users").Create(&User{
    Name: "John",
    Email: "john@example.com",
})

// Обновление записи
err := qb.From("users").
    Where("id = ?", 1).
    Update(&User{Name: "Jane"})

// Удаление записи
err := qb.From("users").
    Where("id = ?", 1).
    Delete()
```

### Сложные запросы
```go
// Объединение таблиц с условиями
var results []Result
found, err := qb.From("orders").
    LeftJoin("users", "users.id = orders.user_id").
    Where("orders.status = ?", "pending").
    WhereNull("orders.deleted_at").
    GroupBy("users.id").
    Having("COUNT(*) > 1").
    Get(&results)

// Подзапросы
subQuery := qb.From("orders").
    Select("user_id").
    Where("status = ?", "completed")

users, err := qb.From("users").
    WhereIn("id", subQuery).
    Get(&users)
```

### Пагинация
```go
// Стандартная пагинация
result, err := qb.From("users").
    Where("active = ?", true).
    Paginate(1, 10, &users)

// Пагинация с курсором
result, err := qb.From("users").
    OrderBy("id", "ASC").
    PaginateWithCursor("", 10, &users)
```

## Примечания
- Все методы безопасны для использования в горутинах
- Поддерживаются PostgreSQL и MySQL
- Рекомендуется использовать подготовленные выражения для предотвращения SQL-инъекций

## Кеширование
```go
// Использование кеша в памяти
qb.SetCache(NewCacheMemory())

// Кеширование запроса
var users []User
found, err := qb.From("users").
    Remember("users_list", time.Minute).
    Get(&users)
```
### MemoryCache
Реализация кеша в памяти.

```go
cache := NewCacheMemory()
```

Основные характеристики:
- Хранение данных в памяти процесса
- Автоматическая очистка просроченных записей
- Потокобезопасность через sync.RWMutex
- Поддержка TTL для записей

### RedisCache
Реализация кеша с использованием Redis.

```go
cache := NewCacheRedis("localhost:6379", "", 0)
```

Особенности:
- Персистентное хранение данных
- Распределенный кеш
- Автоматическое удаление по TTL
- JSON сериализация значений

## События

### EventType
Типы поддерживаемых событий:
- `BeforeCreate` - перед созданием записи
- `AfterCreate` - после создания записи
- `BeforeUpdate` - перед обновлением записи
- `AfterUpdate` - после обновления записи
- `BeforeDelete` - перед удалением записи
- `AfterDelete` - после удаления записи

### Использование событий
```go
// Регистрация обработчика
qb.On(BeforeCreate, func(data any) error {
    // Обработка события
    return nil
})

// Создание записи (автоматически вызовет обработчики)
qb.From("users").Create(user)
```

## Транзакции

### Transaction
Поддержка транзакций с автоматическим откатом при ошибке.

```go
// Простая транзакция
tx, err := qb.Begin()
if err != nil {
    return err
}
defer tx.Rollback()

// Выполнение операций
err = tx.From("users").Create(user)
if err != nil {
    return err
}

return tx.Commit()

// Транзакция с помощью замыкания
err := qb.Transaction(func(tx *Transaction) error {
    // Выполнение операций
    return tx.From("users").Create(user)
})
```

### Особенности транзакций:
- Поддержка вложенных транзакций
- Автоматический откат при панике
- Контекстная поддержка через `BeginContext` и `TransactionContext`
- Безопасность для горутин


## Примеры использования дополнительных возможностей





### Транзакции с контекстом
```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
defer cancel()

err := qb.TransactionContext(ctx, func(tx *Transaction) error {
    // Выполнение операций с таймаутом
    return nil
})
```

## Рекомендации по использованию
1. Используйте транзакции для атомарных операций
2. Применяйте кеширование для часто запрашиваемых данных
3. Регистрируйте обработчики событий для сквозной функциональности
4. Используйте контексты для управления таймаутами и отменой операций
5. Собирайте метрики для мониторинга производительности
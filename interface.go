package qb

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

// Executor определяет интерфейс для выполнения низкоуровневых операций с базой данных.
// Он расширяет sqlx.Ext для именованных запросов и предоставляет методы для получения одной записи,
// нескольких записей и выполнения команд с контекстом или без него.
type Executor interface {
	sqlx.Ext
	// Get извлекает одну запись из базы данных и сканирует ее в dest.
	// Пример:
	// var user User
	// err := db.Get(&user, "SELECT * FROM users WHERE id = ?", 1)
	Get(dest any, query string, args ...any) error
	// Select извлекает несколько записей из базы данных и сканирует их в dest (срез).
	// Пример:
	// var users []User
	// err := db.Select(&users, "SELECT * FROM users WHERE age > ?", 25)
	Select(dest any, query string, args ...any) error
	// NamedExecContext выполняет именованный запрос с контекстом, обычно для INSERT, UPDATE, DELETE.
	// Пример:
	// result, err := db.NamedExecContext(ctx, "INSERT INTO users (name, email) VALUES (:name, :email)", userStruct)
	NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error)
	// ExecContext выполняет запрос с контекстом без возврата строк, обычно для DDL или DML.
	// Пример:
	// result, err := db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", 1)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	// SelectContext извлекает несколько записей с контекстом и сканирует их в dest (срез).
	// Пример:
	// var users []User
	// err := db.SelectContext(ctx, &users, "SELECT * FROM users WHERE status = ?", "active")
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	// GetContext извлекает одну запись с контекстом и сканирует ее в dest.
	// Пример:
	// var user User
	// err := db.GetContext(ctx, "SELECT * FROM users WHERE email = ?", "test@example.com")
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	// QueryRowxContext выполняет запрос с контекстом и возвращает одну строку для сканирования.
	// Пример:
	// row := db.QueryRowxContext(ctx, "SELECT name FROM users WHERE id = ?", 1)
	// var name string
	// err := row.Scan(&name)
	QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row
}

// DBInterface определяет интерфейс для взаимодействия с основным подключением к базе данных.
// Он включает методы для выполнения запросов, получения результатов и управления транзакциями.
type DBInterface interface {
	sqlx.Ext
	// Get извлекает одну запись из базы данных и сканирует ее в dest.
	// Пример:
	// var user User
	// err := db.Get(&user, "SELECT * FROM users WHERE id = ?", 1)
	Get(dest any, query string, args ...any) error
	// GetContext извлекает одну запись с контекстом и сканирует ее в dest.
	// Пример:
	// var user User
	// err := db.GetContext(ctx, "SELECT * FROM users WHERE email = ?", "test@example.com")
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	// Select извлекает несколько записей из базы данных и сканирует их в dest (срез).
	// Пример:
	// var users []User
	// err := db.Select(&users, "SELECT * FROM users WHERE age > ?", 25)
	Select(dest any, query string, args ...any) error
	// SelectContext извлекает несколько записей с контекстом и сканирует их в dest (срез).
	// Пример:
	// var users []User
	// err := db.SelectContext(ctx, &users, "SELECT * FROM users WHERE status = ?", "active")
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	// Exec выполняет запрос без возврата строк, обычно для DDL или DML.
	// Пример:
	// result, err := db.Exec("DELETE FROM users WHERE id = ?", 1)
	Exec(query string, args ...any) (sql.Result, error)
	// ExecContext выполняет запрос с контекстом без возврата строк, обычно для DDL или DML.
	// Пример:
	// result, err := db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", 1)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	// NamedExecContext выполняет именованный запрос с контекстом, обычно для INSERT, UPDATE, DELETE.
	// Пример:
	// result, err := db.NamedExecContext(ctx, "INSERT INTO users (name, email) VALUES (:name, :email)", userStruct)
	NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error)
	// QueryRowxContext выполняет запрос с контекстом и возвращает одну строку для сканирования.
	// Пример:
	// row := db.QueryRowxContext(ctx, "SELECT name FROM users WHERE id = ?", 1)
	// var name string
	// err := row.Scan(&name)
	QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row
	// Beginx начинает новую транзакцию.
	// Пример:
	// tx, err := db.Beginx()
	Beginx() (*sqlx.Tx, error)
	// BeginTxx начинает новую транзакцию с контекстом и опциями.
	// Пример:
	// tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
}

// TxInterface определяет интерфейс для взаимодействия с транзакцией базы данных.
// Он включает методы для выполнения запросов в рамках транзакции и фиксации или отката.
type TxInterface interface {
	sqlx.Ext
	// Get извлекает одну запись из базы данных в рамках транзакции и сканирует ее в dest.
	// Пример:
	// var user User
	// err := tx.Get(&user, "SELECT * FROM users WHERE id = ?", 1)
	Get(dest any, query string, args ...any) error
	// GetContext извлекает одну запись с контекстом в рамках транзакции и сканирует ее в dest.
	// Пример:
	// var user User
	// err := tx.GetContext(ctx, "SELECT * FROM users WHERE email = ?", "test@example.com")
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	// Select извлекает несколько записей из базы данных в рамках транзакции и сканирует их в dest (срез).
	// Пример:
	// var users []User
	// err := tx.Select(&users, "SELECT * FROM users WHERE age > ?", 25)
	Select(dest any, query string, args ...any) error
	// SelectContext извлекает несколько записей с контекстом в рамках транзакции и сканирует их в dest (срез).
	// Пример:
	// var users []User
	// err := tx.SelectContext(ctx, &users, "SELECT * FROM users WHERE status = ?", "active")
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	// Exec выполняет запрос в рамках транзакции без возврата строк.
	// Пример:
	// result, err := tx.Exec("DELETE FROM users WHERE id = ?", 1)
	Exec(query string, args ...any) (sql.Result, error)
	// ExecContext выполняет запрос с контекстом в рамках транзакции без возврата строк.
	// Пример:
	// result, err := tx.ExecContext(ctx, "DELETE FROM users WHERE id = ?", 1)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	// NamedExecContext выполняет именованный запрос с контекстом в рамках транзакции.
	// Пример:
	// result, err := tx.NamedExecContext(ctx, "INSERT INTO users (name, email) VALUES (:name, :email)", userStruct)
	NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error)
	// Commit фиксирует транзакцию.
	// Пример:
	// err := tx.Commit()
	Commit() error
	// Rollback откатывает транзакцию.
	// Пример:
	// err := tx.Rollback()
	Rollback() error
}

// QueryBuilderInterface определяет основной интерфейс для построителя запросов.
// Он предоставляет методы для начала построения запросов, управления логированием, кешированием и транзакциями.
type QueryBuilderInterface interface {
	// From устанавливает имя таблицы для запроса и возвращает BuilderInterface.
	// Пример:
	// qb.From("users")
	From(table string) BuilderInterface
	// Raw выполняет сырой SQL-запрос.
	// Пример:
	// qb.Raw("SELECT * FROM users WHERE id = ?", 1).QueryRow(&user)
	Raw(query string, args ...any) *RawQuery
	// GetDB возвращает базовый DBInterface.
	// Пример:
	// db := qb.GetDB()
	GetDB() DBInterface
	// SetLogger устанавливает пользовательский slog.Logger для логирования операций запросов.
	// Пример:
	// qb.SetLogger(slog.Default())
	SetLogger(logger *slog.Logger)
	// SetMemoryCache настраивает построитель запросов на использование кеша в памяти.
	// Пример:
	// qb.SetMemoryCache()
	SetMemoryCache() *MemoryCache
	// SetRedisCache настраивает построитель запросов на использование кеша Redis.
	// Пример:
	// qb.SetRedisCache("localhost:6379", "", 0)
	SetRedisCache(addr string, password string, db int) (*RedisCache, error)
	// Begin начинает новую транзакцию базы данных.
	// Пример:
	// tx, err := qb.Begin()
	Begin() (*Transaction, error)
	// BeginContext начинает новую транзакцию базы данных с контекстом.
	// Пример:
	// tx, err := qb.BeginContext(ctx)
	BeginContext(ctx context.Context) (*Transaction, error)
	// Transaction выполняет функцию в рамках транзакции, автоматически фиксируя или откатывая ее.
	// Пример:
	// err := qb.Transaction(func(tx *Transaction) error {
	//     // Выполнить операции с использованием tx
	//     return nil
	// })
	Transaction(fn func(*Transaction) error) error
	// TransactionContext выполняет функцию в рамках транзакции с контекстом.
	// Пример:
	// err := qb.TransactionContext(ctx, func(tx *Transaction) error {
	//     // Выполнить операции с использованием tx
	//     return nil
	// })
	TransactionContext(ctx context.Context, fn func(*Transaction) error) error

	// Debug логирует отладочное сообщение для операции запроса.
	// Пример:
	// qb.Debug("Запрос выполнен", time.Now(), "SELECT * FROM users", nil)
	Debug(msg string, start time.Time, query string, args ...any)
	// Info логирует информационное сообщение для операции запроса.
	// Пример:
	// qb.Info("Запрос успешно выполнен", time.Now(), "INSERT INTO products", productData)
	Info(msg string, start time.Time, query string, args ...any)
	// Warn логирует предупреждающее сообщение для операции запроса.
	// Пример:
	// qb.Warn("Обнаружен медленный запрос", time.Now(), "SELECT * FROM large_table", nil)
	Warn(msg string, start time.Time, query string, args ...any)
	// Error логирует сообщение об ошибке для операции запроса.
	// Пример:
	// qb.Error("Ошибка базы данных", time.Now(), "UPDATE orders", err)
	Error(msg string, start time.Time, query string, args ...any)
}

// BuilderInterface определяет всеобъемлющий интерфейс для построения и выполнения SQL-запросов.
// Он включает различные специализированные интерфейсы для CRUD, WHERE, Date, Join и других операций.
type BuilderInterface interface {
	CrudInterface
	WhereInterface
	DateInterface
	WindowInterface
	LockInterface
	AggregateInterface

	JoinInterface
	PaginationInterface
	GroupSortInterface
	SubQueryInterface
	QueryOptionsInterface
	QueueInterface
	EventInterface
	SpecialQueriesInterface

	// Select указывает столбцы для выборки в запросе.
	// Пример:
	// qb.Select("id", "name", "email")
	Select(columns ...string) *Builder
	// Limit устанавливает максимальное количество возвращаемых строк.
	// Пример:
	// qb.Limit(10)
	Limit(limit int) *Builder
	// Offset устанавливает смещение для запроса, пропуская указанное количество строк.
	// Пример:
	// qb.Offset(20)
	Offset(offset int) *Builder

	// SubQuery создает подзапрос из текущего состояния построителя.
	// Пример:
	// sub := qb.From("orders").Select("user_id").Where("amount > ?", 100)
	// qb.From("users").WhereIn("id", sub.SubQuery("sub_users"))
	SubQuery(alias string) *Builder
	// WhereSubQuery добавляет условие WHERE на основе подзапроса.
	// Пример:
	// qb.WhereSubQuery("id", "IN", qb.From("orders").Select("user_id").Where("amount > ?", 100))
	WhereSubQuery(column string, operator string, subQuery *Builder) *Builder
	// Union объединяет текущий запрос с другим с использованием UNION.
	// Пример:
	// qb1 := qb.From("users").Select("name")
	// qb2 := qb.From("customers").Select("name")
	// combined := qb1.Union(qb2)
	Union(other *Builder) *Builder
	// UnionAll объединяет текущий запрос с другим с использованием UNION ALL.
	// Пример:
	// qb1 := qb.From("users").Select("name")
	// qb2 := qb.From("customers").Select("name")
	// combined := qb1.UnionAll(qb2)
	UnionAll(other *Builder) *Builder

	// Pluck извлекает значения одного столбца в целевой срез.
	// Пример:
	// var names []string
	// qb.From("users").Pluck("name", &names)
	Pluck(column string, dest any) error
	// Value извлекает значение одного столбца из первой строки.
	// Пример:
	// name, err := qb.From("users").WhereId(1).Value("name")
	Value(column string) (any, error)
	// Values извлекает значения одного столбца из всех строк в срез.
	// Пример:
	// var ages []int
	// qb.From("users").Values("age", &ages)
	Values(column string) ([]any, error)
	// Chunk обрабатывает результаты запроса частями, вызывая функцию для каждой части.
	// Пример:
	// qb.From("large_table").Chunk(100, func(items any) error {
	//     // Обработать 100 элементов
	//     return nil
	// })
	Chunk(size int, fn func(items any) error) error
	// ChunkContext обрабатывает результаты запроса частями с контекстом.
	// Пример:
	// qb.From("large_table").ChunkContext(ctx, 100, func(ctx context.Context, items any) error {
	//     // Обработать 100 элементов с контекстом
	//     return nil
	// })
	ChunkContext(ctx context.Context, size int, fn func(context.Context, any) error) error
	// WithinGroup добавляет предложение WITHIN GROUP для агрегатных функций.
	// Пример:
	// qb.Select("name", "PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY score)").WithinGroup("score", "ORDER BY score")
	WithinGroup(column string, window string) *Builder
	// Distinct добавляет предложение DISTINCT к оператору SELECT.
	// Пример:
	// qb.Distinct("category_id")
	// qb.Distinct() // SELECT DISTINCT *
	Distinct(columns ...string) *Builder
	// WithTransaction устанавливает построитель для использования существующей транзакции.
	// Пример:
	// tx, _ := qb.Begin()
	// qb.From("users").WithTransaction(tx).UpdateMap(map[string]any{"status": "active"})
	WithTransaction(tx *Transaction) *Builder

	// GeoSearch выполняет геопространственный поиск на основе столбца, точки и радиуса.
	// Пример:
	// qb.From("locations").GeoSearch("coordinates", Point{Lat: 34.0522, Lng: -118.2437}, 1000) // 1000 метров
	GeoSearch(column string, point Point, radius float64) *Builder
	// Search выполняет полнотекстовый поиск по указанным столбцам.
	// Пример:
	// qb.From("articles").Search([]string{"title", "content"}, "Go programming")
	Search(columns []string, query string) *Builder

	// ToSql возвращает сгенерированную строку SQL-запроса и ее аргументы.
	// Пример:
	// sql, args := qb.From("users").Where("age > ?", 25).ToSql()
	ToSql() (string, []any)
	// WithAudit включает аудит для запроса, отслеживая изменения по идентификатору пользователя.
	// Пример:
	// qb.From("products").WithAudit(userID).UpdateMap(map[string]any{"price": 29.99})
	WithAudit(userID any) *Builder
	// ProcessQueue обрабатывает ожидающие операции из очереди.
	// Пример:
	// qb.From("queued_operations").ProcessQueue(func(op QueuedOperation) error {
	//     // Обработать операцию
	//     return nil
	// })
	ProcessQueue(handler func(QueuedOperation) error) error
	// Queue добавляет операцию в очередь для последующей обработки.
	// Пример:
	// qb.From("queued_operations").Queue("send_email", map[string]any{"to": "a@b.com"}, time.Now().Add(time.Hour))
	Queue(operation string, data any, runAt time.Time) error
	// WithMetrics включает сбор метрик производительности запросов.
	// Пример:
	// collector := NewMetricsCollector()
	// qb.From("users").WithMetrics(collector).Get(&users)
	WithMetrics(collector *MetricsCollector) *Builder
	// On регистрирует обработчик события для определенного типа события.
	// Пример:
	// qb.On(BeforeCreate, func(data any) error {
	//     fmt.Println("Перед созданием:", data)
	//     return nil
	// })
	On(event EventType, handler EventHandler)
	// Trigger отправляет событие с связанными данными.
	// Пример:
	// qb.Trigger(AfterUpdate, updatedUser)
	Trigger(event EventType, data any)
	// Context устанавливает контекст для построителя запросов.
	// Пример:
	// qb.Context(ctx).From("users").Get(&users)
	Context(ctx context.Context) *Builder
}

// CacheInterface определяет интерфейс для операций кеширования.
type CacheInterface interface {
	// Get извлекает данные из кеша по ключу и сканирует их в dest. Возвращает true, если найдено.
	// Пример:
	// var user User
	// if cache.Get("user:1", &user) {
	//     // Использовать кешированного пользователя
	// }
	Get(key string, dest any) bool
	// Set сохраняет данные в кеше с заданным ключом и сроком действия.
	// Пример:
	// cache.Set("user:1", user, time.Minute*5)
	Set(key string, value any, expiration time.Duration)
	// Delete удаляет элемент из кеша по ключу.
	// Пример:
	// cache.Delete("user:1")
	Delete(key string)
	// Clear удаляет все элементы из кеша.
	// Пример:
	// cache.Clear()
	Clear()
}

// CrudInterface определяет методы для операций создания, чтения, обновления и удаления.
type CrudInterface interface {
	// Find извлекает одну запись по ее ID.
	// Пример:
	// var user User
	// found, err := qb.From("users").Find(1, &user)
	Find(id any, dest any) (bool, error)
	// FindAsync извлекает одну запись по ID асинхронно.
	// Пример:
	// foundCh, errCh := qb.From("users").FindAsync(1, &user)
	// found := <-foundCh
	// err := <-errCh
	FindAsync(id any, dest any) (chan bool, chan error)
	// Get извлекает все записи, соответствующие текущим условиям запроса.
	// Пример:
	// var users []User
	// found, err := qb.From("users").Where("age > ?", 25).Get(&users)
	Get(dest any) (bool, error)
	// GetAsync извлекает все записи асинхронно.
	// Пример:
	// foundCh, errCh := qb.From("users").GetAsync(&users)
	// found := <-foundCh
	// err := <-errCh
	GetAsync(dest any) (chan bool, chan error)
	// Rows извлекает записи в виде среза map[string]any.
	// Пример:
	// rows, err := qb.From("users").Rows()
	Rows() ([]map[string]any, error)
	// First извлекает первую запись, соответствующую текущим условиям запроса.
	// Пример:
	// var user User
	// found, err := qb.From("users").OrderByAsc("id").First(&user)
	First(dest any) (bool, error)
	// FirstAsync извлекает первую запись асинхронно.
	// Пример:
	// foundCh, errCh := qb.From("users").FirstAsync(&user)
	// found := <-foundCh
	// err := <-errCh
	FirstAsync(dest any) (chan bool, chan error)
	// Create вставляет новую запись в таблицу.
	// Пример:
	// id, err := qb.From("users").Create(&User{Name: "John", Email: "john@example.com"})
	Create(data any, fields ...string) (any, error)
	// CreateAsync вставляет новую запись асинхронно.
	// Пример:
	// idCh, errCh := qb.From("users").CreateAsync(&User{Name: "Jane"})
	// id := <-idCh
	// err := <-errCh
	CreateAsync(data any, fields ...string) (chan any, chan error)
	// CreateMap вставляет новую запись из map[string]any.
	// Пример:
	// id, err := qb.From("users").CreateMap(map[string]any{"name": "Doe", "email": "doe@example.com"})
	CreateMap(data map[string]any) (any, error)
	// CreateMapAsync вставляет новую запись из map асинхронно.
	// Пример:
	// idCh, errCh := qb.From("users").CreateMapAsync(map[string]any{"name": "Alice"})
	// id := <-idCh
	// err := <-errCh
	CreateMapAsync(data map[string]any) (chan any, chan error)
	// Update обновляет записи, соответствующие текущим условиям запроса.
	// Пример:
	// err := qb.From("users").WhereId(1).Update(&User{Name: "Updated Name"})
	Update(data any, fields ...string) error
	// UpdateAsync обновляет записи асинхронно.
	// Пример:
	// errCh := qb.From("users").WhereId(1).UpdateAsync(&User{Name: "Async Update"})
	// err := <-errCh
	UpdateAsync(data any, fields ...string) chan error
	// UpdateMap обновляет записи из map[string]any.
	// Пример:
	// err := qb.From("users").WhereId(1).UpdateMap(map[string]any{"email": "new@example.com"})
	UpdateMap(data map[string]any) error
	// UpdateMapAsync обновляет записи из map асинхронно.
	// Пример:
	// errCh := qb.From("users").WhereId(1).UpdateMapAsync(map[string]any{"email": "async@example.com"})
	// err := <-errCh
	UpdateMapAsync(data map[string]any) chan error
	// Delete удаляет записи, соответствующие текущим условиям запроса.
	// Пример:
	// err := qb.From("users").Where("status = ?", "inactive").Delete()
	Delete() error
	// DeleteAsync удаляет записи асинхронно.
	// Пример:
	// errCh := qb.From("users").WhereId(1).DeleteAsync()
	// err := <-errCh
	DeleteAsync() chan error
	// Increment увеличивает значение числового столбца.
	// Пример:
	// err := qb.From("products").WhereId(1).Increment("quantity", 5)
	Increment(column string, value any) error
	// Decrement уменьшает значение числового столбца.
	// Пример:
	// err := qb.From("products").WhereId(1).Decrement("quantity", 2)
	Decrement(column string, value any) error
	// BatchInsert вставляет несколько записей одним пакетом.
	// Пример:
	// records := []map[string]any{{"name": "A"}, {"name": "B"}}
	// err := qb.From("items").BatchInsert(records)
	BatchInsert(records []map[string]any) error
	// BatchInsertAsync вставляет несколько записей асинхронно.
	// Пример:
	// errCh := qb.From("items").BatchInsertAsync(records)
	// err := <-errCh
	BatchInsertAsync(records []map[string]any) chan error
	// BulkInsert выполняет массовую вставку записей.
	// Пример:
	// records := []map[string]any{{"name": "C"}, {"name": "D"}}
	// err := qb.From("items").BulkInsert(records)
	BulkInsert(records []map[string]any) error
	// BulkInsertAsync выполняет массовую вставку асинхронно.
	// Пример:
	// errCh := qb.From("items").BulkInsertAsync(records)
	// err := <-errCh
	BulkInsertAsync(records []map[string]any) chan error
	// BulkUpdate выполняет массовое обновление записей на основе ключевого столбца.
	// Пример:
	// records := []map[string]any{{"id": 1, "name": "X"}, {"id": 2, "name": "Y"}}
	// err := qb.From("items").BulkUpdate(records, "id")
	BulkUpdate(records []map[string]any, keyColumn string) error
	// BulkUpdateAsync выполняет массовое обновление асинхронно.
	// Пример:
	// errCh := qb.From("items").BulkUpdateAsync(records, "id")
	// err := <-errCh
	BulkUpdateAsync(records []map[string]any, keyColumn string) chan error
	// BatchUpdate обновляет записи пакетами указанного размера.
	// Пример:
	// records := []map[string]any{{"id": 1, "name": "X"}, {"id": 2, "name": "Y"}}
	// err := qb.From("items").BatchUpdate(records, "id", 100)
	BatchUpdate(records []map[string]any, keyColumn string, batchSize int) error
}

// GroupSortInterface определяет методы для группировки и сортировки результатов запроса.
type GroupSortInterface interface {
	// OrderBy добавляет предложение ORDER BY к запросу.
	// Пример:
	// qb.OrderBy("created_at", "DESC")
	OrderBy(column string, direction string) *Builder
	// OrderByAsc добавляет предложение ORDER BY с возрастающим порядком.
	// Пример:
	// qb.OrderByAsc("name")
	OrderByAsc(column string) *Builder
	// OrderByDesc добавляет предложение ORDER BY с убывающим порядком.
	// Пример:
	// qb.OrderByDesc("name")
	OrderByDesc(column string) *Builder
	// GroupBy добавляет предложение GROUP BY к запросу.
	// Пример:
	// qb.GroupBy("category_id", "status")
	GroupBy(columns ...string) *Builder
	// Having добавляет предложение HAVING к запросу.
	// Пример:
	// qb.Having("COUNT(id) > 5")
	Having(condition string) *Builder
	// HavingRaw добавляет сырое предложение HAVING с аргументами.
	// Пример:
	// qb.HavingRaw("SUM(amount) > ?", 1000)
	HavingRaw(sql string, args ...any) *Builder
}

// WhereInterface определяет методы для добавления условий WHERE.
type WhereInterface interface {
	// Where добавляет условие WHERE с оператором AND.
	// Пример:
	// qb.Where("status = ?", "active")
	Where(condition string, args ...any) *Builder
	// WhereId добавляет условие WHERE для столбца 'id'.
	// Пример:
	// qb.WhereId(1)
	WhereId(id any) *Builder
	// OrWhere добавляет условие WHERE с оператором OR.
	// Пример:
	// qb.Where("status = ?", "active").OrWhere("status = ?", "pending")
	OrWhere(condition string, args ...any) *Builder
	// WhereIn добавляет условие WHERE IN.
	// Пример:
	// qb.WhereIn("id", 1, 2, 3)
	WhereIn(column string, values ...any) *Builder
	// WhereGroup добавляет сгруппированное условие WHERE.
	// Пример:
	// qb.WhereGroup(func(q *Builder) {
	//     q.Where("age > ?", 18).OrWhere("status = ?", "admin")
	// })
	WhereGroup(fn func(*Builder)) *Builder
	// OrWhereGroup добавляет сгруппированное условие WHERE с оператором OR.
	// Пример:
	// qb.OrWhereGroup(func(q *Builder) {
	//     q.Where("age < ?", 10).Where("parent_consent = ?", true)
	// })
	OrWhereGroup(fn func(*Builder)) *Builder
	// WhereExists добавляет условие подзапроса WHERE EXISTS.
	// Пример:
	// sub := qb.From("orders").WhereRaw("orders.user_id = users.id")
	// qb.From("users").WhereExists(sub)
	WhereExists(subQuery *Builder) *Builder
	// WhereNotExists добавляет условие подзапроса WHERE NOT EXISTS.
	// Пример:
	// sub := qb.From("orders").WhereRaw("orders.user_id = users.id")
	// qb.From("users").WhereNotExists(sub)
	WhereNotExists(subQuery *Builder) *Builder
	// WhereNull добавляет условие WHERE IS NULL.
	// Пример:
	// qb.WhereNull("deleted_at")
	WhereNull(column string) *Builder
	// WhereNotNull добавляет условие WHERE IS NOT NULL.
	// Пример:
	// qb.WhereNotNull("email")
	WhereNotNull(column string) *Builder
	// WhereBetween добавляет условие WHERE BETWEEN.
	// Пример:
	// qb.WhereBetween("age", 18, 65)
	WhereBetween(column string, start, end any) *Builder
	// WhereNotBetween добавляет условие WHERE NOT BETWEEN.
	// Пример:
	// qb.WhereNotBetween("price", 100, 200)
	WhereNotBetween(column string, start, end any) *Builder
	// WhereRaw добавляет сырое условие WHERE с аргументами.
	// Пример:
	// qb.WhereRaw("id = ? AND name LIKE ?", 1, "J%")
	WhereRaw(sql string, args ...any) *Builder
	// OrWhereRaw добавляет сырое условие WHERE с оператором OR.
	// Пример:
	// qb.OrWhereRaw("status = 'active' OR created_at > ?", time.Now().Add(-time.Hour))
	OrWhereRaw(sql string, args ...any) *Builder
}
type JoinInterface interface {
	// Join добавляет предложение INNER JOIN.
	// Пример:
	// qb.Join("orders", "users.id = orders.user_id")
	Join(table string, condition string) *Builder
	// LeftJoin добавляет предложение LEFT JOIN.
	// Пример:
	// qb.LeftJoin("profiles", "users.id = profiles.user_id")
	LeftJoin(table string, condition string) *Builder
	// RightJoin добавляет предложение RIGHT JOIN.
	// Пример:
	// qb.RightJoin("products", "categories.id = products.category_id")
	RightJoin(table string, condition string) *Builder
	// CrossJoin добавляет предложение CROSS JOIN.
	// Пример:
	// qb.CrossJoin("product_variants")
	CrossJoin(table string) *Builder
	// As устанавливает псевдоним для текущей таблицы в предложении FROM.
	// Пример:
	// qb.From("users").As("u").Select("u.name")
	As(alias string) *Builder
}
type DateInterface interface {
	// WhereDate добавляет условие WHERE, сравнивающее только часть даты столбца.
	// Пример:
	// qb.WhereDate("created_at", "=", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))
	WhereDate(column string, operator string, value time.Time) *Builder
	// WhereBetweenDates добавляет условие WHERE для диапазона дат (включительно).
	// Пример:
	// qb.WhereBetweenDates("order_date", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC))
	WhereBetweenDates(column string, start time.Time, end time.Time) *Builder
	// WhereDateTime добавляет условие WHERE, сравнивающее полную дату и время столбца.
	// Пример:
	// qb.WhereDateTime("updated_at", ">", time.Now().Add(-time.Hour))
	WhereDateTime(column string, operator string, value time.Time) *Builder
	// WhereBetweenDateTime добавляет условие WHERE для диапазона даты и времени (включительно).
	// Пример:
	// qb.WhereBetweenDateTime("log_time", time.Now().Add(-time.Hour), time.Now())
	WhereBetweenDateTime(column string, start time.Time, end time.Time) *Builder
	// WhereYear добавляет условие WHERE, сравнивающее год столбца.
	// Пример:
	// qb.WhereYear("birth_date", "=", 1990)
	WhereYear(column string, operator string, year int) *Builder
	// WhereMonth добавляет условие WHERE, сравнивающее месяц столбца.
	// Пример:
	// qb.WhereMonth("event_date", "=", 7) // Июль
	WhereMonth(column string, operator string, month int) *Builder
	// WhereDay добавляет условие WHERE, сравнивающее день месяца столбца.
	// Пример:
	// qb.WhereDay("delivery_date", "=", 15)
	WhereDay(column string, operator string, day int) *Builder
	// WhereTime добавляет условие WHERE, сравнивающее только часть времени столбца.
	// Пример:
	// qb.WhereTime("appointment_time", ">", time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)) // После 9 утра
	WhereTime(column string, operator string, value time.Time) *Builder
	// WhereDateIsNull добавляет условие WHERE, проверяющее, является ли столбец даты NULL.
	// Пример:
	// qb.WhereDateIsNull("completed_at")
	WhereDateIsNull(column string) *Builder
	// WhereDateIsNotNull добавляет условие WHERE, проверяющее, является ли столбец даты NOT NULL.
	// Пример:
	// qb.WhereDateIsNotNull("published_at")
	WhereDateIsNotNull(column string) *Builder
	// WhereCurrentDate добавляет условие WHERE, сравнивающее столбец даты с текущей датой.
	// Пример:
	// qb.WhereCurrentDate("created_at", "=")
	WhereCurrentDate(column string, operator string) *Builder
	// WhereLastDays добавляет условие WHERE для записей за последние N дней.
	// Пример:
	// qb.WhereLastDays("login_date", 30) // Последние 30 дней
	WhereLastDays(column string, days int) *Builder
	// WhereWeekday добавляет условие WHERE, сравнивающее день недели столбца.
	// Пример:
	// qb.WhereWeekday("meeting_date", "=", 1) // Понедельник
	WhereWeekday(column string, operator string, weekday int) *Builder
	// WhereQuarter добавляет условие WHERE, сравнивающее квартал столбца.
	// Пример:
	// qb.WhereQuarter("fiscal_date", "=", 2)
	WhereQuarter(column string, operator string, quarter int) *Builder
	// WhereWeek добавляет условие WHERE, сравнивающее номер недели столбца.
	// Пример:
	// qb.WhereWeek("delivery_date", "=", 25)
	WhereWeek(column string, operator string, week int) *Builder
	// WhereDateRange добавляет условие WHERE для диапазона дат с опциями включения/исключения границ.
	// Пример:
	// qb.WhereDateRange("period_start", t1, t2, true) // Включительно
	WhereDateRange(column string, start time.Time, end time.Time, inclusive bool) *Builder
	// WhereNextDays добавляет условие WHERE для записей в течение следующих N дней от текущей даты.
	// Пример:
	// qb.WhereNextDays("due_date", 7) // Следующие 7 дней
	WhereNextDays(column string, days int) *Builder
	// WhereDateBetweenColumns добавляет условие WHERE, проверяющее, находится ли дата между значениями двух других столбцов даты.
	// Пример:
	// qb.WhereDateBetweenColumns("event_date", "start_date", "end_date")
	WhereDateBetweenColumns(dateColumn string, startColumn string, endColumn string) *Builder
	// WhereAge добавляет условие WHERE на основе возраста, полученного из столбца даты (например, даты рождения).
	// Пример:
	// qb.WhereAge("birth_date", ">=", 18)
	WhereAge(column string, operator string, age int) *Builder
	// WhereDateDiff добавляет условие WHERE на основе разницы между двумя столбцами даты.
	// Пример:
	// qb.WhereDateDiff("end_date", "start_date", ">", 7) // Разница > 7 дней
	WhereDateDiff(column1 string, column2 string, operator string, days int) *Builder
	// WhereDateTrunc добавляет условие WHERE с усечением даты.
	// Пример:
	// qb.WhereDateTrunc("month", "created_at", "=", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))
	WhereDateTrunc(part string, column string, operator string, value time.Time) *Builder
	// WhereTimeWindow добавляет условие WHERE для временного окна (например, между 9:00 и 17:00).
	// Пример:
	// qb.WhereTimeWindow("log_time", time.Date(0,0,0,9,0,0,0,time.UTC), time.Date(0,0,0,17,0,0,0,time.UTC))
	WhereTimeWindow(column string, startTime, endTime time.Time) *Builder
	// WhereBusinessDays добавляет условие WHERE для фильтрации по рабочим дням (понедельник-пятница).
	// Пример:
	// qb.WhereBusinessDays("delivery_date")
	WhereBusinessDays(column string) *Builder
	// WhereDateFormat добавляет условие WHERE, сравнивающее отформатированную строку даты.
	// Пример:
	// qb.WhereDateFormat("created_at", "%Y-%m-%d", "=", "2023-01-01")
	WhereDateFormat(column string, format string, operator string, value string) *Builder
	// WhereTimeZone добавляет условие WHERE с учетом указанного часового пояса.
	// Пример:
	// qb.WhereTimeZone("event_time", "=", time.Now(), "America/New_York")
	WhereTimeZone(column string, operator string, value time.Time, timezone string) *Builder
}
type WindowInterface interface {
	// Window добавляет общую оконную функцию в предложение SELECT.
	// Пример:
	// qb.Window("SUM(amount)", "category_id", "order_date")
	Window(column string, partition string, orderBy string) *Builder
	// RowNumber добавляет оконную функцию ROW_NUMBER().
	// Пример:
	// qb.RowNumber("category_id", "price DESC", "row_num")
	RowNumber(partition string, orderBy string, alias string) *Builder
	// Rank добавляет оконную функцию RANK().
	// Пример:
	// qb.Rank("department_id", "salary DESC", "rank_in_dept")
	Rank(partition string, orderBy string, alias string) *Builder
	// DenseRank добавляет оконную функцию DENSE_RANK().
	// Пример:
	// qb.DenseRank("score_group", "score DESC", "dense_rank_score")
	DenseRank(partition string, orderBy string, alias string) *Builder
}
type LockInterface interface {
	// LockForUpdate добавляет предложение FOR UPDATE для блокировки выбранных строк для обновления.
	// Пример:
	// qb.LockForUpdate()
	LockForUpdate() *Builder
	// LockForShare добавляет предложение FOR SHARE для блокировки выбранных строк для общего чтения.
	// Пример:
	// qb.LockForShare()
	LockForShare() *Builder
	// SkipLocked добавляет предложение SKIP LOCKED для пропуска строк, которые в данный момент заблокированы.
	// Пример:
	// qb.SkipLocked()
	SkipLocked() *Builder
	// NoWait добавляет предложение NOWAIT для немедленного возврата ошибки, если строки заблокированы.
	// Пример:
	// qb.NoWait()
	NoWait() *Builder
	// Lock добавляет общий режим блокировки к запросу.
	// Пример:
	// qb.Lock("IN SHARE MODE") // Специфично для MySQL
	Lock(mode string) *Builder
}
type AggregateInterface interface {
	// Avg вычисляет среднее значение числового столбца.
	// Пример:
	// avgPrice, err := qb.From("products").Avg("price")
	Avg(column string) (float64, error)
	// Sum вычисляет сумму числового столбца.
	// Пример:
	// totalSales, err := qb.From("orders").Sum("amount")
	Sum(column string) (float64, error)
	// Min находит минимальное значение столбца.
	// Пример:
	// minAge, err := qb.From("users").Min("age")
	Min(column string) (float64, error)
	// Max находит максимальное значение столбца.
	// Пример:
	// maxDate, err := qb.From("events").Max("event_date")
	Max(column string) (float64, error)
	// Count возвращает количество записей, соответствующих запросу.
	// Пример:
	// userCount, err := qb.From("users").Where("status = ?", "active").Count()
	Count() (int64, error)
	// Exists проверяет, существуют ли какие-либо записи, соответствующие запросу.
	// Пример:
	// userExists, err := qb.From("users").Where("email = ?", "test@example.com").Exists()
	Exists() (bool, error)
}

// PaginationInterface определяет методы для пагинации результатов запроса.
type PaginationInterface interface {
	// Paginate выполняет пагинацию на основе смещения.
	// Пример:
	// var users []User
	// result, err := qb.From("users").Paginate(1, 10, &users)
	Paginate(page int, perPage int, dest any) (*PaginationResult, error)
	// PaginateWithToken выполняет пагинацию на основе токена (keyset).
	// Пример:
	// var products []Product
	// result, err := qb.From("products").OrderByAsc("id").PaginateWithToken("", 20, &products)
	PaginateWithToken(token string, limit int, dest any) (*PaginationTokenResult, error)
	// PaginateWithCursor выполняет пагинацию на основе курсора.
	// Пример:
	// var items []Item
	// result, err := qb.From("items").OrderByAsc("id").PaginateWithCursor("", 50, &items)
	PaginateWithCursor(cursor string, limit int, dest any) (*CursorPagination, error)
}

// SubQueryInterface определяет методы для работы с подзапросами.
type SubQueryInterface interface {
	// SubQuery создает подзапрос из текущего состояния построителя.
	// Пример:
	// sub := qb.From("orders").Select("user_id").Where("amount > ?", 100)
	// qb.From("users").WhereIn("id", sub.SubQuery("sub_users"))
	SubQuery(alias string) *Builder
	// WhereSubQuery добавляет условие WHERE на основе подзапроса.
	// Пример:
	// qb.WhereSubQuery("id", "IN", qb.From("orders").Select("user_id").Where("amount > ?", 100))
	WhereSubQuery(column string, operator string, subQuery *Builder) *Builder
	// Union объединяет текущий запрос с другим с использованием UNION.
	// Пример:
	// qb1 := qb.From("users").Select("name")
	// qb2 := qb.From("customers").Select("name")
	// combined := qb1.Union(qb2)
	Union(other *Builder) *Builder
	// UnionAll объединяет текущий запрос с другим с использованием UNION ALL.
	// Пример:
	// qb1 := qb.From("users").Select("name")
	// qb2 := qb.From("customers").Select("name")
	// combined := qb1.UnionAll(qb2)
	UnionAll(other *Builder) *Builder
}

// QueryOptionsInterface определяет методы для дополнительных опций запроса.
type QueryOptionsInterface interface {
	// Distinct добавляет предложение DISTINCT к оператору SELECT.
	// Пример:
	// qb.Distinct("category_id")
	// qb.Distinct() // SELECT DISTINCT *
	Distinct(columns ...string) *Builder
	// WithTransaction устанавливает построитель для использования существующей транзакции.
	// Пример:
	// tx, _ := qb.Begin()
	// qb.From("users").WithTransaction(tx).UpdateMap(map[string]any{"status": "active"})
	WithTransaction(tx *Transaction) *Builder
	// WithAudit включает аудит для запроса, отслеживая изменения по идентификатору пользователя.
	// Пример:
	// qb.From("products").WithAudit(userID).UpdateMap(map[string]any{"price": 29.99})
	WithAudit(userID any) *Builder
	// WithMetrics включает сбор метрик производительности запросов.
	// Пример:
	// collector := NewMetricsCollector()
	// qb.From("users").WithMetrics(collector).Get(&users)
	WithMetrics(collector *MetricsCollector) *Builder
	// Context устанавливает контекст для построителя запросов.
	// Пример:
	// qb.Context(ctx).From("users").Get(&users)
	Context(ctx context.Context) *Builder
}

// QueueInterface определяет методы для работы с фоновой очередью.
type QueueInterface interface {
	// ProcessQueue обрабатывает ожидающие операции из очереди.
	// Пример:
	// qb.From("queued_operations").ProcessQueue(func(op QueuedOperation) error {
	//     // Обработать операцию
	//     return nil
	// })
	ProcessQueue(handler func(QueuedOperation) error) error
	// Queue добавляет операцию в очередь для последующей обработки.
	// Пример:
	// qb.From("queued_operations").Queue("send_email", map[string]any{"to": "a@b.com"}, time.Now().Add(time.Hour))
	Queue(operation string, data any, runAt time.Time) error
}

// EventInterface определяет методы для обработки событий в построителе запросов.
type EventInterface interface {
	// On регистрирует обработчик события для определенного типа события.
	// Пример:
	// qb.On(BeforeCreate, func(data any) error {
	//     fmt.Println("Перед созданием:", data)
	//     return nil
	// })
	On(event EventType, handler EventHandler)
	// Trigger отправляет событие с связанными данными.
	// Пример:
	// qb.Trigger(AfterUpdate, updatedUser)
	Trigger(event EventType, data any)
}

// SpecialQueriesInterface определяет методы для специализированных типов запросов.
type SpecialQueriesInterface interface {
	// GeoSearch выполняет геопространственный поиск на основе столбца, точки и радиуса.
	// Пример:
	// qb.From("locations").GeoSearch("coordinates", Point{Lat: 34.0522, Lng: -118.2437}, 1000) // 1000 метров
	GeoSearch(column string, point Point, radius float64) *Builder
	// Search выполняет полнотекстовый поиск по указанным столбцам.
	// Пример:
	// qb.From("articles").Search([]string{"title", "content"}, "Go programming")
	Search(columns []string, query string) *Builder
}

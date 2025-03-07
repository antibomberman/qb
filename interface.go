package qb

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

// Основные интерфейсы для работы с базой данных
type Executor interface {
	sqlx.Ext
	Get(dest any, query string, args ...any) error
	Select(dest any, query string, args ...any) error
	NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	GetContext(ctx context.Context, dest any, query string, args ...any) error
}

type DBInterface interface {
	sqlx.Ext
	Get(dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	Beginx() (*sqlx.Tx, error)
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
}

// Интерфейс для работы с транзакциями
type TxInterface interface {
	sqlx.Ext
	Get(dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	Commit() error
	Rollback() error
}

// Основной интерфейс построителя запросов
type QueryBuilderInterface interface {
	// Основные методы
	From(table string) BuilderInterface
	Raw(query string, args ...any) *RawQuery
	GetDB() DBInterface
	SetLogger(logger *slog.Logger)

	// Транзакции
	Begin() (*Transaction, error)
	BeginContext(ctx context.Context) (*Transaction, error)
	Transaction(fn func(*Transaction) error) error
	TransactionContext(ctx context.Context, fn func(*Transaction) error) error

	// Логирование
	Debug(msg string, start time.Time, query string, args ...any)
	Info(msg string, start time.Time, query string, args ...any)
	Warn(msg string, start time.Time, query string, args ...any)
	Error(msg string, start time.Time, query string, args ...any)
}

// Интерфейс построителя запросов
type BuilderInterface interface {
	CrudInterface
	WhereInterface
	DateInterface
	WindowInterface
	LockInterface
	AggregateInterface
	SoftDeleteInterface
	JoinInterface
	PaginationInterface
	GroupSortInterface
	SubQueryInterface
	QueryOptionsInterface
	QueueInterface
	EventInterface
	SpecialQueriesInterface

	// Базовые операции
	Select(columns ...string) *Builder
	Limit(limit int) *Builder
	Offset(offset int) *Builder

	// Подзапросы
	SubQuery(alias string) *Builder
	WhereSubQuery(column string, operator string, subQuery *Builder) *Builder
	Union(other *Builder) *Builder
	UnionAll(other *Builder) *Builder

	// Дополнительные операции
	Pluck(column string, dest any) error
	Value(column string) (any, error)
	Values(column string) ([]any, error)
	Chunk(size int, fn func(items any) error) error
	ChunkContext(ctx context.Context, size int, fn func(context.Context, any) error) error
	WithinGroup(column string, window string) *Builder
	Distinct(columns ...string) *Builder
	WithTransaction(tx *Transaction) *Builder

	// Специальные запросы
	GeoSearch(column string, point Point, radius float64) *Builder
	Search(columns []string, query string) *Builder

	// Дополнительный функционал
	WithAudit(userID any) *Builder
	ProcessQueue(handler func(QueuedOperation) error) error
	Queue(operation string, data any, runAt time.Time) error
	WithMetrics(collector *MetricsCollector) *Builder
	On(event EventType, handler EventHandler)
	Trigger(event EventType, data any)
	Context(ctx context.Context) *Builder
}

// Вспомогательные интерфейсы
type CacheInterface interface {
	Get(key string) (any, bool)
	Set(key string, value any, expiration time.Duration)
	Delete(key string)
	Clear() error
}

type CrudInterface interface {
	Find(id any, dest any) (bool, error)
	FindAsync(id any, dest any) (chan bool, chan error)
	Get(dest any) (bool, error)
	GetAsync(dest any) (chan bool, chan error)
	First(dest any) (bool, error)
	FirstAsync(dest any) (chan bool, chan error)
	Create(data any, fields ...string) (any, error)
	CreateAsync(data any, fields ...string) (chan any, chan error)
	CreateMap(data map[string]any) (any, error)
	CreateMapAsync(data map[string]any) (chan any, chan error)
	Update(data any, fields ...string) error
	UpdateAsync(data any, fields ...string) chan error
	UpdateMap(data map[string]any) error
	UpdateMapAsync(data map[string]any) chan error
	Delete() error
	DeleteAsync() chan error
	Increment(column string, value any) error
	Decrement(column string, value any) error
	BatchInsert(records []map[string]any) error
	BatchInsertAsync(records []map[string]any) chan error
	BulkInsert(records []map[string]any) error
	BulkInsertAsync(records []map[string]any) chan error
	BulkUpdate(records []map[string]any, keyColumn string) error
	BulkUpdateAsync(records []map[string]any, keyColumn string) chan error
	BatchUpdate(records []map[string]any, keyColumn string, batchSize int) error
}

type GroupSortInterface interface {
	// Группировка и сортировка
	OrderBy(column string, direction string) *Builder
	GroupBy(columns ...string) *Builder
	Having(condition string) *Builder
	HavingRaw(sql string, args ...any) *Builder
}

type WhereInterface interface {
	Where(condition string, args ...any) *Builder
	WhereId(id any) *Builder
	OrWhere(condition string, args ...any) *Builder
	WhereIn(column string, values ...any) *Builder
	WhereGroup(fn func(*Builder)) *Builder
	OrWhereGroup(fn func(*Builder)) *Builder
	WhereExists(subQuery *Builder) *Builder
	WhereNotExists(subQuery *Builder) *Builder
	WhereNull(column string) *Builder
	WhereNotNull(column string) *Builder
	WhereBetween(column string, start, end any) *Builder
	WhereNotBetween(column string, start, end any) *Builder
	WhereRaw(sql string, args ...any) *Builder
	OrWhereRaw(sql string, args ...any) *Builder
}
type JoinInterface interface {
	Join(table string, condition string) *Builder
	LeftJoin(table string, condition string) *Builder
	RightJoin(table string, condition string) *Builder
	CrossJoin(table string) *Builder
	As(alias string) *Builder
}
type DateInterface interface {
	// Методы для работы с датами
	WhereDate(column string, operator string, value time.Time) *Builder
	WhereBetweenDates(column string, start time.Time, end time.Time) *Builder
	WhereDateTime(column string, operator string, value time.Time) *Builder
	WhereBetweenDateTime(column string, start time.Time, end time.Time) *Builder
	WhereYear(column string, operator string, year int) *Builder
	WhereMonth(column string, operator string, month int) *Builder
	WhereDay(column string, operator string, day int) *Builder
	WhereTime(column string, operator string, value time.Time) *Builder
	WhereDateIsNull(column string) *Builder
	WhereDateIsNotNull(column string) *Builder
	WhereCurrentDate(column string, operator string) *Builder
	WhereLastDays(column string, days int) *Builder
	WhereWeekday(column string, operator string, weekday int) *Builder
	WhereQuarter(column string, operator string, quarter int) *Builder
	WhereWeek(column string, operator string, week int) *Builder
	WhereDateRange(column string, start time.Time, end time.Time, inclusive bool) *Builder
	WhereNextDays(column string, days int) *Builder
	WhereDateBetweenColumns(dateColumn string, startColumn string, endColumn string) *Builder
	WhereAge(column string, operator string, age int) *Builder
	WhereDateDiff(column1 string, column2 string, operator string, days int) *Builder
	WhereDateTrunc(part string, column string, operator string, value time.Time) *Builder
	WhereTimeWindow(column string, startTime, endTime time.Time) *Builder
	WhereBusinessDays(column string) *Builder
	WhereDateFormat(column string, format string, operator string, value string) *Builder
	WhereTimeZone(column string, operator string, value time.Time, timezone string) *Builder
}
type WindowInterface interface {
	Window(column string, partition string, orderBy string) *Builder
	RowNumber(partition string, orderBy string, alias string) *Builder
	Rank(partition string, orderBy string, alias string) *Builder
	DenseRank(partition string, orderBy string, alias string) *Builder
}
type LockInterface interface {
	LockForUpdate() *Builder
	LockForShare() *Builder
	SkipLocked() *Builder
	NoWait() *Builder
	Lock(mode string) *Builder
}
type AggregateInterface interface {
	Avg(column string) (float64, error)
	Sum(column string) (float64, error)
	Min(column string) (float64, error)
	Max(column string) (float64, error)
	Count() (int64, error)
}
type SoftDeleteInterface interface {
	WithTrashed() *Builder
	OnlyTrashed() *Builder
	SoftDelete() error
	Restore() error
}
type PaginationInterface interface {
	Paginate(page int, perPage int, dest any) (*PaginationResult, error)
	PaginateWithToken(token string, limit int, dest any) (*PaginationTokenResult, error)
	PaginateWithCursor(cursor string, limit int, dest any) (*CursorPagination, error)
}

// Создадим новый интерфейс для работы с подзапросами
type SubQueryInterface interface {
	SubQuery(alias string) *Builder
	WhereSubQuery(column string, operator string, subQuery *Builder) *Builder
	Union(other *Builder) *Builder
	UnionAll(other *Builder) *Builder
}

// Создадим интерфейс для дополнительных опций запроса
type QueryOptionsInterface interface {
	Distinct(columns ...string) *Builder
	WithTransaction(tx *Transaction) *Builder
	WithAudit(userID any) *Builder
	WithMetrics(collector *MetricsCollector) *Builder
	Context(ctx context.Context) *Builder
}

// Создадим интерфейс для работы с очередями
type QueueInterface interface {
	ProcessQueue(handler func(QueuedOperation) error) error
	Queue(operation string, data any, runAt time.Time) error
}

// Создадим интерфейс для работы с событиями
type EventInterface interface {
	On(event EventType, handler EventHandler)
	Trigger(event EventType, data any)
}

// Создадим интерфейс для специальных запросов
type SpecialQueriesInterface interface {
	GeoSearch(column string, point Point, radius float64) *Builder
	Search(columns []string, query string) *Builder
}

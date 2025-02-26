package qb

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

//go:generate mockery --name DBInterface
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

//go:generate mockery --name TxInterface
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

//go:generate mockery --name QueryBuilderInterface
type QueryBuilderInterface interface {
	// Основные методы
	Query(table string) BuilderInterface
	Raw(query string, args ...any) *RawQuery
	GetDB() DBInterface
	SetLogger(logger *slog.Logger)

	// Транзакции
	Begin() (*Transaction, error)
	BeginContext(ctx context.Context) (*Transaction, error)
	Transaction(fn func(*Transaction) error) error
	TransactionContext(ctx context.Context, fn func(*Transaction) error) error

	// Логирование
	Debug(start time.Time, query string, args ...any)
	Info(start time.Time, query string, args ...any)
	Warn(start time.Time, query string, args ...any)
	Error(start time.Time, query string, args ...any)
}

//go:generate mockery --name BuilderInterface
type BuilderInterface interface {
	// Построение запросов
	Select(columns ...string) *Builder
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

	// Джойны
	Join(table string, condition string) *Builder
	LeftJoin(table string, condition string) *Builder
	RightJoin(table string, condition string) *Builder
	CrossJoin(table string) *Builder
	As(alias string) *Builder

	// Группировка и сортировка
	OrderBy(column string, direction string) *Builder
	GroupBy(columns ...string) *Builder
	Having(condition string) *Builder
	HavingRaw(sql string, args ...any) *Builder

	// Лимиты и смещение
	Limit(limit int) *Builder
	Offset(offset int) *Builder

	// CRUD операции
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

	// Пакетные операции
	BatchInsert(records []map[string]any) error
	BatchInsertAsync(records []map[string]any) chan error
	BulkInsert(records []map[string]any) error
	BulkInsertAsync(records []map[string]any) chan error
	BulkUpdate(records []map[string]any, keyColumn string) error
	BulkUpdateAsync(records []map[string]any, keyColumn string) chan error
	BatchUpdate(records []map[string]any, keyColumn string, batchSize int) error

	// Подзапросы и объединения
	SubQuery(alias string) *Builder
	WhereSubQuery(column string, operator string, subQuery *Builder) *Builder
	Union(other *Builder) *Builder
	UnionAll(other *Builder) *Builder

	// Блокировки
	LockForUpdate() *Builder
	LockForShare() *Builder
	SkipLocked() *Builder
	NoWait() *Builder
	Lock(mode string) *Builder

	// Оконные функции
	Window(column string, partition string, orderBy string) *Builder
	RowNumber(partition string, orderBy string, alias string) *Builder
	Rank(partition string, orderBy string, alias string) *Builder
	DenseRank(partition string, orderBy string, alias string) *Builder

	// Дополнительные операции
	Increment(column string, value any) error
	Decrement(column string, value any) error
	Pluck(column string, dest any) error
	Value(column string) (any, error)
	Values(column string) ([]any, error)
	Chunk(size int, fn func(items any) error) error
	ChunkContext(ctx context.Context, size int, fn func(context.Context, any) error) error
	WithinGroup(column string, window string) *Builder
	Distinct(columns ...string) *Builder
	WithTransaction(tx *Transaction) *Builder

	// Геопространственные запросы
	GeoSearch(column string, point Point, radius float64) *Builder

	// Полнотекстовый поиск
	Search(columns []string, query string) *Builder

	// Soft Delete
	WithTrashed() *Builder
	OnlyTrashed() *Builder
	SoftDelete() error
	Restore() error

	// Аудит
	WithAudit(userID any) *Builder

	// Очереди
	ProcessQueue(handler func(QueuedOperation) error) error

	// Очереди
	Queue(operation string, data any, runAt time.Time) error

	// Метрики
	WithMetrics(collector *MetricsCollector) *Builder
	// События
	On(event EventType, handler EventHandler)
	Trigger(event EventType, data any)

	// Контекст
	Context(ctx context.Context) *Builder

	// Агрегатные функции
	Avg(column string) (float64, error)
	Sum(column string) (float64, error)
	Min(column string) (float64, error)
	Max(column string) (float64, error)
	Count() (int64, error)
	Exists() (bool, error)

	// Пагинация
	Paginate(page int, perPage int, dest any) (*PaginationResult, error)
	PaginateWithToken(token string, limit int, dest any) (*PaginationTokenResult, error)
	PaginateWithCursor(cursor string, limit int, dest any) (*CursorPagination, error)

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

type QueryBuilder struct {
	db         DBInterface
	driverName string
	cache      CacheInterface
	logger     *slog.Logger
}

func New(driverName string, db *sql.DB) QueryBuilderInterface {
	return NewX(driverName, sqlx.NewDb(db, driverName))
}
func NewX(driverName string, db *sqlx.DB) QueryBuilderInterface {
	return &QueryBuilder{
		db:         db,
		driverName: driverName,
		cache:      NewCacheMemory(),
		logger:     NewDefaultLogger(),
	}
}

func (q *QueryBuilder) Query(table string) BuilderInterface {
	return &Builder{
		tableName:    table,
		db:           q.db,
		queryBuilder: q,
		ctx:          context.TODO(),
	}
}

func (t *Transaction) Query(table string) BuilderInterface {
	return &Builder{
		tableName:    table,
		db:           t.Tx,
		queryBuilder: t.QueryBuilder,
		ctx:          context.TODO(),
	}
}

func (t *Transaction) Raw(query string, args ...any) *RawQuery {
	return &RawQuery{
		query:        query,
		args:         args,
		db:           t.Tx,
		queryBuilder: t.QueryBuilder,
	}
}

func (q *QueryBuilder) GetDB() DBInterface {
	return q.db
}

func (q *QueryBuilder) SetLogger(logger *slog.Logger) {
	q.logger = logger
}

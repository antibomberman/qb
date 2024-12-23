package dblayer

// Реализации для разных БД
type MySQLDialect struct{}
type PostgresDialect struct{}
type SQLiteDialect struct{}

type Dialect interface {
	// GetAutoIncrement() string
	// GetTimestampType() string
	// SupportsJSON() bool
	// GetCreateTableSQL(schema *Schema) string
}

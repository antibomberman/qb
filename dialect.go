package dblayer

// Реализации для разных БД
type MysqlQueryDialect struct{}
type PostgresQueryDialect struct{}
type SqliteQueryDialect struct{}

type QueryDialect interface {
}

type MysqlSchemaDialect struct{}
type PostgresSchemaDialect struct{}
type SqliteSchemaDialect struct{}

// SchemaDialect интерфейс с базовыми методами
type SchemaDialect interface {
	BaseSchemaDialect
	TypesDialect
	IndexDialect
	QuotingDialect
}

type BaseSchemaDialect interface {
	BuildCreateTable(s *Schema) string
	BuildAlterTable(s *Schema) string
	BuildDropTable(dt *DropTable) string
	BuildTruncateTable(tt *TruncateTable) string
	BuildColumnDefinition(col Column) string
	BuildForeignKeyDefinition(fk *ForeignKey) string
	SupportsDropConcurrently() bool
	SupportsRestartIdentity() bool
	SupportsCascade() bool
	SupportsForce() bool
}

type TypesDialect interface {
	GetAutoIncrementType() string
	GetUUIDType() string
	GetBooleanType() string
	GetIntegerType() string
	GetBigIntegerType() string
	GetFloatType() string
	GetDoubleType() string
	GetDecimalType(precision, scale int) string
	GetStringType(length int) string
	GetTextType() string
	GetBinaryType(length int) string
	GetJsonType() string
	GetTimestampType() string
	GetDateType() string
	GetTimeType() string

	// Поддержка функций
	GetCurrentTimestampExpression() string
}

type IndexDialect interface {
	BuildIndexDefinition(name string, columns []string, unique bool) string
	BuildSpatialIndexDefinition(name string, columns []string) string
	BuildFullTextIndexDefinition(name string, columns []string) string
	SupportsSpatialIndex() bool
	SupportsFullTextIndex() bool
}

type QuotingDialect interface {
	QuoteIdentifier(name string) string
	QuoteString(value string) string
}

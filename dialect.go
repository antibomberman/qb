package dblayer

import "strings"

// Реализации для разных БД
type MysqlQueryDialect struct{}
type PostgresQueryDialect struct{}
type SqliteQueryDialect struct{}

type QueryDialect interface {
}

type MysqlSchemaDialect struct{}
type PostgresSchemaDialect struct{}
type SqliteSchemaDialect struct{}

// BaseDialect объединяет общую функциональность
type BaseDialect interface {
	SchemaDialect
	QueryDialect
}

// SchemaDialect разделить на более специализированные интерфейсы
type SchemaDialect interface {
	TableDialect
	ColumnDialect
	ConstraintDialect
	TypeDialect
	QuotingDialect
	IndexDialect
}

type TableDialect interface {
	BuildCreateTable(s *Schema) string
	BuildAlterTable(s *Schema) string
	BuildDropTable(dt *DropTable) string
	BuildTruncateTable(tt *TruncateTable) string
	SupportsDropConcurrently() bool
	SupportsRestartIdentity() bool
	SupportsCascade() bool
	SupportsForce() bool
}

type ColumnDialect interface {
	BuildColumnDefinition(col Column) string
	SupportsColumnPositioning() bool
	SupportsColumnComments() bool
}

type ConstraintDialect interface {
	BuildIndexDefinition(name string, columns []string, unique bool) string
	BuildForeignKeyDefinition(fk *ForeignKey) string
	BuildSpatialIndexDefinition(name string, columns []string) string
	BuildFullTextIndexDefinition(name string, columns []string) string
}

type TypeDialect interface {
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

type BaseDialectImpl struct {
	driverName string
}

// Общие методы для всех диалектов
func (d *BaseDialectImpl) QuoteIdentifier(name string) string {
	return `"` + strings.Replace(name, `"`, `""`, -1) + `"`
}

func (d *BaseDialectImpl) QuoteString(value string) string {
	return "'" + strings.Replace(value, "'", "''", -1) + "'"
}

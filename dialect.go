package dblayer

// Реализации для разных БД
type MysqlQueryDialect struct{}
type PostgresQueryDialect struct{}
type SqliteQueryDialect struct{}

type QueryDialect interface {

	// GetAutoIncrement() string
	// GetTimestampType() string
	// SupportsJSON() bool
	// GetCreateTableSQL(schema *Schema) string

}

type MysqlSchemaDialect struct{}
type PostgresSchemaDialect struct{}
type SqliteSchemaDialect struct{}

type SchemaDialect interface {
	BuildCreateTable(s *Schema) string
	BuildAlterTable(s *Schema) string
	BuildDropTable(dt *DropTable) string
	BuildColumnDefinition(col Column) string
	BuildIndexDefinition(name string, columns []string, unique bool) string
	BuildForeignKeyDefinition(fk *ForeignKey) string
}

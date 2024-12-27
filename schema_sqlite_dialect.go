package dblayer

import (
	"fmt"
	"strings"
)

func (g *SqliteDialect) BuildCreateTable(s *Schema) string {
	var sql strings.Builder

	sql.WriteString("CREATE ")

	sql.WriteString("TABLE ")
	if s.definition.options.ifNotExists {
		sql.WriteString("IF NOT EXISTS ")
	}
	sql.WriteString(s.definition.name)
	sql.WriteString(" (\n")

	// Колонки
	var columns []string
	for _, col := range s.definition.columns {
		columns = append(columns, g.BuildColumnDefinition(col))
	}

	// Первичный ключ
	if len(s.definition.constraints.primaryKey) > 0 {
		columns = append(columns, fmt.Sprintf("PRIMARY KEY (%s)",
			strings.Join(s.definition.constraints.primaryKey, ", ")))
	}

	// Уникальные ключи
	for _, cols := range s.definition.constraints.uniqueKeys {
		columns = append(columns, fmt.Sprintf("UNIQUE (%s)",
			strings.Join(cols, ", ")))
	}

	// Внешние ключи
	for col, fk := range s.definition.constraints.foreignKeys {
		constraint := fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s)",
			col, fk.Table, fk.Column)
		if fk.OnDelete != "" {
			constraint += " ON DELETE " + fk.OnDelete
		}
		if fk.OnUpdate != "" {
			constraint += " ON UPDATE " + fk.OnUpdate
		}
		columns = append(columns, constraint)
	}

	sql.WriteString(strings.Join(columns, ",\n"))
	sql.WriteString("\n)")

	return sql.String()
}

func (g *SqliteDialect) BuildAlterTable(s *Schema) string {
	var commands []string
	for _, cmd := range s.definition.commands {
		commands = append(commands, cmd.Type+" "+cmd.Name)
	}

	return fmt.Sprintf(
		"ALTER TABLE %s\n%s",
		s.definition.name,
		strings.Join(commands, ";\nALTER TABLE "+s.definition.name+" "),
	)
}

func (g *SqliteDialect) BuildDropTable(dt *DropTable) string {
	var sql strings.Builder

	sql.WriteString("DROP ")
	if dt.options.Temporary {
		sql.WriteString("TEMPORARY ")
	}
	sql.WriteString("TABLE ")

	if dt.options.Concurrent && dt.dbl.db.DriverName() == "postgres" {
		sql.WriteString("CONCURRENTLY ")
	}

	if dt.options.IfExists {
		sql.WriteString("IF EXISTS ")
	}

	sql.WriteString(strings.Join(dt.tables, ", "))

	if dt.options.Cascade && dt.dbl.db.DriverName() == "postgres" {
		sql.WriteString(" CASCADE")
	}

	if dt.options.Restrict && dt.dbl.db.DriverName() == "postgres" {
		sql.WriteString(" RESTRICT")
	}

	if dt.options.Force && dt.dbl.db.DriverName() == "mysql" {
		sql.WriteString(" FORCE")
	}

	return sql.String()
}

func (g *SqliteDialect) BuildColumnDefinition(col Column) string {
	var sql strings.Builder

	sql.WriteString(col.Name)
	sql.WriteString(" ")

	// SQLite имеет упрощенную систему типов
	if col.Constraints.AutoIncrement {
		sql.WriteString("INTEGER PRIMARY KEY AUTOINCREMENT")
		return sql.String()
	}

	// Преобразование типов для SQLite
	switch strings.ToUpper(col.Definition.Type) {
	case "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "INTEGER", "BIGINT":
		sql.WriteString("INTEGER")
	case "DECIMAL", "NUMERIC", "REAL", "DOUBLE", "FLOAT":
		sql.WriteString("REAL")
	case "CHAR", "VARCHAR", "BINARY", "VARBINARY", "TINYTEXT", "TEXT", "MEDIUMTEXT", "LONGTEXT", "ENUM", "SET":
		sql.WriteString("TEXT")
	case "DATE", "DATETIME", "TIMESTAMP", "TIME", "YEAR":
		sql.WriteString("TEXT")
	case "BLOB", "TINYBLOB", "MEDIUMBLOB", "LONGBLOB":
		sql.WriteString("BLOB")
	default:
		sql.WriteString(col.Definition.Type)
	}

	if !col.Constraints.Nullable {
		sql.WriteString(" NOT NULL")
	}

	if col.Definition.Default != nil {
		sql.WriteString(fmt.Sprintf(" DEFAULT %v", col.Definition.Default))
	}

	// SQLite не поддерживает ON UPDATE напрямую
	// Можно реализовать через триггеры
	if col.Definition.OnUpdate != "" {
		// ON UPDATE реализуется через триггеры в SQLite
		// Здесь можно добавить генерацию триггера
	}

	return sql.String()
}

func (g *SqliteDialect) BuildIndexDefinition(name string, columns []string, unique bool) string {
	var sql strings.Builder

	sql.WriteString("CREATE ")
	if unique {
		sql.WriteString("UNIQUE ")
	}
	sql.WriteString("INDEX IF NOT EXISTS ")
	sql.WriteString(g.QuoteIdentifier(name))
	sql.WriteString(" ON ")
	// Имя таблицы будет добавлено позже
	sql.WriteString(" (")

	// Цитируем каждую колонку
	quotedColumns := make([]string, len(columns))
	for i, col := range columns {
		quotedColumns[i] = g.QuoteIdentifier(col)
	}
	sql.WriteString(strings.Join(quotedColumns, ", "))
	sql.WriteString(")")

	return sql.String()
}

func (g *SqliteDialect) BuildForeignKeyDefinition(fk *ForeignKey) string {
	var sql strings.Builder

	sql.WriteString("REFERENCES ")
	sql.WriteString(g.QuoteIdentifier(fk.Table))
	sql.WriteString("(")
	sql.WriteString(g.QuoteIdentifier(fk.Column))
	sql.WriteString(")")

	if fk.OnDelete != "" {
		sql.WriteString(" ON DELETE ")
		sql.WriteString(fk.OnDelete)
	}

	if fk.OnUpdate != "" {
		sql.WriteString(" ON UPDATE ")
		sql.WriteString(fk.OnUpdate)
	}

	return sql.String()
}

func (g *SqliteDialect) BuildTruncateTable(tt *TruncateTable) string {
	// SQLite не поддерживает TRUNCATE, используем DELETE
	return fmt.Sprintf("DELETE FROM %s", strings.Join(tt.tables, ", "))
}

func (g *SqliteDialect) SupportsDropConcurrently() bool {
	return false
}

func (g *SqliteDialect) SupportsRestartIdentity() bool {
	return false
}

func (g *SqliteDialect) SupportsCascade() bool {
	return false
}

func (g *SqliteDialect) SupportsForce() bool {
	return false
}

func (g *SqliteDialect) GetAutoIncrementType() string {
	return "INTEGER PRIMARY KEY AUTOINCREMENT"
}

func (g *SqliteDialect) GetUUIDType() string {
	return "TEXT"
}

func (g *SqliteDialect) GetBooleanType() string {
	return "INTEGER"
}

func (g *SqliteDialect) GetIntegerType() string {
	return "INTEGER"
}

func (g *SqliteDialect) GetBigIntegerType() string {
	return "INTEGER"
}

func (g *SqliteDialect) GetFloatType() string {
	return "REAL"
}

func (g *SqliteDialect) GetDoubleType() string {
	return "REAL"
}

func (g *SqliteDialect) GetDecimalType(precision, scale int) string {
	return "REAL"
}

func (g *SqliteDialect) GetStringType(length int) string {
	return "TEXT"
}

func (g *SqliteDialect) GetTextType() string {
	return "TEXT"
}

func (g *SqliteDialect) GetBinaryType(length int) string {
	return "BLOB"
}

func (g *SqliteDialect) GetJsonType() string {
	return "TEXT"
}

func (g *SqliteDialect) GetTimestampType() string {
	return "TEXT"
}

func (g *SqliteDialect) GetDateType() string {
	return "TEXT"
}

func (g *SqliteDialect) GetTimeType() string {
	return "TEXT"
}

func (g *SqliteDialect) GetCurrentTimestampExpression() string {
	return "CURRENT_TIMESTAMP"
}

func (g *SqliteDialect) QuoteIdentifier(name string) string {
	return "\"" + strings.Replace(name, "\"", "\"\"", -1) + "\""
}

func (g *SqliteDialect) QuoteString(value string) string {
	return "'" + strings.Replace(value, "'", "''", -1) + "'"
}

func (g *SqliteDialect) SupportsColumnPositioning() bool {
	return false
}

func (g *SqliteDialect) SupportsEnum() bool {
	return false
}

func (g *SqliteDialect) GetEnumType(values []string) string {
	return "TEXT"
}

func (g *SqliteDialect) SupportsColumnComments() bool {
	return false
}

func (g *SqliteDialect) SupportsSpatialIndex() bool {
	return false
}

func (g *SqliteDialect) SupportsFullTextIndex() bool {
	return false
}

func (g *SqliteDialect) BuildSpatialIndexDefinition(name string, columns []string) string {
	return ""
}

func (g *SqliteDialect) BuildFullTextIndexDefinition(name string, columns []string) string {
	return ""
}

func (g *SqliteDialect) GetSmallIntegerType() string {
	return "INTEGER"
}

func (g *SqliteDialect) GetMediumIntegerType() string {
	return "INTEGER"
}

func (g *SqliteDialect) GetTinyIntegerType() string {
	return "INTEGER"
}

func (g *SqliteDialect) GetMoneyType() string {
	return "REAL"
}

func (g *SqliteDialect) GetCharType(length int) string {
	return "TEXT"
}

func (g *SqliteDialect) GetMediumTextType() string {
	return "TEXT"
}

func (g *SqliteDialect) GetLongTextType() string {
	return "TEXT"
}

func (g *SqliteDialect) GetSetType(values []string) string {
	return "TEXT"
}

func (g *SqliteDialect) GetYearType() string {
	return "INTEGER"
}

func (g *SqliteDialect) GetPointType() string {
	return "TEXT"
}

func (g *SqliteDialect) GetPolygonType() string {
	return "TEXT"
}

func (g *SqliteDialect) GetGeometryType() string {
	return "TEXT"
}

func (g *SqliteDialect) GetIpType() string {
	return "TEXT"
}

func (g *SqliteDialect) GetMacAddressType() string {
	return "TEXT"
}

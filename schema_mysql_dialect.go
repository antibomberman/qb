package dblayer

import (
	"fmt"
	"strings"
)

func (g *MysqlSchemaDialect) BuildCreateTable(s *Schema) string {
	var sql strings.Builder

	sql.WriteString("CREATE ")

	sql.WriteString("TABLE ")
	if s.options.ifNotExists {
		sql.WriteString("IF NOT EXISTS ")
	}
	sql.WriteString(s.name)
	sql.WriteString(" (\n")

	// Колонки
	var columns []string
	for _, col := range s.columns {
		columns = append(columns, s.buildColumn(col))
	}

	// Первичный ключ
	if len(s.constraints.primaryKey) > 0 {
		columns = append(columns, fmt.Sprintf("PRIMARY KEY (%s)",
			strings.Join(s.constraints.primaryKey, ", ")))
	}

	// Уникальные ключи
	for name, cols := range s.constraints.uniqueKeys {
		columns = append(columns, fmt.Sprintf("UNIQUE KEY %s (%s)",
			name, strings.Join(cols, ", ")))
	}

	// Индексы
	for name, cols := range s.constraints.indexes {
		columns = append(columns, fmt.Sprintf("INDEX %s (%s)",
			name, strings.Join(cols, ", ")))
	}

	// Внешние ключи
	for col, fk := range s.constraints.foreignKeys {
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

	// Опции таблицы
	sql.WriteString(fmt.Sprintf(" ENGINE=%s", s.options.engine))
	sql.WriteString(fmt.Sprintf(" DEFAULT CHARSET=%s", s.options.charset))
	sql.WriteString(fmt.Sprintf(" COLLATE=%s", s.options.collate))
	if s.options.comment != "" {
		sql.WriteString(fmt.Sprintf(" COMMENT='%s'", s.options.comment))
	}

	return sql.String()
}

func (g *MysqlSchemaDialect) BuildAlterTable(s *Schema) string {
	return fmt.Sprintf(
		"ALTER TABLE %s\n%s",
		s.name,
		strings.Join(s.commands, ",\n"),
	)
}

func (g *MysqlSchemaDialect) BuildDropTable(dt *DropTable) string {
	var sql strings.Builder

	sql.WriteString("DROP ")
	if dt.options.Temporary {
		sql.WriteString("TEMPORARY ")
	}
	sql.WriteString("TABLE ")

	if dt.options.Concurrent {
		sql.WriteString("CONCURRENTLY ")
	}

	if dt.options.IfExists {
		sql.WriteString("IF EXISTS ")
	}

	sql.WriteString(strings.Join(dt.tables, ", "))

	if dt.options.Force {
		sql.WriteString(" FORCE")
	}

	return sql.String()
}

func (g *MysqlSchemaDialect) BuildColumnDefinition(col Column) string {
	var sql strings.Builder

	sql.WriteString(col.Name)
	sql.WriteString(" ")
	sql.WriteString(col.Definition.Type)

	if col.Definition.Length > 0 {
		sql.WriteString(fmt.Sprintf("(%d)", col.Definition.Length))
	}

	if !col.Constraints.Nullable {
		sql.WriteString(" NOT NULL")
	}

	if col.Definition.Default != nil {
		sql.WriteString(fmt.Sprintf(" DEFAULT %v", col.Definition.Default))
	}

	if col.Constraints.AutoIncrement {
		sql.WriteString(" AUTO_INCREMENT")
	}

	if col.Meta.Comment != "" {
		sql.WriteString(fmt.Sprintf(" COMMENT '%s'", strings.Replace(col.Meta.Comment, "'", "\\'", -1)))
	}

	return sql.String()
}

func (g *MysqlSchemaDialect) BuildTruncateTable(tt *TruncateTable) string {
	var sql strings.Builder
	sql.WriteString("TRUNCATE TABLE ")
	sql.WriteString(strings.Join(tt.tables, ", "))

	if tt.options.Force {
		sql.WriteString(" FORCE")
	}

	return sql.String()
}

func (g *MysqlSchemaDialect) BuildIndexDefinition(name string, columns []string, unique bool) string {
	var sql strings.Builder

	if unique {
		sql.WriteString("UNIQUE ")
	}
	sql.WriteString("INDEX ")
	sql.WriteString(g.QuoteIdentifier(name))
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

func (g *MysqlSchemaDialect) BuildForeignKeyDefinition(fk *ForeignKey) string {
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

func (g *MysqlSchemaDialect) SupportsDropConcurrently() bool {
	return false
}

func (g *MysqlSchemaDialect) SupportsRestartIdentity() bool {
	return false
}

func (g *MysqlSchemaDialect) SupportsCascade() bool {
	return false
}

func (g *MysqlSchemaDialect) SupportsForce() bool {
	return true
}

func (g *MysqlSchemaDialect) GetAutoIncrementType() string {
	return "AUTO_INCREMENT"
}

func (g *MysqlSchemaDialect) GetUUIDType() string {
	return "CHAR(36)"
}

// Типы данных
func (g *MysqlSchemaDialect) GetBooleanType() string {
	return "TINYINT(1)"
}

func (g *MysqlSchemaDialect) GetIntegerType() string {
	return "INT"
}

func (g *MysqlSchemaDialect) GetBigIntegerType() string {
	return "BIGINT"
}

func (g *MysqlSchemaDialect) GetFloatType() string {
	return "FLOAT"
}

func (g *MysqlSchemaDialect) GetDoubleType() string {
	return "DOUBLE"
}

func (g *MysqlSchemaDialect) GetDecimalType(precision, scale int) string {
	return fmt.Sprintf("DECIMAL(%d,%d)", precision, scale)
}

func (g *MysqlSchemaDialect) GetStringType(length int) string {
	return fmt.Sprintf("VARCHAR(%d)", length)
}

func (g *MysqlSchemaDialect) GetTextType() string {
	return "TEXT"
}

func (g *MysqlSchemaDialect) GetBinaryType(length int) string {
	return fmt.Sprintf("BINARY(%d)", length)
}

func (g *MysqlSchemaDialect) GetJsonType() string {
	return "JSON"
}

func (g *MysqlSchemaDialect) GetTimestampType() string {
	return "TIMESTAMP"
}

func (g *MysqlSchemaDialect) GetDateType() string {
	return "DATE"
}

func (g *MysqlSchemaDialect) GetTimeType() string {
	return "TIME"
}

func (g *MysqlSchemaDialect) GetCurrentTimestampExpression() string {
	return "CURRENT_TIMESTAMP"
}

func (g *MysqlSchemaDialect) QuoteIdentifier(name string) string {
	return "`" + strings.Replace(name, "`", "``", -1) + "`"
}

func (g *MysqlSchemaDialect) QuoteString(value string) string {
	return "'" + strings.Replace(value, "'", "''", -1) + "'"
}

func (g *MysqlSchemaDialect) SupportsColumnPositioning() bool {
	return true
}

func (g *MysqlSchemaDialect) SupportsEnum() bool {
	return true
}

func (g *MysqlSchemaDialect) GetEnumType(values []string) string {
	return fmt.Sprintf("ENUM('%s')", strings.Join(values, "','"))
}

func (g *MysqlSchemaDialect) SupportsColumnComments() bool {
	return true
}

// MySQL поддерживает оба типа индексов
func (g *MysqlSchemaDialect) SupportsSpatialIndex() bool {
	return true
}

func (g *MysqlSchemaDialect) SupportsFullTextIndex() bool {
	return true
}

// Добавляем методы для создания специальных индексов
func (g *MysqlSchemaDialect) BuildSpatialIndexDefinition(name string, columns []string) string {
	return fmt.Sprintf("SPATIAL INDEX %s (%s)", name, strings.Join(columns, ", "))
}

func (g *MysqlSchemaDialect) BuildFullTextIndexDefinition(name string, columns []string) string {
	return fmt.Sprintf("FULLTEXT INDEX %s (%s)", name, strings.Join(columns, ", "))
}

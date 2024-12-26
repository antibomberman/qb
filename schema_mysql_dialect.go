package dblayer

import (
	"fmt"
	"strings"
)

func (g *MysqlDialect) BuildCreateTable(s *Schema) string {
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
	for name, cols := range s.definition.constraints.uniqueKeys {
		columns = append(columns, fmt.Sprintf("UNIQUE KEY %s (%s)",
			name, strings.Join(cols, ", ")))
	}

	// Индексы
	for name, cols := range s.definition.constraints.indexes {
		columns = append(columns, fmt.Sprintf("INDEX %s (%s)",
			name, strings.Join(cols, ", ")))
	}

	// ��нешние ключи
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

	// Опции таблицы
	sql.WriteString(fmt.Sprintf(" ENGINE=%s", s.definition.options.engine))
	sql.WriteString(fmt.Sprintf(" DEFAULT CHARSET=%s", s.definition.options.charset))
	sql.WriteString(fmt.Sprintf(" COLLATE=%s", s.definition.options.collate))
	if s.definition.options.comment != "" {
		sql.WriteString(fmt.Sprintf(" COMMENT='%s'", s.definition.options.comment))
	}

	return sql.String()
}

func (g *MysqlDialect) BuildAlterTable(s *Schema) string {
	var commands []string
	for _, cmd := range s.definition.commands {
		commands = append(commands, fmt.Sprintf("%s %s", cmd.Type, cmd.Name))
	}

	return fmt.Sprintf(
		"ALTER TABLE %s\n%s",
		s.definition.name,
		strings.Join(commands, ",\n"),
	)
}

func (g *MysqlDialect) BuildDropTable(dt *DropTable) string {
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

func (g *MysqlDialect) BuildColumnDefinition(col Column) string {
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

	if col.Definition.OnUpdate != "" {
		sql.WriteString(" ON UPDATE " + col.Definition.OnUpdate)
	}

	if col.Constraints.AutoIncrement {
		sql.WriteString(" AUTO_INCREMENT")
	}

	if col.Meta.Comment != "" {
		sql.WriteString(fmt.Sprintf(" COMMENT '%s'", strings.Replace(col.Meta.Comment, "'", "\\'", -1)))
	}

	return sql.String()
}

func (g *MysqlDialect) BuildTruncateTable(tt *TruncateTable) string {
	var sql strings.Builder
	sql.WriteString("TRUNCATE TABLE ")
	sql.WriteString(strings.Join(tt.tables, ", "))

	if tt.options.Force {
		sql.WriteString(" FORCE")
	}

	return sql.String()
}

func (g *MysqlDialect) BuildIndexDefinition(name string, columns []string, unique bool) string {
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

func (g *MysqlDialect) BuildForeignKeyDefinition(fk *ForeignKey) string {
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

func (g *MysqlDialect) SupportsDropConcurrently() bool {
	return false
}

func (g *MysqlDialect) SupportsRestartIdentity() bool {
	return false
}

func (g *MysqlDialect) SupportsCascade() bool {
	return false
}

func (g *MysqlDialect) SupportsForce() bool {
	return true
}

func (g *MysqlDialect) GetAutoIncrementType() string {
	return "AUTO_INCREMENT"
}

func (g *MysqlDialect) GetUUIDType() string {
	return "CHAR(36)"
}

// Типы данных
func (g *MysqlDialect) GetBooleanType() string {
	return "TINYINT(1)"
}

func (g *MysqlDialect) GetIntegerType() string {
	return "INT"
}

func (g *MysqlDialect) GetBigIntegerType() string {
	return "BIGINT"
}

func (g *MysqlDialect) GetFloatType() string {
	return "FLOAT"
}

func (g *MysqlDialect) GetDoubleType() string {
	return "DOUBLE"
}

func (g *MysqlDialect) GetDecimalType(precision, scale int) string {
	return fmt.Sprintf("DECIMAL(%d,%d)", precision, scale)
}

func (g *MysqlDialect) GetStringType(length int) string {
	return fmt.Sprintf("VARCHAR(%d)", length)
}

func (g *MysqlDialect) GetTextType() string {
	return "TEXT"
}

func (g *MysqlDialect) GetBinaryType(length int) string {
	return fmt.Sprintf("BINARY(%d)", length)
}

func (g *MysqlDialect) GetJsonType() string {
	return "JSON"
}

func (g *MysqlDialect) GetTimestampType() string {
	return "TIMESTAMP"
}

func (g *MysqlDialect) GetDateType() string {
	return "DATE"
}

func (g *MysqlDialect) GetTimeType() string {
	return "TIME"
}

func (g *MysqlDialect) GetCurrentTimestampExpression() string {
	return "CURRENT_TIMESTAMP"
}

func (g *MysqlDialect) QuoteIdentifier(name string) string {
	return "`" + strings.Replace(name, "`", "``", -1) + "`"
}

func (g *MysqlDialect) QuoteString(value string) string {
	return "'" + strings.Replace(value, "'", "''", -1) + "'"
}

func (g *MysqlDialect) SupportsColumnPositioning() bool {
	return true
}

func (g *MysqlDialect) SupportsEnum() bool {
	return true
}

func (g *MysqlDialect) GetEnumType(values []string) string {
	return fmt.Sprintf("ENUM('%s')", strings.Join(values, "','"))
}

func (g *MysqlDialect) SupportsColumnComments() bool {
	return true
}

// MySQL поддерживает оба типа индексов
func (g *MysqlDialect) SupportsSpatialIndex() bool {
	return true
}

func (g *MysqlDialect) SupportsFullTextIndex() bool {
	return true
}

// Добавляем методы для создания специальных индексов
func (g *MysqlDialect) BuildSpatialIndexDefinition(name string, columns []string) string {
	return fmt.Sprintf("SPATIAL INDEX %s (%s)", name, strings.Join(columns, ", "))
}

func (g *MysqlDialect) BuildFullTextIndexDefinition(name string, columns []string) string {
	return fmt.Sprintf("FULLTEXT INDEX %s (%s)", name, strings.Join(columns, ", "))
}

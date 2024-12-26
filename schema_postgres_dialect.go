package dblayer

import (
	"fmt"
	"strings"
)

func (g *PostgresSchemaDialect) BuildCreateTable(s *Schema) string {
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
		columns = append(columns, s.buildColumn(col))
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

func (g *PostgresSchemaDialect) BuildAlterTable(s *Schema) string {
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

func (g *PostgresSchemaDialect) BuildDropTable(dt *DropTable) string {
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

	if dt.options.Cascade {
		sql.WriteString(" CASCADE")
	}

	if dt.options.Restrict {
		sql.WriteString(" RESTRICT")
	}

	return sql.String()
}

func (g *PostgresSchemaDialect) BuildColumnDefinition(col Column) string {
	var sql strings.Builder

	sql.WriteString(col.Name)
	sql.WriteString(" ")

	// Особая обработка для PostgreSQL
	if col.Constraints.AutoIncrement {
		sql.WriteString("SERIAL")
		return sql.String()
	}

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

	return sql.String()
}

func (g *PostgresSchemaDialect) BuildIndexDefinition(name string, columns []string, unique bool) string {
	var sql strings.Builder

	if unique {
		sql.WriteString("UNIQUE ")
	}
	sql.WriteString("INDEX ")
	sql.WriteString(g.QuoteIdentifier(name))
	sql.WriteString(" ON ")
	// Имя таблицы будет добавлено позже
	sql.WriteString(" USING btree (")

	// Цитируем каждую колонку
	quotedColumns := make([]string, len(columns))
	for i, col := range columns {
		quotedColumns[i] = g.QuoteIdentifier(col)
	}
	sql.WriteString(strings.Join(quotedColumns, ", "))
	sql.WriteString(")")

	return sql.String()
}

func (g *PostgresSchemaDialect) BuildForeignKeyDefinition(fk *ForeignKey) string {
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

func (g *PostgresSchemaDialect) BuildTruncateTable(tt *TruncateTable) string {
	var sql strings.Builder
	sql.WriteString("TRUNCATE TABLE ")
	sql.WriteString(strings.Join(tt.tables, ", "))

	if tt.options.Restart {
		sql.WriteString(" RESTART IDENTITY")
	}

	if tt.options.Cascade {
		sql.WriteString(" CASCADE")
	}

	return sql.String()
}

func (g *PostgresSchemaDialect) SupportsDropConcurrently() bool {
	return true
}

func (g *PostgresSchemaDialect) SupportsRestartIdentity() bool {
	return true
}

func (g *PostgresSchemaDialect) SupportsCascade() bool {
	return true
}

func (g *PostgresSchemaDialect) SupportsForce() bool {
	return false
}

func (g *PostgresSchemaDialect) GetAutoIncrementType() string {
	return "SERIAL"
}

func (g *PostgresSchemaDialect) GetUUIDType() string {
	return "UUID"
}

func (g *PostgresSchemaDialect) GetBooleanType() string {
	return "BOOLEAN"
}

func (g *PostgresSchemaDialect) GetIntegerType() string {
	return "INTEGER"
}

func (g *PostgresSchemaDialect) GetBigIntegerType() string {
	return "BIGINT"
}

func (g *PostgresSchemaDialect) GetFloatType() string {
	return "REAL"
}

func (g *PostgresSchemaDialect) GetDoubleType() string {
	return "DOUBLE PRECISION"
}

func (g *PostgresSchemaDialect) GetDecimalType(precision, scale int) string {
	return fmt.Sprintf("NUMERIC(%d,%d)", precision, scale)
}

func (g *PostgresSchemaDialect) GetStringType(length int) string {
	return fmt.Sprintf("VARCHAR(%d)", length)
}

func (g *PostgresSchemaDialect) GetTextType() string {
	return "TEXT"
}

func (g *PostgresSchemaDialect) GetBinaryType(length int) string {
	return "BYTEA"
}

func (g *PostgresSchemaDialect) GetJsonType() string {
	return "JSONB"
}

func (g *PostgresSchemaDialect) GetTimestampType() string {
	return "TIMESTAMP"
}

func (g *PostgresSchemaDialect) GetDateType() string {
	return "DATE"
}

func (g *PostgresSchemaDialect) GetTimeType() string {
	return "TIME"
}

func (g *PostgresSchemaDialect) GetCurrentTimestampExpression() string {
	return "CURRENT_TIMESTAMP"
}

func (g *PostgresSchemaDialect) QuoteIdentifier(name string) string {
	return "\"" + strings.Replace(name, "\"", "\"\"", -1) + "\""
}

func (g *PostgresSchemaDialect) QuoteString(value string) string {
	return "'" + strings.Replace(value, "'", "''", -1) + "'"
}

func (g *PostgresSchemaDialect) SupportsColumnPositioning() bool {
	return false
}

func (g *PostgresSchemaDialect) SupportsEnum() bool {
	return true
}

func (g *PostgresSchemaDialect) GetEnumType(values []string) string {
	return "TEXT"
}

func (g *PostgresSchemaDialect) SupportsColumnComments() bool {
	return true
}

func (g *PostgresSchemaDialect) SupportsSpatialIndex() bool {
	return true
}

func (g *PostgresSchemaDialect) SupportsFullTextIndex() bool {
	return true
}

func (g *PostgresSchemaDialect) BuildSpatialIndexDefinition(name string, columns []string) string {
	return fmt.Sprintf("CREATE INDEX %s ON %s USING GIST (%s)",
		name, "%s", strings.Join(columns, ", "))
}

func (g *PostgresSchemaDialect) BuildFullTextIndexDefinition(name string, columns []string) string {
	return fmt.Sprintf("CREATE INDEX %s ON %s USING GIN (to_tsvector('english', %s))",
		name, "%s", strings.Join(columns, " || ' ' || "))
}

package dblayer

import (
	"fmt"
	"strings"
)

func (g *PostgresDialect) BuildCreateTable(s *Schema) string {
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

func (g *PostgresDialect) BuildAlterTable(s *Schema) string {
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

func (g *PostgresDialect) BuildDropTable(dt *DropTable) string {
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

func (g *PostgresDialect) BuildColumnDefinition(col *Column) string {
	var sql strings.Builder

	sql.WriteString(col.Name)
	sql.WriteString(" ")

	// Особая обработка для PostgreSQL
	if col.Constraints.AutoIncrement {
		if col.Constraints.Primary {
			sql.WriteString("SERIAL PRIMARY KEY")
		} else {
			sql.WriteString("SERIAL")
		}
		return sql.String()
	}

	sql.WriteString(col.Definition.Type)

	if col.Definition.Length > 0 {
		sql.WriteString(fmt.Sprintf("(%d)", col.Definition.Length))
	}

	if col.Constraints.NotNull {
		sql.WriteString(" NOT NULL")
	}

	if col.Definition.Default != nil {
		sql.WriteString(fmt.Sprintf(" DEFAULT %v", col.Definition.Default))
	}

	if col.Constraints.Primary {
		sql.WriteString(" PRIMARY KEY")
	}

	if col.Constraints.Unique {
		sql.WriteString(" UNIQUE")
	}

	return sql.String()
}

func (g *PostgresDialect) BuildIndexDefinition(name string, columns []string, unique bool, opts *IndexOptions) string {
	var sql strings.Builder
	if unique {
		sql.WriteString("UNIQUE ")
	}
	sql.WriteString("INDEX ")
	sql.WriteString(g.QuoteIdentifier(name))
	sql.WriteString(" ON ")

	sql.WriteString(" USING btree (")
	quotedColumns := make([]string, len(columns))
	for i, col := range columns {
		quotedColumns[i] = g.QuoteIdentifier(col)
	}
	sql.WriteString(strings.Join(quotedColumns, ", "))
	sql.WriteString(")")

	return sql.String()
}

func (g *PostgresDialect) BuildForeignKeyDefinition(fk *ForeignKey) string {
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

func (g *PostgresDialect) BuildTruncateTable(tt *TruncateTable) string {
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

func (g *PostgresDialect) SupportsDropConcurrently() bool {
	return true
}

func (g *PostgresDialect) SupportsRestartIdentity() bool {
	return true
}

func (g *PostgresDialect) SupportsCascade() bool {
	return true
}

func (g *PostgresDialect) SupportsForce() bool {
	return false
}

func (g *PostgresDialect) GetAutoIncrementType() string {
	return "SERIAL"
}

func (g *PostgresDialect) GetUUIDType() string {
	return "UUID"
}

func (g *PostgresDialect) GetBooleanType() string {
	return "BOOLEAN"
}

func (g *PostgresDialect) GetIntegerType() string {
	return "INTEGER"
}

func (g *PostgresDialect) GetBigIntegerType() string {
	return "BIGINT"
}

func (g *PostgresDialect) GetFloatType() string {
	return "REAL"
}

func (g *PostgresDialect) GetDoubleType() string {
	return "DOUBLE PRECISION"
}

func (g *PostgresDialect) GetDecimalType(precision, scale int) string {
	return fmt.Sprintf("NUMERIC(%d,%d)", precision, scale)
}

func (g *PostgresDialect) GetStringType(length int) string {
	return fmt.Sprintf("VARCHAR(%d)", length)
}

func (g *PostgresDialect) GetTextType() string {
	return "TEXT"
}

func (g *PostgresDialect) GetBinaryType(length int) string {
	return "BYTEA"
}

func (g *PostgresDialect) GetJsonType() string {
	return "JSONB"
}

func (g *PostgresDialect) GetTimestampType() string {
	return "TIMESTAMP"
}

func (g *PostgresDialect) GetDateType() string {
	return "DATE"
}

func (g *PostgresDialect) GetTimeType() string {
	return "TIME"
}

func (g *PostgresDialect) GetCurrentTimestampExpression() string {
	return "CURRENT_TIMESTAMP"
}

func (g *PostgresDialect) QuoteIdentifier(name string) string {
	return "\"" + strings.Replace(name, "\"", "\"\"", -1) + "\""
}

func (g *PostgresDialect) QuoteString(value string) string {
	return "'" + strings.Replace(value, "'", "''", -1) + "'"
}

func (g *PostgresDialect) SupportsColumnPositioning() bool {
	return false
}

func (g *PostgresDialect) SupportsEnum() bool {
	return true
}

func (g *PostgresDialect) GetEnumType(values []string) string {
	return "TEXT"
}

func (g *PostgresDialect) SupportsColumnComments() bool {
	return true
}

func (g *PostgresDialect) SupportsSpatialIndex() bool {
	return true
}

func (g *PostgresDialect) SupportsFullTextIndex() bool {
	return true
}

func (g *PostgresDialect) BuildSpatialIndexDefinition(name string, columns []string) string {
	return fmt.Sprintf("CREATE INDEX %s ON %s USING GIST (%s)",
		name, "%s", strings.Join(columns, ", "))
}

func (g *PostgresDialect) BuildFullTextIndexDefinition(name string, columns []string) string {
	return fmt.Sprintf("CREATE INDEX %s ON %s USING GIN (to_tsvector('english', %s))",
		name, "%s", strings.Join(columns, " || ' ' || "))
}

func (g *PostgresDialect) GetSmallIntegerType() string {
	return "SMALLINT"
}

func (g *PostgresDialect) GetMediumIntegerType() string {
	return "INTEGER"
}

func (g *PostgresDialect) GetTinyIntegerType() string {
	return "SMALLINT"
}

func (g *PostgresDialect) GetMoneyType() string {
	return "MONEY"
}

func (g *PostgresDialect) GetCharType(length int) string {
	return fmt.Sprintf("CHAR(%d)", length)
}

func (g *PostgresDialect) GetMediumTextType() string {
	return "TEXT"
}

func (g *PostgresDialect) GetLongTextType() string {
	return "TEXT"
}

func (g *PostgresDialect) GetSetType(values []string) string {
	return "TEXT[]"
}

func (g *PostgresDialect) GetYearType() string {
	return "SMALLINT"
}

func (g *PostgresDialect) GetPointType() string {
	return "POINT"
}

func (g *PostgresDialect) GetPolygonType() string {
	return "POLYGON"
}

func (g *PostgresDialect) GetGeometryType() string {
	return "GEOMETRY"
}

func (g *PostgresDialect) GetIpType() string {
	return "INET"
}

func (g *PostgresDialect) GetMacAddressType() string {
	return "MACADDR"
}
func (g *PostgresDialect) GetUnsignedType() string {
	return ""
}

func (g *PostgresDialect) CheckColumnExists(table, column string) string {
	return `SELECT COUNT(*) > 0 FROM information_schema.columns 
			WHERE table_schema = 'public' 
			AND table_name = $1 AND column_name = $2`
}

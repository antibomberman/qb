package dialect

import (
	"fmt"
	"github.com/antibomberman/dbl/schema"
	"strings"
)

func (g *PostgresDialect) BuildCreateTable(s *schema.Schema) string {
	var sql strings.Builder

	sql.WriteString("CREATE ")

	sql.WriteString("TABLE ")
	if s.Definition.Options.IfNotExists {
		sql.WriteString("IF NOT EXISTS ")
	}
	sql.WriteString(s.Definition.Name)
	sql.WriteString(" (\n")

	// Колонки
	var columns []string
	for _, col := range s.Definition.Columns {
		columns = append(columns, g.BuildColumnDefinition(col))
	}

	// Первичный ключ
	if len(s.Definition.KeyIndex.PrimaryKey) > 0 {
		columns = append(columns, fmt.Sprintf("PRIMARY KEY (%s)",
			strings.Join(s.Definition.KeyIndex.PrimaryKey, ", ")))
	}

	// Уникальные ключи
	for Name, cols := range s.Definition.KeyIndex.UniqueKeys {
		columns = append(columns, fmt.Sprintf("UNIQUE KEY %s (%s)",
			Name, strings.Join(cols, ", ")))
	}

	// Индексы
	for Name, cols := range s.Definition.KeyIndex.Indexes {
		columns = append(columns, fmt.Sprintf("INDEX %s (%s)",
			Name, strings.Join(cols, ", ")))
	}

	// Внешние ключи
	for col, fk := range s.Definition.KeyIndex.ForeignKeys {
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

func (g *PostgresDialect) BuildAlterTable(s *schema.Schema) string {
	var commands []string
	for _, cmd := range s.Definition.Commands {
		commands = append(commands, fmt.Sprintf("%s %s", cmd.Type, cmd.Name))
	}

	return fmt.Sprintf(
		"ALTER TABLE %s\n%s",
		s.Definition.Name,
		strings.Join(commands, ",\n"),
	)
}

func (g *PostgresDialect) BuildDropTable(dt *schema.DropTable) string {
	var sql strings.Builder

	sql.WriteString("DROP ")
	if dt.Options.Temporary {
		sql.WriteString("TEMPORARY ")
	}
	sql.WriteString("TABLE ")

	if dt.Options.Concurrent {
		sql.WriteString("CONCURRENTLY ")
	}

	if dt.Options.IfExists {
		sql.WriteString("IF EXISTS ")
	}

	sql.WriteString(strings.Join(dt.Tables, ", "))

	if dt.Options.Cascade {
		sql.WriteString(" CASCADE")
	}

	if dt.Options.Restrict {
		sql.WriteString(" RESTRICT")
	}

	return sql.String()
}

func (g *PostgresDialect) BuildColumnDefinition(col *schema.Column) string {
	var sql strings.Builder

	sql.WriteString(col.Name)
	sql.WriteString(" ")

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

	if !col.Constraints.NotNull {
		sql.WriteString(" NOT NULL")
	} else {
		sql.WriteString(" NULL")
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

func (g *PostgresDialect) BuildIndexDefinition(Name string, columns []string, unique bool, opts *IndexOptions) string {
	var sql strings.Builder
	if unique {
		sql.WriteString("UNIQUE ")
	}
	sql.WriteString("INDEX ")
	sql.WriteString(g.QuoteIdentifier(Name))
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

func (g *PostgresDialect) BuildForeignKeyDefinition(fk *schema.Foreign) string {
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

func (g *PostgresDialect) BuildTruncateTable(tt *schema.TruncateTable) string {
	var sql strings.Builder
	sql.WriteString("TRUNCATE TABLE ")
	sql.WriteString(strings.Join(tt.Tables, ", "))

	if tt.Options.Restart {
		sql.WriteString(" RESTART IDENTITY")
	}

	if tt.Options.Cascade {
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

func (g *PostgresDialect) QuoteIdentifier(Name string) string {
	return "\"" + strings.Replace(Name, "\"", "\"\"", -1) + "\""
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

func (g *PostgresDialect) BuildSpatialIndexDefinition(Name string, columns []string) string {
	return fmt.Sprintf("CREATE INDEX %s ON %s USING GIST (%s)",
		Name, "%s", strings.Join(columns, ", "))
}

func (g *PostgresDialect) BuildFullTextIndexDefinition(Name string, columns []string) string {
	return fmt.Sprintf("CREATE INDEX %s ON %s USING GIN (to_tsvector('english', %s))",
		Name, "%s", strings.Join(columns, " || ' ' || "))
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
			AND table_Name = $1 AND column_Name = $2`
}

func (g *PostgresDialect) CheckTableExists(table string) string {
	return `SELECT COUNT(*) > 0 FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_Name = $1`
}

package dialect

import (
	"context"
	"fmt"
	"github.com/antibomberman/dbl/schema"
	"strings"
)

func (d *MysqlDialect) BuildCreateTable(s *schema.Schema) string {
	var sql strings.Builder

	sql.WriteString("CREATE ")

	sql.WriteString("TABLE ")
	if s.Definition.Options.IfNotExists {
		sql.WriteString("IF NOT EXISTS ")
	}
	sql.WriteString(s.Definition.Name)
	sql.WriteString(" (\n")

	// Колонки
	var Columns []string
	for _, col := range s.Definition.Columns {
		Columns = append(Columns, g.BuildColumnDefinition(col))
	}

	// Первичный ключ
	if len(s.Definition.Constraints.PrimaryKey) > 0 {
		Columns = append(Columns, fmt.Sprintf("PRIMARY KEY (%s)",
			strings.Join(s.Definition.Constraints.PrimaryKey, ", ")))
	}

	// Уникальные ключи
	for Name, cols := range s.Definition.Constraints.UniqueKeys {
		Columns = append(Columns, fmt.Sprintf("UNIQUE KEY %s (%s)",
			Name, strings.Join(cols, ", ")))
	}

	// Индексы
	for Name, cols := range s.Definition.Constraints.Indexes {
		Columns = append(Columns, fmt.Sprintf("INDEX %s (%s)",
			Name, strings.Join(cols, ", ")))
	}

	// Внешние ключи
	for col, fk := range s.Definition.Constraints.ForeignKeys {
		constraint := fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s)", col, fk.Table, fk.Column)
		if fk.OnDelete != "" {
			constraint += " ON DELETE " + fk.OnDelete
		}
		if fk.OnUpdate != "" {
			constraint += " ON UPDATE " + fk.OnUpdate
		}
		Columns = append(Columns, constraint)
	}

	sql.WriteString(strings.Join(Columns, ",\n"))
	sql.WriteString("\n)")

	// Опции таблицы
	sql.WriteString(fmt.Sprintf(" ENGINE=%s",
		defaultIfEmpty(s.Definition.Options.Engine, "InnoDB")))
	sql.WriteString(fmt.Sprintf(" DEFAULT CHARSET=%s",
		defaultIfEmpty(s.Definition.Options.Charset, "utf8mb4")))
	sql.WriteString(fmt.Sprintf(" COLLATE=%s",
		defaultIfEmpty(s.Definition.Options.Collate, "utf8mb4_unicode_ci")))
	if s.Definition.Options.Comment != "" {
		sql.WriteString(fmt.Sprintf(" COMMENT='%s'",
			strings.Replace(s.Definition.Options.Comment, "'", "\\'", -1)))
	}

	return sql.String()
}

func defaultIfEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func (d *MysqlDialect) BuildAlterTable(s *schema.Schema) string {
	var commands []string
	for _, cmd := range s.Definition.Commands {
		if cmd.Cmd != "" {
			commands = append(commands, cmd.Cmd)
		} else {
			// Добавляем дополнительные параметры если есть
			cmdStr := fmt.Sprintf("%s %s", cmd.Type, cmd.Name)
			if len(cmd.Columns) > 0 {
				cmdStr += fmt.Sprintf(" (%s)", strings.Join(cmd.Columns, ", "))
			}
			if len(cmd.Options) > 0 {
				for k, v := range cmd.Options {
					cmdStr += fmt.Sprintf(" %s %v", k, v)
				}
			}
			commands = append(commands, cmdStr)
		}
	}

	return fmt.Sprintf(
		"ALTER TABLE %s\n%s",
		s.Definition.Name,
		strings.Join(commands, ",\n"),
	)
}

func (d *MysqlDialect) BuildDropTable(dt *schema.DropTable) string {
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

	sql.WriteString(strings.Join(dt.tables, ", "))

	if dt.Options.Force {
		sql.WriteString(" FORCE")
	}

	return sql.String()
}

func (d *MysqlDialect) BuildColumnDefinition(col *schema.Column) string {
	var sql strings.Builder

	sql.WriteString(col.Name)
	sql.WriteString(" ")
	sql.WriteString(col.Definition.Type)

	if col.Definition.Length > 0 {
		sql.WriteString(fmt.Sprintf("(%d)", col.Definition.Length))
	}

	if col.Constraints.Unsigned {
		sql.WriteString(" UNSIGNED")
	}

	if !col.Constraints.NotNull {
		sql.WriteString(" NOT NULL")

	} else {
		sql.WriteString(" NULL")
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

	if col.Constraints.Primary {
		sql.WriteString(" PRIMARY KEY")
	}

	if col.Constraints.Unique {
		sql.WriteString(" UNIQUE")
	}

	if col.Meta.Comment != "" {
		sql.WriteString(fmt.Sprintf(" COMMENT '%s'", strings.Replace(col.Meta.Comment, "'", "\\'", -1)))
	}

	if col.Definition.Collate != "" {
		sql.WriteString(" COLLATE " + col.Definition.Collate)
	}

	return sql.String()
}

func (d *MysqlDialect) BuildTruncateTable(tt *schema.TruncateTable) string {
	var sql strings.Builder
	sql.WriteString("TRUNCATE TABLE ")
	sql.WriteString(strings.Join(tt.Tables, ", "))

	if tt.Options.Force {
		sql.WriteString(" FORCE")
	}

	return sql.String()
}

type IndexOptions struct {
	Using   string            // BTREE, HASH и т.д.
	Comment string            // Комментарий к индексу
	Visible bool              // Видимость индекса
	Options map[string]string // Дополнительные опции
}

func (d *MysqlDialect) BuildIndexDefinition(Name string, Columns []string, unique bool, opts *IndexOptions) string {
	var sql strings.Builder

	if unique {
		sql.WriteString("UNIQUE ")
	}
	sql.WriteString("INDEX ")
	sql.WriteString(d.QuoteIdentifier(Name))

	if opts != nil && opts.Using != "" {
		sql.WriteString(" USING " + opts.Using)
	}

	sql.WriteString(" (")
	// Поддержка длины индекса для каждой колонки
	quotedColumns := make([]string, len(Columns))
	for i, col := range Columns {
		// Проверяем на наличие длины индекса (format: column(length))
		if parts := strings.Split(col, "("); len(parts) > 1 {
			quotedColumns[i] = d.QuoteIdentifier(parts[0]) + "(" + strings.TrimRight(parts[1], ")")
		} else {
			quotedColumns[i] = d.QuoteIdentifier(col)
		}
	}
	sql.WriteString(strings.Join(quotedColumns, ", "))
	sql.WriteString(")")

	if opts != nil {
		if opts.Comment != "" {
			sql.WriteString(fmt.Sprintf(" COMMENT '%s'",
				strings.Replace(opts.Comment, "'", "\\'", -1)))
		}
		if !opts.Visible {
			sql.WriteString(" INVISIBLE")
		}
		for k, v := range opts.Options {
			sql.WriteString(fmt.Sprintf(" %s %s", k, v))
		}
	}

	return sql.String()
}

func (d *MysqlDialect) BuildForeignKeyDefinition(fk *schema.Foreign) string {
	var sql strings.Builder

	sql.WriteString("REFERENCES ")
	sql.WriteString(d.QuoteIdentifier(fk.Table))
	sql.WriteString("(")
	sql.WriteString(d.QuoteIdentifier(fk.Column))
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

func (d *MysqlDialect) SupportsDropConcurrently() bool {
	return false
}

func (d *MysqlDialect) SupportsRestartIdentity() bool {
	return false
}

func (d *MysqlDialect) SupportsCascade() bool {
	return false
}

func (d *MysqlDialect) SupportsForce() bool {
	return true
}

func (d *MysqlDialect) GetAutoIncrementType() string {
	return "AUTO_INCREMENT"
}

func (d *MysqlDialect) GetUUIDType() string {
	return "CHAR(36)"
}

func (d *MysqlDialect) GetBooleanType() string {
	return "TINYINT(1)"
}

func (d *MysqlDialect) GetIntegerType() string {
	return "INT"
}

func (d *MysqlDialect) GetBigIntegerType() string {
	return "BIGINT"
}

func (d *MysqlDialect) GetFloatType() string {
	return "FLOAT"
}
func (d *MysqlDialect) GetDoubleType() string {
	return "DOUBLE"
}

func (d *MysqlDialect) GetDecimalType(precision, scale int) string {
	return fmt.Sprintf("DECIMAL(%d,%d)", precision, scale)
}

func (d *MysqlDialect) GetStringType(length int) string {
	return fmt.Sprintf("VARCHAR(%d)", length)
}

func (d *MysqlDialect) GetTextType() string {
	return "TEXT"
}
func (d *MysqlDialect) GetBinaryType(length int) string {
	return fmt.Sprintf("BINARY(%d)", length)
}

func (d *MysqlDialect) GetJsonType() string {
	return "JSON"
}

func (d *MysqlDialect) GetTimestampType() string {
	return "TIMESTAMP"
}
func (d *MysqlDialect) GetDateType() string {
	return "DATE"
}
func (d *MysqlDialect) GetTimeType() string {
	return "TIME"
}
func (d *MysqlDialect) GetCurrentTimestampExpression() string {
	return "CURRENT_TIMESTAMP"
}

func (d *MysqlDialect) QuoteIdentifier(Name string) string {
	if len(Name) > 64 {
		// Можно либо обрезать, либо вызвать ошибку
		Name = Name[:64]
	}
	return "`" + strings.Replace(Name, "`", "``", -1) + "`"
}

func (d *MysqlDialect) QuoteString(value string) string {
	return "'" + strings.Replace(value, "'", "''", -1) + "'"
}

func (d *MysqlDialect) SupportsColumnPositioning() bool {
	return true
}

func (d *MysqlDialect) SupportsEnum() bool {
	return true
}
func (d *MysqlDialect) GetEnumType(values []string) string {
	return fmt.Sprintf("ENUM('%s')", strings.Join(values, "','"))
}
func (d *MysqlDialect) SupportsColumnComments() bool {
	return true
}

func (d *MysqlDialect) SupportsSpatialIndex() bool {
	return true
}

func (d *MysqlDialect) SupportsFullTextIndex() bool {
	return true
}

func (d *MysqlDialect) BuildSpatialIndexDefinition(Name string, Columns []string) string {
	return fmt.Sprintf("SPATIAL INDEX %s (%s)", Name, strings.Join(Columns, ", "))
}

func (d *MysqlDialect) BuildFullTextIndexDefinition(Name string, Columns []string) string {
	return fmt.Sprintf("FULLTEXT INDEX %s (%s)", Name, strings.Join(Columns, ", "))
}

func (d *MysqlDialect) GetSmallIntegerType() string {
	return "SMALLINT"
}

func (d *MysqlDialect) GetMediumIntegerType() string {
	return "MEDIUMINT"
}

func (d *MysqlDialect) GetTinyIntegerType() string {
	return "TINYINT"
}

func (d *MysqlDialect) GetMoneyType() string {
	return "DECIMAL(19,4)"
}

func (d *MysqlDialect) GetCharType(length int) string {
	return fmt.Sprintf("CHAR(%d)", length)
}

func (d *MysqlDialect) GetMediumTextType() string {
	return "MEDIUMTEXT"
}

func (d *MysqlDialect) GetLongTextType() string {
	return "LONGTEXT"
}

func (d *MysqlDialect) GetSetType(values []string) string {
	return fmt.Sprintf("SET('%s')", strings.Join(values, "','"))
}

func (d *MysqlDialect) GetYearType() string {
	return "YEAR"
}

func (d *MysqlDialect) GetPointType() string {
	return "POINT"
}

func (d *MysqlDialect) GetPolygonType() string {
	return "POLYGON"
}

func (d *MysqlDialect) GetGeometryType() string {
	return "GEOMETRY"
}

func (d *MysqlDialect) GetIpType() string {
	return "VARCHAR(45)"
}

func (d *MysqlDialect) GetMacAddressType() string {
	return "VARCHAR(17)"
}
func (d *MysqlDialect) GetUnsignedType() string {
	return "UNSIGNED"
}

func (d *MysqlDialect) CheckColumnExists(table, column string) string {
	return `SELECT COUNT(*) > 0 FROM information_schema.Columns 
			WHERE table_schema = DATABASE() 
			AND table_Name = ? AND column_Name = ?`
}

func (d *MysqlDialect) Create(ctx context.Context, data interface{}, fields ...string) (int64, error) {
	return 0, nil
}

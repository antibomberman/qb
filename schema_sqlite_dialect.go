package dblayer

import (
	"fmt"
	"strings"
)

func (g *SqliteSchemaDialect) BuildCreateTable(s *Schema) string {
	var sql strings.Builder

	sql.WriteString("CREATE ")

	sql.WriteString("TABLE ")
	if s.ifNotExists {
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
	if len(s.primaryKey) > 0 {
		columns = append(columns, fmt.Sprintf("PRIMARY KEY (%s)",
			strings.Join(s.primaryKey, ", ")))
	}

	// Уникальные ключи
	for name, cols := range s.uniqueKeys {
		columns = append(columns, fmt.Sprintf("UNIQUE KEY %s (%s)",
			name, strings.Join(cols, ", ")))
	}

	// Индексы
	for name, cols := range s.indexes {
		columns = append(columns, fmt.Sprintf("INDEX %s (%s)",
			name, strings.Join(cols, ", ")))
	}

	// Внешние ключи
	for col, fk := range s.foreignKeys {
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

func (g *SqliteSchemaDialect) BuildAlterTable(s *Schema) string {
	return fmt.Sprintf(
		"ALTER TABLE %s\n%s",
		s.name,
		strings.Join(s.commands, ",\n"),
	)
}

func (g *SqliteSchemaDialect) BuildDropTable(dt *DropTable) string {
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

func (g *SqliteSchemaDialect) BuildColumnDefinition(col Column) string {
	return ""
}

func (g *SqliteSchemaDialect) BuildIndexDefinition(name string, columns []string, unique bool) string {
	return ""
}

func (g *SqliteSchemaDialect) BuildForeignKeyDefinition(fk *ForeignKey) string {
	return ""
}

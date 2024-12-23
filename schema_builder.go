package dblayer

import (
	"fmt"
	"strings"
)

// Column добавляет колонку
func (s *Schema) Column(name string) *ColumnBuilder {
	return &ColumnBuilder{
		schema: s,
		column: Column{Name: name},
	}
}

// Build генерирует SQL запрос
func (s *Schema) _Build() string {
	var sql strings.Builder

	sql.WriteString("CREATE ")
	if s.temporary {
		sql.WriteString("TEMPORARY ")
	}
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

	// Опции таблицы
	if s.dbl.db.DriverName() == "mysql" {
		sql.WriteString(fmt.Sprintf(" ENGINE=%s", s.engine))
		sql.WriteString(fmt.Sprintf(" DEFAULT CHARSET=%s", s.charset))
		sql.WriteString(fmt.Sprintf(" COLLATE=%s", s.collate))
		if s.comment != "" {
			sql.WriteString(fmt.Sprintf(" COMMENT='%s'", s.comment))
		}
	}

	return sql.String()
}

// buildColumn генерирует SQL для колонки
func (s *Schema) buildColumn(col Column) string {
	sql := col.Name + " " + col.Type

	if col.Length > 0 {
		sql += fmt.Sprintf("(%d)", col.Length)
	}

	if !col.Nullable {
		sql += " NOT NULL"
	}

	if col.Default != nil {
		sql += fmt.Sprintf(" DEFAULT %v", col.Default)
	}

	if col.AutoIncrement {
		if s.dbl.db.DriverName() == "mysql" {
			sql += " AUTO_INCREMENT"
		} else if s.dbl.db.DriverName() == "postgres" {
			sql = col.Name + " SERIAL"
		}
	}

	if col.Comment != "" && s.dbl.db.DriverName() == "mysql" {
		sql += fmt.Sprintf(" COMMENT '%s'", col.Comment)
	}

	return sql
}

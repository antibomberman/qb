package dblayer

import (
	"fmt"
	"strings"
)

// Build генерирует SQL запрос
func (s *Schema) BuildAlter() string {
	return fmt.Sprintf(
		"ALTER TABLE %s\n%s",
		s.name,
		strings.Join(s.commands, ",\n"),
	)
}

// buildColumnDefinition генерирует SQL определение колонки
func buildColumnDefinition(col Column) string {
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
		sql += " AUTO_INCREMENT"
	}

	if col.Comment != "" {
		sql += fmt.Sprintf(" COMMENT '%s'", col.Comment)
	}

	return sql
}

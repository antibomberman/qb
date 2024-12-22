package dblayer

import (
	"fmt"
	"strings"
)

// Schema представляет построитель схемы таблицы
type Schema struct {
	table *TableBuilder
}

// BigIncrements добавляет автоинкрементное большое целое
func (s *Schema) BigIncrements(name string) *ColumnBuilder {
	return s.table.Column(name).Type("BIGINT").AutoIncrement().Primary()
}

// String добавляет строковое поле
func (s *Schema) String(name string, length int) *ColumnBuilder {
	return s.table.Column(name).Type("VARCHAR", length)
}

// Enum добавляет поле с перечислением
func (s *Schema) Enum(name string, values []string) *ColumnBuilder {
	return s.table.Column(name).Type(
		fmt.Sprintf("ENUM('%s')", strings.Join(values, "','")),
	)
}

// Timestamp добавляет поле метки времени
func (s *Schema) Timestamp(name string) *ColumnBuilder {
	return s.table.Column(name).Type("TIMESTAMP")
}

// Index добавляет индекс
func (s *Schema) Index(name string, columns ...string) *TableBuilder {
	return s.table.Index(name, columns...)
}

// Comment добавляет комментарий к таблице
func (s *Schema) Comment(comment string) *TableBuilder {
	return s.table.Comment(comment)
}

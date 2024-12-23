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

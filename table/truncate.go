package table

import (
	"fmt"
	"strings"
)

// TruncateOptions опции очистки таблицы
type TruncateOptions struct {
	Cascade          bool
	Restart          bool
	ContinueIdentity bool // только для PostgreSQL
	Restrict         bool
	Force            bool // только для MySQL
}

// TruncateTable очищает таблицу
type TruncateTable struct {
	Table   *Table
	Tables  []string
	Options TruncateOptions
}

// Cascade включает каскадную очистку
func (tt *TruncateTable) Cascade() *TruncateTable {
	tt.Options.Cascade = true
	return tt
}

// RestartIdentity сбрасывает автоинкремент
func (tt *TruncateTable) RestartIdentity() *TruncateTable {
	tt.Options.Restart = true
	return tt
}

// ContinueIdentity продолжает автоинкремент (PostgreSQL)
func (tt *TruncateTable) ContinueIdentity() *TruncateTable {
	tt.Options.ContinueIdentity = true
	return tt
}

// Restrict запрещает очистку при зависимостях
func (tt *TruncateTable) Restrict() *TruncateTable {
	tt.Options.Restrict = true
	return tt
}

// Force принудительная очистка (MySQL)
func (tt *TruncateTable) Force() *TruncateTable {
	tt.Options.Force = true
	return tt
}

// Build генерирует SQL запрос
func (tt *TruncateTable) Build() string {
	var sql strings.Builder

	sql.WriteString("TRUNCATE TABLE ")
	sql.WriteString(strings.Join(tt.Tables, ", "))

	if tt.Table.dialect.SupportsRestartIdentity() {
		if tt.Options.Restart {
			sql.WriteString(" RESTART IDENTITY")
		} else if tt.Options.ContinueIdentity {
			sql.WriteString(" CONTINUE IDENTITY")
		}

		if tt.Options.Cascade {
			sql.WriteString(" CASCADE")
		} else if tt.Options.Restrict {
			sql.WriteString(" RESTRICT")
		}
	}

	if tt.Options.Force && tt.Table.dialect.SupportsForce() {
		sql.WriteString(" FORCE")
	}

	return sql.String()
}

// Execute выполняет очистку таблицы
func (tt *TruncateTable) Execute() error {
	fmt.Println(tt.Build())
	return nil
}

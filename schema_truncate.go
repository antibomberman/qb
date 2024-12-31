package DBL

import "strings"

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
	dbl     *DBL
	tables  []string
	options TruncateOptions
}

// Cascade включает каскадную очистку
func (tt *TruncateTable) Cascade() *TruncateTable {
	tt.options.Cascade = true
	return tt
}

// RestartIdentity сбрасывает автоинкремент
func (tt *TruncateTable) RestartIdentity() *TruncateTable {
	tt.options.Restart = true
	return tt
}

// ContinueIdentity продолжает автоинкремент (PostgreSQL)
func (tt *TruncateTable) ContinueIdentity() *TruncateTable {
	tt.options.ContinueIdentity = true
	return tt
}

// Restrict запрещает очистку при зависимостях
func (tt *TruncateTable) Restrict() *TruncateTable {
	tt.options.Restrict = true
	return tt
}

// Force принудительная очистка (MySQL)
func (tt *TruncateTable) Force() *TruncateTable {
	tt.options.Force = true
	return tt
}

// Build генерирует SQL запрос
func (tt *TruncateTable) Build() string {
	var sql strings.Builder

	sql.WriteString("TRUNCATE TABLE ")
	sql.WriteString(strings.Join(tt.tables, ", "))

	if tt.dbl.dialect.SupportsRestartIdentity() {
		if tt.options.Restart {
			sql.WriteString(" RESTART IDENTITY")
		} else if tt.options.ContinueIdentity {
			sql.WriteString(" CONTINUE IDENTITY")
		}

		if tt.options.Cascade {
			sql.WriteString(" CASCADE")
		} else if tt.options.Restrict {
			sql.WriteString(" RESTRICT")
		}
	}

	if tt.options.Force && tt.dbl.dialect.SupportsForce() {
		sql.WriteString(" FORCE")
	}

	return sql.String()
}

// Execute выполняет очистку таблицы
func (tt *TruncateTable) Execute() error {
	return tt.dbl.Raw(tt.Build()).Exec()
}

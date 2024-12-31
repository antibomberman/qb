package DBL

import (
	"strings"
)

// DropOptions опции удаления таблицы
type DropOptions struct {
	IfExists   bool
	Cascade    bool
	Temporary  bool
	Restrict   bool
	Concurrent bool // только для PostgreSQL
	Force      bool // только для MySQL
}

// DropTable удаляет таблицу
type DropTable struct {
	dbl     *DBL
	tables  []string
	options DropOptions
}

// IfExists добавляет проверку существования
func (dt *DropTable) IfExists() *DropTable {
	dt.options.IfExists = true
	return dt
}

// Cascade включает каскадное удаление
func (dt *DropTable) Cascade() *DropTable {
	dt.options.Cascade = true
	return dt
}

// Temporary указывает на временную таблицу
func (dt *DropTable) Temporary() *DropTable {
	dt.options.Temporary = true
	return dt
}

// Restrict запрещает удаление при зависимостях
func (dt *DropTable) Restrict() *DropTable {
	dt.options.Restrict = true
	return dt
}

// Concurrent включает неблокирующее удаление (PostgreSQL)
func (dt *DropTable) Concurrent() *DropTable {
	dt.options.Concurrent = true
	return dt
}

// Force принудительное удаление (MySQL)
func (dt *DropTable) Force() *DropTable {
	dt.options.Force = true
	return dt
}

// Build генерирует SQL запрос
func (dt *DropTable) Build() string {
	var sql strings.Builder

	sql.WriteString("DROP ")
	if dt.options.Temporary {
		sql.WriteString("TEMPORARY ")
	}
	sql.WriteString("TABLE ")

	if dt.options.Concurrent && dt.dbl.dialect.SupportsDropConcurrently() {
		sql.WriteString("CONCURRENTLY ")
	}

	if dt.options.IfExists {
		sql.WriteString("IF EXISTS ")
	}

	sql.WriteString(strings.Join(dt.tables, ", "))

	if dt.options.Cascade && dt.dbl.dialect.SupportsCascade() {
		sql.WriteString(" CASCADE")
	}

	if dt.options.Restrict && dt.dbl.dialect.SupportsCascade() {
		sql.WriteString(" RESTRICT")
	}

	if dt.options.Force && dt.dbl.dialect.SupportsForce() {
		sql.WriteString(" FORCE")
	}

	return sql.String()
}

// Execute выполняет удаление таблицы
func (dt *DropTable) Execute() error {
	return dt.dbl.Raw(dt.Build()).Exec()
}

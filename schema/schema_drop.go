package schema

import (
	"fmt"
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
	DBL     *DBL
	Tables  []string
	Options DropOptions
}

// IfExists добавляет проверку существования
func (dt *DropTable) IfExists() *DropTable {
	dt.Options.IfExists = true
	return dt
}

// Cascade включает каскадное удаление
func (dt *DropTable) Cascade() *DropTable {
	dt.Options.Cascade = true
	return dt
}

// Temporary указывает на временную таблицу
func (dt *DropTable) Temporary() *DropTable {
	dt.Options.Temporary = true
	return dt
}

// Restrict запрещает удаление при зависимостях
func (dt *DropTable) Restrict() *DropTable {
	dt.Options.Restrict = true
	return dt
}

// Concurrent включает неблокирующее удаление (PostgreSQL)
func (dt *DropTable) Concurrent() *DropTable {
	dt.Options.Concurrent = true
	return dt
}

// Force принудительное удаление (MySQL)
func (dt *DropTable) Force() *DropTable {
	dt.Options.Force = true
	return dt
}

// Build генерирует SQL запрос
func (dt *DropTable) Build() string {
	var sql strings.Builder

	sql.WriteString("DROP ")
	if dt.Options.Temporary {
		sql.WriteString("TEMPORARY ")
	}
	sql.WriteString("TABLE ")

	if dt.Options.Concurrent && dt.DBL.Dialect.SupportsDropConcurrently() {
		sql.WriteString("CONCURRENTLY ")
	}

	if dt.Options.IfExists {
		sql.WriteString("IF EXISTS ")
	}

	sql.WriteString(strings.Join(dt.Tables, ", "))

	if dt.Options.Cascade && dt.DBL.Dialect.SupportsCascade() {
		sql.WriteString(" CASCADE")
	}

	if dt.Options.Restrict && dt.DBL.Dialect.SupportsCascade() {
		sql.WriteString(" RESTRICT")
	}

	if dt.Options.Force && dt.DBL.Dialect.SupportsForce() {
		sql.WriteString(" FORCE")
	}

	return sql.String()
}

// Execute выполняет удаление таблицы
func (dt *DropTable) Execute() error {
	fmt.Println(dt.Build())
	return nil
}

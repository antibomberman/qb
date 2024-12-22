package dblayer

import (
	"fmt"
	"strings"
)

// AlterTable представляет изменение таблицы
type AlterTable struct {
	dbl      *DBLayer
	name     string
	commands []string
}

// AddColumn добавляет колонку
func (at *AlterTable) AddColumn(column Column) *AlterTable {
	position := ""
	if column.After != "" {
		position = fmt.Sprintf(" AFTER %s", column.After)
	} else if column.First {
		position = " FIRST"
	}

	at.commands = append(at.commands, fmt.Sprintf(
		"ADD COLUMN %s%s",
		buildColumnDefinition(column),
		position,
	))
	return at
}

// DropColumn удаляет колонку
func (at *AlterTable) DropColumn(name string) *AlterTable {
	at.commands = append(at.commands, fmt.Sprintf("DROP COLUMN %s", name))
	return at
}

// ModifyColumn изменяет колонку
func (at *AlterTable) ModifyColumn(column Column) *AlterTable {
	at.commands = append(at.commands, fmt.Sprintf(
		"MODIFY COLUMN %s",
		buildColumnDefinition(column),
	))
	return at
}

// RenameColumn переименовывает колонку
func (at *AlterTable) RenameColumn(oldName, newName string) *AlterTable {
	at.commands = append(at.commands, fmt.Sprintf(
		"RENAME COLUMN %s TO %s",
		oldName, newName,
	))
	return at
}

// AddIndex добавляет индекс
func (at *AlterTable) AddIndex(name string, columns []string, unique bool) *AlterTable {
	indexType := "INDEX"
	if unique {
		indexType = "UNIQUE INDEX"
	}
	at.commands = append(at.commands, fmt.Sprintf(
		"ADD %s %s (%s)",
		indexType, name,
		strings.Join(columns, ", "),
	))
	return at
}

// DropIndex удаляет индекс
func (at *AlterTable) DropIndex(name string) *AlterTable {
	at.commands = append(at.commands, fmt.Sprintf("DROP INDEX %s", name))
	return at
}

// AddForeignKey добавляет внешний ключ
func (at *AlterTable) AddForeignKey(name string, column string, reference ForeignKey) *AlterTable {
	cmd := fmt.Sprintf(
		"ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s)",
		name, column, reference.Table, reference.Column,
	)
	if reference.OnDelete != "" {
		cmd += " ON DELETE " + reference.OnDelete
	}
	if reference.OnUpdate != "" {
		cmd += " ON UPDATE " + reference.OnUpdate
	}
	at.commands = append(at.commands, cmd)
	return at
}

// DropForeignKey удаляет внешний ключ
func (at *AlterTable) DropForeignKey(name string) *AlterTable {
	at.commands = append(at.commands, fmt.Sprintf("DROP FOREIGN KEY %s", name))
	return at
}

// RenameTable переименовывает таблицу
func (at *AlterTable) RenameTable(newName string) *AlterTable {
	at.commands = append(at.commands, fmt.Sprintf("RENAME TO %s", newName))
	return at
}

// ChangeEngine меняет движок таблицы
func (at *AlterTable) ChangeEngine(engine string) *AlterTable {
	at.commands = append(at.commands, fmt.Sprintf("ENGINE = %s", engine))
	return at
}

// ChangeCharset меняет кодировку
func (at *AlterTable) ChangeCharset(charset, collate string) *AlterTable {
	at.commands = append(at.commands, fmt.Sprintf(
		"CHARACTER SET = %s COLLATE = %s",
		charset, collate,
	))
	return at
}

// Build генерирует SQL запрос
func (at *AlterTable) Build() string {
	return fmt.Sprintf(
		"ALTER TABLE %s\n%s",
		at.name,
		strings.Join(at.commands, ",\n"),
	)
}

// Execute выполняет изменение таблицы
func (at *AlterTable) Execute() error {
	return at.dbl.Raw(at.Build()).Exec()
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

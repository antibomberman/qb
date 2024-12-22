package dblayer

import (
	"fmt"
	"strings"
)

// Column представляет колонку таблицы
type Column struct {
	Name          string
	Type          string
	Length        int
	Nullable      bool
	Default       interface{}
	AutoIncrement bool
	Primary       bool
	Unique        bool
	Index         bool
	Comment       string
	After         string
	First         bool
	References    *ForeignKey
}

// ForeignKey представляет внешний ключ
type ForeignKey struct {
	Table    string
	Column   string
	OnDelete string
	OnUpdate string
}

// TableBuilder построитель таблиц
type TableBuilder struct {
	dbl         *DBLayer
	name        string
	columns     []Column
	primaryKey  []string
	uniqueKeys  map[string][]string
	indexes     map[string][]string
	foreignKeys map[string]*ForeignKey
	engine      string
	charset     string
	collate     string
	comment     string
	temporary   bool
	ifNotExists bool
}

// Column добавляет колонку
func (tb *TableBuilder) Column(name string) *ColumnBuilder {
	return &ColumnBuilder{
		table:  tb,
		column: Column{Name: name},
	}
}

// PrimaryKey устанавливает первичный ключ
func (tb *TableBuilder) PrimaryKey(columns ...string) *TableBuilder {
	tb.primaryKey = columns
	return tb
}

// UniqueKey добавляет уникальный ключ
func (tb *TableBuilder) UniqueKey(name string, columns ...string) *TableBuilder {
	tb.uniqueKeys[name] = columns
	return tb
}

// Index добавляет индекс
func (tb *TableBuilder) Index(name string, columns ...string) *TableBuilder {
	tb.indexes[name] = columns
	return tb
}

// ForeignKey добавляет внешний ключ
func (tb *TableBuilder) ForeignKey(column, refTable, refColumn string) *ForeignKeyBuilder {
	return &ForeignKeyBuilder{
		table: tb,
		fk: &ForeignKey{
			Table:  refTable,
			Column: refColumn,
		},
		column: column,
	}
}

// Engine устанавливает движок таблицы
func (tb *TableBuilder) Engine(engine string) *TableBuilder {
	tb.engine = engine
	return tb
}

// Charset устанавливает кодировку
func (tb *TableBuilder) Charset(charset string) *TableBuilder {
	tb.charset = charset
	return tb
}

// Collate устанавливает сравнение
func (tb *TableBuilder) Collate(collate string) *TableBuilder {
	tb.collate = collate
	return tb
}

// Comment добавляет комментарий
func (tb *TableBuilder) Comment(comment string) *TableBuilder {
	tb.comment = comment
	return tb
}

// Temporary делает таблицу временной
func (tb *TableBuilder) Temporary() *TableBuilder {
	tb.temporary = true
	return tb
}

// IfNotExists добавляет проверку существования
func (tb *TableBuilder) IfNotExists() *TableBuilder {
	tb.ifNotExists = true
	return tb
}

// Build генерирует SQL запрос
func (tb *TableBuilder) Build() string {
	var sql strings.Builder

	sql.WriteString("CREATE ")
	if tb.temporary {
		sql.WriteString("TEMPORARY ")
	}
	sql.WriteString("TABLE ")
	if tb.ifNotExists {
		sql.WriteString("IF NOT EXISTS ")
	}
	sql.WriteString(tb.name)
	sql.WriteString(" (\n")

	// Колонки
	var columns []string
	for _, col := range tb.columns {
		columns = append(columns, tb.buildColumn(col))
	}

	// Первичный ключ
	if len(tb.primaryKey) > 0 {
		columns = append(columns, fmt.Sprintf("PRIMARY KEY (%s)",
			strings.Join(tb.primaryKey, ", ")))
	}

	// Уникальные ключи
	for name, cols := range tb.uniqueKeys {
		columns = append(columns, fmt.Sprintf("UNIQUE KEY %s (%s)",
			name, strings.Join(cols, ", ")))
	}

	// Индексы
	for name, cols := range tb.indexes {
		columns = append(columns, fmt.Sprintf("INDEX %s (%s)",
			name, strings.Join(cols, ", ")))
	}

	// Внешние ключи
	for col, fk := range tb.foreignKeys {
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
	if tb.dbl.db.DriverName() == "mysql" {
		sql.WriteString(fmt.Sprintf(" ENGINE=%s", tb.engine))
		sql.WriteString(fmt.Sprintf(" DEFAULT CHARSET=%s", tb.charset))
		sql.WriteString(fmt.Sprintf(" COLLATE=%s", tb.collate))
		if tb.comment != "" {
			sql.WriteString(fmt.Sprintf(" COMMENT='%s'", tb.comment))
		}
	}

	return sql.String()
}

// ColumnBuilder построитель колонок
type ColumnBuilder struct {
	table  *TableBuilder
	column Column
}

func (cb *ColumnBuilder) Type(typ string, length ...int) *ColumnBuilder {
	cb.column.Type = typ
	if len(length) > 0 {
		cb.column.Length = length[0]
	}
	return cb
}

func (cb *ColumnBuilder) Nullable() *ColumnBuilder {
	cb.column.Nullable = true
	return cb
}

func (cb *ColumnBuilder) Default(value interface{}) *ColumnBuilder {
	cb.column.Default = value
	return cb
}

func (cb *ColumnBuilder) AutoIncrement() *ColumnBuilder {
	cb.column.AutoIncrement = true
	return cb
}

func (cb *ColumnBuilder) Primary() *ColumnBuilder {
	cb.column.Primary = true
	return cb
}

func (cb *ColumnBuilder) Unique() *ColumnBuilder {
	cb.column.Unique = true
	return cb
}

func (cb *ColumnBuilder) Index() *ColumnBuilder {
	cb.column.Index = true
	return cb
}

func (cb *ColumnBuilder) Comment(comment string) *ColumnBuilder {
	cb.column.Comment = comment
	return cb
}

func (cb *ColumnBuilder) After(column string) *ColumnBuilder {
	cb.column.After = column
	return cb
}

func (cb *ColumnBuilder) First() *ColumnBuilder {
	cb.column.First = true
	return cb
}

func (cb *ColumnBuilder) References(table, column string) *ColumnBuilder {
	cb.column.References = &ForeignKey{
		Table:  table,
		Column: column,
	}
	return cb
}

// Add добавляет колонку в таблицу
func (cb *ColumnBuilder) Add() *TableBuilder {
	cb.table.columns = append(cb.table.columns, cb.column)
	return cb.table
}

// ForeignKeyBuilder построитель внешних ключей
type ForeignKeyBuilder struct {
	table  *TableBuilder
	fk     *ForeignKey
	column string
}

func (fkb *ForeignKeyBuilder) OnDelete(action string) *ForeignKeyBuilder {
	fkb.fk.OnDelete = action
	return fkb
}

func (fkb *ForeignKeyBuilder) OnUpdate(action string) *ForeignKeyBuilder {
	fkb.fk.OnUpdate = action
	return fkb
}

func (fkb *ForeignKeyBuilder) Add() *TableBuilder {
	fkb.table.foreignKeys[fkb.column] = fkb.fk
	return fkb.table
}

// buildColumn генерирует SQL для колонки
func (tb *TableBuilder) buildColumn(col Column) string {
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
		if tb.dbl.db.DriverName() == "mysql" {
			sql += " AUTO_INCREMENT"
		} else if tb.dbl.db.DriverName() == "postgres" {
			sql = col.Name + " SERIAL"
		}
	}

	if col.Comment != "" && tb.dbl.db.DriverName() == "mysql" {
		sql += fmt.Sprintf(" COMMENT '%s'", col.Comment)
	}

	return sql
}

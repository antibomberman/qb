package dblayer

import "fmt"

// ForeignKey представляет внешний ключ
type ForeignKey struct {
	Table    string
	Column   string
	OnDelete string
	OnUpdate string
}

// ForeignKeyBuilder построитель внешних ключей

type ForeignKeyBuilder struct {
	schema *Schema
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

func (fkb *ForeignKeyBuilder) Add() *Schema {
	fkb.schema.definition.constraints.foreignKeys[fkb.column] = fkb.fk
	return fkb.schema
}

// ForeignKey добавляет внешний ключ
func (s *Schema) ForeignKey(column, refTable, refColumn string) *ForeignKeyBuilder {
	return &ForeignKeyBuilder{
		schema: s,
		fk: &ForeignKey{
			Table:  refTable,
			Column: refColumn,
		},
		column: column,
	}
}

// AddForeignKey добавляет внешний ключ
func (s *Schema) AddForeignKey(name string, column string, reference ForeignKey) *Schema {
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
	s.definition.commands = append(s.definition.commands, Command{
		Type: "ADD CONSTRAINT",
		Name: name,
		Cmd:  cmd,
	})
	return s
}

func (cb *ColumnBuilder) References(table, column string) *ColumnBuilder {
	cb.column.Constraints.References = &ForeignKey{
		Table:  table,
		Column: column,
	}
	return cb
}

// DropForeignKey удаляет внешний ключ
func (s *Schema) DropForeignKey(name string) *Schema {
	s.definition.commands = append(s.definition.commands, Command{
		Type: "DROP CONSTRAINT",
		Name: name,
		Cmd:  fmt.Sprintf("DROP FOREIGN KEY %s", name),
	})
	return s
}

const (
	CASCADE     = "CASCADE"
	RESTRICT    = "RESTRICT"
	SET_NULL    = "SET NULL"
	NO_ACTION   = "NO ACTION"
	SET_DEFAULT = "SET DEFAULT"
)

package table

import (
	"fmt"
	"strings"
)

const (
	CASCADE     = "CASCADE"
	RESTRICT    = "RESTRICT"
	SET_NULL    = "SET NULL"
	NO_ACTION   = "NO ACTION"
	SET_DEFAULT = "SET DEFAULT"
)

// Foreign представляет внешний ключ
type Foreign struct {
	Table    string
	Column   string
	OnDelete string
	OnUpdate string
}

// ForeignBuilder построитель внешних ключей

type ForeignBuilder struct {
	schema *Builder
	fk     *Foreign
	column string
}

func (b *Builder) foreign(column, refTable, refColumn string) *ForeignBuilder {
	fk := &Foreign{
		Table:  refTable,
		Column: refColumn,
	}
	b.Definition.KeyIndex.ForeignKeys[column] = fk
	return &ForeignBuilder{
		schema: b,
		fk:     fk,
		column: column,
	}
}
func (b *Builder) Foreign(column string) *ForeignBuilder {
	fk := &Foreign{}
	b.Definition.KeyIndex.ForeignKeys[column] = fk
	return &ForeignBuilder{
		schema: b,
		fk:     fk,
		column: column,
	}
}
func (c *ColumnBuilder) Foreign(name string) *ForeignBuilder {

	refTable := strings.TrimSuffix(name, "_id")
	if !strings.HasSuffix(refTable, "s") {
		refTable += "s"
	}

	return c.Builder.foreign(c.Column.Name, refTable, "id")
}
func (c *ForeignBuilder) References(table string, column string) *ForeignBuilder {
	c.fk.Table = table
	c.fk.Column = column
	return c
}

func (b *Builder) ForeignId(name string) *ForeignBuilder {
	b.BigInteger(name).Unsigned()

	refTable := strings.TrimSuffix(name, "_id")
	if !strings.HasSuffix(refTable, "s") {
		refTable += "s"
	}

	return b.foreign(name, refTable, "id")
}

func (fkb *ForeignBuilder) CascadeOnDelete() *ForeignBuilder {
	fkb.fk.OnDelete = CASCADE
	return fkb
}

func (fkb *ForeignBuilder) CascadeOnUpdate() *ForeignBuilder {
	fkb.fk.OnUpdate = CASCADE
	return fkb
}

func (fkb *ForeignBuilder) RestrictOnDelete() *ForeignBuilder {
	fkb.fk.OnDelete = RESTRICT
	return fkb
}

func (fkb *ForeignBuilder) RestrictOnUpdate() *ForeignBuilder {
	fkb.fk.OnUpdate = RESTRICT
	return fkb
}
func (fkb *ForeignBuilder) NullOnDelete() *ForeignBuilder {
	fkb.fk.OnDelete = SET_NULL
	return fkb
}

func (fkb *ForeignBuilder) NullOnUpdate() *ForeignBuilder {
	fkb.fk.OnUpdate = SET_NULL
	return fkb
}

func (fkb *ForeignBuilder) NoActionOnDelete() *ForeignBuilder {
	fkb.fk.OnDelete = NO_ACTION
	return fkb
}

func (fkb *ForeignBuilder) NoActionOnUpdate() *ForeignBuilder {
	fkb.fk.OnUpdate = NO_ACTION
	return fkb
}

func (fkb *ForeignBuilder) SetDefaultOnDelete() *ForeignBuilder {
	fkb.fk.OnDelete = SET_DEFAULT
	return fkb
}

func (fkb *ForeignBuilder) SetDefaultOnUpdate() *ForeignBuilder {
	fkb.fk.OnUpdate = SET_DEFAULT
	return fkb
}

// DropForeignKey удаляет внешний ключ
func (b *Builder) DropForeignKey(name string) *Builder {
	b.Definition.Commands = append(b.Definition.Commands, &Command{
		Type: "DROP CONSTRAINT",
		Name: name,
		Cmd:  fmt.Sprintf("DROP FOREIGN KEY %s", name),
	})
	return b
}

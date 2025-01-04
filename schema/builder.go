package schema

import (
	"fmt"
)

// Builder с более четкой структурой и инкапсуляцией
type Builder struct {
	Schema     *Schema
	Definition Definition
}

type Definition struct {
	Name     string
	Mode     string // "create" или "update"
	Columns  []*Column
	Commands []*Command
	KeyIndex KeyIndex
	Options  TableOptions
}

type KeyIndex struct {
	PrimaryKey  []string
	UniqueKeys  map[string][]string
	Indexes     map[string][]string
	ForeignKeys map[string]*Foreign
}

// &Command для update
type Command struct {
	Type    string
	Name    string
	Cmd     string
	Columns []string
	Options map[string]interface{}
}
type TableOptions struct {
	Engine      string
	Charset     string
	Collate     string
	Comment     string
	IfNotExists bool
}

func (b *Builder) BuildCreate() string {
	return b.Schema.Dialect.BuildCreateTable(b)
}
func (b *Builder) BuildAlter() string {
	return b.Schema.Dialect.BuildAlterTable(b)
}

// Добавляем методы для обновления
func (b *Builder) RenameColumn(from, to string) *Builder {
	if b.Definition.Mode == "update" {
		b.Definition.Commands = append(b.Definition.Commands, &Command{
			Type: "RENAME COLUMN",
			Name: from,
			Cmd:  fmt.Sprintf("RENAME COLUMN %s TO %s", from, to),
		})
	}
	return b
}

// UniqueIndex добавляет уникальный индекс
func (b *Builder) UniqueIndex(name string, columns ...string) *Builder {
	return b.UniqueKey(name, columns...)
}

// FullText добавляет полнотекстовый индекс
func (b *Builder) FullText(name string, columns ...string) *Builder {
	if b.Schema.DB.DriverName() == "mysql" {
		b.Definition.KeyIndex.Indexes[name] = columns
		return b
	}
	return b
}

// PrimaryKey устанавливает первичный ключ
func (b *Builder) PrimaryKey(columns ...string) *Builder {
	b.Definition.KeyIndex.PrimaryKey = columns
	return b
}

// UniqueKey добавляет уникальный ключ
func (b *Builder) UniqueKey(name string, columns ...string) *Builder {
	b.Definition.KeyIndex.UniqueKeys[name] = columns
	return b
}

// Engine устанавливает движок таблицы
func (b *Builder) Engine(Engine string) *Builder {
	b.Definition.Options.Engine = Engine
	return b
}

// Charset устанавливает кодировку
func (b *Builder) Charset(Charset string) *Builder {
	b.Definition.Options.Charset = Charset
	return b
}

// Collate устанавливает сравнение
func (b *Builder) Collate(Collate string) *Builder {
	b.Definition.Options.Collate = Collate
	return b
}

// Comment добавляет комментарий
func (b *Builder) Comment(Comment string) *Builder {
	b.Definition.Options.Comment = Comment
	return b
}

// IfNotExists добавляет проверку существования
func (b *Builder) IfNotExists() *Builder {
	b.Definition.Options.IfNotExists = true
	return b
}

// DropColumn удаляет колонку
//
//	func (s *Builder) DropColumn(name string) *Builder {
//		s.builder.AddConstraint(Constraint{
//			Type:    "DROP COLUMN",
//			Columns: []string{name},
//		})
//		return s
//	}
//
// DropColumn удаляет колонку
func (b *Builder) DropColumn(name string) *Builder {
	if b.Definition.Mode == "update" {
		b.Definition.Commands = append(b.Definition.Commands, &Command{
			Type: "DROP COLUMN",
			Name: name,
			Cmd:  fmt.Sprintf("DROP COLUMN %s", name),
		})
	}
	return b
}

// AddIndex добавляет индекс
func (b *Builder) AddIndex(name string, columns []string, unique bool) *Builder {
	b.Definition.Commands = append(b.Definition.Commands, &Command{
		Type:    "ADD INDEX",
		Name:    name,
		Columns: columns,
		//Unique:  unique,
	})
	return b
}

// DropIndex удаляет индекс
func (b *Builder) DropIndex(name string) *Builder {
	b.Definition.Commands = append(b.Definition.Commands, &Command{
		Type: "DROP INDEX",
		Name: name,
	})
	return b
}

// RenameTable переименует таблицу
func (b *Builder) RenameTable(newName string) *Builder {
	b.Definition.Commands = append(b.Definition.Commands, &Command{
		Type: "RENAME TABlE",
		Name: newName,
		Cmd:  fmt.Sprintf("RENAME TABLE %s TO %s", b.Definition.Name, newName),
	})
	return b
}

// ChangeEngine меняет движок таблицы
func (b *Builder) ChangeEngine(Engine string) *Builder {
	b.Definition.Options.Engine = Engine
	return b
}

// ChangeCharset меняет кодировку
func (b *Builder) ChangeCharset(Charset, Collate string) *Builder {
	b.Definition.Options.Charset = Charset
	b.Definition.Options.Collate = Collate
	return b
}

// Изменяем метод buildColumn
func (b *Builder) buildColumn(col *Column) string {
	return b.Schema.Dialect.BuildColumnDefinition(col)
}

// Добавляем новые методы для индексов
func (b *Builder) SpatialIndex(name string, columns ...string) *Builder {
	if b.Schema.Dialect.SupportsSpatialIndex() {
		if b.Definition.Mode == "create" {
			b.Definition.Commands = append(b.Definition.Commands, &Command{
				Type:    "ADD SPATIAL INDEX",
				Name:    name,
				Columns: columns,
			})
		} else {
			b.Definition.Commands = append(b.Definition.Commands, &Command{
				Type:    "ADD SPATIAL INDEX",
				Name:    name,
				Columns: columns,
			})
		}
	}
	return b
}

func (b *Builder) FullTextIndex(name string, columns ...string) *Builder {
	if b.Schema.Dialect.SupportsFullTextIndex() {
		if b.Definition.Mode == "create" {
			b.Definition.Commands = append(b.Definition.Commands, &Command{
				Type:    "ADD FULLTEXT INDEX",
				Name:    name,
				Columns: columns,
			})
		} else {
			b.Definition.Commands = append(b.Definition.Commands, &Command{
				Type:    "ADD FULLTEXT INDEX",
				Name:    name,
				Columns: columns,
			})
		}
	}
	return b
}

// Timestamps добавляет поля created_at и updated_at
func (b *Builder) Timestamps() *Builder {
	b.Timestamp("created_at").Nullable().Default(b.Schema.Dialect.GetCurrentTimestampExpression())
	b.Timestamp("updated_at").Nullable().Default(b.Schema.Dialect.GetCurrentTimestampExpression()).OnUpdate(b.Schema.Dialect.GetCurrentTimestampExpression()).Nullable()
	return b
}

// SoftDeletes добавляет поле deleted_at для мягкого удаления
func (b *Builder) SoftDeletes() *Builder {
	b.Timestamp("deleted_at").Nullable()
	return b
}

// Morphs добавляет поля для полиморфных отношений
func (b *Builder) Morphs(name string) *Builder {
	b.BigInteger(name + "_id")
	b.String(name+"_type", 255)
	b.Index(name+"_index", name+"_id", name+"_type")
	return b
}

// NullableMorphs добавляет nullable поля для полиморфных отношений
func (b *Builder) NullableMorphs(name string) *Builder {
	b.BigInteger(name + "_id").Nullable()
	b.String(name+"_type", 255).Nullable()
	b.Index(name+"_index", name+"_id", name+"_type")
	return b
}

// Version добавляет поле для версионирования
func (b *Builder) Version() *Builder {
	b.Integer("version").Default(1)
	return b
}

// ID добавляет поле ID с автоинкрементом и первичным ключом
func (b *Builder) ID() *ColumnBuilder {
	return b.BigInteger("id").Unsigned().AutoIncrement().Primary()
}

package schema

import (
	"fmt"
)

// Schema с более четкой структурой и инкапсуляцией
type Schema struct {
	dbl        *DBL
	Definition SchemaDefinition
	builder    SchemaBuilder
}

type SchemaDefinition struct {
	Name        string
	Mode        string // "create" или "update"
	Columns     []*Column
	Commands    []Command
	Constraints Constraints
	Options     TableOptions
}

type SchemaBuilder interface {
	AddColumn(col Column)
	AddConstraint(constraint Constraint)
	SetOption(key string, value interface{})
	Build() string
}

type Constraints struct {
	PrimaryKey  []string
	UniqueKeys  map[string][]string
	Indexes     map[string][]string
	ForeignKeys map[string]*Foreign
}

type TableOptions struct {
	Engine      string
	Charset     string
	Collate     string
	Comment     string
	IfNotExists bool
}

// Constraint представляет ограничение таблицы
type Constraint struct {
	Type          string
	Name          string
	Columns       []string
	NewName       string
	NewDefinition ColumnDefinition
	Unique        bool
}

// Command представляет команду изменения схемы
type Command struct {
	Type    string
	Name    string
	Cmd     string
	Columns []string
	Options map[string]interface{}
}

func (s *Schema) BuildCreate() string {
	return s.dbl.Dialect.BuildCreateTable(s)
}
func (s *Schema) BuildAlter() string {
	return s.dbl.Dialect.BuildAlterTable(s)
}

// Добавляем методы для обновления
func (s *Schema) RenameColumn(from, to string) *Schema {
	if s.Definition.Mode == "update" {
		s.builder.AddConstraint(Constraint{
			Type:    "RENAME COLUMN",
			Columns: []string{from},
			NewName: to,
		})
	}
	return s
}

// UniqueIndex добавляет уникальный индекс
func (s *Schema) UniqueIndex(name string, columns ...string) *Schema {
	return s.UniqueKey(name, columns...)
}

// FullText добавляет полнотекстовый индекс
func (s *Schema) FullText(name string, columns ...string) *Schema {
	if s.dbl.DB.DriverName() == "mysql" {
		s.Definition.Constraints.Indexes[name] = columns
		return s
	}
	return s
}

// PrimaryKey устанавливает первичный ключ
func (s *Schema) PrimaryKey(columns ...string) *Schema {
	s.Definition.Constraints.PrimaryKey = columns
	return s
}

// UniqueKey добавляет уникальный ключ
func (s *Schema) UniqueKey(name string, columns ...string) *Schema {
	s.Definition.Constraints.UniqueKeys[name] = columns
	return s
}

// Engine устанавливает движок таблицы
func (s *Schema) Engine(Engine string) *Schema {
	s.Definition.Options.Engine = Engine
	return s
}

// Charset устанавливает кодировку
func (s *Schema) Charset(Charset string) *Schema {
	s.Definition.Options.Charset = Charset
	return s
}

// Collate устанавливает сравнение
func (s *Schema) Collate(Collate string) *Schema {
	s.Definition.Options.Collate = Collate
	return s
}

// Comment добавляет комментарий
func (s *Schema) Comment(Comment string) *Schema {
	s.Definition.Options.Comment = Comment
	return s
}

// IfNotExists добавляет проверку существования
func (s *Schema) IfNotExists() *Schema {
	s.Definition.Options.IfNotExists = true
	return s
}

// DropColumn удаляет колонку
//
//	func (s *Schema) DropColumn(name string) *Schema {
//		s.builder.AddConstraint(Constraint{
//			Type:    "DROP COLUMN",
//			Columns: []string{name},
//		})
//		return s
//	}
//
// DropColumn удаляет колонку
func (s *Schema) DropColumn(name string) *Schema {
	if s.Definition.Mode == "update" {
		s.Definition.Commands = append(s.Definition.Commands, Command{
			Type: "DROP COLUMN",
			Name: name,
			Cmd:  fmt.Sprintf("DROP COLUMN %s", name),
		})
	}
	return s
}

// AddIndex добавляет индекс
func (s *Schema) AddIndex(name string, columns []string, unique bool) *Schema {
	s.builder.AddConstraint(Constraint{
		Type:    "ADD INDEX",
		Name:    name,
		Columns: columns,
		Unique:  unique,
	})
	return s
}

// DropIndex удаляет индекс
func (s *Schema) DropIndex(name string) *Schema {
	s.builder.AddConstraint(Constraint{
		Type: "DROP INDEX",
		Name: name,
	})
	return s
}

// RenameTable переименует таблицу
func (s *Schema) RenameTable(newName string) *Schema {
	s.builder.AddConstraint(Constraint{
		Type:    "RENAME TO",
		NewName: newName,
	})
	return s
}

// ChangeEngine меняет движок таблицы
func (s *Schema) ChangeEngine(Engine string) *Schema {
	s.builder.SetOption("Engine", Engine)
	return s
}

// ChangeCharset меняет кодировку
func (s *Schema) ChangeCharset(Charset, Collate string) *Schema {
	s.Definition.Options.Charset = Charset
	s.Definition.Options.Collate = Collate
	return s
}

// Изменяем метод buildColumn
func (s *Schema) buildColumn(col *Column) string {
	return s.dbl.Dialect.BuildColumnDefinition(col)
}

// Добавляем новые методы для индексов
func (s *Schema) SpatialIndex(name string, columns ...string) *Schema {
	if s.dbl.Dialect.SupportsSpatialIndex() {
		if s.Definition.Mode == "create" {
			s.builder.AddConstraint(Constraint{
				Type:    "ADD SPATIAL INDEX",
				Name:    name,
				Columns: columns,
			})
		} else {
			s.builder.AddConstraint(Constraint{
				Type:    "ADD SPATIAL INDEX",
				Name:    name,
				Columns: columns,
			})
		}
	}
	return s
}

func (s *Schema) FullTextIndex(name string, columns ...string) *Schema {
	if s.dbl.Dialect.SupportsFullTextIndex() {
		if s.Definition.Mode == "create" {
			s.builder.AddConstraint(Constraint{
				Type:    "ADD FULLTEXT INDEX",
				Name:    name,
				Columns: columns,
			})
		} else {
			s.builder.AddConstraint(Constraint{
				Type:    "ADD FULLTEXT INDEX",
				Name:    name,
				Columns: columns,
			})
		}
	}
	return s
}

// Timestamps добавляет поля created_at и updated_at
func (s *Schema) Timestamps() *Schema {
	s.Timestamp("created_at").Nullable().Default(s.dbl.Dialect.GetCurrentTimestampExpression())
	s.Timestamp("updated_at").Nullable().Default(s.dbl.Dialect.GetCurrentTimestampExpression()).OnUpdate(s.dbl.Dialect.GetCurrentTimestampExpression()).Nullable()
	return s
}

// SoftDeletes добавляет поле deleted_at для мягкого удаления
func (s *Schema) SoftDeletes() *Schema {
	s.Timestamp("deleted_at").Nullable()
	return s
}

// Morphs добавляет поля для полиморфных отношений
func (s *Schema) Morphs(name string) *Schema {
	s.BigInteger(name + "_id")
	s.String(name+"_type", 255)
	s.Index(name+"_index", name+"_id", name+"_type")
	return s
}

// NullableMorphs добавляет nullable поля для полиморфных отношений
func (s *Schema) NullableMorphs(name string) *Schema {
	s.BigInteger(name + "_id").Nullable()
	s.String(name+"_type", 255).Nullable()
	s.Index(name+"_index", name+"_id", name+"_type")
	return s
}

// Version добавляет поле для версионирования
func (s *Schema) Version() *Schema {
	s.Integer("version").Default(1)
	return s
}

// ID добавляет поле ID с автоинкрементом и первичным ключом
func (s *Schema) ID() *ColumnBuilder {
	return s.BigInteger("id").Unsigned().AutoIncrement().Primary()
}

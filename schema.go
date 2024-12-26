package dblayer

// Schema с более четкой структурой и инкапсуляцией
type Schema struct {
	dbl        *DBLayer
	definition SchemaDefinition
	builder    SchemaBuilder
}

type SchemaDefinition struct {
	name        string
	mode        string // "create" или "update"
	columns     []Column
	commands    []Command
	constraints Constraints
	options     TableOptions
}

type SchemaBuilder interface {
	AddColumn(col Column)
	AddConstraint(constraint Constraint)
	SetOption(key string, value interface{})
	Build() string
}

type Constraints struct {
	primaryKey  []string
	uniqueKeys  map[string][]string
	indexes     map[string][]string
	foreignKeys map[string]*ForeignKey
}

type TableOptions struct {
	engine      string
	charset     string
	collate     string
	comment     string
	ifNotExists bool
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
	return s.dbl.schemaDialect.BuildCreateTable(s)
}
func (s *Schema) BuildAlter() string {
	return s.dbl.schemaDialect.BuildAlterTable(s)
}

// Добавляем методы для обновления
func (s *Schema) RenameColumn(from, to string) *Schema {
	if s.definition.mode == "update" {
		s.builder.AddConstraint(Constraint{
			Type:    "RENAME COLUMN",
			Columns: []string{from},
			NewName: to,
		})
	}
	return s
}

// Timestamps добавляет поля created_at и updated_at
func (s *Schema) Timestamps() *Schema {
	s.Timestamp("created_at").Default("CURRENT_TIMESTAMP")
	s.Timestamp("updated_at").Nullable()
	return s
}

// Morphs добавляет поля для полиморфных отношений
func (s *Schema) Morphs(name string) *Schema {
	s.Integer(name + "_id")
	s.String(name+"_type", 255)
	s.Index("idx_"+name, name+"_id", name+"_type")
	return s
}

// UniqueIndex добавляет уникальный индекс
func (s *Schema) UniqueIndex(name string, columns ...string) *Schema {
	return s.UniqueKey(name, columns...)
}

// FullText добавляет полнотекстовый индекс
func (s *Schema) FullText(name string, columns ...string) *Schema {
	if s.dbl.db.DriverName() == "mysql" {
		s.definition.constraints.indexes[name] = columns
		return s
	}
	return s
}

// Audit добавляет поля аудита
func (s *Schema) Audit() *Schema {
	s.ForeignKey("created_by", "users", "id")
	s.ForeignKey("updated_by", "users", "id")
	s.ForeignKey("deleted_by", "users", "id")
	s.Timestamps()
	s.SoftDeletes()
	return s
}

// PrimaryKey устанавливает первичный ключ
func (s *Schema) PrimaryKey(columns ...string) *Schema {
	s.definition.constraints.primaryKey = columns
	return s
}

// UniqueKey добавляет уникальный ключ
func (s *Schema) UniqueKey(name string, columns ...string) *Schema {
	s.definition.constraints.uniqueKeys[name] = columns
	return s
}

// Engine устанавливает движок таблицы
func (s *Schema) Engine(engine string) *Schema {
	s.definition.options.engine = engine
	return s
}

// Charset устанавливает кодировку
func (s *Schema) Charset(charset string) *Schema {
	s.definition.options.charset = charset
	return s
}

// Collate устанавливает сравнение
func (s *Schema) Collate(collate string) *Schema {
	s.definition.options.collate = collate
	return s
}

// Comment добавляет комментарий
func (s *Schema) Comment(comment string) *Schema {
	s.definition.options.comment = comment
	return s
}

// IfNotExists добавляет проверку существования
func (s *Schema) IfNotExists() *Schema {
	s.definition.options.ifNotExists = true
	return s
}

// DropColumn удаляет колонку
func (s *Schema) DropColumn(name string) *Schema {
	s.builder.AddConstraint(Constraint{
		Type:    "DROP COLUMN",
		Columns: []string{name},
	})
	return s
}

// ModifyColumn изменяет колонку
func (s *Schema) ModifyColumn(column Column) *Schema {
	s.builder.AddConstraint(Constraint{
		Type:          "MODIFY COLUMN",
		Columns:       []string{column.Name},
		NewDefinition: column.Definition,
	})
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

// RenameTable переимен��вывает таблицу
func (s *Schema) RenameTable(newName string) *Schema {
	s.builder.AddConstraint(Constraint{
		Type:    "RENAME TO",
		NewName: newName,
	})
	return s
}

// ChangeEngine меняет движок таблицы
func (s *Schema) ChangeEngine(engine string) *Schema {
	s.builder.SetOption("ENGINE", engine)
	return s
}

// ChangeCharset меняет кодировку
func (s *Schema) ChangeCharset(charset, collate string) *Schema {
	s.definition.options.charset = charset
	s.definition.options.collate = collate
	return s
}

// Изменяем метод buildColumn
func (s *Schema) buildColumn(col Column) string {
	return s.dbl.schemaDialect.BuildColumnDefinition(col)
}

// Изменяем метод Uuid
func (s *Schema) Uuid(name string) *ColumnBuilder {
	return s.addColumn(Column{
		Name: name,
		Definition: ColumnDefinition{
			Type: s.dbl.schemaDialect.GetUUIDType(),
		},
	})
}

// Добавляем новые методы для индексов
func (s *Schema) SpatialIndex(name string, columns ...string) *Schema {
	if s.dbl.schemaDialect.SupportsSpatialIndex() {
		if s.definition.mode == "create" {
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
	if s.dbl.schemaDialect.SupportsFullTextIndex() {
		if s.definition.mode == "create" {
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

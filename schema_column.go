package dblayer

import (
	"fmt"
	"strings"
)

// Column с группировкой связанных полей
type Column struct {
	Name        string
	Definition  ColumnDefinition
	Constraints ColumnConstraints
	Position    ColumnPosition
	Meta        ColumnMeta
}

type ColumnDefinition struct {
	Type     string
	Length   int
	Default  interface{}
	OnUpdate string
	Collate  string
}

type ColumnConstraints struct {
	NotNull       bool
	Unsigned      bool
	AutoIncrement bool
	Primary       bool
	Unique        bool
	Index         bool
	References    *Foreign
}

type ColumnPosition struct {
	After string
	First bool
}

type ColumnMeta struct {
	Comment string
}

// ColumnBuilder построитель колонок
type ColumnBuilder struct {
	schema *Schema
	column *Column
}

// Column добавляет колонку
func (s *Schema) Column(name string) *ColumnBuilder {
	return &ColumnBuilder{
		schema: s,
		column: &Column{Name: name},
	}
}

func (s *Schema) addColumn(col *Column) *ColumnBuilder {
	if s.definition.mode == "create" {
		s.definition.columns = append(s.definition.columns, col)
	} else {
		var exists bool
		query := s.dbl.dialect.CheckColumnExists(s.definition.name, col.Name)
		err := s.dbl.db.QueryRow(query, s.definition.name, col.Name).Scan(&exists)
		if err != nil {
			// Обработка ошибки
			return &ColumnBuilder{schema: s, column: col}
		}

		cmd := ""
		if exists {
			cmd = fmt.Sprintf("MODIFY COLUMN %s", s.dbl.dialect.BuildColumnDefinition(col))
		} else {
			cmd = fmt.Sprintf("ADD COLUMN %s", s.dbl.dialect.BuildColumnDefinition(col))
		}

		s.definition.commands = append(s.definition.commands, Command{
			Type: cmd,
			Name: col.Name,
			Cmd:  cmd,
		})
	}
	return &ColumnBuilder{schema: s, column: col}
}

// AddColumn добавляет колонку
func (s *Schema) AddColumn(column *Column) *ColumnBuilder {
	position := ""
	if column.Position.After != "" {
		position = fmt.Sprintf(" AFTER %s", column.Position.After)
	} else if column.Position.First {
		position = " FIRST"
	}

	s.definition.commands = append(s.definition.commands, Command{
		Type: "ADD COLUMN",
		Name: column.Name,
		Cmd: fmt.Sprintf(
			"%s%s",
			s.dbl.dialect.BuildColumnDefinition(column),
			position,
		),
	})
	return &ColumnBuilder{schema: s, column: column}
}

// BASE TYPES
// String добавляет строковое поле
func (s *Schema) String(name string, length int) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       name,
		Definition: ColumnDefinition{Type: "VARCHAR", Length: length},
	})
}

// Enum добавляет поле с перечислением
func (s *Schema) Enum(name string, values []string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       name,
		Definition: ColumnDefinition{Type: fmt.Sprintf("ENUM('%s')", strings.Join(values, "','"))},
	})
}

// Timestamp добавляет поле метки времени
func (s *Schema) Timestamp(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       name,
		Definition: ColumnDefinition{Type: "TIMESTAMP"},
	})
}

// Index добавляет индекс
func (s *Schema) Index(name string, columns ...string) *Schema {
	s.definition.constraints.indexes[name] = columns
	return s
}

// TinyInteger добавляет малое целое
func (s *Schema) TinyInteger(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       name,
		Definition: ColumnDefinition{Type: "TINYINT"},
	})
}

// SmallInteger добавляет малое целое
func (s *Schema) SmallInteger(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       name,
		Definition: ColumnDefinition{Type: s.dbl.dialect.GetSmallIntegerType()},
	})
}

// MediumInteger добавляет среднее целое
func (s *Schema) MediumInteger(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       name,
		Definition: ColumnDefinition{Type: s.dbl.dialect.GetMediumIntegerType()},
	})
}

// Integer добавляет целочисленное поле
func (s *Schema) Integer(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       name,
		Definition: ColumnDefinition{Type: "INT"},
	})
}

// BigInteger добавляет большое целое
func (s *Schema) BigInteger(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       name,
		Definition: ColumnDefinition{Type: s.dbl.dialect.GetBigIntegerType()},
	})
}

// Boolean добавляет логическое поле
func (s *Schema) Boolean(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       name,
		Definition: ColumnDefinition{Type: s.dbl.dialect.GetBooleanType()},
	})
}

// Text добавляет текстовое поле
func (s *Schema) Text(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       name,
		Definition: ColumnDefinition{Type: "TEXT"},
	})
}

// Date добавляет поле даты
func (s *Schema) Date(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       name,
		Definition: ColumnDefinition{Type: "DATE"},
	})
}

// DateTime добавляет поле даты и времени
func (s *Schema) DateTime(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       name,
		Definition: ColumnDefinition{Type: "DATETIME"},
	})
}

// Decimal добавляет десятичное поле
func (s *Schema) Decimal(name string, precision, scale int) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       name,
		Definition: ColumnDefinition{Type: fmt.Sprintf("DECIMAL(%d,%d)", precision, scale)},
	})
}

// Json добавляет JSON поле
func (s *Schema) Json(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       name,
		Definition: ColumnDefinition{Type: "JSON"},
	})
}

// Binary добавляет бинарное поле
func (s *Schema) Binary(name string, length int) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       name,
		Definition: ColumnDefinition{Type: "BINARY", Length: length},
	})
}

// Float добавляет поле с плавающей точкой
func (s *Schema) Float(name string, precision, scale int) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       name,
		Definition: ColumnDefinition{Type: fmt.Sprintf("FLOAT(%d,%d)", precision, scale)},
	})
}

// MediumText добавляет поле MEDIUMTEXT
func (s *Schema) MediumText(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: name,
		Definition: ColumnDefinition{
			Type: s.dbl.dialect.GetMediumTextType(),
		},
	})
}

// LongText добавляет поле LONGTEXT
func (s *Schema) LongText(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: name,
		Definition: ColumnDefinition{
			Type: s.dbl.dialect.GetLongTextType(),
		},
	})
}

// Char добавляет поле фиксированной длины
func (s *Schema) Char(name string, length int) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       name,
		Definition: ColumnDefinition{Type: "CHAR", Length: length},
	})
}

// ForeignKey добавляет внешний ключ
//func (s *Schema) ForeignKey(name string, table string, column string) *ColumnBuilder {
//	return s.addColumn(&Column{Name: name, Type: "BIGINT", References: &ForeignKey{Table: table, Column: column}})
//}

// ADDITIONAL TYPES
// Money добавляет денежное поле
func (s *Schema) Money(name string) *ColumnBuilder {
	return s.Decimal(name, 19, 4)
}

// Price добавляет поле цены
func (s *Schema) Price(name string) *ColumnBuilder {
	return s.Decimal(name, 10, 2)
}

// Percentage добавляет поле процентов
func (s *Schema) Percentage(name string) *ColumnBuilder {
	return s.Decimal(name, 5, 2)
}

// Status добавляет поле статуса
func (s *Schema) Status(name string, statuses []string) *ColumnBuilder {
	return s.Enum(name, statuses).Default(statuses[0])
}

// Slug добавляет поле для URL-совместимой строки
func (s *Schema) Slug(name string) *ColumnBuilder {
	return s.String(name, 255).Unique()
}

// Phone добавляет поле телефона
func (s *Schema) Phone(name string) *ColumnBuilder {
	return s.String(name, 20)
}

// Color добавляет поле цвета (HEX)
func (s *Schema) Color(name string) *ColumnBuilder {
	return s.Char(name, 7)
}

// Language добавляет поле языка
func (s *Schema) Language() *ColumnBuilder {
	return s.Char("lang", 2)
}

// Country добавляет поле страны
func (s *Schema) Country() *ColumnBuilder {
	return s.Char("country", 2)
}

// Currency добавляет поле валюты
func (s *Schema) Currency() *ColumnBuilder {
	return s.Char("currency", 3)
}

// Timezone добавляет поле часового пояса
func (s *Schema) Timezone() *ColumnBuilder {
	return s.String("timezone", 64)
}

//func (s *Schema) ID(name string) *ColumnBuilder {
//	schema.BigInteger("id").Unsigned().AutoIncrement().Primary()
//	return s.addColumn(&Column{
//		Name:       name,
//		Definition: ColumnDefinition{Type: s.dbl.dialect.GetBigIntegerType()},
//	})
//}

// Year добавляет поле года
func (s *Schema) Year(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: name,
		Definition: ColumnDefinition{
			Type: s.dbl.dialect.GetYearType(),
		},
	})
}

// Time добавляет поле времени
func (s *Schema) Time(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       name,
		Definition: ColumnDefinition{Type: "TIME"},
	})
}

// Ip добавляет поле для IP-адреса
func (s *Schema) Ip(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: name,
		Definition: ColumnDefinition{
			Type: s.dbl.dialect.GetIpType(),
		},
	})
}

// MacAddress добавляет поле для MAC-адреса
func (s *Schema) MacAddress(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: name,
		Definition: ColumnDefinition{
			Type: s.dbl.dialect.GetMacAddressType(),
		},
	})
}

// Point добавляет геометрическое поле точки
func (s *Schema) Point(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: name,
		Definition: ColumnDefinition{
			Type: s.dbl.dialect.GetPointType(),
		},
	})
}

// Polygon добавляет геометрическое поле полигона
func (s *Schema) Polygon(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: name,
		Definition: ColumnDefinition{
			Type: s.dbl.dialect.GetPolygonType(),
		},
	})
}

// Set добавляет поле SET
func (s *Schema) Set(name string, values []string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: name,
		Definition: ColumnDefinition{
			Type: s.dbl.dialect.GetSetType(values),
		},
	})
}

// RememberToken добавляет поле для токена remember_token
func (s *Schema) RememberToken() *ColumnBuilder {
	return s.String("remember_token", 100).Nullable()
}

func (cb *ColumnBuilder) Type(typ string, length ...int) *ColumnBuilder {
	cb.column.Definition.Type = typ
	if len(length) > 0 {
		cb.column.Definition.Length = length[0]
	}
	return cb
}

func (cb *ColumnBuilder) Nullable() *ColumnBuilder {
	cb.column.Constraints.NotNull = false
	return cb
}

func (cb *ColumnBuilder) Default(value interface{}) *ColumnBuilder {
	cb.column.Definition.Default = value
	return cb
}

func (cb *ColumnBuilder) AutoIncrement() *ColumnBuilder {
	cb.column.Constraints.AutoIncrement = true
	return cb
}
func (cb *ColumnBuilder) Unsigned() *ColumnBuilder {
	cb.column.Constraints.Unsigned = true
	return cb
}

func (cb *ColumnBuilder) Primary() *ColumnBuilder {
	cb.column.Constraints.Primary = true
	return cb
}

func (cb *ColumnBuilder) Unique() *ColumnBuilder {
	cb.column.Constraints.Unique = true
	return cb
}

func (cb *ColumnBuilder) Index() *ColumnBuilder {
	cb.column.Constraints.Index = true
	return cb
}

func (cb *ColumnBuilder) Comment(comment string) *ColumnBuilder {
	cb.column.Meta.Comment = comment
	return cb
}

func (cb *ColumnBuilder) After(column string) *ColumnBuilder {
	cb.column.Position.After = column
	return cb
}

func (cb *ColumnBuilder) First() *ColumnBuilder {
	cb.column.Position.First = true
	return cb
}

// OnUpdate добавляет условие ON UPDATE
func (cb *ColumnBuilder) OnUpdate(value string) *ColumnBuilder {
	cb.column.Definition.OnUpdate = value
	return cb
}

// NotNull устанавливает колонку как NOT NULL
func (cb *ColumnBuilder) NotNull() *ColumnBuilder {
	cb.column.Constraints.NotNull = true
	return cb
}

// Geometry добавляет геометрическое поле
func (s *Schema) Geometry(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: name,
		Definition: ColumnDefinition{
			Type: s.dbl.dialect.GetGeometryType(),
		},
	})
}

// UUID добавляет поле UUID
func (s *Schema) UUID(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: name,
		Definition: ColumnDefinition{
			Type: s.dbl.dialect.GetUUIDType(),
		},
	})
}

// Double доба��ляет поле с двойной точностью
func (s *Schema) Double(name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: name,
		Definition: ColumnDefinition{
			Type: s.dbl.dialect.GetDoubleType(),
		},
	})
}

// Password добавляет поле для хранения хэша пароля
func (s *Schema) Password(name string) *ColumnBuilder {
	return s.String(name, 60) // Достаточно для bcrypt
}

// Email добавляет поле email
func (s *Schema) Email(name string) *ColumnBuilder {
	return s.String(name, 255)
}

// Url добавляет поле URL
func (s *Schema) Url(name string) *ColumnBuilder {
	return s.String(name, 2083) // Максимальная длина URL в IE
}

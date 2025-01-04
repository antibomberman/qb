package schema

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
	Schema *Schema
	Column *Column
}

// Column добавляет колонку
func (s *Schema) Column(Name string) *ColumnBuilder {
	return &ColumnBuilder{
		Schema: s,
		Column: &Column{Name: Name},
	}
}

func (s *Schema) addColumn(col *Column) *ColumnBuilder {
	if s.Definition.Mode == "create" {
		s.Definition.Columns = append(s.Definition.Columns, col)
	} else {
		var exists bool
		query := s.dbl.Dialect.CheckColumnExists(s.Definition.Name, col.Name)
		err := s.dbl.DB.QueryRow(query, s.Definition.Name, col.Name).Scan(&exists)
		if err != nil {
			// Обработка ошибки
			return &ColumnBuilder{Schema: s, Column: col}
		}

		cmd := ""
		if exists {
			cmd = fmt.Sprintf("MODIFY Column %s", s.dbl.Dialect.BuildColumnDefinition(col))
		} else {
			cmd = fmt.Sprintf("ADD Column %s", s.dbl.Dialect.BuildColumnDefinition(col))
		}

		s.Definition.Commands = append(s.Definition.Commands, &Command{
			Type: cmd,
			Name: col.Name,
			Cmd:  cmd,
		})
	}
	return &ColumnBuilder{Schema: s, Column: col}
}

// AddColumn добавляет колонку
func (s *Schema) AddColumn(Column *Column) *ColumnBuilder {
	position := ""
	if Column.Position.After != "" {
		position = fmt.Sprintf(" AFTER %s", Column.Position.After)
	} else if Column.Position.First {
		position = " FIRST"
	}

	s.Definition.Commands = append(s.Definition.Commands, &Command{
		Type: "ADD Column",
		Name: Column.Name,
		Cmd: fmt.Sprintf(
			"%s%s",
			s.dbl.Dialect.BuildColumnDefinition(Column),
			position,
		),
	})
	return &ColumnBuilder{Schema: s, Column: Column}
}

// BASE TYPES
// String добавляет строковое поле
func (s *Schema) String(Name string, length int) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "VARCHAR", Length: length},
	})
}

// Enum добавляет поле с перечислением
func (s *Schema) Enum(Name string, values []string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: fmt.Sprintf("ENUM('%s')", strings.Join(values, "','"))},
	})
}

// Timestamp добавляет поле метки времени
func (s *Schema) Timestamp(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "TIMESTAMP"},
	})
}

// Index добавляет индекс
func (s *Schema) Index(Name string, Columns ...string) *Schema {
	s.Definition.KeyIndex.Indexes[Name] = Columns
	return s
}

// TinyInteger добавляет малое целое
func (s *Schema) TinyInteger(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "TINYINT"},
	})
}

// SmallInteger добавляет малое целое
func (s *Schema) SmallInteger(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: s.dbl.Dialect.GetSmallIntegerType()},
	})
}

// MediumInteger добавляет среднее целое
func (s *Schema) MediumInteger(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: s.dbl.Dialect.GetMediumIntegerType()},
	})
}

// Integer добавляет целочисленное поле
func (s *Schema) Integer(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "INT"},
	})
}

// BigInteger добавляет большое целое
func (s *Schema) BigInteger(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: s.dbl.Dialect.GetBigIntegerType()},
	})
}

// Boolean добавляет логическое поле
func (s *Schema) Boolean(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: s.dbl.Dialect.GetBooleanType()},
	})
}

// Text добавляет текстовое поле
func (s *Schema) Text(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "TEXT"},
	})
}

// Date добавляет поле даты
func (s *Schema) Date(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "DATE"},
	})
}

// DateTime добавляет поле даты и времени
func (s *Schema) DateTime(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "DATETIME"},
	})
}

// Decimal добавляет десятичное поле
func (s *Schema) Decimal(Name string, precision, scale int) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: fmt.Sprintf("DECIMAL(%d,%d)", precision, scale)},
	})
}

// Json добавляет JSON поле
func (s *Schema) Json(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "JSON"},
	})
}

// Binary добавляет бинарное поле
func (s *Schema) Binary(Name string, length int) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "BINARY", Length: length},
	})
}

// Float добавляет поле с плавающей точкой
func (s *Schema) Float(Name string, precision, scale int) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: fmt.Sprintf("FLOAT(%d,%d)", precision, scale)},
	})
}

// MediumText добавляет поле MEDIUMTEXT
func (s *Schema) MediumText(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: s.dbl.Dialect.GetMediumTextType(),
		},
	})
}

// LongText добавляет поле LONGTEXT
func (s *Schema) LongText(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: s.dbl.Dialect.GetLongTextType(),
		},
	})
}

// Char добавляет поле фиксированной длины
func (s *Schema) Char(Name string, length int) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "CHAR", Length: length},
	})
}

// ForeignKey добавляет внешний ключ
//func (s *Schema) ForeignKey(Name string, table string, Column string) *ColumnBuilder {
//	return s.addColumn(&Column{Name: Name, Type: "BIGINT", References: &ForeignKey{Table: table, Column: Column}})
//}

// ADDITIONAL TYPES
// Money добавляет денежное поле
func (s *Schema) Money(Name string) *ColumnBuilder {
	return s.Decimal(Name, 19, 4)
}

// Price добавляет поле цены
func (s *Schema) Price(Name string) *ColumnBuilder {
	return s.Decimal(Name, 10, 2)
}

// Percentage добавляет поле процентов
func (s *Schema) Percentage(Name string) *ColumnBuilder {
	return s.Decimal(Name, 5, 2)
}

// Status добавляет поле статуса
func (s *Schema) Status(Name string, statuses []string) *ColumnBuilder {
	return s.Enum(Name, statuses).Default(statuses[0])
}

// Slug добавляет поле для URL-совместимой строки
func (s *Schema) Slug(Name string) *ColumnBuilder {
	return s.String(Name, 255).Unique()
}

// Phone добавляет поле телефона
func (s *Schema) Phone(Name string) *ColumnBuilder {
	return s.String(Name, 20)
}

// Color добавляет поле цвета (HEX)
func (s *Schema) Color(Name string) *ColumnBuilder {
	return s.Char(Name, 7)
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

//func (s *Schema) ID(Name string) *ColumnBuilder {
//	Schema.BigInteger("id").Unsigned().AutoIncrement().Primary()
//	return s.addColumn(&Column{
//		Name:       Name,
//		Definition: ColumnDefinition{Type: s.dbl.Dialect.GetBigIntegerType()},
//	})
//}

// Year добавляет поле года
func (s *Schema) Year(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: s.dbl.Dialect.GetYearType(),
		},
	})
}

// Time добавляет поле времени
func (s *Schema) Time(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "TIME"},
	})
}

// Ip добавляет поле для IP-адреса
func (s *Schema) Ip(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: s.dbl.Dialect.GetIpType(),
		},
	})
}

// MacAddress добавляет поле для MAC-адреса
func (s *Schema) MacAddress(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: s.dbl.Dialect.GetMacAddressType(),
		},
	})
}

// Point добавляет геометрическое поле точки
func (s *Schema) Point(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: s.dbl.Dialect.GetPointType(),
		},
	})
}

// Polygon добавляет геометрическое поле полигона
func (s *Schema) Polygon(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: s.dbl.Dialect.GetPolygonType(),
		},
	})
}

// Set добавляет поле SET
func (s *Schema) Set(Name string, values []string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: s.dbl.Dialect.GetSetType(values),
		},
	})
}

// RememberToken добавляет поле для токена remember_token
func (s *Schema) RememberToken() *ColumnBuilder {
	return s.String("remember_token", 100).Nullable()
}

func (cb *ColumnBuilder) Type(typ string, length ...int) *ColumnBuilder {
	cb.Column.Definition.Type = typ
	if len(length) > 0 {
		cb.Column.Definition.Length = length[0]
	}
	return cb
}

func (cb *ColumnBuilder) Default(value interface{}) *ColumnBuilder {
	cb.Column.Definition.Default = value
	return cb
}

func (cb *ColumnBuilder) AutoIncrement() *ColumnBuilder {
	cb.Column.Constraints.AutoIncrement = true
	return cb
}
func (cb *ColumnBuilder) Unsigned() *ColumnBuilder {
	cb.Column.Constraints.Unsigned = true
	return cb
}

func (cb *ColumnBuilder) Primary() *ColumnBuilder {
	cb.Column.Constraints.Primary = true
	return cb
}

func (cb *ColumnBuilder) Unique() *ColumnBuilder {
	cb.Column.Constraints.Unique = true
	return cb
}

func (cb *ColumnBuilder) Index() *ColumnBuilder {
	cb.Column.Constraints.Index = true
	return cb
}

func (cb *ColumnBuilder) Comment(comment string) *ColumnBuilder {
	cb.Column.Meta.Comment = comment
	return cb
}

func (cb *ColumnBuilder) After(Column string) *ColumnBuilder {
	cb.Column.Position.After = Column
	return cb
}

func (cb *ColumnBuilder) First() *ColumnBuilder {
	cb.Column.Position.First = true
	return cb
}

// OnUpdate добавляет условие ON UPDATE
func (cb *ColumnBuilder) OnUpdate(value string) *ColumnBuilder {
	cb.Column.Definition.OnUpdate = value
	return cb
}

// NotNull устанавливает колонку как NOT NULL
func (cb *ColumnBuilder) NotNull() *ColumnBuilder {
	cb.Column.Constraints.NotNull = true
	return cb
}
func (cb *ColumnBuilder) Nullable() *ColumnBuilder {
	cb.Column.Constraints.NotNull = false
	return cb
}

// Geometry добавляет геометрическое поле
func (s *Schema) Geometry(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: s.dbl.Dialect.GetGeometryType(),
		},
	})
}

// UUID добавляет поле UUID
func (s *Schema) UUID(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: s.dbl.Dialect.GetUUIDType(),
		},
	})
}

// Double доба��ляет поле с двойной точностью
func (s *Schema) Double(Name string) *ColumnBuilder {
	return s.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: s.dbl.Dialect.GetDoubleType(),
		},
	})
}

// Password добавляет поле для хранения хэша пароля
func (s *Schema) Password(Name string) *ColumnBuilder {
	return s.String(Name, 60) // Достаточно для bcrypt
}

// Email добавляет поле email
func (s *Schema) Email(Name string) *ColumnBuilder {
	return s.String(Name, 255)
}

// Url добавляет поле URL
func (s *Schema) Url(Name string) *ColumnBuilder {
	return s.String(Name, 2083) // Максимальная длина URL в IE
}

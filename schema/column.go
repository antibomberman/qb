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
	Builder *Builder
	Column  *Column
}

// Column добавляет колонку
func (b *Builder) Column(Name string) *ColumnBuilder {
	return &ColumnBuilder{
		Builder: b,
		Column:  &Column{Name: Name},
	}
}

func (b *Builder) addColumn(col *Column) *ColumnBuilder {
	if b.Definition.Mode == "create" {
		b.Definition.Columns = append(b.Definition.Columns, col)
	} else {
		var exists bool
		query := b.SchemaBuilder.Dialect.CheckColumnExists(b.Definition.Name, col.Name)
		err := b.SchemaBuilder.DB.QueryRow(query, b.Definition.Name, col.Name).Scan(&exists)
		if err != nil {
			// Обработка ошибки
			return &ColumnBuilder{Builder: b, Column: col}
		}

		cmd := ""
		if exists {
			cmd = fmt.Sprintf("MODIFY Column %s", b.SchemaBuilder.Dialect.BuildColumnDefinition(col))
		} else {
			cmd = fmt.Sprintf("ADD Column %s", b.SchemaBuilder.Dialect.BuildColumnDefinition(col))
		}

		b.Definition.Commands = append(b.Definition.Commands, &Command{
			Type: cmd,
			Name: col.Name,
			Cmd:  cmd,
		})
	}
	return &ColumnBuilder{Builder: b, Column: col}
}

// AddColumn добавляет колонку
func (b *Builder) AddColumn(Column *Column) *ColumnBuilder {
	position := ""
	if Column.Position.After != "" {
		position = fmt.Sprintf(" AFTER %s", Column.Position.After)
	} else if Column.Position.First {
		position = " FIRST"
	}

	b.Definition.Commands = append(b.Definition.Commands, &Command{
		Type: "ADD Column",
		Name: Column.Name,
		Cmd: fmt.Sprintf(
			"%s%s",
			b.SchemaBuilder.Dialect.BuildColumnDefinition(Column),
			position,
		),
	})
	return &ColumnBuilder{Builder: b, Column: Column}
}

// BASE TYPES
// String добавляет строковое поле
func (b *Builder) String(Name string, length int) *ColumnBuilder {
	return b.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "VARCHAR", Length: length},
	})
}

// Enum добавляет поле с перечислением
func (b *Builder) Enum(Name string, values []string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: fmt.Sprintf("ENUM('%s')", strings.Join(values, "','"))},
	})
}

// Timestamp добавляет поле метки времени
func (b *Builder) Timestamp(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "TIMESTAMP"},
	})
}

// Index добавляет индекс
func (b *Builder) Index(Name string, Columns ...string) *Builder {
	b.Definition.KeyIndex.Indexes[Name] = Columns
	return b
}

// TinyInteger добавляет малое целое
func (b *Builder) TinyInteger(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "TINYINT"},
	})
}

// SmallInteger добавляет малое целое
func (b *Builder) SmallInteger(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: b.SchemaBuilder.Dialect.GetSmallIntegerType()},
	})
}

// MediumInteger добавляет среднее целое
func (b *Builder) MediumInteger(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: b.SchemaBuilder.Dialect.GetMediumIntegerType()},
	})
}

// Integer добавляет целочисленное поле
func (b *Builder) Integer(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "INT"},
	})
}

// BigInteger добавляет большое целое
func (b *Builder) BigInteger(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: b.SchemaBuilder.Dialect.GetBigIntegerType()},
	})
}

// Boolean добавляет логическое поле
func (b *Builder) Boolean(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: b.SchemaBuilder.Dialect.GetBooleanType()},
	})
}

// Text добавляет текстовое поле
func (b *Builder) Text(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "TEXT"},
	})
}

// Date добавляет поле даты
func (b *Builder) Date(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "DATE"},
	})
}

// DateTime добавляет поле даты и времени
func (b *Builder) DateTime(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "DATETIME"},
	})
}

// Decimal добавляет десятичное поле
func (b *Builder) Decimal(Name string, precision, scale int) *ColumnBuilder {
	return b.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: fmt.Sprintf("DECIMAL(%d,%d)", precision, scale)},
	})
}

// Json добавляет JSON поле
func (b *Builder) Json(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "JSON"},
	})
}

// Binary добавляет бинарное поле
func (b *Builder) Binary(Name string, length int) *ColumnBuilder {
	return b.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "BINARY", Length: length},
	})
}

// Float добавляет поле с плавающей точкой
func (b *Builder) Float(Name string, precision, scale int) *ColumnBuilder {
	return b.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: fmt.Sprintf("FLOAT(%d,%d)", precision, scale)},
	})
}

// MediumText добавляет поле MEDIUMTEXT
func (b *Builder) MediumText(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: b.SchemaBuilder.Dialect.GetMediumTextType(),
		},
	})
}

// LongText добавляет поле LONGTEXT
func (b *Builder) LongText(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: b.SchemaBuilder.Dialect.GetLongTextType(),
		},
	})
}

// Char добавляет поле фиксированной длины
func (b *Builder) Char(Name string, length int) *ColumnBuilder {
	return b.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "CHAR", Length: length},
	})
}

// ForeignKey добавляет внешний ключ
//func (b *Builder) ForeignKey(Name string, table string, Column string) *ColumnBuilder {
//	return b.addColumn(&Column{Name: Name, Type: "BIGINT", References: &ForeignKey{Table: table, Column: Column}})
//}

// ADDITIONAL TYPES
// Money добавляет денежное поле
func (b *Builder) Money(Name string) *ColumnBuilder {
	return b.Decimal(Name, 19, 4)
}

// Price добавляет поле цены
func (b *Builder) Price(Name string) *ColumnBuilder {
	return b.Decimal(Name, 10, 2)
}

// Percentage добавляет поле процентов
func (b *Builder) Percentage(Name string) *ColumnBuilder {
	return b.Decimal(Name, 5, 2)
}

// Status добавляет поле статуса
func (b *Builder) Status(Name string, statuses []string) *ColumnBuilder {
	return b.Enum(Name, statuses).Default(statuses[0])
}

// Slug добавляет поле для URL-совместимой строки
func (b *Builder) Slug(Name string) *ColumnBuilder {
	return b.String(Name, 255).Unique()
}

// Phone добавляет поле телефона
func (b *Builder) Phone(Name string) *ColumnBuilder {
	return b.String(Name, 20)
}

// Color добавляет поле цвета (HEX)
func (b *Builder) Color(Name string) *ColumnBuilder {
	return b.Char(Name, 7)
}

// Language добавляет поле языка
func (b *Builder) Language() *ColumnBuilder {
	return b.Char("lang", 2)
}

// Country добавляет поле страны
func (b *Builder) Country() *ColumnBuilder {
	return b.Char("country", 2)
}

// Currency добавляет поле валюты
func (b *Builder) Currency() *ColumnBuilder {
	return b.Char("currency", 3)
}

// Timezone добавляет поле часового пояса
func (b *Builder) Timezone() *ColumnBuilder {
	return b.String("timezone", 64)
}

//func (b *Builder) ID(Name string) *ColumnBuilder {
//	SchemaBuilder.BigInteger("id").Unsigned().AutoIncrement().Primary()
//	return b.addColumn(&Column{
//		Name:       Name,
//		Definition: ColumnDefinition{Type: b.SchemaBuilder.Dialect.GetBigIntegerType()},
//	})
//}

// Year добавляет поле года
func (b *Builder) Year(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: b.SchemaBuilder.Dialect.GetYearType(),
		},
	})
}

// Time добавляет поле времени
func (b *Builder) Time(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name:       Name,
		Definition: ColumnDefinition{Type: "TIME"},
	})
}

// Ip добавляет поле для IP-адреса
func (b *Builder) Ip(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: b.SchemaBuilder.Dialect.GetIpType(),
		},
	})
}

// MacAddress добавляет поле для MAC-адреса
func (b *Builder) MacAddress(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: b.SchemaBuilder.Dialect.GetMacAddressType(),
		},
	})
}

// Point добавляет геометрическое поле точки
func (b *Builder) Point(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: b.SchemaBuilder.Dialect.GetPointType(),
		},
	})
}

// Polygon добавляет геометрическое поле полигона
func (b *Builder) Polygon(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: b.SchemaBuilder.Dialect.GetPolygonType(),
		},
	})
}

// Set добавляет поле SET
func (b *Builder) Set(Name string, values []string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: b.SchemaBuilder.Dialect.GetSetType(values),
		},
	})
}

// RememberToken добавляет поле для токена remember_token
func (b *Builder) RememberToken() *ColumnBuilder {
	return b.String("remember_token", 100).Nullable()
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
func (b *Builder) Geometry(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: b.SchemaBuilder.Dialect.GetGeometryType(),
		},
	})
}

// UUID добавляет поле UUID
func (b *Builder) UUID(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: b.SchemaBuilder.Dialect.GetUUIDType(),
		},
	})
}

// Double доба��ляет поле с двойной точностью
func (b *Builder) Double(Name string) *ColumnBuilder {
	return b.addColumn(&Column{
		Name: Name,
		Definition: ColumnDefinition{
			Type: b.SchemaBuilder.Dialect.GetDoubleType(),
		},
	})
}

// Password добавляет поле для хранения хэша пароля
func (b *Builder) Password(Name string) *ColumnBuilder {
	return b.String(Name, 60) // Достаточно для bcrypt
}

// Email добавляет поле email
func (b *Builder) Email(Name string) *ColumnBuilder {
	return b.String(Name, 255)
}

// Url добавляет поле URL
func (b *Builder) Url(Name string) *ColumnBuilder {
	return b.String(Name, 2083) // Максимальная длина URL в IE
}

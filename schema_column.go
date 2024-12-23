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

// ColumnBuilder построитель колонок
type ColumnBuilder struct {
	schema *Schema
	column Column
}

func (s *Schema) addColumn(col Column) *ColumnBuilder {
	if s.mode == "create" {
		s.columns = append(s.columns, col)
	} else {
		s.AddColumn(col)
	}
	return &ColumnBuilder{column: col}
}

func (cb *ColumnBuilder) Add() *Schema {
	cb.schema.addColumn(cb.column)
	return cb.schema
}

// AddColumn добавляет колонку
func (s *Schema) AddColumn(column Column) *ColumnBuilder {
	position := ""
	if column.After != "" {
		position = fmt.Sprintf(" AFTER %s", column.After)
	} else if column.First {
		position = " FIRST"
	}

	s.commands = append(s.commands, fmt.Sprintf(
		"ADD COLUMN %s%s",
		buildColumnDefinition(column),
		position,
	))
	return &ColumnBuilder{column: column}
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

func (s *Schema) ID() *ColumnBuilder {
	return s.addColumn(Column{Name: "id", Type: "BIGINT", AutoIncrement: true, Primary: true})
}

// BigIncrements добавляет автоинкрементное большое целое
func (s *Schema) BigIncrements(name string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "BIGINT", AutoIncrement: true, Primary: true})
}

// String добавляет строковое поле
func (s *Schema) String(name string, length int) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "VARCHAR", Length: length})
}

// Enum добавляет поле с перечислением
func (s *Schema) Enum(name string, values []string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: fmt.Sprintf("ENUM('%s')", strings.Join(values, "','"))})
}

// Timestamp добавляет поле метки времени
func (s *Schema) Timestamp(name string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "TIMESTAMP"})
}

// Index добавляет индекс
func (s *Schema) Index(name string, columns ...string) *Schema {
	s.indexes[name] = columns
	return s
}

// Integer добавляет целочисленное поле
func (s *Schema) Integer(name string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "INT"})
}

// TinyInteger добавляет малое целое
func (s *Schema) TinyInteger(name string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "TINYINT"})
}

// Boolean добавляет логическое поле
func (s *Schema) Boolean(name string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "BOOLEAN"})
}

// Text добавляет текстовое поле
func (s *Schema) Text(name string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "TEXT"})
}

// Date добавляет поле даты
func (s *Schema) Date(name string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "DATE"})
}

// DateTime добавляет поле даты и времени
func (s *Schema) DateTime(name string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "DATETIME"})
}

// Decimal добавляет десятичное поле
func (s *Schema) Decimal(name string, precision, scale int) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: fmt.Sprintf("DECIMAL(%d,%d)", precision, scale)})
}

// Json добавляет JSON поле
func (s *Schema) Json(name string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "JSON"})
}

// Binary добавляет бинарное поле
func (s *Schema) Binary(name string, length int) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "BINARY", Length: length})
}

// Float добавляет поле с плавающей точкой
func (s *Schema) Float(name string, precision, scale int) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: fmt.Sprintf("FLOAT(%d,%d)", precision, scale)})
}

//// ForeignKey добавляет внешний ключ
//func (s *Schema) ForeignKey(name string, table string, column string) *ColumnBuilder {
//	return s.addColumn(Column{Name: name, Type: "BIGINT", References: &ForeignKey{Table: table, Column: column}})
//}

// Uuid добавляет поле UUID
func (s *Schema) Uuid(name string) *ColumnBuilder {
	if s.dbl.db.DriverName() == "postgres" {
		return s.addColumn(Column{Name: name, Type: "UUID"})
	}
	return s.addColumn(Column{Name: name, Type: "CHAR", Length: 36})
}

// Computed добавляет вычисляемую колонку (для MySQL 5.7+)
func (s *Schema) Computed(name string, expression string) *ColumnBuilder {
	if s.dbl.db.DriverName() == "mysql" {
		return s.addColumn(Column{Name: name, Type: fmt.Sprintf("AS (%s) STORED", expression)})
	}
	return nil
}

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

// SoftDeletes добавляет поле deleted_at для мягкого удаления
func (s *Schema) SoftDeletes() *ColumnBuilder {
	return s.addColumn(Column{Name: "deleted_at", Type: "TIMESTAMP", Nullable: true})
}

// MediumText добавляет поле MEDIUMTEXT
func (s *Schema) MediumText(name string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "MEDIUMTEXT"})
}

// LongText добавляет поле LONGTEXT
func (s *Schema) LongText(name string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "LONGTEXT"})
}

// Char добавляет поле фиксированной длины
func (s *Schema) Char(name string, length int) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "CHAR", Length: length})
}

// SmallInteger добавляет малое целое
func (s *Schema) SmallInteger(name string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "SMALLINT"})
}

// MediumInteger добавляет среднее целое
func (s *Schema) MediumInteger(name string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "MEDIUMINT"})
}

// Year добавляет поле года
func (s *Schema) Year(name string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "YEAR"})
}

// Time добавляет поле времени
func (s *Schema) Time(name string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "TIME"})
}

// Ip добавляет поле для IP-адреса
func (s *Schema) Ip(name string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "VARCHAR", Length: 45})
}

// MacAddress добавляет поле для MAC-адреса
func (s *Schema) MacAddress(name string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "VARCHAR", Length: 17})
}

// Point добавляет геометрическое поле точки
func (s *Schema) Point(name string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "POINT"})
}

// Polygon добавляет геометрическое поле полигона
func (s *Schema) Polygon(name string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "POLYGON"})
}

// Set добавляет поле SET
func (s *Schema) Set(name string, values []string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: fmt.Sprintf("SET('%s')", strings.Join(values, "','"))})
}

// RememberToken добавляет поле для токена remember_token
func (s *Schema) RememberToken() *ColumnBuilder {
	return s.String("remember_token", 100).Nullable()
}

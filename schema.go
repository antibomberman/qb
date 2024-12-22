package dblayer

import (
	"fmt"
	"strings"
)

// Schema представляет построитель схемы таблицы
type Schema struct {
	table *TableBuilder
	alter *AlterTable // Добавляем поле для ALTER TABLE
	mode  string      // "create" или "update"
}

// Модифицируем существующие методы для поддержки обоих режимов
func (s *Schema) addColumn(col Column) *ColumnBuilder {
	if s.mode == "create" {
		s.table.columns = append(s.table.columns, col)
	} else {
		s.alter.AddColumn(col)
	}
	return &ColumnBuilder{column: col}
}

// Добавляем методы для обновления
func (s *Schema) RenameColumn(from, to string) *Schema {
	if s.mode == "update" {
		s.alter.RenameColumn(from, to)
	}
	return s
}

func (s *Schema) DropColumn(name string) *Schema {
	if s.mode == "update" {
		s.alter.DropColumn(name)
	}
	return s
}

func (s *Schema) ModifyColumn(name string, fn func(*ColumnBuilder)) *Schema {
	if s.mode == "update" {
		cb := &ColumnBuilder{column: Column{Name: name}}
		fn(cb)
		s.alter.ModifyColumn(cb.column)
	}
	return s
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
func (s *Schema) Index(name string, columns ...string) *TableBuilder {
	return s.table.Index(name, columns...)
}

// Comment добавляет комментарий к таблице
func (s *Schema) Comment(comment string) *TableBuilder {
	return s.table.Comment(comment)
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

// ForeignId добавляет внешний ключ
func (s *Schema) ForeignId(name string, table string) *ColumnBuilder {
	return s.addColumn(Column{Name: name, Type: "BIGINT", References: &ForeignKey{Table: table, Column: "id"}})
}

// Uuid добавляет поле UUID
func (s *Schema) Uuid(name string) *ColumnBuilder {
	if s.table.dbl.db.DriverName() == "postgres" {
		return s.addColumn(Column{Name: name, Type: "UUID"})
	}
	return s.addColumn(Column{Name: name, Type: "CHAR", Length: 36})
}

// Timestamps добавляет поля created_at и updated_at
func (s *Schema) Timestamps() *Schema {
	s.Timestamp("created_at").Default("CURRENT_TIMESTAMP")
	s.Timestamp("updated_at").Nullable()
	return s
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

// Morphs добавляет поля для полиморфных отношений
func (s *Schema) Morphs(name string) *Schema {
	s.Integer(name + "_id")
	s.String(name+"_type", 255)
	s.Index("idx_"+name, name+"_id", name+"_type")
	return s
}

// UniqueIndex добавляет уникальный индекс
func (s *Schema) UniqueIndex(name string, columns ...string) *TableBuilder {
	return s.table.UniqueKey(name, columns...)
}

// FullText добавляет полнотекстовый индекс
func (s *Schema) FullText(name string, columns ...string) *TableBuilder {
	// Реализация зависит от типа БД
	if s.table.dbl.db.DriverName() == "mysql" {
		s.table.indexes[name] = columns
		return s.table
	}
	return s.table
}

// Computed добавляет вычисляемую колонку (для MySQL 5.7+)
func (s *Schema) Computed(name string, expression string) *ColumnBuilder {
	if s.table.dbl.db.DriverName() == "mysql" {
		return s.addColumn(Column{Name: name, Type: fmt.Sprintf("AS (%s) STORED", expression)})
	}
	return nil
}

// Virtual добавляет виртуальную колонку
func (s *Schema) Virtual(name string, expression string) *ColumnBuilder {
	if s.table.dbl.db.DriverName() == "mysql" {
		return s.addColumn(Column{Name: name, Type: fmt.Sprintf("AS (%s) VIRTUAL", expression)})
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
func (s *Schema) Language(name string) *ColumnBuilder {
	return s.Char(name, 2)
}

// Country добавляет поле страны
func (s *Schema) Country(name string) *ColumnBuilder {
	return s.Char(name, 2)
}

// Currency добавляет поле валюты
func (s *Schema) Currency(name string) *ColumnBuilder {
	return s.Char(name, 3)
}

// Timezone добавляет поле часового пояса
func (s *Schema) Timezone(name string) *ColumnBuilder {
	return s.String(name, 64)
}

// Dimensions добавляет поля размеров
func (s *Schema) Dimensions(prefix string) *Schema {
	s.Decimal(prefix+"_length", 8, 2)
	s.Decimal(prefix+"_width", 8, 2)
	s.Decimal(prefix+"_height", 8, 2)
	return s
}

// Address добавляет поля адреса
func (s *Schema) Address() *Schema {
	s.String("country", 2)
	s.String("city", 100)
	s.String("street", 255)
	s.String("house", 20)
	s.String("apartment", 20).Nullable()
	s.String("postal_code", 20)
	return s
}

// Seo добавляет поля для SEO
func (s *Schema) Seo() *Schema {
	s.String("meta_title", 255).Nullable()
	s.String("meta_description", 255).Nullable()
	s.String("meta_keywords", 255).Nullable()
	return s
}

// Audit добавляет поля аудита
func (s *Schema) Audit() *Schema {
	s.ForeignId("created_by", "users").Nullable()
	s.ForeignId("updated_by", "users").Nullable()
	s.ForeignId("deleted_by", "users").Nullable()
	s.Timestamps()
	s.SoftDeletes()
	return s
}

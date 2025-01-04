package schema

import (
	"fmt"
	"github.com/antibomberman/dbl/dialect"
	"github.com/jmoiron/sqlx"
)

type Schema struct {
	DB         *sqlx.DB
	DriverName string
	Dialect    dialect.Dialect
}

func (s *Schema) setDialect() {
	switch s.DriverName {
	case "mysql":
		s.Dialect = &dialect.MysqlDialect{}
	case "postgres":
		s.Dialect = &dialect.PostgresDialect{}
	case "sqlite":
		s.Dialect = &dialect.SqliteDialect{}
	}
}
func NewSchema(db *sqlx.DB, driverName string) *Schema {
	schema := &Schema{
		DB:         db,
		DriverName: driverName,
	}
	switch driverName {
	case "mysql":
		schema.Dialect = &dialect.MysqlDialect{}
	case "postgres":
		schema.Dialect = &dialect.PostgresDialect{}
	case "sqlite":
		schema.Dialect = &dialect.SqliteDialect{}
	}
	return schema

}

func (s *Schema) TruncateTable(tables ...string) *TruncateTable {
	tb := &TruncateTable{
		Schema: s,
		Tables: tables,
	}

	return tb
}

func (s *Schema) DropTable(tables ...string) *DropTable {
	return &DropTable{
		Schema: s,
		Tables: tables,
	}
}

func (s *Schema) CreateTable(name string, fn func(builder *Builder)) error {
	builder := &Builder{
		Schema: s,
		Definition: Definition{
			Name: name,
			Mode: "create",
			Options: TableOptions{
				IfNotExists: false,
			},
			// Инициализируем все maps в constraints
			KeyIndex: KeyIndex{
				PrimaryKey:  make([]string, 0),
				UniqueKeys:  make(map[string][]string),
				Indexes:     make(map[string][]string),
				ForeignKeys: make(map[string]*Foreign),
			},
		},
	}

	fn(builder)
	_, err := s.DB.Exec(builder.BuildCreate())
	return err
}

func (s *Schema) CreateTableIfNotExists(name string, fn func(builder *Builder)) error {
	schema := &Builder{
		Schema: s,
		Definition: Definition{
			Name: name,
			Options: TableOptions{
				IfNotExists: true,
			},
			Mode: "create",
			KeyIndex: KeyIndex{
				PrimaryKey:  make([]string, 0),
				UniqueKeys:  make(map[string][]string),
				Indexes:     make(map[string][]string),
				ForeignKeys: make(map[string]*Foreign),
			},
		},
	}

	fn(schema)
	fmt.Println(schema.BuildCreate())
	_, err := s.DB.Exec(schema.BuildCreate())
	return err
}

func (s *Schema) UpdateTable(name string, fn func(builder *Builder)) error {
	schema := &Builder{
		Schema: s,
		Definition: Definition{
			Name: name,
			Mode: "update",
		},
	}

	fn(schema)

	_, err := s.DB.Exec(schema.BuildAlter())
	return err
}

//TODO  SHOW TABLES
//Mysql
//show table `SHOW TABLES LIKE 'users'`
//check table `DESCRIBE`
//

//postgresql
// show table `SELECT EXISTS ( SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users' )`
//check table `SELECT column_name, data_type, character_maximum_length FROM information_schema.columns WHERE table_name = 'users'``

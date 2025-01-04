package schema

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

type Schema struct {
	DB         *sqlx.DB
	DriverName string
	Dialect    Dialect
}

func NewSchemaBuilder(db *sqlx.DB, driverName string) *Schema {
	schema := &Schema{
		DB:         db,
		DriverName: driverName,
	}
	switch driverName {
	case "mysql":
		schema.Dialect = &MysqlDialect{}
	case "postgres":
		schema.Dialect = &PostgresDialect{}
	case "sqlite":
		schema.Dialect = &SqliteDialect{}
	}
	return schema

}

func (s *Schema) TruncateTable(tables ...string) error {
	tb := &TruncateTable{
		Schema: s,
		Tables: tables,
	}

	_, err := s.DB.Exec(tb.Build())
	return err
}

func (s *Schema) DropTable(tables ...string) error {
	dt := &DropTable{
		Schema: s,
		Tables: tables,
	}
	_, err := s.DB.Exec(dt.Build())
	return err
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

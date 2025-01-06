package table

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

type DatabaseConfig struct {
	DB         *sqlx.DB
	DriverName string
	Dialect    Dialect
}
type TableBuilder struct {
	*DatabaseConfig
	Name string
}

func New(db *sqlx.DB, driverName string) *DatabaseConfig {
	tableBuilder := &DatabaseConfig{
		DB:         db,
		DriverName: driverName,
	}
	switch driverName {
	case "mysql":
		tableBuilder.Dialect = &MysqlDialect{}
	case "postgres":
		tableBuilder.Dialect = &PostgresDialect{}
	case "sqlite":
		tableBuilder.Dialect = &SqliteDialect{}
	}
	return tableBuilder

}

func (dc *DatabaseConfig) Table(name string) *TableBuilder {
	return &TableBuilder{
		Name:           name,
		DatabaseConfig: dc,
	}
}

func (s *TableBuilder) Truncate() error {
	tb := &TruncateTable{
		TableBuilder: s,
		Tables:       []string{s.Name},
	}

	_, err := s.DB.Exec(tb.Build())
	return err
}

func (s *TableBuilder) Drop() error {
	dt := &DropTable{
		TableBuilder: s,
		Tables:       []string{s.Name},
	}
	_, err := s.DB.Exec(dt.Build())
	return err
}

func (s *TableBuilder) Create(fn func(builder *Builder)) error {
	builder := &Builder{
		TableBuilder: s,
		Definition: Definition{
			Name: s.Name,
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

func (s *TableBuilder) CreateIfNotExists(fn func(builder *Builder)) error {
	schema := &Builder{
		TableBuilder: s,
		Definition: Definition{
			Name: s.Name,
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

func (s *TableBuilder) Update(name string, fn func(builder *Builder)) error {
	schema := &Builder{
		TableBuilder: s,
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

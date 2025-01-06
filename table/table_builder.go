package table

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type TableBuilder struct {
	db         *sqlx.DB
	driverName string
	dialect    Dialect
}
type Table struct {
	*TableBuilder
	Name string
}

func New(driverName string, db *sqlx.DB) *TableBuilder {
	dc := &TableBuilder{
		db:         db,
		driverName: driverName,
	}
	switch driverName {
	case "mysql":
		dc.dialect = &MysqlDialect{}
	case "postgres":
		dc.dialect = &PostgresDialect{}
	case "sqlite":
		dc.dialect = &SqliteDialect{}
	}
	return dc

}

func (t *TableBuilder) Table(name string) *Table {
	return &Table{
		Name:         name,
		TableBuilder: t,
	}
}

func (s *Table) Truncate() error {
	tb := &TruncateTable{
		Table:  s,
		Tables: []string{s.Name},
	}

	_, err := s.db.Exec(tb.Build())
	return err
}

func (s *Table) Drop() error {
	dt := &DropTable{
		Table:  s,
		Tables: []string{s.Name},
	}
	_, err := s.db.Exec(dt.Build())
	return err
}

func (s *Table) Create(fn func(builder *Builder)) error {
	builder := &Builder{
		Table: s,
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
	_, err := s.db.Exec(builder.BuildCreate())
	return err
}

func (s *Table) CreateIfNotExists(fn func(builder *Builder)) error {
	schema := &Builder{
		Table: s,
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
	_, err := s.db.Exec(schema.BuildCreate())
	return err
}

func (s *Table) Update(name string, fn func(builder *Builder)) error {
	schema := &Builder{
		Table: s,
		Definition: Definition{
			Name: name,
			Mode: "update",
		},
	}

	fn(schema)

	_, err := s.db.Exec(schema.BuildAlter())
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

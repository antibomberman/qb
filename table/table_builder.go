package table

import (
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

func (s *Table) Truncate() (string, error) {
	tb := &TruncateTable{
		Table:  s,
		Tables: []string{s.Name},
	}

	sql := tb.Build()
	_, err := s.db.Exec(sql)
	return sql, err
}

func (s *Table) Drop() (string, error) {
	dt := &DropTable{
		Table:  s,
		Tables: []string{s.Name},
	}
	sql := dt.Build()
	_, err := s.db.Exec(sql)
	return sql, err
}

func (s *Table) Create(fn func(builder *Builder)) (string, error) {
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
	sql := builder.BuildCreate()
	_, err := s.db.Exec(sql)
	return sql, err
}

func (s *Table) CreateIfNotExists(fn func(builder *Builder)) (string, error) {
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
	sql := schema.BuildCreate()
	_, err := s.db.Exec(sql)
	return sql, err
}

func (s *Table) Update(name string, fn func(builder *Builder)) (string, error) {
	schema := &Builder{
		Table: s,
		Definition: Definition{
			Name: name,
			Mode: "update",
		},
	}

	fn(schema)

	sql := schema.BuildAlter()
	_, err := s.db.Exec(sql)
	return sql, err
}

//TODO  SHOW TABLES
//Mysql
//show table `SHOW TABLES LIKE 'users'`
//check table `DESCRIBE`
//

//postgresql
// show table `SELECT EXISTS ( SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users' )`
//check table `SELECT column_name, data_type, character_maximum_length FROM information_schema.columns WHERE table_name = 'users'``

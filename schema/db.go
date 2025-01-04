package schema

import (
	"fmt"
	"github.com/antibomberman/dbl/dialect"
	"github.com/jmoiron/sqlx"
)

type DBL struct {
	DB         *sqlx.DB
	DriverName string
	Dialect    dialect.Dialect
}

func (dbl *DBL) setDialect() {
	switch dbl.DriverName {
	case "mysql":
		dbl.Dialect = &dialect.MysqlDialect{}
	case "postgres":
		dbl.Dialect = &dialect.PostgresDialect{}
	case "sqlite":
		dbl.Dialect = &dialect.SqliteDialect{}
	}
}

func (dbl *DBL) Truncate(tables ...string) *TruncateTable {
	return &TruncateTable{
		DBL:    dbl,
		Tables: tables,
	}
}

func (dbl *DBL) Drop(tables ...string) *DropTable {
	return &DropTable{
		DBL:    dbl,
		Tables: tables,
	}
}

func (dbl *DBL) CreateTable(name string, fn func(*Schema)) error {
	schema := &Schema{
		dbl: dbl,
		Definition: SchemaDefinition{
			Name: name,
			Mode: "create",
			Options: TableOptions{
				IfNotExists: false,
			},
			// Инициализируем все maps в constraints
			Constraints: Constraints{
				PrimaryKey:  make([]string, 0),
				UniqueKeys:  make(map[string][]string),
				Indexes:     make(map[string][]string),
				ForeignKeys: make(map[string]*Foreign),
			},
		},
	}

	fn(schema)
	_, err := dbl.DB.Exec(schema.BuildCreate())
	return err
}

func (dbl *DBL) CreateTableIfNotExists(name string, fn func(*Schema)) error {
	schema := &Schema{
		dbl: dbl,
		Definition: SchemaDefinition{
			Name: name,
			Options: TableOptions{
				IfNotExists: true,
			},
			Mode: "create",
			Constraints: Constraints{
				PrimaryKey:  make([]string, 0),
				UniqueKeys:  make(map[string][]string),
				Indexes:     make(map[string][]string),
				ForeignKeys: make(map[string]*Foreign),
			},
		},
	}

	fn(schema)
	fmt.Println(schema.BuildCreate())
	_, err := dbl.DB.Exec(schema.BuildCreate())
	return err
}

func (dbl *DBL) UpdateTable(name string, fn func(*Schema)) error {
	schema := &Schema{
		dbl: dbl,
		Definition: SchemaDefinition{
			Name: name,
			Mode: "update",
		},
	}

	fn(schema)

	_, err := dbl.DB.Exec(schema.BuildAlter())
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

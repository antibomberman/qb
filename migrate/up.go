package migrate

import (
	"github.com/antibomberman/dblayer/migrate/parser"
	"github.com/antibomberman/dblayer/migrate/utils"
	"log"
	"os"
)

func (m *MigrateBuilder) Up() error {
	//check exist dir migrations
	if _, err := os.Stat(utils.MigrationDir); os.IsNotExist(err) {
		err := os.Mkdir("migrations", 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	files, err := parser.Files()
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		upSQL, _, err := parser.SqlFile(utils.MigrationDir + "/" + file)
		if err != nil {

			return err
		}

		err = m.queryBuilder.Raw(upSQL).Exec()
		if err != nil {
			return err
		}

	}
	return nil

	//dbl, err := ConnectDB()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//err = dbl.CreateTableIfNotExists("migrations", func(schema *DBL.Schema) {
	//	schema.ID()
	//	schema.String("name", 255)
	//	schema.Text("up")
	//	schema.Text("down")
	//	schema.String("status", 256).Default("init")
	//	schema.Integer("version").Default(0)
	//	schema.Timestamps()
	//})
	//if err != nil {
	//	log.Fatal(err)
	//}

	//get all migrations from migrations dir

	//get all migrations from migrations table

	//compare migrations

	//apply migrations

	//update migrations table

}

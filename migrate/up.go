package migrate

import (
	"github.com/antibomberman/dblayer/migrate/parser"
	"log"
	"os"
)

func (m *MigrateBuilder) Up() {
	//check exist dir migrations
	if _, err := os.Stat("migrations"); os.IsNotExist(err) {
		err := os.Mkdir("migrations", 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	files, err := parser.Files("sql")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		upSQL, _, err := parser.SqlFile("migrations/" + file)
		if err != nil {
			log.Fatal(err)
		}

		err = m.queryBuilder.Raw(upSQL).Exec()
		if err != nil {
			panic(err)
		}

	}

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

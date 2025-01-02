package internal

import (
	"fmt"
	DBL "github.com/antibomberman/dbl"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var UpCmd = &cobra.Command{
	Use:   "up",
	Short: "Применить все миграции",
	Run: func(cmd *cobra.Command, args []string) {
		//check exist dir migrations

		if _, err := os.Stat("migrations"); os.IsNotExist(err) {
			err := os.Mkdir("migrations", 0755)
			if err != nil {
				log.Fatal(err)
			}
		}

		//create table migrations

		dbl, err := ConnectDB()
		if err != nil {
			log.Fatal(err)
		}
		err = dbl.CreateTableIfNotExists("migrations", func(schema *DBL.Schema) {
			schema.ID()
			schema.String("name", 255)
			schema.Text("up")
			schema.Text("down")
			schema.String("status", 256).Default("init")
			schema.Integer("version").Default(0)
			schema.Timestamps()
		})
		if err != nil {
			log.Fatal(err)
		}

		//get all migrations from migrations dir

		//get all migrations from migrations table

		//compare migrations

		//apply migrations

		//update migrations table

	},
}

func getSqlFiles() {

	files, err := os.ReadDir("migrations")
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		fmt.Println(f.Name())
	}
}

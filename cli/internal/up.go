package internal

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"strings"
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
		files, err := getFiles("sql")
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range files {
			upSQL, downSQL, err := parseSqlFromFile("migrations/" + file)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(upSQL)
			fmt.Println(downSQL)
		}
		//create table migrations

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

	},
}

func getFiles(extType string) ([]string, error) {
	files, err := os.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var newFiles []string
	for _, f := range files {
		if filepath.Ext(f.Name())[1:] == extType {
			newFiles = append(newFiles, f.Name())
		}
	}

	return newFiles, nil
}
func parseSqlFromFile(filename string) (string, string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", "", fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Разделяем содержимое на строки
	lines := strings.Split(string(content), "\n")

	var upSQL, downSQL strings.Builder
	var isUp bool

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Определяем секцию UP или DOWN
		if strings.HasPrefix(trimmedLine, "-- DBL UP") {
			isUp = true
			continue
		} else if strings.HasPrefix(trimmedLine, "-- DNL DOWN") {
			isUp = false
			continue
		}

		// Пропускаем пустые строки
		if trimmedLine == "" {
			continue
		}

		// Добавляем строку в соответствующую секцию
		if isUp {
			upSQL.WriteString(line + "\n")
		} else {
			downSQL.WriteString(line + "\n")
		}
	}

	// Проверяем, что обе секции не пустые
	if upSQL.Len() == 0 || downSQL.Len() == 0 {
		return "", "", fmt.Errorf("missing UP or DOWN section in file %s", filename)
	}

	return strings.TrimSpace(upSQL.String()), strings.TrimSpace(downSQL.String()), nil
}

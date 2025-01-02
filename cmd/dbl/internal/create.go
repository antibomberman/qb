package internal

import (
	"fmt"
	DBL "github.com/antibomberman/dbl"
	"log"
	"os"
	"time"
)

var defaultContent = `
-- DBL UP
CREATE TABLE IF NOT EXISTS %s (
	id bigint unsigned NOT NULL  PRIMARY KEY AUTO_INCREMENT COMMENT 'id',


    version int unsigned NOT NULL DEFAULT 0 COMMENT 'version',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated_at',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT 'deleted_at'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- DNL DOWN
DROP TABLE IF EXISTS %s;
`

func Create(name string) error {
	// Генерируем имя файла в формате: YYYYMMDDHHMMSS_название_миграции.sql
	timestamp := time.Now().Format("20060102150405")
	fileName := fmt.Sprintf("%s_%s.sql", timestamp, name)

	// Создаем папку migrations, если ее нет
	if _, err := os.Stat("migrations"); os.IsNotExist(err) {
		err := os.Mkdir("migrations", 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	// Создаем файл миграции
	file, err := os.Create(fmt.Sprintf("migrations/%s", fileName))
	if err != nil {
	}
	_, err = file.WriteString(fmt.Sprintf(defaultContent, name, name))
	if err != nil {
		return err
	}
	defer file.Close()

	return err
}

func migrationTable() {

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
}

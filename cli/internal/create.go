package internal

import (
	"fmt"
	"github.com/spf13/cobra"
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

var CreateCmd = &cobra.Command{
	Use:   "create [название_миграции] ",
	Short: "Создать новую миграцию",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Необходимо указать название миграции")
			return
		}
		GenerateDefaultFiles()

		name := args[0]
		timestamp := time.Now().Format("20060102150405")
		fileName := fmt.Sprintf("%s_%s.sql", timestamp, name)

		file, err := os.Create(fmt.Sprintf("migrations/%s", fileName))
		if err != nil {
			log.Fatal(err)
		}
		_, err = file.WriteString(fmt.Sprintf(defaultContent, name, name))
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

	},
}

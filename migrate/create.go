package migrate

import (
	"fmt"
	"log"
	"os"
	"time"
)

var defaultContent = `
-- UP
CREATE TABLE IF NOT EXISTS %s (
	id bigint unsigned NOT NULL  PRIMARY KEY AUTO_INCREMENT COMMENT 'id',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated_at',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT 'deleted_at'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- DOWN
DROP TABLE IF EXISTS %s;
`

func (*MigrateBuilder) Create(name string) (string, error) {
	timestamp := time.Now().Format("20060102150405")
	fileName := fmt.Sprintf("%s_%s.sql", timestamp, name)
	filePath := fmt.Sprintf("migrations/%s", fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	_, err = file.WriteString(fmt.Sprintf(defaultContent, name, name))
	if err != nil {
		return "", err

	}
	defer file.Close()
	return fileName, nil
}
func InitDir() {
	if _, err := os.Stat("migrations"); os.IsNotExist(err) {
		err := os.Mkdir("migrations", 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
}

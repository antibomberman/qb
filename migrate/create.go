package migrate

import (
	"fmt"
	"github.com/antibomberman/dblayer/migrate/utils"
	"log"
	"os"
	"time"
)

func (*MigrateBuilder) CreateSql(name string) (string, error) {
	timestamp := time.Now().Format("20060102150405")
	fileName := fmt.Sprintf("%s_%s.sql", timestamp, name)
	filePath := fmt.Sprintf("%s/%s", utils.MigrationDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	_, err = file.WriteString(utils.GetDefaultSqlContent(name))
	if err != nil {
		return "", err

	}
	defer file.Close()
	return fileName, nil
}
func (*MigrateBuilder) CreateGo(name string) (string, error) {
	timestamp := time.Now().Format("20060102150405")
	migrateName := fmt.Sprintf("%s_%s", timestamp, name)
	fileName := fmt.Sprintf("%s.go", migrateName)
	filePath := fmt.Sprintf("%s/%s", utils.MigrationDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	_, err = file.WriteString(utils.GetDefaultGoContent(name, utils.Upper(name)+timestamp))
	if err != nil {
		return "", err

	}
	defer file.Close()
	return fileName, nil
}

func InitDir() {
	if _, err := os.Stat(utils.MigrationDir); os.IsNotExist(err) {
		err := os.Mkdir(utils.MigrationDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
}

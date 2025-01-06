package utils

import "fmt"

const defaultSqlContent = `
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
const defaultGoContent = `package %s

import (
	q "github.com/antibomberman/dblayer/query"
	t "github.com/antibomberman/dblayer/table"
)

func Up%s (tb *t.TableBuilder, qb *q.QueryBuilder) (string, error) {
	return tb.Table("%s").CreateIfNotExists(func(b *t.Builder) {
		b.ID()

		b.Timestamps()
		b.SoftDeletes()
	})

}

func Down%s (sb *t.TableBuilder, qb *q.QueryBuilder) (string, error)m {
	return sb.Table("%s").Drop()
}
`
const defaultGoMainContent = `package main

import (
	"github.com/antibomberman/dblayer"
	_ "github.com/antibomberman/dblayer/cli/migrations"
	"github.com/antibomberman/dblayer/migrate"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	dbl, err := dblayer.ConnectEnv()
	if err != nil {
		log.Fatal(err)
	}

	for _, mFile := range migrate.GetAllMigrations() {
		_, err = (*mFile.UpFunc)(dbl.GetTableBuilder(), dbl.GetQueryBuilder())
		if err != nil {
			log.Fatal(err)
		}
	}
}
`

func GetDefaultSqlContent(tableName string) string {
	return fmt.Sprintf(defaultSqlContent, tableName, tableName)
}

func GetDefaultGoContent(tableName, migrationName string) string {
	return fmt.Sprintf(defaultGoContent, MigrationDir, migrationName, tableName, migrationName, tableName)
}
func GetDefaultGoMainContent() string {
	return defaultGoMainContent
}

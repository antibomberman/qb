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
	q "github.com/antibomberman/dblayer/table"
	t "github.com/antibomberman/dblayer/table"
)

func Up%s (tb *t.TableBuilder, qb *q.TableBuilder) error {
	return tb.Table("%s").CreateIfNotExists(func(b *t.Builder) {
		b.ID()

		b.Timestamps()
		b.SoftDeletes()
	})

}

func Down%s (sb *t.TableBuilder, qb *q.TableBuilder) error {
	return sb.Table("%s").Drop()
}
`

func GetDefaultSqlContent(tableName string) string {
	return fmt.Sprintf(defaultSqlContent, tableName, tableName)
}

func GetDefaultGoContent(tableName, migrationName string) string {
	return fmt.Sprintf(defaultGoContent, MigrationDir, migrationName, tableName, migrationName, tableName)
}

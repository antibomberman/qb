package migrate

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"time"

	"github.com/antibomberman/dblayer/migrate/parser"
	q "github.com/antibomberman/dblayer/query"
	t "github.com/antibomberman/dblayer/table"
)

type MigrationFunc func(tb *t.TableBuilder, qb *q.QueryBuilder) (string, error)
type Migration struct {
	ID       int64  `db:"id"`
	ExtType  string `db:"ext_type"`
	Name     string `db:"name"`
	Version  string `db:"version"`
	UpSql    string `db:"up"`
	DownSql  string `db:"down"`
	UpFunc   *MigrationFunc
	DownFunc *MigrationFunc

	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

var migrations = make(map[string]Migration)

func (m *MigrateBuilder) Up() error {

	cmd := exec.Command("go", "run", "cmd/migrate/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ошибка выполнения миграции: %w", err)
	}

	return nil
}

// register go files
func RegisterMigration(version string, up, down MigrationFunc) {
	migrations[version] = Migration{
		ExtType:  "go",
		UpFunc:   &up,
		DownFunc: &down,
	}
}

func GetMigration(version string) (Migration, bool) {
	m, ok := migrations[version]
	return m, ok
}

func GetAllMigrations() map[string]Migration {
	return migrations
}
func GetSortKeys() []string {
	keys := make([]string, 0, len(migrations))
	for k := range migrations {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func RegisterSqlMigrations() error {
	migrationFiles, err := parser.Files()
	if err != nil {
		return err
	}
	for _, migrationFile := range migrationFiles {

		if migrationFile.ExtType == "sql" {
			upSql, downSql, err := parser.GetSqlFromMigrationFile(migrationFile.Path)
			if err != nil {
				return err
			}
			migrations[migrationFile.Version] = Migration{
				Version: migrationFile.Version,
				Name:    migrationFile.Name,
				ExtType: "sql",
				UpSql:   upSql,
				DownSql: downSql,
			}
		}

	}
	return nil

}

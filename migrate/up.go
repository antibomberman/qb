package migrate

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/antibomberman/dblayer/migrate/parser"
	q "github.com/antibomberman/dblayer/query"
	t "github.com/antibomberman/dblayer/table"
)

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

type MigrationFunc func(tb *t.TableBuilder, qb *q.QueryBuilder) error
type Migration struct {
	ID       int64 `db:"id"`
	ExtType  string
	Name     string `db:"name"`
	UpSql    string `db:"up"`
	DownSql  string `db:"down"`
	UpFunc   *MigrationFunc
	DownFunc *MigrationFunc
	Status   string `db:"status"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

var migrations = make(map[string]Migration)

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

func Run() error {
	migrationFiles, err := parser.Files()
	if err != nil {
		return err
	}
	for _, migrationFile := range migrationFiles {

		if migrationFile.ExtType == "sql" {
			upSql, downSql, err := parser.SqlFile(migrationFile.Path)
			if err != nil {
				return err
			}
			migrations[migrationFile.Name] = Migration{
				ExtType:  "sql",
				UpSql:    upSql,
				DownSql:  downSql,
				UpFunc:   nil,
				DownFunc: nil,
			}
		}

	}
	return nil

}

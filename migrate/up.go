package migrate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"time"

	q "github.com/antibomberman/dblayer/query"
	t "github.com/antibomberman/dblayer/table"
)

func (m *MigrateBuilder) Up() error {
	//нужно запускать команду
	//go run cmd/migrate/main.go

	cmd := exec.Command("go", "run", "cmd/migrate/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ошибка выполнения миграции: %w", err)
	}

	return nil

	// migrationFiles, err := parser.Files()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// for _, migrationFile := range migrationFiles {
	// 	if migrationFile.ExtType == "go" {

	// 		//if err := downFunc(tb, qb); err != nil {
	// 		//	fmt.Println("Down error:", err)
	// 		//}
	// 	}
	// 	if migrationFile.ExtType == "sql" {
	// 		upSql, downSql, err := parser.SqlFile(migrationFile.Path)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		fmt.Println(upSql)
	// 		fmt.Println(downSql)

	// 	}

	// }
	// return nil

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

}

type MigrationFunc func(tb *t.TableBuilder, qb *q.QueryBuilder) error
type Migration struct {
	ID       int64 `db:"id"`
	ExtType  string
	Name     string `db:"name"`
	Up       string `db:"up"`
	Down     string `db:"down"`
	UpFunc   MigrationFunc
	DownFunc MigrationFunc
	Status   string `db:"status"`
	Version  int    `db:"version"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

var migrations = make(map[string]Migration)

func RegisterMigration(version string, up, down MigrationFunc) {
	migrations[version] = Migration{
		UpFunc:   up,
		DownFunc: down,
	}
}

func GetMigration(version string) (Migration, bool) {
	m, ok := migrations[version]
	return m, ok
}

func GetAllMigrations() map[string]Migration {
	return migrations
}

func (m *MigrateBuilder) TestRead() {
	// Путь к директории с миграциями
	migrationsDir := "./migrations"

	// Временная директория для .so файлов
	tmpDir := "./tmp"
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	// Находим все файлы миграций
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.go"))
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		baseName := filepath.Base(file)
		soPath := filepath.Join(tmpDir, baseName+".so")

		// Компилируем файл как плагин
		cmd := exec.Command("go", "build",
			"-buildmode=plugin",
			"-o", soPath,
			file,
		)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Ошибка компиляции %s: %v\n", file, err)
			continue
		}

		// Загружаем плагин
		p, err := plugin.Open(soPath)
		if err != nil {
			fmt.Printf("Ошибка загрузки плагина %s: %v\n", soPath, err)
			continue
		}

		// Получаем функции миграции
		// Предполагаем, что имя функции содержится в имени файла
		// Например: 20250106123010_tasks.go -> UpTasks20250106123010
		upFunc, err := p.Lookup("UpTasks20250106123010")
		if err != nil {
			fmt.Printf("Функция Up не найдена в %s: %v\n", file, err)
			continue
		}

		// Приводим к правильному типу
		up, ok := upFunc.(func(*t.TableBuilder, *q.QueryBuilder) error)
		if !ok {
			fmt.Printf("Неверный тип функции Up в %s\n", file)
			continue
		}

		tb := &t.TableBuilder{}
		qb := &q.QueryBuilder{}

		if err := up(tb, qb); err != nil {
			fmt.Printf("Ошибка выполнения миграции %s: %v\n", file, err)
			continue
		}

		fmt.Printf("Успешно выполнена миграция %s\n", file)
	}
}

// Выполняем все миграции
//for _, migration := range Migrations {
//	fmt.Printf("Running migration version: %s\n", migration.Version)
//	if err := migration.Up(db); err != nil {
//		log.Fatalf("Failed to apply migration version %s: %v", migration.Version, err)
//	}
//}

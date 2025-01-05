package internal

import (
	"fmt"
	sb "github.com/antibomberman/dblayer/schema"
	"github.com/jmoiron/sqlx"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Migration struct {
	Version int
	Up      func(*sb.Schema) error
	Down    func(*sb.Schema) error
}

type Migrator struct {
	schema     *sb.Schema
	migrations map[int]*Migration
}

func NewMigrator(db *sqlx.DB) *Migrator {

	return &Migrator{
		schema:     sb.NewSchemaBuilder(db, "mysql"),
		migrations: make(map[int]*Migration),
	}
}

// LoadMigrations загружает все миграции из директории
func (m *Migrator) LoadMigrations(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".base") {
			continue
		}

		// Извлекаем номер версии из имени файла (например, "00002" из "00002_create_users.base")
		versionStr := strings.Split(file.Name(), "_")[0]
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			return fmt.Errorf("invalid migration file name %s: %w", file.Name(), err)
		}

		// Загружаем файл миграции
		filePath := filepath.Join(dir, file.Name())
		migration, err := m.loadMigrationFile(filePath, version)
		if err != nil {
			return fmt.Errorf("failed to load migration %s: %w", file.Name(), err)
		}

		m.migrations[version] = migration
	}

	return nil
}

func (m *Migrator) loadMigrationFile(filePath string, version int) (*Migration, error) {
	// Здесь мы можем использовать base/parser для загрузки функций
	// Но для примера просто создадим заглушку
	return &Migration{
		Version: version,
		Up:      nil, // Тут должна быть реальная функция
		Down:    nil, // Тут должна быть реальная функция
	}, nil
}

// MigrateUp выполняет все миграции по порядку
func (m *Migrator) MigrateUp() error {
	// Получаем отсортированные версии
	versions := make([]int, 0, len(m.migrations))
	for version := range m.migrations {
		versions = append(versions, version)
	}
	sort.Ints(versions)

	for _, version := range versions {
		migration := m.migrations[version]
		if err := migration.Up(m.schema); err != nil {
			return fmt.Errorf("failed to run migration %d: %w", version, err)
		}
		fmt.Printf("Successfully ran migration %d\n", version)
	}
	return nil
}

// MigrateDown откатывает все миграции в обратном порядке
func (m *Migrator) MigrateDown() error {
	versions := make([]int, 0, len(m.migrations))
	for version := range m.migrations {
		versions = append(versions, version)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(versions)))

	for _, version := range versions {
		migration := m.migrations[version]
		if err := migration.Down(m.schema); err != nil {
			return fmt.Errorf("failed to rollback migration %d: %w", version, err)
		}
		fmt.Printf("Successfully rolled back migration %d\n", version)
	}
	return nil
}

// MigrateTo выполняет миграции до указанной версии
func (m *Migrator) MigrateTo(targetVersion int) error {
	versions := make([]int, 0, len(m.migrations))
	for version := range m.migrations {
		if version <= targetVersion {
			versions = append(versions, version)
		}
	}
	sort.Ints(versions)

	for _, version := range versions {
		migration := m.migrations[version]
		if err := migration.Up(m.schema); err != nil {
			return fmt.Errorf("failed to run migration %d: %w", version, err)
		}
		fmt.Printf("Successfully ran migration %d\n", version)
	}
	return nil
}

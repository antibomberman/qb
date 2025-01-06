package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/antibomberman/dblayer/migrate/utils"
)

type MigrationsFiles struct {
	Path    string
	ExtType string
	Name    string
}

func SqlFile(filename string) (string, string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", "", fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Разделяем содержимое на строки
	lines := strings.Split(string(content), "\n")

	var upSQL, downSQL strings.Builder
	var isUp bool

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Определяем секцию UP или DOWN
		if strings.HasPrefix(trimmedLine, "-- UP") {
			isUp = true
			continue
		} else if strings.HasPrefix(trimmedLine, "-- DOWN") {
			isUp = false
			continue
		}

		// Пропускаем пустые строки
		if trimmedLine == "" {
			continue
		}

		// Добавляем строку в соответствующую секцию
		if isUp {
			upSQL.WriteString(line + "\n")
		} else {
			downSQL.WriteString(line + "\n")
		}
	}

	// Проверяем, что обе секции не пустые
	if upSQL.Len() == 0 || downSQL.Len() == 0 {
		return "", "", fmt.Errorf("missing UP or DOWN section in file %s", filename)
	}

	return strings.TrimSpace(upSQL.String()), strings.TrimSpace(downSQL.String()), nil
}

func Files() ([]MigrationsFiles, error) {
	files, err := os.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var newFiles []MigrationsFiles

	for _, f := range files {
		extType := filepath.Ext(f.Name())[1:]
		if extType == "go" || extType == "sql" {
			newFiles = append(newFiles, MigrationsFiles{
				Path:    filepath.Join(utils.MigrationDir, "/", f.Name()),
				ExtType: extType,
				Name:    f.Name(),
			})
		}

	}

	return newFiles, nil
}

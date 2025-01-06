package parser

import (
	"fmt"
	t "github.com/antibomberman/dblayer/table"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func GoFile(filename string) (up func(builder *t.TableBuilder) error, down func(builder *t.TableBuilder) error, err error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse file: %w", err)
	}

	var upFound, downFound bool

	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if strings.HasPrefix(fn.Name.Name, "Up") {
				upFound = true
				up = func(s *t.TableBuilder) error {
					return s.Table("users").CreateIfNotExists(func(b *t.Builder) {
						b.String("name", 255)
						b.String("email", 255)
						b.String("password", 255)
						b.Timestamps()
					})
				}
			} else if strings.HasPrefix(fn.Name.Name, "Down") {
				downFound = true
				down = func(s *t.TableBuilder) error {
					return s.Table("users").Drop()
				}
			}
		}
		return true
	})

	if !upFound || !downFound {
		return nil, nil, fmt.Errorf("up or down function not found")
	}

	return up, down, nil
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

func Files() (map[string]string, error) {
	files, err := os.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var newFiles map[string]string

	for _, f := range files {
		newFiles[filepath.Ext(f.Name())[1:]] = f.Name()
	}

	return newFiles, nil
}

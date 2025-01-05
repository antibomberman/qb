package internal

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"
)

func parseGoMigration(filename string) (upFunc string, downFunc string, err error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", "", fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	fileContent := string(content)

	// Ищем Up функцию
	upRe := regexp.MustCompile(`func\s+Up\d+\([^)]+\)\s*error\s*{([\s\S]*?)}`)
	upMatches := upRe.FindStringSubmatch(fileContent)
	if len(upMatches) < 2 {
		return "", "", fmt.Errorf("up function not found in file")
	}

	// Ищем Down функцию
	downRe := regexp.MustCompile(`func\s+Down\d+\([^)]+\)\s*error\s*{([\s\S]*?)}`)
	downMatches := downRe.FindStringSubmatch(fileContent)
	if len(downMatches) < 2 {
		return "", "", fmt.Errorf("down function not found in file")
	}

	// Очищаем код от лишних пробелов
	upFunc = strings.TrimSpace(upMatches[1])
	downFunc = strings.TrimSpace(downMatches[1])

	return upFunc, downFunc, nil
}
func parseGoMigrationWithAst(filename string) (upFunc string, downFunc string, err error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse file: %w", err)
	}

	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			// Получаем содержимое файла для извлечения тела функции
			content, err := os.ReadFile(filename)
			if err != nil {
				return "", "", fmt.Errorf("failed to read file: %w", err)
			}
			lines := strings.Split(string(content), "\n")

			if strings.HasPrefix(fn.Name.Name, "Up") {
				start := fset.Position(fn.Body.Lbrace).Line
				end := fset.Position(fn.Body.Rbrace).Line
				upFunc = strings.Join(lines[start:end], "\n")
			} else if strings.HasPrefix(fn.Name.Name, "Down") {
				start := fset.Position(fn.Body.Lbrace).Line
				end := fset.Position(fn.Body.Rbrace).Line
				downFunc = strings.Join(lines[start:end], "\n")
			}
		}
	}

	if upFunc == "" || downFunc == "" {
		return "", "", fmt.Errorf("up or down function not found")
	}

	return strings.TrimSpace(upFunc), strings.TrimSpace(downFunc), nil
}

package internal

import (
	"fmt"
	sb "github.com/antibomberman/dblayer/schema"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func parseGoMigrationWithAst(filename string) (up func(*sb.Schema) error, down func(*sb.Schema) error, err error) {
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
				up = func(s *sb.Schema) error {
					return s.CreateTableIfNotExists("users", func(b *sb.Builder) {
						b.String("name", 255)
						b.String("email", 255)
						b.String("password", 255)
						b.Timestamps()
					})
				}
			} else if strings.HasPrefix(fn.Name.Name, "Down") {
				downFound = true
				down = func(s *sb.Schema) error {
					return s.DropTable("users")
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

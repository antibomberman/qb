package internal

import (
	"fmt"
	t "github.com/antibomberman/dblayer/table"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func parseGoMigrationWithAst(filename string) (up func(builder *t.TableBuilder) error, down func(builder *t.TableBuilder) error, err error) {
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

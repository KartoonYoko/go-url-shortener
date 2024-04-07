/*
Package osexitcheck анализатор, запрещающий использовать прямой вызов os.Exit в функции main пакета main.
*/
package osexitcheck

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// OsExitAnalyzer запрещает использовать прямой вызов os.Exit в функции main пакета main.
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osexitcheck",
	Doc:  "check for os.Exit in main function of main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}
	for _, file := range pass.Files {
		// функцией ast.Inspect проходим по всем узлам AST
		ast.Inspect(file, func(node ast.Node) bool {
			if file.Name.Name != "main" {
				return true
			}

			if x, ok := node.(*ast.ExprStmt); ok {
				if call, ok := x.X.(*ast.CallExpr); ok {
					if fun, ok := call.Fun.(*ast.SelectorExpr); ok {
						if fun.X.(*ast.Ident).Name == "os" && fun.Sel.Name == "Exit" {
							pass.Reportf(x.Pos(), "os exit in main function of main package")
						}
					}
				}
			}

			return true
		})
	}

	return nil, nil
}

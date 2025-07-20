package staticlint

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// NoOsExitAnalyzer - анализатор, который проверяет отсутствие вызовов os.Exit в функции main пакета main.
var NoOsExitAnalyzer = &analysis.Analyzer{
	Name: "no_os_exit_in_main",
	Doc:  "Проверка отсутствия вызова os.Exit в main",
	Run:  run,
}

// run запускает синтаксический анализ файлов
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if pass.Pkg.Name() != "main" {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
				ast.Inspect(fn.Body, func(nn ast.Node) bool {
					if call, ok := nn.(*ast.CallExpr); ok {
						if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
							if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "os" && sel.Sel.Name == "Exit" {
								pass.Reportf(call.Pos(), "использование os.Exit запрещено в main функции пакета main")
							}
						}
					}
					return true
				})
			}
			return true
		})
	}
	return nil, nil
}

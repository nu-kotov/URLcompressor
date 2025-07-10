package staticlint

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	mychecks := []*analysis.Analyzer{
		shadow.Analyzer,
		printf.Analyzer,
		structtag.Analyzer,
		NoOsExitAnalyzer,
	}
	for _, v := range staticcheck.Analyzers {
		if v.Analyzer.Name[:2] == "SA" || v.Analyzer.Name == "S1000" {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	multichecker.Main(mychecks...)
}

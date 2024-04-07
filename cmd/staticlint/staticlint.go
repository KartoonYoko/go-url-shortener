/*
Package staticlint это статический анализатор соответсвующий следующим требованиям:

- стандартных статических анализаторов пакета golang.org/x/tools/go/analysis/passes;
- всех анализаторов класса SA пакета staticcheck.io;
- не менее одного анализатора остальных классов пакета staticcheck.io;
- двух или более любых публичных анализаторов на ваш выбор.
*/
package staticlint

import (
	"strings"

	"github.com/KartoonYoko/go-url-shortener/cmd/staticlint/osexitcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	allAnalyzers := []*analysis.Analyzer{
		// анализаторы пакета golang.org/x/tools/go/analysis/passes
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		appends.Analyzer,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		defers.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		framepointer.Analyzer,
	}

	allAnalyzers = append(allAnalyzers, osexitcheck.OsExitAnalyzer)

	checks := map[string]bool{
		"S1000": true,
		"S1001": true,
	}

	for _, v := range staticcheck.Analyzers {
		// всех анализаторов класса SA пакета staticcheck.io
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			allAnalyzers = append(allAnalyzers, v.Analyzer)
		} else if checks[v.Analyzer.Name] {
			allAnalyzers = append(allAnalyzers, v.Analyzer)
		}
	}

	multichecker.Main(
		allAnalyzers...,
	)
}

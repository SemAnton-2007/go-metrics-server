/*
Мультичекер объединяет несколько статических анализаторов:
1. Стандартные анализаторы пакета golang.org/x/tools/go/analysis/passes
2. Все анализаторы класса SA пакета staticcheck.io
3. Публичные анализаторы: bodyclose и errcheck
4. Кастомный анализатор noexit

Запуск:

	go run cmd/staticlint/*.go ./...

Описание анализаторов:

Стандартные анализаторы:

	asmdecl, assign, atomic, bools, buildtag, cgocall, composite,
	copylock, errorsas, httpresponse, loopclosure, lostcancel,
	nilfunc, printf, shift, structtag, tests, unmarshal,
	unreachable, unsafeptr, unusedresult

SA (staticcheck):

	Все анализаторы класса SA (безопасность, корректность, стиль)

bodyclose: Проверяет закрытие тела HTTP-ответа.
errcheck: Обнаруживает необработанные ошибки.
noexit: Запрещает прямой вызов os.Exit в main.
*/
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"

	"honnef.co/go/tools/staticcheck"
)

func main() {
	analyzers := []*analysis.Analyzer{
		// Стандартные анализаторы
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,

		// Публичные анализаторы
		bodyclose.Analyzer,
		errcheck.Analyzer,

		// Кастомный анализатор
		NoExitAnalyzer,
	}

	// Добавляем все SA-анализаторы
	for _, v := range staticcheck.Analyzers {
		if len(v.Analyzer.Name) > 2 && v.Analyzer.Name[0:2] == "SA" {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	multichecker.Main(analyzers...)
}

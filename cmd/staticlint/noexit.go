package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var NoExitAnalyzer = &analysis.Analyzer{
	Name: "noexit",
	Doc:  "forbid direct calls to os.Exit in main function",
	Run:  runNoExit,
}

func runNoExit(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			if fd, ok := n.(*ast.FuncDecl); ok && fd.Name.Name == "main" {
				checkForExitCalls(pass, fd.Body)
				return false
			}
			return true
		})
	}
	return nil, nil
}

func checkForExitCalls(pass *analysis.Pass, body *ast.BlockStmt) {
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if isOsExitCall(call) {
			pass.Reportf(call.Pos(), "direct call to os.Exit in main function forbidden")
		}
		return true
	})
}

func isOsExitCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "os" {
			return sel.Sel.Name == "Exit"
		}
	}
	return false
}

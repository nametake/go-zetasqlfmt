package zetasqlfmt

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

func FindGoFiles(directory string, fn func(path string)) error {
	if err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if info.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		fn(path)
		return nil
	}); err != nil {
		return fmt.Errorf("failed to walk directory %s: %v", directory, err)
	}

	return nil
}

type FormatResult struct {
	Output  []byte
	Changed bool
}

func Format(path string) (*FormatResult, error) {
	cfg := &packages.Config{
		Mode: packages.LoadAllSyntax,
	}

	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("failed to load packages")
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected exactly one package")
	}

	pkg := pkgs[0]
	if len(pkg.Syntax) != 1 {
		return nil, fmt.Errorf("expected exactly one file")
	}
	file := pkg.Syntax[0]

	basicLitExprs := make([]*ast.BasicLit, 0)
	ast.Inspect(file, func(n ast.Node) bool {
		compositeLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		selectorExpr, ok := compositeLit.Type.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		use, ok := pkg.TypesInfo.Uses[selectorExpr.Sel]
		if !ok {
			return true
		}

		if use.Type().String() != "cloud.google.com/go/spanner.Statement" {
			return true
		}

		for _, elt := range compositeLit.Elts {
			elt, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := elt.Key.(*ast.Ident)
			if !ok {
				continue
			}
			if key.Name != "SQL" {
				continue
			}
			value, ok := elt.Value.(*ast.BasicLit)
			if !ok {
				continue
			}

			basicLitExprs = append(basicLitExprs, value)
		}
		return true
	})

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, pkg.Fset, file); err != nil {
		return nil, fmt.Errorf("%s: failed to print AST: %v", path, err)
	}

	result, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("%s: failed to format source: %v", path, err)
	}

	return &FormatResult{
		Output:  result,
		Changed: true,
	}, nil
}

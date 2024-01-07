package zetasqlfmt

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-zetasql"
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

	if len(basicLitExprs) == 0 {
		return &FormatResult{
			Output:  []byte{},
			Changed: false,
		}, nil
	}

	for _, basicLitExpr := range basicLitExprs {
		query := trimQuotes(basicLitExpr.Value)
		query = fillFormatVerbs(query)

		fmt.Println(query)

		output, err := zetasql.FormatSQL(query)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to format SQL: %v", path, err)
		}

		fmt.Println(output)

		output = restoreFormatVerbs(output)
		basicLitExpr.Value = wrapQuotes(output)
	}

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

func trimQuotes(s string) string {
	if len(s) < 2 {
		return s
	}
	if (s[0] != '"' && s[0] != '`') || (s[len(s)-1] != '"' && s[len(s)-1] != '`') {
		return s
	}
	return s[1 : len(s)-1]
}

func wrapQuotes(s string) string {
	if strings.Contains(s, "\n") {
		return fmt.Sprintf("`\n%s\n`", s)
	}
	return fmt.Sprintf("\"%s\"", s)
}

func fillFormatVerbs(sql string) string {
	dummyValues := make([]any, 0)
	isVerb := false
	for _, char := range sql {
		if char == '%' {
			isVerb = true
		} else if isVerb {
			switch char {
			case 'd':
				dummyValues = append(dummyValues, -999)
			case 'v':
				dummyValues = append(dummyValues, "_DUMMY_VALUE_")
			case 's':
				dummyValues = append(dummyValues, "_DUMMY_STRING_")
			}
			isVerb = false
		}
	}

	return fmt.Sprintf(sql, dummyValues...)
}

func restoreFormatVerbs(sql string) string {
	sql = strings.ReplaceAll(sql, "-999", "%d")
	sql = strings.ReplaceAll(sql, "_DUMMY_VALUE_", "%v")
	sql = strings.ReplaceAll(sql, "_DUMMY_STRING_", "%s")
	return sql
}

package zetasqlfmt

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goccy/go-zetasql"
	"golang.org/x/tools/go/packages"
)

func FindGoFiles(directory string, fn func(path string)) error {
	if err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			switch info.Name() {
			case "testdata", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		if strings.HasSuffix(path, "_gen.go") {
			return nil
		}
		fn(path)
		return nil
	}); err != nil {
		return fmt.Errorf("failed to walk directory %s: %v", directory, err)
	}

	return nil
}

type FormatError struct {
	Message string
	PosText string
}

func (e *FormatError) String() string {
	return fmt.Sprintf("%s:\n%s", e.PosText, e.Message)
}

type FormatResult struct {
	Path    string
	Output  []byte
	Errors  []*FormatError
	Changed bool
}

func Forma(pkg *packages.Package, file *ast.File) (*FormatResult, error) {
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

			switch valueExpr := elt.Value.(type) {
			case *ast.BasicLit:
				basicLitExprs = append(basicLitExprs, valueExpr)
			case *ast.CallExpr:
				callExpr, ok := valueExpr.Fun.(*ast.SelectorExpr)
				if !ok {
					continue
				}
				fn, ok := pkg.TypesInfo.Uses[callExpr.Sel]
				if !ok {
					continue
				}
				if fn.Pkg().Path() != "fmt" || fn.Name() != "Sprintf" {
					continue
				}
				if len(valueExpr.Args) < 1 {
					return true
				}
				argExpr := valueExpr.Args[0]
				v, ok := argExpr.(*ast.BasicLit)
				if !ok {
					return true
				}

				basicLitExprs = append(basicLitExprs, v)
			default:
			}

		}
		return true
	})

	errors := make([]*FormatError, 0, len(basicLitExprs))
	if len(basicLitExprs) == 0 {
		return &FormatResult{
			Output:  nil,
			Errors:  errors,
			Changed: false,
		}, nil
	}

	for _, basicLitExpr := range basicLitExprs {
		query := trimQuotes(basicLitExpr.Value)
		query = fillFormatVerbs(query)

		output, err := zetasql.FormatSQL(query)
		if err != nil {
			errors = append(errors, &FormatError{
				Message: err.Error(),
				PosText: pkg.Fset.Position(basicLitExpr.Pos()).String(),
			})
			continue
		}

		output = restoreFormatVerbs(output)
		basicLitExpr.Value = wrapQuotes(output)
	}

	if len(errors) == len(basicLitExprs) {
		return &FormatResult{
			Output:  nil,
			Errors:  errors,
			Changed: false,
		}, nil
	}

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, pkg.Fset, file); err != nil {
		return nil, fmt.Errorf("%s: failed to print AST: %v", pkg.Fset.Position(file.Pos()), err)
	}

	result, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("%s: failed to format source: %v", pkg.Fset.Position(file.Pos()), err)
	}

	return &FormatResult{
		Output:  result,
		Errors:  errors,
		Changed: true,
	}, nil
}

func FormatOld(path string) (*FormatResult, error) {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
	}

	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: path = %s: %v", path, err)
	}
	// NOTE If the package has syntax errors, it will be ignored.
	// if packages.PrintErrors(pkgs) > 0 {
	// 	return nil, fmt.Errorf("failed to load packages: path = %s: %v", path, err)
	// }
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected exactly one package: %s", path)
	}

	pkg := pkgs[0]

	if len(pkg.Syntax) != 1 {
		// for test.go file
		if len(pkg.Syntax) == 0 {
			return &FormatResult{
				Output:  []byte{},
				Errors:  []*FormatError{},
				Changed: false,
			}, nil
		}
		return nil, fmt.Errorf("expected exactly one file: %s", path)
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

			switch valueExpr := elt.Value.(type) {
			case *ast.BasicLit:
				basicLitExprs = append(basicLitExprs, valueExpr)
			case *ast.CallExpr:
				callExpr, ok := valueExpr.Fun.(*ast.SelectorExpr)
				if !ok {
					continue
				}
				fn, ok := pkg.TypesInfo.Uses[callExpr.Sel]
				if !ok {
					continue
				}
				if fn.Pkg().Path() != "fmt" || fn.Name() != "Sprintf" {
					continue
				}
				if len(valueExpr.Args) < 1 {
					return true
				}
				argExpr := valueExpr.Args[0]
				v, ok := argExpr.(*ast.BasicLit)
				if !ok {
					return true
				}

				basicLitExprs = append(basicLitExprs, v)
			default:
			}

		}
		return true
	})

	errors := make([]*FormatError, 0, len(basicLitExprs))
	if len(basicLitExprs) == 0 {
		return &FormatResult{
			Output:  nil,
			Errors:  errors,
			Changed: false,
		}, nil
	}

	for _, basicLitExpr := range basicLitExprs {
		query := trimQuotes(basicLitExpr.Value)
		query = fillFormatVerbs(query)

		output, err := zetasql.FormatSQL(query)
		if err != nil {
			errors = append(errors, &FormatError{
				Message: err.Error(),
				PosText: pkg.Fset.Position(basicLitExpr.Pos()).String(),
			})
			continue
		}

		output = restoreFormatVerbs(output)
		basicLitExpr.Value = wrapQuotes(output)
	}

	if len(errors) == len(basicLitExprs) {
		return &FormatResult{
			Output:  nil,
			Errors:  errors,
			Changed: false,
		}, nil
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
		Errors:  errors,
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
	if strings.Contains(s, "`") {
		return fmt.Sprintf("\"%s\"", removeNewlines(s))
	} else if strings.Contains(s, "\n") {
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

func removeNewlines(input string) string {
	// remove multiple spaces
	re := regexp.MustCompile(`\s{2,}`)
	result := re.ReplaceAllString(input, " ")
	// remove newlines
	result = strings.ReplaceAll(result, "\n", " ")
	result = strings.TrimSpace(result)
	return result
}

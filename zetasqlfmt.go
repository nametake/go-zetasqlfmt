package zetasqlfmt

import (
	"fmt"
	"go/ast"
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
	// source, err := os.ReadFile(path)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to read file %s: %v", path, err)
	// }
	//
	// fset := token.NewFileSet()
	//
	// f, err := parser.ParseFile(fset, path, source, parser.ParseComments)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to parse file %s: %v", path, err)
	// }

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
	fmt.Println(pkg.Name)
	// fmt.Printf("%+v\n\n", pkg.TypesInfo.Uses)
	// fmt.Printf("%+v\n\n", pkg.TypesInfo.Types)
	// fmt.Printf("%+v\n\n", pkg.TypesInfo.Defs)
	// fmt.Printf("%+v\n\n", pkg.TypesInfo.Scopes)
	fmt.Printf("%+v\n\n", pkg.Syntax)
	for i, f := range pkg.TypesInfo.Uses {
		fmt.Println("------")
		if f.Type().String() == "cloud.google.com/go/spanner.Statement" {
			fmt.Println("@@@@@@@")
			fmt.Println(i)
		}
	}
	for _, f := range pkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			// fmt.Printf("%+v\n", n)
			switch n := n.(type) {
			case *ast.CompositeLit:
				fmt.Println("AAAAAAAAAAAAAAAAAAAAAA")
				sel, ok := n.Type.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				fmt.Println(pkg.TypesInfo.Uses[sel.Sel])
				x, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}
				fmt.Println(pkg.TypesInfo.Uses[x])
				fmt.Println("BBBBBBBBBBBBBBBBBbb")

			case *ast.Ident:
				//	 fmt.Println("------")
				use, ok := pkg.TypesInfo.Uses[n]
				if !ok {
					return true
				}
				if use.Type().String() == "cloud.google.com/go/spanner.Statement" {
					fmt.Println("@@@@@@@")
					fmt.Println(n.Name)
					fmt.Println("1111111111", use)
				}
				// fmt.Println(n.Name)
				// fmt.Println("3333333333", pkg.TypesInfo.Uses[n])
				// case ast.Expr:
				//	 fmt.Println("------")
				//	 fmt.Println(pkg.TypesInfo.Types[n])
			}
			return true
		})
	}

	return nil, nil
}

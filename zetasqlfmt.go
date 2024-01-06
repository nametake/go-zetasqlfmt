package zetasqlfmt

import (
	"fmt"
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
	for _, pkg := range pkgs {
		fmt.Println("@", pkg.TypesInfo.Uses)
		fmt.Println("@", pkg.Types)
		for k, v := range pkg.TypesInfo.Uses {
			// fmt.Println("KEY", k.Name, "VALUE", v)
			fmt.Println("KEY", k)
			fmt.Println("VALUE", v.Name(), v.Type().Underlying())
		}
	}

	// conf := types.Config{Importer: importer.Default()}

	// info := &types.Info{
	// 	Uses: make(map[*ast.Ident]types.Object),
	// }

	// o := f.Scope.Lookup("Statement")
	// fmt.Println(o)
	// pkg, err := conf.Check(path, fset, []*ast.File{f}, info)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to type check file %s: %v", path, err)
	// }
	//
	// fmt.Println(pkg)

	return nil, nil
}

package zetasqlfmt

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
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
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", path, err)
	}

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, path, source, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %v", path, err)
	}

	conf := types.Config{Importer: importer.Default()}

	info := &types.Info{
		Uses: make(map[*ast.Ident]types.Object),
	}

	pkg, err := conf.Check(path, fset, []*ast.File{f}, info)
	if err != nil {
		return nil, fmt.Errorf("failed to type check file %s: %v", path, err)
	}

	fmt.Println(pkg)

	return nil, nil
}

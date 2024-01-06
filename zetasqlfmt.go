package zetasqlfmt

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
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
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", path, err)
	}

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, path, source, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %v", path, err)
	}

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
	fmt.Println("-------------- Package -------------- ")
	fmt.Println(pkg.Name)
	fmt.Println(f.Name.Name)
	objs := make([]*types.Struct, 0)

	// fmt.Printf("%+v\n", pkg.TypesInfo.Instances)
	// fmt.Printf("%+v\n", pkg.TypesInfo.Selections)
	// fmt.Printf("%+v\n", pkg.TypesInfo.Scopes)
	// fmt.Printf("%+v\n", pkg.TypesInfo.Defs)
	fmt.Printf("%+v\n", pkg.TypesInfo.Uses)
	// fmt.Printf("%+v\n", pkg.TypesInfo.Implicits)

	fmt.Println("-------------- Types -------------- ")
	fmt.Printf("%v\n", pkg.TypesInfo)
	fmt.Printf("%+v\n", pkg.TypesInfo)
	fmt.Printf("%#v\n", pkg.TypesInfo)

	for _, typ := range pkg.TypesInfo.Types {
		fmt.Println("-------------- Type -------------- ")
		fmt.Println(typ)
		fmt.Println(typ.Type.Underlying())

		// if s, ok := typ.Type.Underlying().(*types.TypeAndValue); ok {
		// 	fmt.Println("-------------- Struct -------------- ")
		// }
	}

	for _, selection := range pkg.TypesInfo.Selections {
		fmt.Println("-------------- Selection -------------- ")
		fmt.Println(selection)
	}

	for ident, object := range pkg.TypesInfo.Uses {
		if object.Name() == "SQL" {
			fmt.Println("-------------- SQL -------------- ")
			fmt.Println(ident, object)
			fmt.Println(object.Type().String())
			fmt.Println(object.Pkg())
			fmt.Println(object.Parent())
			fmt.Println(object.Id())
			fmt.Println(object.Id())
		}
		if object.Type().String() == "cloud.google.com/go/spanner.Statement" {
			ast.Inspect(ident, func(n ast.Node) bool {
				fmt.Println("-------------- Node -------------- ")
				fmt.Println(n)
				return true
			})
			if s, ok := object.Type().Underlying().(*types.Struct); ok {
				fmt.Println("Ident:", ident.Obj)
				objs = append(objs, s)
			}
		}
	}

	fmt.Println("-------------- Structs -------------- ")
	fmt.Println(objs)
	for _, obj := range objs {
		sql := obj.Field(0)
		fmt.Println(sql.Type().Underlying())
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

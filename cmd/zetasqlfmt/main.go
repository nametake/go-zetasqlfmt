package main

import (
	"flag"
	"fmt"
	"go/ast"
	"os"
	"sync"

	"github.com/nametake/go-zetasqlfmt"
	"golang.org/x/tools/go/packages"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] directory\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("No directory specified.")
		flag.Usage()
		os.Exit(1)
	}
	dir := args[0]

	if err := run(dir); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(dir string) error {
	waitGroup := sync.WaitGroup{}

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedFiles,
	}
	pkgs, err := packages.Load(cfg, dir)
	if err != nil {
		return fmt.Errorf("failed to load packages: path = %s: %v", dir, err)
	}

	errCount := 0
	format := func(pkg *packages.Package, file *ast.File, wg *sync.WaitGroup) {
		defer func() {
			wg.Done()
		}()

		result, err := zetasqlfmt.Format(pkg, file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		if len(result.Errors) > 0 {
			for _, err := range result.Errors {
				errCount += 1
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
		}
		if !result.Changed {
			return
		}

		if err := os.WriteFile(result.Path, result.Output, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			waitGroup.Add(1)
			go format(pkg, file, &waitGroup)
		}
	}

	waitGroup.Wait()

	if errCount > 0 {
		return fmt.Errorf("failed to format %d files", errCount)
	}

	return nil
}

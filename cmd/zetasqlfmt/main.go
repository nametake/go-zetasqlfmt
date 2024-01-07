package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/nametake/go-zetasqlfmt"
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
	errMsgCh := make(chan *zetasqlfmt.FormatError)
	wg := &sync.WaitGroup{}

	fn := func(path string, ch chan *zetasqlfmt.FormatError, wg *sync.WaitGroup) {
		defer wg.Done()

		result, err := zetasqlfmt.Format(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		if len(result.Errors) > 0 {
			for _, err := range result.Errors {
				ch <- err
			}
		}
		if result.Changed {
			if err := os.WriteFile(path, result.Output, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
		}
	}

	if err := zetasqlfmt.FindGoFiles(dir, func(path string) {
		wg.Add(1)
		go fn(path, errMsgCh, wg)
	}); err != nil {
		return fmt.Errorf("failed to find go files: %v", err)
	}

	wg.Wait()
	close(errMsgCh)

	count := 0
	if err := <-errMsgCh; err != nil {
		count += 1
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	if count > 0 {
		return fmt.Errorf("failed to format %d files", count)
	}

	return nil
}

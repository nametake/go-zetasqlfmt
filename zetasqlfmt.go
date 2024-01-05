package zetasqlfmt

import (
	"fmt"
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

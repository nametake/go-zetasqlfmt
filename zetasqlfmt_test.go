package zetasqlfmt

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFindGoFiles(t *testing.T) {
	expected := []string{
		"testdata/files/dir/file1.go",
		"testdata/files/file1.go",
		"testdata/files/file2.go",
	}

	actuals := make([]string, 0)
	fn := func(path string) {
		actuals = append(actuals, path)
	}

	if err := FindGoFiles("testdata/files", fn); err != nil {
		t.Fatalf("findGoFiles(%q) returned unexpected error: %v", "testdata", err)
	}

	if diff := cmp.Diff(expected, actuals); diff != "" {
		t.Errorf("FindGoFiles(%q) returned unexpected files (-want +got):\n%s", "testdata", diff)
	}
}

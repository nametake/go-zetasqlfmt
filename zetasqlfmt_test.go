package zetasqlfmt

import (
	"os"
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

func TestFormat(t *testing.T) {
	tests := []struct {
		filePath   string
		goldenFile string
		want       *FormatResult
	}{
		{
			filePath:   "testdata/simple.go",
			goldenFile: "testdata/simple_golden.go",
			want: &FormatResult{
				Changed: true,
			},
		},
	}

	for _, test := range tests {
		got, err := Format(test.filePath)
		if err != nil {
			t.Errorf("Format(%q) returned unexpected error: %v", test.filePath, err)
			continue
		}

		golden, err := os.ReadFile(test.goldenFile)
		if err != nil {
			t.Errorf("failed to read golden file %q: %v", test.goldenFile, err)
		}
		if test.want.Changed {
			test.want.Output = golden
		}

		if diff := cmp.Diff(test.want, got); diff != "" {
			t.Errorf("Format(%q) returned unexpected result (-want +got):\n%s", test.filePath, diff)
		}
	}
}

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
	// for cloud.google.com/go/spanner module
	if err := os.Chdir("testdata"); err != nil {
		t.Fatalf("failed to change directory to testdata: %v", err)
	}

	tests := []struct {
		filePath   string
		goldenFile string
		want       *FormatResult
	}{
		{
			filePath:   "simple.go",
			goldenFile: "simple_golden.go",
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

func TestTrimQuotes(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{
		{
			arg:  `"SELECT * FROM TABLE_A;"`,
			want: `SELECT * FROM TABLE_A;`,
		},
		{
			arg:  "`SELECT * FROM TABLE_A;`",
			want: `SELECT * FROM TABLE_A;`,
		},
	}

	for _, test := range tests {
		t.Run(test.arg, func(t *testing.T) {
			got := trimQuotes(test.arg)
			if got != test.want {
				t.Errorf("trimQuotes(%q) = %q, want %q", test.arg, got, test.want)
			}
		})
	}
}

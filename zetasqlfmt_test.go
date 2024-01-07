package zetasqlfmt

import (
	"os"
	"strings"
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
		{
			filePath:   "sprintf.go",
			goldenFile: "sprintf_golden.go",
			want: &FormatResult{
				Changed: true,
			},
		},
		{
			filePath:   "invalid_sql.go",
			goldenFile: "",
			want: &FormatResult{
				Errors: []*FormatError{
					{
						Message: `INVALID_ARGUMENT: Syntax error: Expected end of input but got identifier "FROM_TABLE" [at 1:10]
SELECT * FROM_TABLE;
         ^
Syntax error: Unexpected end of statement [at 1:21]
SELECT * FROM_TABLE;
                    ^`,
						PosText: "invalid_sql.go:9:11",
					},
				},
				Changed: false,
			},
		},
		{
			filePath:   "no_sql.go",
			goldenFile: "",
			want: &FormatResult{
				Changed: false,
			},
		},
	}

	opt := cmp.Comparer(func(x, y *FormatError) bool {
		if len(x.PosText) < len(y.PosText) {
			return x.Message == y.Message && strings.HasSuffix(y.PosText, x.PosText)
		}
		return x.Message == y.Message && strings.HasSuffix(x.PosText, y.PosText)
	})

	for _, test := range tests {
		if test.want.Changed {
			golden, err := os.ReadFile(test.goldenFile)
			if err != nil {
				t.Errorf("failed to read golden file %q: %v", test.goldenFile, err)
			}
			test.want.Output = golden
		}

		got, err := Format(test.filePath)
		if err != nil {
			t.Errorf("Format(%q) returned unexpected error: %v", test.filePath, err)
			continue
		}

		if diff := cmp.Diff(test.want, got, opt); diff != "" {
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

func TestFillFormatVerbs(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{
		{
			arg:  "SELECT * FROM TABLE ORDER BY %s;",
			want: "SELECT * FROM TABLE ORDER BY _DUMMY_STRING_;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY %v;",
			want: "SELECT * FROM TABLE ORDER BY _DUMMY_VALUE_;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY %d;",
			want: "SELECT * FROM TABLE ORDER BY -999;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY %v %v;",
			want: "SELECT * FROM TABLE ORDER BY _DUMMY_VALUE_ _DUMMY_VALUE_;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY %s %s;",
			want: "SELECT * FROM TABLE ORDER BY _DUMMY_STRING_ _DUMMY_STRING_;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY %v %s;",
			want: "SELECT * FROM TABLE ORDER BY _DUMMY_VALUE_ _DUMMY_STRING_;",
		},
	}
	for _, test := range tests {
		t.Run(test.arg, func(t *testing.T) {
			got := fillFormatVerbs(test.arg)
			if got != test.want {
				t.Errorf("fillFormatVerbs(%q) = %q, want %q", test.arg, got, test.want)
			}
		})
	}
}

func TestRestoreFormatVerbs(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{
		{
			arg:  "SELECT * FROM TABLE ORDER BY _DUMMY_STRING_;",
			want: "SELECT * FROM TABLE ORDER BY %s;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY _DUMMY_VALUE_;",
			want: "SELECT * FROM TABLE ORDER BY %v;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY -999;",
			want: "SELECT * FROM TABLE ORDER BY %d;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY _DUMMY_VALUE_ _DUMMY_VALUE_;",
			want: "SELECT * FROM TABLE ORDER BY %v %v;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY _DUMMY_STRING_ _DUMMY_STRING_;",
			want: "SELECT * FROM TABLE ORDER BY %s %s;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY _DUMMY_VALUE_ _DUMMY_STRING_;",
			want: "SELECT * FROM TABLE ORDER BY %v %s;",
		},
	}
	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			got := restoreFormatVerbs(test.arg)
			if got != test.want {
				t.Errorf("restoreFormatVerbs(%q) = %q, want %q", test.arg, got, test.want)
			}
		})
	}
}

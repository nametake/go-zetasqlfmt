package zetasqlfmt

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/go/packages"
)

func TestFormat(t *testing.T) {
	// for cloud.google.com/go/spanner module
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	if err := os.Chdir("testdata"); err != nil {
		t.Fatalf("failed to change directory to testdata: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(currentDir); err != nil {
			t.Fatalf("failed to change directory to %q: %v", currentDir, err)
		}
	})

	testdataDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	addDirPrefix := func(s string) string {
		return fmt.Sprintf("%s/%s", testdataDir, s)
	}

	tests := []struct {
		filePath   string
		goldenFile string
		option     *Option
		want       *FormatResult
	}{
		{
			filePath:   "simple.go",
			goldenFile: "simple_golden.go",
			option:     &Option{NoSemicolon: false},
			want: &FormatResult{
				Path:    addDirPrefix("simple.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			filePath:   "sprintf.go",
			goldenFile: "sprintf_golden.go",
			option:     &Option{NoSemicolon: false},
			want: &FormatResult{
				Path:    addDirPrefix("sprintf.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			filePath:   "backquote.go",
			goldenFile: "backquote_golden.go",
			option:     &Option{NoSemicolon: false},
			want: &FormatResult{
				Path:    addDirPrefix("backquote.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			filePath:   "nosemicolon.go",
			goldenFile: "nosemicolon_golden.go",
			option:     &Option{NoSemicolon: true},
			want: &FormatResult{
				Path:    addDirPrefix("nosemicolon.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			filePath:   "invalid_sql.go",
			goldenFile: "",
			option:     &Option{NoSemicolon: false},
			want: &FormatResult{
				Path:    addDirPrefix("invalid_sql.go"),
				Changed: false,
				Errors: []*FormatError{
					{
						Message: `INVALID_ARGUMENT: Syntax error: Expected end of input but got identifier "FROM_TABLE" [at 1:10]
SELECT * FROM_TABLE;
         ^
Syntax error: Unexpected end of statement [at 1:21]
SELECT * FROM_TABLE;
                    ^`,
						PosText: addDirPrefix("invalid_sql.go:9:11"),
					},
				},
			},
		},
		{
			filePath:   "include_invalid_sql.go",
			goldenFile: "include_invalid_sql_golden.go",
			option:     &Option{NoSemicolon: false},
			want: &FormatResult{
				Path:    addDirPrefix("include_invalid_sql.go"),
				Changed: true,
				Errors: []*FormatError{
					{
						Message: `INVALID_ARGUMENT: Syntax error: Expected end of input but got identifier "FROM_TABLE" [at 1:10]
SELECT * FROM_TABLE;
         ^
Syntax error: Unexpected end of statement [at 1:21]
SELECT * FROM_TABLE;
                    ^`,
						PosText: addDirPrefix("include_invalid_sql.go:9:11"),
					},
				},
			},
		},
		{
			filePath:   "functions.go",
			goldenFile: "functions_golden.go",
			option:     &Option{NoSemicolon: false},
			want: &FormatResult{
				Path:    addDirPrefix("functions.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			filePath:   "undefined_type.go",
			goldenFile: "undefined_type_golden.go",
			option:     &Option{NoSemicolon: false},
			want: &FormatResult{
				Path:    addDirPrefix("undefined_type.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			filePath:   "no_sql.go",
			goldenFile: "",
			option:     &Option{NoSemicolon: false},
			want: &FormatResult{
				Path:    addDirPrefix("no_sql.go"),
				Changed: false,
				Errors:  []*FormatError{},
			},
		},
	}

	for _, test := range tests {
		if test.want.Changed {
			golden, err := os.ReadFile(test.goldenFile)
			if err != nil {
				t.Errorf("failed to read golden file %q: %v", test.goldenFile, err)
			}
			test.want.Output = golden
		}

		cfg := &packages.Config{
			Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedFiles,
		}
		pkgs, err := packages.Load(cfg, test.filePath)
		if err != nil {
			t.Errorf("failed to load packages: path = %s: %v", test.filePath, err)
		}
		if len(pkgs) != 1 {
			t.Errorf("expected exactly one package: %s", test.filePath)
		}

		pkg := pkgs[0]

		if len(pkg.Syntax) != 1 {
			t.Errorf("expected exactly one file: %s", test.filePath)
		}

		file := pkg.Syntax[0]

		got, err := Format(pkg, file, test.option)
		if err != nil {
			t.Errorf("Format(%q) returned unexpected error: %v", test.filePath, err)
			continue
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

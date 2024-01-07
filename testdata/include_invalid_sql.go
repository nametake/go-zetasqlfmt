package testdata

import (
	"cloud.google.com/go/spanner"
)

func Foo() *spanner.Statement {
	return &spanner.Statement{
		SQL:    "SELECT * FROM_TABLE;",
		Params: map[string]interface{}{},
	}
}

func Bar() *spanner.Statement {
	return &spanner.Statement{
		SQL:    "SELECT * FROM TABLE;",
		Params: map[string]interface{}{},
	}
}

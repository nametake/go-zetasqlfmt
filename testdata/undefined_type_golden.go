package testdata

import (
	"cloud.google.com/go/spanner"
)

func Foo() *spanner.Statement {
	var _ FooID
	return &spanner.Statement{
		SQL: `
SELECT
  *
FROM
  TABLE;
`, Params: map[string]interface{}{},
	}
}

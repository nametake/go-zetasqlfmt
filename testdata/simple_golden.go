package simple

import (
	"cloud.google.com/go/spanner"
)

func SQL() *spanner.Statement {
	return &spanner.Statement{
		SQL:    "SELECT * FROM TABLE;",
		Params: map[string]interface{}{},
	}
}

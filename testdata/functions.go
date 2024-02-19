package testdata

import (
	"cloud.google.com/go/spanner"
)

func CurrentDate() *spanner.Statement {
	return &spanner.Statement{
		SQL:    "SELECT CAST(CURRENT_DATE() AS STRING) AS current_date;",
		Params: map[string]interface{}{},
	}
}

func ArraySum() *spanner.Statement {
	return &spanner.Statement{
		SQL:    "SELECT ARRAY_SUM([1, 2, 3, 4, 5, NULL, 4, 3, 2, 1]) as sum;",
		Params: map[string]interface{}{},
	}
}

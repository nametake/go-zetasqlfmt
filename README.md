# go-zetasqlfmt

go-zetasqlfmt is a tool for formatting ZetaSQL within Go code.

It formats the SQL of the `cloud.google.com/go/spanner.Statement` structure using [go-zetasql](https://github.com/goccy/go-zetasql).

## Installation

```console
go install github.com/nametake/go-zetasqlfmt/cmd/zetasqlfmt@latest
```

If you don't want to see warning logs, you can suppress them by specifying CXXFLAGS as follows:

```console
CGO_CXXFLAGS="$(go env CGO_CXXFLAGS) -Wno-deprecated" go install github.com/nametake/go-zetasqlfmt/cmd/zetasqlfmt@latest
```

If you use Docker, you can use the following commands without installing zetasqlfmt:

```console
docker run --rm -v ".:/app" ghcr.io/nametake/go-zetasqlfmt:main ./...
```

## Usage

You can format ZetaSQL by running the command within a project containing `cloud.google.com/go/spanner.Statement`.

Example code to be formatted:

```go
package testdata

import (
	"cloud.google.com/go/spanner"
)

func Foo() *spanner.Statement {
	return &spanner.Statement{
		SQL:    "SELECT * FROM TABLE;",
		Params: map[string]interface{}{},
	}
}
```

Command execution:

```console
zetasqlfmt ./...
```

Formatted code:

```go
package testdata

import (
	"cloud.google.com/go/spanner"
)

func Foo() *spanner.Statement {
	return &spanner.Statement{
		SQL: `
SELECT
  *
FROM
  TABLE;
`, Params: map[string]interface{}{},
	}
}
```

## Options

```console
-nosemicolon
      no semicolon
```

## Formatting Specifications

- [go-zetasql](https://github.com/goccy/go-zetasql) is used for SQL formatting.
- If backticks `` ` `` are present in the formatted SQL, it removes line breaks and compresses multiple spaces into a single space.
- If backticks `` ` `` are not present but line breaks are, it wraps the SQL with `` ` `` and line breaks.
- If the SQL is formatted using fmt.Sprintf, it replaces it with dummy values before formatting.

In the case of dynamically using ORDER BY with `cloud.google.com/go/spanner`,
a [method using fmt.Sprintf](https://github.com/googleapis/google-cloud-go/issues/6496) has been suggest.

Therefore, in go-zetasqlfmt, in cases like the one below, it replaces format verbs with dummy values before formatting:

```go
func Foo() *spanner.Statement {
	return &spanner.Statement{
		SQL:    fmt.Sprintf("SELECT * FROM TABLE ORDER BY %s;", "CreatedAt"),
		Params: map[string]interface{}{},
	}
}
```

The SQL is transformed as follows during formatting:

```
SELECT * FROM TABLE ORDER BY _DUMMY_STRING_;
```

The conversion table is as follows:

| Verbs | Dummy String     |
| ---   | ---              |
| `%s`  | `_DUMMY_STRING_` |
| `%v`  | `_DUMMY_VALUE_`  |
| `%d`  | `-999`           |

The replaced dummy values are then converted back to the original format verbs after formatting.
Therefore, if the original SQL contains strings from the conversion table, unintended behavior may occur.

### Links

- [go-zetasqlfmt-action](https://github.com/nametake/go-zetasqlfmt-action)

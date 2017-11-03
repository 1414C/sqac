# sqac


sqac is a simple overlay to provide a common interface to attached mssql, mysql, postgres or sqlite databases.

- create tables, supporting default, nullable, start, primary-key, index tags
- drop tables
- destructive rest of tables
- create indexes
- drop indexes
- alter tables
- add columns
- set sequence, auto-increment or identity nextval
- Standard go sql, j moirons sqlx db access

* Testing

```bash

go test -v -l -db <dbtype> sqac_test.go

```

Postgres:
```bash

$ go test -v -db pg sqac_test.go

```

MySQL:
```bash

$ go test -v -db mysql sqac_test.go

```

MSSQL:
```bash

$ go test -v -db mssql sqac_test.go

```

SQLite:
```bash

$ go test -v -db sqlite sqac_test.go

```

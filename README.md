# sqac


sqac is a simple overlay to provide a common interface to attached mssql, mysql, postgres, sqlite or SAP Hana databases.

- create tables, supporting default, nullable, start, primary-key, index tags
- drop tables
- destructive reset of tables
- create indexes
- drop indexes
- alter tables via column, index and sequence additions
- set sequence, auto-increment or identity nextval
- Standard go sql, jmoirons sqlx db access
- generic CRUD entity operations

* Testing

```bash

go test -v -l -db <dbtype> sqac_test.go

```

Postgres:
```bash

$ go test -v -db postgres sqac_test.go

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


Autoincrement fields should be designated as rgen:"primary_key:inc"

- [ ]SQLite stores timestamps as UTC, so clients would need to convert back to the local timezone on a read.
- [ ]Consider saving all time as UTC
- [ ]Consider converting all time reads as Local
- [ ]This is not perfect, as hand-written SQL will not pass the requests through the CrudInfo conversions.  Problem.
- [ ]HDB ExistsTable should include SCHEMA field in selection?
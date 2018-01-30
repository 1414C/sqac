# sqac

sqac is a simple overlay to provide a common interface to attached mssql, mysql, postgres, sqlite or SAP Hana database.

- create tables, supporting default, nullable, start, primary-key, index tags
- drop tables
- destructive reset of tables
- create indexes
- drop indexes
- alter tables via column, index and sequence additions
- set sequence, auto-increment or identity nextval
- Standard go sql, jmoirons sqlx db access
- generic CRUD entity operations
- set commands (/$count /$orderby=<field_name> $limit=n; $offset=n; ($asc|$desc))

* passing: pg, mssql, mysql, hdb, sqlite
* refactor non-indempotent SQLite Foreign-Key test to use a closure
* consider parsing the stored create schema when adding / dropping a foreign-key on SQLite tables (== dangerous)
* add cascade to Drops?
* refactor foreign-key create/drop

* Testing / TODO
* examine the $desc orderby when limit / offset is used in postgres with selection parameter (weirdness)
* change from timestamp with TZ to timestamp and ensure timestamps are in UTC before submitting to the db
* examine view support
* remove debugging output
* remove extraneaous getSet-type methods
* correct CreateTables where no primary-key is present (trailing ,_ )

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

SAP Hana:
```bash

$ go test -v -db hdb sqac_test.go

```

- [x]Support unique constraint on single-fields
- [ ]Support unique constraines on grouped fields(?)
- [ ]Other constraints(?)
- [x]Complete sql/sqlx query/exec wrapper tests
- [x]Autoincrement fields should be designated as sqac:"primary_key:inc"
- [ ]SQLite stores timestamps as UTC, so clients would need to convert back to the local timezone on a read.
- [ ]Consider saving all time as UTC
- [ ]Consider converting all time reads as Local
- [x]This is not perfect, as hand-written SQL will not pass the requests through the CrudInfo conversions.
- [ ]HDB ExistsTable should include SCHEMA field in selection?
- [x]Really consider what to do with nullable fields
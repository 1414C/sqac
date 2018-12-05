# sqac

sqac is a simple overlay to provide a common interface to an attached mssql, mysql, postgres, sqlite or SAP Hana database.

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
- comprehensive test cases

* passing: pg, mssql, mysql, hdb, sqlite
* refactor non-indempotent SQLite Foreign-Key test to use a closure
* consider parsing the stored create schema when adding / dropping a foreign-key on SQLite tables (dangerous?)
* add cascade to Drops?

* Testing / TODO
* examine the $desc orderby when limit / offset is used in postgres with selection parameter (odd)
* change from timestamp with TZ to timestamp and ensure timestamps are in UTC before submitting to the db
* examine view support
* remove extraneous getSet-type methods

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
- [ ]Support unique constraints on grouped fields(?)
- [x]Complete sql/sqlx query/exec wrapper tests
- [x]Auto-increment fields should be designated as sqac:"primary_key:inc"
- [ ]SQLite stores timestamps as UTC, so clients would need to convert back to the local timezone on a read.
- [x]Consider saving all time as UTC
- [ ]Consider converting all time reads as Local
- [x]This is not perfect, as hand-written SQL will not pass the requests through the CrudInfo conversions.
- [ ]HDB ExistsTable should include SCHEMA field in selection?
- [x]Really consider what to do with nullable fields
- [ ]It would be nice to replace the fmt.Sprintf(...) calls in the DDL and DML constructions with inline strconv.XXXX, as the performance seems to be 2-4x better.  oops.  In practical terms we are dealing with 10's of ns here, but under high load it could be a thing.  Consider doing this when implementing DB2 support.

<br></br>
## sqac Quickstart Example

The following example illustrates the general usage form of the sqac library:

```golang

package main

import (
	"flag"
	"fmt"

	"github.com/1414C/sqac"
	// "github.com/1414C/sqac/common"
	_ "github.com/SAP/go-hdb/driver"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

    dbFlag := flag.String("db", "sqlite", "db-type for connection")
    csFlag := flag.String("cs", "testdb.sqlite", "connection-string for the database")
	logFlag := flag.Bool("l", false, "activate sqac detail logging to stdout")
	dbLogFlag := flag.Bool("dbl", false, "activate DDL/DML logging to stdout)")
	flag.Parse()

	// This will be the central access-point to the ORM and should be made
	// available in all locations where access to the persistent storage
	// (database) is required.
	var (
		Handle sqac.PublicDB
	)

	// Create a PublicDB instance.  Check the Create method, as the return parameter contains
	// not only an implementation of PublicDB targeting the db-type/db, but also a pointer
	// facilitating access to the db via jmoiron's sqlx package.
	Handle = sqac.Create(*dbFlag, *logFlag, *dbLogFlag, *cs)

	// Execute a call to get the name of the db-driver being used.  At this point, any method
	// contained in the sqac.PublicDB interface may be called.
	driverName := Handle.GetDBDriverName()
	fmt.Println("driverName:", driverName)

	// Close the connection.
	Handle.Close()
}

```

Execute the test program as follows using sqlite.  Note that the sample program makes no
effort to validate the flag parameters.

```bash

  go run -db sqlite -cs testdb.sqlite main.go

```


## Connection Strings
sqac presently supports MSSQL, MySql, PostgreSQL, Sqlite3 and the SAP Hana database.  You will
need to know the db user-name / password, as well as the address:port and name of the database.

### MSSQL Connection String

```golang

  cs := "sqlserver://SA:my_passwd@localhost:1401?database=my_dbname"

```

### MySQL Connection String

```golang

  cs := "my_uname:my_passwd@tcp(192.168.1.10:3306)/my_dbname?charset=utf8&parseTime=True&loc=Local"

```

### PostgreSQL Connection String

```golang

  cs := "host=127.0.0.1 user=my_uname dbname=my_dbname sslmode=disable password=my_passwd"

```

### Sqlite3 Connection String

```golang

  cs := "my_db_file.sqlite"

  // or

  cs = "my_db_file.db"

```


### SAP Hana Connection String

```golang

  cs := "hdb://my_uname:my_passwd@192.168.111.45:30015"

```

## Table Declarations

sqac table-declarations are informed by go structs with json-style tags indicating
column attributes.  A complete list of tags/column attributes follows:

|  Database               | JSON Value for db_dialect field    |
|-------------------------|------------------------------------|
| Postgres                | "db_dialect": "postgres"           |
| MSSQL (2008+)           | "db_dialect": "mssql"              |
| SAP Hana                | "db_dialect": "hdb"                |
| SQLite3                 | "db_dialect": "sqlite3"            |
| MySQL / MariaDB         | "db_dialect": "mysql"              |



## Table DDL-Type Operations

A small example declaring sqac table 'depot' follows:

```golang

    type Depot struct {
        DepotNum   int       `db:"depot_num" sqac:"primary_key:inc"`
        CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now();"`
        Region     string    `db:"region" sqac:"nullable:false;default:YYC"`
        Province   string    `db:"province" sqac:"nullable:false;default:AB"`
        Country    string    `db:"country" sqac:"nullable:false;default:CA"`
        }

```

The 'db:' tag is used to declare the column name in the database, while the 'sqac:' tag
is used to inform the sqac runtime how to setup the column attributes in the database.

A breakdown of the column attributes follows:

'depot_num' is declared as an auto-incrementing primary key called starting at 0.  
'create_date' is declared as a non-nullable column using sqac date function 'now()' as a default value.
'region' is declared as a non-nullable column with a default value of "YYC".
'province' is declared as a non-nullable column with a default value of "AB".
'country' is declared as a non-nullable column with a default value of "CA".

An excerpt from 'a_sqac_test.go' illustrates how the sqac method PublicDB.CreateTables
is used to create new tables in the database:

```golang

// TestCreateTableBasic
//
// Create table depot via CreateTables(i ...interface{})
// Verify table creation via ExistsTable(tn string)
// Perform negative validation be checking for non-existant
// table "table_name" via ExistsTable(tn string)
//
func TestCreateTableBasic(t *testing.T) {

    type Depot struct {
        DepotNum   int       `db:"depot_num" sqac:"primary_key:inc"`
        CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now();"`
        Region     string    `db:"region" sqac:"nullable:false;default:YYC"`
        Province   string    `db:"province" sqac:"nullable:false;default:AB"`
        Country    string    `db:"country" sqac:"nullable:false;default:CA"`
    }

    err := Handle.CreateTables(Depot{})
    if err != nil {
        t.Errorf("%s", err.Error())
    }

    // determine the table name as per the table creation logic
    tn := common.GetTableName(Depot{})     // "depot"

    // expect that table depot exists
    if !Handle.ExistsTable(tn) {
        t.Errorf("table %s was not created", tn)
    }
}

```

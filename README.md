# sqac

Sqac is a simple overlay to provide a common interface to an attached mssql, mysql, postgres, sqlite or SAP Hana database.

- create tables, supporting default, nullable, start, primary-key, index tags
- drop tables
- destructive reset of tables
- create indexes
- drop indexes
- alter tables via column, index and sequence additions
- set sequence, auto-increment or identity nextval
- supports db access through standard go sql drivers and jmoirons sqlx package
- generic CRUD entity operations
- UTC timestamps used internally for all time types
- set commands (/$count /$orderby=<field_name> $limit=n; $offset=n; ($asc|$desc))
- comprehensive test cases

## Outstanding TODO's

- [ ] refactor to fold the larger methods down to a more readable and reasonable size.  The buildTablSchema methods are monolithic blocks b/c at the time I was thinking of function-call cost and the uncertainty I had around the way the compiler inlines.
- [ ] refactor non-idempotent SQLite Foreign-Key test to use a closure
- [ ] consider parsing the stored create schema when adding / dropping a foreign-key on SQLite tables
- [ ] add cascade to Drops?
- [ ] examine the $desc orderby when limit / offset is used in postgres with selection parameter (odd)
- [ ] change from timestamp with TZ to timestamp and ensure timestamps are in UTC before submitting to the db
- [ ] examine view support
- [ ] consider the consumption of SAP CDS
- [ ] remove extraneous getSet-type methods
- [ ] ProcessSchema does not return an error; ProcessTransaction does?  Noticed this in DropIndex.  Inconsistent.
- [ ] Support unique constraints on grouped fields(?)
- [ ] Consider an option where all time reads are returned as Local
- [ ] HDB ExistsTable should include SCHEMA field in selection?
- [ ] It would be nice to replace the fmt.Sprintf(...) calls in the DDL and DML constructions with inline strconv.XXXX.  In practical terms we are dealing with 10's of ns here, but it could be a thing.  Consider doing this when implementing DB2 support.

<br>

## Installation

Install sqac via go get:

```bash
go get -u github.com/sqac
```

<br>

Ensure that you have also installed the drivers for the databases you plan to use.  Supported drivers include:

| Driver Name               | Driver Location                   |
|---------------------------|-----------------------------------|
|SAP Hana Database Driver   | github.com/SAP/go-hdb/driver      |
|MSSQL Database Driver      | github.com/denisenkom/go-mssqldb  |
|MySQL Database Driver      | github.com/go-sql-driver/mysql    |
|PostgreSQL Database Driver | github.com/lib/pq                 |
|SQLite3 Database Driver    | github.com/mattn/go-sqlite3       |

<br>

Verify the installation by running the included test suite against sqlite.  Test execution will create a 'testdb.sqlite' database file in the sqac directory.  The tests are not entirely idempotent and the testdb.sqlite file will not be cleaned up.  This is by design as the tests were used for debugging purposes during the development.  It would be a simple matter to tidy this up.

```bash
go test -v -db sqlite
```

If testing against sqlite is not an option, the test suite may be run against any of the supported database systems.  When running against a non-sqlite db, a connection string must be supplied via the *cs* flag.  See the [Database Connection Strings](https://1414c.github.io/sqac/getting-started/gs-content-e/) section in the documentation for database-specific connection string formats.  As an example, a postgres test could be run using the following command string:

```bash
go test -v -db postgres -cs "host=127.0.0.1 user=my_uname dbname=my_dbname sslmode=disable password=my_passwd"
```

<br>

## Accessing Documentation

The API is somewhat documented via comments in the code, but it is best to use the [official sqac documentation](https://1414c.github.io/sqac/).  If code comments are more to your liking, run the *godoc* command as shown below:

```bash
godoc -http=:6061
```

Once the *godoc* server has started, hit http://localhost:6061/pkg/github.com/1414C/sqac/ for sqac API documentation.

<br>

## Quickstart

The following example illustrates the general usage of the sqac library.  

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

  // valid dbFlag values: {hdb, sqlite, mssql, mysql, postgres}
  dbFlag := flag.String("db", "sqlite", "db-type for connection")
  // see ConnectionStrings in this document for valid csFlag value formats
  csFlag := flag.String("cs", "testdb.sqlite", "connection-string for the database")
  // the logging is verbose and targetted at debugging
  logFlag := flag.Bool("l", false, "activate sqac detail logging to stdout")
  // the db logging provides a close approximation to the commands issued to the db
  dbLogFlag := flag.Bool("dbl", false, "activate DDL/DML logging to stdout)")
  flag.Parse()

  // This will be the central access-point to the ORM and should be made
  // available in all locations where access to the persistent storage
  // (database) is required.
  var (
    Handle sqac.PublicDB
  )

  // Declare a struct to use as a source for table declaration.
  type Depot struct {
      DepotNum   int       `db:"depot_num" sqac:"primary_key:inc"`
      CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now();"`
      Region     string    `db:"region" sqac:"nullable:false;default:YYC"`
      Province   string    `db:"province" sqac:"nullable:false;default:AB"`
      Country    string    `db:"country" sqac:"nullable:false;default:CA"`
  }

  // Create a PublicDB instance.  Check the Create method, as the return parameter contains
  // not only an implementation of PublicDB targeting the db-type/db, but also a pointer
  // facilitating access to the db via jmoiron's sqlx package.  This is useful if you wish
  // to access the sql/sqlx APIs directly.
  Handle = sqac.Create(*dbFlag, *logFlag, *dbLogFlag, *cs)

  // Execute a call to get the name of the db-driver being used.  At this point, any method
  // contained in the sqac.PublicDB interface may be called.
  driverName := Handle.GetDBDriverName()
  fmt.Println("driverName:", driverName)

  // Create a new table in the database
  err := Handle.CreateTables(Depot{})
  if err != nil {
    t.Errorf("%s", err.Error())
  }

  // Determine the table name as per the table creation logic
  tn := common.GetTableName(Depot{})

  // Expect that table depot exists
  if !Handle.ExistsTable(tn) {
    t.Errorf("table %s was not created", tn)
  }

  // Drop the table
  err = Handle.DropTables(Depot{})
  if err != nil {
    t.Errorf("table %s was not dropped", tn)
  }

  // Close the connection.
  Handle.Close()
}
```

Execute the sample program as follows using sqlite.  Note that the sample program makes no
effort to validate the flag parameters.

```bash
go run -db sqlite -cs testdb.sqlite main.go
```

<br>

## Sqac makes use of

- [sqlx](https://jmoiron.github.io/sqlx/)
- [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql/)
- [lib/pq](https://github.com/lib/pq)
- [go-sqlite3](http://mattn.github.io/go-sqlite3/)
- [go-mssqldb](https://github.com/denisenkom/go-mssqldb)
- [go-hdb](https://github.com/SAP/go-hdb)
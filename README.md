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
go test -v -db postgres sqac_test.go
```

MySQL:

```bash
go test -v -db mysql sqac_test.go
```

MSSQL:

```bash
go test -v -db mssql sqac_test.go
```

SQLite:

```bash
go test -v -db sqlite sqac_test.go
```

SAP Hana:

```bash
go test -v -db hdb sqac_test.go
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

## Installation

Install sqac via go get:
```bash
go get -u github.com/sqac
```
<br>
Ensure that you have also installed the drivers for the databases you plan to use.  Supported drivers:

| Driver Name               | Driver Location                   |
|---------------------------|-----------------------------------|
|SAP Hana Database Driver   | github.com/SAP/go-hdb/driver      |
|MSSQL Database Driver      | github.com/denisenkom/go-mssqldb  |
|MySQL Database Driver      | github.com/go-sql-driver/mysql    |
|PostgreSQL Database Driver | github.com/lib/pq                 |
|SQLite3 Database Driver    | github.com/mattn/go-sqlite3       |
<br>
Verify the installation by running the included test suite against sqlite.  Test execution will create a 'testdb.sqlite' database file in the sqac directory.  The tests are not entirely idempotent and the testdb.sqlite file will not be cleaned up.  This is by design as the tests were used
for debugging purposes during the development.  It would be a simple matter to tidy this up.

```bash
go test -v -db sqlite
```

If running against sqlite is not an option, the test suite may be run against any of the supported database systems.  When running against a non-sqlite db, a connection string must be supplied via the *cs* flag.  See the Connection Strings section for database-specific connection string formats.

```bash
go test -v -db pg -cs "host=127.0.0.1 user=my_uname dbname=my_dbname sslmode=disable password=my_passwd"
```

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

## Connection Strings
sqac presently supports MSSQL, MySQL, PostgreSQL, Sqlite3 and the SAP Hana database.  You will
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
<br>

## Table Declarations

sqac table-declarations are informed by go structs with json-style struct-tags indicating
column attributes.  Two tags are used: 'db:' and 'sqac:'.

The 'db:' tag is used to declare the database column name.  This is typically the snake_case
conversion of the go struct field-name.

The 'sqac:' tag us used to declare column attributes.  A list of the supported attributes
follows:

|  sqac tag               | Description                        |
|-------------------------|------------------------------------|
| **"primary_key:"**  | This tag is used to declare that the specified column should be used as a primary-key in the generated database table.   There are a few variations in its use:  <br><br> **"primary_key:inc"** declares the primary-key as auto-incrementing from 0 in the database table schema: <br> `db:"depot_num" sqac:"primary_key:inc"` <br><br>**"primary_key:"** declares the primary-key as a non-auto-incrementing primary-key in the database schema: <br> `db:"depot_num" sqac:"primary_key:"` <br><br>  **"primary_key:inc;start:90000000"** declares the primary-key as auto-incrementing starting from 900000000: <br> `db:"depot_num" sqac:"primary_key:inc;start:90000000"` <br><br> It is possible to assign the "primary_key:" tag to more than one column in a table's model declaration.  The column containing the first occurrence of the tag (top-down) will be created as the actual primary-key in the database.  The collection of column declarations containing the "primary_key:" tag will be used to create a unique index on the DB table.  This is useful in cases where one is migrating data from a source system that has the concept of compound table keys.  For example, the following model excerpt would result in the creation of "depot_num" as the table's primary-key as well as the creation of a unique index containing "depot_num", "depot_bay", "create_date": <br><br> DepotNum            int       `db:"depot_num" sqac:"primary_key:inc;start:90000000"`<br> DepotBay            int       `db:"depot_bay" sqac:"primary_key:"` <br> CreateDate          time.Time `db:"create_date" sqac:"nullable:false;default:now();index:unique"`<br><br>**Notes:** auto-incrementing primary-keys increment by 1 and must always be declared as go-type **int**.|
| **"nullable:"**         | This tag is used to declare that the specified column is nullable in the database. <br>  Allowed values are *true* or *false*. <br> `db:"region" sqac:"nullable:false"` or <br>  `db:"region" sqac:"nullable:true"` <br><br> **Notes:** If this tag is not specified, the column is declared as nullable with the exception of columns declared with the "primary_key:" tag.              |
| **"default:**           | The "default:" tag is used to declare a default value for the column in the database table schema.  Default values are used as per the implementation of the SQL DEFAULT keyword in the target DBMS. <br>  `db:"region" sqac:"nullable:false;default:YYC"` <br><br> **Notes:** This tag supports the use of static values for all column-types, as well as a small set of date-time functions: "default:now()" / "default:make_timestamptz(9999, 12, 31, 23, 59, 59.9)" / "default:eot()"|
| **"index:"**            | Single column indexes can be declared via the "index:" tag.  The example index declarations require only the "index:unique / non-unique" pair in the column's sqac-tag.  The following column declaration results in the creation of a unique index on table column "create_date": <br> `db:"create_date" sqac:"nullable:false;default:now();index:unique"` <br><br>  A non-unique single column index for the same column is declared as follows: <br> `db:"create_date" sqac:"nullable:false;default:now();index:non-unique"` <br><br><br> Multi-column indexes can also be declared via the index tag.  The following example illustrates the declaration of a non-unique two-column index containing columns "new_column1" and "new_column2": <br> `db:"new_column1" sqac:"nullable:false;index:idx_depot_new_column1_new_column2"` <br> `db:"new_column2" sqac:"nullable:false;default:0;index:idx_depot_new_column1_new_column2"` <br>|
| **"fkey:**   | Foreign-keys can be declared between table columns.  The following example results in the creation of a foreign-key between the table's "warehouse_id" column and reference column "id" in table "warehouse". <br><br> WarehouseID uint64 `db:"warehouse_id" json:"warehouse_id" sqac:"nullable:false;fkey:warehouse(id)"` <br><br> This example is not very clear - see the full example code excerpt in the Table Declaration Examples section.     |
| Non-persistent column   | There are scenarios where model columns may not be persisted in the database.  If a column is to be determined at runtime by the consuming application (for example), the following syntax may be used: <br> NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`              |
<br>

### Simple Table Declaration

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
<br><br>

### Comprehensive Table Declaration

A more comprehensive declaration of table "depot":

```golang
type DepotCreate struct {
  DepotNum            int       `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
  DepotBay            int       `db:"depot_bay" sqac:"primary_key:"`
  CreateDate          time.Time `db:"create_date" sqac:"nullable:false;default:now();index:unique"`
  Region              string    `db:"region" sqac:"nullable:false;default:YYC"`
  Province            string    `db:"province" sqac:"nullable:false;default:AB"`
  Country             string    `db:"country" sqac:"nullable:true;default:CA"`
  NewColumn1          string    `db:"new_column1" sqac:"nullable:false"`
  NewColumn2          int64     `db:"new_column2" sqac:"nullable:false"`
  NewColumn3          float64   `db:"new_column3" sqac:"nullable:false;default:0.0"`
  IntDefaultZero      int       `db:"int_default_zero" sqac:"nullable:false;default:0"`
  IntDefault42        int       `db:"int_default42" sqac:"nullable:false;default:42"`
  FldOne              int       `db:"fld_one" sqac:"nullable:false;default:0;index:idx_depotcreate_fld_one_fld_two"`
  FldTwo              int       `db:"fld_two" sqac:"nullable:false;default:0;index:idx_depotcreate_fld_one_fld_two"`
  TimeCol             time.Time `db:"time_col" sqac:"nullable:false"`
  TimeColNow          time.Time `db:"time_col_now" sqac:"nullable:false;default:now()"`
  TimeColEot          time.Time `db:"time_col_eot" sqac:"nullable:false;default:eot"`
  IntZeroValNoDefault int       `db:"int_zero_val_no_default" sqac:"nullable:false"`
  NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
}
```

### Table Declaration Using Nested Structs

Table declarations may also contain nested structs:

```golang
type Triplet struct {
  TripOne   string `db:"trip_one" sqac:"nullable:false"`
  TripTwo   int64  `db:"trip_two" sqac:"nullable:false;default:0"`
  Tripthree string `db:"trip_three" sqac:"nullable:false"`
}

type Equipment struct {
  EquipmentNum   int64     `db:"equipment_num" sqac:"primary_key:inc;start:55550000"`
  ValidFrom      time.Time `db:"valid_from" sqac:"primary_key;nullable:false;default:now()"`
  ValidTo        time.Time `db:"valid_to" sqac:"primary_key;nullable:false;default:make_timestamptz(9999, 12, 31, 23, 59, 59.9)"`
  CreatedAt      time.Time `db:"created_at" sqac:"nullable:false;default:now()"`
  InspectionAt   time.Time `db:"inspection_at" sqac:"nullable:true"`
  MaterialNum    int       `db:"material_num" sqac:"index:idx_material_num_serial_num"`
  Description    string    `db:"description" sqac:"sqac:nullable:false"`
  SerialNum      string    `db:"serial_num" sqac:"index:idx_material_num_serial_num"`
  Triplet        // structs can be nested to any level
}
```

### Table Declaration With Foreign-Key

```golang
type Warehouse struct {
    ID       uint64 `db:"id" json:"id" sqac:"primary_key:inc;start:40000000"`
    City     string `db:"city" json:"city" sqac:"nullable:false;default:Calgary"`
    Quadrant string `db:"quadrant" json:"quadrant" sqac:"nullable:false;default:SE"`
}

type Product struct {
    ID          uint64 `db:"id" json:"id" sqac:"primary_key:inc;start:95000000"`
    ProductName string `db:"product_name" json:"product_name" sqac:"nullable:false;default:unknown"`
    ProductCode string `db:"product_code" json:"product_code" sqac:"nullable:false;default:0000-0000-00"`
    UOM         string `db:"uom" json:"uom" sqac:"nullable:false;default:EA"`
    WarehouseID uint64 `db:"warehouse_id" json:"warehouse_id" sqac:"nullable:false;fkey:warehouse(id)"`
}
```

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

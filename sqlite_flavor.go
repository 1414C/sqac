package sqac

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// SQLiteFlavor is a sqlite3-specific implementation.
// Methods defined in the PublicDB interface of struct-type
// BaseFlavor are called by default for SQLiteFlavor. If
// the method as it exists in the BaseFlavor implementation
// is not compatible with the schema-syntax required by
// SQLite, the method in question may be overridden.
// Overriding (redefining) a BaseFlavor method may be
// accomplished through the addition of a matching method
// signature and implementation on the SQLiteFlavor
// struct-type.
type SQLiteFlavor struct {
	BaseFlavor

	//================================================================
	// possible local SQLite-specific overrides
	//================================================================
	// GetDBDriverName() string
	// CreateTables(i ...interface{}) error
	// DropTables(i ...interface{}) error
	// AlterTables(i ...interface{}) error
	// ExistsTable(i interface{}) bool
	// ExistsColumn(tn string, cn string, ct string) bool
	// CreateIndex(tn string, in string) error
	// DropIndex(tn string, in string) error
	// ExistsIndex(tn string, in string) bool
	// CreateSequence(sn string, start string) error
	// DropSequence(sn string) error
	// ExistsSequence(sn string) bool
}

// CreateTables creates tables on the sqlite3 database referenced
// by slf.DB.
func (slf *SQLiteFlavor) CreateTables(i ...interface{}) error {

	for t, ent := range i {

		ftr := reflect.TypeOf(ent)
		if slf.log {
			fmt.Println("CreateTable() entity type:", ftr)
		}

		// determine the table name
		tn := reflect.TypeOf(i[t]).String() // models.ProfileHeader{} for example
		if strings.Contains(tn, ".") {
			el := strings.Split(tn, ".")
			tn = strings.ToLower(el[len(el)-1])
		} else {
			tn = strings.ToLower(tn)
		}
		if tn == "" {
			return fmt.Errorf("unable to determine table name in slf.CreateTables")
		}

		// if the table is found to exist, skip the creation
		// and move on to the next table in the list.
		if slf.ExistsTable(tn) {
			if slf.log {
				fmt.Printf("CreateTable - table %s exists - skipping...\n", tn)
			}
			continue
		}

		tc := slf.buildTablSchema(tn, i[t])
		if slf.log {
			fmt.Println("====================================================================")
			fmt.Println("TABLE SCHEMA:", tc.tblSchema)
			fmt.Println()
			for _, v := range tc.seq {
				fmt.Println("SEQUENCE:", v)
			}
			fmt.Println()
			for k, v := range tc.ind {
				fmt.Printf("INDEX: k:%s	fields:%v  unique:%v tableName:%s\n", k, v.IndexFields, v.Unique, v.TableName)
			}
			fmt.Println()
			fmt.Println("PRIMARY KEYS:", tc.pk)
			fmt.Println()
			for _, v := range tc.flDef {
				fmt.Printf("FIELD DEF: fname:%s, ftype:%s, gotype:%s \n", v.FName, v.FType, v.GoType)
				for _, p := range v.RgenPairs {
					fmt.Printf("FIELD PROPERTY: %s, %v\n", p.Name, p.Value)
				}
				fmt.Println("------")
			}
			fmt.Println()
			fmt.Println("ERROR:", tc.err)
			fmt.Println("====================================================================")
		}
		slf.db.MustExec(tc.tblSchema)
		for _, sq := range tc.seq {
			start, _ := strconv.Atoi(sq.Value)
			slf.AlterSequenceStart(sq.Name, start)
		}
		for k, in := range tc.ind {
			slf.CreateIndex(k, in)
		}
	}
	return nil
}

// buildTableSchema builds a CREATE TABLE schema for the SQLite DB
// and returns it to the caller, along with the components determined from
// the db and rgen struct-tags.  this method is used in CreateTables
// and AlterTables methods.
func (slf *SQLiteFlavor) buildTablSchema(tn string, ent interface{}) TblComponents {

	qt := slf.GetDBQuote()
	pKeys := ""
	var sequences []RgenPair
	indexes := make(map[string]IndexInfo)
	tableSchema := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s%s%s (", qt, tn, qt)

	// get a list of the field names, go-types and db attributes.
	// TagReader is a common function across db-flavors. For
	// this reason, the db-specific-data-type for each field
	// is determined locally.
	fldef, err := TagReader(ent, nil)
	if err != nil {
		panic(err)
	}

	// set the SQLite field-types and build the table schema,
	// as well as any other schemas that are needed to support
	// the table definition. In all cases any foreign-key or
	// index requirements must be deferred until all other
	// artifacts have been created successfully.
	// SQLite has basic types more along the lines of Postgres.
	// int64   bigint
	// uint64  bigint
	// int32   integer
	// int16   integer
	// int8    integer
	// int     integer
	// uint32  integer
	// uint16  integer
	// uint8   integer
	// uint	   integer
	// byte    integer
	// rune    integer
	// float32  real
	// float64  real
	// string  varchar(255) (or 256?)
	// date/time  datetime - need now() function equivalent - also function to create a date-time from a string

	// support composite primary-keys (sort-of) by relying on the ROWID
	// property of SQLite to autoincrement the one-and-only PRIMARY KEY field
	// at time of INSERT.  Twice as fast as AUTOINCREMENT.  Use UNIQUE constraint
	// and NOT NULL to artificially create the equivalient of a composite PRIMARY
	// KEY.
	// For example:
	// CREATE TABLE IF NOT EXISTS "DoubleKey4" (
	// 	"KeyOne" integer PRIMARY KEY,
	// 	"KeyTwo" integer NOT NULL,
	// 	"Description" VARCHAR(255),
	// 	UNIQUE("KeyOne", "KeyTwo") );

	//========================================================================================================
	// DROP INDEX IF EXISTS idx_double_key4_new_column3;

	// DROP TABLE IF EXISTS "DoubleKey4";

	// CREATE TABLE IF NOT EXISTS "DoubleKey4" (
	// "KeyOne" integer PRIMARY KEY AUTOINCREMENT,
	// "KeyTwo" integer NOT NULL,
	// "CreateDTUTC" datetime DEFAULT (datetime('now')),
	// "ExpiryDT" datetime DEFAULT(datetime('now','+2 years')),
	// "EOTDT" datetime DEFAULT('9999-12-31 23:59:59'),
	// "Description" VARCHAR(255),
	// "DefaultedText" VARCHAR(255) DEFAULT 'fiddlesticks',
	// "DefaultedFloat" real NOT NULL DEFAULT 4.335,
	// UNIQUE("KeyOne", "KeyTwo") );

	// INSERT OR FAIL INTO "DoubleKey4" (KeyTwo, Description) VALUES ( 40,"Second Record");

	// ALTER TABLE "DoubleKey4" ADD COLUMN "NewColumn2" bigint;
	// ALTER TABLE "DoubleKey4" ADD COLUMN "NewColumn3" integer;
	// ALTER TABLE "DoubleKey4" ADD COLUMN "NewColumn4" bool;
	// ALTER TABLE "DoubleKey4" ADD COLUMN "NewColumn5" integer;

	// CREATE UNIQUE INDEX idx_double_key4_new_column3 ON "DoubleKey4"("NewColumn2");

	// CREATE INDEX idx_double_key4_new_column4_new_column5 ON "DoubleKey4"("NewColumn4, NewColumn5");

	// SELECT * FROM sqlite_sequence WHERE "name" = "DoubleKey4";

	// UPDATE "sqlite_sequence" SET "seq" = 50000000 WHERE "name" = "DoubleKey4";

	// SELECT * FROM sqlite_sequence WHERE "name" = "DoubleKey4";

	// SELECT * FROM "DoubleKey4";
	//========================================================================================================

	for idx, fd := range fldef {

		var col ColComponents

		col.fName = fd.FName
		col.fType = ""
		col.fPrimaryKey = ""
		col.fDefault = ""
		col.fNullable = ""

		// int64   bigint
		// uint64  bigint
		// int32   integer
		// int16   integer
		// int8    integer
		// int     integer
		// uint32  integer
		// uint16  integer
		// uint8   integer
		// uint	   integer
		// byte    integer
		// rune    integer
		// float32  real
		// float64  real
		// string  varchar(255) (or 256?)
		// date/time  datetime - need now() function equivalent - also function to create a date-time from a string

		switch fd.GoType {
		case "int64", "uint64":
			col.fType = "bigint"

		case "uint", "uint8", "uint16", "uint32", "int",
			"int8", "int16", "int32", "rune", "byte":
			col.fType = "integer"

		case "float32", "float64":
			col.fType = "real"

		case "bool":
			col.fType = "bool"

		case "string":
			col.fType = "varchar(255)"

		case "time.Time", "*time.Time":
			col.fType = "datetime"

		default:
			err := fmt.Errorf("go type %s is not presently supported", fldef[idx].FType)
			panic(err)
		}
		fldef[idx].FType = col.fType

		// read rgen tag pairs and apply
		seqName := ""
		if !strings.Contains(fd.GoType, "*time.Time") {

			for _, p := range fd.RgenPairs {

				switch p.Name {
				case "primary_key":

					col.fPrimaryKey = "PRIMARY KEY"
					pKeys = fmt.Sprintf("%s %s%s%s,", pKeys, qt, fd.FName, qt)
					// pKeys = pKeys + fd.FName + ","

					// int-type primary keys will autoincrement based on ROWID,
					// but the speed increase comes with the cost of losing control
					// of the starting point of the range at time of table creation.
					// addtionally, relying on the default keygen opens the door
					// to the reuse of deleted keys - which is largely problematic
					// (for me at least).  for this reason, AUTOINCREMENT is used.
					// if AUTOINCREMENT is requested on an int64/uint64, downcast
					// the db-field-type to integer.
					if p.Value == "inc" && strings.Contains(fd.GoType, "int") {
						if strings.Contains(fd.GoType, "64") {
							fldef[idx].FType = "integer"
							col.fType = "integer"
						}
						col.fAutoInc = true
					}

				case "start":
					start, err := strconv.Atoi(p.Value)
					if err != nil {
						panic(err)
					}
					if seqName == "" && start > 0 {
						seqName = tn
						sequences = append(sequences, RgenPair{Name: seqName, Value: p.Value})
					}

				case "default":
					if fd.GoType == "string" {
						col.fDefault = fmt.Sprintf("DEFAULT '%s'", p.Value)
					} else {
						col.fDefault = fmt.Sprintf("DEFAULT %s", p.Value)
					}

					if fd.GoType == "time.Time" {
						switch p.Value {
						case "now()":
							p.Value = "(datetime('now'))"
						case "eot":
							p.Value = "('9999-12-31 23:59:59')"
						default:

						}
						col.fDefault = fmt.Sprintf("DEFAULT %s", p.Value)
					}

				case "nullable":
					if p.Value == "false" {
						col.fNullable = "NOT NULL"
					}

				case "index":
					switch p.Value {
					case "non-unique":
						indexes = slf.processIndexTag(indexes, tn, fd.FName, "idx_"+fd.FName, false, true)

					case "unique":
						indexes = slf.processIndexTag(indexes, tn, fd.FName, "idx_"+fd.FName, true, true)

					default:
						indexes = slf.processIndexTag(indexes, tn, fd.FName, p.Value, false, false)
					}

				default:

				}
			}
		} else { // *time.Time only supports default directive
			for _, p := range fd.RgenPairs {
				if p.Name == "default" {
					switch p.Value {
					case "now()":
						p.Value = "(datetime('now'))"
					case "eot":
						p.Value = "('9999-12-31 23:59:59')"
					default:

					}
					col.fDefault = fmt.Sprintf("DEFAULT %s", p.Value)
				}
			}
		}
		fldef[idx].FType = col.fType

		// add the current column to the schema
		tableSchema = tableSchema + fmt.Sprintf("%s%s%s %s", qt, col.fName, qt, col.fType)
		if col.fPrimaryKey != "" {
			tableSchema = tableSchema + fmt.Sprintf(" %s", col.fPrimaryKey)
		}
		if col.fAutoInc == true {
			tableSchema = tableSchema + " AUTOINCREMENT"
		}
		if col.fNullable != "" {
			tableSchema = tableSchema + " " + col.fNullable
		}
		if col.fDefault != "" {
			tableSchema = tableSchema + " " + col.fDefault
		}
		tableSchema = tableSchema + ", "
	}

	if tableSchema != "" && pKeys == "" {
		tableSchema = strings.TrimSuffix(tableSchema, ",")
		tableSchema = tableSchema + ")"
	}
	if tableSchema != "" && pKeys != "" {
		pKeys = strings.TrimSuffix(pKeys, ",")
		tableSchema = tableSchema + fmt.Sprintf("UNIQUE(%s) );", pKeys)
	}

	// pass out the CREATE TABLE schema, and component info
	return TblComponents{
		tblSchema: tableSchema,
		flDef:     fldef,
		seq:       sequences,
		ind:       indexes,
		pk:        pKeys,
		err:       err,
	}
}

// ExistsTable checks that the specified table exists in the SQLite database file.
func (slf *SQLiteFlavor) ExistsTable(tn string) bool {

	n := 0
	reqQuery := fmt.Sprintf("SELECT COUNT(*) FROM sqlite_master WHERE type=\"table\" AND name=\"%s\";", tn)
	err := slf.db.QueryRow(reqQuery).Scan(&n)
	if err != nil {
		return false
	}
	if n == 0 {
		return false
	}
	return true
}

// DropTables drops tables on the SQLite db if they exist, based on
// the provided list of go struct definitions.
func (slf *SQLiteFlavor) DropTables(i ...interface{}) error {

	dropSchema := ""
	for t := range i {

		// determine the table name
		tn := reflect.TypeOf(i[t]).String() // models.ProfileHeader{} for example
		if strings.Contains(tn, ".") {
			el := strings.Split(tn, ".")
			tn = strings.ToLower(el[len(el)-1])
		} else {
			tn = strings.ToLower(tn)
		}
		if tn == "" {
			return fmt.Errorf("unable to determine table name in slf.DropTables")
		}

		// if the table is found to exist, add a DROP statement
		// to the dropSchema string and move on to the next
		// table in the list.
		if slf.ExistsTable(tn) {
			if slf.log {
				fmt.Printf("table %s exists - adding to drop schema...\n", tn)
			}
			// submit 1 at a time for mysql
			dropSchema = dropSchema + fmt.Sprintf("DROP TABLE IF EXISTS %s; ", tn)
			slf.ProcessSchema(dropSchema)
			dropSchema = ""
		}
	}
	return nil
}

// DropIndex drops the specfied index on the connected SQLite database.  SQLite does
// not require the table name to drop an index, but it is provided in order to
// comply with the PublicDB interface definition.
func (slf *SQLiteFlavor) DropIndex(tn string, in string) error {

	indexSchema := fmt.Sprintf("DROP INDEX IF EXISTS %s;", in)
	slf.ProcessSchema(indexSchema)
	return nil
}

// ExistsIndex checks the connected SQLite database for the presence
// of the specified index.  This method is typically not required
// for SQLite, as the 'IF EXISTS' syntax is widely supported.
func (slf *SQLiteFlavor) ExistsIndex(tn string, in string) bool {

	n := 0
	indQuery := fmt.Sprintf("SELECT COUNT(*) FROM sqlite_master WHERE \"type\" = \"index\" AND \"name\" = \"%s\";", in)
	slf.db.QueryRow(indQuery).Scan(&n)
	if n > 0 {
		return true
	}
	return false
}

// ExistsColumn checks the current SQLite database file and
// returns true if the named table-column is found to exist.
// this checks the column name only, not the column data-type
// or properties.
func (slf *SQLiteFlavor) ExistsColumn(tn string, cn string) bool {

	if slf.ExistsTable(tn) {

		// colQuery := fmt.Sprintf("PRAGMA table_info(\"%s\")", tn)  // does not work - annoying
		// without using the built-in PRAGMA, we have to rely on the table creation SQL
		// that is stored in the sqlite_master table - not very exact.
		sqlString := ""
		colQuery := fmt.Sprintf("SELECT \"sql\" FROM \"sqlite_master\" WHERE \"type\" = \"table\" AND \"name\" = \"%s\"", tn)
		slf.db.QueryRow(colQuery).Scan(&sqlString)
		if sqlString == "" {
			return false
		}

		if strings.Contains(sqlString, cn) {
			return true
		}
	}
	return false
}

// DestructiveResetTables drops tables on the SQLite db file if they exist,
// as well as any related objects such as sequences.  this is
// useful if you wish to regenerated your table and the
// number-range used by an auto-incementing primary key.
func (slf *SQLiteFlavor) DestructiveResetTables(i ...interface{}) error {

	err := slf.DropTables(i...)
	if err != nil {
		return err
	}
	err = slf.CreateTables(i...)
	if err != nil {
		return err
	}
	return nil
}

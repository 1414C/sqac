package sqac

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/1414C/sqac/common"
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

// GetDBName returns the name of the currently connected db
func (slf *SQLiteFlavor) GetDBName() (dbName string) {

	dbNum := ""
	dbMain := ""
	qs := "PRAGMA database_list;"
	slf.QsLog(qs)

	row := slf.db.QueryRow(qs)
	if row != nil {
		err := row.Scan(&dbNum, &dbMain, &dbName)
		if err != nil {
			panic(err)
		}
	}
	return dbName
}

// CreateTables creates tables on the sqlite3 database referenced
// by slf.DB.
func (slf *SQLiteFlavor) CreateTables(i ...interface{}) error {

	for t, ent := range i {

		ftr := reflect.TypeOf(ent)
		if slf.log {
			log.Println("CreateTable() entity type:", ftr)
		}

		// determine the table name
		tn := common.GetTableName(i[t])
		if tn == "" {
			return fmt.Errorf("unable to determine table name in slf.CreateTables")
		}

		// if the table is found to exist, skip the creation
		// and move on to the next table in the list.
		if slf.ExistsTable(tn) {
			if slf.log {
				log.Printf("CreateTable - table %s exists - skipping...\n", tn)
			}
			continue
		}

		// get all the table parts and build the create schema
		tc := slf.buildTablSchema(tn, i[t], false)
		slf.QsLog(tc.tblSchema)

		// execute the create schema against the db
		slf.db.MustExec(tc.tblSchema)
		for _, sq := range tc.seq {
			start, _ := strconv.Atoi(sq.Value)
			slf.AlterSequenceStart(sq.Name, start-1)
		}
		for k, in := range tc.ind {
			slf.CreateIndex(k, in)
		}
	}
	return nil
}

// AlterTables alters tables on the SQLite database referenced
// by slf.DB.
func (slf *SQLiteFlavor) AlterTables(i ...interface{}) error {

	for t, ent := range i {

		// determine the table name
		tn := common.GetTableName(i[t])
		if tn == "" {
			return fmt.Errorf("unable to determine table name in slf.AlterTables")
		}

		// if the table does not exist, call CreateTables
		// if the table does exist, examine it and perform
		// alterations if necessary
		if !slf.ExistsTable(tn) {
			slf.CreateTables(ent)
			continue
		}

		// build the altered table schema and get its components
		tc := slf.buildTablSchema(tn, i[t], true)

		// go through the latest version of the model and check each
		// field against its definition in the database.
		var cols []string

		for _, fd := range tc.flDef {
			// new columns first
			if !slf.ExistsColumn(tn, fd.FName) && fd.NoDB == false {

				colSchema := "ALTER TABLE " + tn + " ADD COLUMN " + fd.FName + " " + fd.FType
				for _, p := range fd.SqacPairs {
					switch p.Name {
					case "primary_key":
						// abort - adding primary key
						panic(fmt.Errorf("aborting - cannot add a primary-key (table-field %s-%s) through migration", tn, fd.FName))

					case "default":
						if fd.UnderGoType == "string" {
							colSchema = colSchema + " DEFAULT '" + p.Value + "'"
						} else {
							colSchema = colSchema + " DEFAULT " + p.Value
						}

					case "nullable":
						if p.Value == "false" {
							colSchema = colSchema + " DEFAULT " + p.Value
						}

					default:

					}
				}
				cols = append(cols, colSchema+";")
			}
		}

		// ALTER TABLE ADD COLUMN ...
		if len(cols) > 0 {
			for _, c := range cols {
				if slf.IsLog() {
					log.Println(c)
				}
				slf.ProcessSchema(c)
			}
		}

		// add indexes if required
		for k, v := range tc.ind {
			if !slf.ExistsIndex(v.TableName, k) {
				slf.CreateIndex(k, v)
			}
		}

		// add foreign-keys if required - this is quite intensive, as the table
		// will go through copy, drop, recreate, reload cycle for each foreign-
		// key.  it would be possible to react only to the first 'new' foreign-
		// key in the list, as the entire model will be processed during an
		// ADD CONSTRAINT ... FOREIGN KEY ... ...(..) operation.
		for _, v := range tc.fkey {
			fkn, err := common.GetFKeyName(ent, v.FromTable, v.RefTable, v.FromField, v.RefField)
			if err != nil {
				return err
			}
			fkExists, _ := slf.ExistsForeignKeyByName(ent, fkn)
			if !fkExists {
				err = slf.CreateForeignKey(ent, v.FromTable, v.RefTable, v.FromField, v.RefField)
				if err != nil {
					log.Println(err)
					return err
				}
			}
		}
	}
	return nil
}

// buildTableSchema builds a CREATE TABLE schema for the SQLite DB
// and returns it to the caller, along with the components determined from
// the db and sqac struct-tags.  this method is used in CreateTables
// and AlterTables methods.
func (slf *SQLiteFlavor) buildTablSchema(tn string, ent interface{}, isAlter bool) TblComponents {

	qt := slf.GetDBQuote()
	pKeys := ""
	var sequences []common.SqacPair
	indexes := make(map[string]IndexInfo)
	fKeys := make([]FKeyInfo, 0)
	tableSchema := "CREATE TABLE IF NOT EXISTS " + qt + tn + qt + " ("

	// get a list of the field names, go-types and db attributes.
	// TagReader is a common function across db-flavors. For
	// this reason, the db-specific-data-type for each field
	// is determined locally.
	fldef, err := common.TagReader(ent, nil)
	if err != nil {
		panic(err)
	}

	// set the SQLite field-types and build the table schema,
	// as well as any other schemas that are needed to support
	// the table definition. In all cases any foreign-key or
	// index requirements must be deferred until all other
	// artifacts have been created successfully.
	// SQLite has basic types more along the lines of Postgres.

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

		// date/time  datetime - need now() function equivalent - also function to create a date-time from a string

		// if the field has been marked as NoDB, continue with the next field
		if fd.NoDB == true {
			continue
		}

		switch fd.UnderGoType {
		case "int64", "uint64":
			col.fType = "bigint"

		case "uint", "uint8", "uint16", "uint32", "int",
			"int8", "int16", "int32", "rune", "byte":
			col.fType = "integer"

		case "float32", "float64":
			col.fType = "real"

		case "bool":
			col.fType = "boolean"

		case "string":
			col.fType = "varchar(255)"

		case "time.Time":
			col.fType = "datetime"

		default:
			err := fmt.Errorf("go type %s is not presently supported", fldef[idx].FType)
			panic(err)
		}
		fldef[idx].FType = col.fType

		// read sqac tag pairs and apply
		seqName := ""

		if !strings.Contains(fd.GoType, "*time.Time") {

			for _, p := range fd.SqacPairs {

				switch p.Name {
				case "primary_key":

					pKeys = pKeys + " " + qt + fd.FName + qt + ","

					// int-type primary keys will autoincrement based on ROWID,
					// but the speed increase comes with the cost of losing control
					// of the starting point of the range at time of table creation.
					// addtionally, relying on the default keygen opens the door
					// to the reuse of deleted keys - which is largely problematic
					// (for me at least).  for this reason, AUTOINCREMENT is used.
					// if AUTOINCREMENT is requested on an int64/uint64, downcast
					// the db-field-type to integer.
					if p.Value == "inc" && strings.Contains(fd.UnderGoType, "int") {
						if col.uType != "" {
							log.Printf("WARNING: %s auto-incrementing primary-key field %s has user-specified db_type: %s  user-type is ignored. \n", common.GetTableName(ent), col.fName, col.uType)
							col.uType = ""
						}
						col.fPrimaryKey = "PRIMARY KEY"
						if strings.Contains(fd.UnderGoType, "64") {
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
						sequences = append(sequences, common.SqacPair{Name: seqName, Value: p.Value})
					}

				case "default":
					if fd.UnderGoType == "string" {
						col.fDefault = "DEFAULT '" + p.Value + "'"
					} else {
						col.fDefault = "DEFAULT " + p.Value
					}

					if fd.UnderGoType == "bool" {
						switch p.Value {
						case "TRUE", "true":
							p.Value = "1"
						case "FALSE", "false":
							p.Value = "0"
						default:

						}
						col.fDefault = "DEFAULT " + p.Value
					}

					if fd.UnderGoType == "time.Time" {
						switch p.Value {
						case "now()":
							p.Value = "(datetime('now'))"
						case "eot()":
							p.Value = "('9999-12-31 23:59:59')"
						default:

						}
						col.fDefault = "DEFAULT " + p.Value
					}

				case "nullable":
					if p.Value == "false" {
						col.fNullable = "NOT NULL"
					}

				case "constraint":
					if p.Value == "unique" {
						col.fUniqueConstraint = "UNIQUE"
					}

				case "index":
					switch p.Value {
					case "non-unique":
						indexes = slf.processIndexTag(indexes, tn, fd.FName, "idx_", false, true)

					case "unique":
						indexes = slf.processIndexTag(indexes, tn, fd.FName, "idx_", true, true)

					default:
						indexes = slf.processIndexTag(indexes, tn, fd.FName, p.Value, false, false)
					}

				case "fkey":
					fKeys = slf.processFKeyTag(fKeys, tn, fd.FName, p.Value)

				default:

				}
			}
		} else { // *time.Time only supports default directive
			for _, p := range fd.SqacPairs {
				if p.Name == "default" {
					switch p.Value {
					case "now()":
						p.Value = "(datetime('now'))"
					case "eot()":
						p.Value = "('9999-12-31 23:59:59')"
					default:

					}
					col.fDefault = "DEFAULT " + p.Value
				}

				if p.Name == "primary_key" {
					pKeys = pKeys + " " + qt + fd.FName + qt + ","
				}

				if p.Name == "fkey" {
					fKeys = slf.processFKeyTag(fKeys, tn, fd.FName, p.Value)
				}

			}
		}
		fldef[idx].FType = col.fType

		// add the current column to the schema
		tableSchema = tableSchema + qt + col.fName + qt + " " + col.fType
		if col.fPrimaryKey != "" {
			tableSchema = tableSchema + " " + col.fPrimaryKey
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
		if col.fUniqueConstraint != "" {
			tableSchema = tableSchema + " " + col.fUniqueConstraint
		}
		tableSchema = tableSchema + ", "
	}

	if tableSchema != "" && pKeys == "" {
		tableSchema = strings.TrimSpace(tableSchema)
		tableSchema = strings.TrimSuffix(tableSchema, ",")
	}
	if tableSchema != "" && pKeys != "" {
		pKeys = strings.TrimSuffix(pKeys, ",")
		tableSchema = tableSchema + "UNIQUE(" + pKeys + ")"
	}

	// SQLite needs the foreign-key constraints added in the CREATE TABLE schema when
	// building the schema for a CreateTables call.  In an AlterTables call, it is not
	// possible to add a new foreign-key to a SQLite table, so the foreign-key constraints
	// are left out of the schema altogether.  Existing foreign-keys should stay in-place
	// and a list of foreign-keys to add is exported for use with the copy, drop, recreate,
	// reload process used to add/drop foreign-keys.
	if !isAlter {
		if tableSchema != "" && len(fKeys) > 0 {
			tableSchema = strings.TrimSpace(tableSchema)
			lv := tableSchema[len(tableSchema)-1:]
			if lv != "," {
				tableSchema = tableSchema + ","
			}

			for _, v := range fKeys {
				fkn, err := common.GetFKeyName(nil, v.FromTable, v.RefTable, v.FromField, v.RefField)
				if err != nil {
					log.Printf("WARNING: unable to determine foreign-key-name based on %v.  SKIPPING.", v)
					continue
				}
				tableSchema = tableSchema + " CONSTRAINT " + fkn + " FOREIGN KEY (" + v.FromField + ") REFERENCES " + v.RefTable + "(" + v.RefField + "),"
			}
		}
	}

	if tableSchema != "" {
		tableSchema = strings.TrimSpace(tableSchema)
		tableSchema = strings.TrimSuffix(tableSchema, ",")
		tableSchema = tableSchema + ");"
	}

	// fill the return structure passing out the CREATE TABLE schema, and component info
	rc := TblComponents{
		tblSchema: tableSchema,
		flDef:     fldef,
		seq:       sequences,
		ind:       indexes,
		pk:        pKeys,
		err:       err,
	}

	if isAlter && len(fKeys) > 0 {
		rc.fkey = fKeys
	}

	if slf.log {
		rc.Log()
	}
	return rc
}

// DropTables drops tables on the SQLite db if they exist, based on
// the provided list of go struct definitions.
func (slf *SQLiteFlavor) DropTables(i ...interface{}) error {

	dropSchema := ""
	for t := range i {

		// determine the table name
		tn := common.GetTableName(i[t])
		if tn == "" {
			return fmt.Errorf("unable to determine table name in slf.DropTables")
		}

		// if the table is found to exist, add a DROP statement
		// to the dropSchema string and move on to the next
		// table in the list.
		if slf.ExistsTable(tn) {
			if slf.log {
				log.Printf("table %s exists - adding to drop schema...\n", tn)
			}
			dropSchema = dropSchema + "DROP TABLE IF EXISTS " + tn + ";"
			slf.ProcessSchema(dropSchema)
			dropSchema = ""
		}
	}
	return nil
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

// ExistsTable checks that the specified table exists in the SQLite database file.
func (slf *SQLiteFlavor) ExistsTable(tn string) bool {

	n := 0
	reqQuery := "SELECT COUNT(*) FROM sqlite_master WHERE type=\"table\" AND name=\"" + tn + "\";"
	slf.QsLog(reqQuery)
	err := slf.db.QueryRow(reqQuery).Scan(&n)
	if err != nil {
		return false
	}
	if n == 0 {
		return false
	}
	return true
}

// DropIndex drops the specfied index on the connected SQLite database.  SQLite does
// not require the table name to drop an index, but it is provided in order to
// comply with the PublicDB interface definition.
func (slf *SQLiteFlavor) DropIndex(tn string, in string) error {

	indexSchema := "DROP INDEX IF EXISTS " + in + ";"
	slf.ProcessSchema(indexSchema)
	return nil
}

// ExistsIndex checks the connected SQLite database for the presence
// of the specified index.  This method is typically not required
// for SQLite, as the 'IF EXISTS' syntax is widely supported.
func (slf *SQLiteFlavor) ExistsIndex(tn string, in string) bool {

	n := 0
	indQuery := "SELECT COUNT(*) FROM sqlite_master WHERE \"type\" = \"index\" AND \"name\" = \"" + in + "\";"
	slf.QsLog(indQuery)

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
		// that is stored in the sqlite_master table.
		sqlString := ""
		colQuery := "SELECT \"sql\" FROM sqlite_master WHERE \"type\" = \"table\" AND \"name\" = \"" + tn + "\""
		slf.QsLog(colQuery)

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

// AlterSequenceStart may be used to make changes to the start value of the
// auto-increment field on the currently connected SQLite database file.
// This method is intended to be called at the time of table-creation, as
// updating the current value of the SQLite auto-increment may cause
// unanticipated difficulties if the target table already contains
// records.
func (slf *SQLiteFlavor) AlterSequenceStart(name string, start int) error {

	asQuery := "UPDATE sqlite_sequence SET seq = " + strconv.Itoa(start) + " WHERE name = '" + name + "';"
	slf.QsLog(asQuery)

	result, err := slf.Exec(asQuery)
	if err == nil {
		ra, err := result.RowsAffected()
		if err == nil && ra > 0 {
			log.Println("ra==", ra)
			return nil
		}
	}

	asQuery = "INSERT INTO sqlite_sequence (name,seq) VALUES ('" + name + "', " + strconv.Itoa(start) + ");"
	slf.QsLog(asQuery)

	result, err = slf.Exec(asQuery)
	if err != nil {
		return err
	}
	ra, err := result.RowsAffected()
	if err != nil || ra < 1 {
		return err
	}
	return nil
}

// GetNextSequenceValue is used primarily for testing.  It returns
// the current value of the SQLite auto-increment field for the named
// table.
func (slf *SQLiteFlavor) GetNextSequenceValue(name string) (int, error) {

	seq := 0
	if slf.ExistsTable(name) {

		// colQuery := fmt.Sprintf("PRAGMA table_info(\"%s\")", tn)  // does not work - annoying
		// without using the built-in PRAGMA, we have to rely on the table creation SQL
		// that is stored in the sqlite_master table - not very exact.
		seqQuery := "SELECT \"seq\" FROM sqlite_sequence WHERE \"name\" = '" + name + "'"
		slf.QsLog(seqQuery)

		err := slf.db.QueryRow(seqQuery).Scan(&seq)
		if err != nil {
			return 0, err
		}
		return seq + 1, nil
	}
	return seq, nil
}

// CreateForeignKey creates a foreign key on an existing column in the database table
// specified by the i / ft parameter.  SQLite does not support the addition of a
// foreign-key via ALTER TABLE, so the existing table has to be copied to a backup,
// and a new table created (hence parameter i) with the foreign-key constraint in
// the CREATE TABLE ... command.  Foreign-key constraints are temporarily disabled
// on the db for the duration of the transaction processing.
// THIS SHOULD NOT BE CALLED DIRECTLY.  IT IS FAR SAFER IN THE SQLITE CASE TO UPDATE
// THE SQAC-TAGS ON THE TABLE'S MODEL.
func (slf *SQLiteFlavor) CreateForeignKey(i interface{}, ft, rt, ff, rf string) error {

	bakTn := ""
	q := ""

	// sql transaction command buffer
	cmds := make([]string, 0)

	// confirm the table name
	tn := common.GetTableName(i)
	if tn == "" || tn != ft {
		return fmt.Errorf("unable to confirm table name in slf.CreateForeignKey")
	}

	// if the table is found to exist, copy it to a temp backup table - Sprintf for legibility
	if slf.ExistsTable(tn) {
		bakTn = fmt.Sprintf("_%s_bak", ft)
		q = "DROP TABLE IF EXISTS " + bakTn + ";"
		cmds = append(cmds, q)
		q = fmt.Sprintf("ALTER TABLE %s RENAME TO _%s_bak;", ft, ft)
		cmds = append(cmds, q)
		q = "DROP TABLE IF EXISTS " + tn + ";"
		cmds = append(cmds, q)
	}

	// determine the new fk constraint-name
	fkn, err := common.GetFKeyName(i, ft, rt, ff, rf)
	if err != nil {
		return err
	}

	// build the new foreign-key constraint clause
	fkc := fmt.Sprintf(" CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s)", fkn, ff, rt, rf)

	// build the new table schema with foreign-key constraint
	tc := slf.buildTablSchema(tn, i, false)
	q = strings.TrimSuffix(tc.tblSchema, ");")
	q = strings.TrimSpace(q)
	lv := q[len(q)-1:]
	if lv != "," {
		q = q + ","
	}
	q = fmt.Sprintf("%s%s%s", q, fkc, ");")
	cmds = append(cmds, q)

	// copy the data back - note that this can happen even if there is a
	// foreign-key violation due to the prior PRAGMA.  :)
	if bakTn != "" {
		q = "INSERT INTO " + ft + " SELECT * FROM " + bakTn + ";"
		cmds = append(cmds, q)

		// drop the backup table directly
		q = "DROP TABLE IF EXISTS " + bakTn + ";"
		cmds = append(cmds, q)
	}

	// disable foreign-key checks
	_, err = slf.Exec("PRAGMA foreign_keys=off;")
	if err != nil {
		return err
	}

	// submit the transaction buffer
	err = slf.ProcessTransaction(cmds)
	if err != nil {
		// attempt to reactivate foreign-key constraints
		_, fkErr := slf.Exec("PRAGMA foreign_keys=on;")
		if fkErr != nil {
			log.Println("WARNING: FOREIGN KEY CONSTRAINTS ARE PRESENTLY DEACATIVATED!")
		}
		return err
	}

	// reactivate foreign-key constraints
	_, err = slf.Exec("PRAGMA foreign_keys=on;")
	if err != nil {
		log.Println("WARNING: FOREIGN KEY CONSTRAINTS MAY PRESENTLY BE DEACATIVATED!")
		return err
	}
	return nil
}

// DropForeignKey drops a foreign-key on an existing column.  Since SQLite does not
// support the addition or deletion of foreign-key relationships on existing tables,
// the existing table is copied to a backup table, dropped and then recreated using
// the sqac model information contained in (i).  It follows that in order for
// a foreign-key to be dropped, it must be removed from the sqac tag in the model
// definition.
func (slf *SQLiteFlavor) DropForeignKey(i interface{}, ft, fkn string) error {

	bakTn := ""
	q := ""

	// sql transaction command buffer
	cmds := make([]string, 0)

	// confirm the table name
	tn := common.GetTableName(i)
	if tn == "" || tn != ft {
		return fmt.Errorf("unable to confirm table name in slf.DropForeignKey")
	}

	// if the table is found to exist, copy it to a temp backup table
	if slf.ExistsTable(tn) {
		bakTn = fmt.Sprintf("_%s_bak", ft)
		q = "DROP TABLE IF EXISTS " + bakTn + ";"
		cmds = append(cmds, q)
		q = fmt.Sprintf("ALTER TABLE %s RENAME TO _%s_bak;", ft, ft)
		cmds = append(cmds, q)
		q = "DROP TABLE IF EXISTS " + tn + ";"
		cmds = append(cmds, q)
	}

	// build the new table schema without the foreign-key constraint (must be omitted from model)
	tc := slf.buildTablSchema(tn, i, false)
	cmds = append(cmds, tc.tblSchema)

	// copy the data back
	if bakTn != "" {
		q = "INSERT INTO " + ft + " SELECT * FROM " + bakTn + ";"
		cmds = append(cmds, q)

		// drop the backup table directly
		q = "DROP TABLE IF EXISTS " + bakTn + ";"
		cmds = append(cmds, q)
	}

	// disable foreign-key checks to start the transaction processing
	qs := "PRAGMA foreign_keys=off;"
	slf.QsLog(qs)
	_, err := slf.Exec(qs)
	if err != nil {
		return err
	}

	// submit the transaction buffer
	err = slf.ProcessTransaction(cmds)
	if err != nil {
		// attempt to reactivate foreign-key constraints
		_, fkErr := slf.Exec("PRAGMA foreign_keys=on;")
		if fkErr != nil {
			log.Println("WARNING: FOREIGN KEY CONSTRAINTS ARE PRESENTLY DEACATIVATED!")
		}
		return err
	}

	// reactivate foreign-key constraints
	qs = "PRAGMA foreign_keys=on;"
	slf.QsLog(qs)
	_, err = slf.Exec(qs)
	if err != nil {
		log.Println("WARNING: FOREIGN KEY CONSTRAINTS MAY PRESENTLY BE DEACATIVATED!")
		return err
	}
	return nil
}

// ExistsForeignKeyByName checks to see if the named foreign-key exists on the
// table corresponding to provided sqac model (i).
func (slf *SQLiteFlavor) ExistsForeignKeyByName(i interface{}, fkn string) (bool, error) {

	var count uint64
	tn := common.GetTableName(i)

	fkQuery := fmt.Sprintf("SELECT COUNT(*) FROM sqlite_master WHERE tbl_name='%s' AND sql like'%%%s%%';", tn, fkn)
	slf.QsLog(fkQuery)

	err := slf.Get(&count, fkQuery)
	if err != nil {
		return false, nil
	}

	if count > 0 {
		return true, nil
	}
	return false, nil
}

// ExistsForeignKeyByFields checks to see if a foreign-key exists between the named
// tables and fields.
func (slf *SQLiteFlavor) ExistsForeignKeyByFields(i interface{}, ft, rt, ff, rf string) (bool, error) {

	fkn, err := common.GetFKeyName(i, ft, rt, ff, rf)
	if err != nil {
		return false, err
	}
	return slf.ExistsForeignKeyByName(i, fkn)
}

//================================================================
// CRUD ops
//================================================================

// Create the entity (single-row) on the database
func (slf *SQLiteFlavor) Create(ent interface{}) error {

	var info CrudInfo
	info.ent = ent
	info.log = false
	info.mode = "C"

	err := slf.BuildComponents(&info)
	if err != nil {
		return err
	}

	// build the sqlite insert query
	insFlds := ""
	insVals := ""
	for k, v := range info.fldMap {
		if v == "DEFAULT" {
			continue
		}
		insFlds = insFlds + " " + k + ", "
		insVals = fmt.Sprintf("%s %s, ", insVals, v)
	}
	insFlds = strings.TrimSuffix(insFlds, ", ")
	insVals = strings.TrimSuffix(insVals, ", ")

	// build the sqlite insert query
	insQuery := "INSERT OR FAIL INTO " + info.tn + " (" + insFlds + ") VALUES (" + insVals + ");"
	slf.QsLog(insQuery)

	// clear the source data - deals with non-persistet columns
	e := reflect.ValueOf(info.ent).Elem()
	e.Set(reflect.Zero(e.Type()))

	// attempt the insert and read the result back into info.resultMap
	result, err := slf.db.Exec(insQuery)
	if err != nil {
		return err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	selQuery := "SELECT * FROM " + info.tn + " WHERE " + info.incKeyName + " = " + strconv.FormatInt(lastID, 10) + " LIMIT 1;"
	slf.QsLog(selQuery)

	err = slf.db.QueryRowx(selQuery).StructScan(info.ent) //.MapScan(info.resultMap) // SliceScan
	if err != nil {
		return err
	}
	info.entValue = reflect.ValueOf(info.ent)
	return nil
}

// Update an existing entity (single-row) on the database
func (slf *SQLiteFlavor) Update(ent interface{}) error {

	var info CrudInfo
	info.ent = ent
	info.log = false
	info.mode = "U"

	err := slf.BuildComponents(&info)
	if err != nil {
		return err
	}

	keyList := ""
	for k, s := range info.keyMap {

		fType := reflect.TypeOf(s).String()
		if slf.IsLog() {
			log.Printf("key: %v, value: %v\n", k, s)
			log.Println("TYPE:", fType)
		}

		if fType == "string" { // also applies to time.Time at this point due to .Format()
			keyList = fmt.Sprintf("%s %s = '%v' AND", keyList, k, s)
		} else {
			keyList = fmt.Sprintf("%s %s = %v AND", keyList, k, s)
		}
		// fmt.Printf("TypeOfKey: %v, keyName: %s\n", reflect.TypeOf(s), k)
	}
	keyList = strings.TrimSuffix(keyList, " AND")

	colList := ""
	for k, v := range info.fldMap {
		colList = fmt.Sprintf("%s %s = %s, ", colList, k, v)
	}
	colList = strings.TrimSuffix(colList, ", ")

	updQuery := "UPDATE OR FAIL " + info.tn + " SET " + colList + " WHERE " + keyList + ";"
	slf.QsLog(updQuery)

	// clear the source data - deals with non-persistent columns
	e := reflect.ValueOf(info.ent).Elem()
	e.Set(reflect.Zero(e.Type()))

	// attempt the update and check for errors
	_, err = slf.db.Exec(updQuery)
	if err != nil {
		return err
	}

	// read the updated row
	selQuery := "SELECT * FROM " + info.tn + " WHERE " + keyList + " LIMIT 1;"
	slf.QsLog(selQuery)

	err = slf.db.QueryRowx(selQuery).StructScan(info.ent) //.MapScan(info.resultMap) // SliceScan
	if err != nil {
		return err
	}
	info.entValue = reflect.ValueOf(info.ent)
	return nil
}

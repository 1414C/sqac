package sqac

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// HDBFlavor is a SAP Hana-specific implementation, where
// the Hana DB is approached as a traditional SQL-92 compliant
// database.  As such, some of the nice HDB things are left out.
// Methods defined in the PublicDB interface of struct-type
// BaseFlavor are called by default for HDBFlavor. If
// the method as it exists in the BaseFlavor implementation
// is not compatible with the schema-syntax required by
// HDB, the method in question may be overridden.
// Overriding (redefining) a BaseFlavor method may be
// accomplished through the addition of a matching method
// signature and implementation on the HDBFlavor
// struct-type.
type HDBFlavor struct {
	BaseFlavor

	//================================================================
	// possible local HDB-specific overrides
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

// things to deal with:
// rgen:"primary_key:inc;start:55550000"
// rgen:"nullable:false"
// rgen:"default:0"
// rgen:"index:idx_material_num_serial_num
// rgen:"index:unique/non-unique"
// timestamp syntax and functions
// - pg now() equivalent
// - pg make_timestamptz(9999, 12, 31, 23, 59, 59.9) equivalent

//=======================================================================================
// SQL Commands
//=======================================================================================

// SELECT Column, Column, COUNT(*)
//	FROM  Table
//	WHERE Condition
//	GROUP BY Column, Column
// 	HAVING Group_Condition
//	ORDER BY Column ASC, Column DESC;

// SELECT a, 'b', "c", 1, '2', "3" FROM "4";
//
//	a -> existing column
//  'b' -> artificial result column
//	"c" -> existing column named c
//	1	-> artificial column with 1 as a numeric constant
//	'2' -> artificial result column with string 2 as value in each row
//	"3"	-> existing column named 3
//	"4"	-> existing table named 4

// SELECT Name, Overtime * 60 FROM Official;
// SELECT Name, ADD_YEARS(Birthday, 18) As "18th Birthday" FROM Owner;
// SELECT Name, DAYNAME(ADD_YEARS(Birthday, ROUND(ABS(-18.2)))) AS Weekday FROM Owner;

// YEAR(Date) 						-> Year
// ADD_YEARS(Date, n)				-> n years later
// DAYNAME(Date)					-> weekday
// CURRENT_DATE						-> current date
// ABS(Number)						-> absolute value
// ROUND(Number)					-> rounding
// SQRT(Number)						-> square root
// UPPER(String)					-> convert to upper case
// SUBSTR(String, Start, Length)	-> cut out of a string (substring)
// LENGTH(String)                   -> length of a string

// SELECT Official.Name FROM Official;

// INSERT INTO Table VALUES (Value, Value, Value);
// INSERT INTO Table(Column, Column) VALUES (Value, Value);

// UPDATE Table SET Column = Value, Column = Value, Column = Value WHERE Condition;

// DELETE FROM Table WHERE Condition;
//=======================================================================================

//=======================================================================================
// DDL
//=======================================================================================
//
// data-types
//
// TINYINT			-> 0 - 255
// SMALLINT			-> -32768 - 32767
// INTEGER			-> -2147483648 - 2147483647
// BIGINT			-> big ....
//
// -> DECIMAL(p,s)
// SMALLDECIMAL		-> -369 to 368
// DECIMAL			-> -6111 to 6176
// REAL				-> 32-bit
// DOUBLE			-> 64-bit
//
// VARCHAR(n)		-> ASCII string maxlen (n <= 5000)
// NVARCHAR(n)		-> Unicode string maxlen (n <= 5000)
// ALPHANUM			-> Alpanumeric (n <= 127)
// SHORTTEXT		-> Unicode string maxlen (n <= 5000) special text/string search features
//
// DATE				-> 'YYYY-MM-DD'
// TIME				-> 'HH:MM:SS'
// SECONDDATE		-> 'YYYY-MM-DD HH:MM:SS'
// TIMESTAMP		-> '2012-05-21 18:00:57.1234567'
//
// VARBINARY		-> binary data maxlen (n <= 5000)
// BLOB				-> blob (max 2Gb)
// CLOB				-> long ASCII character string (max 2Gb)
// NCLOB			-> long unicode character string (max 2Gb)
// TEXT				-> long unicode character string (max 2Gb)
//
//
// CREATE COLUMN
// ALTER TABLE
// RENAME TABLE
// DROP TABLE

// CreateTables creates tables on the mysql database referenced
// by hf.DB.
func (hf *HDBFlavor) CreateTables(i ...interface{}) error {

	for t, ent := range i {

		ftr := reflect.TypeOf(ent)
		if hf.log {
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
			return fmt.Errorf("unable to determine table name in myf.CreateTables")
		}

		// if the table is found to exist, skip the creation
		// and move on to the next table in the list.
		if hf.ExistsTable(tn) {
			if hf.log {
				fmt.Printf("CreateTable - table %s exists - skipping...\n", tn)
			}
			continue
		}

		// build the create table schema and return all of the table info
		tc := hf.buildTablSchema(tn, i[t])

		// create the table on the db
		hf.db.MustExec(tc.tblSchema)
		for _, sq := range tc.seq {
			start, _ := strconv.Atoi(sq.Value)
			hf.AlterSequenceStart(sq.Name, start)
		}
		for k, in := range tc.ind {
			hf.CreateIndex(k, in)
		}
	}
	return nil
}

// AlterTables alters tables on the HDB database referenced
// by hf.DB.
func (hf *HDBFlavor) AlterTables(i ...interface{}) error {

	for t, ent := range i {

		// ftr := reflect.TypeOf(ent)

		// determine the table name
		tn := reflect.TypeOf(i[t]).String() // models.ProfileHeader{} for example
		if strings.Contains(tn, ".") {
			el := strings.Split(tn, ".")
			tn = strings.ToLower(el[len(el)-1])
		} else {
			tn = strings.ToLower(tn)
		}
		if tn == "" {
			return fmt.Errorf("unable to determine table name in hf.AlterTables")
		}

		// if the table does not exist, call CreateTables
		// if the table does exist, examine it and perform
		// alterations if neccessary
		if !hf.ExistsTable(tn) {
			hf.CreateTables(ent)
			continue
		}

		// build the altered table schema and get its components
		tc := hf.buildTablSchema(tn, i[t])

		// go through the latest version of the model and check each
		// field against its definition in the database.
		qt := hf.GetDBQuote()
		alterSchema := fmt.Sprintf("ALTER TABLE %s%s%s ADD ", qt, tn, qt)
		var cols []string

		for _, fd := range tc.flDef {
			// new columns first
			if !hf.ExistsColumn(tn, fd.FName) && fd.NoDB == false {

				colSchema := fmt.Sprintf("%s%s%s %s", qt, fd.FName, qt, fd.FType)
				for _, p := range fd.RgenPairs {
					switch p.Name {
					case "primary_key":
						// abort - adding primary key
						panic(fmt.Errorf("aborting - cannot add a primary-key (table-field %s-%s) through migration", tn, fd.FName))

					case "default":
						switch fd.GoType {
						case "string":
							colSchema = fmt.Sprintf("%s DEFAULT '%s'", colSchema, p.Value)

						case "bool":
							switch p.Value {
							case "TRUE", "true":
								p.Value = "1"

							case "FALSE", "false":
								p.Value = "0"

							default:

							}

						default:
							colSchema = fmt.Sprintf("%s DEFAULT %s", colSchema, p.Value)
						}

					case "nullable":
						if p.Value == "false" {
							colSchema = fmt.Sprintf("%s NOT NULL", colSchema)
						}

					default:

					}
				}
				cols = append(cols, colSchema+",")
			}
		}

		// ALTER TABLE ADD COLUMNS...
		if len(cols) > 0 {
			for _, c := range cols {
				alterSchema = fmt.Sprintf("%s %s", alterSchema, c)
			}

			alterSchema = strings.TrimSuffix(alterSchema, ",")
			hf.ProcessSchema(alterSchema)
		}

		// add indexes if required
		for k, v := range tc.ind {
			if !hf.ExistsIndex(v.TableName, k) {
				hf.CreateIndex(k, v)
			}
		}
	}
	return nil
}

// buildTableSchema builds a CREATE TABLE schema for the HDB DB
// and returns it to the caller, along with the components determined from
// the db and rgen struct-tags.  this method is used in CreateTables
// and AlterTables methods.
func (hf *HDBFlavor) buildTablSchema(tn string, ent interface{}) TblComponents {

	qt := hf.GetDBQuote()
	pKeys := ""
	var sequences []RgenPair
	indexes := make(map[string]IndexInfo)
	tableSchema := fmt.Sprintf("CREATE TABLE %s%s%s (", qt, tn, qt)

	// get a list of the field names, go-types and db attributes.
	// TagReader is a common function across db-flavors. For
	// this reason, the db-specific-data-type for each field
	// is determined locally.
	fldef, err := TagReader(ent, nil)
	if err != nil {
		panic(err)
	}

	// set the HDB field-types and build the table schema,
	// as well as any other schemas that are needed to support
	// the table definition. In all cases any foreign-key or
	// index requirements must be deferred until all other
	// artifacts have been created successfully.
	for idx, fd := range fldef {

		var col ColComponents

		col.fName = fd.FName
		col.fType = ""
		col.fPrimaryKey = ""
		col.fDefault = ""
		col.fNullable = ""

		// https://stackoverflow.com/questions/168736/how-do-you-set-a-default-value-for-a-mysql-datetime-column

		// if the field has been marked as NoDB, continue with the next field
		if fd.NoDB == true {
			continue
		}

		switch fd.GoType {
		case "int64", "uint64":
			col.fType = "bigint"

		case "int32", "uint32", "int", "uint":
			col.fType = "int"

		case "int16", "uint16":
			col.fType = "smallint"

		case "int8", "uint8", "byte", "rune":
			col.fType = "tinyint"

		case "float32", "float64":
			col.fType = "numeric" // default precision is 18

		case "bool":
			col.fType = "bit"

		case "string":
			col.fType = "varchar(255)" //

		case "time.Time", "*time.Time":
			col.fType = "datetime2"

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

					if p.Value == "inc" {
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
							p.Value = "GETDATE()"
						case "eot":
							p.Value = "'9999-12-31 23:59:59.999'"
						default:

						}
						col.fDefault = fmt.Sprintf("DEFAULT %s", p.Value)
					}

					if fd.GoType == "bool" {
						switch p.Value {
						case "TRUE", "true":
							p.Value = "1"

						case "FALSE", "false":
							p.Value = "0"

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
						indexes = hf.processIndexTag(indexes, tn, fd.FName, "idx_", false, true)

					case "unique":
						indexes = hf.processIndexTag(indexes, tn, fd.FName, "idx_", true, true)

					default:
						indexes = hf.processIndexTag(indexes, tn, fd.FName, p.Value, false, false)
					}

				default:

				}
			}
		} else { // *time.Time only supports default directive
			for _, p := range fd.RgenPairs {
				if p.Name == "default" {
					switch p.Value {
					case "now()":
						p.Value = "GETDATE()"
					case "eot":
						p.Value = "'9999-12-31 23:59:59.999'"
					default:

					}
					col.fDefault = fmt.Sprintf("DEFAULT %s", p.Value)
				}
			}
		}
		fldef[idx].FType = col.fType

		// add the current column to the schema
		tableSchema = tableSchema + fmt.Sprintf("%s%s%s %s", qt, col.fName, qt, col.fType)
		if col.fAutoInc == true {
			tableSchema = tableSchema + " IDENTITY(1,1)"
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
		tableSchema = tableSchema + fmt.Sprintf("PRIMARY KEY (%s) )", pKeys)
	}
	tableSchema = tableSchema + ";"

	// fill the return structure passing out the CREATE TABLE schema, and component info
	rc := TblComponents{
		tblSchema: tableSchema,
		flDef:     fldef,
		seq:       sequences,
		ind:       indexes,
		pk:        pKeys,
		err:       err,
	}

	if hf.log {
		rc.Log()
	}
	return rc
}

// DropTables drops tables on the db if they exist, based on
// the provided list of go struct definitions.
func (hf *HDBFlavor) DropTables(i ...interface{}) error {

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
			return fmt.Errorf("unable to determine table name in hf.DropTables")
		}

		// if the table is found to exist, add a DROP statement
		// to the dropSchema string and move on to the next
		// table in the list.
		if hf.ExistsTable(tn) {
			if hf.log {
				fmt.Printf("table %s exists - adding to drop schema...\n", tn)
			}
			// submit 1 at a time for mysql
			dropSchema = dropSchema + fmt.Sprintf("DROP TABLE %s; ", tn)
			hf.ProcessSchema(dropSchema)
			dropSchema = ""
		}
	}
	return nil
}

// ExistsTable checks the currently connected database and
// returns true if the named table is found to exist.
func (hf *HDBFlavor) ExistsTable(tn string) bool {

	n := 0
	etQuery := fmt.Sprintf("SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = 'dbo' AND TABLE_NAME = '%s';", tn)
	hf.db.QueryRow(etQuery).Scan(&n)
	if n > 0 {
		return true
	}
	return false
}

// GetDBName returns the name of the currently connected db
func (hf *HDBFlavor) GetDBName() (dbName string) {

	row := hf.db.QueryRow("SELECT DB_NAME()")
	if row != nil {
		err := row.Scan(&dbName)
		if err != nil {
			panic(err)
		}
	}
	return dbName
}

// ExistsIndex checks the connected database for the presence
// of the specified index.
func (hf *HDBFlavor) ExistsIndex(tn string, in string) bool {

	n := 0
	// hf.db.QueryRow("SELECT count(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE table_schema = ? AND table_name = ? AND index_name = ?", bf.GetDBName(), tn, in).Scan(&n)
	hf.db.QueryRow("SELECT COUNT(*) FROM sys.indexes WHERE name=? AND object_id = OBJECT_ID(?);", in, tn).Scan(&n)
	if n > 0 {
		return true
	}
	return false
}

// DropIndex drops the specfied index on the connected database.
func (hf *HDBFlavor) DropIndex(tn string, in string) error {

	if hf.ExistsIndex(tn, in) {
		indexSchema := fmt.Sprintf("DROP INDEX %s ON %s;", in, tn)
		hf.ProcessSchema(indexSchema)
		return nil
	}
	return nil
}

// ExistsColumn checks the currently connected database and
// returns true if the named table-column is found to exist.
// this checks the column name only, not the column data-type
// or properties.
func (hf *HDBFlavor) ExistsColumn(tn string, cn string) bool {

	n := 0
	if hf.ExistsTable(tn) {
		hf.db.QueryRow("SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = ? AND column_name = ?;", tn, cn).Scan(&n)
		if n > 0 {
			return true
		}
	}
	return false
}

// DestructiveResetTables drops tables on the HDB db if they exist,
// as well as any related objects such as sequences.  this is
// useful if you wish to regenerated your table and the
// number-range used by an auto-incementing primary key.
func (hf *HDBFlavor) DestructiveResetTables(i ...interface{}) error {

	err := hf.DropTables(i...)
	if err != nil {
		return err
	}
	err = hf.CreateTables(i...)
	if err != nil {
		return err
	}
	return nil
}

// AlterSequenceStart may be used to make changes to the start value of the
// named identity-field on the currently connected HDB database.
func (hf *HDBFlavor) AlterSequenceStart(name string, start int) error {

	// reseed the primary key
	// DBCC CHECKIDENT ('dbo.depot', RESEED, 50000000);
	alterSequenceSchema := fmt.Sprintf("DBCC CHECKIDENT (%s, RESEED, %d)", name, start)
	hf.ProcessSchema(alterSequenceSchema)
	return nil
}

// GetNextSequenceValue is used primarily for testing.  It returns
// the current value of the HDB identity (auto-increment) field for
// the named table.
func (hf *HDBFlavor) GetNextSequenceValue(name string) (int, error) {

	seq := 0
	if hf.ExistsTable(name) {

		seqQuery := fmt.Sprintf("SELECT IDENT_CURRENT( '%s' );", name)
		err := hf.db.QueryRow(seqQuery).Scan(&seq)
		if err != nil {
			return 0, err
		}
		return seq, nil
	}
	return seq, nil
}

//================================================================
// CRUD ops
//================================================================

// Create the entity (single-row) on the database
func (hf *HDBFlavor) Create(ent interface{}) error {

	var info CrudInfo
	info.ent = ent
	info.log = false
	info.mode = "C"

	err := hf.BuildComponents(&info)
	if err != nil {
		return err
	}

	// build the hdb insert query
	insFlds := "("
	insVals := "("
	for k, v := range info.fldMap {
		if v == "DEFAULT" {
			continue
		}
		insFlds = fmt.Sprintf("%s %s, ", insFlds, k)
		insVals = fmt.Sprintf("%s %s, ", insVals, v)
	}
	insFlds = strings.TrimSuffix(insFlds, ", ") + ")"
	insVals = strings.TrimSuffix(insVals, ", ") + ")"

	// build the hdb insert query
	insQuery := fmt.Sprintf("INSERT INTO %s %s VALUES %s;", info.tn, insFlds, insVals)
	fmt.Println(insQuery)

	// attempt the insert and read the result back into info.resultMap
	result, err := hf.db.Exec(insQuery)
	if err != nil {
		return err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	selQuery := fmt.Sprintf("SELECT * FROM %s WHERE %s = %v;", info.tn, info.incKeyName, lastID)
	err = hf.db.QueryRowx(selQuery).MapScan(info.resultMap) // SliceScan
	if err != nil {
		return err
	}

	// fill the underlying structure of the interface ptr with the
	// fields returned from the database.
	err = hf.FormatReturn(&info)
	if err != nil {
		return err
	}
	return nil
}

// Update an existing entity (single-row) on the database
func (hf *HDBFlavor) Update(ent interface{}) error {

	var info CrudInfo
	info.ent = ent
	info.log = false
	info.mode = "U"

	err := hf.BuildComponents(&info)
	if err != nil {
		return err
	}

	keyList := ""
	for k, s := range info.keyMap {

		fType := reflect.TypeOf(s).String()
		if hf.IsLog() {
			fmt.Printf("key: %v, value: %v\n", k, s)
			fmt.Println("TYPE:", fType)
		}

		if fType == "string" {
			keyList = fmt.Sprintf("%s %s = '%v' AND", keyList, k, s)
		} else {
			keyList = fmt.Sprintf("%s %s = %v AND", keyList, k, s)
		}
	}
	keyList = strings.TrimSuffix(keyList, " AND")

	colList := ""
	for k, v := range info.fldMap {
		colList = fmt.Sprintf("%s %s = %s, ", colList, k, v)
	}
	colList = strings.TrimSuffix(colList, ", ")

	updQuery := fmt.Sprintf("UPDATE %s SET %s WHERE %s;", info.tn, colList, keyList)
	fmt.Println(updQuery)

	// attempt the update and check for errors
	_, err = hf.db.Exec(updQuery)
	if err != nil {
		return err
	}

	// read the updated row
	selQuery := fmt.Sprintf("SELECT * FROM %s WHERE %v;", info.tn, keyList)
	err = hf.db.QueryRowx(selQuery).MapScan(info.resultMap) // SliceScan
	if err != nil {
		return err
	}

	// fill the underlying structure of the interface ptr with the
	// fields returned from the database.
	err = hf.FormatReturn(&info)
	if err != nil {
		return err
	}
	return nil
}

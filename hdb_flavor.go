package sqac

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/1414C/sqac/common"
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
// sqac:"primary_key:inc;start:55550000"
// sqac:"nullable:false"
// sqac:"default:0"
// sqac:"index:idx_material_num_serial_num
// sqac:"index:unique/non-unique"
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
// CREATE COLUMN TABLE TableName
//		(Column Data_Type,
//		 Column Data_Type UNIQUE NOT NULL,
//		 Column Data_Type NOT NULL,
//		 Column Data_Type DEFAULT Default-Value,
//		 PRIMARY KEY(Column, Column));
//
// CREATE COLUMN TABLE TableName
//		(Column Data_Type PRIMARY KEY,
//		 Column1 Data_Type,
//		 Column2 Data_Type NOT NULL,
//		 Column Data_Type DEFAULT Default-Value,
//		 UNIQUE(Column1, Column2));
//
//
// CREATE COLUMN TABLE <table_name>
//		(<column_name> <num_data_type PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
//		 <column_name> <data_type> NOT NULL);
//
//====================================================================================
//	DDL defaults
//====================================================================================
// <default_value_clause> ::= DEFAULT <default_value_exp>
//
// <default_value_exp> ::=
//  NULL
//  | <string_literal>
//  | <signed_numeric_literal>
//  | <unsigned_numeric_literal>
//  | <datetime_value_function>
//
// <datetime_value_function> ::=
//  | CURRENT_DATE
//  | CURRENT_TIME
//  | CURRENT_TIMESTAMP
//  | CURRENT_UTCDATE
//  | CURRENT_UTCTIME
//  | CURRENT_UTCTIMESTAMP
//
//
//	select column_name, column_id from table_columns where table_name = 'SOME_NAMES'
//
// +-------------+-----------+
// | COLUMN_NAME | COLUMN_ID |
// +-------------+-----------+
// | ID          |    145210 |
// | NAME        |    145211 |
// +-------------+-----------+
//
// select * from sequences where sequence_name like ‘%145210%’
//
// +-------------+---------------------------+--------------+--------------+-----------+------------------+--------------+-----------+-------------------------------------------+------------+
// | SCHEMA_NAME |  SEQUENCE_NAME            | SEQUENCE_OID | START_NUMBER | MIN_VALUE | MAX_VALUE        | INCREMENT_BY | IS_CYCLED | RESET_BY_QUERY                            | CACHE_SIZE |
// +-------------+---------------------------+--------------+--------------+-----------+------------------+--------------+-----------+-------------------------------------------+------------+
// | SYSTEM      | _SYS_SEQUENCE_145210_#0_# |    145215    |     1        |    1      |       4611686018 |              |           |                                           |            |
// +-------------+---------------------------+--------------+--------------+-----------+------------------+--------------+-----------+-------------------------------------------+------------+
//
// SELECT SYSTEM."_SYS_SEQUENCE_145210_#0_#".CURRVAL FROM DUMMY;
//
// CREATE SEQUENCE STS.STS_SEQUENCE START WITH 1; -- OLD WAY?
//
// ALTER TABLE
//
// ALTER TABLE Table
//		ADD (Column Data_Type.
//			 Column Data_Type NOT NULL);
//
//
// DROP TABLE
//
// DROP TABLE Table;
//
//
// CREATE INDEX
//
// CREATE INDEX Access_Path
//		ON TableName (Column1, Column2);
//
// CREATE UNIQUE INDEX Access_Path
//		ON TableName (Column1, Column2);
//
// DROP INDEX AccessPath;
//
//
// Acces Metadata
//
// Sys.Tables
// Sys.Table_Columns
// Sys.Views
// Sys.View_Columns
// Sys.Indexes
// Sys.Index_Columns
//

type hdbSeqTyp struct {
	TableName string
	FieldName string
	Start     int
	SeqName   string
}

// GetDBName returns the first db in the list - should only be one?
func (hf *HDBFlavor) GetDBName() (dbName string) {

	row := hf.db.QueryRow("SELECT DATABASE_NAME FROM Sys.M_Databases;")
	if row != nil {
		err := row.Scan(&dbName)
		if err != nil {
			panic(err)
		}
	}
	return dbName
}

// createTables creates tables on the postgres database referenced
// by hf.DB.  This internally visible version is able to defer
// foreign-key creation if called with calledFromAlter = true.
func (hf *HDBFlavor) createTables(calledFromAlter bool, i ...interface{}) ([]ForeignKeyBuffer, error) {

	var tc TblComponents
	fkBuffer := make([]ForeignKeyBuffer, 0)

	// get the list of table Model{}s
	di := i[0].([]interface{})

	for t, ent := range di {

		ftr := reflect.TypeOf(ent)
		if hf.log {
			fmt.Println("CreateTable() entity type:", ftr)
		}

		// determine the table name
		tn := common.GetTableName(di[t])
		if tn == "" {
			return nil, fmt.Errorf("unable to determine table name in hf.CreateTables")
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
		tc = hf.buildTablSchema(tn, di[t])
		hf.QsLog(tc.tblSchema)

		// create the table on the db
		hf.db.MustExec(tc.tblSchema)

		// deal with the auto-incrementing by creating sequence manually
		for _, sq := range tc.seq {
			if hf.IsLog() {
				fmt.Printf("sequenceVals: %v\n", sq)
			}
			vals := strings.Split(sq.Value, " ")
			if len(vals) != 4 {
				return nil, fmt.Errorf("insufficient information to create sequence %s - got %v", sq.Name, sq.Value)
			}

			start, err := strconv.Atoi(vals[2])
			if err != nil {
				return nil, fmt.Errorf("%v is not a valid start value for sequence %s", vals[2], sq.Name)
			}

			seqDef := &hdbSeqTyp{
				TableName: strings.Replace(vals[0], "{", "", 1),
				FieldName: vals[1],
				Start:     start,
				SeqName:   strings.Replace(vals[3], "}", "", 1),
			}
			hf.CreateSequence(seqDef.SeqName, seqDef.Start)
			// delete the existing procedure if it exists
			// create a new procedure
			// for _, v := range tc.flDef {
			// 	fmt.Println("fldef:", v)
			// }
			// create a procedure
			// err = hf.createInsertSP(*seqDef, tc.flDef)
			// if err != nil {
			// 	return err
			// }
			// // procName := "procINSERT" + tn
			// procDDL := `CREATE PROCEDURE procINSERTDEPOT(
			// 			  IN col_one VARCHAR(255),
			// 			  IN col_two INTEGER,
			// 			  OUT id     INTEGER)
			// 			  LANGUAGE SQLSCRIPT AS
			// 			  BEGIN
			// 				id := 42;
			// 			  END;
			// `
			// fmt.Println(procDDL)

			// // attempt to create the procedure on the db
			// _, err = hf.db.Exec(procDDL)
			// if err != nil {
			// 	panic(err)
			// }
		}

		// create the table indices
		for k, in := range tc.ind {
			hf.CreateIndex(k, in)
		}

		// add foreign-key information to the buffer
		for _, v := range tc.fkey {
			fkv := ForeignKeyBuffer{
				ent:    ent,
				fkinfo: v,
			}
			fkBuffer = append(fkBuffer, fkv)
		}
	}

	// create the foreign-keys if any and if flag 'calledFromAlter = false'
	// attempt to create the foreign-key, but maybe do not hit a hard-fail
	// if FK creation fails.  When called from within AlterTable, creation
	// of new tables in the list is carried out first - by this method.  It
	// is possbile that a column required by for new foreign-key has yet to
	// be added to one of the tables pending alteration.  A soft failure
	// for FK creation issues seems approriate here, and the data for the
	// failed FK creation is added to the fkBuffer and passed back to the
	// called (AlterTable), where the FK creation can be tried again
	// following the completion of the table alterations.
	if calledFromAlter == false {
		for _, v := range fkBuffer {
			// fmt.Println()
			// fmt.Println()
			// fmt.Println("CALLING CreateForeignKey")
			// fmt.Println()
			// fmt.Println()
			err := hf.CreateForeignKey(v.ent, v.fkinfo.FromTable, v.fkinfo.RefTable, v.fkinfo.FromField, v.fkinfo.RefField)
			// fmt.Println("CreateForeignKey Got:", err)
			if err != nil {
				log.Printf("CreateForeignKey failed.  got: %v", err)
				return nil, err
			}
		}
	} else {
		return fkBuffer, nil // fkBuffer will always be !nil, but may be len==0
	}
	return nil, nil
}

// buildTableSchema builds a CREATE TABLE schema for the HDB DB and returns it
// to the caller, along with the components determined from the db and sqac
// struct-tags.  this method is used in CreateTables and AlterTables methods.
func (hf *HDBFlavor) buildTablSchema(tn string, ent interface{}) TblComponents {

	qt := hf.GetDBQuote()
	pKeys := ""
	var sequences []common.SqacPair
	var hdbSeq hdbSeqTyp

	indexes := make(map[string]IndexInfo)
	fKeys := make([]FKeyInfo, 0)
	tableSchema := fmt.Sprintf("CREATE COLUMN TABLE %s%s%s (", qt, tn, qt)

	// get a list of the field names, go-types and db attributes.
	// TagReader is a common function across db-flavors. For
	// this reason, the db-specific-data-type for each field
	// is determined locally.
	fldef, err := common.TagReader(ent, nil)
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
		col.fStart = 0

		// https://stackoverflow.com/questions/168736/how-do-you-set-a-default-value-for-a-mysql-datetime-column

		// if the field has been marked as NoDB, continue with the next field
		if fd.NoDB == true {
			continue
		}

		switch fd.UnderGoType {
		case "int64", "uint64":
			col.fType = "bigint"

		case "int32", "uint32", "int", "uint":
			col.fType = "int"

		case "int16", "uint16":
			col.fType = "smallint"

		case "int8", "uint8", "byte", "rune":
			col.fType = "tinyint"

		case "float64":
			col.fType = "double"

		case "float32":
			col.fType = "real"

		case "bool":
			col.fType = "boolean" // ??

		case "string":
			col.fType = "nvarchar(255)" //

		case "time.Time":
			col.fType = "timestamp"

		default:
			err := fmt.Errorf("go type %s is not presently supported", fldef[idx].FType)
			panic(err)
		}
		fldef[idx].FType = col.fType

		// read sqac tag pairs and apply
		// seqName := ""
		if !strings.Contains(fd.GoType, "*time.Time") {

			for _, p := range fd.SqacPairs {

				switch p.Name {
				case "primary_key":

					col.fPrimaryKey = "PRIMARY KEY"
					pKeys = fmt.Sprintf("%s %s%s%s,", pKeys, qt, fd.FName, qt)

					if p.Value == "inc" {
						hdbSeq.TableName = strings.ToUpper(tn)
						hdbSeq.FieldName = strings.ToUpper(fd.FName)
						hdbSeq.SeqName = fmt.Sprintf("SEQ_%s_%s", hdbSeq.TableName, hdbSeq.FieldName)
						if hdbSeq.Start == 0 {
							hdbSeq.Start = 1
						}
						col.fAutoInc = true
					}

				case "start":
					start, err := strconv.Atoi(p.Value)
					if err != nil {
						panic(err)
					}
					if start > 0 {
						hdbSeq.Start = start
						col.fStart = start
					}

				case "default":
					if fd.UnderGoType == "string" {
						col.fDefault = fmt.Sprintf("DEFAULT '%s'", p.Value)
					} else {
						col.fDefault = fmt.Sprintf("DEFAULT %s", p.Value)
					}

					if fd.UnderGoType == "time.Time" {
						switch p.Value {
						case "now()":
							p.Value = "CURRENT_UTCTIMESTAMP"
						case "eot":
							p.Value = "'9999-12-31 23:59:59.99999'"
						default:

						}
						col.fDefault = fmt.Sprintf("DEFAULT %s", p.Value)
					}

					// if fd.GoType == "bool" {
					// 	switch p.Value {
					// 	case "TRUE", "true":
					// 		p.Value = "1"
					// 	case "FALSE", "false":
					// 		p.Value = "0"
					// 	default:
					// 	}
					// 	col.fDefault = fmt.Sprintf("DEFAULT %s", p.Value)
					// }

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
						indexes = hf.processIndexTag(indexes, tn, fd.FName, "idx_", false, true)

					case "unique":
						indexes = hf.processIndexTag(indexes, tn, fd.FName, "idx_", true, true)

					default:
						indexes = hf.processIndexTag(indexes, tn, fd.FName, p.Value, false, false)
					}

				case "fkey":
					fKeys = hf.processFKeyTag(fKeys, tn, fd.FName, p.Value)

				default:

				}
			}
		} else { // *time.Time only supports default directive
			for _, p := range fd.SqacPairs {
				if p.Name == "default" {
					switch p.Value {
					case "now()":
						p.Value = "CURRENT_UTCTIMESTAMP"
					case "eot":
						p.Value = "'9999-12-31 23:59:59.99999'"
					default:

					}
					col.fDefault = fmt.Sprintf("DEFAULT %s", p.Value)
				}
			}
		}
		fldef[idx].FType = col.fType

		// record the sequence(no start-value)
		if col.fAutoInc {
			sequences = append(sequences, common.SqacPair{Name: hdbSeq.SeqName, Value: fmt.Sprintf("%v", hdbSeq)})
			hdbSeq = hdbSeqTyp{}
		}

		// add the current column to the schema
		tableSchema = tableSchema + fmt.Sprintf("%s%s%s %s", qt, col.fName, qt, col.fType)
		if col.fAutoInc == true {
			//tableSchema = tableSchema + " GENERATED BY DEFAULT AS IDENTITY"
			//if col.fStart > 1 {
			//	tableSchema = fmt.Sprintf("%s (START WITH %d)", tableSchema, col.fStart)
			//}
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
		fkey:      fKeys,
		pk:        pKeys,
		err:       err,
	}

	if hf.log {
		rc.Log()
	}
	return rc
}

// CreateTables creates tables on the hdb database referenced
// by hf.DB.
func (hf *HDBFlavor) CreateTables(i ...interface{}) error {

	// call createTables specifying that the call has not originated
	// from within the AlterTables(...) method.
	_, err := hf.createTables(false, i)
	if err != nil {
		return err
	}
	return nil
}

func (hf *HDBFlavor) createInsertSP(seqDef hdbSeqTyp, fldef []common.FieldDef) error {

	type tmplDataTyp struct {
		Header hdbSeqTyp
		Fields []common.FieldDef
	}

	var tmplData tmplDataTyp
	tmplData.Header = seqDef
	tmplData.Fields = fldef

	fmt.Println("seqDef:", seqDef)
	for _, v := range fldef {
		fmt.Println(v)
	}

	spTemplate := template.New("Entity Insert SP")
	spTemplate, err := template.ParseFiles("templates/createInsertSP.gotmpl")
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	err = spTemplate.Execute(&buf, tmplData)
	if err != nil {
		return err
	}

	procDDL := buf.String()
	fmt.Println(procDDL)

	// procName := "procInsert" + seqDef.tableName
	// procDDL := fmt.Sprintf(`CREATE PROCEDURE SMACLEOD.%s(
	// 						  IN col_one VARCHAR(255),
	// 						  IN col_two INTEGER,
	// 						  OUT id	 INTEGER)
	// 						  LANGUAGE SQLSCRIPT AS
	// 						  BEGIN
	//                             SELECT

	// 						  END;

	// 	`, procName)

	// attempt to create the procedure on the db
	_, err = hf.db.Exec(procDDL)
	if err != nil {
		return err
	}
	return nil
}

// DropTables drops tables on the db if they exist, based on
// the provided list of go struct definitions.
func (hf *HDBFlavor) DropTables(i ...interface{}) error {

	dropSchema := ""
	for t := range i {

		// determine the table name
		tn := common.GetTableName(i[t])
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
			dropSchema = dropSchema + fmt.Sprintf("DROP TABLE %s; ", strings.ToUpper(tn))
			hf.ProcessSchema(dropSchema)
			dropSchema = ""
		}
	}
	return nil
}

// AlterTables alters tables on the HDB database referenced
// by hf.DB.
func (hf *HDBFlavor) AlterTables(i ...interface{}) error {

	var err error
	fkBuffer := make([]ForeignKeyBuffer, 0)
	ci := make([]interface{}, 0)
	ai := make([]interface{}, 0)

	// construct create-table and alter-table buffers
	for t := range i {

		// ftr := reflect.TypeOf(ent)

		// determine the table name
		tn := common.GetTableName(i[t])
		if tn == "" {
			return fmt.Errorf("unable to determine table name in pf.AlterTables")
		}

		// if the table does not exist, add the Model{} definition to
		// the CreateTables buffer (ci).
		// if the table does exist, add the Model{} defintion to  the
		// AlterTables buffer (ai).
		if !hf.ExistsTable(tn) {
			ci = append(ci, i[t])
		} else {
			ai = append(ai, i[t])
		}
	}

	// if create-tables buffer 'ci' contains any entries, call createTables and
	// take note of any returned foreign-key definitions.
	if len(ci) > 0 {
		fkBuffer, err = hf.createTables(true, ci)
		if err != nil {
			return err
		}
	}

	// if alter-tables buffer 'ai' constains any entries, process the table
	// deltas and take note of any new foreign-key definitions.
	for t, ent := range ai {

		// ftr := reflect.TypeOf(ent)

		// determine the table name
		tn := common.GetTableName(ai[t])
		if tn == "" {
			return fmt.Errorf("unable to determine table name in hf.AlterTables")
		}

		// build the altered table schema and get its components
		tc := hf.buildTablSchema(tn, ai[t])

		// go through the latest version of the model and check each
		// field against its definition in the database.
		qt := hf.GetDBQuote()
		alterSchema := fmt.Sprintf("ALTER TABLE %s%s%s ADD (", qt, tn, qt)
		var cols []string

		for _, fd := range tc.flDef {
			// new columns first
			if !hf.ExistsColumn(tn, fd.FName) && fd.NoDB == false {

				colSchema := fmt.Sprintf("%s%s%s %s", qt, fd.FName, qt, fd.FType)
				for _, p := range fd.SqacPairs {
					switch p.Name {
					case "primary_key":
						// abort - adding primary key
						panic(fmt.Errorf("aborting - cannot add a primary-key (table-field %s-%s) through migration", tn, fd.FName))

					case "default":
						switch fd.UnderGoType {
						case "string":
							colSchema = fmt.Sprintf("%s DEFAULT '%s'", colSchema, p.Value)

						case "bool":
							switch p.Value {
							case "TRUE", "true":
								p.Value = "1"

							case "FALSE", "false":
								p.Value = "0"

							default:
								// nil
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

			alterSchema = strings.TrimSuffix(alterSchema, ",") + ");"
			hf.ProcessSchema(alterSchema)
		}

		// add indexes if required
		for k, v := range tc.ind {
			if !hf.ExistsIndex(v.TableName, k) {
				hf.CreateIndex(k, v)
			}
		}

		// add to the list of foreign-keys
		for _, v := range tc.fkey {
			fkb := ForeignKeyBuffer{
				ent:    ent,
				fkinfo: v,
			}
			fkBuffer = append(fkBuffer, fkb)
		}
	}

	// all table alterations and creations have been completed at this point, with the
	// exception of the foreign-key creations.  iterate over the fkBuffer, check for
	// the existance of each foreign-key and create those that do not yet exist.
	for _, v := range fkBuffer {
		fkn, err := common.GetFKeyName(v.ent, v.fkinfo.FromTable, v.fkinfo.RefTable, v.fkinfo.FromField, v.fkinfo.RefField)
		if err != nil {
			return err
		}
		fkExists, _ := hf.ExistsForeignKeyByName(v.ent, fkn)
		if !fkExists {
			err = hf.CreateForeignKey(v.ent, v.fkinfo.FromTable, v.fkinfo.RefTable, v.fkinfo.FromField, v.fkinfo.RefField)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
	}
	return nil
}

// ExistsTable checks the currently connected database and
// returns true if the named table is found to exist.
func (hf *HDBFlavor) ExistsTable(tn string) bool {

	n := 0
	etQuery := fmt.Sprintf("SELECT COUNT(*) FROM Sys.Tables WHERE TABLE_NAME = '%s';", strings.ToUpper(tn))
	hf.QsLog(etQuery)

	hf.db.QueryRow(etQuery).Scan(&n)
	if n > 0 {
		return true
	}
	return false
}

// ExistsIndex checks the connected database for the presence
// of the specified index.
func (hf *HDBFlavor) ExistsIndex(tn string, in string) bool {

	n := 0
	hf.QsLog("SELECT COUNT(*) FROM sys.indexes WHERE index_name=? AND table_name = ?;", strings.ToUpper(in), strings.ToUpper(tn))

	hf.db.QueryRow("SELECT COUNT(*) FROM sys.indexes WHERE index_name=? AND table_name = ?;", strings.ToUpper(in), strings.ToUpper(tn)).Scan(&n)
	if n > 0 {
		return true
	}
	return false
}

// DropIndex drops the specfied index on the connected database.
func (hf *HDBFlavor) DropIndex(tn string, in string) error {

	if hf.ExistsIndex(tn, in) {
		indexSchema := fmt.Sprintf("DROP INDEX %s;", strings.ToUpper(in))
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
	tn = strings.ToUpper(tn)
	if hf.ExistsTable(tn) {

		hf.QsLog("SELECT COUNT(*) FROM Sys.Table_Columns WHERE table_name = ? AND column_name = ?;", tn, strings.ToUpper(cn))

		// hf.db.QueryRow("SELECT COUNT(*) FROM Sys.Table_Columns WHERE Schema_Name = ? AND table_name = ? AND column_name = ?;", tn, cn).Scan(&n)
		hf.db.QueryRow("SELECT COUNT(*) FROM Sys.Table_Columns WHERE table_name = ? AND column_name = ?;", tn, strings.ToUpper(cn)).Scan(&n)
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

// getSequenceNames splits the incoming name field on the '+' sign
// and then assigns the resulting values to tn and fn respectively.
func (hf *HDBFlavor) getSequenceName(name string) (seqName string, err error) {

	// need table and column name - set as tn+fn in name
	tn := ""
	fn := ""
	var colID int

	names := strings.Split(name, "+")
	if len(names) == 2 {
		tn = strings.ToUpper(names[0])
		fn = strings.ToUpper(names[1])
	} else {
		return "", fmt.Errorf("expected table_name and field_name: got %v", names)
	}

	// identify the column_id associated with the sequence
	seqQuery := fmt.Sprintf("SELECT column_id FROM table_columns WHERE table_name = '%s' and column_name = '%s';", tn, fn)
	hf.QsLog(seqQuery)

	err = hf.db.QueryRowx(seqQuery).Scan(&colID)
	if err != nil {
		return "", err
	}

	if colID == 0 {
		return "", fmt.Errorf("could not find sequence for table %s / field %s", tn, fn)
	}

	// identify the requested sequence's name
	seqSearchVal := fmt.Sprintf("%%%v%%", colID) // % escapes %
	seqNameQuery := fmt.Sprintf("SELECT sequence_name FROM Sys.Sequences WHERE SEQUENCE_NAME LIKE '%s'", seqSearchVal)
	hf.QsLog(seqNameQuery)

	err = hf.db.QueryRow(seqNameQuery).Scan(&seqName)
	if err != nil {
		return "", err
	}
	return seqName, nil
}

// ExistsSequence is used to check for the existence of the named
// sequence in HDB.
func (hf *HDBFlavor) ExistsSequence(sn string) bool {

	// search for sequence by name
	seqCount := 0
	seqNameQuery := fmt.Sprintf("SELECT COUNT(*) FROM Sys.Sequences WHERE SEQUENCE_NAME = '%s'", sn)
	hf.QsLog(seqNameQuery)

	err := hf.db.QueryRow(seqNameQuery).Scan(&seqCount)
	if err != nil {
		panic(err)
	}

	if seqCount > 0 {
		return true
	}
	return false
}

// CreateSequence is used to create a sequence for use with HDB
// Identity columns.
func (hf *HDBFlavor) CreateSequence(sn string, start int) {

	// check for and drop existing sequence if exists
	if hf.ExistsSequence(sn) {
		err := hf.DropSequence(sn)
		if err != nil {
			panic(err)
		}
	}

	// build the sequence creation DDL
	crtSequence := fmt.Sprintf("CREATE SEQUENCE %s START WITH %d INCREMENT BY 1;", sn, start)
	hf.QsLog(crtSequence)

	// attempt to create the sequence on the db
	_, err := hf.db.Exec(crtSequence)
	if err != nil {
		panic(err)
	}
}

// DropSequence is used to drop an existing sequence in HDB.
func (hf *HDBFlavor) DropSequence(sn string) error {

	// build the sequence creation DDL
	dropSequence := fmt.Sprintf("DROP SEQUENCE %s;", sn)
	hf.QsLog(dropSequence)

	// attempt to drop the sequence from the db
	_, err := hf.db.Exec(dropSequence)
	if err != nil {
		return err
	}
	return nil
}

// GetNextSequenceValue is used primarily for testing.  It returns
// the next value of the named HDB identity (auto-increment) field
// in the named table.  this is not a reliable way to get the inserted
// id in a multi-transaction environment.
func (hf *HDBFlavor) GetNextSequenceValue(name string) (int, error) {

	var nextVal int
	nextQuery := fmt.Sprintf("SELECT %s.NEXTVAL FROM dummy;", name)
	hf.QsLog(nextQuery)

	err := hf.db.QueryRow(nextQuery).Scan(&nextVal)
	if err != nil {
		return 0, err
	}
	return nextVal, nil
}

// ExistsForeignKeyByName checks to see if the named foreign-key exists on the
// table corresponding to provided sqac model (i).
func (hf *HDBFlavor) ExistsForeignKeyByName(i interface{}, fkn string) (bool, error) {

	var count uint64
	tn := strings.ToUpper(common.GetTableName(i))

	fkQuery := fmt.Sprintf("SELECT COUNT(*) FROM Sys.Referential_Constraints WHERE TABLE_NAME='%s' AND CONSTRAINT_NAME='%s';", tn, strings.ToUpper(fkn))
	hf.QsLog(fkQuery)

	err := hf.Get(&count, fkQuery)
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
func (hf *HDBFlavor) ExistsForeignKeyByFields(i interface{}, ft, rt, ff, rf string) (bool, error) {

	fkn, err := common.GetFKeyName(i, ft, rt, ff, rf)
	if err != nil {
		return false, err
	}

	return hf.ExistsForeignKeyByName(i, strings.ToUpper(fkn))
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
	incKey := 0

	err := hf.BuildComponents(&info)
	if err != nil {
		return err
	}

	// build the hdb insert query
	insFlds := "("
	insVals := "("

	if hf.IsLog() {
		fmt.Println("info.incKeyName:", info.incKeyName)
	}

	for k, v := range info.fldMap {

		// pull an id - this is ugly, but hdb does not have a reliable
		// mechanism to report a new row-id.  Dynamic SQL in a SP may
		// be better, but opens the door for bad behaviour.  Try this
		// for now.
		if strings.Compare(k, info.incKeyName) == 0 {
			keyQuery := fmt.Sprintf("SELECT SEQ_%s_%s.NEXTVAL FROM DUMMY;", strings.ToUpper(info.tn), strings.ToUpper(info.incKeyName))
			err = hf.db.QueryRowx(keyQuery).Scan(&incKey)
			if err != nil {
				return err
			}
			insFlds = fmt.Sprintf("%s%s, ", insFlds, k)
			insVals = fmt.Sprintf("%s%v, ", insVals, incKey)
			continue
		}

		if v == "DEFAULT" {
			continue
		}
		insFlds = fmt.Sprintf("%s%s, ", insFlds, k)
		insVals = fmt.Sprintf("%s%s, ", insVals, v)
	}

	insFlds = strings.TrimSuffix(insFlds, ", ") + ")"
	insVals = strings.TrimSuffix(insVals, ", ") + ")"

	// build the hdb insert query
	insQuery := fmt.Sprintf("INSERT INTO %s %s VALUES %s;", info.tn, insFlds, insVals)
	hf.QsLog(insQuery)

	// clear the source data - deals with non-persistet columns
	e := reflect.ValueOf(info.ent).Elem()
	e.Set(reflect.Zero(e.Type()))

	// attempt the insert and read the result back into info.resultMap
	_, err = hf.db.Exec(insQuery)
	if err != nil {
		return err
	}

	selQuery := fmt.Sprintf("SELECT * FROM %s WHERE %s = %v;", info.tn, info.incKeyName, incKey)
	hf.QsLog(selQuery)

	err = hf.db.QueryRowx(selQuery).StructScan(info.ent) //.MapScan(info.resultMap) // SliceScan
	if err != nil {
		return err
	}
	info.entValue = reflect.ValueOf(info.ent)
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
	hf.QsLog(updQuery)

	// clear the source data - deals with non-persistet columns
	e := reflect.ValueOf(info.ent).Elem()
	e.Set(reflect.Zero(e.Type()))

	// attempt the update and check for errors
	_, err = hf.db.Exec(updQuery)
	if err != nil {
		return err
	}

	// read the updated row
	selQuery := fmt.Sprintf("SELECT * FROM %s WHERE %v;", info.tn, keyList)
	hf.QsLog(selQuery)

	err = hf.db.QueryRowx(selQuery).StructScan(info.ent) //.MapScan(info.resultMap) // SliceScan
	if err != nil {
		return err
	}
	info.entValue = reflect.ValueOf(info.ent)
	return nil
}

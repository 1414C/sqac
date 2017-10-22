package sqac

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	// "fmt"
)

// MySQLFlavor is a MySQL-specific implementation.
// Methods defined in the PublicDB interface of struct-type
// BaseFlavor are called by default for MySQLFlavor. If
// the method as it exists in the BaseFlavor implementation
// is not compatible with the schema-syntax required by
// MySQL, the method in question may be overridden.
// Overriding (redefining) a BaseFlavor method may be
// accomplished through the addition of a matching method
// signature and implementation on the MySQLFlavor
// struct-type.
type MySQLFlavor struct {
	BaseFlavor

	//================================================================
	// possible local MySQL-specific overrides
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

//================================================================
// PublicDB Interface methods implemented in BaseFlavor:
//
// ExistsTable(tn string) bool
// ExistsColumn(tn string, cn string) bool
// ExistsIndex(tn string, in string) bool
// CreateIndex(in string, index IndexInfo) error - experiemental
// DropIndex(in string) error - experimental
//
//
//================================================================

// CreateTables creates tables on the postgres database referenced
// by pf.DB.
func (myf *MySQLFlavor) CreateTables(i ...interface{}) error {

	for t, ent := range i {

		ftr := reflect.TypeOf(ent)
		if myf.log {
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
		if myf.ExistsTable(tn) {
			if myf.log {
				fmt.Printf("CreateTable - table %s exists - skipping...\n", tn)
			}
			continue
		}

		tc := myf.buildTablSchema(tn, i[t])
		if myf.log {
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
		myf.db.MustExec(tc.tblSchema)
		for _, sq := range tc.seq {
			start, _ := strconv.Atoi(sq.Value)
			myf.AlterSequenceStart(sq.Name, start)
		}
		for k, in := range tc.ind {
			myf.CreateIndex(k, in)
		}
	}
	return nil
}

// buildTableSchema builds a CREATE TABLE schema for the MySQL DB (MariaDB),
// and returns it to the caller, along with the components determined from
// the db and rgen struct-tags.  this method is used in CreateTables
// and AlterTables methods.
func (myf *MySQLFlavor) buildTablSchema(tn string, ent interface{}) TblComponents {

	qt := myf.GetDBQuote()
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

	// set the MySQL field-types and build the table schema,
	// as well as any other schemas that are needed to support
	// the table definition. In all cases any foreign-key or
	// index requirements must be deferred until all other
	// artifacts have been created successfully.
	// MySQL has more data-types than postgres - db ints are
	// signed / unsigned, and they account for integer sizes
	// down to 8 bits:
	// TINYINT = 1 byte
	// SMALLINT = 2 bytes
	// MEDIUMINT = 3 bytes
	// INT = 4 bytes
	// BIGINT = 8 bytes
	// DOUBLE = 8 bytes (floating-point, so we go for widest DB type here)
	// https://mariadb.com/kb/en/library/data-types/
	// future: https://dev.mysql.com/doc/refman/5.7/en/spatial-extensions.html
	for idx, fd := range fldef {

		var col ColComponents

		col.fName = fd.FName
		col.fType = ""
		col.fPrimaryKey = ""
		col.fDefault = ""
		col.fNullable = ""

		// things to deal with:
		// rgen:"primary_key:inc;start:55550000"
		// rgen:"nullable:false"
		// rgen:"default:0"
		// rgen:"index:idx_material_num_serial_num
		// rgen:"index:unique/non-unique"
		// timestamp syntax and functions
		// - pg now() equivalent
		// - pg make_timestamptz(9999, 12, 31, 23, 59, 59.9) equivalent

		// uint8  - TINYINT - range must be smaller than int8?
		// uint16 - SMALLINT
		// uint32 - INT
		// uint64 - BIGINT

		// int8  - TINYINT
		// int16 - SMALLINT
		// int32 - INT
		// int64 - BIGINT

		// float32 - DOUBLE
		// float64 - DOUBLE

		// bool - BOOLEAN - (alias for TINYINT(1))

		// rune - INT (32-bits - unicode and stuff)
		// byte - TINTINT - (8-bits and stuff)

		// string - VARCHAR(255) - (uses 1-byte for length-prefix in record prefix)
		// string - VARCHAR(256) - (uses 2-bytes for length-prefix; use for strings
		//                      	that may exceed 255 bytes->out to max 65,535 bytes)

		// TIMESTAMP - also look at YYYYMMDD format, which seems to be native

		// autoincrement - https://mariadb.com/kb/en/library/auto_increment/
		// spatial - POINT, MULTIPOINT, POLYGON (future)  https://mariadb.com/kb/en/library/geometry-types/

		// CREATE TABLE `test_default_four` (
		// 	`int16_key` bigint NOT NULL AUTO_INCREMENT,
		// 	 `int32_field` int NOT NULL DEFAULT 0,
		// 	`description` varchar(255) DEFAULT 'test',
		// 	PRIMARY KEY (`int16_key`)
		//   ) ENGINE=InnoDB DEFAULT CHARSET=latin1

		// https://stackoverflow.com/questions/168736/how-do-you-set-a-default-value-for-a-mysql-datetime-column

		switch fd.GoType {
		case "int", "int16", "int32", "rune":
			col.fType = "int"

		case "int64":
			col.fType = "bigint"

		case "int8":
			col.fType = "tinyint"

		case "uint", "uint16", "uint32":
			col.fType = "int unsigned"

		case "uint64":
			col.fType = "bigint unsigned"

		case "uint8", "byte":
			col.fType = "tinyint"

		case "float32", "float64":
			col.fType = "double"

		case "bool":
			col.fType = "boolean" // or tinyint(1)?

		case "string":
			col.fType = "varchar(255)" //

		case "time.Time", "*time.Time":
			col.fType = "timestamp"

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
					if fd.GoType == "time.Time" && p.Value == "eot" {
						p.Value = "TIMESTAMP('2003-12-31 12:00:00')"
						col.fDefault = fmt.Sprintf("DEFAULT %s", p.Value)
					}

				case "nullable":
					if p.Value == "false" {
						col.fNullable = "NOT NULL"
					}

				case "index":
					switch p.Value {
					case "non-unique":
						indexes = myf.processIndexTag(indexes, tn, fd.FName, "idx_"+fd.FName, false, true)

					case "unique":
						indexes = myf.processIndexTag(indexes, tn, fd.FName, "idx_"+fd.FName, true, true)

					default:
						indexes = myf.processIndexTag(indexes, tn, fd.FName, p.Value, false, false)
					}

				default:

				}
			}
		} else { // *time.Time only supports default directive
			for _, p := range fd.RgenPairs {
				if p.Name == "default" {
					if p.Value == "eot" {
						p.Value = "TIMESTAMP('2003-12-31 12:00:00')"
					}
					col.fDefault = fmt.Sprintf("DEFAULT %s", p.Value)
				}

			}
		}
		fldef[idx].FType = col.fType

		// add the current column to the schema
		tableSchema = tableSchema + fmt.Sprintf("%s%s%s %s", qt, col.fName, qt, col.fType)
		if col.fAutoInc == true {
			tableSchema = tableSchema + " AUTO_INCREMENT"
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
	tableSchema = tableSchema + " ENGINE=InnoDB DEFAULT CHARSET=latin1;"

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

// AlterTables alters tables on the MySQL database referenced
// by myf.DB.
func (myf *MySQLFlavor) AlterTables(i ...interface{}) error {

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
			return fmt.Errorf("unable to determine table name in myf.AlterTables")
		}

		// if the table does not exist, call CreateTables
		// if the table does exist, examine it and perform
		// alterations if neccessary
		if !myf.ExistsTable(tn) {
			myf.CreateTables(ent)
			continue
		}

		// build the altered table schema and get its components
		tc := myf.buildTablSchema(tn, i[t])
		if myf.log {
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

		// go through the latest version of the model and check each
		// field against its definition in the database.
		qt := myf.GetDBQuote()
		alterSchema := fmt.Sprintf("ALTER TABLE %s%s%s", qt, tn, qt)
		var cols []string

		for _, fd := range tc.flDef {
			// new columns first
			if !myf.ExistsColumn(tn, fd.FName) {

				colSchema := fmt.Sprintf("ADD COLUMN %s%s%s %s", qt, fd.FName, qt, fd.FType)
				for _, p := range fd.RgenPairs {
					switch p.Name {
					case "primary_key":
						// abort - adding primary key
						panic(fmt.Errorf("aborting - cannot add a primary-key (table-field %s-%s) through migration", tn, fd.FName))

					case "default":
						if fd.GoType == "string" {
							colSchema = fmt.Sprintf("%s DEFAULT '%s'", colSchema, p.Value)
						} else {
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
			myf.ProcessSchema(alterSchema)
		}

		// add indexes if required
		for k, v := range tc.ind {
			if !myf.ExistsIndex(v.TableName, k) {
				myf.CreateIndex(k, v)
			}
		}
	}
	return nil
}

// DropIndex drops the specfied index on the connected database.
func (myf *MySQLFlavor) DropIndex(tn string, in string) error {

	if myf.ExistsIndex(tn, in) {
		indexSchema := fmt.Sprintf("DROP INDEX %s ON %s;", in, tn)
		myf.ProcessSchema(indexSchema)
		return nil
	}
	return nil
}

// DestructiveResetTables drops tables on the MySQL db if they exist,
// as well as any related objects such as sequences.  this is
// useful if you wish to regenerated your table and the
// number-range used by an auto-incementing primary key.
func (myf *MySQLFlavor) DestructiveResetTables(i ...interface{}) error {

	err := myf.DropTables(i...)
	if err != nil {
		return err
	}
	err = myf.CreateTables(i...)
	if err != nil {
		return err
	}
	return nil
}

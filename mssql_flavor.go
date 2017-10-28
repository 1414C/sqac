package sqac

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// MSSQLFlavor is a MWSQL-specific implementation.
// Methods defined in the PublicDB interface of struct-type
// BaseFlavor are called by default for MSSQLFlavor. If
// the method as it exists in the BaseFlavor implementation
// is not compatible with the schema-syntax required by
// MSSQL, the method in question may be overridden.
// Overriding (redefining) a BaseFlavor method may be
// accomplished through the addition of a matching method
// signature and implementation on the MSSQLFlavor
// struct-type.
type MSSQLFlavor struct {
	BaseFlavor

	//================================================================
	// possible local MSSQL-specific overrides
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

// CreateTables creates tables on the mysql database referenced
// by msf.DB.
func (msf *MSSQLFlavor) CreateTables(i ...interface{}) error {

	for t, ent := range i {

		ftr := reflect.TypeOf(ent)
		if msf.log {
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
		if msf.ExistsTable(tn) {
			if msf.log {
				fmt.Printf("CreateTable - table %s exists - skipping...\n", tn)
			}
			continue
		}

		// build the create table schema and return all of the table info
		tc := msf.buildTablSchema(tn, i[t])

		// create the table on the db
		msf.db.MustExec(tc.tblSchema)
		for _, sq := range tc.seq {
			start, _ := strconv.Atoi(sq.Value)
			msf.AlterSequenceStart(sq.Name, start)
		}
		for k, in := range tc.ind {
			msf.CreateIndex(k, in)
		}
	}
	return nil
}

// buildTableSchema builds a CREATE TABLE schema for the MSSQL DB
// and returns it to the caller, along with the components determined from
// the db and rgen struct-tags.  this method is used in CreateTables
// and AlterTables methods.
func (msf *MSSQLFlavor) buildTablSchema(tn string, ent interface{}) TblComponents {

	qt := msf.GetDBQuote()
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

	// set the MSSQL field-types and build the table schema,
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
						indexes = msf.processIndexTag(indexes, tn, fd.FName, "idx_"+fd.FName, false, true)

					case "unique":
						indexes = msf.processIndexTag(indexes, tn, fd.FName, "idx_"+fd.FName, true, true)

					default:
						indexes = msf.processIndexTag(indexes, tn, fd.FName, p.Value, false, false)
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

	if msf.log {
		rc.Log()
	}
	return rc
}

// ExistsTable checks the currently connected database and
// returns true if the named table is found to exist.
func (msf *MSSQLFlavor) ExistsTable(tn string) bool {

	n := 0
	etQuery := fmt.Sprintf("SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = 'dbo' AND TABLE_NAME = '%s';", tn)
	msf.db.QueryRow(etQuery).Scan(&n)
	if n > 0 {
		return true
	}
	return false
}

// GetDBName returns the name of the currently connected db
func (msf *MSSQLFlavor) GetDBName() (dbName string) {

	row := msf.db.QueryRow("SELECT DB_NAME()")
	if row != nil {
		err := row.Scan(&dbName)
		if err != nil {
			panic(err)
		}
	}
	return dbName
}

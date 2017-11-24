package sqac

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/1414C/sqac/common"
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

// CreateTables creates tables on the mysql database referenced
// by myf.DB.
func (myf *MySQLFlavor) CreateTables(i ...interface{}) error {

	for t, ent := range i {

		ftr := reflect.TypeOf(ent)
		if myf.log {
			fmt.Println("CreateTable() entity type:", ftr)
		}

		// determine the table name
		tn := common.GetTableName(i[t])
		// tn := reflect.TypeOf(i[t]).String() // models.ProfileHeader{} for example
		// if strings.Contains(tn, ".") {
		// 	el := strings.Split(tn, ".")
		// 	tn = strings.ToLower(el[len(el)-1])
		// } else {
		// 	tn = strings.ToLower(tn)
		// }
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

		// build the create table schema and return all of the table info
		tc := myf.buildTablSchema(tn, i[t])

		// create the table on the db
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
	var sequences []common.RgenPair
	indexes := make(map[string]IndexInfo)
	tableSchema := fmt.Sprintf("CREATE TABLE %s%s%s (", qt, tn, qt)

	// get a list of the field names, go-types and db attributes.
	// TagReader is a common function across db-flavors. For
	// this reason, the db-specific-data-type for each field
	// is determined locally.
	fldef, err := common.TagReader(ent, nil)
	if err != nil {
		panic(err)
	}

	// set the MySQL field-types and build the table schema,
	// as well as any other schemas that are needed to support
	// the table definition. In all cases any foreign-key or
	// index requirements must be deferred until all other
	// artifacts have been created successfully.
	// https://mariadb.com/kb/en/library/data-types/
	// future: https://dev.mysql.com/doc/refman/5.7/en/spatial-extensions.html
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

		switch fd.UnderGoType { //fd.GoType {
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

		case "time.Time":
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
						sequences = append(sequences, common.RgenPair{Name: seqName, Value: p.Value})
					}

				case "default":
					if fd.UnderGoType == "string" {
						col.fDefault = fmt.Sprintf("DEFAULT '%s'", p.Value)
					} else {
						col.fDefault = fmt.Sprintf("DEFAULT %s", p.Value)
					}
					if fd.UnderGoType == "time.Time" && p.Value == "eot" {
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
						indexes = myf.processIndexTag(indexes, tn, fd.FName, "idx_", false, true)

					case "unique":
						indexes = myf.processIndexTag(indexes, tn, fd.FName, "idx_", true, true)

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

	// fill the return structure passing out the CREATE TABLE schema, and component info
	rc := TblComponents{
		tblSchema: tableSchema,
		flDef:     fldef,
		seq:       sequences,
		ind:       indexes,
		pk:        pKeys,
		err:       err,
	}

	if myf.log {
		rc.Log()
	}
	return rc
}

// AlterTables alters tables on the MySQL database referenced
// by myf.DB.
func (myf *MySQLFlavor) AlterTables(i ...interface{}) error {

	for t, ent := range i {

		// ftr := reflect.TypeOf(ent)

		// determine the table name
		tn := common.GetTableName(i[t])
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

		// go through the latest version of the model and check each
		// field against its definition in the database.
		qt := myf.GetDBQuote()
		alterSchema := fmt.Sprintf("ALTER TABLE %s%s%s", qt, tn, qt)
		var cols []string

		for _, fd := range tc.flDef {
			// new columns first
			if !myf.ExistsColumn(tn, fd.FName) && fd.NoDB == false {

				colSchema := fmt.Sprintf("ADD COLUMN %s%s%s %s", qt, fd.FName, qt, fd.FType)
				for _, p := range fd.RgenPairs {
					switch p.Name {
					case "primary_key":
						// abort - adding primary key
						panic(fmt.Errorf("aborting - cannot add a primary-key (table-field %s-%s) through migration", tn, fd.FName))

					case "default":
						if fd.UnderGoType == "string" {
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

// AlterSequenceStart may be used to make changes to the start value
// of the named auto_increment field in the MySQL database.  Note
// that this is intended to deal with auto-incrementing primary
// keys only.  It is possible in MySQL to setup a non-primary-key
// field as auto_increment as follows:
//
//   ALTER TABLE users ADD id INT UNSIGNED NOT NULL AUTO_INCREMENT, ADD INDEX (id);
//
//  This is not presently supported.
func (myf *MySQLFlavor) AlterSequenceStart(name string, start int) error {

	// ALTER TABLE users AUTO_INCREMENT=1001;
	alterSequenceSchema := fmt.Sprintf(" ALTER TABLE %s AUTO_INCREMENT=%d;", name, start)
	myf.ProcessSchema(alterSequenceSchema)
	return nil
}

// GetNextSequenceValue is used primarily for testing.  It returns
// the current value of the MySQL auto-increment field for the named
// table.
func (myf *MySQLFlavor) GetNextSequenceValue(name string) (int, error) {

	seq := 0
	if myf.ExistsTable(name) {

		seqQuery := fmt.Sprintf("SELECT `AUTO_INCREMENT` FROM  INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s';", myf.GetDBName(), name)
		err := myf.db.QueryRow(seqQuery).Scan(&seq)
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
func (myf *MySQLFlavor) Create(ent interface{}) error {

	var info CrudInfo
	info.ent = ent
	info.log = false
	info.mode = "C"

	err := myf.BuildComponents(&info)
	if err != nil {
		return err
	}

	// build the mysql insert query
	insQuery := fmt.Sprintf("INSERT INTO %s", info.tn)
	insQuery = fmt.Sprintf("%s %s VALUES %s;", insQuery, info.fList, info.vList)
	fmt.Println(insQuery)

	// attempt the insert and read the result back into info.resultMap
	result, err := myf.db.Exec(insQuery)
	if err != nil {
		return err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	selQuery := fmt.Sprintf("SELECT * FROM %s WHERE %s = %d LIMIT 1;", info.tn, info.incKeyName, lastID)
	err = myf.db.QueryRowx(selQuery).StructScan(info.ent) // .MapScan(info.resultMap) // SliceScan
	if err != nil {
		return err
	}
	info.entValue = reflect.ValueOf(info.ent)
	return nil
}

// Update an existing entity (single-row) on the database
func (myf *MySQLFlavor) Update(ent interface{}) error {

	var info CrudInfo
	info.ent = ent
	info.log = false
	info.mode = "U"

	err := myf.BuildComponents(&info)
	if err != nil {
		return err
	}

	keyList := ""
	for k, s := range info.keyMap {

		fType := reflect.TypeOf(s).String()
		if myf.IsLog() {
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
	_, err = myf.db.Exec(updQuery)
	if err != nil {
		return err
	}

	// read the updated row
	selQuery := fmt.Sprintf("SELECT * FROM %s WHERE %s LIMIT 1;", info.tn, keyList)
	err = myf.db.QueryRowx(selQuery).StructScan(info.ent) // .MapScan(info.resultMap) // SliceScan
	if err != nil {
		return err
	}
	info.entValue = reflect.ValueOf(info.ent)
	return nil
}

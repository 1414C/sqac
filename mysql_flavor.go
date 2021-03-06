package sqac

import (
	"fmt"
	"log"
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

// createTables creates tables on the postgres database referenced
// by pf.DB.  This internally visible version is able to defer
// foreign-key creation if called with calledFromAlter = true.
func (myf *MySQLFlavor) createTables(calledFromAlter bool, i ...interface{}) ([]ForeignKeyBuffer, error) {

	var tc TblComponents
	fkBuffer := make([]ForeignKeyBuffer, 0)

	// get the list of table Model{}s
	di := i[0].([]interface{})
	for t, ent := range di {

		ftr := reflect.TypeOf(ent)
		if myf.log {
			log.Println("CreateTable() entity type:", ftr)
		}

		// determine the table name
		tn := common.GetTableName(di[t])
		if tn == "" {
			return nil, fmt.Errorf("unable to determine table name in myf.createTables")
		}

		// if the table is found to exist, skip the creation
		// and move on to the next table in the list.
		if myf.ExistsTable(tn) {
			if myf.log {
				log.Printf("createTable - table %s exists - skipping...\n", tn)
			}
			continue
		}

		// build the create table schema and return all of the table info
		tc = myf.buildTablSchema(tn, di[t])
		myf.QsLog(tc.tblSchema)

		// create the table on the db
		myf.db.MustExec(tc.tblSchema)
		for _, sq := range tc.seq {
			start, _ := strconv.Atoi(sq.Value)
			myf.AlterSequenceStart(sq.Name, start)
		}

		// create the table indices
		for k, in := range tc.ind {
			myf.CreateIndex(k, in)
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
			err := myf.CreateForeignKey(v.ent, v.fkinfo.FromTable, v.fkinfo.RefTable, v.fkinfo.FromField, v.fkinfo.RefField)
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

// buildTableSchema builds a CREATE TABLE schema for the MySQL DB (MariaDB),
// and returns it to the caller, along with the components determined from
// the db and sqac struct-tags.  this method is used in CreateTables
// and AlterTables methods.
func (myf *MySQLFlavor) buildTablSchema(tn string, ent interface{}) TblComponents {

	qt := myf.GetDBQuote()
	pKeys := ""
	var sequences []common.SqacPair
	indexes := make(map[string]IndexInfo)
	fKeys := make([]FKeyInfo, 0)
	tableSchema := "CREATE TABLE " + qt + tn + qt + "("

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

		// read sqac tag pairs and apply
		seqName := ""
		if !strings.Contains(fd.GoType, "*time.Time") {

			for _, p := range fd.SqacPairs {

				switch p.Name {
				case "primary_key":

					col.fPrimaryKey = "PRIMARY KEY"
					pKeys = pKeys + " " + qt + fd.FName + qt + ","

					if p.Value == "inc" {
						// warn that user-specified db_type type will be ignored
						if col.uType != "" {
							log.Printf("WARNING: %s auto-incrementing primary-key field %s has user-specified db_type: %s  user-type is ignored. \n", common.GetTableName(ent), col.fName, col.uType)
							col.uType = ""
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
					if fd.UnderGoType == "time.Time" && p.Value == "eot()" {
						p.Value = "TIMESTAMP('2038-01-09 03:14:07')"
						col.fDefault = "DEFAULT " + p.Value
					}

				case "constraint":
					if p.Value == "unique" {
						col.fUniqueConstraint = "UNIQUE"
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

				case "fkey":
					fKeys = myf.processFKeyTag(fKeys, tn, fd.FName, p.Value)

				default:

				}
			}
		} else { // *time.Time only supports default directive
			for _, p := range fd.SqacPairs {
				if p.Name == "default" {
					if p.Value == "eot()" {
						// maximum mysql ts - consider using DATETIME,
						// although DT use would not be consistent with
						// the other db dialects in sqac.
						p.Value = "TIMESTAMP('2038-01-09 03:14:07')"
					}
					col.fDefault = "DEFAULT " + p.Value
				}

			}
		}
		fldef[idx].FType = col.fType

		// add the current column to the schema
		if col.uType != "" {
			tableSchema = tableSchema + qt + col.fName + qt + " " + col.uType
		} else {
			tableSchema = tableSchema + qt + col.fName + qt + " " + col.fType
		}
		if col.fAutoInc == true {
			tableSchema = tableSchema + " AUTO_INCREMENT"
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
		tableSchema = tableSchema + "PRIMARY KEY (" + pKeys + ") )"
	}
	tableSchema = tableSchema + " ENGINE=InnoDB DEFAULT CHARSET=latin1;"

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

	if myf.log {
		rc.Log()
	}
	return rc
}

// CreateTables creates tables on the mysql database referenced
// by myf.DB.
func (myf *MySQLFlavor) CreateTables(i ...interface{}) error {

	// call createTables specifying that the call has not originated
	// from within the AlterTables(...) method.
	_, err := myf.createTables(false, i)
	if err != nil {
		return err
	}
	return nil
}

// AlterTables alters tables on the MySQL database referenced
// by myf.DB.
func (myf *MySQLFlavor) AlterTables(i ...interface{}) error {

	var err error
	fkBuffer := make([]ForeignKeyBuffer, 0)
	ci := make([]interface{}, 0)
	ai := make([]interface{}, 0)

	// construct create-table and alter-table buffers
	for t := range i {

		// determine the table name
		tn := common.GetTableName(i[t])
		if tn == "" {
			return fmt.Errorf("unable to determine table name in pf.AlterTables")
		}

		// if the table does not exist, add the Model{} definition to
		// the CreateTables buffer (ci).
		// if the table does exist, add the Model{} definition to  the
		// AlterTables buffer (ai).
		if !myf.ExistsTable(tn) {
			ci = append(ci, i[t])
		} else {
			ai = append(ai, i[t])
		}
	}

	// if create-tables buffer 'ci' contains any entries, call createTables and
	// take note of any returned foreign-key definitions.
	if len(ci) > 0 {
		fkBuffer, err = myf.createTables(true, ci)
		if err != nil {
			return err
		}
	}

	// if alter-tables buffer 'ai' constains any entries, process the table
	// deltas and take note of any new foreign-key definitions.
	for t, ent := range ai {

		// determine the table name
		tn := common.GetTableName(ai[t])
		if tn == "" {
			return fmt.Errorf("unable to determine table name in myf.AlterTables")
		}

		// build the alter-table schema and get its components
		tc := myf.buildTablSchema(tn, ai[t])

		// go through the latest version of the model and check each
		// field against its definition in the database.
		qt := myf.GetDBQuote()
		alterSchema := "ALTER TABLE " + qt + tn + qt
		var cols []string

		for _, fd := range tc.flDef {
			// new columns first
			if !myf.ExistsColumn(tn, fd.FName) && fd.NoDB == false {

				colSchema := "ADD COLUMN " + qt + fd.FName + qt + " " + fd.FType
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
							colSchema = colSchema + " NOT NULL"
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
				alterSchema = alterSchema + " " + c
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
	// the existence of each foreign-key and create those that do not yet exist.
	for _, v := range fkBuffer {
		fkn, err := common.GetFKeyName(v.ent, v.fkinfo.FromTable, v.fkinfo.RefTable, v.fkinfo.FromField, v.fkinfo.RefField)
		if err != nil {
			return err
		}
		fkExists, _ := myf.ExistsForeignKeyByName(v.ent, fkn)
		if !fkExists {
			err = myf.CreateForeignKey(v.ent, v.fkinfo.FromTable, v.fkinfo.RefTable, v.fkinfo.FromField, v.fkinfo.RefField)
			if err != nil {
				log.Println(err)
				return err
			}
		}
	}
	return nil
}

// DropIndex drops the specfied index on the connected database.
func (myf *MySQLFlavor) DropIndex(tn string, in string) error {

	if myf.ExistsIndex(tn, in) {
		indexSchema := "DROP INDEX " + in + " ON " + tn + ";"
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
	alterSequenceSchema := " ALTER TABLE " + name + " AUTO_INCREMENT=" + strconv.Itoa(start) + ";"
	myf.ProcessSchema(alterSequenceSchema)
	return nil
}

// GetNextSequenceValue is used primarily for testing.  It returns
// the current value of the MySQL auto-increment field for the named
// table.
func (myf *MySQLFlavor) GetNextSequenceValue(name string) (int, error) {

	seq := 0
	if myf.ExistsTable(name) {

		seqQuery := "SELECT `AUTO_INCREMENT` FROM  INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = '" + myf.GetDBName() + "' AND TABLE_NAME = '" + name + "';"
		myf.QsLog(seqQuery)

		err := myf.db.QueryRow(seqQuery).Scan(&seq)
		if err != nil {
			return 0, err
		}
		return seq, nil
	}
	return seq, nil
}

// DropForeignKey drops a foreign-key on an existing column
func (myf *MySQLFlavor) DropForeignKey(i interface{}, ft, fkn string) error {

	// mysql: SELECT COUNT(*) FROM information_schema.table_constraints WHERE constraint_name='user__fk__store_id' AND table_name='client';
	schema := "ALTER TABLE " + ft + " DROP FOREIGN KEY " + fkn
	myf.QsLog(schema)

	_, err := myf.Exec(schema)
	if err != nil {
		return err
	}
	return nil
}

// ExistsForeignKeyByName checks to see if the named foreign-key exists on the
// table corresponding to provided sqac model (i).
func (myf *MySQLFlavor) ExistsForeignKeyByName(i interface{}, fkn string) (bool, error) {

	var count uint64
	tn := common.GetTableName(i)

	fkQuery := "SELECT COUNT(*) FROM information_schema.table_constraints WHERE constraint_name='" + fkn + "' AND table_name='" + tn + "';"
	myf.QsLog(fkQuery)

	err := myf.Get(&count, fkQuery)
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
func (myf *MySQLFlavor) ExistsForeignKeyByFields(i interface{}, ft, rt, ff, rf string) (bool, error) {

	fkn, err := common.GetFKeyName(i, ft, rt, ff, rf)
	if err != nil {
		return false, err
	}
	return myf.ExistsForeignKeyByName(i, fkn)
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
	insQuery := "INSERT INTO " + info.tn + " " + info.fList + " VALUES " + info.vList + ";"
	myf.QsLog(insQuery)

	// clear the source data - deals with non-persistet columns
	e := reflect.ValueOf(info.ent).Elem()
	e.Set(reflect.Zero(e.Type()))

	// attempt the insert and read the result back into info.resultMap
	result, err := myf.db.Exec(insQuery)
	if err != nil {
		return err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	selQuery := "SELECT * FROM " + info.tn + " WHERE " + info.incKeyName + " = " + strconv.FormatInt(lastID, 10) + " LIMIT 1;"
	myf.QsLog(selQuery)

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
			log.Printf("CRUD UPDATE key: %v, value: %v\n", k, s)
			log.Println("CRUD UPDATED TYPE:", fType)
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

	updQuery := "UPDATE " + info.tn + " SET " + colList + " WHERE " + keyList + ";"
	myf.QsLog(updQuery)

	// clear the source data - deals with non-persistet columns
	e := reflect.ValueOf(info.ent).Elem()
	e.Set(reflect.Zero(e.Type()))

	// attempt the update and check for errors
	_, err = myf.db.Exec(updQuery)
	if err != nil {
		return err
	}

	// read the updated row
	selQuery := "SELECT * FROM " + info.tn + " WHERE " + keyList + " LIMIT 1;"
	myf.QsLog(selQuery)

	err = myf.db.QueryRowx(selQuery).StructScan(info.ent) // .MapScan(info.resultMap) // SliceScan
	if err != nil {
		return err
	}
	info.entValue = reflect.ValueOf(info.ent)
	return nil
}

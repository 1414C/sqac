package sqac

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/1414C/sqac/common"
	"github.com/jmoiron/sqlx"
)

// MSSQLFlavor is a MSSQL-specific implementation.
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
// sqac:"primary_key:inc;start:55550000"
// sqac:"nullable:false"
// sqac:"default:0"
// sqac:"index:idx_material_num_serial_num
// sqac:"index:unique/non-unique"
// timestamp syntax and functions
// - pg now() equivalent
// - pg make_timestamptz(9999, 12, 31, 23, 59, 59.9) equivalent

// createTables creates tables on the postgres database referenced
// by pf.DB.  This internally visible version is able to defer
// foreign-key creation if called with calledFromAlter = true.
func (msf *MSSQLFlavor) createTables(calledFromAlter bool, i ...interface{}) ([]ForeignKeyBuffer, error) {

	var tc TblComponents
	fkBuffer := make([]ForeignKeyBuffer, 0)

	// get the list of table Model{}s
	di := i[0].([]interface{})
	for t, ent := range di {

		ftr := reflect.TypeOf(ent)
		if msf.log {
			log.Println("CreateTable() entity type:", ftr)
		}

		// determine the table name
		tn := common.GetTableName(di[t])
		if tn == "" {
			return nil, fmt.Errorf("unable to determine table name in myf.CreateTables")
		}

		// if the table is found to exist, skip the creation
		// and move on to the next table in the list.
		if msf.ExistsTable(tn) {
			if msf.log {
				fmt.Printf("createTable - table %s exists - skipping...\n", tn)
			}
			continue
		}

		// build the create table schema and return all of the table info
		tc = msf.buildTablSchema(tn, di[t])
		msf.QsLog(tc.tblSchema)

		// create the table on the db
		msf.db.MustExec(tc.tblSchema)
		for _, sq := range tc.seq {
			start, _ := strconv.Atoi(sq.Value)
			msf.AlterSequenceStart(sq.Name, start)
		}

		// create the table indices
		for k, in := range tc.ind {
			msf.CreateIndex(k, in)
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
			err := msf.CreateForeignKey(v.ent, v.fkinfo.FromTable, v.fkinfo.RefTable, v.fkinfo.FromField, v.fkinfo.RefField)
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

// buildTableSchema builds a CREATE TABLE schema for the MSSQL DB
// and returns it to the caller, along with the components determined from
// the db and sqac struct-tags.  this method is used in CreateTables
// and AlterTables methods.
func (msf *MSSQLFlavor) buildTablSchema(tn string, ent interface{}) TblComponents {

	qt := msf.GetDBQuote()
	pKeys := ""
	var sequences []common.SqacPair
	indexes := make(map[string]IndexInfo)
	fKeys := make([]FKeyInfo, 0)
	tableSchema := fmt.Sprintf("CREATE TABLE %s%s%s (", qt, tn, qt)

	// get a list of the field names, go-types and db attributes.
	// TagReader is a common function across db-flavors. For
	// this reason, the db-specific-data-type for each field
	// is determined locally.
	fldef, err := common.TagReader(ent, nil)
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

		switch fd.UnderGoType {
		case "int64", "uint64":
			col.fType = "bigint"

		case "int32", "uint32", "int", "uint":
			col.fType = "int"

		case "int16", "uint16":
			col.fType = "smallint"

		case "int8", "uint8", "byte", "rune":
			col.fType = "tinyint"

		case "float32", "float64":
			col.fType = "numeric(38,7)" // default precision is 18

		case "bool":
			col.fType = "bit"

		case "string":
			col.fType = "varchar(255)" //

		case "time.Time":
			col.fType = "datetime2"

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
						sequences = append(sequences, common.SqacPair{Name: seqName, Value: p.Value})
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
							p.Value = "GETDATE()"
						case "eot":
							p.Value = "'9999-12-31 23:59:59.999'"
						default:

						}
						col.fDefault = fmt.Sprintf("DEFAULT %s", p.Value)
					}

					if fd.UnderGoType == "bool" {
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

				case "constraint":
					if p.Value == "unique" {
						col.fUniqueConstraint = "UNIQUE"
					}

				case "index":
					switch p.Value {
					case "non-unique":
						indexes = msf.processIndexTag(indexes, tn, fd.FName, "idx_", false, true)

					case "unique":
						indexes = msf.processIndexTag(indexes, tn, fd.FName, "idx_", true, true)

					default:
						indexes = msf.processIndexTag(indexes, tn, fd.FName, p.Value, false, false)
					}

				case "fkey":
					fKeys = msf.processFKeyTag(fKeys, tn, fd.FName, p.Value)

				default:

				}
			}
		} else { // *time.Time only supports default directive
			for _, p := range fd.SqacPairs {
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

	if msf.log {
		rc.Log()
	}
	return rc
}

// CreateTables creates tables on the mysql database referenced
// by msf.DB.
func (msf *MSSQLFlavor) CreateTables(i ...interface{}) error {

	// call createTables specifying that the call has not originated
	// from within the AlterTables(...) method.
	_, err := msf.createTables(false, i)
	if err != nil {
		return err
	}
	return nil
}

// DropTables drops tables on the db if they exist, based on
// the provided list of go struct definitions.
func (msf *MSSQLFlavor) DropTables(i ...interface{}) error {

	dropSchema := ""
	for t := range i {

		// determine the table name
		tn := common.GetTableName(i[t])
		if tn == "" {
			return fmt.Errorf("unable to determine table name in msf.DropTables")
		}

		// if the table is found to exist, add a DROP statement
		// to the dropSchema string and move on to the next
		// table in the list.
		if msf.ExistsTable(tn) {
			if msf.log {
				fmt.Printf("table %s exists - adding to drop schema...\n", tn)
			}
			// submit 1 at a time for mysql
			dropSchema = dropSchema + fmt.Sprintf("DROP TABLE %s; ", tn)
			msf.ProcessSchema(dropSchema)
			dropSchema = ""
		}
	}
	return nil
}

// AlterTables alters tables on the MSSQL database referenced
// by msf.DB.
func (msf *MSSQLFlavor) AlterTables(i ...interface{}) error {

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
		if !msf.ExistsTable(tn) {
			ci = append(ci, i[t])
		} else {
			ai = append(ai, i[t])
		}
	}

	// if create-tables buffer 'ci' contains any entries, call createTables and
	// take note of any returned foreign-key definitions.
	if len(ci) > 0 {
		fkBuffer, err = msf.createTables(true, ci)
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
			return fmt.Errorf("unable to determine table name in msf.AlterTables")
		}

		// if the table does not exist, call CreateTables
		// if the table does exist, examine it and perform
		// alterations if neccessary
		if !msf.ExistsTable(tn) {
			msf.CreateTables(ent)
			continue
		}

		// build the altered table schema and get its components
		tc := msf.buildTablSchema(tn, ai[t])

		// go through the latest version of the model and check each
		// field against its definition in the database.
		qt := msf.GetDBQuote()
		alterSchema := fmt.Sprintf("ALTER TABLE %s%s%s ADD ", qt, tn, qt)
		var cols []string

		for _, fd := range tc.flDef {
			// new columns first
			if !msf.ExistsColumn(tn, fd.FName) && fd.NoDB == false {

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
			msf.ProcessSchema(alterSchema)
		}

		// add indexes if required
		for k, v := range tc.ind {
			if !msf.ExistsIndex(v.TableName, k) {
				msf.CreateIndex(k, v)
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
		fkExists, _ := msf.ExistsForeignKeyByName(v.ent, fkn)
		if !fkExists {
			err = msf.CreateForeignKey(v.ent, v.fkinfo.FromTable, v.fkinfo.RefTable, v.fkinfo.FromField, v.fkinfo.RefField)
			if err != nil {
				log.Println(err)
				return err
			}
		}
	}
	return nil
}

// ExistsTable checks the currently connected database and
// returns true if the named table is found to exist.
func (msf *MSSQLFlavor) ExistsTable(tn string) bool {

	n := 0
	etQuery := fmt.Sprintf("SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = 'dbo' AND TABLE_NAME = '%s';", tn)
	msf.QsLog(etQuery)
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

// ExistsIndex checks the connected database for the presence
// of the specified index.
func (msf *MSSQLFlavor) ExistsIndex(tn string, in string) bool {

	n := 0
	// msf.db.QueryRow("SELECT count(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE table_schema = ? AND table_name = ? AND index_name = ?", bf.GetDBName(), tn, in).Scan(&n)
	msf.QsLog("SELECT COUNT(*) FROM sys.indexes WHERE name=? AND object_id = OBJECT_ID(?);", in, tn)
	msf.db.QueryRow("SELECT COUNT(*) FROM sys.indexes WHERE name=? AND object_id = OBJECT_ID(?);", in, tn).Scan(&n)
	if n > 0 {
		return true
	}
	return false
}

// DropIndex drops the specfied index on the connected database.
func (msf *MSSQLFlavor) DropIndex(tn string, in string) error {

	if msf.ExistsIndex(tn, in) {
		indexSchema := fmt.Sprintf("DROP INDEX %s ON %s;", in, tn)
		msf.ProcessSchema(indexSchema)
		return nil
	}
	return nil
}

// ExistsColumn checks the currently connected database and
// returns true if the named table-column is found to exist.
// this checks the column name only, not the column data-type
// or properties.
func (msf *MSSQLFlavor) ExistsColumn(tn string, cn string) bool {

	n := 0
	if msf.ExistsTable(tn) {
		msf.QsLog("SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = ? AND column_name = ?;", tn, cn)
		msf.db.QueryRow("SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = ? AND column_name = ?;", tn, cn).Scan(&n)
		if n > 0 {
			return true
		}
	}
	return false
}

// DestructiveResetTables drops tables on the MSSQL db if they exist,
// as well as any related objects such as sequences.  this is
// useful if you wish to regenerated your table and the
// number-range used by an auto-incementing primary key.
func (msf *MSSQLFlavor) DestructiveResetTables(i ...interface{}) error {

	err := msf.DropTables(i...)
	if err != nil {
		return err
	}
	err = msf.CreateTables(i...)
	if err != nil {
		return err
	}
	return nil
}

// AlterSequenceStart may be used to make changes to the start value of the
// named identity-field on the currently connected MSSQL database.
func (msf *MSSQLFlavor) AlterSequenceStart(name string, start int) error {

	// reseed the primary key
	// DBCC CHECKIDENT ('dbo.depot', RESEED, 50000000);
	alterSequenceSchema := fmt.Sprintf("DBCC CHECKIDENT (%s, RESEED, %d)", name, start)
	msf.ProcessSchema(alterSequenceSchema)
	return nil
}

// GetNextSequenceValue is used primarily for testing.  It returns
// the current value of the MSSQL identity (auto-increment) field for
// the named table.
func (msf *MSSQLFlavor) GetNextSequenceValue(name string) (int, error) {

	seq := 0
	if msf.ExistsTable(name) {
		seqQuery := fmt.Sprintf("SELECT IDENT_CURRENT( '%s' );", name)
		msf.QsLog(seqQuery)
		err := msf.db.QueryRow(seqQuery).Scan(&seq)
		if err != nil {
			return 0, err
		}
		return seq, nil
	}
	return seq, nil
}

// ExistsForeignKeyByName checks to see if the named foreign-key exists on the
// table corresponding to provided sqac model (i).
func (msf *MSSQLFlavor) ExistsForeignKeyByName(i interface{}, fkn string) (bool, error) {

	var count uint64

	fkQuery := fmt.Sprintf("SELECT COUNT(*) FROM INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS WHERE CONSTRAINT_NAME = '%s';", fkn)
	msf.QsLog(fkQuery)

	err := msf.Get(&count, fkQuery)
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
func (msf *MSSQLFlavor) ExistsForeignKeyByFields(i interface{}, ft, rt, ff, rf string) (bool, error) {

	fkn, err := common.GetFKeyName(i, ft, rt, ff, rf)
	if err != nil {
		return false, err
	}

	return msf.ExistsForeignKeyByName(i, fkn)
}

//================================================================
// CRUD ops
//================================================================

// Create the entity (single-row) on the database
func (msf *MSSQLFlavor) Create(ent interface{}) error {

	var info CrudInfo
	info.ent = ent
	info.log = false
	info.mode = "C"

	err := msf.BuildComponents(&info)
	if err != nil {
		return err
	}

	// build the mssql insert query
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

	// build the mssql insert query
	insQuery := fmt.Sprintf("INSERT INTO %s %s VALUES %s;", info.tn, insFlds, insVals)

	// clear the source data - deals with non-persistet columns
	e := reflect.ValueOf(info.ent).Elem()
	e.Set(reflect.Zero(e.Type()))

	// attempt the insert and read the result back into info.resultMap
	result, err := msf.db.Exec(insQuery)
	if err != nil {
		return err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	selQuery := fmt.Sprintf("SELECT * FROM %s WHERE %s = %v;", info.tn, info.incKeyName, lastID)
	msf.QsLog(selQuery)
	err = msf.db.QueryRowx(selQuery).StructScan(info.ent) //.MapScan(info.resultMap) // SliceScan
	if err != nil {
		return err
	}
	info.entValue = reflect.ValueOf(info.ent)
	return nil
}

// Update an existing entity (single-row) on the database
func (msf *MSSQLFlavor) Update(ent interface{}) error {

	var info CrudInfo
	info.ent = ent
	info.log = false
	info.mode = "U"

	err := msf.BuildComponents(&info)
	if err != nil {
		return err
	}

	keyList := ""
	for k, s := range info.keyMap {

		fType := reflect.TypeOf(s).String()
		if msf.IsLog() {
			log.Printf("key: %v, value: %v\n", k, s)
			log.Println("TYPE:", fType)
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
	msf.QsLog(updQuery)

	// clear the source data - deals with non-persistet columns
	e := reflect.ValueOf(info.ent).Elem()
	e.Set(reflect.Zero(e.Type()))

	// attempt the update and check for errors
	_, err = msf.db.Exec(updQuery)
	if err != nil {
		return err
	}

	// read the updated row
	selQuery := fmt.Sprintf("SELECT * FROM %s WHERE %v;", info.tn, keyList)
	msf.QsLog(selQuery)
	err = msf.db.QueryRowx(selQuery).StructScan(info.ent) // .MapScan(info.resultMap) // SliceScan
	if err != nil {
		return err
	}
	info.entValue = reflect.ValueOf(info.ent)
	return nil
}

// GetEntitiesWithCommands is the experimental replacement for all get-set ops
func (msf *MSSQLFlavor) GetEntitiesWithCommands(ents interface{}, params []common.GetParam, cmdMap map[string]interface{}) (interface{}, error) {

	var err error
	var count uint64
	var row *sqlx.Row
	paramString := ""
	selQuery := ""

	// get the underlying data type of the interface{}
	entTypeElem := reflect.TypeOf(ents).Elem()
	// fmt.Println("entTypeElem:", entTypeElem)

	// create a struct from the type
	testVar := reflect.New(entTypeElem)

	// determine the db table name
	tn := common.GetTableName(ents)

	// are there any parameters to include in the query?
	var pv []interface{}
	if params != nil && len(params) > 0 {
		paramString = " WHERE"
		for i := range params {
			paramString = fmt.Sprintf("%s %s %s ? %s", paramString, common.CamelToSnake(params[i].FieldName), params[i].Operand, params[i].NextOperator)
			pv = append(pv, params[i].ParamValue)
		}
	}
	if msf.log {
		log.Println("constructed paramString:", paramString)
	}

	// received a $count command?  this supercedes all, as it should not
	// be mixed with any other $<commands>.
	_, ok := cmdMap["count"]
	if ok {
		if paramString == "" {
			selQuery = fmt.Sprintf("SELECT COUNT(*) FROM %s;", tn)
			msf.QsLog(selQuery)
			row = msf.ExecuteQueryRowx(selQuery)
		} else {
			selQuery = fmt.Sprintf("SELECT COUNT(*) FROM %s%s;", tn, paramString)
			msf.QsLog(selQuery)
			row = msf.ExecuteQueryRowx(selQuery, pv...)
		}

		err = row.Scan(&count)
		if err != nil {
			log.Fatal(err)
		}
		return count, nil
	}

	// no $count command - build query
	var obString string
	var limitString string
	var offsetString string
	var adString string

	// received $orderby command?
	obField, ok := cmdMap["orderby"]
	if ok {
		obString = fmt.Sprintf(" ORDER BY %s", obField.(string))
	}

	// received $asc command?
	_, ok = cmdMap["asc"]
	if ok {
		adString = " ASC"
	}

	// received $desc command?
	_, ok = cmdMap["desc"]
	if ok {
		adString = " DESC"
	}

	// received $offset command?
	offField, ok := cmdMap["offset"]
	if ok {
		offsetString = fmt.Sprintf(" OFFSET %v ROWS", offField)
	}

	// received $limit command?
	limField, ok := cmdMap["limit"]
	if ok {
		if offsetString != "" {
			limitString = fmt.Sprintf(" FETCH NEXT %v ROWS ONLY", limField)
		} else {
			limitString = fmt.Sprintf("TOP(%v)", limField)
		}
	}

	// -- SELECT COUNT(*) FROM library;
	// -- SELECT * FROM library;
	// -- SELECT * FROM library LIMIT 2;
	// -- SELECT * FROM library OFFSET 2;
	// -- SELECT * FROM library LIMIT 2 OFFSET 1;
	// -- SELECT * FROM library ORDER BY ID DESC;
	// -- SELECT * FROM library ORDER BY ID ASC;
	// -- SELECT * FROM library ORDER BY name ASC;
	// -- SELECT * FROM library ORDER BY ID ASC LIMIT 2 OFFSET 2;

	// if $asc or $desc were specifed with no $orderby, default to order by id
	if obString == "" && adString != "" {
		obString = " ORDER BY id"
	}

	if offsetString != "" && obString == "" {
		obString = " ORDER BY id"
	}

	if limitString != "" && offsetString == "" {
		selQuery = fmt.Sprintf("SELECT %v * FROM %s%s", limitString, tn, paramString)
	} else {
		selQuery = fmt.Sprintf("SELECT * FROM %s%s", tn, paramString)
	}
	selQuery = msf.db.Rebind(selQuery)

	// use SELECT (TOP n) * ...
	if limitString != "" && offsetString == "" {
		selQuery = fmt.Sprintf("%s%s%s;", selQuery, obString, adString)
	} else {
		selQuery = fmt.Sprintf("%s%s%s%s%s;", selQuery, obString, adString, offsetString, limitString)
	}
	// selQuery = fmt.Sprintf("%s%s%s%s%s;", selQuery, obString, adString, offsetString, limitString)
	msf.QsLog(selQuery)

	// read the rows
	// fmt.Println("pv...", pv)
	rows, err := msf.db.Queryx(selQuery, pv...)
	if err != nil {
		log.Printf("GetEntities for table &s returned error: %v\n", err.Error())
		return nil, err
	}
	defer rows.Close()

	// iterate over the rows collection and put the results
	// into the ents interface (slice)
	entsv := reflect.ValueOf(ents)
	for rows.Next() {
		err = rows.StructScan(testVar.Interface())
		if err != nil {
			log.Println("scan error:", err)
			return nil, err
		}
		// fmt.Println(testVar)
		entsv = reflect.Append(entsv, testVar.Elem())
	}

	ents = entsv.Interface()
	// fmt.Println("ents:", ents)
	return entsv.Interface(), nil
}

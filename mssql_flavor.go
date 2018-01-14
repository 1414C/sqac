package sqac

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/1414C/sqac/common"
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

// CreateTables creates tables on the mysql database referenced
// by msf.DB.
func (msf *MSSQLFlavor) CreateTables(i ...interface{}) error {

	for t, ent := range i {

		ftr := reflect.TypeOf(ent)
		if msf.log {
			fmt.Println("CreateTable() entity type:", ftr)
		}

		// determine the table name
		tn := common.GetTableName(i[t])
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
		msf.QsLog(tc.tblSchema)

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

// AlterTables alters tables on the MSSQL database referenced
// by msf.DB.
func (msf *MSSQLFlavor) AlterTables(i ...interface{}) error {

	for t, ent := range i {

		// ftr := reflect.TypeOf(ent)

		// determine the table name
		tn := common.GetTableName(i[t])
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
		tc := msf.buildTablSchema(tn, i[t])

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
	}
	return nil
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
	fmt.Println(insQuery)

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
func (msf *MSSQLFlavor) GetEntitiesWithCommands(ents interface{}, cmdMap map[string]interface{}) (interface{}, error) {

	fmt.Println()
	fmt.Println("GetEntitiesWithCommands received cmdMap:", cmdMap)
	fmt.Println()
	return nil, nil
}

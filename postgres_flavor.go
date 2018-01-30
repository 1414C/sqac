package sqac

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/1414C/sqac/common"
)

// PostgresFlavor is a postgres-specific implementation.
// Methods defined in the PublicDB interface of struct-type
// BaseFlavor are called by default for PostgresFlavor. If
// the method as it exists in the BaseFlavor implementation
// is not compatible with the schema-syntax required by
// Postgres, the method in question may be overridden.
// Overriding (redefining) a BaseFlavor method may be
// accomplished through the addition of a matching method
// signature and implementation on the PostgresFlavor
// struct-type.
type PostgresFlavor struct {
	BaseFlavor

	//================================================================
	// possible local Postgres-specific overrides
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
func (pf *PostgresFlavor) GetDBName() (dbName string) {

	row := pf.db.QueryRow("SELECT current_database();")
	if row != nil {
		err := row.Scan(&dbName)
		if err != nil {
			panic(err)
		}
	}
	return dbName
}

// createTables creates tables on the postgres database referenced
// by pf.DB.  This internally visible version is able to defer
// foreign-key creation if called with calledFromAlter = true.
func (pf *PostgresFlavor) createTables(calledFromAlter bool, i ...interface{}) ([]ForeignKeyBuffer, error) {

	var tc TblComponents
	fkBuffer := make([]ForeignKeyBuffer, 0)

	// get the list of table Model{}s
	di := i[0].([]interface{})
	for t, ent := range di {

		ftr := reflect.TypeOf(ent)
		if pf.log {
			fmt.Println("CreateTable() entity type:", ftr)
		}

		// determine the table name
		tn := common.GetTableName(di[t])
		if tn == "" {
			return nil, fmt.Errorf("unable to determine table name in pf.createTables")
		}

		// if the table is found to exist, skip the creation
		// and move on to the next table in the list.
		if pf.ExistsTable(tn) {
			if pf.log {
				fmt.Printf("createTable - table %s exists - skipping...\n", tn)
			}
			continue
		}

		// build the create table schema and return all of the table info
		tc = pf.buildTablSchema(tn, di[t])
		pf.QsLog(tc.tblSchema)

		// create the table on the db
		pf.db.MustExec(tc.tblSchema)
		for _, sq := range tc.seq {
			start, _ := strconv.Atoi(sq.Value)
			pf.AlterSequenceStart(sq.Name, start)
		}

		// create the table indices
		for k, in := range tc.ind {
			pf.CreateIndex(k, in)
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
			err := pf.CreateForeignKey(v.ent, v.fkinfo.FromTable, v.fkinfo.RefTable, v.fkinfo.FromField, v.fkinfo.RefField)
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

// buildTableSchema builds a CREATE TABLE schema for the Postgres DB, and
// returns it to the caller, along with the components determined from
// the db and sqac struct-tags.  this method is used in CreateTables
// and AlterTables methods.
func (pf *PostgresFlavor) buildTablSchema(tn string, ent interface{}) TblComponents {

	pKeys := ""
	var sequences []common.SqacPair
	indexes := make(map[string]IndexInfo)
	fKeys := make([]FKeyInfo, 0)
	tableSchema := fmt.Sprintf("CREATE TABLE %s (", tn)

	// get a list of the field names, go-types and db attributes.
	// TagReader is a common function across db-flavors. For
	// this reason, the db-specific-data-type for each field
	// is determined locally.
	fldef, err := common.TagReader(ent, nil)
	if err != nil {
		panic(err)
	}

	// set the Postgres field-types and build the table schema,
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

		// if the field has been marked as NoDB, continue with the next field
		if fd.NoDB == true {
			continue
		}

		switch fd.UnderGoType { // fd.GoType {
		case "uint", "uint8", "uint16", "uint32", "uint64",
			"int", "int8", "int16", "int32", "int64", "rune", "byte":

			if strings.Contains(fd.UnderGoType, "64") {
				col.fType = "bigint"
			} else {
				col.fType = "integer"
			}

			// read sqac tag pairs and apply
			seqName := ""
			for _, p := range fd.SqacPairs {

				switch p.Name {
				case "primary_key":

					col.fPrimaryKey = "PRIMARY KEY"
					pKeys = pKeys + fd.FName + ","

					if p.Value == "inc" {
						if strings.Contains(fd.UnderGoType, "64") {
							col.fType = "bigserial"
						} else {
							col.fType = "serial"
						}
					}

				case "start":
					start, err := strconv.Atoi(p.Value)
					if err != nil {
						panic(err)
					}
					if seqName == "" && start > 0 {
						seqName = tn + "_" + fd.FName + "_seq"
						sequences = append(sequences, common.SqacPair{Name: seqName, Value: p.Value})
					}

				case "default":
					col.fDefault = fmt.Sprintf("DEFAULT %s", p.Value)

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
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_", false, true)

					case "unique":
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_", true, true)

					default:
						indexes = pf.processIndexTag(indexes, tn, fd.FName, p.Value, false, false)
					}

				case "fkey":
					fKeys = pf.processFKeyTag(fKeys, tn, fd.FName, p.Value)

				default:

				}
			}
			fldef[idx].FType = col.fType

		case "string":
			col.fType = "text"

			for _, p := range fd.SqacPairs {
				switch p.Name {
				case "primary_key":
					col.fPrimaryKey = "PRIMARY KEY"
					pKeys = pKeys + fd.FName + ","

				case "nullable":
					if p.Value == "false" {
						col.fNullable = "NOT NULL"
					}

				case "default":
					col.fDefault = fmt.Sprintf("DEFAULT '%s'", p.Value)

				case "constraint":
					if p.Value == "unique" {
						col.fUniqueConstraint = "UNIQUE"
					}

				case "index":

					switch p.Value {
					case "non-unique":
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_", false, true)

					case "unique":
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_", true, true)

					default:
						indexes = pf.processIndexTag(indexes, tn, fd.FName, p.Value, false, false)
					}

				case "fkey":
					fKeys = pf.processFKeyTag(fKeys, tn, fd.FName, p.Value)

				default:

				}
			}
			fldef[idx].FType = col.fType

		case "float32", "float64":
			col.fType = "numeric"

			for _, p := range fd.SqacPairs {
				switch p.Name {
				case "primary_key":
					col.fPrimaryKey = "PRIMARY KEY"
					pKeys = pKeys + fd.FName + ","

				case "nullable":
					if p.Value == "false" {
						col.fNullable = "NOT NULL"
					}

				case "default":
					col.fDefault = fmt.Sprintf("DEFAULT '%s'", p.Value)

				case "constraint":
					if p.Value == "unique" {
						col.fUniqueConstraint = "UNIQUE"
					}

				case "index":
					switch p.Value {
					case "non-unique":
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_", false, true)

					case "unique":
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_", true, true)

					default:
						indexes = pf.processIndexTag(indexes, tn, fd.FName, p.Value, false, false)
					}

				case "fkey":
					fKeys = pf.processFKeyTag(fKeys, tn, fd.FName, p.Value)

				default:

				}
			}
			fldef[idx].FType = col.fType

		case "bool":
			col.fType = "boolean"

			for _, p := range fd.SqacPairs {
				switch p.Name {
				case "primary_key":
					pKeys = pKeys + fd.FName + ","

				case "default":
					col.fDefault = fmt.Sprintf("DEFAULT %s", p.Value)

				case "nullable":
					if p.Value == "false" {
						col.fNullable = "NOT NULL"
					}

				case "index":
					switch p.Value {
					case "non-unique":
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_", false, true)

					case "unique":
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_", true, true)

					default:
						indexes = pf.processIndexTag(indexes, tn, fd.FName, p.Value, false, false)
					}

				case "fkey":
					fKeys = pf.processFKeyTag(fKeys, tn, fd.FName, p.Value)

				default:

				}
			}
			fldef[idx].FType = col.fType

		case "time.Time":
			col.fType = "timestamp with time zone"

			for _, p := range fd.SqacPairs {
				switch p.Name {
				case "primary_key":
					col.fPrimaryKey = "PRIMARY KEY"
					pKeys = pKeys + fd.FName + ","

				case "default":
					if p.Value != "eot" {
						col.fDefault = fmt.Sprintf("DEFAULT %s", p.Value)
					} else {
						col.fDefault = fmt.Sprintf("DEFAULT %s", "make_timestamptz(9999, 12, 31, 23, 59, 59.9)")
					}

				case "index":
					switch p.Value {
					case "non-unique":
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_", false, true)

					case "unique":
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_", true, true)

					default:
						indexes = pf.processIndexTag(indexes, tn, fd.FName, p.Value, false, false)
					}

				case "fkey":
					fKeys = pf.processFKeyTag(fKeys, tn, fd.FName, p.Value)

				default:

				}
			}
			fldef[idx].FType = col.fType

		// this is always nullable, and consequently the following are
		// not supported default value, use as a primary key, use as an index.
		case "*time.Time":
			col.fType = "timestamp with time zone"
			for _, p := range fd.SqacPairs {
				switch p.Name {
				case "default":
					if p.Value != "eot" {
						col.fDefault = fmt.Sprintf("DEFAULT %s", p.Value)
					} else {
						col.fDefault = fmt.Sprintf("DEFAULT %s", "make_timestamptz(9999, 12, 31, 23, 59, 59.9)")
					}

				case "fkey":
					fKeys = pf.processFKeyTag(fKeys, tn, fd.FName, p.Value)

				default:
					// do nothing with other tag directives
				}
			}
		default:
			err := fmt.Errorf("go type %s is not presently supported", fldef[idx].FType)
			panic(err)
		}

		// add the current column to the schema
		tableSchema = tableSchema + fmt.Sprintf("%s %s", col.fName, col.fType)
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
		tableSchema = tableSchema + ");"
	}
	if tableSchema != "" && pKeys != "" {
		pKeys = strings.TrimSuffix(pKeys, ",")
		tableSchema = tableSchema + fmt.Sprintf("CONSTRAINT %s_pkey PRIMARY KEY (%s) );", strings.ToLower(tn), pKeys)
	}

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

	if pf.log {
		rc.Log()
	}
	return rc
}

// CreateTables creates tables on the postgres database referenced
// by pf.DB.
func (pf *PostgresFlavor) CreateTables(i ...interface{}) error {

	// call createTables specifying that the call has not originated
	// from within the AlterTables(...) method.
	_, err := pf.createTables(false, i)
	if err != nil {
		return err
	}
	return nil
}

// DropTables drops tables on the postgres database referenced
// by pf.DB.
func (pf *PostgresFlavor) DropTables(i ...interface{}) error {

	dropSchema := ""

	for t := range i {

		// determine the table name
		tn := common.GetTableName(i[t])
		if tn == "" {
			return fmt.Errorf("unable to determine table name in pf.DropTables")
		}

		// if the table is found to exist, add a DROP statement
		// to the dropSchema string and move on to the next
		// table in the list.
		if pf.ExistsTable(tn) {
			if pf.log {
				fmt.Printf("table %s exists - adding to drop schema...\n", tn)
			}
			dropSchema = dropSchema + fmt.Sprintf("DROP TABLE IF EXISTS %s;", tn)
		}
	}
	if dropSchema != "" {
		pf.ProcessSchema(dropSchema)
	}
	return nil
}

// AlterTables alters tables on the Postgres database referenced
// by pf.DB.
func (pf *PostgresFlavor) AlterTables(i ...interface{}) error {

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
		// if the table does exist, add the Model{} defintion to  the
		// AlterTables buffer (ai).
		if !pf.ExistsTable(tn) {
			ci = append(ci, i[t])
		} else {
			ai = append(ai, i[t])
		}
	}

	// if create-tables buffer 'ci' contains any entries, call createTables and
	// take note of any returned foreign-key definitions.
	if len(ci) > 0 {
		fkBuffer, err = pf.createTables(true, ci)
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
			return fmt.Errorf("unable to determine table name in pf.AlterTables")
		}

		// build the alter-table schema and get its components
		tc := pf.buildTablSchema(tn, ai[t])

		// go through the latest version of the model and check each
		// field against its definition in the database.
		alterSchema := fmt.Sprintf("ALTER TABLE IF EXISTS %s", tn)
		var cols []string

		for _, fd := range tc.flDef {
			// new columns first
			if !pf.ExistsColumn(tn, fd.FName) && fd.NoDB == false {

				colSchema := fmt.Sprintf("ADD COLUMN %s %s", fd.FName, fd.FType)
				for _, p := range fd.SqacPairs {
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
			pf.ProcessSchema(alterSchema)
		}

		// add indexes if required
		for k, v := range tc.ind {
			if !pf.ExistsIndex(v.TableName, k) {
				pf.CreateIndex(k, v)
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
		fkExists, _ := pf.ExistsForeignKeyByName(v.ent, fkn)
		if !fkExists {
			err = pf.CreateForeignKey(v.ent, v.fkinfo.FromTable, v.fkinfo.RefTable, v.fkinfo.FromField, v.fkinfo.RefField)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
	}
	return nil
}

// DestructiveResetTables drops tables on the db if they exist,
// as well as any related objects such as sequences.  this is
// useful if you wish to regenerated your table and the
// number-range used by an auto-incementing primary key.
func (pf *PostgresFlavor) DestructiveResetTables(i ...interface{}) error {

	err := pf.DropTables(i...)
	if err != nil {
		return err
	}
	err = pf.CreateTables(i...)
	if err != nil {
		return err
	}
	return nil
}

// ExistsTable checks the public schema of the connected Postgres
// DB for the existance of the provided table name.  Note that
// the use of to_regclass(<obj_name>) checks for the existance of
// *any* object in the public schema that has that name.  If obj/name
// consistency is maintained, this approach is fine.
func (pf *PostgresFlavor) ExistsTable(tn string) bool {

	reqQuery := fmt.Sprintf("SELECT to_regclass('public.%s');", tn)
	pf.QsLog(reqQuery)
	rows, err := pf.db.Query(reqQuery)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var s string
	for rows.Next() {
		err = rows.Scan(&s)
		if err != nil {
			return false
		}
		return true
	}
	return false
}

// ExistsColumn checks for the existance of the specified
// table-column checking for the column name. this is
// rather incomplete, but in many cases where there is
// a type-discrepancy it is neccessary to drop and recreate
// the column - or the entire table if a key is involved.
// consider also that pg requies autoincrement fields to
// be specified as 'serial' or 'bigserial', but then goes
// on to report them as 'integer' in the actual db-scema.
func (pf *PostgresFlavor) ExistsColumn(tn string, cn string) bool {

	n := 0
	pf.QsLog("SELECT count(*) FROM INFORMATION_SCHEMA.columns WHERE table_name = ? AND column_name = ? AND table_schema = CURRENT_SCHEMA()", tn, cn)
	row := pf.db.QueryRow("SELECT count(*) FROM INFORMATION_SCHEMA.columns WHERE table_name = $1 AND column_name = $2 AND table_schema = CURRENT_SCHEMA()", tn, cn)
	if row != nil {
		row.Scan(&n)
		if n > 0 {
			return true
		}
	}
	return false
}

// ExistsIndex checks the connected Postgres database for the presence
// of the specified index - assuming that the index-type has not
// been adjusted...
func (pf *PostgresFlavor) ExistsIndex(tn string, in string) bool {

	n := 0
	pf.QsLog("SELECT count(*) FROM pg_indexes WHERE tablename = ? AND indexname = ? AND schemaname = CURRENT_SCHEMA()", tn, in)
	err := pf.db.QueryRow("SELECT count(*) FROM pg_indexes WHERE tablename = $1 AND indexname = $2 AND schemaname = CURRENT_SCHEMA()", tn, in).Scan(&n)
	if err != nil {
		return false
	}
	if n > 0 {
		return true
	}
	return false
}

// DropIndex drops the specfied index on the connected Postgres database.
// tn is ignored for Postgres.
func (pf *PostgresFlavor) DropIndex(tn string, in string) error {

	indexSchema := fmt.Sprintf("DROP INDEX IF EXISTS %s;", in)
	pf.ProcessSchema(indexSchema)
	return nil
}

// ExistsSequence checks the public schema of the connected Postgres
// DB for the existance of the provided sequence name.
func (pf *PostgresFlavor) ExistsSequence(sn string) bool {

	var params []interface{}
	reqQuery := fmt.Sprintf("SELECT relname FROM pg_class WHERE relkind = 'S' AND relname::name = $1")
	params = append(params, sn)
	pf.QsLog(reqQuery, params...)

	rows, err := pf.db.Query(reqQuery, params...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var s string
	for rows.Next() {
		err = rows.Scan(&s)
		if err != nil {
			return false
		}
		return true
	}
	return false
}

// CreateSequence creates the required sequence on the connected Postgres
// database in the public schema.  Panics on error.
func (pf *PostgresFlavor) CreateSequence(sn string, start int) {

	seqSchema := fmt.Sprintf("CREATE SEQUENCE %s START %d;", sn, start)
	pf.ProcessSchema(seqSchema)
}

// AlterSequenceStart adjusts the starting value of the named sequence.  This should
// be called very carefully, preferably only at the time that the table/sequence is
// created on the db.  There are no safeguards here.
func (pf *PostgresFlavor) AlterSequenceStart(sn string, start int) error {

	seqSchema := fmt.Sprintf("ALTER SEQUENCE IF EXISTS %s RESTART WITH %d;", sn, start)
	pf.ProcessSchema(seqSchema)
	return nil
}

// GetNextSequenceValue is used primarily for testing.  It returns
// the current value of the sequence assigned to the primary-key of the
// of the named Postgres table.  Although it is possible to assign
// Postgres sequences to non-primary-key fields (composite key gen),
// sqac handle auto-increment as a primary-key constraint only.
func (pf *PostgresFlavor) GetNextSequenceValue(name string) (int, error) {

	// determine the column name of the primary key
	pKeyQuery := fmt.Sprintf("SELECT c.column_name, c.ordinal_position FROM information_schema.key_column_usage AS c LEFT JOIN information_schema.table_constraints AS t ON t.constraint_name = c.constraint_name WHERE t.table_name = '%s' AND t.constraint_type = 'PRIMARY KEY';", name)
	var keyColumn string
	var keyColumnPos int
	pf.QsLog(pKeyQuery)
	pf.db.QueryRow(pKeyQuery).Scan(&keyColumn, &keyColumnPos)
	if keyColumn == "" {
		return 0, fmt.Errorf("could not identify primary-key column for table %s", name)
	}

	// Postgres sequences have format '<tablename>_<keyColumn>_seq'
	seqName := fmt.Sprintf("%s_%s_seq", name, keyColumn)

	if pf.ExistsSequence(seqName) {
		seq := 0
		seqQuery := fmt.Sprintf("SELECT nextval('%s');", seqName)
		pf.QsLog(seqQuery)

		err := pf.db.QueryRow(seqQuery).Scan(&seq)
		if err != nil {
			return 0, err
		}
		return seq, nil
	}
	return 0, nil
}

// ExistsForeignKeyByName checks to see if the named foreign-key exists on the
// table corresponding to provided sqac model (i).
func (pf *PostgresFlavor) ExistsForeignKeyByName(i interface{}, fkn string) (bool, error) {

	var count uint64
	tn := common.GetTableName(i)

	fkQuery := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.table_constraints WHERE constraint_name='%s' AND table_name='%s';", fkn, tn)
	pf.QsLog(fkQuery)

	err := pf.Get(&count, fkQuery)
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
func (pf *PostgresFlavor) ExistsForeignKeyByFields(i interface{}, ft, rt, ff, rf string) (bool, error) {

	fkn, err := common.GetFKeyName(i, ft, rt, ff, rf)
	if err != nil {
		return false, err
	}

	return pf.ExistsForeignKeyByName(i, fkn)
}

//================================================================
// CRUD ops
//================================================================

// Create the entity (single-row) on the database
func (pf *PostgresFlavor) Create(ent interface{}) error {

	var info CrudInfo
	info.ent = ent
	info.log = false
	info.mode = "C"

	err := pf.BuildComponents(&info)
	if err != nil {
		return err
	}

	// build the postgres insert query
	insQuery := fmt.Sprintf("INSERT INTO %s", info.tn)
	insQuery = fmt.Sprintf("%s %s VALUES %s RETURNING *;", insQuery, info.fList, info.vList)
	pf.QsLog(insQuery)

	// clear the source data - deals with non-persistet columns
	e := reflect.ValueOf(info.ent).Elem()
	e.Set(reflect.Zero(e.Type()))

	// attempt the insert and read the result back into info.resultMap
	err = pf.db.QueryRowx(insQuery).StructScan(info.ent) //.MapScan(info.resultMap) // SliceScan
	if err != nil {
		return err
	}
	info.entValue = reflect.ValueOf(info.ent)
	return nil
}

// Update an existing entity (single-row) on the database
func (pf *PostgresFlavor) Update(ent interface{}) error {

	var info CrudInfo
	info.ent = ent
	info.log = false
	info.mode = "U"

	err := pf.BuildComponents(&info)
	if err != nil {
		return err
	}

	keyList := ""
	for k, s := range info.keyMap {

		fType := reflect.TypeOf(s).String()
		if pf.IsLog() {
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
	keyList = keyList + " RETURNING *;"

	updQuery := fmt.Sprintf("UPDATE %s SET", info.tn)
	updQuery = fmt.Sprintf("%s %s = %s WHERE%s", updQuery, info.fList, info.vList, keyList)
	pf.QsLog(updQuery)

	// clear the source data - deals with non-persistet columns
	e := reflect.ValueOf(info.ent).Elem()
	e.Set(reflect.Zero(e.Type()))

	// attempt the update and read result back into resultMap
	err = pf.db.QueryRowx(updQuery).StructScan(info.ent) //.MapScan(info.resultMap) // SliceScan
	if err != nil {
		return err
	}
	info.entValue = reflect.ValueOf(info.ent)
	return nil
}

// // GetEntitiesWithCommands is the experimental replacement for all get-set ops
//func (pf *PostgresFlavor) GetEntitiesWithCommands(ents interface{}, params []common.GetParam, cmdMap map[string]interface{}) (interface{}, error) {

// 	fmt.Println()
// 	fmt.Println("GetEntitiesWithCommands received params:", params)
// 	fmt.Println("GetEntitiesWithCommands received cmdMap:", cmdMap)
// 	fmt.Println()

// 	var err error
// 	var count uint64
// 	var row *sqlx.Row
// 	paramString := ""
// 	selQuery := ""

// 	// get the underlying data type of the interface{}
// 	entTypeElem := reflect.TypeOf(ents).Elem()
// 	// fmt.Println("entTypeElem:", entTypeElem)

// 	// create a struct from the type
// 	testVar := reflect.New(entTypeElem)

// 	// determine the db table name
// 	tn := common.GetTableName(ents)

// 	// are there any parameters to include in the query?
// 	var pv []interface{}
// 	if params != nil && len(params) > 0 {
// 		paramString = " WHERE"
// 		for i := range params {
// 			paramString = fmt.Sprintf("%s %s %s ? %s", paramString, common.CamelToSnake(params[i].FieldName), params[i].Operand, params[i].NextOperator)
// 			pv = append(pv, params[i].ParamValue)
// 		}
// 	}
// 	fmt.Println("constructed paramString:", paramString)

// 	// received a $count command?  this supercedes all, as it should not
// 	// be mixed with any other $<commands>.
// 	_, ok := cmdMap["count"]
// 	if ok {
// 		if paramString == "" {
// 			selQuery = fmt.Sprintf("SELECT COUNT(*) FROM %s;", tn)
// 			pf.QsLog(selQuery)
// 			row = pf.ExecuteQueryRowx(selQuery)
// 		} else {
// 			selQuery = fmt.Sprintf("SELECT COUNT(*) FROM %s%s;", tn, paramString)
// 			pf.QsLog(selQuery)
// 			row = pf.ExecuteQueryRowx(selQuery, pv...)
// 		}

// 		err = row.Scan(&count)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		return count, nil
// 	}

// 	// no $count command - build query
// 	var obString string
// 	var limitString string
// 	var offsetString string
// 	var adString string

// 	// received $orderby command?
// 	obField, ok := cmdMap["orderby"]
// 	if ok {
// 		obString = fmt.Sprintf(" ORDER BY %s", obField.(string))
// 	}

// 	// received $asc command?
// 	_, ok = cmdMap["asc"]
// 	if ok {
// 		adString = " ASC"
// 	}

// 	// received $desc command?
// 	_, ok = cmdMap["desc"]
// 	if ok {
// 		adString = " DESC"
// 	}

// 	// received $limit command?
// 	limField, ok := cmdMap["limit"]
// 	if ok {
// 		limitString = fmt.Sprintf(" LIMIT %v", limField)
// 	}

// 	// received $offset command?
// 	offField, ok := cmdMap["offset"]
// 	if ok {
// 		offsetString = fmt.Sprintf(" OFFSET %v", offField)
// 	}

// 	// -- SELECT COUNT(*) FROM library;
// 	// -- SELECT * FROM library;
// 	// -- SELECT * FROM library LIMIT 2;
// 	// -- SELECT * FROM library OFFSET 2;
// 	// -- SELECT * FROM library LIMIT 2 OFFSET 1;
// 	// -- SELECT * FROM library ORDER BY ID DESC;
// 	// -- SELECT * FROM library ORDER BY ID ASC;
// 	// -- SELECT * FROM library ORDER BY name ASC;
// 	// -- SELECT * FROM library ORDER BY ID ASC LIMIT 2 OFFSET 2;

// 	// if $asc or $desc were specifed with no $orderby, default to order by id
// 	if obString == "" && adString != "" {
// 		obString = " ORDER BY id"
// 	}

// 	selQuery = fmt.Sprintf("SELECT * FROM %s%s", tn, paramString)
// 	selQuery = pf.db.Rebind(selQuery)
// 	fmt.Println("rebound selQuery:", selQuery)

// 	selQuery = fmt.Sprintf("%s%s%s%s%s;", selQuery, obString, adString, limitString, offsetString)
// 	fmt.Println("selQuery fully constructed:", selQuery)
// 	pf.QsLog(selQuery)

// 	// read the rows
// 	fmt.Println("pv...", pv)
// 	rows, err := pf.db.Queryx(selQuery, pv...)
// 	if err != nil {
// 		log.Printf("GetEntities for table &s returned error: %v\n", err.Error())
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	// iterate over the rows collection and put the results
// 	// into the ents interface (slice)
// 	entsv := reflect.ValueOf(ents)
// 	for rows.Next() {
// 		err = rows.StructScan(testVar.Interface())
// 		if err != nil {
// 			fmt.Println("scan error:", err)
// 			return nil, err
// 		}
// 		// fmt.Println(testVar)
// 		entsv = reflect.Append(entsv, testVar.Elem())
// 	}

// 	ents = entsv.Interface()
// 	// fmt.Println("ents:", ents)
// 	return entsv.Interface(), nil
// }

package sqac

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
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

// CreateTables creates tables on the postgres database referenced
// by pf.DB.
func (pf *PostgresFlavor) CreateTables(i ...interface{}) error {

	for t, ent := range i {

		ftr := reflect.TypeOf(ent)
		if pf.log {
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
			return fmt.Errorf("unable to determine table name in pf.CreateTables")
		}

		// if the table is found to exist, skip the creation
		// and move on to the next table in the list.
		if pf.ExistsTable(tn) {
			if pf.log {
				fmt.Printf("CreateTable - table %s exists - skipping...\n", tn)
			}
			continue
		}

		// build the create table schema and return all of the table info
		tc := pf.buildTablSchema(tn, i[t])

		// create the table on the db
		pf.db.MustExec(tc.tblSchema)
		for _, sq := range tc.seq {
			start, _ := strconv.Atoi(sq.Value)
			pf.AlterSequenceStart(sq.Name, start)
		}
		for k, in := range tc.ind {
			pf.CreateIndex(k, in)
		}
	}
	return nil
}

// buildTableSchema builds a CREATE TABLE schema for the Postgres DB, and
// returns it to the caller, along with the components determined from
// the db and rgen struct-tags.  this method is used in CreateTables
// and AlterTables methods.
func (pf *PostgresFlavor) buildTablSchema(tn string, ent interface{}) TblComponents {

	pKeys := ""
	var sequences []RgenPair
	indexes := make(map[string]IndexInfo)
	tableSchema := fmt.Sprintf("CREATE TABLE %s (", tn)

	// get a list of the field names, go-types and db attributes.
	// TagReader is a common function across db-flavors. For
	// this reason, the db-specific-data-type for each field
	// is determined locally.
	fldef, err := TagReader(ent, nil)
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

		switch fd.GoType {
		case "uint", "uint8", "uint16", "uint32", "uint64",
			"int", "int8", "int16", "int32", "int64", "rune", "byte":

			if strings.Contains(fd.GoType, "64") {
				col.fType = "bigint"
			} else {
				col.fType = "integer"
			}

			// read rgen tag pairs and apply
			seqName := ""
			for _, p := range fd.RgenPairs {

				switch p.Name {
				case "primary_key":

					col.fPrimaryKey = "PRIMARY KEY"
					pKeys = pKeys + fd.FName + ","

					if p.Value == "inc" {
						if strings.Contains(fd.GoType, "64") {
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
						sequences = append(sequences, RgenPair{Name: seqName, Value: p.Value})
					}

				case "default":
					col.fDefault = fmt.Sprintf("DEFAULT %s", p.Value)

				case "nullable":
					if p.Value == "false" {
						col.fNullable = "NOT NULL"
					}

				case "index":
					switch p.Value {
					case "non-unique":
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_"+fd.FName, false, true)

					case "unique":
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_"+fd.FName, true, true)

					default:
						indexes = pf.processIndexTag(indexes, tn, fd.FName, p.Value, false, false)
					}

				default:

				}
			}
			fldef[idx].FType = col.fType

		case "string":
			col.fType = "text"

			for _, p := range fd.RgenPairs {
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

				case "index":

					switch p.Value {
					case "non-unique":
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_"+fd.FName, false, true)

					case "unique":
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_"+fd.FName, true, true)

					default:
						indexes = pf.processIndexTag(indexes, tn, fd.FName, p.Value, false, false)
					}

				default:

				}
			}
			fldef[idx].FType = col.fType

		case "float32", "float64":
			col.fType = "numeric"

			for _, p := range fd.RgenPairs {
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

				case "index":
					switch p.Value {
					case "non-unique":
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_"+fd.FName, false, true)

					case "unique":
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_"+fd.FName, true, true)

					default:
						indexes = pf.processIndexTag(indexes, tn, fd.FName, p.Value, false, false)
					}

				default:

				}
			}
			fldef[idx].FType = col.fType

		case "bool":
			col.fType = "boolean"

			for _, p := range fd.RgenPairs {
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
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_"+fd.FName, false, true)

					case "unique":
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_"+fd.FName, true, true)

					default:
						indexes = pf.processIndexTag(indexes, tn, fd.FName, p.Value, false, false)
					}

				default:

				}
			}
			fldef[idx].FType = col.fType

		case "time.Time":
			col.fType = "timestamp with time zone"

			for _, p := range fd.RgenPairs {
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
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_"+fd.FName, false, true)

					case "unique":
						indexes = pf.processIndexTag(indexes, tn, fd.FName, "idx_"+fd.FName, true, true)

					default:
						indexes = pf.processIndexTag(indexes, tn, fd.FName, p.Value, false, false)
					}

				default:

				}
			}
			fldef[idx].FType = col.fType

		// this is always nullable, and consequently the following are
		// not supported default value, use as a primary key, use as an index.
		case "*time.Time":
			col.fType = "timestamp with time zone"
			for _, p := range fd.RgenPairs {
				switch p.Name {
				case "default":
					if p.Value != "eot" {
						col.fDefault = fmt.Sprintf("DEFAULT %s", p.Value)
					} else {
						col.fDefault = fmt.Sprintf("DEFAULT %s", "make_timestamptz(9999, 12, 31, 23, 59, 59.9)")
					}
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
		tableSchema = tableSchema + ", "
	}

	if tableSchema != "" && pKeys == "" {
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
		pk:        pKeys,
		err:       err,
	}

	if pf.log {
		rc.Log()
	}
	return rc
}

// DropTables drops tables on the postgres database referenced
// by pf.DB.
func (pf *PostgresFlavor) DropTables(i ...interface{}) error {

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
			return fmt.Errorf("unable to determine table name in pf.AlterTables")
		}

		// if the table does not exist, call CreateTables
		// if the table does exist, examine it and perform
		// alterations if neccessary
		if !pf.ExistsTable(tn) {
			pf.CreateTables(ent)
			continue
		}

		// build the altered table schema and get its components
		tc := pf.buildTablSchema(tn, i[t])

		// go through the latest version of the model and check each
		// field against its definition in the database.
		alterSchema := fmt.Sprintf("ALTER TABLE IF EXISTS %s", tn)
		var cols []string

		for _, fd := range tc.flDef {
			// new columns first
			if !pf.ExistsColumn(tn, fd.FName) && fd.NoDB == false {

				colSchema := fmt.Sprintf("ADD COLUMN %s %s", fd.FName, fd.FType)
				for _, p := range fd.RgenPairs {
					switch p.Name {
					case "primary_key":
						// abort - adding primary key
						panic(fmt.Errorf("aborting - cannot add a primary-key (table-field %s-%s) through migration", tn, fd.FName))

					case "default":
						colSchema = fmt.Sprintf("%s DEFAULT %s", colSchema, p.Value)

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
	fetch, err := pf.db.Query(reqQuery)
	if err != nil {
		panic(err)
	}

	var s string
	for fetch.Next() {
		err = fetch.Scan(&s)
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
	row := pf.db.QueryRow("SELECT count(*) FROM pg_indexes WHERE tablename = $1 AND indexname = $2 AND schemaname = CURRENT_SCHEMA()", tn, in)
	if row != nil {
		row.Scan(&n)
		if n > 0 {
			return true
		}
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
	fetch, err := pf.db.Query(reqQuery, params...)
	if err != nil {
		panic(err)
	}

	var s string
	for fetch.Next() {
		err = fetch.Scan(&s)
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
	pf.db.QueryRow(pKeyQuery).Scan(&keyColumn, &keyColumnPos)
	if keyColumn == "" {
		return 0, fmt.Errorf("could not identify primary-key column for table %s", name)
	}

	// Postgres sequences have format '<tablename>_<keyColumn>_seq'
	seqName := fmt.Sprintf("%s_%s_seq", name, keyColumn)

	if pf.ExistsSequence(seqName) {
		seq := 0
		seqQuery := fmt.Sprintf("SELECT nextval('%s');", seqName)
		err := pf.db.QueryRow(seqQuery).Scan(&seq)
		if err != nil {
			return 0, err
		}
		return seq, nil
	}
	return 0, nil
}

//================================================================
// CRUD ops :(
//================================================================

// Create - Create the entity (single-row) on the database
func (pf *PostgresFlavor) Create(ent interface{}) error {

	// http://speakmy.name/2014/09/14/modifying-interfaced-go-struct/
	// get the underlying Type of the interface ptr
	stype := reflect.TypeOf(ent).Elem()
	if pf.IsLog() {
		fmt.Println("stype:", stype)
	}

	// check that the interface type passed in was a struct
	if stype.Kind() != reflect.Struct {
		return fmt.Errorf("only struct{} types can be passed in for table creation.  got %s", stype.Kind())
	}

	// read the tags for the struct underlying the interface ptr
	flDef, err := TagReader(ent, stype)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if pf.IsLog() {
		fmt.Println("flDef:", flDef)
	}

	// determine the table name as per the table creation logic
	tn := reflect.TypeOf(ent).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

	insQuery := fmt.Sprintf("INSERT INTO %s", tn)
	fList := "("
	vList := "("

	// get the value that the interface ptr ent points to
	// i.e. the struct holding the data for insertion
	v := reflect.ValueOf(ent).Elem()
	if pf.IsLog() {
		fmt.Println("value of data in struct for insertion:", v)
	}

	// what to do with rgen tags
	// primary key:inc - do not fill
	// primary key:""  - do nothing
	// default - DEFAULT keyword for field
	// nullable - if no and nil value, fill with default value for nullable type
	// insQuery = "INSERT INTO depot (depot_num, region, province) VALUES (DEFAULT,'YVR','AB');"
	// https: //stackoverflow.com/questions/18926303/iterate-through-the-fields-of-a-struct-in-go
	// entity-type in Create CRUD call: sqac_test.Depot
	// {depot_num  int false [{primary_key inc} {start 90000000}]}
	// {depot_bay  int false [{primary_key }]}
	// {create_date  time.Time false [{nullable false} {default now()} {index unique}]}
	// {region  string false [{nullable false} {default YYC}]}
	// {province  string false [{nullable false} {default AB}]}
	// {country  string false [{nullable true} {default CA}]}
	// {new_column1  string false [{nullable false}]}
	// {new_column2  int64 false [{nullable false}]}
	// {new_column3  float64 false [{nullable false} {default 0.0}]}
	// {non_persistent_column  string true []}

	// iterate over the entity-struct metadata
	for i, fd := range flDef {
		if pf.IsLog() {
			fmt.Println(fd)
		}
		if fd.NoDB == true {
			continue
		}
		bDefault := false
		bPkeyInc := false
		bNullable := false

		// set the field attribute indicators
		for _, t := range fd.RgenPairs {
			switch t.Name {
			case "primary_key":
				if t.Value == "inc" {
					bPkeyInc = true
				}
			case "default":
				bDefault = true
			case "nullable":
				if t.Value == "true" || t.Value == "TRUE" {
					bNullable = true
				}
			default:

			}
		}

		// get the value of the current entity field
		fv := v.Field(i).Interface()
		fvr := v.Field(i)
		switch fd.GoType {
		case "int", "uint", "int8", "uint8", "int16", "uint16", "int32", "uint32", "int64", "uint64", "rune", "byte":
			if bDefault == true && fvr.IsNil() {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}
			if bPkeyInc == true {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}
			if bNullable == false && fv == nil {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%d, ", vList, 0)
				continue
			}
			// in all other cases, just use the given value making the
			// assumption that the int-type field contains an int-type
			fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
			vList = fmt.Sprintf("%s%d, ", vList, fvr.Int())
			continue

		case "float32", "float64":
			if bDefault == true && fv == nil {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}
			if bPkeyInc == true {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}
			if bNullable == false && fv == nil {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%f, ", vList, 0.0)
				continue
			}
			// in all other cases, just use the given value making the
			// assumption that the float-type field contains a float-type
			fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
			vList = fmt.Sprintf("%s%f, ", vList, fvr.Float())
			continue

		case "string":
			if bDefault == true && fv == nil {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}
			if bPkeyInc == true {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}
			if bNullable == false && fv == nil {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "''")
				continue
			}
			// in all other cases, just use the given value making the
			// assumption that the string-type field contains a string-type
			fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
			vList = fmt.Sprintf("%s'%s', ", vList, fv)
			continue

		case "time.Time", "*time.Time":
			if bDefault == true {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}
			if bPkeyInc == true {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}
			if bNullable == false && fv == nil {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "make_timestamptz(0000, 00, 00, 00, 00, 00.0")
				continue
			}
			fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
			vList = fmt.Sprintf("%s%v, ", vList, fv)
			continue

		default:

		}
	}

	// build the insert query string
	fList = strings.TrimSuffix(fList, ", ")
	fList = fmt.Sprintf("%s%s", fList, ")")
	vList = strings.TrimSuffix(vList, ", ")
	vList = fmt.Sprintf("%s%s", vList, ") RETURNING *;") // depot_num
	insQuery = fmt.Sprintf("%s %s VALUES %s", insQuery, fList, vList)
	if pf.IsLog() {
		fmt.Println(insQuery)
	}

	// attempt the insert and read result back into resultMap
	resultMap := make(map[string]interface{})
	err = pf.db.QueryRowx(insQuery).MapScan(resultMap) // SliceScan
	if err != nil {
		fmt.Println(err) //?
	}

	if pf.IsLog() {
		for k, r := range resultMap {
			fmt.Println(k, r)
		}
		fmt.Println("TYPEOF ent:", reflect.TypeOf(ent)) // sqac_test.Depot
	}

	values := make([]interface{}, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i).Interface()

		np, _ := stype.Field(i).Tag.Lookup("rgen")
		if np == "-" {
			continue
		}

		fn := stype.Field(i).Name                // GoName
		st := stype.Field(i).Tag                 // structTag
		ft, _ := stype.Field(i).Tag.Lookup("db") // snake_name
		tp := stype.Field(i).Type.String()       // field-type as String

		if pf.IsLog() {
			fmt.Println("NAME:", fn)
			fmt.Println("TAG:", st)
			fmt.Println("DB FIELD NAME:", ft)
			fmt.Println("FIELD-TYPE:", tp)
		}

		// get the reflect.Value of the current field in the ent struct
		fv := reflect.ValueOf(ent).Elem().FieldByName(fn)
		if !fv.IsValid() {
			panic(fmt.Errorf("invalid field %s in struct %s", fn, st))
		}

		// check if the reflect.Value can be updated and set the returned
		// db field value from the resultMap.
		if fv.CanSet() {
			switch tp {

			case "int", "int8", "int16", "int32", "int64":
				fv.SetInt(resultMap[ft].(int64))

			case "uint", "uint8", "uint16", "uint32", "uint64", "rune", "byte":
				fv.SetUint(resultMap[ft].(uint64))

			case "float32", "float64":
				s := fmt.Sprintf("%s", resultMap[ft].([]byte))
				f, err := strconv.ParseFloat(s, 64)
				if err != nil {
					fmt.Printf("%s", err)
				}
				if pf.IsLog() {
					fmt.Println("float value:", f)
				}
				fv.SetFloat(f)

			case "string":
				fv.SetString(resultMap[ft].(string))

			case "time.Time":
				fv.Set(reflect.ValueOf(resultMap[ft].(time.Time)))

			case "*time.Time":
				fv.Set(reflect.ValueOf(resultMap[ft].(*time.Time)))

			default:
				fmt.Printf("UNSUPPORTED TYPE:%s\n", tp)
				// try
				// fv.Set(reflect.ValueOf(resultMap[ft].(stype.Field(i).Type)))

			}
		} else {
			fmt.Printf("CANNOT SET %s:\n", fn)
		}

	}
	if pf.IsLog() {
		fmt.Println(values)
		fmt.Println("ENT:", ent)
	}
	return nil
}

// Update - Update an existing entity (single-row) on the database
func (pf *PostgresFlavor) Update(ent interface{}) error {
	return nil
}

// Delete - Delete an existing entity (single-row) on the database
func (pf *PostgresFlavor) Delete(key interface{}) error { // (id uint) error
	return nil
}

// GetEntity - get an existing entity from the db using the primary
// key definition.
func (pf *PostgresFlavor) GetEntity(key interface{}) interface{} {
	return nil
}

// GetEntities retrieves all entities of the requested type
// from the database.  the returned list is unfiltered and
// not pageable for now.
func (pf *PostgresFlavor) GetEntities() []interface{} {
	return nil
}

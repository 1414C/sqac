package sqac

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/1414C/sqac/common"
	"github.com/jmoiron/sqlx"
)

// IndexInfo contains index definitions as read from the rgen:"index" tags
type IndexInfo struct {
	TableName   string
	Unique      bool
	IndexFields []string
}

// ColComponents is used to capture the field properties from rgen: tags
// during table creation and table alteration activities.
type ColComponents struct {
	fName             string
	fType             string
	fPrimaryKey       string
	fAutoInc          bool // not used for Postgres
	fStart            int  // only used for HDB
	fDefault          string
	fUniqueConstraint string
	fNullable         string
}

// TblComponents is used as a collector structure for internal table
// create / alter processing.
type TblComponents struct {
	tblSchema string
	flDef     []common.FieldDef
	seq       []common.RgenPair
	ind       map[string]IndexInfo
	pk        string
	err       error
}

// Log dumps all of the raw table components to stdout is called for CreateTable
// and AlterTable operations if the main sqac logging has been activated via
// BaseFlavor.Log(true).
func (tc *TblComponents) Log() {
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
		fmt.Printf("FIELD DEF: fname:%s, ftype:%s, gotype:%s ,nodb:%v\n", v.FName, v.FType, v.GoType, v.NoDB)
		for _, p := range v.RgenPairs {
			fmt.Printf("FIELD PROPERTY: %s, %v\n", p.Name, p.Value)
		}
		fmt.Println("------")
	}
	fmt.Println()
	fmt.Println("ERROR:", tc.err)
	fmt.Println("====================================================================")
}

// CTick CBackTick and CDblQuote specify the quote
// style for for db field encapsulation in CREATE
// and ALTER table schemas
const CTick = "'"
const CBackTick = "`"
const CDblQuote = "\""

// PublicDB exposes functions for db schema operations
type PublicDB interface {

	// postgres, sqlite, mariadb, hdb, hana etc.
	GetDBDriverName() string

	// activate / check logging
	Log(b bool)
	IsLog() bool
	DBLog(b bool)
	IsDBLog() bool

	// set the *sqlx.DB handle in the PublicDB interface
	SetDB(db *sqlx.DB)
	GetDB() *sqlx.DB

	// GetDBName reports the name of the currently connected db for
	// information_schema access
	GetDBName() string

	// GetDBQuote reports the quoting preference for db-query construction.
	// ' vs ` vs " for example
	GetDBQuote() string

	// Close the db-connection
	Close() error

	// set / get the max idle sqlx db-connections and max open sqlx db-connections
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)

	GetRelations(tn string) []string

	// i=db/rgen tagged go struct-type
	CreateTables(i ...interface{}) error
	DropTables(i ...interface{}) error
	AlterTables(i ...interface{}) error
	DestructiveResetTables(i ...interface{}) error
	ExistsTable(tn string) bool

	// tn=tableName, cn=columnName
	ExistsColumn(tn string, cn string) bool

	// tn=tableName, in=indexName
	CreateIndex(in string, index IndexInfo) error
	DropIndex(tn string, in string) error
	ExistsIndex(tn string, in string) bool

	// sn=sequenceName, start=start-value, name is used to hold
	// the name of the sequence, autoincrement or identity
	// field name.  the use of name depends on which db system
	// has been connected.
	CreateSequence(sn string, start int)
	AlterSequenceStart(name string, start int) error
	GetNextSequenceValue(name string) (int, error)
	// select pg_get_serial_sequence('public.some_table', 'some_column');
	DropSequence(sn string) error
	ExistsSequence(sn string) bool

	// CreateForeignKey(...) error
	// BuildForeignKeyName(...) error
	// DropForeignKey(...) error
	// ExistsForeignKey(...) bool

	ProcessSchema(schema string)
	ProcessSchemaList(sList []string)
	ProcessTransaction(tList []string) error

	// sql package access
	ExecuteQueryRow(queryString string, qParams ...interface{}) *sql.Row
	ExecuteQuery(queryString string, qParams ...interface{}) (*sql.Rows, error)
	Exec(queryString string, args ...interface{}) (sql.Result, error)

	// sqlx package access
	ExecuteQueryRowx(queryString string, qParams ...interface{}) *sqlx.Row
	ExecuteQueryx(queryString string, qParams ...interface{}) (*sqlx.Rows, error)
	Get(dst interface{}, queryString string, args ...interface{}) error
	Select(dst interface{}, queryString string, args ...interface{}) error

	// Boolean conversions
	BoolToDBBool(b bool) *int
	DBBoolToBool(interface{}) bool
	TimeToFormattedString(i interface{}) string

	// CRUD ops :(
	Create(ent interface{}) error
	Update(ent interface{}) error
	Delete(ent interface{}) error                      // (id uint) error
	GetEntity(ent interface{}) error                   // pass ptr to type containing key information
	GetEntities(ents interface{}) (interface{}, error) // tn == tableName
	GetEntities2(ge GetEnt) error
	GetEntities3(ents interface{})
}

// ensure consistency of interface implementation
var _ PublicDB = &BaseFlavor{}

// BaseFlavor is a supporting struct for interface PublicDB
type BaseFlavor struct {
	db    *sqlx.DB
	log   bool
	dbLog bool
	PublicDB
}

// Log sets the logging status
func (bf *BaseFlavor) Log(b bool) {
	bf.log = b
}

// IsLog reports whether logging is active
func (bf *BaseFlavor) IsLog() bool {
	return bf.log
}

// DBLog sets the db-access-logging status
func (bf *BaseFlavor) DBLog(b bool) {
	bf.dbLog = b
}

// IsDBLog reports whether db-access-logging is active
func (bf *BaseFlavor) IsDBLog() bool {
	return bf.dbLog
}

// QsLog is used to log SQL statements to stdout.  Statements are text approximations
// of what was sent to the database.  For the most part they should be correct, but
// quoting around parameter values is rudimentary.
func (bf *BaseFlavor) QsLog(queryString string, qParams ...interface{}) {
	if !bf.dbLog {
		return
	}

	if qParams != nil {
		for _, v := range qParams {
			switch v.(type) {
			case string:
				rVal := ""
				if bf.GetDBDriverName() == "sqlite3" {
					rVal = "\"" + v.(string) + "\""
				} else {
					rVal = "'" + v.(string) + "'"
				}
				queryString = strings.Replace(queryString, "?", rVal, 1)
			default:
				queryString = strings.Replace(queryString, "?", "%v", 1)
				queryString = fmt.Sprintf(reflect.ValueOf(queryString).String(), v)
				// queryString = strings.Replace(queryString, "?", v.(string), 1)
			}
		}
		fmt.Println(queryString)
		return
	}
	log.Println(queryString)
}

// SetDB sets the sqlx.DB connection in the
// db-flavor environment.
func (bf *BaseFlavor) SetDB(db *sqlx.DB) {
	bf.db = db
}

// GetDB returns a *sqlx.DB pointer if one has
// been set in the db-flavor environment.
func (bf *BaseFlavor) GetDB() *sqlx.DB {
	return bf.db
}

// Close closes the db-connection
func (bf *BaseFlavor) Close() error {
	err := bf.db.Close()
	if err != nil {
		log.Println("failed to close db connection")
		return err
	}
	return nil
}

// GetDBName returns the name of the currently connected db
func (bf *BaseFlavor) GetDBName() (dbName string) {

	row := bf.db.QueryRow("SELECT DATABASE()")
	if row != nil {
		err := row.Scan(&dbName)
		if err != nil {
			panic(err)
		}
	}
	return dbName
}

// GetDBQuote reports the quoting preference for db-query construction.
// this does not refer to quoted strings, but rather to the older(?)
// style of quoting table field-names in query-strings such as:
// SELECT "f1" FROM "t1" WHERE "v1" = <some_criteria>.
// in practice, it seems you can get away without quoting, but
// it is a nod to backward compatibility and it standardizes on
// an approach.
// ' vs ` vs " for example
func (bf *BaseFlavor) GetDBQuote() string {

	switch bf.GetDBDriverName() {
	case "postgres":
		return CTick

	case "mysql":
		return CBackTick

	case "sqlite":
		return CDblQuote

	case "mssql":
		return ""

	case "hdb":
		return ""

	default:
		return CDblQuote
	}
}

// SetMaxIdleConns calls sqlx.SetMaxIdleConns
func (bf *BaseFlavor) SetMaxIdleConns(n int) {
	bf.db.SetMaxIdleConns(n)
}

// SetMaxOpenConns calls sqlx.SetMaxOpenConns
func (bf *BaseFlavor) SetMaxOpenConns(n int) {
	bf.db.SetMaxOpenConns(n)
}

// GetDBDriverName returns the name of the current db-driver
func (bf *BaseFlavor) GetDBDriverName() string {
	return bf.db.DriverName()
}

// BoolToDBBool converts a go-bool value into the
// DB bool representation.  called for DB's that
// do not support a true/false boolean type.
func (bf *BaseFlavor) BoolToDBBool(b bool) *int {

	var r int

	switch b {
	case true:
		r = 1
		return &r

	case false:
		r = 0
		return &r

	default:
		return nil
	}
}

// DBBoolToBool converts from the DB representation
// of a bool into the go-bool type.  The is called for
// DB's that do not support a true/false boolean type.
func (bf *BaseFlavor) DBBoolToBool(i interface{}) bool {

	switch i.(type) {
	case string:
		if i.(string) == "TRUE" || i.(string) == "true" {
			return true
		}
		return false

	case int:
		if i.(int) == 1 {
			return true
		}
		return false

	case int64:
		if i.(int64) == 1 {
			return true
		}
		return false

	default:
		return false
	}
}

// GetRelations is designed to take a tablename and use it
// to determine a list of related objects.  this is just an
// idea, and the functionality will reqiure more than the
// return of a []string.
func (bf *BaseFlavor) GetRelations(tn string) []string {

	return nil
}

// CreateTables creates tables on the db based on
// the provided list of go struct definitions.
func (bf *BaseFlavor) CreateTables(i ...interface{}) error {

	// handled in each db flavor
	return fmt.Errorf("method CreateTables has not been implemented for %s", bf.GetDBDriverName())
}

// DropTables drops tables on the db if they exist, based on
// the provided list of go struct definitions.
func (bf *BaseFlavor) DropTables(i ...interface{}) error {

	dropSchema := ""
	for t := range i {

		// determine the table name
		tn := common.GetTableName(i[t])
		if tn == "" {
			return fmt.Errorf("unable to determine table name in bf.DropTables")
		}

		// if the table is found to exist, add a DROP statement
		// to the dropSchema string and move on to the next
		// table in the list.
		if bf.ExistsTable(tn) {
			if bf.log {
				fmt.Printf("table %s exists - adding to drop schema...\n", tn)
			}
			// submit 1 at a time for mysql
			dropSchema = dropSchema + fmt.Sprintf("DROP TABLE %s; ", tn)
			bf.ProcessSchema(dropSchema)
			dropSchema = ""
		}
	}
	return nil
}

// AlterTables alters tables on the db based on
// the provided list of go struct definitions.
func (bf *BaseFlavor) AlterTables(i ...interface{}) error {

	return fmt.Errorf("method AlterTables has not been implemented for %s", bf.GetDBDriverName())
}

// DestructiveResetTables drops tables on the db if they exist,
// as well as any related objects such as sequences.  this is
// useful if you wish to regenerated your table and the
// number-range used by an auto-incementing primary key.
func (bf *BaseFlavor) DestructiveResetTables(i ...interface{}) error {

	return fmt.Errorf("method DestructiveResetTable has not been implemented for %s", bf.GetDBDriverName())
}

// ExistsTable checks the currently connected database and
// returns true if the named table is found to exist.
func (bf *BaseFlavor) ExistsTable(tn string) bool {

	//	SELECT * FROM information_schema.TABLES	WHERE table_schema = 'jsonddl' AND table_name = 'equipment';
	n := 0
	bf.db.QueryRow("SELECT count(*) FROM INFORMATION_SCHEMA.TABLES WHERE table_schema = ? AND table_name = ?;", bf.GetDBName(), tn).Scan(&n)
	if n > 0 {
		return true
	}
	return false
}

// ExistsColumn checks the currently connected database and
// returns true if the named table-column is found to exist.
// this checks the column name only, not the column data-type
// or properties.
func (bf *BaseFlavor) ExistsColumn(tn string, cn string) bool {

	// SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = 'jsonddl' AND table_name = 'equipment' AND column_name = 'description';
	n := 0
	if bf.ExistsTable(tn) {
		bf.db.QueryRow("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = ? AND table_name = ? AND column_name = ?;", bf.GetDBName(), tn, cn).Scan(&n)
		if n > 0 {
			return true
		}
	}
	return false
}

// CreateIndex creates the index contained in the incoming
// IndexInfo structure.  indexes will be created as non-unique
// by default, and in multi-field situations, the fields will
// added to the index in the order they are contained in the
// IndexInfo.[]IndexFields slice.
func (bf *BaseFlavor) CreateIndex(in string, index IndexInfo) error {

	// CREATE INDEX idx_material_num_int_example ON `equipment`(material_num, int_example)
	fList := ""
	indexSchema := ""

	if len(index.IndexFields) == 1 {
		fList = index.IndexFields[0]
		// in = "idx_" + fList
		in = "idx_" + index.TableName + "_" + fList
	} else {
		for _, f := range index.IndexFields {
			fList = fmt.Sprintf("%s%s,", fList, f)
		}
		fList = strings.TrimSuffix(fList, ",")
	}

	if !index.Unique {
		indexSchema = fmt.Sprintf("CREATE INDEX %s ON %s (%s)", in, index.TableName, fList)
	} else {
		indexSchema = fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s)", in, index.TableName, fList)
	}

	bf.ProcessSchema(indexSchema)
	return nil
}

// DropIndex drops the specfied index on the connected database.
func (bf *BaseFlavor) DropIndex(tn string, in string) error {

	return fmt.Errorf("method DropIndex has not been implemented for %s", bf.GetDBDriverName())
}

// ExistsIndex checks the connected database for the presence
// of the specified index.
func (bf *BaseFlavor) ExistsIndex(tn string, in string) bool {

	n := 0
	bf.db.QueryRow("SELECT count(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE table_schema = ? AND table_name = ? AND index_name = ?", bf.GetDBName(), tn, in).Scan(&n)
	if n > 0 {
		return true
	}
	return false
}

// CreateSequence may be used to create a new sequence on the
// currently connected database.
func (bf *BaseFlavor) CreateSequence(sn string, start int) {

	log.Printf("method CreateSequence has not been implemented for %s\n", bf.GetDBDriverName())
	return
}

// AlterSequenceStart may be used to make changes to the start value
// of the named sequence, autoincrement or identity field depending
// on the manner in which the currently connected database flavour
// handles key generation.
func (bf *BaseFlavor) AlterSequenceStart(name string, start int) error {

	return fmt.Errorf("AlterSequenceStart has not been implemented for %s", bf.GetDBName())
}

// DropSequence may be used to drop the named sequence on the currently
// connected database.  This is probably not needed, as we are now
// creating sequences on postgres in a more correct manner.
// select pg_get_serial_sequence('public.some_table', 'some_column');
func (bf *BaseFlavor) DropSequence(sn string) error {

	return fmt.Errorf("DropSequence has not been implemented for %s", bf.GetDBDriverName())
}

// ExistsSequence checks for the presence of the named sequence on
// the currently connected database.
func (bf *BaseFlavor) ExistsSequence(sn string) bool {

	log.Printf("method ExistsSequence has not been implemented for %s\n", bf.GetDBDriverName())
	return false
}

// GetNextSequenceValue returns the next value of the named or derived
// sequence, auto-increment or identity field depending on which
// db-system is presently being used.
func (bf *BaseFlavor) GetNextSequenceValue(name string) (int, error) {

	return 0, fmt.Errorf("ExistsSequence has not been implemented for %s", bf.GetDBDriverName())
}

//===============================================================================
// SQL Schema Processing
//===============================================================================

// ProcessSchema processes the schema against the connected DB.
func (bf *BaseFlavor) ProcessSchema(schema string) {

	// MustExec panics on error, so just call it
	// bf.DB.MustExec(schema)
	bf.QsLog(schema)
	result, err := bf.db.Exec(schema)
	if err != nil {
		fmt.Println("err:", err)
	}

	// not all db's support rows affected, so reading it is
	// for interests sake only.
	ra, err := result.RowsAffected()
	if err != nil {
		return
	}

	if bf.log {
		fmt.Printf("%d rows affected.\n", ra)
	}
}

// ProcessSchemaList processes the schemas contained in sList
// in the order in which they were provided.  Schemas are
// executed against the connected DB.
func (bf *BaseFlavor) ProcessSchemaList(sList []string) {

	// bf.DB.MustExec(query string, args ...interface{})
	return
}

//===============================================================================
// SQL Query Processing
//===============================================================================

// ExecuteQueryRow processes the single-row query contained in queryString
// against the connected DB using sql/database.
func (bf *BaseFlavor) ExecuteQueryRow(queryString string, qParams ...interface{}) *sql.Row {

	if qParams != nil {
		queryString = bf.db.Rebind(queryString)
		bf.QsLog(queryString, qParams...)
		return bf.db.QueryRow(queryString, qParams...)
	}
	bf.QsLog(queryString)
	return bf.db.QueryRow(queryString)
}

// ExecuteQuery processes the multi-row query contained in queryString
// against the connected DB using sql/database.
func (bf *BaseFlavor) ExecuteQuery(queryString string, qParams ...interface{}) (*sql.Rows, error) {

	var rows *sql.Rows
	var err error

	if qParams != nil {
		queryString = bf.db.Rebind(queryString)
		bf.QsLog(queryString, qParams...)
		rows, err = bf.db.Query(queryString, qParams...)
	} else {
		bf.QsLog(queryString)
		rows, err = bf.db.Query(queryString)
	}
	return rows, err
}

// ExecuteQueryRowx processes the single-row query contained in queryString
// against the connected DB using sqlx.
func (bf *BaseFlavor) ExecuteQueryRowx(queryString string, qParams ...interface{}) *sqlx.Row {

	if qParams != nil {
		queryString = bf.db.Rebind(queryString)
		bf.QsLog(queryString, qParams...)
		return bf.db.QueryRowx(queryString, qParams...)
	}
	bf.QsLog(queryString)
	return bf.db.QueryRowx(queryString)
}

// ExecuteQueryx processes the multi-row query contained in queryString
// against the connected DB using sqlx.
func (bf *BaseFlavor) ExecuteQueryx(queryString string, qParams ...interface{}) (*sqlx.Rows, error) {

	var rows *sqlx.Rows
	var err error

	if qParams != nil {
		queryString = bf.db.Rebind(queryString)
		bf.QsLog(queryString, qParams...)
		rows, err = bf.db.Queryx(queryString, qParams...)
	} else {
		bf.QsLog(queryString)
		rows, err = bf.db.Queryx(queryString)
	}
	return rows, err
}

// Get reads a single row into the dst interface.
// This calls sqlx.Get(...)
func (bf *BaseFlavor) Get(dst interface{}, queryString string, args ...interface{}) error {

	if args != nil {
		queryString = bf.db.Rebind(queryString)
		bf.QsLog(queryString, args...)
		return bf.db.Get(dst, queryString, args...)
	}
	bf.QsLog(queryString)
	return bf.db.Get(dst, queryString)
}

// Select reads some rows into the dst interface.
// This calls sqlx.Select(...)
func (bf *BaseFlavor) Select(dst interface{}, queryString string, args ...interface{}) error {

	if args != nil {
		queryString = bf.db.Rebind(queryString)
		bf.QsLog(queryString, args...)
		return bf.db.Select(dst, queryString, args...)
	}
	bf.QsLog(queryString)
	return bf.db.Select(dst, queryString)
}

// Exec runs the queryString against the connected db
func (bf *BaseFlavor) Exec(queryString string, args ...interface{}) (sql.Result, error) {

	var result sql.Result
	var err error

	if args != nil {
		queryString = bf.db.Rebind(queryString)
		bf.QsLog(queryString, args...)
		result, err = bf.db.Exec(queryString, args...)
	} else {
		bf.QsLog(queryString)
		result, err = bf.db.Exec(queryString)
	}
	return result, err
}

// ProcessTransaction processes the list of commands as a transaction.
// If any of the commands encounter an error, the transaction will be
// cancelled via a Rollback and the error message will be returned to
// the caller.  It is assumed that tList contains bound queryStrings.
func (bf *BaseFlavor) ProcessTransaction(tList []string) error {

	tx, err := bf.db.Begin()
	if err != nil {
		return err
	}
	for _, s := range tList {
		_, err = tx.Exec(s, nil)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

// processIndexTag is used to create or add to an entry in the working indexes map that is
// being built in a CreateTable or AlterTable method.
func (bf *BaseFlavor) processIndexTag(iMap map[string]IndexInfo, tableName string, fieldName string,
	indexName string, unique bool, singleField bool) map[string]IndexInfo {

	var fldIndex IndexInfo

	// single column index
	if singleField {
		fldIndex.TableName = tableName
		fldIndex.IndexFields = append(fldIndex.IndexFields, fieldName)
		if unique {
			fldIndex.Unique = true
		} else {
			fldIndex.Unique = false
		}
		indexName = fmt.Sprintf("%s%s_%s", indexName, tableName, fieldName)
		iMap[indexName] = fldIndex
		return iMap
	}

	// multi-column indexes where the index-name is in the map
	fldIndex, ok := iMap[indexName]
	if ok {
		fldIndex.IndexFields = append(fldIndex.IndexFields, fieldName)
		iMap[indexName] = fldIndex
		return iMap
	}

	// add a multi-column index to the map
	fldIndex.TableName = tableName
	fldIndex.IndexFields = append(fldIndex.IndexFields, fieldName)
	fldIndex.Unique = false
	iMap[indexName] = fldIndex
	return iMap
}

// Delete - Delete an existing entity (single-row) on the database using the full-key
func (bf *BaseFlavor) Delete(ent interface{}) error { // (id uint) error

	var info CrudInfo
	info.ent = ent
	info.log = false
	info.mode = "D"

	err := bf.BuildComponents(&info)
	if err != nil {
		return err
	}

	keyList := ""
	for k, s := range info.keyMap {

		fType := reflect.TypeOf(s).String()
		if bf.IsLog() {
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
	delQuery := fmt.Sprintf("DELETE FROM %s", info.tn)
	delQuery = fmt.Sprintf("%s WHERE%s;", delQuery, keyList)
	bf.QsLog(delQuery)

	// attempt the delete and read result back into resultMap
	row := bf.db.QueryRowx(delQuery)
	if row.Err() != nil {
		return err
	}
	return nil
}

// GetEntity - get an existing entity from the db using the primary
// key definition.  The entire key should be provided, although
// providing a partial key will not generate an (obvious) error.
func (bf *BaseFlavor) GetEntity(ent interface{}) error {

	var info CrudInfo
	info.ent = ent
	info.log = false
	info.mode = "G"

	err := bf.BuildComponents(&info)
	if err != nil {
		return err
	}

	keyList := ""
	for k, s := range info.keyMap {

		fType := reflect.TypeOf(s).String()
		if bf.IsLog() {
			fmt.Printf("key: %v, value: %v\n", k, s)
			fmt.Println("TYPE:", fType)
		}

		if fType == "string" {
			keyList = fmt.Sprintf("%s %s = '%v' AND", keyList, k, s)
		} else {
			keyList = fmt.Sprintf("%s %s = %v AND", keyList, k, s)
		}
		keyList = strings.TrimSuffix(keyList, " AND")

		selQuery := fmt.Sprintf("SELECT * FROM %s", info.tn)
		selQuery = fmt.Sprintf("%s WHERE%s;", selQuery, keyList)
		bf.QsLog(selQuery)

		// attempt read the entity row
		err := bf.db.QueryRowx(selQuery).StructScan(info.ent) //.MapScan(info.resultMap) // SliceScan
		if err != nil {
			return err
		}
		info.entValue = reflect.ValueOf(info.ent)
		return nil
	}
	return nil
}

// GetEntities - this is one way to do it, but reflection feels like a mistake
// in the CRUD code.  consider an alternative approach whereby a CRUD interface
// is declared using the model.<ent> as the basis.
//
// type CRUDEnt{} interface
//
//  Create(h *sqac.Handle)
//  Read()
//  Update()
//  Delete()
//  GetEntities()
//
//  Each model would have a corresponding interface implementation,
//  based on its own data.  The interface object would be passed from
//  the application side into the CRUD part of the ORM.  The ORM would
//  then provide a sqac.Handle that would be used inside the interface
//  implementation to perform the required CRUD task.  It follows that
//  the data for the Create, Read, Update, Delete would need to be set
//  inside the interface object prior to its being passed to the ORM.
//  The interface object would be passed by reference, allowing the
//  results to be 'passed' back to the caller without any fuss or
//  delay.  Advantage - no reflection.  Disadvantage - it seems rather
//  messy, as the extraction of the retrieved data would likely be
//  though the struct or slice of structs the interface is implemented
//  on.
func (bf *BaseFlavor) GetEntities(ents interface{}) (interface{}, error) {

	// get the underlying data type of the interface{}
	entTypeElem := reflect.TypeOf(ents).Elem()
	// fmt.Println("entTypeElem:", entTypeElem)

	// create a struct from the type
	testVar := reflect.New(entTypeElem)

	// determine the db table name
	tn := common.GetTableName(ents)

	selQuery := fmt.Sprintf("SELECT * FROM %s;", tn)
	bf.QsLog(selQuery)

	// read the rows
	rows, err := bf.db.Queryx(selQuery)
	if err != nil {
		log.Printf("GetEntities for table &s returned error: %v\n", err.Error())
		return nil, err
	}

	// iterate over the rows collection and put the results
	// into the ents interface (slice)
	entsv := reflect.ValueOf(ents)
	for rows.Next() {
		err = rows.StructScan(testVar.Interface())
		if err != nil {
			fmt.Println("scan error:", err)
			return nil, err
		}
		// fmt.Println(testVar)
		entsv = reflect.Append(entsv, testVar.Elem())
	}

	ents = entsv.Interface()
	// fmt.Println("ents:", ents)
	return entsv.Interface(), nil
}

// GetEntities2 attempts to retrieve all entities based
// on the internal implementation of GetEnt.  GetEnt
// exposes a single method (Exec) to execute the request.
// All this because go can only go so far with meta-type
// programming.
func (bf *BaseFlavor) GetEntities2(ge GetEnt) error {

	// Exec() should contain whatever SQL related code
	// is required to satisfy GetEntities2 for the underlying
	// model.<struc> or model.[]<struct> type.
	err := ge.Exec(bf)
	if err != nil {
		return err
	}
	if bf.IsLog() {
		fmt.Println("bf.GetEntities2 following Exec() contained: ", ge)
	}
	return nil
}

// GetEntities3 is experimental
func (bf *BaseFlavor) GetEntities3(ents interface{}) {

	// get the underlying data type of the interface{}
	sliceTypeElem := reflect.TypeOf(ents).Elem()
	// fmt.Println("entTypeElem:", entTypeElem)

	t := reflect.Indirect(reflect.ValueOf(ents)).Type().Elem()
	fmt.Println("t:", t)

	// create a struct from the type
	testVar := reflect.New(t)

	// determine the db table name
	tn := common.GetTableName(ents)

	selQuery := fmt.Sprintf("SELECT * FROM %s;", tn)
	bf.QsLog(selQuery)

	// read the rows
	rows, err := bf.db.Queryx(selQuery)
	if err != nil {
		log.Printf("GetEntities for table &s returned error: %v\n", err.Error())
		// return err
	}

	slice := reflect.MakeSlice(sliceTypeElem, 0, 0)
	for rows.Next() {
		err = rows.StructScan(testVar.Interface())
		if err != nil {
			fmt.Println("scan error:", err)
		}

		fmt.Println(testVar)
		slice = reflect.Append(slice, testVar.Elem())
	}

	fmt.Println("slice:", slice)
	// entsv := reflect.ValueOf(ents)
	// fmt.Println("entsv:", entsv)
	// *ents.(*interface{}) = slice.Interface() // reflect.ValueOf(slice)
	// *ents = slice
	fmt.Println("ents:", ents)
}

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

// IndexInfo contains index definitions as read from the sqac:"index" tags
type IndexInfo struct {
	TableName   string
	Unique      bool
	IndexFields []string
}

// FKeyInfo holds foreign-key defs read from the sqac:"fkey" tags
// sqac:"fkey:ref_table(ref_field)"
type FKeyInfo struct {
	FromTable string
	FromField string
	RefTable  string
	RefField  string
	FKeyName  string
}

// ForeignKeyBuffer is used to hold deferred foreign-key information
// pending the creation of all tables submitted in a CreateTable(...)
// or AlterTable(...) call.
type ForeignKeyBuffer struct {
	ent    interface{}
	fkinfo FKeyInfo
}

// ColComponents is used to capture the field properties from sqac: tags
// during table creation and table alteration activities.
type ColComponents struct {
	fName             string
	fType             string
	uType             string // user-specified DB type
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
	seq       []common.SqacPair
	ind       map[string]IndexInfo
	fkey      []FKeyInfo
	pk        string
	err       error
}

// Log dumps all of the raw table components to stdout is called for CreateTable
// and AlterTable operations if the main sqac logging has been activated via
// BaseFlavor.Log(true).
func (tc *TblComponents) Log() {
	log.Println("====================================================================")
	log.Println("TABLE SCHEMA:", tc.tblSchema)
	log.Println()
	for _, v := range tc.seq {
		log.Println("SEQUENCE:", v)
	}
	log.Println("--")
	for k, v := range tc.ind {
		log.Printf("INDEX: k:%s	fields:%v  unique:%v tableName:%s\n", k, v.IndexFields, v.Unique, v.TableName)
	}
	log.Println("--")
	log.Println("PRIMARY KEYS:", tc.pk)
	log.Println("--")
	for _, v := range tc.flDef {
		log.Printf("FIELD DEF: fname:%s, ftype:%s, gotype:%s ,nodb:%v\n", v.FName, v.FType, v.GoType, v.NoDB)
		for _, p := range v.SqacPairs {
			log.Printf("FIELD PROPERTY: %s, %v\n", p.Name, p.Value)
		}
		log.Println("------")
	}
	log.Println("--")
	log.Println("ERROR:", tc.err)
	log.Println("====================================================================")
}

// CTick CBackTick and CDblQuote specify the quote
// style for for db field encapsulation in CREATE
// and ALTER table schemas
const CTick = "'"
const CBackTick = "`"
const CDblQuote = "\""

// PublicDB exposes functions for db related operations.
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
	// information_schema access.  File-based databases like
	// sqlite report the name as the absolute path to the location
	// of their database file.
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

	// i=db/sqac tagged go struct-type
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

	// CreateForeignKey(Entity{}, foreignkeytable, reftable, fkfield, reffield)
	// &Entity{} (i) is only needed for SQLite - okay to pass nil in other cases.
	CreateForeignKey(i interface{}, ft, rt, ff, rf string) error
	DropForeignKey(i interface{}, ft, fkn string) error
	ExistsForeignKeyByName(i interface{}, fkn string) (bool, error)
	ExistsForeignKeyByFields(i interface{}, ft, rt, ff, rf string) (bool, error)

	// process DDL/DML commands
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
	Delete(ent interface{}) error    // (id uint) error
	GetEntity(ent interface{}) error // pass ptr to type containing key information
	GetEntities(ents interface{}) (interface{}, error)
	GetEntities2(ge GetEnt) error
	GetEntities4(ents interface{})
	GetEntitiesWithCommandsIP(ents interface{}, params []common.GetParam, cmdMap map[string]interface{}) (uint64, error)
	GetEntitiesWithCommands(ents interface{}, params []common.GetParam, cmdMap map[string]interface{}) (interface{}, error)
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
			log.Println("unable to determine DBName!")
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
// it is a nod to backward compatibility for existing db installs.
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
				log.Printf("table %s exists - adding to drop schema...\n", tn)
			}
			// submit 1 at a time for mysql
			dropSchema = dropSchema + "DROP TABLE " + tn + ";"
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

	n := 0
	qs := "SELECT count(*) FROM INFORMATION_SCHEMA.TABLES WHERE table_schema = ? AND table_name = ?;"
	dbName := bf.GetDBName()

	bf.QsLog(qs, dbName)
	bf.db.QueryRow(qs, dbName, tn).Scan(&n)
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

	n := 0
	qs := "SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = ? AND table_name = ? AND column_name = ?;"
	dbName := bf.GetDBName()

	if bf.ExistsTable(tn) {
		bf.QsLog(qs, dbName, tn, cn)
		bf.db.QueryRow(qs, dbName, tn, cn).Scan(&n)
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
		in = "idx_" + index.TableName + "_" + fList
	} else {
		for _, f := range index.IndexFields {
			fList = fList + f + ", "
		}
		fList = strings.TrimSuffix(fList, ", ")
	}

	if !index.Unique {
		indexSchema = "CREATE INDEX " + in + " ON " + index.TableName + " (" + fList + ");"
	} else {
		indexSchema = "CREATE UNIQUE INDEX " + in + " ON " + index.TableName + " (" + fList + ");"
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
	qs := "SELECT count(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE table_schema = ? AND table_name = ? AND index_name = ?"
	dbName := bf.GetDBName()

	bf.QsLog(qs, dbName, tn, in)
	bf.db.QueryRow(qs, dbName, tn, in).Scan(&n)
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

// CreateForeignKey creates a foreign-key on an existing column.
func (bf *BaseFlavor) CreateForeignKey(i interface{}, ft, rt, ff, rf string) error {

	schema := "ALTER TABLE " + ft + " ADD CONSTRAINT " + "fk_" + ft + "_" + rt + "_" + rf + " FOREIGN KEY(" + ff + ")" + " REFERENCES " + rt + "(" + rf + ");"
	bf.QsLog(schema)

	_, err := bf.Exec(schema)
	if err != nil {
		return err
	}
	return nil
}

// DropForeignKey drops a foreign-key on an existing column
func (bf *BaseFlavor) DropForeignKey(i interface{}, ft, fkn string) error {

	schema := "ALTER TABLE " + ft + " DROP CONSTRAINT " + fkn + ";"
	bf.QsLog(schema)

	_, err := bf.Exec(schema)
	if err != nil {
		return err
	}
	return nil
}

// ExistsForeignKeyByName checks to see if the named foreign-key exists on the
// table corresponding to provided sqac model (i).
func (bf *BaseFlavor) ExistsForeignKeyByName(i interface{}, fkn string) (bool, error) {

	return false, fmt.Errorf("ExistsForeignKeyByName(...) has not been implemented for %s", bf.GetDBDriverName())
}

// ExistsForeignKeyByFields checks to see if a foreign-key exists between the named
// tables and fields.
func (bf *BaseFlavor) ExistsForeignKeyByFields(i interface{}, ft, rt, ff, rf string) (bool, error) {

	return false, fmt.Errorf("ExistsForeignKeyByFields(...) has not been implemented for %s", bf.GetDBDriverName())
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
		log.Println("ProcessSchema err:", err)
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
// DEPRECATED: USER ProcessTransactionList
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
		bf.QsLog(queryString, args...)
		queryString = bf.db.Rebind(queryString)
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

	// begin the transaction
	tx, err := bf.db.Begin()
	if err != nil {
		return err
	}

	// execute each command in the transaction set
	for _, s := range tList {
		bf.QsLog(s)
		_, err = tx.Exec(s, nil)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// commit
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

// processFKeyTag places the data for foreign-key creation in a slice that will be
// added to the tc (TblComponents) struct.  Foreign-keys are always created after all
// tables in a Create / Alter set have been processed in order to provide the greatest
// chance that the corresponding ref-table/field exist.  Could add the constraint name
// here...
func (bf *BaseFlavor) processFKeyTag(fkeys []FKeyInfo, ft, ff, rv string) []FKeyInfo {

	tf := strings.Split(rv, "(")
	if len(tf) != 2 {
		// panic for now
		panic(fmt.Sprintf("unable to parse foreign-key sqac tag: %v", rv))
	}

	rt := tf[0]
	rf := tf[1]
	rt = strings.TrimSpace(rt)
	rf = strings.Replace(rf, ")", "", 2)
	rf = strings.TrimSpace(rf)

	fk := FKeyInfo{
		FromTable: ft, // from-table
		FromField: ff, // from-field
		RefTable:  rt, // ref-table
		RefField:  rf, // ref-field
	}
	return append(fkeys, fk)
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
		indexName = indexName + tableName + "_" + fieldName
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

// Delete - CRUD Delete an existing entity (single-row) on the database using the full-key
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
		if fType == "string" {
			keyList = fmt.Sprintf("%s %s = '%v' AND", keyList, k, s)
		} else {
			keyList = fmt.Sprintf("%s %s = %v AND", keyList, k, s)
		}
	}

	keyList = strings.TrimSuffix(keyList, " AND")
	delQuery := "DELETE FROM " + info.tn + " WHERE " + keyList + ";"
	bf.QsLog(delQuery)

	result, err := bf.db.Exec(delQuery)
	if err != nil {
		log.Println("CRUD Delete error:", err)
	}

	ra, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if bf.log {
		fmt.Printf("%d rows affected.\n", ra)
	}
	return nil
}

// GetEntity - CRUD GetEntity gets an existing entity from the db using the primary
// key definition.  It is expected that ID will have been populated in the body by
// the caller.
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
			log.Printf("CRUD GET ENTITY key: %v, value: %v\n", k, s)
			log.Println("CRUD GET ENTITY TYPE:", fType)
		}

		// not worth coding the detailed cases here - use Sprintf %v
		if fType == "string" {
			keyList = fmt.Sprintf("%s %s = '%v' AND", keyList, k, s)
		} else {
			keyList = fmt.Sprintf("%s %s = %v AND", keyList, k, s)
		}
		keyList = strings.TrimSuffix(keyList, " AND")

		selQuery := "SELECT * FROM " + info.tn + " WHERE " + keyList + ";"
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

// GetEntities is experimental - use GetEntitiesWithCommands.
func (bf *BaseFlavor) GetEntities(ents interface{}) (interface{}, error) {

	// get the underlying data type of the interface{}
	entTypeElem := reflect.TypeOf(ents).Elem()
	// fmt.Println("entTypeElem:", entTypeElem)

	// create a struct from the type
	testVar := reflect.New(entTypeElem)

	// determine the db table name
	tn := common.GetTableName(ents)

	selQuery := "SELECT * FROM " + tn + ";"
	bf.QsLog(selQuery)

	// read the rows
	rows, err := bf.db.Queryx(selQuery)
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
			log.Println("GetEntities scan error:", err)
			return nil, err
		}
		entsv = reflect.Append(entsv, testVar.Elem())
	}

	ents = entsv.Interface()
	// fmt.Println("ents:", ents)
	return entsv.Interface(), nil
}

// GetEntities2 attempts to retrieve all entities based on the internal implementation of GetEnt.
// GetEnt exposes a single method (Exec) to execute the request.  All this because go can only go
// so far with meta-type programming in go before you get buried in reflection.
// ge allows you to pass a sqac handle into get entities, then you can do what you need to do.
// GetEntities2 has been replaced by GetEntitiesWithCommands, but can be used if you want a clean
// looking API that is pretty quick (very light use of reflection).
// That said, it is a --dirty-- way of doing things.
func (bf *BaseFlavor) GetEntities2(ge GetEnt) error {

	// Exec() should contain whatever SQL related code
	// is required to satisfy GetEntities2 for the underlying
	// model.<struc> or model.[]<struct> type.
	err := ge.Exec(bf)
	if err != nil {
		return err
	}
	if bf.IsLog() {
		log.Println("bf.GetEntities2 following Exec() contained: ", ge)
	}
	return nil
}

// GetEntities4 is experimental - use GetEntitiesByCommands or GetEntitiesByCommandsIP.
//
// This method uses alot of reflection to permit the retrieval of the equivalent of
// []interface{} where interface{} can be taken to mean Model{}.  This can be used,
// but may prove to be a slow way of doing things.
// A quick internet search on []interface{} will turn up all sorts of acrimony.  Notice
// that the method signature is still interface{}?  Not very transparent.
func (bf *BaseFlavor) GetEntities4(ents interface{}) {

	// get the underlying data type of the interface{} ([]ModelEtc)
	sliceTypeElem := reflect.TypeOf(ents).Elem()

	// get the underlying (struct?) type of the slice
	t := reflect.Indirect(reflect.ValueOf(ents)).Type().Elem()
	// fmt.Println("t:", t)

	// create a struct from the type
	dstRow := reflect.New(t)

	// determine the db table name
	tn := common.GetTableName(ents)

	selQuery := "SELECT * FROM " + tn + ";"
	bf.QsLog(selQuery)

	// read the rows
	rows, err := bf.db.Queryx(selQuery)
	if err != nil {
		log.Printf("GetEntities for table %s returned error: %v\n", tn, err.Error())
		// return err
	}
	defer rows.Close()

	// this is where it happens for GetEntities4(...) and why I am  not too
	// satisfied with generic programming in go:
	eValue := reflect.ValueOf(ents)
	for eValue.Kind() == reflect.Ptr {
		eValue = eValue.Elem()
	}

	results := eValue
	resultType := results.Type().Elem()
	results.Set(reflect.MakeSlice(results.Type(), 0, 0))

	if resultType.Kind() == reflect.Ptr {
		resultType = resultType.Elem()
	}

	slice := reflect.MakeSlice(sliceTypeElem, 0, 0)
	for rows.Next() {
		err = rows.StructScan(dstRow.Interface())
		if err != nil {
			log.Println("GetEntities4 scan error:", err)
		}
		slice = reflect.Append(slice, dstRow.Elem())
		results.Set(reflect.Append(results, dstRow.Elem()))
	}
	// fmt.Println("slice:", slice)
	// fmt.Println("")
	// fmt.Println("results:", results)
	// fmt.Println("ents:", ents)
}

// GetEntitiesWithCommandsIP uses alot of reflection to permit the retrieval of the equivalent
// of []interface{} where interface{} can be taken to mean Model{}.  This can be used, but may
// not be the fastest way to do the selection in terms of processing in the go runtime.
// A quick internet search on []interface{} will turn up all sorts of acrimony related to
// refelection and using interface{} as []interface.
// However, this is the Get method that is preferred, as the caller can simply pass the address
// of a slice of the requested table-type in the ents parameter, then read the resulting
// slice directly in their program following method execution.
// Each DB needs slightly different handling due to differences in OFFSET / LIMIT / TOP support.
// This is a mostly common version, but MSSQL has its own specific implementation due to
// some extra differences in transact-SQL.
func (bf *BaseFlavor) GetEntitiesWithCommandsIP(ents interface{}, params []common.GetParam, cmdMap map[string]interface{}) (result uint64, err error) {

	var count uint64
	var row *sqlx.Row
	paramString := ""
	selQuery := ""

	// get the underlying data type of the interface{} ([]ModelEtc)
	sliceTypeElem := reflect.TypeOf(ents).Elem()

	// get the underlying (struct?) type of the slice
	t := reflect.Indirect(reflect.ValueOf(ents)).Type().Elem()

	// create a struct from the type
	dstRow := reflect.New(t)

	// determine the db table name
	tn := common.GetTableName(ents)

	// are there any parameters to include in the query?
	var pv []interface{}
	if params != nil && len(params) > 0 {
		paramString = " WHERE"
		for i := range params {
			paramString = paramString + " " + common.CamelToSnake(params[i].FieldName) + " " + params[i].Operand + " ? " + params[i].NextOperator
			pv = append(pv, params[i].ParamValue)
		}
	}

	// received a $count command?  this supercedes all, as it should not
	// be mixed with any other $<commands>.
	_, ok := cmdMap["count"]
	if ok {
		if paramString == "" {
			selQuery = "SELECT COUNT(*) FROM " + tn + ";"
			bf.QsLog(selQuery)
			row = bf.ExecuteQueryRowx(selQuery)
		} else {
			selQuery = "SELECT COUNT(*) FROM " + tn + paramString + ";"
			fmt.Println("S1:", selQuery)
			bf.QsLog(selQuery)
			row = bf.ExecuteQueryRowx(selQuery, pv...)
		}

		err = row.Scan(&count)
		if err != nil {
			return 0, err
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
		obString = " ORDER BY " + obField.(string)
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

	// received $limit command?
	limField, ok := cmdMap["limit"]
	if ok {
		limitString = fmt.Sprintf(" LIMIT %v", limField)
	}

	// received $offset command?  some db's require a limit with offset....
	offField, ok := cmdMap["offset"]
	if ok {
		switch bf.GetDBDriverName() {
		case "sqlite3":
			// set -1 for open-ended limit
			if limitString == "" {
				limitString = " LIMIT -1"
			}
		case "mysql":
			// set 18446744073709551615 for open-ended limit :P
			if limitString == "" {
				limitString = " LIMIT 18446744073709551615"
			}
		case "hdb":
			if limitString == "" {
				limitString = " LIMIT null"
			}
		case "mssql":
			// handled in mssql_flavor

		default:

		}
		offsetString = fmt.Sprintf(" OFFSET %v", offField)
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

	selQuery = "SELECT * FROM " + tn + paramString
	selQuery = bf.db.Rebind(selQuery)
	selQuery = selQuery + obString + adString + limitString + offsetString + ";"
	bf.QsLog(selQuery)

	// read the rows
	rows, err := bf.db.Queryx(selQuery, pv...)
	if err != nil {
		log.Printf("GetEntities for table %s returned error: %v\n", tn, err.Error())
		return 0, err
	}

	defer rows.Close()

	// this is where it happens for GetEntities5(...) and why I am  not too
	// satisfied with generic programming in go:
	eValue := reflect.ValueOf(ents)
	for eValue.Kind() == reflect.Ptr {
		eValue = eValue.Elem()
	}

	results := eValue
	resultType := results.Type().Elem()
	results.Set(reflect.MakeSlice(results.Type(), 0, 0))

	if resultType.Kind() == reflect.Ptr {
		resultType = resultType.Elem()
	}

	var c uint64
	slice := reflect.MakeSlice(sliceTypeElem, 0, 0)
	for rows.Next() {
		err = rows.StructScan(dstRow.Interface())
		if err != nil {
			log.Println("GetEntitiesWithCommandsIP scan error:", err)
			return 0, err
		}
		slice = reflect.Append(slice, dstRow.Elem())
		results.Set(reflect.Append(results, dstRow.Elem()))
		c++
	}
	// fmt.Println("slice:", slice)
	// fmt.Println("")
	// fmt.Println("results:", results)
	// fmt.Println("ents:", ents)
	return c, nil
}

// GetEntitiesWithCommands can be used as a get for lists of entities.  Each DB needs
// slightly different handling due to differences in OFFSET / LIMIT / TOP support.
// This is a mostly common version, but MSSQL has its own specific implementation due to
// some extra differences in transact-SQL.  This method still requires that the caller
// perform a type-assertion on the returned interface{} ([]interface{}) parameter.
func (bf *BaseFlavor) GetEntitiesWithCommands(ents interface{}, params []common.GetParam, cmdMap map[string]interface{}) (interface{}, error) {

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
			paramString = paramString + " " + common.CamelToSnake(params[i].FieldName) + " " + params[i].Operand + " ? " + params[i].NextOperator
			pv = append(pv, params[i].ParamValue)
		}
	}

	// received a $count command?  this supercedes all, as it should not
	// be mixed with any other $<commands>.
	_, ok := cmdMap["count"]
	if ok {
		if paramString == "" {
			selQuery = "SELECT COUNT(*) FROM " + tn + ";"
			bf.QsLog(selQuery)
			row = bf.ExecuteQueryRowx(selQuery)
		} else {
			selQuery = "SELECT COUNT(*) FROM " + tn + paramString + ";"
			fmt.Println("S1:", selQuery)
			bf.QsLog(selQuery)
			row = bf.ExecuteQueryRowx(selQuery, pv...)
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
		obString = " ORDER BY " + obField.(string)
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

	// received $limit command?
	limField, ok := cmdMap["limit"]
	if ok {
		limitString = fmt.Sprintf(" LIMIT %v", limField)
	}

	// received $offset command?  some db's require a limit with offset....
	offField, ok := cmdMap["offset"]
	if ok {
		switch bf.GetDBDriverName() {
		case "sqlite3":
			// set -1 for open-ended limit
			if limitString == "" {
				limitString = " LIMIT -1"
			}
		case "mysql":
			// set 18446744073709551615 for open-ended limit :P
			if limitString == "" {
				limitString = " LIMIT 18446744073709551615"
			}
		case "hdb":
			if limitString == "" {
				limitString = " LIMIT null"
			}
		case "mssql":
			// handled in mssql_flavor

		default:

		}
		offsetString = fmt.Sprintf(" OFFSET %v", offField)
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

	selQuery = "SELECT * FROM " + tn + paramString
	selQuery = bf.db.Rebind(selQuery)
	selQuery = selQuery + obString + adString + limitString + offsetString + ";"
	bf.QsLog(selQuery)

	// read the rows
	rows, err := bf.db.Queryx(selQuery, pv...)
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
			log.Println("GetEntitiesWithCommand scan error:", err)
			return nil, err
		}
		entsv = reflect.Append(entsv, testVar.Elem())
	}
	ents = entsv.Interface()
	return entsv.Interface(), nil
}

// this is where it happens for GetEntities4(...)
// func indirect(reflectValue reflect.Value) reflect.Value {
// 	for reflectValue.Kind() == reflect.Ptr {
// 		reflectValue = reflectValue.Elem()
// 	}
// 	return reflectValue
// }

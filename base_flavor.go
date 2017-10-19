package sqac

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

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
	fName       string
	fType       string
	fPrimaryKey string
	fDefault    string
	fNullable   string
}

// TblComponents is used as a collector structure for internal table
// create / alter processing.
type TblComponents struct {
	tblSchema string
	flDef     []FieldDef
	seq       []RgenPair
	ind       map[string]IndexInfo
	pk        string
	err       error
}

// PublicDB exposes functions for db schema operations
type PublicDB interface {
	InBase()
	InDB()

	// postgres, sqlite, mariadb, db2, hana etc.
	GetDBDriverName() string

	// activate / check logging
	Log(b bool)
	IsLog() bool

	// set the *sqlx.DB handle in the PublicDB interface
	SetDB(db *sqlx.DB)
	GetDB() *sqlx.DB

	// get the currently connected db-name for infomation_schema lookups
	GetDBName() string

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
	DropIndex(in string) error
	ExistsIndex(tn string, in string) bool

	// sn=sequenceName, start=start-value
	CreateSequence(sn string, start int)
	AlterSequenceStart(sn string, start int) error
	// select pg_get_serial_sequence('public.some_table', 'some_column');
	DropSequence(sn string) error
	ExistsSequence(sn string) bool

	// CreateForeignKey(...) error
	// BuildForeignKeyName(...) error
	// DropForeignKey(...) error
	// ExistsForeignKey(...) bool

	ProcessSchema(schema string)
	ProcessSchemaList(sList []string)

	// sql package access
	ExecuteQueryRow(queryString string, qParams ...interface{}) *sql.Row
	ExecuteQuery(queryString string, qParams ...interface{}) (*sql.Rows, error)
	Exec(queryString string, args ...interface{}) (sql.Result, error)

	// sqlx package access
	ExecuteQueryRowx(queryString string, qParams ...interface{}) *sqlx.Row
	ExecuteQueryx(queryString string, qParams ...interface{}) (*sqlx.Rows, error)
	Get(dst interface{}, queryString string, args ...interface{}) error
	Select(dst interface{}, queryString string, args ...interface{}) error
}

// ensure consistency of interface implementation
var _ PublicDB = &BaseFlavor{}

// BaseFlavor is a supporting struct for interface PublicDB
type BaseFlavor struct {
	db  *sqlx.DB
	log bool
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

func (bf *BaseFlavor) InBase() {
	fmt.Println("InBase() called in BF")
}

func (bf *BaseFlavor) InDB() {
	fmt.Println("InDB called in BF")
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
	return nil
}

// DropTables drops tables on the db if they exist, based on
// the provided list of go struct definitions.
func (bf *BaseFlavor) DropTables(i ...interface{}) error {

	dropSchema := ""
	fmt.Println("in bf.DropTables")
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
		if bf.ExistsTable(tn) {
			if bf.log {
				fmt.Printf("table %s exists - adding to drop schema...\n", tn)
			}
			dropSchema = dropSchema + fmt.Sprintf("DROP TABLE IF EXISTS %s;", tn)
		}
	}
	if dropSchema != "" {
		bf.ProcessSchema(dropSchema)
	}
	return nil
}

// AlterTables alters tables on the db based on
// the provided list of go struct definitions.
func (bf *BaseFlavor) AlterTables(i ...interface{}) error {

	return nil
}

// DestructiveResetTables drops tables on the db if they exist,
// as well as any related objects such as sequences.  this is
// useful if you wish to regenerated your table and the
// number-range used by an auto-incementing primary key.
// func (bf *BaseFlavor) DestructiveResetTables(i ...interface{}) error {

// 	err := bf.DropTables(i...)
// 	if err != nil {
// 		return err
// 	}
// 	err = bf.CreateTables(i...)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

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
	bf.db.QueryRow("SELECT count(*) FROM INFORMATION_SCHEMA.TABLES WHERE table_schema = ? AND table_name = ?;", bf.GetDBName(), tn, cn).Scan(&n)
	if n > 0 {
		return true
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
		in = "idx_" + fList
	} else {
		for _, f := range index.IndexFields {
			fList = fmt.Sprintf("%s %s,", fList, f)
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
func (bf *BaseFlavor) DropIndex(in string) error {

	indexSchema := fmt.Sprintf("DROP INDEX IF EXISTS %s;", in)
	bf.ProcessSchema(indexSchema)
	return nil
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

	return
}

// AlterSequenceStart may be used to make changes to the start
// value of the named sequence on the currently connected database.
func (bf *BaseFlavor) AlterSequenceStart(sn string, start int) error {

	return nil
}

// DropSequence may be used to drop the named sequence on the currently
// connected database.  This is probably not needed, as we are now
// creating sequences on postgres in a more correct manner.
// select pg_get_serial_sequence('public.some_table', 'some_column');
func (bf *BaseFlavor) DropSequence(sn string) error {

	return nil
}

// ExistsSequence checks for the presence of the named sequence on
// the currently connected database.
func (bf *BaseFlavor) ExistsSequence(sn string) bool {

	return false
}

//===============================================================================
// SQL Schema Processing
//===============================================================================

// ProcessSchema processes the schema against the connected DB.
func (bf *BaseFlavor) ProcessSchema(schema string) {

	// MustExec panics on error, so just call it
	// bf.DB.MustExec(schema)
	if bf.log {
		fmt.Println(schema)
	}
	result, err := bf.db.Exec(schema)
	if err != nil {
		fmt.Println("err:", err)
	}
	ra, err := result.RowsAffected()
	if err != nil {
		fmt.Println("err:", err)
	} else {
		if bf.log {
			fmt.Printf("%d rows affected.\n", ra)
		}
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

	queryString = bf.db.Rebind(queryString)
	row := bf.db.QueryRow(queryString, qParams)
	return row
}

// ExecuteQuery processes the multi-row query contained in queryString
// against the connected DB using sql/database.
func (bf *BaseFlavor) ExecuteQuery(queryString string, qParams ...interface{}) (*sql.Rows, error) {

	queryString = bf.db.Rebind(queryString)
	rows, err := bf.db.Query(queryString, qParams)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// ExecuteQueryRowx processes the single-row query contained in queryString
// against the connected DB using sqlx.
func (bf *BaseFlavor) ExecuteQueryRowx(queryString string, qParams ...interface{}) *sqlx.Row {

	queryString = bf.db.Rebind(queryString)
	row := bf.db.QueryRowx(queryString, qParams)
	return row
}

// ExecuteQueryx processes the multi-row query contained in queryString
// against the connected DB using sqlx.
func (bf *BaseFlavor) ExecuteQueryx(queryString string, qParams ...interface{}) (*sqlx.Rows, error) {

	queryString = bf.db.Rebind(queryString)
	rows, err := bf.db.Queryx(queryString, qParams)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// Get reads a single row into the dst interface.
// This calls sqlx.Get(...)
func (bf *BaseFlavor) Get(dst interface{}, queryString string, args ...interface{}) error {

	queryString = bf.db.Rebind(queryString)
	err := bf.db.Get(dst, queryString, args...)
	if err != nil {
		return err
	}
	return nil
}

// Select reads some rows into the dst interface.
// This calls sqlx.Select(...)
func (bf *BaseFlavor) Select(dst interface{}, queryString string, args ...interface{}) error {

	queryString = bf.db.Rebind(queryString)
	err := bf.db.Select(dst, queryString, args...)
	if err != nil {
		return err
	}
	return nil
}

// Exec runs the queryString against the connected db
func (bf *BaseFlavor) Exec(queryString string, args ...interface{}) (sql.Result, error) {

	queryString = bf.db.Rebind(queryString)
	result, err := bf.db.Exec(queryString, args...)
	if err != nil {
		return nil, err
	}
	return result, nil
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

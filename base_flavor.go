package sqac

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	// "reflect"
	"strings"
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

	// postgres, sqlite, mariadb, db2, hana etc.
	GetDBDriverName() string

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
	CreateIndex(index IndexInfo) error
	DropIndex(in string) error
	ExistsIndex(tn string, in string) bool

	// sn=sequenceName, start=start-value
	CreateSequence(sn string, start int) error
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

	ExecuteQueryRow(queryString string, qParams ...interface{}) *sql.Row
	ExecuteQuery(queryString string, qParams ...interface{}) (*sql.Rows, error)
	ExecuteQueryRowx(queryString string, qParams ...interface{}) *sqlx.Row
	ExecuteQueryx(queryString string, qParams ...interface{}) (*sqlx.Rows, error)
}

// BaseFlavor is a supporting struct for interface PublicDB
type BaseFlavor struct {
	DB  *sqlx.DB
	Log bool
	PublicDB
}

// GetDBDriverName returns the name of the driver associcated
// with the currently connected database.
func (bf *BaseFlavor) GetDBDriverName() string {

	return "footle"
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
func (bf *BaseFlavor) DestructiveResetTables(i ...interface{}) error {

	return nil
}

// ExistsTable checks the currently connected database and
// returns true if the named table is found to exist.
func (bf *BaseFlavor) ExistsTable(tn string) bool {

	return false
}

// ExistsColumn checks the currently connected database and
// returns true if the named table-column is found to exist.
// this checks the column name only, not the column data-type
// or properties.
func (bf *BaseFlavor) ExistsColumn(tn string, cn string) bool {

	return true
}

// CreateIndex creates the index contained in the incoming
// IndexInfo structure.  indexes will be created as non-unique
// by default, and in multi-field situations, the fields will
// added to the index in the order they are contained in the
// IndexInfo.[]IndexFields slice.
func (bf *BaseFlavor) CreateIndex(indexName string, index IndexInfo) error {

	// CREATE INDEX idx_material_num_int_example ON `equipment`(material_num, int_example)
	fList := ""
	indexSchema := ""

	if len(index.IndexFields) == 1 {
		fList = index.IndexFields[0]
		indexName = "idx_" + fList
	} else {
		for _, f := range index.IndexFields {
			fList = fmt.Sprintf("%s %s,", fList, f)
		}
		fList = strings.TrimSuffix(fList, ",")
	}

	if !index.Unique {
		indexSchema = fmt.Sprintf("CREATE INDEX %s ON %s (%s)", indexName, index.TableName, fList)
	} else {
		indexSchema = fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s)", indexName, index.TableName, fList)
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

	return false
}

// CreateSequence may be used to create a new sequence on the
// currently connected database.
func (bf *BaseFlavor) CreateSequence(sn string, start int) error {

	return nil
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
	if bf.Log {
		fmt.Println(schema)
	}
	result, err := bf.DB.Exec(schema)
	if err != nil {
		fmt.Println("err:", err)
	}
	ra, err := result.RowsAffected()
	if err != nil {
		fmt.Println("err:", err)
	} else {
		if bf.Log {
			fmt.Printf("%d rows affected.\n", ra)
		}
	}
}

// ProcessSchemaList processes the schemas contained in sList
// in the order in which they were provided.  Schemas are
// executed against the connected DB.
func (bf *BaseFlavor) ProcessSchemaList(sList []string) error {

	// bf.DB.MustExec(query string, args ...interface{})
	return nil
}

//===============================================================================
// SQL Query Processing
//===============================================================================

// ExecuteQueryRow processes the single-row query contained in queryString
// against the connected DB using sql/database.
func (bf *BaseFlavor) ExecuteQueryRow(queryString string, qParams ...interface{}) *sql.Row {

	queryString = bf.DB.Rebind(queryString)
	row := bf.DB.QueryRow(queryString, qParams)
	return row
}

// ExecuteQuery processes the multi-row query contained in queryString
// against the connected DB using sql/database.
func (bf *BaseFlavor) ExecuteQuery(queryString string, qParams ...interface{}) (*sql.Rows, error) {

	queryString = bf.DB.Rebind(queryString)
	rows, err := bf.DB.Query(queryString, qParams)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// ExecuteQueryRowx processes the single-row query contained in queryString
// against the connected DB using sqlx.
func (bf *BaseFlavor) ExecuteQueryRowx(queryString string, qParams ...interface{}) *sqlx.Row {

	queryString = bf.DB.Rebind(queryString)
	row := bf.DB.QueryRowx(queryString, qParams)
	return row
}

// ExecuteQueryx processes the multi-row query contained in queryString
// against the connected DB using sqlx.
func (bf *BaseFlavor) ExecuteQueryx(queryString string, qParams ...interface{}) (*sqlx.Rows, error) {

	queryString = bf.DB.Rebind(queryString)
	rows, err := bf.DB.Queryx(queryString, qParams)
	if err != nil {
		return nil, err
	}
	return rows, nil
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

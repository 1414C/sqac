package sqac_test

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/1414C/sqac"
	"github.com/1414C/sqac/common"
	_ "github.com/SAP/go-hdb/driver"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	Handle sqac.PublicDB
)

// ============================================================================================================================
// GetEntities test artifacts
// ============================================================================================================================
type DepotGetEntities2 struct {
	DepotNum            int       `json:"depot_num" db:"depot_num" sqac:"primary_key:inc;start:90000000"`
	DepotBay            int       `json:"depot_bay" db:"depot_bay" sqac:"primary_key:"`
	CreateDate          time.Time `json:"create_date" db:"create_date" sqac:"nullable:false;default:now();index:nonUnique"`
	Region              string    `json:"region" db:"region" sqac:"nullable:false;default:YYC"`
	Province            string    `json:"province" db:"province" sqac:"nullable:false;default:AB"`
	Country             string    `json:"country" db:"country" sqac:"nullable:true;default:CA"`
	NewColumn1          string    `json:"new_column1" db:"new_column1" sqac:"nullable:false"`
	NewColumn2          int64     `json:"new_column2" db:"new_column2" sqac:"nullable:false"`
	NewColumn3          float64   `json:"new_column3" db:"new_column3" sqac:"nullable:false;default:0.0"`
	IntDefaultZero      int       `json:"int_default_zero" db:"int_default_zero" sqac:"nullable:false;default:0"`
	IntDefault42        int       `json:"int_default42" db:"int_default42" sqac:"nullable:false;default:42"`
	IntZeroValNoDefault int       `json:"int_zero_val_no_default" db:"int_zero_val_no_default" sqac:"nullable:false"`
	NonPersistentColumn string    `json:"non_persistent_column" db:"non_persistent_column" sqac:"-"`
}

type DepotGetEntitiesTab struct {
	ents []DepotGetEntities2
	sqac.GetEnt
}

// Implement the sqac.GetEnt{} interface for DepotGetEntities2
func (dget *DepotGetEntitiesTab) Exec(sqh sqac.PublicDB) error {

	selQuery := "SELECT * FROM depotgetentities2;"

	// read the table rows
	rows, err := sqh.ExecuteQueryx(selQuery)
	if err != nil {
		log.Printf("GetEntities for table depotgetentities2 returned error: %v\n", err.Error())
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var ent DepotGetEntities2
		err = rows.StructScan(&ent)
		if err != nil {
			fmt.Printf("error reading rows: %v\n", err)
			return err
		}
		dget.ents = append(dget.ents, ent)
	}
	if sqh.IsLog() {
		fmt.Println("DepotGetEntitiesTab.Exec() got:", dget.ents)
	}
	return nil
}

//============================================================================================================================
// GetEntities test artifacts end
//============================================================================================================================

// TestMain - the go test entry point
func TestMain(m *testing.M) {

	// parse flags
	dbFlag := flag.String("db", "postgres", "db to connect to")
	logFlag := flag.Bool("l", false, "activate sqac detail logging to stdout")
	dbLogFlag := flag.Bool("dbl", false, "activate DDL/DML logging to stdout)")
	flag.Parse()

	var cs string
	switch *dbFlag {
	case "postgres":
		cs = "host=127.0.0.1 user=godev dbname=sqactst sslmode=disable password=gogogo123"
	case "mysql":
		cs = "godev:gogogo123@tcp(localhost:3306)/sqactst?charset=utf8&parseTime=True&loc=Local"
	case "sqlite":
		cs = "testdb.sqlite"
	case "mssql":
		cs = "sqlserver://SA:gogogo123@localhost:1433?database=sqlx"
	case "db2":
		cs = ""
	case "hdb":
		cs = "hdb://SYSTEM:WTBHana1!@192.168.112.35:39017"
		//cs = "hdb://hxeadm:HXEHana1@192.168.112.35:39017"
		//cs = "hdb://godev:gogogo123@your.hanadb.com:30015"
	default:
		cs = ""
	}
	Handle = sqac.Create(*dbFlag, *logFlag, *dbLogFlag, cs)

	// run the tests
	code := m.Run()

	os.Exit(code)
}

// TestGetDBDriverName
//
// Check that a driver name is returned
func TestGetDBDriverName(t *testing.T) {
	driverName := Handle.GetDBDriverName()
	if driverName == "" {
		t.Errorf("unable to determine db driver name")
	}
	if Handle.IsLog() {
		fmt.Println("db driver name:", driverName)
	}
}

// TestGetDBName
//
// Check that a db name is known
func TestGetDBName(t *testing.T) {
	dbName := Handle.GetDBName()
	if dbName == "" {
		t.Errorf("unable to determine db name")
	}
	if Handle.IsLog() {
		fmt.Println("db name:", dbName)
	}
}

// TestExistsTableNegative

// Test for non-existent table 'Footle'

func TestExistsTableNegative(t *testing.T) {

	type Footle struct {
		KeyNum      int       `db:"key_num" sqac:"primary_key:inc"`
		CreateDate  time.Time `db:"create_date" sqac:"nullable:false;default:now();"`
		Description string    `db:"description" sqac:"nullable:false;default:"`
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(Footle{})

	// expect that table depot does not exist
	if Handle.ExistsTable(tn) {
		t.Errorf("table %s was found when it was not expected", tn)
	}
}

// TestCreateTableBasic
//
// Create table depot via CreateTables(i ...interface{})
// Verify table creation via ExistsTable(tn string)
// Perform negative validation be checking for non-existent
//
//	table "abcdefg" via ExistsTable(tn string)
func TestCreateTableBasic(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" sqac:"primary_key:inc"`
		CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now();"`
		Region     string    `db:"region" sqac:"nullable:false;default:YYC"`
		Province   string    `db:"province" sqac:"nullable:false;default:AB"`
		Country    string    `db:"country" sqac:"nullable:false;default:CA"`
	}

	err := Handle.CreateTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(Depot{})

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s was not created", tn)
	}
}

// TestDropTablesBasic
//
// Drop table depot via DropTables(i ...interface{})
func TestDropTablesBasic(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" sqac:"primary_key:inc"`
		CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now();"`
		Region     string    `db:"region" sqac:"nullable:false;default:YYC"`
		Province   string    `db:"province" sqac:"nullable:false;default:AB"`
		Country    string    `db:"country" sqac:"nullable:false;default:CA"`
	}

	err := Handle.DropTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(Depot{})

	// expect that table depot has been dropped
	if Handle.ExistsTable(tn) {
		t.Errorf("table %s was not dropped", tn)
	}

	err = Handle.DropTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

func TestCreateTableCompoundKey(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" sqac:"primary_key:inc"`
		DepotBay   int       `db:"depot_bay" sqac:"primary_key:"`
		CreateDate time.Time `db:"create_date" sqac:"default:now()"`
		ExpiryDate time.Time `db:"expiry_date" sqac:"default:eot()"`
		Region     string    `db:"region" sqac:"nullable:true;default:YYC"`
		Province   string    `db:"province" sqac:"nullable:false;default:AB"`
		Country    string    `db:"country" sqac:"nullable:false;default:CA"`
	}

	err := Handle.CreateTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(Depot{})

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s was not created", tn)
	}

	// create a new record via the CRUD Create call
	var depot = Depot{
		DepotBay: 1,
		Region:   "YYC",
		Province: "AB",
		Country:  "CA",
	}

	err = Handle.Create(&depot)
	if err != nil {
		t.Errorf(err.Error())
	}

	depot.DepotBay = 1
	depot.Region = "YEG"
	err = Handle.Create(&depot)
	if err != nil {
		t.Errorf(err.Error())
	}
	// time.Sleep(15 * time.Second)
	if Handle.IsLog() {
		fmt.Printf("INSERTING: %v\n", depot)
		fmt.Printf("TEST GOT: %v\n", depot)
	}

	err = Handle.DropTables(Depot{})
	if err != nil {
		t.Errorf("TestCreateTableCompoundKey: %s", err.Error())
	}
}

// TestCreateTableNonIncKey
//
// Create table depot via CreateTables(i ...interface{})
// Verify table creation via ExistsTable(tn string)
// Perform negative validation be checking for non-existent
//
//	table "abcdefg" via ExistsTable(tn string)
func TestCreateTableNonIncKey(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" sqac:"primary_key:"` // non-incrementing key
		CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now();"`
		Region     string    `db:"region" sqac:"nullable:false;default:YYC"`
		Province   string    `db:"province" sqac:"nullable:false;default:AB"`
		Country    string    `db:"country" sqac:"nullable:false;default:CA"`
	}

	// ensure table does not exist
	err := Handle.DropTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	err = Handle.CreateTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(Depot{})

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s was not created", tn)
	}

	err = Handle.DropTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCreateTableNoKey
//
// Create table depot via CreateTables(i ...interface{})
// Verify table creation via ExistsTable(tn string)
// Perform negative validation be checking for non-existent
//
//	table "abcdefg" via ExistsTable(tn string)
func TestCreateTableNoKey(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" sqac:"nullable:false"` // not a key
		CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now();"`
		Region     string    `db:"region" sqac:"nullable:false;default:YYC"`
		Province   string    `db:"province" sqac:"nullable:false;default:AB"`
		Country    string    `db:"country" sqac:"nullable:false;default:CA"`
	}

	// ensure table does not exist
	err := Handle.DropTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	err = Handle.CreateTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(Depot{})

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s was not created", tn)
	}

	err = Handle.DropTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCreateTableWithAlterSequence
//
// Create table depot via CreateTables(i ...interface{})
// Verify table creation via ExistsTable(tn string)
// Perform negative validation be checking for non-existent
//
//	table "abcdefg" via ExistsTable(tn string)
func TestCreateTableWithAlterSequence(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now()"`
		Region     string    `db:"region" sqac:"nullable:false;default:YYC"`
		Province   string    `db:"province" sqac:"nullable:false;default:AB"`
		Country    string    `db:"country" sqac:"nullable:false;default:CA"`
	}

	err := Handle.CreateTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the  table creation logic
	tn := common.GetTableName(Depot{})

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s was not created", tn)
	}

	// check the next value of the auto-increment, sequence or
	// identity field depending on db-system.
	var hdbName string
	var seq int
	if Handle.GetDBDriverName() != "hdb" {
		seq, err = Handle.GetNextSequenceValue(tn)
	} else {
		hdbName = fmt.Sprintf("SEQ_%s_%s", strings.ToUpper(tn), "DEPOT_NUM")
		seq, err = Handle.GetNextSequenceValue(hdbName)
	}
	if err != nil {
		t.Errorf(err.Error())
	}

	if seq != 90000000 {
		t.Errorf("expected value of 90000000, got %d", seq)
	}

	// Drop the depot table
	err = Handle.DropTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCreateTablesWithInclude
//
// Create table equipment via CreateTables(i ...interface{})
// and verify that flat structs can be included in the
// table creation.
func TestCreateTablesWithInclude(t *testing.T) {

	// type Triplet struct {
	// 	TripOne   string `db:"trip_one" sqac:"nullable:false"`
	// 	TripTwo   int64  `db:"trip_two" sqac:"nullable:false;default:0"`
	// 	Tripthree string `db:"trip_three" sqac:"nullable:false"`
	// }

	// type Equipment struct {
	// 	EquipmentNum   int64     `db:"equipment_num" sqac:"primary_key:inc;start:55550000"`
	// 	ValidFrom      time.Time `db:"valid_from" sqac:"primary_key;nullable:false;default:now()"`
	// 	ValidTo        time.Time `db:"valid_to" sqac:"primary_key;nullable:false;default:make_timestamptz(9999, 12, 31, 23, 59, 59.9)"`
	// 	CreatedAt      time.Time `db:"created_at" sqac:"nullable:false;default:now()"`
	// 	InspectionAt   time.Time `db:"inspection_at" sqac:"nullable:true"`
	// 	MaterialNum    int       `db:"material_num" sqac:"index:idx_material_num_serial_num"`
	// 	Description    string    `db:"description" sqac:"sqac:nullable:false"`
	// 	SerialNum      string    `db:"serial_num" sqac:"index:idx_material_num_serial_num"`
	// 	IntExample     int       `db:"int_example" sqac:"nullable:false;default:0"`
	// 	Int64Example   int64     `db:"int64_example" sqac:"nullable:false;default:0"`
	// 	Int32Example   int32     `db:"int32_example" sqac:"nullable:false;default:0"`
	// 	Int16Example   int16     `db:"int16_example" sqac:"nullable:false;default:0"`
	// 	Int8Example    int8      `db:"int8_example" sqac:"nullable:false;default:0"`
	// 	UIntExample    uint      `db:"uint_example" sqac:"nullable:false;default:0"`
	// 	UInt64Example  uint64    `db:"uint64_example" sqac:"nullable:false;default:0"`
	// 	UInt32Example  uint32    `db:"uint32_example" sqac:"nullable:false;default:0"`
	// 	UInt16Example  uint16    `db:"uint16_example" sqac:"nullable:false;default:0"`
	// 	UInt8Example   uint8     `db:"uint8_example" sqac:"nullable:false;default:0"`
	// 	Float32Example float32   `db:"float32_example" sqac:"nullable:false;default:0.0"`
	// 	Float64Example float64   `db:"float64_example" sqac:"nullable:false;default:0.0"`
	// 	BoolExample    bool      `db:"bool_example" sqac:"nullable:false;default:false"`
	// 	RuneExample    rune      `db:"rune_example" sqac:"nullable:true"`
	// 	ByteExample    byte      `db:"byte_example" sqac:"nullable:true"`
	// 	Triplet
	// }

	err := Handle.CreateTables(Equipment{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(Equipment{})

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s was not created", tn)
	}

	// drop the equipment table
	err = Handle.DropTables(Equipment{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCreateUniqueColumnConstraintFromModel
//
// Create table depot via CreateTables(i ...interface{})
// Create a single-field non-unique index based on model
// attributes.
func TestCreateUniqueColumnConstraintFromModel(t *testing.T) {

	type DepotConstraint struct {
		DepotNum   int       `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now();index:non-unique"`
		Region     string    `db:"region" sqac:"nullable:false;default:YYC"`
		Province   string    `db:"province" sqac:"nullable:false;default:AB"`
		Country    string    `db:"country" sqac:"nullable:false;default:CA"`
		LotID      uint      `db:"lot_id" sqac:"nullable:false;constraint:unique"`
	}

	err := Handle.DropTables(DepotConstraint{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	err = Handle.CreateTables(DepotConstraint{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(DepotConstraint{})

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s was not created", tn)
	}

	// if !Handle.ExistsIndex(tn, "idx_depot_create_date") {
	// 	t.Errorf("expected unique index idx_depot_create_date - got: ")
	// }

	// drop the depotconstraint table
	err = Handle.DropTables(DepotConstraint{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestExistsIndexNegative
//
// Check to see if index:
// idx_province_country exists on
// table depot.
func TestExistsIndexNegative(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now()"`
		Region     string    `db:"region" sqac:"nullable:false;default:YYC"`
		Province   string    `db:"province" sqac:"nullable:false;default:AB"`
		Country    string    `db:"country" sqac:"nullable:false;default:CA"`
		NewColumn1 string    `db:"new_column1" sqac:"nullable:false"`
		NewColumn2 int64     `db:"new_column2" sqac:"nullable:false;default:0"`
		NewColumn3 float64   `db:"new_column3" sqac:"nullable:false;default:0.0"`
	}

	// drop table depot
	// err := Handle.DropTables(Depot{})
	// if err != nil {
	// 	t.Errorf("%s", err.Error())
	// }

	// ensure that table depot exists
	err := Handle.AlterTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(Depot{})

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	if Handle.ExistsIndex(tn, "idx_depot_province_country") {
		t.Errorf("index %s was found on table %s, but was not expected", "idx_depot_province_country", tn)
	}

	// drop the depot table
	err = Handle.DropTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCreateSingleUniqueIndexFromModel
//
// Create table depot via CreateTables(i ...interface{})
// Create a single-field unique index based on model
// attributes.
func TestCreateSingleUniqueIndexFromModel(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now();index:unique"`
		Region     string    `db:"region" sqac:"nullable:false;default:YYC"`
		Province   string    `db:"province" sqac:"nullable:false;default:AB"`
		Country    string    `db:"country" sqac:"nullable:false;default:CA"`
		NewColumn1 string    `db:"new_column1" sqac:"nullable:false"`
		NewColumn2 int64     `db:"new_column2" sqac:"nullable:false;default:0"`
		NewColumn3 float64   `db:"new_column3" sqac:"nullable:false;default:0.0"`
	}

	err := Handle.CreateTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(Depot{})

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s was not created", tn)
	}

	if !Handle.ExistsIndex(tn, "idx_depot_create_date") {
		t.Errorf("expected unique index idx_depot_create_date - got: ")
	}

	// drop the depot table
	err = Handle.DropTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCreateSingleNonUniqueIndexFromModel
//
// Create table depot via CreateTables(i ...interface{})
// Create a single-field non-unique index based on model
// attributes.
func TestCreateSingleNonUniqueIndexFromModel(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now();index:non-unique"`
		Region     string    `db:"region" sqac:"nullable:false;default:YYC"`
		Province   string    `db:"province" sqac:"nullable:false;default:AB"`
		Country    string    `db:"country" sqac:"nullable:false;default:CA"`
		NewColumn1 string    `db:"new_column1" sqac:"nullable:false"`
		NewColumn2 int64     `db:"new_column2" sqac:"nullable:false;default:0"`
		NewColumn3 float64   `db:"new_column3" sqac:"nullable:false;default:0.0"`
	}

	err := Handle.CreateTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(Depot{})

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s was not created", tn)
	}

	if !Handle.ExistsIndex(tn, "idx_depot_create_date") {
		t.Errorf("expected unique index idx_depot_create_date - got: ")
	}

	// drop the depot table
	err = Handle.DropTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCreateSimpleCompositeIndex
//
// Create table depot via CreateTables(i ...interface{})
// Create an index not based on model attributes, but
// based on a constructed sqac.IndexInfo struct.
func TestCreateSimpleCompositeIndex(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now()"`
		Region     string    `db:"region" sqac:"nullable:false;default:YYC"`
		Province   string    `db:"province" sqac:"nullable:false;default:AB"`
		Country    string    `db:"country" sqac:"nullable:false;default:CA"`
		NewColumn1 string    `db:"new_column1" sqac:"nullable:false"`
		NewColumn2 int64     `db:"new_column2" sqac:"nullable:false;default:0"`
		NewColumn3 float64   `db:"new_column3" sqac:"nullable:false;default:0.0"`
	}

	err := Handle.CreateTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(Depot{})

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s was not created", tn)
	}

	var indexInfo sqac.IndexInfo
	indexInfo.TableName = tn
	indexInfo.Unique = false
	indexInfo.IndexFields = []string{"province", "country"}
	err = Handle.CreateIndex("idx_depot_province_country", indexInfo)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestExistsIndexPositive
//
// Check to see if index:
// idx_province_country exists on
// table depot.
func TestExistsIndexPositive(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now();index:unique"`
		Region     string    `db:"region" sqac:"nullable:false;default:YYC"`
		Province   string    `db:"province" sqac:"nullable:false;default:AB"`
		Country    string    `db:"country" sqac:"nullable:false;default:CA"`
		NewColumn1 string    `db:"new_column1" sqac:"nullable:false"`
		NewColumn2 int64     `db:"new_column2" sqac:"nullable:false;default:0"`
		NewColumn3 float64   `db:"new_column3" sqac:"nullable:false;default:0.0"`
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(Depot{})

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	if !Handle.ExistsIndex(tn, "idx_depot_province_country") {
		t.Errorf("index %s was not found on table %s", "idx_depot_province_country", tn)
	}
}

// TestDropIndex
//
// Drop index "idx_province_country" on table depot via
// DropIndex(in string)
// Call ExistsIndex(tn string, in string) bool to
// verify that the index has been dropped.
func TestDropIndex(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now();index:unique"`
		Region     string    `db:"region" sqac:"nullable:false;default:YYC"`
		Province   string    `db:"province" sqac:"nullable:false;default:AB"`
		Country    string    `db:"country" sqac:"nullable:false;default:CA"`
		NewColumn1 string    `db:"new_column1" sqac:"nullable:false"`
		NewColumn2 int64     `db:"new_column2" sqac:"nullable:false;default:0"`
		NewColumn3 float64   `db:"new_column3" sqac:"nullable:false;default:0.0"`
	}

	// ensure table depot exists
	err := Handle.AlterTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(Depot{})

	if !Handle.ExistsIndex(tn, "idx_depot_province_country") {
		t.Errorf("index %s was not found on table %s", "idx_depot_province_country", tn)
	}

	err = Handle.DropIndex(tn, "idx_depot_province_country")
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	if Handle.ExistsIndex(tn, "idx_depot_province_country") {
		t.Errorf("drop of index %s did not succeed on table %s", "idx_province_country", tn)
	}

	err = Handle.DropTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCreateCompositeIndexFromModel
//
// Create table depot via CreateTables(i ...interface{})
// Create an index not based on model attributes, but
// based on a constructed sqac.IndexInfo struct.
func TestCreateCompositeIndexFromModel(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now()"`
		Region     string    `db:"region" sqac:"nullable:false;default:YYC"`
		Province   string    `db:"province" sqac:"nullable:false;default:AB"`
		Country    string    `db:"country" sqac:"nullable:false;default:CA"`
		NewColumn1 string    `db:"new_column1" sqac:"nullable:false;index:idx_depot_new_column1_new_column2"`
		NewColumn2 int64     `db:"new_column2" sqac:"nullable:false;default:0;index:idx_depot_new_column1_new_column2"`
		NewColumn3 float64   `db:"new_column3" sqac:"nullable:false;default:0.0"`
	}

	err := Handle.CreateTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(Depot{})

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s was not created", tn)
	}

	if !Handle.ExistsIndex(tn, "idx_depot_new_column1_new_column2") {
		t.Errorf("index %s was not found on table %s", "idx_depot_new_column1_new_column2", tn)
	}

	err = Handle.DropTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestExistsColumn
//
// Verify that table depot exists on the database
// by calling AlterTables(Depot{})
// Verify that column 'region' exists in the table
// via ExistsColumn(tn string, cn string) bool.
// Verify that column 'footle' does not exist in the
// table via ExistsColumn(tn string, cn string) bool.
func TestExistsColumn(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now();index:unique"`
		Region     string    `db:"region" sqac:"nullable:false;default:YYC;index:non-unique"`
		Province   string    `db:"province" sqac:"nullable:false;default:AB"`
		Country    string    `db:"country" sqac:"nullable:true;default:CA"`
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(Depot{})

	// ensure table exists in db - create via alter
	err := Handle.AlterTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// check for column 'region'
	if !Handle.ExistsColumn(tn, "region") {
		t.Errorf("column %s should have been found in table %s", "region", tn)
	}

	// check for column 'footle'
	if Handle.ExistsColumn(tn, "footle") {
		t.Errorf("column %s should not have been found in table %s", "footle", tn)
	}
}

// TestAlterTables
//
// Alter table depot via AlterTables(i ...interface{})
// Add three columns:
//   - NewColumn1 string    `db:"new_column1" sqac:"nullable:false"`
//   - NewColumn2 int64     `db:"new_column2" sqac:"nullable:false;default:0"`
//   - NewColumn3 float64   `db:"new_column3" sqac:"nullable:false;default:0.0"`
func TestAlterTables(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now();index:unique"`
		Region     string    `db:"region" sqac:"nullable:false;default:YYC;index:non-unique"`
		Province   string    `db:"province" sqac:"nullable:false;default:AB"`
		Country    string    `db:"country" sqac:"nullable:false;default:CA"`
		NewColumn1 string    `db:"new_column1" sqac:"nullable:false;default:nc1_default;index:non-unique"`
		NewColumn2 int64     `db:"new_column2" sqac:"nullable:false;default:0;index:idx_new_column2_new_column3"`
		NewColumn3 float64   `db:"new_column3" sqac:"nullable:false;default:0.0;index:idx_new_column2_new_column3"`
	}

	err := Handle.AlterTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(Depot{})

	if !Handle.ExistsColumn("depot", "new_column1") {
		t.Errorf("new_column1 was expected to exist but does not ")
	}

	if !Handle.ExistsColumn("depot", "new_column2") {
		t.Errorf("new_column2 was expected to exist but does not ")
	}

	if !Handle.ExistsColumn("depot", "new_column3") {
		t.Errorf("new_column3 was expected to exist but does not ")
	}

	r := Handle.ExistsIndex(tn, "idx_new_column2_new_column3")
	if r == false {
		t.Errorf("index idx_new_column2_new_column3 was not not found following alter table call on table %s", tn)
	}

	r = Handle.ExistsIndex(tn, "idx_depot_region")
	if r == false {
		t.Errorf("index idx_depot_region was not not found following alter table call on table %s", tn)
	}

}

// TestDestructiveResetTables
//
// Verify that tables depot and equipment exist on the db
// by calling AlterTables(i ...interface{})
// Drop and recreate tables depot and equipment on the db
// by calling DestructiveResetTables(i ...interface{})
// Select * from db tables depot and equipment to verify
// that they exist and contain no records.
func TestDestructiveResetTables(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now();index:unique"`
		Region     string    `db:"region" sqac:"nullable:false;default:YYC"`
		Province   string    `db:"province" sqac:"nullable:false;default:AB"`
		Country    string    `db:"country" sqac:"nullable:false;default:CA"`
		Active     bool      `db:"active" sqac:"nullable:false;default:true"`
	}

	// determine the table names
	tns := make([]string, 0)
	tn := common.GetTableName(Depot{})
	tns = append(tns, tn)

	tn = common.GetTableName(Equipment{})
	tns = append(tns, tn)

	// ensure tables exist in db
	err := Handle.AlterTables(Depot{}, Equipment{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that tables depot and equipment exist
	for _, n := range tns {
		if !Handle.ExistsTable(n) {
			t.Errorf("table %s does not exist", n)
		}
	}

	// drop both tables via DestructiveReset
	err = Handle.DestructiveResetTables(Depot{}, Equipment{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that tables depot and equipment have been recreated
	for _, n := range tns {
		if !Handle.ExistsTable(n) {
			t.Errorf("table %s does not exist", n)
		}
	}

	d := []Depot{}
	err = Handle.Select(&d, "SELECT * FROM depot;")
	if err != nil {
		t.Errorf("error reading from table depot - got %s\n", err.Error())
	}
	if len(d) > 0 {
		t.Errorf("table depot contained records after attempted DestructiveReset - got %v\n", d)
	}

	e := []Equipment{}
	err = Handle.Select(&e, "SELECT * FROM equipment;")
	if err != nil {
		t.Errorf("error reading from table equipment - got %s\n", err.Error())
	}
	if len(e) > 0 {
		t.Errorf("table equipment contained records after attempted DestructiveReset - got %v\n", d)
	}

	err = Handle.DropTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestQueryOps
//
// This test is designed to illustrate the handling of database reads when dealing with db fields that
// contain null values.  Create db table depot based on an updated Depot struct containing a number of
// nullable and non-defaulted fields.
// Insert a new record containing null-values into db table depot.
// Declare struct DepotN{} as a parallel structure to Depot{} making use of sql.Null<type> fields in
// place of the gotypes for the nullable fields.
// Note that DepotN{} also contains one *string pointer type instead of sql.NullString in order
// to demonstrate a different way to handle the situation.
// Read all the records (1) from db table depot assigning them to a slice declared as type DepotN.
// Iterate over the record(s) contained in the result set and take note of the manner in which the
// nullable field values are accessed / converted from nil values to their base-type's default value.
// In this example, the Valid bool flag in the nullable field is not checked, as it is typically(?)
// okay to simply ask for base-type default through .Sting, .Int64, .Float64 or .Bool.
func TestQueryOps(t *testing.T) {

	type QOps struct {
		OpNum       int       `db:"op_num" sqac:"primary_key:inc;start:70000000"`
		CreateDate  time.Time `db:"create_date" sqac:"nullable:false;default:now();index:unique"`
		Description string    `db:"description" sqac:"nullable:false;default:initial"`
	}

	// create table qops
	err := Handle.CreateTables(QOps{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table names as per the table creation logic
	tn := common.GetTableName(QOps{})

	// ensure table qops exists in db
	if !Handle.ExistsTable(tn) {
		t.Errorf("%s", err.Error())
	}

	// insert a new record containing null-values into db-table qops
	// sql.NullBool, sql.NullFloat64, sql.NullInt64, sql.NullString
	insQuery := ""
	switch Handle.GetDBDriverName() {
	case "postgres", "mysql":
		insQuery = "INSERT INTO qops (op_num, create_date, description) VALUES (DEFAULT, DEFAULT, 'test_value');"
	case "sqlite3":
		insQuery = "INSERT INTO qops (description) VALUES ('test_value');"
	case "mssql":
		// INSERT INTO Persons(name, age) values('Bob', 20)
		insQuery = "INSERT INTO qops (description) VALUES ('test_value');"
	case "hdb":
		var incKey int
		keyQuery := "SELECT SEQ_QOPS_OP_NUM.NEXTVAL FROM DUMMY;"
		err = Handle.ExecuteQueryRowx(keyQuery).Scan(&incKey)
		if err != nil {
			t.Errorf(err.Error())
		}
		insQuery = fmt.Sprintf("INSERT INTO qops (op_num, description) VALUES (%d, 'test_value');", incKey)
	default:
		insQuery = "INSERT INTO qops (op_num, create_date, description) VALUES (DEFAULT, DEFAULT, 'test_value');"
	}
	_, err = Handle.Exec(insQuery)
	if err != nil {
		t.Errorf("error inserting into table qops - got %s", err.Error())
	}

	// read all records from db-table qops into a QOps struct using Select
	qo := []QOps{}
	err = Handle.Select(&qo, "SELECT * FROM qops;")
	if err != nil {
		t.Errorf("error reading from table qops - got %s", err.Error())
	}

	if len(qo) == 0 {
		t.Errorf("expected 1 record in table qops - got 0")
	}

	if Handle.IsLog() {
		fmt.Println("sqlx.Select got:", qo)
		for _, v := range qo {
			fmt.Println("got op_num:", v.OpNum)
			fmt.Println("got create_date:", v.CreateDate)
			fmt.Println("got description:", v.Description)
		}
	}

	// read all records from db-table qops into a QOps struct where description == 'test_value' using Select
	qo = []QOps{}
	err = Handle.Select(&qo, "SELECT * FROM qops WHERE description = ?;", "test_value")
	if err != nil {
		t.Errorf("error reading from table qops - got %s", err.Error())
	}

	if len(qo) == 0 {
		t.Errorf("expected 1 record in table qops - got 0")
	}

	if Handle.IsLog() {
		fmt.Println("sqlx.Select got:", qo)
		for _, v := range qo {
			fmt.Println("got op_num:", v.OpNum)
			fmt.Println("got create_date:", v.CreateDate)
			fmt.Println("got description:", v.Description)
		}
	}

	// read all records from db-table qops into a QOps struct using sql.Query
	qos := QOps{}
	rows, err := Handle.ExecuteQuery("SELECT * FROM qops;")
	if err != nil {
		t.Errorf("error reading from table qops  - got %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&qos.OpNum, &qos.CreateDate, &qos.Description)
		if err != nil {
			t.Errorf("error reading from table qops  - got %s", err.Error())
		}
	}

	if Handle.IsLog() {
		fmt.Println("sql.Query got:", qo)
		for _, v := range qo {
			fmt.Println("got op_num:", v.OpNum)
			fmt.Println("got create_date:", v.CreateDate)
			fmt.Println("got description:", v.Description)
		}
	}

	// read all records from db-table qops into a QOps struct using sql.Query with a parameter
	qos = QOps{}
	rows, err = Handle.ExecuteQuery("SELECT * FROM qops WHERE description = ?;", "test_value")
	if err != nil {
		t.Errorf("error reading from table qops  - got %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&qos.OpNum, &qos.CreateDate, &qos.Description)
		if err != nil {
			t.Errorf("error reading from table qops  - got %s", err.Error())
		}
	}

	if Handle.IsLog() {
		fmt.Println("sql.Query got:", qo)
		for _, v := range qo {
			fmt.Println("got op_num:", v.OpNum)
			fmt.Println("got create_date:", v.CreateDate)
			fmt.Println("got description:", v.Description)
		}
	}

	// read all records from db-table qops into a QOps struct using sqlx.Queryx
	qos = QOps{}
	sqlxRows, err := Handle.ExecuteQueryx("SELECT * FROM qops;")
	if err != nil {
		t.Errorf("error reading from table qops  - got %s", err.Error())
	}
	defer sqlxRows.Close()

	for sqlxRows.Next() {
		err = sqlxRows.Scan(&qos.OpNum, &qos.CreateDate, &qos.Description)
		if err != nil {
			t.Errorf("error reading from table qops  - got %s", err.Error())
		}
	}

	if Handle.IsLog() {
		fmt.Println("sql.Queryx got:", qo)
		for _, v := range qo {
			fmt.Println("got op_num:", v.OpNum)
			fmt.Println("got create_date:", v.CreateDate)
			fmt.Println("got description:", v.Description)
		}
	}

	// read all records from db-table qops into a QOps struct using sqlx.Queryx with a parameter
	qos = QOps{}
	sqlxRows, err = Handle.ExecuteQueryx("SELECT * FROM qops WHERE description = ?;", "test_value")
	if err != nil {
		t.Errorf("error reading from table qops  - got %s", err.Error())
	}
	defer sqlxRows.Close()

	for sqlxRows.Next() {
		err = sqlxRows.Scan(&qos.OpNum, &qos.CreateDate, &qos.Description)
		if err != nil {
			t.Errorf("error reading from table qops  - got %s", err.Error())
		}
	}

	if Handle.IsLog() {
		fmt.Println("sql.Queryx got:", qo)
		for _, v := range qo {
			fmt.Println("got op_num:", v.OpNum)
			fmt.Println("got create_date:", v.CreateDate)
			fmt.Println("got description:", v.Description)
		}
	}

}

// TestNullableValues
//
// This test is designed to illustrate the handling of database reads when dealing with db fields that
// contain null values.  Create db table depot based on an updated Depot struct containing a number of
// nullable and non-defaulted fields.
// Insert a new record containing null-values into db table depot.
// Declare struct DepotN{} as a parallel structure to Depot{} making use of sql.Null<type> fields in
// place of the gotypes for the nullable fields.
// Note that DepotN{} also contains one *string pointer type instead of sql.NullString in order
// to demonstrate a different way to handle the situation.
// Read all the records (1) from db table depot assigning them to a slice declared as type DepotN.
// Iterate over the record(s) contained in the result set and take note of the manner in which the
// nullable field values are accessed / converted from nil values to their base-type's default value.
// In this example, the Valid bool flag in the nullable field is not checked, as it is typically(?)
// okay to simply ask for base-type default through .Sting, .Int64, .Float64 or .Bool.
func TestNullableValues(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" sqac:"nullable:false;default:now();index:unique"`
		Region     string    `db:"region" sqac:"nullable:false;default:YYC"`
		MemOnly    string    `db:"mem_only" sqac:"-"`
		Province   string    `db:"province" sqac:"nullable:false;default:AB"`
		Country    string    `db:"country" sqac:"nullable:true;"`    // nullable
		NewColumn1 string    `db:"new_column1" sqac:"nullable:true"` // nullable
		NewColumn2 int64     `db:"new_column2" sqac:"nullable:true"` // nullable
		NewColumn3 float64   `db:"new_column3" sqac:"nullable:true"` // nullable
		Active     bool      `db:"active" sqac:"nullable:true"`      // nullable
	}

	// create table depot
	err := Handle.CreateTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table names as per the table creation logic
	tn := common.GetTableName(Depot{})

	// ensure table depot exists in db
	if !Handle.ExistsTable(tn) {
		t.Errorf("%s", err.Error())
	}

	// insert a new record containing null-values into db-table depot
	// sql.NullBool, sql.NullFloat64, sql.NullInt64, sql.NullString
	insQuery := ""
	switch Handle.GetDBDriverName() {
	case "postgres", "mysql":
		insQuery = "INSERT INTO depot (depot_num, region, province) VALUES (DEFAULT,'YVR','AB');"
	case "sqlite3":
		insQuery = "INSERT INTO depot (region, province) VALUES ('YVR','AB');"
	case "mssql":
		// INSERT INTO Persons(name, age) values('Bob', 20)
		insQuery = "INSERT INTO depot (region, province) VALUES ('YVR','AB');"
	case "hdb":
		var incKey int
		keyQuery := "SELECT SEQ_DEPOT_DEPOT_NUM.NEXTVAL FROM DUMMY;"
		err = Handle.ExecuteQueryRowx(keyQuery).Scan(&incKey)
		if err != nil {
			t.Errorf(err.Error())
		}
		insQuery = fmt.Sprintf("INSERT INTO depot (depot_num, region, province) VALUES (%d, 'YVR','AB');", incKey)
	default:
		insQuery = "INSERT INTO depot (depot_num, region, province) VALUES (DEFAULT, 'YVR','AB');"
	}
	_, err = Handle.Exec(insQuery)
	if err != nil {
		t.Errorf("error inserting into table depot - got %s", err.Error())
	}
	// if Handle.IsLog() {
	// 	ra, _ := result.RowsAffected()
	// 	fmt.Printf("%d rows affected.\n", ra)
	// }

	// deal with nullable fields via sql.Null<type> and one *string to
	// illustrate different ways of handling the db-nulls. A parallel
	// Depot struct is defined:
	type DepotN struct {
		DepotNum   int             `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		CreateDate time.Time       `db:"create_date" sqac:"nullable:false;default:now();index:unique"`
		Region     string          `db:"region" sqac:"nullable:false;default:YYC"`
		MemOnly    string          `db:"mem_only" sqac:"-"`
		Province   string          `db:"province" sqac:"nullable:false;default:AB"`
		Country    sql.NullString  `db:"country" sqac:"nullable:true;"`
		NewColumn1 *string         `db:"new_column1" sqac:"nullable:true"`
		NewColumn2 sql.NullInt64   `db:"new_column2" sqac:"nullable:true"`
		NewColumn3 sql.NullFloat64 `db:"new_column3" sqac:"nullable:true"`
		Active     sql.NullBool    `db:"active" sqac:"nullable:true"`
	}

	// read records from db-table depot into a DepotN struct
	dn := []DepotN{}
	err = Handle.Select(&dn, "SELECT * FROM depot;")
	if err != nil {
		t.Errorf("error reading from table depot - got %s", err.Error())
	}

	if Handle.IsLog() {
		fmt.Println("got:", dn)
		for i, v := range dn {

			fmt.Println("got depot_num:", v.DepotNum)
			fmt.Println("got create_date:", v.CreateDate)
			fmt.Println("got region:", v.Region)
			fmt.Println("got province:", v.Province)
			fmt.Println("got country:", v.Country)
			fmt.Println("got mem only:", v.MemOnly)

			fmt.Printf("record %d contains %s in the sql.NullString.String\n", i, v.Country.String)
			if v.Country.String == "" {
				fmt.Println("v.Country.String contained ''")
			}

			if v.NewColumn1 != nil {
				fmt.Printf("record %d contains %v in its *string pointer\n", i, *v.NewColumn1)
			} else {
				fmt.Println("v.NewColumn1  contained nil")
			}

			fmt.Printf("record %d contains %d in the sql.NullInt64.Int64\n", i, v.NewColumn2.Int64)
			if v.NewColumn2.Int64 == 0 {
				fmt.Println("v.NewColumn2.Int64 contained 0")
			}

			fmt.Printf("record %d contains %f in the sql.NullFloat64.Float64\n", i, v.NewColumn3.Float64)
			if v.NewColumn3.Float64 == 0 {
				fmt.Println("v.NewColumn3.Float64 contained 0")
			}

			fmt.Printf("record %d contains %v in the sql.NullBool.Bool\n", i, v.Active.Bool)
			if v.Active.Bool == false {
				fmt.Println("v.Active.Bool contained false")
			}
		}
	}
}

// TestNonPersistentColumn
//
// Test the non-persistent column support, indicated
// in the SqacTags by Name == "-" and Value = "".
func TestNonPersistentColumn(t *testing.T) {

	type Depot struct {
		DepotNum            int       `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		CreateDate          time.Time `db:"create_date" sqac:"nullable:false;default:now();index:unique"`
		Region              string    `db:"region" sqac:"nullable:false;default:YYC"`
		Province            string    `db:"province" sqac:"nullable:false;default:AB"`
		Country             string    `db:"country" sqac:"nullable:true;default:CA"`
		NewColumn1          string    `db:"new_column1" sqac:"nullable:false"`
		NewColumn2          int64     `db:"new_column2" sqac:"nullable:false"`
		NewColumn3          float64   `db:"new_column3" sqac:"nullable:false;default:0.0"`
		NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	}

	// drop table depot
	err := Handle.DropTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table names as per the table creation logic
	tn := common.GetTableName(Depot{})

	// create table depot in the db
	err = Handle.CreateTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// check for column 'region'
	if !Handle.ExistsColumn(tn, "region") {
		t.Errorf("column %s should have been found in table %s", "region", tn)
	}

	// check for column 'non_persistent_column'
	if Handle.ExistsColumn(tn, "non_persistent_column") {
		t.Errorf("column %s should not have been found in table %s", "non_persistent_column", tn)
	}

	// drop table depot
	err = Handle.DropTables(Depot{})
	if err != nil {
		t.Errorf("failed to drop table %s", tn)
	}
}

// TestTimeSimple
//
// Test time implementation
func TestTimeSimple(t *testing.T) {

	type DepotTime struct {
		DepotNum            int        `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		DepotBay            int        `db:"depot_bay" sqac:"primary_key:"`
		Region              string     `db:"region" sqac:"nullable:false;defalt:YYC"`
		TimeColUTC          time.Time  `db:"time_col_utc" sqac:"nullable:false"`
		TimeNowLocal        time.Time  `db:"time_now_local" sqac:"nullable:false"`
		TimeNowUTC          time.Time  `db:"time_now_utc" sqac:"nullable:false;default:now()"`
		TimeColNowDflt      time.Time  `db:"time_col_now_dflt" sqac:"nullable:false;default:now()"`
		TimeColEot          time.Time  `db:"time_col_eot" sqac:"nullable:false;default:eot()"`
		TimeNull            *time.Time `db:"time_null" sqac:"nullable:true"`
		TimeNotNull         *time.Time `db:"time_not_null" sqac:"nullable:true"`
		TimeNullWithDefault *time.Time `db:"time_null_with_default" sqac:"nullable:true;default:eot()"`
	}

	// determine the table names as per the table creation logic
	tn := common.GetTableName(DepotTime{})

	// drop table depottime if it exists
	err := Handle.DropTables(DepotTime{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table depottime
	err = Handle.CreateTables(DepotTime{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table depottime exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create a new record via the CRUD Create call
	tNotNull := new(time.Time)
	*tNotNull = time.Now().Local()

	// st := Handle.TimeToFormattedString(tNotNull)
	// fmt.Println("st:", st)
	// tNow := time.Now().Local()
	// stNow := Handle.TimeToFormattedString(tNow)
	// fmt.Println("stNow", stNow)

	var depottime = DepotTime{
		Region:       "YYC",
		TimeColUTC:   time.Date(1970, 01, 01, 11, 00, 00, 651387237, time.UTC), // time.Now(),
		TimeNowLocal: time.Now().Local(),
		TimeNowUTC:   time.Now().UTC(),
		// TimeNull:     nil,         omission == nil
		TimeNotNull: tNotNull,
		// TimeNullWithDefault: nil,
	}

	err = Handle.Create(&depottime)
	if err != nil {
		t.Errorf(err.Error())
	}

	if Handle.IsLog() {
		fmt.Printf("INSERTING: %v\n\n", depottime)
		fmt.Printf("TEST GOT: %v\n\n", depottime)
	}

	dt2 := DepotTime{
		DepotNum: depottime.DepotNum,
	}

	err = Handle.GetEntity(&dt2)
	if err != nil {
		t.Errorf("GetEntity failed with: %v", err)
	}

	if Handle.IsLog() {
		fmt.Printf("GetEntity for key %v returned: %v\n", dt2.DepotNum, dt2)
	}
}

// TestCRUDCreate
//
// Test CRUD Create
func TestCRUDCreate(t *testing.T) {

	type DepotCreate struct {
		DepotNum            int       `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		DepotBay            int       `db:"depot_bay" sqac:"primary_key:"`
		CreateDate          time.Time `db:"create_date" sqac:"nullable:false;default:now();index:unique"`
		Region              string    `db:"region" sqac:"nullable:false;default:YYC"`
		Province            string    `db:"province" sqac:"nullable:false;default:AB"`
		Country             string    `db:"country" sqac:"nullable:true;default:CA"`
		NewColumn1          string    `db:"new_column1" sqac:"nullable:false"`
		NewColumn2          int64     `db:"new_column2" sqac:"nullable:false"`
		NewColumn3          float64   `db:"new_column3" sqac:"nullable:false;default:0.0"`
		IntDefaultZero      int       `db:"int_default_zero" sqac:"nullable:false;default:0"`
		IntDefault42        int       `db:"int_default42" sqac:"nullable:false;default:42"`
		FldOne              int       `db:"fld_one" sqac:"nullable:false;default:0;index:idx_depotcreate_fld_one_fld_two"`
		FldTwo              int       `db:"fld_two" sqac:"nullable:false;default:0;index:idx_depotcreate_fld_one_fld_two"`
		TimeCol             time.Time `db:"time_col" sqac:"nullable:false"`
		TimeColNow          time.Time `db:"time_col_now" sqac:"nullable:false;default:now()"`
		TimeColEot          time.Time `db:"time_col_eot" sqac:"nullable:false;default:eot()"`
		IntZeroValNoDefault int       `db:"int_zero_val_no_default" sqac:"nullable:false"`
		NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	}

	// determine the table names as per the table creation logic
	tn := common.GetTableName(DepotCreate{})

	// create table depot
	err := Handle.CreateTables(DepotCreate{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create a new record via the CRUD Create call
	var depot = DepotCreate{
		Region:              "YYC",
		NewColumn1:          "string_value",
		NewColumn2:          9999,
		NewColumn3:          45.33,
		NonPersistentColumn: "0123456789abcdef",
		TimeCol:             time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC), // time.Now(),
	}

	err = Handle.Create(&depot)
	if err != nil {
		t.Errorf(err.Error())
	}
	if Handle.IsLog() {
		fmt.Printf("INSERTING: %v\n", depot)
		fmt.Printf("TEST GOT: %v\n", depot)
	}

	// err = Handle.DropTables(DepotCreate{})
	// if err != nil {
	// 	t.Errorf("failed to drop table %s", tn)
	// }
}

// TestCRUDUpdate
//
// Test CRUD Update
func TestCRUDUpdate(t *testing.T) {

	type Depot struct {
		DepotNum             int       `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		DepotBay             int       `db:"depot_bay" sqac:"primary_key:"`
		TestKeyDate          time.Time `db:"test_key_date" sqac:"primary_key:;default:now()"`
		CreateDate           time.Time `db:"create_date" sqac:"nullable:false;default:now();index:unique"`
		Region               string    `db:"region" sqac:"nullable:false;default:YYC"`
		Province             string    `db:"province" sqac:"nullable:false;default:AB"`
		Country              string    `db:"country" sqac:"nullable:true;default:CA"`
		NewColumn1           string    `db:"new_column1" sqac:"nullable:false"`
		NewColumn2           int64     `db:"new_column2" sqac:"nullable:false"`
		NewColumn3           float64   `db:"new_column3" sqac:"nullable:false;default:0.0"`
		FldOne               int       `db:"fld_one" sqac:"nullable:false;default:0;index:idx_depot_fld_one_fld_two"`
		FldTwo               int       `db:"fld_two" sqac:"nullable:false;default:0;index:idx_depot_fld_one_fld_two"`
		NonPersistentColumn  string    `db:"non_persistent_column" sqac:"-"`
		NonPersistentColumn2 string    `db:"non_persistent_column" sqac:"-"`
	}

	// determine the table names as per the table creation logic
	tn := common.GetTableName(Depot{})

	err := Handle.DestructiveResetTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create a new record via the CRUD Create call
	var depot = Depot{
		Region:              "YYC",
		NewColumn1:          "string_value",
		NewColumn2:          9999,
		NewColumn3:          45.33,
		NonPersistentColumn: "0123456789abcdef",
	}

	// create a new record in the depot table
	err = Handle.Create(&depot)
	if err != nil {
		t.Errorf(err.Error())
	}

	if Handle.IsLog() {
		fmt.Printf("INSERT to table %s returned: %v\n", tn, depot)
	}

	// check that the primary-key field has a value
	if depot.DepotNum == 0 {
		t.Errorf("insert to table %s failed", tn)
	}

	// update the existing record in the depot table
	depot.Region = "YYZ"               // YYC -> YYZ
	depot.Province = "ON"              // AB -> ON
	depot.NewColumn1 = "updated_value" // "string_value" -> "updated_value"
	depot.NewColumn2 = 1111            // 9999 -> 1111
	depot.NewColumn3 = 3333.5556       // 45.33 -> 3333.5556
	depot.NonPersistentColumn = "this value will not get stored in the db"

	err = Handle.Update(&depot)
	if err != nil {
		t.Errorf(err.Error())
	}

	if Handle.IsLog() {
		fmt.Printf("UPDATE to table %s returned: %v\n", tn, depot)
	}

	err = Handle.DropTables(Depot{})
	if err != nil {
		t.Errorf("failed to drop table %s", tn)
	}
}

// TestCRUDDelete
//
// Test CRUD Delete
func TestCRUDDelete(t *testing.T) {

	type DepotDelete struct {
		DepotNum            int       `db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		DepotBay            int       `db:"depot_bay" sqac:"primary_key:"`
		CreateDate          time.Time `db:"create_date" sqac:"nullable:false;default:now();index:unique"`
		Region              string    `db:"region" sqac:"nullable:false;default:YYC"`
		Province            string    `db:"province" sqac:"nullable:false;default:AB"`
		Country             string    `db:"country" sqac:"nullable:true;default:CA"`
		NewColumn1          string    `db:"new_column1" sqac:"nullable:false"`
		NewColumn2          int64     `db:"new_column2" sqac:"nullable:false"`
		NewColumn3          float64   `db:"new_column3" sqac:"nullable:false;default:0.0"`
		IntDefaultZero      int       `db:"int_default_zero" sqac:"nullable:false;default:0"`
		IntDefault42        int       `db:"int_default42" sqac:"nullable:false;default:42"`
		IntZeroValNoDefault int       `db:"int_zero_val_no_default" sqac:"nullable:false"`
		FldOne              int       `db:"fld_one" sqac:"nullable:false;default:0;index:idx_depotdelete_fld_one_fld_two"`
		FldTwo              int       `db:"fld_two" sqac:"nullable:false;default:0;index:idx_depotdelete_fld_one_fld_two"`
		NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	}

	// determine the table names as per the table creation logic
	tn := common.GetTableName(DepotDelete{})

	// create table depot if it does not exist
	err := Handle.CreateTables(DepotDelete{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create a new record via the CRUD Create call
	var depot = DepotDelete{
		Region:              "YYC",
		NewColumn1:          "string_value",
		NewColumn2:          9999,
		NewColumn3:          45.33,
		NonPersistentColumn: "0123456789abcdef",
	}

	err = Handle.Create(&depot)
	if err != nil {
		t.Errorf(err.Error())
	}

	err = Handle.Delete(&depot)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGet
//
// Test CRUD Get
func TestCRUDGet(t *testing.T) {

	type DepotGet struct {
		DepotNum            int       `json:"depot_num" db:"depot_num" sqac:"primary_key:inc;start:90000000"`
		DepotBay            int       `json:"depot_bay" db:"depot_bay" sqac:"primary_key:"`
		CreateDate          time.Time `json:"create_date" db:"create_date" sqac:"nullable:false;default:now();index:unique"`
		Region              string    `json:"region" db:"region" sqac:"nullable:false;default:YYC"`
		Province            string    `json:"province" db:"province" sqac:"nullable:false;default:AB"`
		Country             string    `json:"country" db:"country" sqac:"nullable:true;default:CA"`
		NewColumn1          string    `json:"new_column1" db:"new_column1" sqac:"nullable:false"`
		NewColumn2          int64     `json:"new_column2" db:"new_column2" sqac:"nullable:false"`
		NewColumn3          float64   `json:"new_column3" db:"new_column3" sqac:"nullable:false;default:0.0"`
		IntDefaultZero      int       `json:"int_default_zero" db:"int_default_zero" sqac:"nullable:false;default:0"`
		IntDefault42        int       `json:"int_default42" db:"int_default42" sqac:"nullable:false;default:42"`
		FldOne              int       `json:"fld_one" db:"fld_one" sqac:"nullable:false;default:0;index:idx_depotget_fld_one_fld_two"`
		FldTwo              int       `json:"fld_two" db:"fld_two" sqac:"nullable:false;default:0;index:idx_depotget_fld_one_fld_two"`
		IntZeroValNoDefault int       `json:"int_zero_val_no_default" db:"int_zero_val_no_default" sqac:"nullable:false"`
		NonPersistentColumn string    `json:"non_persistent_column" db:"non_persistent_column" sqac:"-"`
	}

	// determine the table names as per the table creation logic
	tn := common.GetTableName(DepotGet{})

	// create table depot
	err := Handle.CreateTables(DepotGet{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table depotget exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create a new record via the CRUD Create call
	var depot = DepotGet{
		Region:              "YYC",
		NewColumn1:          "string_value",
		NewColumn2:          9999,
		NewColumn3:          45.33,
		NonPersistentColumn: "0123456789abcdef",
	}

	err = Handle.Create(&depot)
	if err != nil {
		t.Errorf(err.Error())
	}

	// create a struct to read into and populate the keys
	depotRead := DepotGet{
		DepotNum: depot.DepotNum,
		DepotBay: depot.DepotBay,
	}

	err = Handle.GetEntity(&depotRead)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	if depotRead.Region != "YYC" {
		t.Errorf("depotRead.Region error!")
	}
	// err = Handle.DropTables(Depot{})
	// if err != nil {
	// 	t.Errorf("failed to drop table %s", tn)
	// }
}

// TestCRUDGetEntities
//
// Test CRUD Get
func TestCRUDGetEntities(t *testing.T) {

	// determine the table names as per the table creation logic
	tn := common.GetTableName(DepotGetEntities2{})

	// drop table depotgetentities
	err := Handle.DropTables(DepotGetEntities2{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table depotgetentities
	err = Handle.CreateTables(DepotGetEntities2{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table depotgetentities2 exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create a new record via the CRUD Create call
	var depotgetentities = DepotGetEntities2{
		Region:              "YYC",
		NewColumn1:          "string_value",
		NewColumn2:          9999,
		NewColumn3:          45.33,
		NonPersistentColumn: "0123456789abcdef",
	}

	err = Handle.Create(&depotgetentities)
	if err != nil {
		t.Errorf(err.Error())
	}

	depotgetentities2 := DepotGetEntities2{
		Region:              "YVR",
		NewColumn1:          "vancouver",
		NewColumn2:          8888,
		NewColumn3:          4642.22,
		NonPersistentColumn: "don't save me",
	}

	err = Handle.Create(&depotgetentities2)
	if err != nil {
		t.Errorf(err.Error())
	}

	// create a slice to read into
	depotRead := []DepotGetEntities2{}

	result, err := Handle.GetEntities(depotRead)
	if err != nil {
		t.Errorf(err.Error())
	}

	if Handle.IsLog() {
		fmt.Println("DEPOTREAD:", depotRead)
	}

	depotReadResult := reflect.ValueOf(result)
	for i := 0; i < depotReadResult.Len(); i++ {
		if Handle.IsLog() {
			fmt.Printf("index[%d]: %v\n", i, depotReadResult.Index(i))
		}
		depotRead = append(depotRead, depotReadResult.Index(i).Interface().(DepotGetEntities2))
	}

	if len(depotRead) == 0 {
		t.Errorf("failed to read any entities from test table DepotGetEntities2")
	}
}

// TestCRUDGetEntities2
//
// Test CRUD Get
func TestCRUDGetEntities2(t *testing.T) {

	// determine the table names as per the table creation logic
	tn := common.GetTableName(DepotGetEntities2{})

	// drop table depotgetentities2
	err := Handle.DropTables(DepotGetEntities2{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table depotgetentities
	err = Handle.CreateTables(DepotGetEntities2{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table depotgetentities2 exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create a new record via the CRUD Create call
	var depotgetentities = DepotGetEntities2{
		Region:              "YYC",
		NewColumn1:          "string_value",
		NewColumn2:          9999,
		NewColumn3:          45.33,
		NonPersistentColumn: "0123456789abcdef",
	}

	err = Handle.Create(&depotgetentities)
	if err != nil {
		t.Errorf(err.Error())
	}

	depotgetentities2 := DepotGetEntities2{
		Region:              "YVR",
		NewColumn1:          "vancouver",
		NewColumn2:          8888,
		NewColumn3:          46488887772.22,
		NonPersistentColumn: "don't save me",
	}

	err = Handle.Create(&depotgetentities2)
	if err != nil {
		t.Errorf(err.Error())
	}

	// create an implementation of the DepotGetEntitiesTab struct.
	// this struct contains a slice for the query results, as well as
	// an implementation of the sqac.GetEnt{} interface.  See the Exec()
	// method implementation at the top of this file for details.

	depotRead := DepotGetEntitiesTab{}

	err = Handle.GetEntities2(&depotRead)
	if err != nil {
		t.Errorf(err.Error())
	}

	if Handle.IsLog() {
		fmt.Println("depotRead:", depotRead)
		for _, v := range depotRead.ents {
			fmt.Println(v)
		}
	}

	if len(depotRead.ents) == 0 {
		t.Errorf("failed to read any entities from test table DepotGetEntities2")
	}

}

// TestCRUDGetEntities4
//
// Test CRUD Get
func TestCRUDGetEntities4(t *testing.T) {

	// determine the table names as per the table creation logic
	tn := common.GetTableName(DepotGetEntities2{})

	// drop table depotgetentities
	err := Handle.DropTables(DepotGetEntities2{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table depotgetentities
	err = Handle.CreateTables(DepotGetEntities2{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table depotgetentities2 exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create a new record via the CRUD Create call
	var depotgetentities = DepotGetEntities2{
		Region:              "YYC",
		NewColumn1:          "string_value",
		NewColumn2:          9999,
		NewColumn3:          45.33,
		NonPersistentColumn: "0123456789abcdef",
	}

	err = Handle.Create(&depotgetentities)
	if err != nil {
		t.Errorf(err.Error())
	}

	depotgetentities2 := DepotGetEntities2{
		Region:              "YVR",
		NewColumn1:          "vancouver",
		NewColumn2:          8888,
		NewColumn3:          464773.22,
		NonPersistentColumn: "don't save me",
	}

	err = Handle.Create(&depotgetentities2)
	if err != nil {
		t.Errorf(err.Error())
	}

	// create a slice to read into
	depotRead := []DepotGetEntities2{}

	Handle.GetEntities4(&depotRead)
	// fmt.Println("DEPOTREAD:", depotRead)
}

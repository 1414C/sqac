package sqac_test

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/1414C/sqac"
	_ "github.com/SAP/go-hdb/driver"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	Handle sqac.PublicDB
)

func TestMain(m *testing.M) {

	// parse flags
	dbFlag := flag.String("db", "pg", "db to connect to")
	logFlag := flag.Bool("l", false, "activate sqac logging")
	flag.Parse()

	// // select the db implementation
	// switch *dbFlag {
	// case "pg":
	// 	pgh := new(sqac.PostgresFlavor)
	// 	Handle = pgh
	// 	db, err := sqac.Open("postgres", "host=127.0.0.1 user=godev dbname=sqlx sslmode=disable password=gogogo123")
	// 	if err != nil {
	// 		log.Fatalf("%s\n", err.Error())
	// 	}
	// 	Handle.SetDB(db)
	// 	defer db.Close()

	// case "mysql":
	// 	myh := new(sqac.MySQLFlavor)
	// 	Handle = myh
	// 	db, err := sqac.Open("mysql", "stevem:gogogo123@tcp(192.168.1.50:3306)/sqlx?charset=utf8&parseTime=True&loc=Local")
	// 	if err != nil {
	// 		log.Fatalf("%s\n", err.Error())
	// 	}
	// 	Handle.SetDB(db)
	// 	defer db.Close()

	// case "sqlite":
	// 	sqh := new(sqac.SQLiteFlavor)
	// 	Handle = sqh
	// 	db, err := sqac.Open("sqlite3", "testdb.sqlite")
	// 	if err != nil {
	// 		log.Fatalf("%s\n", err.Error())
	// 	}
	// 	Handle.SetDB(db)
	// 	defer db.Close()

	// case "mssql":
	// 	msh := new(sqac.MSSQLFlavor)
	// 	Handle = msh
	// 	db, err := sqac.Open("mssql", "sqlserver://SA:Bunny123!!@localhost:1401?database=sqlx")
	// 	if err != nil {
	// 		log.Fatalf("%s\n", err.Error())
	// 	}
	// 	Handle.SetDB(db)
	// 	err = db.Ping()
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	defer db.Close()

	// case "db2":

	// case "go-hdb":

	// default:

	// }

	// // detailed logging?
	// if *logFlag {
	// 	Handle.Log(true)
	// } else {
	// 	Handle.Log(false)
	// }

	var cs string
	switch *dbFlag {
	case "pg":
		cs = "host=127.0.0.1 user=godev dbname=sqlx sslmode=disable password=gogogo123"
	case "mysql":
		cs = "stevem:gogogo123@tcp(192.168.1.50:3306)/sqlx?charset=utf8&parseTime=True&loc=Local"
	case "sqlite":
		cs = "testdb.sqlite"
	case "mssql":
		cs = "sqlserver://SA:Bunny123!!@localhost:1401?database=sqlx"
	case "db2":
		cs = ""
	case "hdb":
		cs = "hdb://SMACLEOD:Blockhead1@clkhana01.lab.clockwork.ca:30047"
	default:
		cs = ""
	}
	Handle = sqac.Create(*dbFlag, *logFlag, cs)

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

// // TestTimeSimple
// //
// // Test time implementation
// func TestTimeSimple(t *testing.T) {

// 	type DepotTime struct {
// 		DepotNum       int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
// 		DepotBay       int       `db:"depot_bay" rgen:"primary_key:"`
// 		Region         string    `db:"region" rgen:"nullable:false;defalt:YYC"`
// 		TimeColUTC     time.Time `db:"time_col_utc" rgen:"nullable:false"`
// 		TimeNowLocal   time.Time `db:"time_now_local" rgen:"nullable:false"`
// 		TimeNowUTC     time.Time `db:"time_now_utc" rgen:"nullable:false;default:now()"`
// 		TimeColNowDflt time.Time `db:"time_col_now_dflt" rgen:"nullable:false;default:now()"`
// 		TimeColEot     time.Time `db:"time_col_eot" rgen:"nullable:false;default:eot"`
// 	}

// 	// determine the table names as per the
// 	// table creation logic
// 	tn := reflect.TypeOf(DepotTime{}).String()
// 	if strings.Contains(tn, ".") {
// 		el := strings.Split(tn, ".")
// 		tn = strings.ToLower(el[len(el)-1])
// 	} else {
// 		tn = strings.ToLower(tn)
// 	}

// 	// drop table depottime if it exists
// 	err := Handle.DropTables(DepotTime{})
// 	if err != nil {
// 		t.Errorf("%s", err.Error())
// 	}

// 	// create table depottime
// 	err = Handle.CreateTables(DepotTime{})
// 	if err != nil {
// 		t.Errorf("%s", err.Error())
// 	}

// 	// expect that table depottime exists
// 	if !Handle.ExistsTable(tn) {
// 		t.Errorf("table %s does not exist", tn)
// 	}

// 	// create a new record via the CRUD Create call
// 	var depot = DepotTime{
// 		Region:       "YYC",
// 		TimeColUTC:   time.Date(1970, 01, 01, 11, 00, 00, 651387237, time.UTC), // time.Now(),
// 		TimeNowLocal: time.Now().Local(),
// 		TimeNowUTC:   time.Now().UTC(),
// 	}

// 	err = Handle.Create(&depot)
// 	if err != nil {
// 		t.Errorf(err.Error())
// 	}
// 	fmt.Println("")

// 	if Handle.IsLog() {
// 		fmt.Printf("INSERTING: %v\n\n", depot)
// 		fmt.Printf("TEST GOT: %v\n\n", depot)
// 	}
// 	fmt.Printf("TEST GOT: %v\n\n", depot)
// 	// os.Exit(0)
// }

// TestExistsTableNegative
//
// Test for non-existent table 'Footle'
//
func TestExistsTableNegative(t *testing.T) {

	type Footle struct {
		KeyNum      int       `db:"key_num" rgen:"primary_key:inc"`
		CreateDate  time.Time `db:"create_date" rgen:"nullable:false;default:now();"`
		Description string    `db:"description" rgen:"nullable:false;default:"`
	}

	// determine the table name as per the
	// table creation logic
	tn := reflect.TypeOf(Footle{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

	// expect that table depot does not exist
	if Handle.ExistsTable(tn) {
		t.Errorf("table %s was found when it was not expected", tn)
	}
}

// TestCreateTableBasic
//
// Create table depot via CreateTables(i ...interface{})
// Verify table creation via ExistsTable(tn string)
// Perform negative validation be checking for non-existant
// 	table "abcdefg" via ExistsTable(tn string)
//
func TestCreateTableBasic(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" rgen:"primary_key:inc"`
		CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now();"`
		Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:false;default:CA"`
	}

	err := Handle.CreateTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the
	// table creation logic
	tn := reflect.TypeOf(Depot{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s was not created", tn)
	}
}

// TestDropTablesBasic
//
// Drop table depot via DropTables(i ...interface{})
//
func TestDropTablesBasic(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" rgen:"primary_key:inc"`
		CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now();"`
		Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:false;default:CA"`
	}

	err := Handle.DropTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the
	// table creation logic
	tn := reflect.TypeOf(Depot{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

	// expect that table depot has been dropped
	if Handle.ExistsTable(tn) {
		t.Errorf("table %s was not dropped", tn)
	}
}

// TestCreateTableWithAlterSequence
//
// Create table depot via CreateTables(i ...interface{})
// Verify table creation via ExistsTable(tn string)
// Perform negative validation be checking for non-existant
// 	table "abcdefg" via ExistsTable(tn string)
//
func TestCreateTableWithAlterSequence(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now()"`
		Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:false;default:CA"`
	}

	err := Handle.CreateTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the
	// table creation logic
	tn := reflect.TypeOf(Depot{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

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
		hdbName = tn + "+" + "depot_num"
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
//
func TestCreateTablesWithInclude(t *testing.T) {

	// type Triplet struct {
	// 	TripOne   string `db:"trip_one" rgen:"nullable:false"`
	// 	TripTwo   int64  `db:"trip_two" rgen:"nullable:false;default:0"`
	// 	Tripthree string `db:"trip_three" rgen:"nullable:false"`
	// }

	// type Equipment struct {
	// 	EquipmentNum   int64     `db:"equipment_num" rgen:"primary_key:inc;start:55550000"`
	// 	ValidFrom      time.Time `db:"valid_from" rgen:"primary_key;nullable:false;default:now()"`
	// 	ValidTo        time.Time `db:"valid_to" rgen:"primary_key;nullable:false;default:make_timestamptz(9999, 12, 31, 23, 59, 59.9)"`
	// 	CreatedAt      time.Time `db:"created_at" rgen:"nullable:false;default:now()"`
	// 	InspectionAt   time.Time `db:"inspeaction_at" rgen:"nullable:true"`
	// 	MaterialNum    int       `db:"material_num" rgen:"index:idx_material_num_serial_num"`
	// 	Description    string    `db:"description" rgen:"rgen:nullable:false"`
	// 	SerialNum      string    `db:"serial_num" rgen:"index:idx_material_num_serial_num"`
	// 	IntExample     int       `db:"int_example" rgen:"nullable:false;default:0"`
	// 	Int64Example   int64     `db:"int64_example" rgen:"nullable:false;default:0"`
	// 	Int32Example   int32     `db:"int32_example" rgen:"nullable:false;default:0"`
	// 	Int16Example   int16     `db:"int16_example" rgen:"nullable:false;default:0"`
	// 	Int8Example    int8      `db:"int8_example" rgen:"nullable:false;default:0"`
	// 	UIntExample    uint      `db:"uint_example" rgen:"nullable:false;default:0"`
	// 	UInt64Example  uint64    `db:"uint64_example" rgen:"nullable:false;default:0"`
	// 	UInt32Example  uint32    `db:"uint32_example" rgen:"nullable:false;default:0"`
	// 	UInt16Example  uint16    `db:"uint16_example" rgen:"nullable:false;default:0"`
	// 	UInt8Example   uint8     `db:"uint8_example" rgen:"nullable:false;default:0"`
	// 	Float32Example float32   `db:"float32_example" rgen:"nullable:false;default:0.0"`
	// 	Float64Example float64   `db:"float64_example" rgen:"nullable:false;default:0.0"`
	// 	BoolExample    bool      `db:"bool_example" rgen:"nullable:false;default:false"`
	// 	RuneExample    rune      `db:"rune_example" rgen:"nullable:true"`
	// 	ByteExample    byte      `db:"byte_example" rgen:"nullable:true"`
	// 	Triplet
	// }

	err := Handle.CreateTables(Equipment{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the
	// table creation logic
	tn := reflect.TypeOf(Equipment{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

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

// TestExistsIndexNegative
//
// Check to see if index:
// idx_province_country exists on
// table depot.
func TestExistsIndexNegative(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now()"`
		Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:false;default:CA"`
		NewColumn1 string    `db:"new_column1" rgen:"nullable:false"`
		NewColumn2 int64     `db:"new_column2" rgen:"nullable:false;default:0"`
		NewColumn3 float64   `db:"new_column3" rgen:"nullable:false;default:0.0"`
	}

	// ensure that table depot exists
	err := Handle.AlterTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the
	// table creation logic
	tn := reflect.TypeOf(Depot{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

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
		DepotNum   int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
		Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:false;default:CA"`
		NewColumn1 string    `db:"new_column1" rgen:"nullable:false"`
		NewColumn2 int64     `db:"new_column2" rgen:"nullable:false;default:0"`
		NewColumn3 float64   `db:"new_column3" rgen:"nullable:false;default:0.0"`
	}

	err := Handle.CreateTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the
	// table creation logic
	tn := reflect.TypeOf(Depot{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

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
		DepotNum   int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now();index:non-unique"`
		Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:false;default:CA"`
		NewColumn1 string    `db:"new_column1" rgen:"nullable:false"`
		NewColumn2 int64     `db:"new_column2" rgen:"nullable:false;default:0"`
		NewColumn3 float64   `db:"new_column3" rgen:"nullable:false;default:0.0"`
	}

	err := Handle.CreateTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the
	// table creation logic
	tn := reflect.TypeOf(Depot{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

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
		DepotNum   int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now()"`
		Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:false;default:CA"`
		NewColumn1 string    `db:"new_column1" rgen:"nullable:false"`
		NewColumn2 int64     `db:"new_column2" rgen:"nullable:false;default:0"`
		NewColumn3 float64   `db:"new_column3" rgen:"nullable:false;default:0.0"`
	}

	err := Handle.CreateTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the
	// table creation logic
	tn := reflect.TypeOf(Depot{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

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
		DepotNum   int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
		Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:false;default:CA"`
		NewColumn1 string    `db:"new_column1" rgen:"nullable:false"`
		NewColumn2 int64     `db:"new_column2" rgen:"nullable:false;default:0"`
		NewColumn3 float64   `db:"new_column3" rgen:"nullable:false;default:0.0"`
	}

	// determine the table name as per the
	// table creation logic
	tn := reflect.TypeOf(Depot{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

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
		DepotNum   int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
		Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:false;default:CA"`
		NewColumn1 string    `db:"new_column1" rgen:"nullable:false"`
		NewColumn2 int64     `db:"new_column2" rgen:"nullable:false;default:0"`
		NewColumn3 float64   `db:"new_column3" rgen:"nullable:false;default:0.0"`
	}

	// ensure table depot exists
	err := Handle.AlterTables(Depot{})
	if err != nil {
		fmt.Println("GOT ERROR!")
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the
	// table creation logic
	tn := reflect.TypeOf(Depot{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

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
		DepotNum   int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now()"`
		Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:false;default:CA"`
		NewColumn1 string    `db:"new_column1" rgen:"nullable:false;index:idx_depot_new_column1_new_column2"`
		NewColumn2 int64     `db:"new_column2" rgen:"nullable:false;default:0;index:idx_depot_new_column1_new_column2"`
		NewColumn3 float64   `db:"new_column3" rgen:"nullable:false;default:0.0"`
	}

	err := Handle.CreateTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the
	// table creation logic
	tn := reflect.TypeOf(Depot{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

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
		DepotNum   int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
		Region     string    `db:"region" rgen:"nullable:false;default:YYC;index:non-unique"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:true;default:CA"`
	}

	// determine the table name as per the
	// table creation logic
	tn := reflect.TypeOf(Depot{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

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
//  - NewColumn1 string    `db:"new_column1" rgen:"nullable:false"`
//	- NewColumn2 int64     `db:"new_column2" rgen:"nullable:false;default:0"`
//  - NewColumn3 float64   `db:"new_column3" rgen:"nullable:false;default:0.0"`
//
func TestAlterTables(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
		Region     string    `db:"region" rgen:"nullable:false;default:YYC;index:non-unique"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:false;default:CA"`
		NewColumn1 string    `db:"new_column1" rgen:"nullable:false;default:nc1_default;index:non-unique"`
		NewColumn2 int64     `db:"new_column2" rgen:"nullable:false;default:0;index:idx_new_column2_new_column3"`
		NewColumn3 float64   `db:"new_column3" rgen:"nullable:false;default:0.0;index:idx_new_column2_new_column3"`
	}

	err := Handle.AlterTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := reflect.TypeOf(Depot{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

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
		DepotNum   int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
		Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:false;default:CA"`
		Active     bool      `db:"active" rgen:"nullable:false;default:true"`
	}

	// determine the table names as per the
	// table creation logic
	tns := make([]string, 0)
	tn := reflect.TypeOf(Depot{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}
	tns = append(tns, tn)

	tn = reflect.TypeOf(Equipment{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}
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

// TestNullableValues
//
// This test is designed to illustrate the handling of
// database reads when dealing with db fields that
// contain null values.
// Create db table depot based on an updated Depot
// struct containing a number of nullable and non-
// defaulted fields.
// Insert a new record containing null-values into
// db table depot.
// Declare struct DepotN{} as a parallel structure to
// Depot{} making use of sql.Null<type> fields in
// place of the gotypes for the nullable fields.
// Note that DepotN{} also contains one *string
// pointer type instead of sql.NullString in order
// to demonstrate a different way to handle the
// situation.
// Read all the records (1) from db table depot
// assigning them to a slice declared as type
// DepotN.
// Iterate over the record(s) contained in the
// result set and take note of the manner in
// which the nullable field values are accessed /
// converted from nil values to their base-type's
// default value.  In this example, the Valid
// bool flag in the nullable field is not checked,
// as it is typically(?) okay to simply ask for
// base-type default through .Sting, .Int64,
// .Float64 or .Bool.
func TestNullableValues(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
		Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
		MemOnly    string    `db:"mem_only" rgen:"-"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:true;"`    // nullable
		NewColumn1 string    `db:"new_column1" rgen:"nullable:true"` // nullable
		NewColumn2 int64     `db:"new_column2" rgen:"nullable:true"` // nullable
		NewColumn3 float64   `db:"new_column3" rgen:"nullable:true"` // nullable
		Active     bool      `db:"active" rgen:"nullable:true"`      // nullable
	}

	// create table depot
	err := Handle.CreateTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table names as per the
	// table creation logic
	tn := reflect.TypeOf(Depot{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

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
		incKey := 0
		keyQuery := "SELECT SEQ_DEPOT_DEPOT_NUM.NEXTVAL FROM DUMMY;"
		err = Handle.ExecuteQueryRowx(keyQuery).Scan(&incKey)
		if err != nil {
			t.Errorf(err.Error())
		}
		insQuery = fmt.Sprintf("INSERT INTO depot (depot_num, region, province) VALUES ('%d, YVR','AB');", incKey)
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
		DepotNum   int             `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		CreateDate time.Time       `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
		Region     string          `db:"region" rgen:"nullable:false;default:YYC"`
		MemOnly    string          `db:"mem_only" rgen:"-"`
		Province   string          `db:"province" rgen:"nullable:false;default:AB"`
		Country    sql.NullString  `db:"country" rgen:"nullable:true;"`
		NewColumn1 *string         `db:"new_column1" rgen:"nullable:true"`
		NewColumn2 sql.NullInt64   `db:"new_column2" rgen:"nullable:true"`
		NewColumn3 sql.NullFloat64 `db:"new_column3" rgen:"nullable:true"`
		Active     sql.NullBool    `db:"active" rgen:"nullable:true"`
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
// in the RgenTags by Name == "-" and Value = "".
func TestNonPersistentColumn(t *testing.T) {

	type Depot struct {
		DepotNum            int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		CreateDate          time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
		Region              string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province            string    `db:"province" rgen:"nullable:false;default:AB"`
		Country             string    `db:"country" rgen:"nullable:true;default:CA"`
		NewColumn1          string    `db:"new_column1" rgen:"nullable:false"`
		NewColumn2          int64     `db:"new_column2" rgen:"nullable:false"`
		NewColumn3          float64   `db:"new_column3" rgen:"nullable:false;default:0.0"`
		NonPersistentColumn string    `db:"non_persistent_column" rgen:"-"`
	}

	// drop table depot
	err := Handle.DropTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table names as per the
	// table creation logic
	tn := reflect.TypeOf(Depot{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

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

// TestCRUDCreate
//
// Test CRUD Create
func TestCRUDCreate(t *testing.T) {

	type DepotCreate struct {
		DepotNum            int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		DepotBay            int       `db:"depot_bay" rgen:"primary_key:"`
		CreateDate          time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
		Region              string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province            string    `db:"province" rgen:"nullable:false;default:AB"`
		Country             string    `db:"country" rgen:"nullable:true;default:CA"`
		NewColumn1          string    `db:"new_column1" rgen:"nullable:false"`
		NewColumn2          int64     `db:"new_column2" rgen:"nullable:false"`
		NewColumn3          float64   `db:"new_column3" rgen:"nullable:false;default:0.0"`
		IntDefaultZero      int       `db:"int_default_zero" rgen:"nullable:false;default:0"`
		IntDefault42        int       `db:"int_default42" rgen:"nullable:false;default:42"`
		FldOne              int       `db:"fld_one" rgen:"nullable:false;default:0;index:idx_depotcreate_fld_one_fld_two"`
		FldTwo              int       `db:"fld_two" rgen:"nullable:false;default:0;index:idx_depotcreate_fld_one_fld_two"`
		TimeCol             time.Time `db:"time_col" rgen:"nullable:false"`
		TimeColNow          time.Time `db:"time_col_now" rgen:"nullable:false;default:now()"`
		TimeColEot          time.Time `db:"time_col_eot" rgen:"nullable:false;default:eot"`
		IntZeroValNoDefault int       `db:"int_zero_val_no_default" rgen:"nullable:false"`
		NonPersistentColumn string    `db:"non_persistent_column" rgen:"-"`
	}

	// determine the table names as per the
	// table creation logic
	tn := reflect.TypeOf(DepotCreate{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

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
	fmt.Printf("TEST GOT: %v\n", depot)

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
		DepotNum             int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		DepotBay             int       `db:"depot_bay" rgen:"primary_key:"`
		TestKeyDate          time.Time `db:"test_key_date" rgen:"primary_key:;default:now()"`
		CreateDate           time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
		Region               string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province             string    `db:"province" rgen:"nullable:false;default:AB"`
		Country              string    `db:"country" rgen:"nullable:true;default:CA"`
		NewColumn1           string    `db:"new_column1" rgen:"nullable:false"`
		NewColumn2           int64     `db:"new_column2" rgen:"nullable:false"`
		NewColumn3           float64   `db:"new_column3" rgen:"nullable:false;default:0.0"`
		FldOne               int       `db:"fld_one" rgen:"nullable:false;default:0;index:idx_depot_fld_one_fld_two"`
		FldTwo               int       `db:"fld_two" rgen:"nullable:false;default:0;index:idx_depot_fld_one_fld_two"`
		NonPersistentColumn  string    `db:"non_persistent_column" rgen:"-"`
		NonPersistentColumn2 string    `db:"non_persistent_column" rgen:"nullable:true;-"`
	}

	// determine the table names as per the
	// table creation logic
	tn := reflect.TypeOf(Depot{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

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
	fmt.Printf("INSERT to table %s returned: %v\n", tn, depot)

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
	fmt.Printf("UPDATE to table %s returned: %v\n", tn, depot)

	// err = Handle.DropTables(Depot{})
	// if err != nil {
	// 	t.Errorf("failed to drop table %s", tn)
	// }
}

// TestCRUDDelete
//
// Test CRUD Delete
func TestCRUDDelete(t *testing.T) {

	type DepotDelete struct {
		DepotNum            int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		DepotBay            int       `db:"depot_bay" rgen:"primary_key:"`
		CreateDate          time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
		Region              string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province            string    `db:"province" rgen:"nullable:false;default:AB"`
		Country             string    `db:"country" rgen:"nullable:true;default:CA"`
		NewColumn1          string    `db:"new_column1" rgen:"nullable:false"`
		NewColumn2          int64     `db:"new_column2" rgen:"nullable:false"`
		NewColumn3          float64   `db:"new_column3" rgen:"nullable:false;default:0.0"`
		IntDefaultZero      int       `db:"int_default_zero" rgen:"nullable:false;default:0"`
		IntDefault42        int       `db:"int_default42" rgen:"nullable:false;default:42"`
		IntZeroValNoDefault int       `db:"int_zero_val_no_default" rgen:"nullable:false"`
		FldOne              int       `db:"fld_one" rgen:"nullable:false;default:0;index:idx_depotdelete_fld_one_fld_two"`
		FldTwo              int       `db:"fld_two" rgen:"nullable:false;default:0;index:idx_depotdelete_fld_one_fld_two"`
		NonPersistentColumn string    `db:"non_persistent_column" rgen:"-"`
	}

	// determine the table names as per the
	// table creation logic
	tn := reflect.TypeOf(DepotDelete{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

	// create table depot
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
	fmt.Printf("INSERTED: %v\n", depot)

	err = Handle.Delete(&depot)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	err = Handle.DropTables(DepotDelete{})
	if err != nil {
		t.Errorf("failed to drop table %s", tn)
	}
}

// TestCRUDGet
//
// Test CRUD Get
func TestCRUDGet(t *testing.T) {

	type DepotGet struct {
		DepotNum            int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		DepotBay            int       `db:"depot_bay" rgen:"primary_key:"`
		CreateDate          time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
		Region              string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province            string    `db:"province" rgen:"nullable:false;default:AB"`
		Country             string    `db:"country" rgen:"nullable:true;default:CA"`
		NewColumn1          string    `db:"new_column1" rgen:"nullable:false"`
		NewColumn2          int64     `db:"new_column2" rgen:"nullable:false"`
		NewColumn3          float64   `db:"new_column3" rgen:"nullable:false;default:0.0"`
		IntDefaultZero      int       `db:"int_default_zero" rgen:"nullable:false;default:0"`
		IntDefault42        int       `db:"int_default42" rgen:"nullable:false;default:42"`
		FldOne              int       `db:"fld_one" rgen:"nullable:false;default:0;index:idx_depotget_fld_one_fld_two"`
		FldTwo              int       `db:"fld_two" rgen:"nullable:false;default:0;index:idx_depotget_fld_one_fld_two"`
		IntZeroValNoDefault int       `db:"int_zero_val_no_default" rgen:"nullable:false"`
		NonPersistentColumn string    `db:"non_persistent_column" rgen:"-"`
	}

	// determine the table names as per the
	// table creation logic
	tn := reflect.TypeOf(DepotGet{}).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

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
	fmt.Printf("INSERTED: %v\n", depot)

	// create a struct to read into and populate the keys
	depotRead := DepotGet{
		DepotNum: depot.DepotNum,
		DepotBay: depot.DepotBay,
	}

	err = Handle.GetEntity(&depotRead)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	fmt.Println("GetEntity() returned:", depotRead)

	// err = Handle.DropTables(Depot{})
	// if err != nil {
	// 	t.Errorf("failed to drop table %s", tn)
	// }
}

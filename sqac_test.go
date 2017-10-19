package sqac_test

import (
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/1414C/sqac"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// // SessionData contains session management vars
// type SessionData struct {
// 	db  *sqlx.DB
// 	log bool
// }
type dbac struct {
	DB   *sqlx.DB
	Log  bool
	Hndl sqac.PublicDB
}

var (
	dbAccess dbac
	Handle   sqac.PublicDB
)

func TestMain(m *testing.M) {

	// parse flags
	dbFlag := flag.String("db", "pg", "db to connect to")
	logFlag := flag.Bool("l", false, "activate sqac logging")
	flag.Parse()

	// select the db implementation
	switch *dbFlag {
	case "pg":
		pgh := new(sqac.PostgresFlavor)
		Handle = pgh
		db, err := sqac.Open("postgres", "host=127.0.0.1 user=godev dbname=sqlx sslmode=disable password=gogogo123")
		if err != nil {
			log.Fatalf("%s\n", err.Error())
		}
		Handle.SetDB(db)
		if *logFlag {
			Handle.Log(true)
		} else {
			Handle.Log(false)
		}

	case "mysql":
		myh := new(sqac.MySQLFlavor)
		Handle = myh
		db, err := sqac.Open("mysql", "stevem:gogogo123@tcp(192.168.1.50:3306)/jsonddl?charset=utf8&parseTime=True&loc=Local")
		if err != nil {
			log.Fatalf("%s\n", err.Error())
		}
		Handle.SetDB(db)
		Handle.Log(false)

	case "sqlite":
		sqh := new(sqac.SQLiteFlavor)
		Handle = sqh
		db, err := sqac.Open("sqlite3", "testdb.sqlite")
		if err != nil {
			log.Fatalf("%s\n", err.Error())
		}
		Handle.SetDB(db)
		Handle.Log(false)

	case "db2":

	case "go-hdb":

	default:

	}

	// run the tests
	code := m.Run()

	// cleanup

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
}

// TestCreateTables
//
// Create table depot via CreateTables(i ...interface{})
// Verify table creation via ExistsTable(tn string)
// Perform negative validation be checking for non-existant
// 	table "abcdefg" via ExistsTable(tn string)
//
func TestCreateTables(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
		Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:true;"`
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

	// negative verification of ExistsTable(tn string)
	if Handle.ExistsTable("abcdefg") {
		t.Errorf("table %s should not exist - check ExistsTable implementation for db %s", "abcdefg", Handle.GetDBDriverName())
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
		Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:true;"`
		NewColumn1 string    `db:"new_column1" rgen:"nullable:false"`
		NewColumn2 int64     `db:"new_column2" rgen:"nullable:false;default:0"`
		NewColumn3 float64   `db:"new_column3" rgen:"nullable:false;default:0.0"`
	}

	err := Handle.AlterTables(Depot{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestDropTables
//
// Drop table depot via DropTables(i ...interface{})
//
func TestDropTables(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
		Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:true;"`
		NewColumn1 string    `db:"new_column1" rgen:"nullable:false"`
		NewColumn2 int64     `db:"new_column2" rgen:"nullable:false;default:0"`
		NewColumn3 float64   `db:"new_column3" rgen:"nullable:false;default:0.0"`
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

// TestCreateIndex
//
// Create table depot via CreateTables(i ...interface{})
// Create an index not based on model attributes, but
// based on a constructed sqac.IndexInfo struct.
func TestCreateIndex(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
		Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:true;"`
		NewColumn1 string    `db:"new_column1" rgen:"nullable:false"`
		NewColumn2 int64     `db:"new_column2" rgen:"nullable:false;default:0"`
		NewColumn3 float64   `db:"new_column3" rgen:"nullable:false;default:0.0"`
	}

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
		t.Errorf("table %s was not created", tn)
	}

	var indexInfo sqac.IndexInfo
	indexInfo.TableName = tn
	indexInfo.Unique = false
	indexInfo.IndexFields = []string{"province", "country"}
	err = Handle.CreateIndex("idx_province_country", indexInfo)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestExistsIndex
//
// Check to see if index:
// idx_province_country exists on
// table depot.
func TestExistsIndex(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
		Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:true;"`
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

	if !Handle.ExistsIndex(tn, "idx_province_country") {
		t.Errorf("index %s was not found on table %s", "idx_province_country", tn)
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
		Country    string    `db:"country" rgen:"nullable:true;"`
		NewColumn1 string    `db:"new_column1" rgen:"nullable:false"`
		NewColumn2 int64     `db:"new_column2" rgen:"nullable:false;default:0"`
		NewColumn3 float64   `db:"new_column3" rgen:"nullable:false;default:0.0"`
	}

	// ensure table depot exists
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

	if !Handle.ExistsIndex(tn, "idx_province_country") {
		t.Errorf("index %s was not found on table %s", "idx_province_country", tn)
	}

	err = Handle.DropIndex("idx_province_country")
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	if Handle.ExistsIndex(tn, "idx_province_country") {
		t.Errorf("drop of index %s did not succeed on table %s", "idx_province_country", tn)
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
		Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:true;"`
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

	// ensute table exists in db
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

func TestDestructiveReset(t *testing.T) {

	type Depot struct {
		DepotNum   int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
		Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:true;"`
		NewColumn1 string    `db:"new_column1" rgen:"nullable:false"`
		NewColumn2 int64     `db:"new_column2" rgen:"nullable:false;default:0"`
		NewColumn3 float64   `db:"new_column3" rgen:"nullable:false;default:0.0"`
	}

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
	// 	UInt64Example  uint64    `db:"u_int64_example" rgen:"nullable:false;default:0"`
	// 	UInt32Example  uint32    `db:"u_int32_example" rgen:"nullable:false;default:0"`
	// 	UInt16Example  uint16    `db:"u_int16_example" rgen:"nullable:false;default:0"`
	// 	UInt8Example   uint8     `db:"u_int8_example" rgen:"nullable:false;default:0"`
	// 	Float32Example float32   `db:"float32_example" rgen:"nullable:false;default:0.0"`
	// 	Float64Example float64   `db:"float64_example" rgen:"nullable:false;default:0.0"`
	// 	BoolExample    bool      `db:"bool_example" rgen:"nullable:false;default:false"`
	// 	RuneExample    rune      `db:"rune_example" rgen:"nullable:true"`
	// 	ByteExample    byte      `db:"byte_example" rgen:"nullable:true"`
	// 	Triplet
	// }

	// type Triplet struct {
	// 	TripOne   string `db:"trip_one" rgen:"nullable:false"`
	// 	TripTwo   int64  `db:"trip_two" rgen:"nullable:false;default:0"`
	// 	Tripthree string `db:"trip_three" rgen:"nullable:false"`
	// }

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
		t.Errorf("error reading from table depot - got %s", err.Error())
	}
	fmt.Println("got:", d)

	// here is an example of dealing with an empty null string
	// insert the record without a value in the nullable
	// country field.  the read will fail, due to the string
	// type of the country-field in the target structure.
	// either *string for nullable, or create a local struct
	// that uses nullTypes for nullable fields:
	//   sql.NullBool, sql.NullFloat64, sql.NullInt64, sql.NullString
	// _, err = Handle.Exec("INSERT INTO depot (depot_num, region, province, new_column1) VALUES (DEFAULT, 'YVR','AB', 100) RETURNING depot_num;")
	_, err = Handle.Exec("INSERT INTO depot (depot_num, region, province, country, new_column1) VALUES (DEFAULT, 'YVR','AB','CA', 100) RETURNING depot_num;")
	if err != nil {
		t.Errorf("error inserting into table depot - got %s", err.Error())
	}
	//fmt.Printf("%d rows affected.\n", result.RowsAffected())

	err = Handle.Select(&d, "SELECT * FROM depot;")
	if err != nil {
		t.Errorf("error reading from table depot - got %s", err.Error())
	}
	fmt.Println("got:", d)
}

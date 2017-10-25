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
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

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

	case "mysql":
		myh := new(sqac.MySQLFlavor)
		Handle = myh
		db, err := sqac.Open("mysql", "stevem:gogogo123@tcp(192.168.1.50:3306)/sqlx?charset=utf8&parseTime=True&loc=Local")
		if err != nil {
			log.Fatalf("%s\n", err.Error())
		}
		Handle.SetDB(db)

	case "sqlite":
		sqh := new(sqac.SQLiteFlavor)
		Handle = sqh
		db, err := sqac.Open("sqlite3", "testdb.sqlite")
		if err != nil {
			log.Fatalf("%s\n", err.Error())
		}
		Handle.SetDB(db)

	case "db2":

	case "go-hdb":

	default:

	}

	// detailed logging?
	if *logFlag {
		Handle.Log(true)
	} else {
		Handle.Log(false)
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
		Country    string    `db:"country" rgen:"nullable:false;default:CA"`
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
		Country    string    `db:"country" rgen:"nullable:false;default:CA"`
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
		Country    string    `db:"country" rgen:"nullable:false;default:CA"`
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

	err = Handle.DropIndex(tn, "idx_province_country")
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
		Country    string    `db:"country" rgen:"nullable:true;default:CA"`
		NewColumn1 string    `db:"new_column1" rgen:"nullable:false"`
		NewColumn2 int64     `db:"new_column2" rgen:"nullable:false"`
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
		Province   string    `db:"province" rgen:"nullable:false;default:AB"`
		Country    string    `db:"country" rgen:"nullable:true;"`    // nullable
		NewColumn1 string    `db:"new_column1" rgen:"nullable:true"` // nullable
		NewColumn2 int64     `db:"new_column2" rgen:"nullable:true"` // nullable
		NewColumn3 float64   `db:"new_column3" rgen:"nullable:true"` // nullable
		Active     bool      `db:"active" rgen:"nullable:true"`      // nullable
	}

	// drop and recreate table depot via DestructiveReset
	err := Handle.DestructiveResetTables(Depot{})
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
		insQuery = "INSERT INTO depot (depot_num, region, province) VALUES (DEFAULT, 'YVR','AB');"
	case "sqlite3":
		insQuery = "INSERT INTO depot (region, province) VALUES ('YVR','AB');"
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

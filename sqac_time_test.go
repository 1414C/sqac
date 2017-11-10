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
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	// dbAccess dbac
	Handle sqac.PublicDB
)

func TestMain(m *testing.M) {

	// parse flags
	dbFlag := flag.String("db", "pg", "db to connect to")
	logFlag := flag.Bool("l", false, "activate sqac logging")
	flag.Parse()

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
		cs = ""
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

// TestCRUDCreate
//
// Test CRUD Create
func TestCRUDCreate(t *testing.T) {

	type DepotCreate struct {
		DepotNum            int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
		DepotBay            int       `db:"depot_bay" rgen:"primary_key:"`
		CreateDate          time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
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

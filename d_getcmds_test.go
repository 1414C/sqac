package sqac_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/1414C/sqac"
	"github.com/1414C/sqac/common"
)

// TestCRUDGetEntitiesWithCommandsOpenSelect
//
// Test CRUD GetSet
// call with no parameters and no commands
func TestCRUDGetEntitiesWithCommandsOpenSelect(t *testing.T) {

	// type GetCmdTest struct {
	// 	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	// 	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	// 	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:unique"`
	// 	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	// 	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	// 	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	// 	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	// 	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	// 	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	// 	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
	// }

	// determine the table names as per the table creation logic
	tn := common.GetTableName(GetCmdTest{})

	// drop table getcmdtest
	err := Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table getcmdtest
	err = Handle.CreateTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table getcmdtest exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create

	// create new records via the CRUD Create call
	for i := 0; i < 8; i++ {

		rec := GetCmdTest{
			FldOneInt:     i,
			FldTwoString:  "string",
			FldThreeFloat: (float64(i) * 2.356),
			FldFourBool:   true,
		}

		if i%2 == 0 {
			f5 := "string_value"
			rec.FldFiveString = &f5
			f6 := (float64(i) * 4.8783)
			rec.FldSixFloat = &f6
			f7 := false
			rec.FldSevenBool = &f7
		}

		// create a record
		err = Handle.Create(&rec)
		if err != nil {
			t.Errorf(err.Error())
		}
	}

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	result, err := Handle.GetEntitiesWithCommands(recRead, nil, nil)
	switch result.(type) {
	case []GetCmdTest:
		recRead = result.([]GetCmdTest)
		if len(recRead) != 8 {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOpenSelect: expected 8 records, got: %v", len(recRead))
		}
	case uint64:
		// valid result, but a fail in this case
	case int64:
		// possible result, but a fail in this case
	default:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsOpenSelect")
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesWithCommandsSelectUint
//
// Test CRUD GetSet
// call with single parameter (id = 4) and no commands
func TestCRUDGetEntitiesWithCommandsSelectUint(t *testing.T) {

	// type GetCmdTest struct {
	// 	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	// 	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	// 	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:unique"`
	// 	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	// 	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	// 	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	// 	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	// 	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	// 	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	// 	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
	// }

	// determine the table names as per the table creation logic
	tn := common.GetTableName(GetCmdTest{})

	// drop table getcmdtest
	err := Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table getcmdtest
	err = Handle.CreateTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table getcmdtest exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create new records via the CRUD Create call
	for i := 0; i < 8; i++ {

		rec := GetCmdTest{
			FldOneInt:     i,
			FldTwoString:  "string",
			FldThreeFloat: (float64(i) * 2.356),
			FldFourBool:   true,
		}

		if i%2 == 0 {
			f5 := "string_value"
			rec.FldFiveString = &f5
			f6 := (float64(i) * 4.8783)
			rec.FldSixFloat = &f6
			f7 := false
			rec.FldSevenBool = &f7
		}

		// create a record
		err = Handle.Create(&rec)
		if err != nil {
			t.Errorf(err.Error())
		}
	}

	// create a slice to read into
	recRead := []GetCmdTest{}

	p := sqac.GetParam{
		FieldName:    "id",
		Operand:      "=",
		ParamValue:   90000004,
		NextOperator: "",
	}

	pa := []sqac.GetParam{}
	pa = append(pa, p)

	// call with no parameters and no commands
	result, err := Handle.GetEntitiesWithCommands(recRead, pa, nil)
	if err != nil {
		fmt.Println()
		log.Println("GetEntitiesWithCommands error:", err)
		fmt.Println()
	}
	switch result.(type) {
	case []GetCmdTest:
		recRead = result.([]GetCmdTest)
		if len(recRead) != 1 {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsSelectUint: expected 1 records, got: %v", len(recRead))
		}
		if len(recRead) > 0 {
			if recRead[0].ID != 90000004 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsSelectUint: expected 1 record with key ID == 90000004, got: %v", recRead[0].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsSelectUint: expected 1 records, got: %v", len(recRead))
		}
	case uint64:
		// valid result, but a fail in this case
	case int64:
		// possible result, but a fail in this case
	default:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsSelectUint")
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesWithCommandsSelectString
//
// Test CRUD GetSet
// call with single parameter fld_two_string="Record Two",and no commands
func TestCRUDGetEntitiesWithCommandsSelectString(t *testing.T) {

	// type GetCmdTest struct {
	// 	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	// 	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	// 	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:unique"`
	// 	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	// 	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	// 	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	// 	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	// 	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	// 	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	// 	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
	// }

	// determine the table names as per the table creation logic
	tn := common.GetTableName(GetCmdTest{})

	// drop table getcmdtest
	err := Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table getcmdtest
	err = Handle.CreateTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table getcmdtest exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create the test records
	createGetCmdTestRecs(t)

	// set a parameter for id
	p := sqac.GetParam{
		FieldName:    "fld_two_string",
		Operand:      "=",
		ParamValue:   "CCCCCC",
		NextOperator: "",
	}

	pa := []sqac.GetParam{}
	pa = append(pa, p)

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	result, err := Handle.GetEntitiesWithCommands(recRead, pa, nil)
	switch result.(type) {
	case []GetCmdTest:
		recRead = result.([]GetCmdTest)
		if len(recRead) != 1 {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsSelectString: expected 1 records, got: %v", len(recRead))
		}
		if len(recRead) > 0 {
			if recRead[0].ID != 90000002 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsSelectString: expected 1 record with key ID == 90000002, got: %v", recRead[0].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsSelectString: expected 1 records, got: %v", len(recRead))
		}
	case uint64:
		// valid result, but a fail in this case
	case int64:
		// possible result, but a fail in this case
	default:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsSelectString")
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesWithCommandsSelectCount
//
// Test CRUD GetSet
// call with command /$count
func TestCRUDGetEntitiesWithCommandsSelectCount(t *testing.T) {

	// type GetCmdTest struct {
	// 	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	// 	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	// 	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:unique"`
	// 	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	// 	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	// 	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	// 	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	// 	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	// 	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	// 	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
	// }

	// determine the table names as per the table creation logic
	tn := common.GetTableName(GetCmdTest{})

	// drop table getcmdtest
	err := Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table getcmdtest
	err = Handle.CreateTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table getcmdtest exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create the test records
	createGetCmdTestRecs(t)

	// // set a parameter for id
	// p := common.GetParam{
	// 	FieldName:    "fld_two_string",
	// 	Operand:      "=",
	// 	ParamValue:   "Record Two",
	// 	NextOperator: "",
	// }

	// pa := []common.GetParam{}
	// pa = append(pa, p)

	// set a selection limit = 4
	cmdMap := make(map[string]interface{})
	cmdMap["count"] = nil

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	result, err := Handle.GetEntitiesWithCommands(recRead, nil, cmdMap)
	switch result.(type) {
	case []GetCmdTest:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsSelectCount: got wrong type")
	case uint64:
		c, ok := result.(uint64)
		if c != 8 || !ok {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsSelectCount: expected 8 records: got %v", result)
		}
		// valid result
	case int64:
		// sort-of-valid
	default:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsSelectCount: got wrong type")
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesWithCommandsLimit
//
// Test CRUD GetSet
// call with command /$limit=4
func TestCRUDGetEntitiesWithCommandsLimit(t *testing.T) {

	// type GetCmdTest struct {
	// 	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	// 	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	// 	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:unique"`
	// 	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	// 	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	// 	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	// 	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	// 	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	// 	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	// 	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
	// }

	// determine the table names as per the table creation logic
	tn := common.GetTableName(GetCmdTest{})

	// drop table getcmdtest
	err := Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table getcmdtest
	err = Handle.CreateTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table getcmdtest exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create the test records
	createGetCmdTestRecs(t)

	// // set a parameter for id
	// p := common.GetParam{
	// 	FieldName:    "fld_two_string",
	// 	Operand:      "=",
	// 	ParamValue:   "Record Two",
	// 	NextOperator: "",
	// }

	// pa := []common.GetParam{}
	// pa = append(pa, p)

	// set a selection limit = 4
	cmdMap := make(map[string]interface{})
	cmdMap["limit"] = "4"

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters command $limit=4
	result, err := Handle.GetEntitiesWithCommands(recRead, nil, cmdMap)
	switch result.(type) {
	case []GetCmdTest:
		recRead = result.([]GetCmdTest)
		if len(recRead) != 4 {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsLimit: expected 4 records, got: %v", len(recRead))
		}
	case uint64:
		// valid result, but a fail in this case
	case int64:
		// possible result, but a fail in this case
	default:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsLimit")
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesWithCommandsLimitDesc
//
// Test CRUD GetSet
// call with command /$limit=4$desc
func TestCRUDGetEntitiesWithCommandsLimitDesc(t *testing.T) {

	// type GetCmdTest struct {
	// 	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	// 	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	// 	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:unique"`
	// 	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	// 	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	// 	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	// 	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	// 	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	// 	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	// 	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
	// }

	// determine the table names as per the table creation logic
	tn := common.GetTableName(GetCmdTest{})

	// drop table getcmdtest
	err := Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table getcmdtest
	err = Handle.CreateTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table getcmdtest exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create the test records
	createGetCmdTestRecs(t)

	// // set a parameter for id
	// p := common.GetParam{
	// 	FieldName:    "fld_two_string",
	// 	Operand:      "=",
	// 	ParamValue:   "Record Two",
	// 	NextOperator: "",
	// }

	// pa := []common.GetParam{}
	// pa = append(pa, p)

	// set a selection limit = 4
	cmdMap := make(map[string]interface{})
	cmdMap["limit"] = "4"
	cmdMap["desc"] = nil

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	result, err := Handle.GetEntitiesWithCommands(recRead, nil, cmdMap)
	switch result.(type) {
	case []GetCmdTest:
		recRead = result.([]GetCmdTest)
		if len(recRead) != 4 {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsLimitDesc: expected 4 records, got: %v", len(recRead))
		}
		if len(recRead) > 0 {
			if recRead[0].ID != 90000007 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsLimitDesc: expected result[0] record with key ID == 90000007, got: %v", recRead[0].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsLimitDesc: expected 4 records, got: %v", len(recRead))
		}
	case uint64:
		// valid result, but a fail in this case
	case int64:
		// possible result, but a fail in this case
	default:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsLimitDesc")
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesWithCommandsLimitAsc
//
// Test CRUD GetSet
// call with command /$limit=4$asc
func TestCRUDGetEntitiesWithCommandsLimitAsc(t *testing.T) {

	// type GetCmdTest struct {
	// 	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	// 	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	// 	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:unique"`
	// 	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	// 	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	// 	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	// 	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	// 	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	// 	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	// 	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
	// }

	// determine the table names as per the table creation logic
	tn := common.GetTableName(GetCmdTest{})

	// drop table getcmdtest
	err := Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table getcmdtest
	err = Handle.CreateTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table getcmdtest exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create the test records
	createGetCmdTestRecs(t)

	// // set a parameter for id
	// p := common.GetParam{
	// 	FieldName:    "fld_two_string",
	// 	Operand:      "=",
	// 	ParamValue:   "Record Two",
	// 	NextOperator: "",
	// }

	// pa := []common.GetParam{}
	// pa = append(pa, p)

	// set a selection limit = 4
	cmdMap := make(map[string]interface{})
	cmdMap["limit"] = "4"
	cmdMap["asc"] = nil

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	result, err := Handle.GetEntitiesWithCommands(recRead, nil, cmdMap)
	switch result.(type) {
	case []GetCmdTest:
		recRead = result.([]GetCmdTest)
		if len(recRead) != 4 {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsLimitAsc: expected 4 records, got: %v", len(recRead))
		}
		if len(recRead) > 0 {
			if recRead[0].ID != 90000000 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsLimitAsc: expected result[0] record with key ID == 90000000, got: %v", recRead[0].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsLimitAsc: expected 4 records, got: %v", len(recRead))
		}
	case uint64:
		// valid result, but a fail in this case
	case int64:
		// possible result, but a fail in this case
	default:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsLimitAsc")
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesWithCommandsOffset
//
// Test CRUD GetSet
// call with command /$offset=2
func TestCRUDGetEntitiesWithCommandsOffset(t *testing.T) {

	// type GetCmdTest struct {
	// 	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	// 	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	// 	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:unique"`
	// 	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	// 	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	// 	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	// 	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	// 	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	// 	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	// 	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
	// }

	// determine the table names as per the table creation logic
	tn := common.GetTableName(GetCmdTest{})

	// drop table getcmdtest
	err := Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table getcmdtest
	err = Handle.CreateTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table getcmdtest exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create the test records
	createGetCmdTestRecs(t)

	// // set a parameter for id
	// p := common.GetParam{
	// 	FieldName:    "fld_two_string",
	// 	Operand:      "=",
	// 	ParamValue:   "Record Two",
	// 	NextOperator: "",
	// }

	// pa := []common.GetParam{}
	// pa = append(pa, p)

	// set an offset of 2
	cmdMap := make(map[string]interface{})
	cmdMap["offset"] = 2

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	result, err := Handle.GetEntitiesWithCommands(recRead, nil, cmdMap)
	switch result.(type) {
	case []GetCmdTest:
		recRead = result.([]GetCmdTest)
		if len(recRead) != 6 {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOffset: expected 6 records, got: %v", len(recRead))
		}
		if len(recRead) > 0 {
			if recRead[0].ID != 90000002 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsOffset: expected result[0] record with key ID == 90000002, got: %v", recRead[0].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOffset: expected 6 records, got: %v", len(recRead))
		}
	case uint64:
		// valid result, but a fail in this case
	case int64:
		// possible result, but a fail in this case
	default:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsOffset")
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesWithCommandsOffsetDesc
//
// Test CRUD GetSet
// call with command /$offset=2$desc
func TestCRUDGetEntitiesWithCommandsOffsetDesc(t *testing.T) {

	// type GetCmdTest struct {
	// 	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	// 	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	// 	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:unique"`
	// 	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	// 	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	// 	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	// 	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	// 	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	// 	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	// 	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
	// }

	// determine the table names as per the table creation logic
	tn := common.GetTableName(GetCmdTest{})

	// drop table getcmdtest
	err := Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table getcmdtest
	err = Handle.CreateTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table getcmdtest exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create the test records
	createGetCmdTestRecs(t)

	// // set a parameter for id
	// p := common.GetParam{
	// 	FieldName:    "fld_two_string",
	// 	Operand:      "=",
	// 	ParamValue:   "Record Two",
	// 	NextOperator: "",
	// }

	// pa := []common.GetParam{}
	// pa = append(pa, p)

	// set an offset of 2
	cmdMap := make(map[string]interface{})
	cmdMap["offset"] = 2
	cmdMap["desc"] = nil

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	result, err := Handle.GetEntitiesWithCommands(recRead, nil, cmdMap)
	switch result.(type) {
	case []GetCmdTest:
		recRead = result.([]GetCmdTest)
		if len(recRead) != 6 {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOffsetDesc: expected 6 records, got: %v", len(recRead))
		}
		if len(recRead) > 0 {
			if recRead[0].ID != 90000005 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsOffsetDesc: expected result[0] record with key ID == 90000005, got: %v", recRead[0].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOffsetDesc: expected 6 records, got: %v", len(recRead))
		}
	case uint64:
		// valid result, but a fail in this case
	case int64:
		// possible result, but a fail in this case
	default:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsOffsetDesc")
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesWithCommandsOffsetLimit
//
// Test CRUD GetSet
// call with command /$offset=2$limit=4
func TestCRUDGetEntitiesWithCommandsOffsetLimit(t *testing.T) {

	// type GetCmdTest struct {
	// 	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	// 	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	// 	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:unique"`
	// 	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	// 	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	// 	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	// 	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	// 	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	// 	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	// 	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
	// }

	// determine the table names as per the table creation logic
	tn := common.GetTableName(GetCmdTest{})

	// drop table getcmdtest
	err := Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table getcmdtest
	err = Handle.CreateTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table getcmdtest exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create the test records
	createGetCmdTestRecs(t)

	// // set a parameter for id
	// p := common.GetParam{
	// 	FieldName:    "fld_two_string",
	// 	Operand:      "=",
	// 	ParamValue:   "Record Two",
	// 	NextOperator: "",
	// }

	// pa := []common.GetParam{}
	// pa = append(pa, p)

	// set an offset of 2, limit of 4
	cmdMap := make(map[string]interface{})
	cmdMap["offset"] = 2
	cmdMap["limit"] = 4

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	result, err := Handle.GetEntitiesWithCommands(recRead, nil, cmdMap)
	switch result.(type) {
	case []GetCmdTest:
		recRead = result.([]GetCmdTest)
		if len(recRead) != 4 {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOffsetLimit: expected 4 records, got: %v", len(recRead))
		}
		if len(recRead) > 0 {
			if recRead[0].ID != 90000002 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsOffsetLimit: expected result[0] record with key ID == 90000002, got: %v", recRead[0].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOffsetLimit: expected 4 records, got: %v", len(recRead))
		}
	case uint64:
		// valid result, but a fail in this case
	case int64:
		// possible result, but a fail in this case
	default:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsOffsetLimit")
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesWithCommandsOffsetLimitDesc
//
// Test CRUD GetSet
// call with command /$offset=2$limit=4$desc
func TestCRUDGetEntitiesWithCommandsOffsetLimitDesc(t *testing.T) {

	// type GetCmdTest struct {
	// 	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	// 	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	// 	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:unique"`
	// 	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	// 	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	// 	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	// 	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	// 	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	// 	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	// 	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
	// }

	// determine the table names as per the table creation logic
	tn := common.GetTableName(GetCmdTest{})

	// drop table getcmdtest
	err := Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table getcmdtest
	err = Handle.CreateTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table getcmdtest exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create the test records
	createGetCmdTestRecs(t)

	// // set a parameter for id
	// p := common.GetParam{
	// 	FieldName:    "fld_two_string",
	// 	Operand:      "=",
	// 	ParamValue:   "Record Two",
	// 	NextOperator: "",
	// }

	// pa := []common.GetParam{}
	// pa = append(pa, p)

	// set an offset of 2, limit of 4, order by id descending
	cmdMap := make(map[string]interface{})
	cmdMap["offset"] = 2
	cmdMap["limit"] = 4
	cmdMap["desc"] = nil

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	result, err := Handle.GetEntitiesWithCommands(recRead, nil, cmdMap)
	switch result.(type) {
	case []GetCmdTest:
		recRead = result.([]GetCmdTest)
		if len(recRead) != 4 {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOffsetLimitDesc: expected 4 records, got: %v", len(recRead))
		}
		if len(recRead) > 0 {
			if recRead[0].ID != 90000005 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsOffsetLimitDesc: expected result[0] record with key ID == 90000005, got: %v", recRead[0].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOffsetLimitDesc: expected 4 records, got: %v", len(recRead))
		}
	case uint64:
		// valid result, but a fail in this case
	case int64:
		// possible result, but a fail in this case
	default:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsOffsetLimitDesc")
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesWithCommandstOrderBy
//
// Test CRUD GetSet
// call with command /$orderby=name
func TestCRUDGetEntitiesWithCommandsOrderBy(t *testing.T) {

	// type GetCmdTest struct {
	// 	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	// 	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	// 	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:unique"`
	// 	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	// 	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	// 	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	// 	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	// 	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	// 	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	// 	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
	// }

	// determine the table names as per the table creation logic
	tn := common.GetTableName(GetCmdTest{})

	// drop table getcmdtest
	err := Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table getcmdtest
	err = Handle.CreateTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table getcmdtest exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create the test records
	createGetCmdTestRecs(t)

	// // set a parameter for id
	// p := common.GetParam{
	// 	FieldName:    "fld_two_string",
	// 	Operand:      "=",
	// 	ParamValue:   "Record Two",
	// 	NextOperator: "",
	// }

	// pa := []common.GetParam{}
	// pa = append(pa, p)

	// set $orderby=name
	cmdMap := make(map[string]interface{})
	cmdMap["orderby"] = "fld_two_string"

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	result, err := Handle.GetEntitiesWithCommands(recRead, nil, cmdMap)
	switch result.(type) {
	case []GetCmdTest:
		recRead = result.([]GetCmdTest)
		if len(recRead) != 8 {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderBy: expected 8 records, got: %v", len(recRead))
		}
		if len(recRead) > 0 {
			if recRead[0].ID != 90000000 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderBy: expected result[0] record with key ID == 90000000 got: %v", recRead[0].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderBy: expected 8 records, got: %v", len(recRead))
		}
	case uint64:
		// valid result, but a fail in this case
	case int64:
		// possible result, but a fail in this case
	default:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderBy")
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesWithCommandsOrderByDesc
//
// Test CRUD GetSet
// call with command /$orderby=name$desc
func TestCRUDGetEntitiesWithCommandsOrderByDesc(t *testing.T) {

	// type GetCmdTest struct {
	// 	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	// 	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	// 	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:unique"`
	// 	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	// 	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	// 	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	// 	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	// 	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	// 	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	// 	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
	// }

	// determine the table names as per the table creation logic
	tn := common.GetTableName(GetCmdTest{})

	// drop table getcmdtest
	err := Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table getcmdtest
	err = Handle.CreateTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table getcmdtest exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create the test records
	createGetCmdTestRecs(t)

	// // set a parameter for id
	// p := common.GetParam{
	// 	FieldName:    "fld_two_string",
	// 	Operand:      "=",
	// 	ParamValue:   "Record Two",
	// 	NextOperator: "",
	// }

	// pa := []common.GetParam{}
	// pa = append(pa, p)

	// set $orderby=name
	cmdMap := make(map[string]interface{})
	cmdMap["orderby"] = "fld_two_string"
	cmdMap["desc"] = nil

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	result, err := Handle.GetEntitiesWithCommands(recRead, nil, cmdMap)
	switch result.(type) {
	case []GetCmdTest:
		recRead = result.([]GetCmdTest)
		if len(recRead) != 8 {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByDesc: expected 8 records, got: %v", len(recRead))
		}
		if len(recRead) > 0 {
			if recRead[0].ID != 90000007 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByDesc: expected result[0] record with key ID == 90000007 got: %v", recRead[0].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByDesc: expected 8 records, got: %v", len(recRead))
		}
	case uint64:
		// valid result, but a fail in this case
	case int64:
		// possible result, but a fail in this case
	default:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByDesc")
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesWithCommandsParamOrderBy
//
// Test CRUD GetSet
// call with command /$orderby=name /parameters: id > 90000002
func TestCRUDGetEntitiesWithCommandsParamOrderBy(t *testing.T) {

	// type GetCmdTest struct {
	// 	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	// 	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	// 	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:unique"`
	// 	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	// 	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	// 	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	// 	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	// 	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	// 	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	// 	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
	// }

	// determine the table names as per the table creation logic
	tn := common.GetTableName(GetCmdTest{})

	// drop table getcmdtest
	err := Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table getcmdtest
	err = Handle.CreateTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table getcmdtest exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create the test records
	createGetCmdTestRecs(t)

	// set a parameter for id
	p := sqac.GetParam{
		FieldName:    "id",
		Operand:      ">",
		ParamValue:   90000002,
		NextOperator: "",
	}

	pa := []sqac.GetParam{}
	pa = append(pa, p)

	// set $orderby=name
	cmdMap := make(map[string]interface{})
	cmdMap["orderby"] = "fld_two_string"
	cmdMap["asc"] = nil

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	result, err := Handle.GetEntitiesWithCommands(recRead, pa, cmdMap)
	switch result.(type) {
	case []GetCmdTest:
		recRead = result.([]GetCmdTest)
		if len(recRead) != 5 {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsParamOrderBy: expected 5 records, got: %v", len(recRead))
		}
		if len(recRead) > 0 {
			if recRead[0].ID != 90000003 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsParamOrderBy: expected result[0] record with key ID == 90000003 got: %v", recRead[0].ID)
			}
			if recRead[4].ID != 90000007 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsParamOrderBy: expected result[0] record with key ID == 90000007 got: %v", recRead[4].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsParamOrderBy: expected 5 records, got: %v", len(recRead))
		}
	case uint64:
		// valid result, but a fail in this case
	case int64:
		// possible result, but a fail in this case
	default:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsParamOrderBy")
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesWithCommandsParamOrderByDesc
//
// Test CRUD GetSet
// call with command /$orderby=name /parameters: id > 90000002
func TestCRUDGetEntitiesWithCommandsParamOrderByDesc(t *testing.T) {

	// type GetCmdTest struct {
	// 	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	// 	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	// 	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:unique"`
	// 	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	// 	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	// 	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	// 	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	// 	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	// 	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	// 	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
	// }

	// determine the table names as per the table creation logic
	tn := common.GetTableName(GetCmdTest{})

	// drop table getcmdtest
	err := Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table getcmdtest
	err = Handle.CreateTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table getcmdtest exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create the test records
	createGetCmdTestRecs(t)

	// set a parameter for id
	p := sqac.GetParam{
		FieldName:    "id",
		Operand:      ">",
		ParamValue:   90000002,
		NextOperator: "",
	}

	pa := []sqac.GetParam{}
	pa = append(pa, p)

	// set $orderby=name
	cmdMap := make(map[string]interface{})
	cmdMap["orderby"] = "fld_two_string"
	cmdMap["desc"] = nil

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	result, err := Handle.GetEntitiesWithCommands(recRead, pa, cmdMap)
	switch result.(type) {
	case []GetCmdTest:
		recRead = result.([]GetCmdTest)
		if len(recRead) != 5 {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsParamOrderByDesc: expected 5 records, got: %v", len(recRead))
		}
		if len(recRead) > 0 {
			if recRead[0].ID != 90000007 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsParamOrderByDesc: expected result[0] record with key ID == 90000007 got: %v", recRead[0].ID)
			}
			if recRead[4].ID != 90000003 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsParamOrderByDesc: expected result[0] record with key ID == 90000003 got: %v", recRead[4].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsParamOrderByDesc: expected 5 records, got: %v", len(recRead))
		}
	case uint64:
		// valid result, but a fail in this case
	case int64:
		// possible result, but a fail in this case
	default:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsParamOrderByDesc")
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesWithCommandsOrderByDescLimit
//
// Test CRUD GetSet
// call with command /$orderby=name$desc$limit=3
func TestCRUDGetEntitiesWithCommandsOrderByDescLimit(t *testing.T) {

	// type GetCmdTest struct {
	// 	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	// 	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	// 	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:unique"`
	// 	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	// 	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	// 	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	// 	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	// 	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	// 	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	// 	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
	// }

	// determine the table names as per the table creation logic
	tn := common.GetTableName(GetCmdTest{})

	// drop table getcmdtest
	err := Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table getcmdtest
	err = Handle.CreateTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table getcmdtest exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create the test records
	createGetCmdTestRecs(t)

	// // set a parameter for id
	// p := common.GetParam{
	// 	FieldName:    "fld_two_string",
	// 	Operand:      "=",
	// 	ParamValue:   "Record Two",
	// 	NextOperator: "",
	// }

	// pa := []common.GetParam{}
	// pa = append(pa, p)

	// set $orderby=name
	cmdMap := make(map[string]interface{})
	cmdMap["orderby"] = "fld_two_string"
	cmdMap["desc"] = nil
	cmdMap["limit"] = 3

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	result, err := Handle.GetEntitiesWithCommands(recRead, nil, cmdMap)
	switch result.(type) {
	case []GetCmdTest:
		recRead = result.([]GetCmdTest)
		if len(recRead) != 3 {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByDescLimit: expected 3 records, got: %v", len(recRead))
		}
		if len(recRead) > 0 {
			if recRead[0].ID != 90000007 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByDescLimit: expected result[0] record with key ID == 90000007 got: %v", recRead[0].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByDescLimit: expected 3 records, got: %v", len(recRead))
		}
		if len(recRead) == 3 {
			if recRead[2].ID != 90000005 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByDescLimit: expected result[2] record with key ID == 90000005 got: %v", recRead[2].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByDescLimit: expected 3 records, got: %v", len(recRead))
		}
	case uint64:
		// valid result, but a fail in this case
	case int64:
		// possible result, but a fail in this case
	default:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByDescLimit")
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesWithCommandsOrderByDescLimitOffset
//
// Test CRUD GetSet
// call with command /$orderby=name$desc$limit=3$offset=2
func TestCRUDGetEntitiesWithCommandsOrderByDescLimitOffset(t *testing.T) {

	// type GetCmdTest struct {
	// 	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	// 	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	// 	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:unique"`
	// 	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	// 	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	// 	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	// 	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	// 	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	// 	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	// 	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
	// }

	// determine the table names as per the table creation logic
	tn := common.GetTableName(GetCmdTest{})

	// drop table getcmdtest
	err := Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table getcmdtest
	err = Handle.CreateTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table getcmdtest exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create the test records
	createGetCmdTestRecs(t)

	// // set a parameter for id
	// p := common.GetParam{
	// 	FieldName:    "fld_two_string",
	// 	Operand:      "=",
	// 	ParamValue:   "Record Two",
	// 	NextOperator: "",
	// }

	// pa := []common.GetParam{}
	// pa = append(pa, p)

	// set $orderby=name$descending$limit=3$offset=2
	cmdMap := make(map[string]interface{})
	cmdMap["orderby"] = "fld_two_string"
	cmdMap["desc"] = nil
	cmdMap["limit"] = 3
	cmdMap["offset"] = 2

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	result, err := Handle.GetEntitiesWithCommands(recRead, nil, cmdMap)
	switch result.(type) {
	case []GetCmdTest:
		recRead = result.([]GetCmdTest)
		if len(recRead) != 3 {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByDescLimitOffset: expected 3 records, got: %v", len(recRead))
		}
		if len(recRead) > 0 {
			if recRead[0].ID != 90000005 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByDescLimitOffset: expected result[0] record with key ID == 90000005 got: %v", recRead[0].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByDescLimitOffset: expected 3 records, got: %v", len(recRead))
		}
		if len(recRead) == 3 {
			if recRead[2].ID != 90000003 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByDescLimitOffset: expected result[2] record with key ID == 90000003 got: %v", recRead[2].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByDescLimitOffset: expected 3 records, got: %v", len(recRead))
		}
	case uint64:
		// valid result, but a fail in this case
	case int64:
		// possible result, but a fail in this case
	default:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByDescLimitOffset")
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesWithCommandsOrderByAscLimitOffset
//
// Test CRUD GetSet
// call with command /$orderby=name$asc$limit=3$offset=2
func TestCRUDGetEntitiesWithCommandsOrderByAscLimitOffset(t *testing.T) {

	// type GetCmdTest struct {
	// 	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	// 	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	// 	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:unique"`
	// 	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	// 	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	// 	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	// 	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	// 	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	// 	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	// 	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
	// }

	// determine the table names as per the table creation logic
	tn := common.GetTableName(GetCmdTest{})

	// drop table getcmdtest
	err := Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table getcmdtest
	err = Handle.CreateTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table getcmdtest exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create the test records
	createGetCmdTestRecs(t)

	// // set a parameter for id
	// p := common.GetParam{
	// 	FieldName:    "fld_two_string",
	// 	Operand:      "=",
	// 	ParamValue:   "Record Two",
	// 	NextOperator: "",
	// }

	// pa := []common.GetParam{}
	// pa = append(pa, p)

	// set $orderby=name$descending$limit=3$offset=2
	cmdMap := make(map[string]interface{})
	cmdMap["orderby"] = "fld_two_string"
	cmdMap["asc"] = nil
	cmdMap["limit"] = 3
	cmdMap["offset"] = 2

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	result, err := Handle.GetEntitiesWithCommands(recRead, nil, cmdMap)
	switch result.(type) {
	case []GetCmdTest:
		recRead = result.([]GetCmdTest)
		if len(recRead) != 3 {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByAscLimitOffset: expected 3 records, got: %v", len(recRead))
		}
		if len(recRead) > 0 {
			if recRead[0].ID != 90000002 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByAscLimitOffset: expected result[0] record with key ID == 90000002 got: %v", recRead[0].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByAscLimitOffset: expected 3 records, got: %v", len(recRead))
		}
		if len(recRead) == 3 {
			if recRead[2].ID != 90000004 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByAscLimitOffset: expected result[2] record with key ID == 90000004 got: %v", recRead[2].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByAscLimitOffset: expected 3 records, got: %v", len(recRead))
		}
	case uint64:
		// valid result, but a fail in this case
	case int64:
		// possible result, but a fail in this case
	default:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsOrderByAscLimitOffset")
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesWithCommandsSelectOrderBy
//
// Test CRUD GetSet
// call with command /$orderby=name
func TestCRUDGetEntitiesWithCommandsTestOffsetOrderByAscLimitOffset(t *testing.T) {

	// type GetCmdTest struct {
	// 	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	// 	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	// 	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:unique"`
	// 	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	// 	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	// 	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	// 	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	// 	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	// 	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	// 	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
	// }

	// determine the table names as per the table creation logic
	tn := common.GetTableName(GetCmdTest{})

	// drop table getcmdtest
	err := Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// create table getcmdtest
	err = Handle.CreateTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table getcmdtest exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s does not exist", tn)
	}

	// create the test records
	createGetCmdTestRecs(t)

	// // set a parameter for id
	// p := common.GetParam{
	// 	FieldName:    "fld_two_string",
	// 	Operand:      "=",
	// 	ParamValue:   "Record Two",
	// 	NextOperator: "",
	// }

	// pa := []common.GetParam{}
	// pa = append(pa, p)

	// set $orderby=name$descending$limit=3$offset=2
	cmdMap := make(map[string]interface{})
	cmdMap["orderby"] = "fld_two_string"
	cmdMap["asc"] = nil
	cmdMap["limit"] = 3
	cmdMap["offset"] = 2

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	result, err := Handle.GetEntitiesWithCommands(recRead, nil, cmdMap)
	switch result.(type) {
	case []GetCmdTest:
		recRead = result.([]GetCmdTest)
		if len(recRead) != 3 {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsTestOrderByAscLimitOffset: expected 3 records, got: %v", len(recRead))
		}
		if len(recRead) > 0 {
			if recRead[0].ID != 90000002 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsTestOrderByAscLimitOffset: expected result[0] record with key ID == 90000002 got: %v", recRead[0].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsTestOrderByAscLimitOffset: expected 3 records, got: %v", len(recRead))
		}
		if len(recRead) == 3 {
			if recRead[2].ID != 90000004 {
				t.Errorf("error: TestCRUDGetEntitiesWithCommandsTestOrderByAscLimitOffset: expected result[2] record with key ID == 90000004 got: %v", recRead[2].ID)
			}
		} else {
			t.Errorf("error: TestCRUDGetEntitiesWithCommandsTestOrderByAscLimitOffset: expected 3 records, got: %v", len(recRead))
		}
	case uint64:
		// valid result, but a fail in this case
	case int64:
		// possible result, but a fail in this case
	default:
		t.Errorf("error: TestCRUDGetEntitiesWithCommandsTestOrderByAscLimitOffset")
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

func createGetCmdTestRecs(t *testing.T) {
	// create new records via the CRUD Create call
	rec := GetCmdTest{
		FldOneInt:     1,
		FldTwoString:  "AAAAAA",
		FldThreeFloat: 10.2922,
		FldFourBool:   true,
	}
	f5 := "Optional String Value"
	rec.FldFiveString = &f5
	f6 := 99.312345
	rec.FldSixFloat = &f6
	f7 := false
	rec.FldSevenBool = &f7

	// create a record
	err := Handle.Create(&rec)
	if err != nil {
		t.Errorf(err.Error())
	}

	// create record 2
	rec = GetCmdTest{
		FldOneInt:     2,
		FldTwoString:  "BBBBBB",
		FldThreeFloat: 20.2762,
		FldFourBool:   true,
	}
	f5 = "Optional String Value Two"
	rec.FldFiveString = &f5
	f6 = 200.12121
	rec.FldSixFloat = &f6
	f7 = false
	rec.FldSevenBool = &f7

	err = Handle.Create(&rec)
	if err != nil {
		t.Errorf(err.Error())
	}

	// create record 3
	rec = GetCmdTest{
		FldOneInt:     3,
		FldTwoString:  "CCCCCC",
		FldThreeFloat: 30.3385,
		FldFourBool:   true,
	}
	f5 = "Optional String Value Three"
	rec.FldFiveString = &f5
	f6 = 300.31313
	rec.FldSixFloat = &f6
	f7 = false
	rec.FldSevenBool = &f7

	err = Handle.Create(&rec)
	if err != nil {
		t.Errorf(err.Error())
	}

	// create record 4
	rec = GetCmdTest{
		FldOneInt:     4,
		FldTwoString:  "DDDDDD",
		FldThreeFloat: 40.75757,
		FldFourBool:   true,
	}
	f5 = "Optional String Value Four"
	rec.FldFiveString = &f5
	f6 = 400.1414
	rec.FldSixFloat = &f6
	f7 = false
	rec.FldSevenBool = &f7

	err = Handle.Create(&rec)
	if err != nil {
		t.Errorf(err.Error())
	}

	// create record 5
	rec = GetCmdTest{
		FldOneInt:     5,
		FldTwoString:  "EEEEEE",
		FldThreeFloat: 50.58585,
		FldFourBool:   true,
	}
	f5 = "Optional String Value Five"
	rec.FldFiveString = &f5
	f6 = 400.1414
	rec.FldSixFloat = &f6
	f7 = false
	rec.FldSevenBool = &f7

	err = Handle.Create(&rec)
	if err != nil {
		t.Errorf(err.Error())
	}

	// create record 6
	rec = GetCmdTest{
		FldOneInt:     6,
		FldTwoString:  "FFFFFF",
		FldThreeFloat: 60.6767,
		FldFourBool:   true,
	}
	f5 = "Optional String Value Six"
	rec.FldFiveString = &f5
	f6 = 600.1616
	rec.FldSixFloat = &f6
	f7 = false
	rec.FldSevenBool = &f7

	err = Handle.Create(&rec)
	if err != nil {
		t.Errorf(err.Error())
	}

	// create record 7
	rec = GetCmdTest{
		FldOneInt:     7,
		FldTwoString:  "GGGGGG",
		FldThreeFloat: 70.73737,
		FldFourBool:   true,
	}
	f5 = "Optional String Value Seven"
	rec.FldFiveString = &f5
	f6 = 700.12224
	rec.FldSixFloat = &f6
	f7 = false
	rec.FldSevenBool = &f7

	err = Handle.Create(&rec)
	if err != nil {
		t.Errorf(err.Error())
	}

	// create record 8
	rec = GetCmdTest{
		FldOneInt:     8,
		FldTwoString:  "HHHHHH",
		FldThreeFloat: 80.85858,
		FldFourBool:   true,
	}
	f5 = "Optional String Value Eight"
	rec.FldFiveString = &f5
	f6 = 800.1818
	rec.FldSixFloat = &f6
	f7 = false
	rec.FldSevenBool = &f7

	err = Handle.Create(&rec)
	if err != nil {
		t.Errorf(err.Error())
	}
}

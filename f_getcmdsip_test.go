package sqac_test

import (
	"testing"

	"github.com/1414C/sqac"
	"github.com/1414C/sqac/common"
)

// TestCRUDGetEntities5OpenSelect
//
// Test CRUD GetSet
// call with no parameters and no commands
func TestCRUDGetEntitiesCPOpenSelect(t *testing.T) {

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
	_, err = Handle.GetEntitiesCP(&recRead, nil, nil)
	if err != nil {
		t.Errorf("error: TestCRUDGetEntitiesCPOpenSelect: %v", err)
	}

	if len(recRead) != 8 {
		t.Errorf("error: TestCRUDGetEntitiesCPOpenSelect: expected 8 records, got: %v", len(recRead))
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesCPSelectUint
//
// Test CRUD GetSet
// call with single parameter (id = 4) and no commands
func TestCRUDGetEntitiesCPSelectUint(t *testing.T) {

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

	// call with a single parameter and no commands
	_, err = Handle.GetEntitiesCP(&recRead, pa, nil)
	if err != nil {
		t.Errorf("error: TestCRUDGetEntitiesCPSelectUint: %v", err)
	}

	if len(recRead) > 0 {
		if recRead[0].ID != 90000004 {
			t.Errorf("error: TestCRUDGetEntitiesCPSelectUint: expected 1 record with key ID == 90000004, got: %v", recRead[0].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPSelectUint: expected 1 records, got: %v", len(recRead))
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesCPSelectString
//
// Test CRUD GetSet
// call with single parameter fld_two_string="Record Two",and no commands
func TestCRUDGetEntitiesCPSelectString(t *testing.T) {

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
	_, err = Handle.GetEntitiesCP(&recRead, pa, nil)
	if err != nil {
		t.Errorf("error: TestCRUDGetEntitiesCPSelectString: %v", err)
	}

	if len(recRead) > 0 {
		if recRead[0].ID != 90000002 {
			t.Errorf("error: TestCRUDGetEntitiesCPSelectString: expected 1 record with key ID == 90000002, got: %v", recRead[0].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPSelectString: expected 1 records, got: %v", len(recRead))
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesCPSelectCount
//
// Test CRUD GetSet
// call with command /$count
func TestCRUDGetEntitiesCPSelectCount(t *testing.T) {

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

	// set a selection limit = 4
	cmdMap := make(map[string]interface{})
	cmdMap["count"] = nil

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	result, err := Handle.GetEntitiesCP(&recRead, nil, cmdMap)
	if err != nil {
		t.Errorf("error: TestCRUDGetEntitiesCPSelectCount: %v", err)
	}

	if result != 8 {
		t.Errorf("error: TestCRUDGetEntitiesCPSelectCount: expected 8 records: got %v", result)
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesCPLimit
//
// Test CRUD GetSet
// call with command /$limit=4
func TestCRUDGetEntitiesCPLimit(t *testing.T) {

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

	// set a selection limit = 4
	cmdMap := make(map[string]interface{})
	cmdMap["limit"] = "4"

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters command $limit=4
	_, err = Handle.GetEntitiesCP(&recRead, nil, cmdMap)
	if err != nil {
		t.Errorf("error: TestCRUDGetEntitiesCPLimit: %v", err)
	}

	if len(recRead) != 4 {
		t.Errorf("error: TestCRUDGetEntitiesCPLimit: expected 4 records, got: %v", len(recRead))
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesCPLimitDesc
//
// Test CRUD GetSet
// call with command /$limit=4$desc
func TestCRUDGetEntitiesCPLimitDesc(t *testing.T) {

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

	// set a selection limit = 4
	cmdMap := make(map[string]interface{})
	cmdMap["limit"] = "4"
	cmdMap["desc"] = nil

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	_, err = Handle.GetEntitiesCP(&recRead, nil, cmdMap)
	if err != nil {
		t.Errorf("error: TestCRUDGetEntitiesCPLimitDesc: %v", err)
	}

	if len(recRead) != 4 {
		t.Errorf("error: TestCRUDGetEntitiesCPLimitDesc: expected 4 records, got: %v", len(recRead))
	}
	if len(recRead) > 0 {
		if recRead[0].ID != 90000007 {
			t.Errorf("error: TestCRUDGetEntitiesCPLimitDesc: expected result[0] record with key ID == 90000007, got: %v", recRead[0].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPLimitDesc: expected 4 records, got: %v", len(recRead))
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesCPLimitAsc
//
// Test CRUD GetSet
// call with command /$limit=4$asc
func TestCRUDGetEntitiesCPLimitAsc(t *testing.T) {

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

	// set a selection limit = 4
	cmdMap := make(map[string]interface{})
	cmdMap["limit"] = "4"
	cmdMap["asc"] = nil

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	_, err = Handle.GetEntitiesCP(&recRead, nil, cmdMap)
	if err != nil {
		t.Errorf("error: TestCRUDGetEntitiesCPLimitAsc: %v", err)
	}
	if len(recRead) != 4 {
		t.Errorf("error: TestCRUDGetEntitiesCPLimitAsc: expected 4 records, got: %v", len(recRead))
	}
	if len(recRead) > 0 {
		if recRead[0].ID != 90000000 {
			t.Errorf("error: TestCRUDGetEntitiesCPLimitAsc: expected result[0] record with key ID == 90000000, got: %v", recRead[0].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPLimitAsc: expected 4 records, got: %v", len(recRead))
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesCPOffset
//
// Test CRUD GetSet
// call with command /$offset=2
func TestCRUDGetEntitiesCPOffset(t *testing.T) {

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

	// set an offset of 2
	cmdMap := make(map[string]interface{})
	cmdMap["offset"] = 2

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	_, err = Handle.GetEntitiesCP(&recRead, nil, cmdMap)
	if err != nil {
		t.Errorf("error: TestCRUDGetEntitiesCPOffset: %v", err)
	}

	if len(recRead) != 6 {
		t.Errorf("error: TestCRUDGetEntitiesCPOffset: expected 6 records, got: %v", len(recRead))
	}
	if len(recRead) > 0 {
		if recRead[0].ID != 90000002 {
			t.Errorf("error: TestCRUDGetEntitiesCPOffset: expected result[0] record with key ID == 90000002, got: %v", recRead[0].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPOffset: expected 6 records, got: %v", len(recRead))
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesCPOffsetDesc
//
// Test CRUD GetSet
// call with command /$offset=2$desc
func TestCRUDGetEntitiesCPOffsetDesc(t *testing.T) {

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

	// set an offset of 2
	cmdMap := make(map[string]interface{})
	cmdMap["offset"] = 2
	cmdMap["desc"] = nil

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	_, err = Handle.GetEntitiesCP(&recRead, nil, cmdMap)
	if err != nil {
		t.Errorf("error: TestCRUDGetEntitiesCPOffsetDesc: %v", err)
	}

	if len(recRead) != 6 {
		t.Errorf("error: TestCRUDGetEntitiesCPOffsetDesc: expected 6 records, got: %v", len(recRead))
	}
	if len(recRead) > 0 {
		if recRead[0].ID != 90000005 {
			t.Errorf("error: TestCRUDGetEntitiesCPOffsetDesc: expected result[0] record with key ID == 90000005, got: %v", recRead[0].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPOffsetDesc: expected 6 records, got: %v", len(recRead))
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesCPOffsetLimit
//
// Test CRUD GetSet
// call with command /$offset=2$limit=4
func TestCRUDGetEntitiesCPOffsetLimit(t *testing.T) {

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

	// set an offset of 2, limit of 4
	cmdMap := make(map[string]interface{})
	cmdMap["offset"] = 2
	cmdMap["limit"] = 4

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	_, err = Handle.GetEntitiesCP(&recRead, nil, cmdMap)
	if err != nil {
		t.Errorf("error: TestCRUDGetEntitiesCPOffsetLimit: %v", err)
	}

	if len(recRead) != 4 {
		t.Errorf("error: TestCRUDGetEntitiesCPOffsetLimit: expected 4 records, got: %v", len(recRead))
	}
	if len(recRead) > 0 {
		if recRead[0].ID != 90000002 {
			t.Errorf("error: TestCRUDGetEntitiesCPOffsetLimit: expected result[0] record with key ID == 90000002, got: %v", recRead[0].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPOffsetLimit: expected 4 records, got: %v", len(recRead))
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesCPOffsetLimitDesc
//
// Test CRUD GetSet
// call with command /$offset=2$limit=4$desc
func TestCRUDGetEntitiesCPOffsetLimitDesc(t *testing.T) {

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

	// set an offset of 2, limit of 4, order by id descending
	cmdMap := make(map[string]interface{})
	cmdMap["offset"] = 2
	cmdMap["limit"] = 4
	cmdMap["desc"] = nil

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	_, err = Handle.GetEntitiesCP(&recRead, nil, cmdMap)
	if err != nil {
		t.Errorf("error: TestCRUDGetEntitiesCPOffsetLimitDesc: %v", err)
	}

	if len(recRead) != 4 {
		t.Errorf("error: TestCRUDGetEntitiesCPOffsetLimitDesc: expected 4 records, got: %v", len(recRead))
	}
	if len(recRead) > 0 {
		if recRead[0].ID != 90000005 {
			t.Errorf("error: TestCRUDGetEntitiesCPOffsetLimitDesc: expected result[0] record with key ID == 90000005, got: %v", recRead[0].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPOffsetLimitDesc: expected 4 records, got: %v", len(recRead))
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesCPOrderBy
//
// Test CRUD GetSet
// call with command /$orderby=name
func TestCRUDGetEntitiesCPOrderBy(t *testing.T) {

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

	// set $orderby=name
	cmdMap := make(map[string]interface{})
	cmdMap["orderby"] = "fld_two_string"

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	_, err = Handle.GetEntitiesCP(&recRead, nil, cmdMap)
	if err != nil {
		t.Errorf("error: TestCRUDGetEntitiesCPOrderBy: %v", err)
	}

	if len(recRead) != 8 {
		t.Errorf("error: TestCRUDGetEntitiesCPOrderBy: expected 8 records, got: %v", len(recRead))
	}
	if len(recRead) > 0 {
		if recRead[0].ID != 90000000 {
			t.Errorf("error: TestCRUDGetEntitiesCPOrderBy: expected result[0] record with key ID == 90000000 got: %v", recRead[0].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPOrderBy: expected 8 records, got: %v", len(recRead))
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesCPOrderByDesc
//
// Test CRUD GetSet
// call with command /$orderby=name$desc
func TestCRUDGetEntitiesCPOrderByDesc(t *testing.T) {

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

	// set $orderby=name
	cmdMap := make(map[string]interface{})
	cmdMap["orderby"] = "fld_two_string"
	cmdMap["desc"] = nil

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	_, err = Handle.GetEntitiesCP(&recRead, nil, cmdMap)
	if err != nil {
		t.Errorf("error: TestCRUDGetEntitiesCPOrderByDesc: %v", err)
	}

	if len(recRead) != 8 {
		t.Errorf("error: TestCRUDGetEntitiesCPOrderByDesc: expected 8 records, got: %v", len(recRead))
	}
	if len(recRead) > 0 {
		if recRead[0].ID != 90000007 {
			t.Errorf("error: TestCRUDGetEntitiesCPOrderByDesc: expected result[0] record with key ID == 90000007 got: %v", recRead[0].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPOrderByDesc: expected 8 records, got: %v", len(recRead))
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesCPParamOrderBy
//
// Test CRUD GetSet
// call with command /$orderby=name /parameters: id > 90000002
func TestCRUDGetEntitiesCPParamOrderBy(t *testing.T) {

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
	_, err = Handle.GetEntitiesCP(&recRead, pa, cmdMap)
	if err != nil {
		t.Errorf("error: TestCRUDGetEntitiesCPParamOrderBy: %v", err)
	}

	if len(recRead) != 5 {
		t.Errorf("error: TestCRUDGetEntitiesCPParamOrderBy: expected 5 records, got: %v", len(recRead))
	}
	if len(recRead) > 0 {
		if recRead[0].ID != 90000003 {
			t.Errorf("error: TestCRUDGetEntitiesCPParamOrderBy: expected result[0] record with key ID == 90000003 got: %v", recRead[0].ID)
		}
		if recRead[4].ID != 90000007 {
			t.Errorf("error: TestCRUDGetEntitiesCPParamOrderBy: expected result[0] record with key ID == 90000007 got: %v", recRead[4].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPParamOrderBy: expected 5 records, got: %v", len(recRead))
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesCPParamOrderByDesc
//
// Test CRUD GetSet
// call with command /$orderby=name /parameters: id > 90000002
func TestCRUDGetEntitiesCPParamOrderByDesc(t *testing.T) {

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
	_, err = Handle.GetEntitiesCP(&recRead, pa, cmdMap)
	if err != nil {
		t.Errorf("error: TestCRUDGetEntitiesCPParamOrderByDesc: %v", err)
	}

	if len(recRead) != 5 {
		t.Errorf("error: TestCRUDGetEntitiesCPParamOrderByDesc: expected 5 records, got: %v", len(recRead))
	}
	if len(recRead) > 0 {
		if recRead[0].ID != 90000007 {
			t.Errorf("error: TestCRUDGetEntitiesCPParamOrderByDesc: expected result[0] record with key ID == 90000007 got: %v", recRead[0].ID)
		}
		if recRead[4].ID != 90000003 {
			t.Errorf("error: TestCRUDGetEntitiesCPParamOrderByDesc: expected result[0] record with key ID == 90000003 got: %v", recRead[4].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPParamOrderByDesc: expected 5 records, got: %v", len(recRead))
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesCPOrderByDescLimit
//
// Test CRUD GetSet
// call with command /$orderby=name$desc$limit=3
func TestCRUDGetEntitiesCPOrderByDescLimit(t *testing.T) {

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

	// set $orderby=name
	cmdMap := make(map[string]interface{})
	cmdMap["orderby"] = "fld_two_string"
	cmdMap["desc"] = nil
	cmdMap["limit"] = 3

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	_, err = Handle.GetEntitiesCP(&recRead, nil, cmdMap)
	if err != nil {
		t.Errorf("error: TestCRUDGetEntitiesCPOrderByDescLimit: %v", err)
	}

	if len(recRead) != 3 {
		t.Errorf("error: TestCRUDGetEntitiesCPOrderByDescLimit: expected 3 records, got: %v", len(recRead))
	}
	if len(recRead) > 0 {
		if recRead[0].ID != 90000007 {
			t.Errorf("error: TestCRUDGetEntitiesCPOrderByDescLimit: expected result[0] record with key ID == 90000007 got: %v", recRead[0].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPOrderByDescLimit: expected 3 records, got: %v", len(recRead))
	}
	if len(recRead) == 3 {
		if recRead[2].ID != 90000005 {
			t.Errorf("error: TestCRUDGetEntitiesCPOrderByDescLimit: expected result[2] record with key ID == 90000005 got: %v", recRead[2].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPOrderByDescLimit: expected 3 records, got: %v", len(recRead))
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesCPOrderByDescLimitOffset
//
// Test CRUD GetSet
// call with command /$orderby=name$desc$limit=3$offset=2
func TestCRUDGetEntitiesCPOrderByDescLimitOffset(t *testing.T) {

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

	// set $orderby=name$descending$limit=3$offset=2
	cmdMap := make(map[string]interface{})
	cmdMap["orderby"] = "fld_two_string"
	cmdMap["desc"] = nil
	cmdMap["limit"] = 3
	cmdMap["offset"] = 2

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	_, err = Handle.GetEntitiesCP(&recRead, nil, cmdMap)
	if err != nil {
		t.Errorf("error: TestCRUDGetEntitiesCPOrderByDescLimitOffset: %v", err)
	}

	if len(recRead) != 3 {
		t.Errorf("error: TestCRUDGetEntitiesCPOrderByDescLimitOffset: expected 3 records, got: %v", len(recRead))
	}
	if len(recRead) > 0 {
		if recRead[0].ID != 90000005 {
			t.Errorf("error: TestCRUDGetEntitiesCPOrderByDescLimitOffset: expected result[0] record with key ID == 90000005 got: %v", recRead[0].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPOrderByDescLimitOffset: expected 3 records, got: %v", len(recRead))
	}
	if len(recRead) == 3 {
		if recRead[2].ID != 90000003 {
			t.Errorf("error: TestCRUDGetEntitiesCPOrderByDescLimitOffset: expected result[2] record with key ID == 90000003 got: %v", recRead[2].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPOrderByDescLimitOffset: expected 3 records, got: %v", len(recRead))
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntitiesCPOrderByAscLimitOffset
//
// Test CRUD GetSet
// call with command /$orderby=name$asc$limit=3$offset=2
func TestCRUDGetEntitiesCPOrderByAscLimitOffset(t *testing.T) {

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

	// set $orderby=name$descending$limit=3$offset=2
	cmdMap := make(map[string]interface{})
	cmdMap["orderby"] = "fld_two_string"
	cmdMap["asc"] = nil
	cmdMap["limit"] = 3
	cmdMap["offset"] = 2

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	_, err = Handle.GetEntitiesCP(&recRead, nil, cmdMap)
	if err != nil {
		t.Errorf("error: TestCRUDGetEntitiesCPOrderByAscLimitOffset: %v", err)
	}

	if len(recRead) != 3 {
		t.Errorf("error: TestCRUDGetEntitiesCPOrderByAscLimitOffset: expected 3 records, got: %v", len(recRead))
	}
	if len(recRead) > 0 {
		if recRead[0].ID != 90000002 {
			t.Errorf("error: TestCRUDGetEntitiesCPOrderByAscLimitOffset: expected result[0] record with key ID == 90000002 got: %v", recRead[0].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPOrderByAscLimitOffset: expected 3 records, got: %v", len(recRead))
	}
	if len(recRead) == 3 {
		if recRead[2].ID != 90000004 {
			t.Errorf("error: TestCRUDGetEntitiesCPOrderByAscLimitOffset: expected result[2] record with key ID == 90000004 got: %v", recRead[2].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPOrderByAscLimitOffset: expected 3 records, got: %v", len(recRead))
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// TestCRUDGetEntities5SelectOrderBy
//
// Test CRUD GetSet
// call with command /$orderby=name
func TestCRUDGetEntitiesCPTestOffsetOrderByAscLimitOffset(t *testing.T) {

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

	// set $orderby=name$descending$limit=3$offset=2
	cmdMap := make(map[string]interface{})
	cmdMap["orderby"] = "fld_two_string"
	cmdMap["asc"] = nil
	cmdMap["limit"] = 3
	cmdMap["offset"] = 2

	// create a slice to read into
	recRead := []GetCmdTest{}

	// call with no parameters and no commands
	_, err = Handle.GetEntitiesCP(&recRead, nil, cmdMap)
	if err != nil {
		t.Errorf("error: TestCRUDGetEntitiesCPTestOrderByAscLimitOffset: %v", err)
	}

	if len(recRead) != 3 {
		t.Errorf("error: TestCRUDGetEntitiesCPTestOrderByAscLimitOffset: expected 3 records, got: %v", len(recRead))
	}
	if len(recRead) > 0 {
		if recRead[0].ID != 90000002 {
			t.Errorf("error: TestCRUDGetEntitiesCPTestOrderByAscLimitOffset: expected result[0] record with key ID == 90000002 got: %v", recRead[0].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPTestOrderByAscLimitOffset: expected 3 records, got: %v", len(recRead))
	}
	if len(recRead) == 3 {
		if recRead[2].ID != 90000004 {
			t.Errorf("error: TestCRUDGetEntitiesCPTestOrderByAscLimitOffset: expected result[2] record with key ID == 90000004 got: %v", recRead[2].ID)
		}
	} else {
		t.Errorf("error: TestCRUDGetEntitiesCPTestOrderByAscLimitOffset: expected 3 records, got: %v", len(recRead))
	}

	// drop table getcmdtest
	err = Handle.DropTables(GetCmdTest{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

// func createGetCmdTestRecs(t *testing.T) {
// 	// create new records via the CRUD Create call
// 	rec := GetCmdTest{
// 		FldOneInt:     1,
// 		FldTwoString:  "AAAAAA",
// 		FldThreeFloat: 10.2922,
// 		FldFourBool:   true,
// 	}
// 	f5 := "Optional String Value"
// 	rec.FldFiveString = &f5
// 	f6 := 99.312345
// 	rec.FldSixFloat = &f6
// 	f7 := false
// 	rec.FldSevenBool = &f7

// 	// create a record
// 	err := Handle.Create(&rec)
// 	if err != nil {
// 		t.Errorf(err.Error())
// 	}

// 	// create record 2
// 	rec = GetCmdTest{
// 		FldOneInt:     2,
// 		FldTwoString:  "BBBBBB",
// 		FldThreeFloat: 20.2762,
// 		FldFourBool:   true,
// 	}
// 	f5 = "Optional String Value Two"
// 	rec.FldFiveString = &f5
// 	f6 = 200.12121
// 	rec.FldSixFloat = &f6
// 	f7 = false
// 	rec.FldSevenBool = &f7

// 	err = Handle.Create(&rec)
// 	if err != nil {
// 		t.Errorf(err.Error())
// 	}

// 	// create record 3
// 	rec = GetCmdTest{
// 		FldOneInt:     3,
// 		FldTwoString:  "CCCCCC",
// 		FldThreeFloat: 30.3385,
// 		FldFourBool:   true,
// 	}
// 	f5 = "Optional String Value Three"
// 	rec.FldFiveString = &f5
// 	f6 = 300.31313
// 	rec.FldSixFloat = &f6
// 	f7 = false
// 	rec.FldSevenBool = &f7

// 	err = Handle.Create(&rec)
// 	if err != nil {
// 		t.Errorf(err.Error())
// 	}

// 	// create record 4
// 	rec = GetCmdTest{
// 		FldOneInt:     4,
// 		FldTwoString:  "DDDDDD",
// 		FldThreeFloat: 40.75757,
// 		FldFourBool:   true,
// 	}
// 	f5 = "Optional String Value Four"
// 	rec.FldFiveString = &f5
// 	f6 = 400.1414
// 	rec.FldSixFloat = &f6
// 	f7 = false
// 	rec.FldSevenBool = &f7

// 	err = Handle.Create(&rec)
// 	if err != nil {
// 		t.Errorf(err.Error())
// 	}

// 	// create record 5
// 	rec = GetCmdTest{
// 		FldOneInt:     5,
// 		FldTwoString:  "EEEEEE",
// 		FldThreeFloat: 50.58585,
// 		FldFourBool:   true,
// 	}
// 	f5 = "Optional String Value Five"
// 	rec.FldFiveString = &f5
// 	f6 = 400.1414
// 	rec.FldSixFloat = &f6
// 	f7 = false
// 	rec.FldSevenBool = &f7

// 	err = Handle.Create(&rec)
// 	if err != nil {
// 		t.Errorf(err.Error())
// 	}

// 	// create record 6
// 	rec = GetCmdTest{
// 		FldOneInt:     6,
// 		FldTwoString:  "FFFFFF",
// 		FldThreeFloat: 60.6767,
// 		FldFourBool:   true,
// 	}
// 	f5 = "Optional String Value Six"
// 	rec.FldFiveString = &f5
// 	f6 = 600.1616
// 	rec.FldSixFloat = &f6
// 	f7 = false
// 	rec.FldSevenBool = &f7

// 	err = Handle.Create(&rec)
// 	if err != nil {
// 		t.Errorf(err.Error())
// 	}

// 	// create record 7
// 	rec = GetCmdTest{
// 		FldOneInt:     7,
// 		FldTwoString:  "GGGGGG",
// 		FldThreeFloat: 70.73737,
// 		FldFourBool:   true,
// 	}
// 	f5 = "Optional String Value Seven"
// 	rec.FldFiveString = &f5
// 	f6 = 700.12224
// 	rec.FldSixFloat = &f6
// 	f7 = false
// 	rec.FldSevenBool = &f7

// 	err = Handle.Create(&rec)
// 	if err != nil {
// 		t.Errorf(err.Error())
// 	}

// 	// create record 8
// 	rec = GetCmdTest{
// 		FldOneInt:     8,
// 		FldTwoString:  "HHHHHH",
// 		FldThreeFloat: 80.85858,
// 		FldFourBool:   true,
// 	}
// 	f5 = "Optional String Value Eight"
// 	rec.FldFiveString = &f5
// 	f6 = 800.1818
// 	rec.FldSixFloat = &f6
// 	f7 = false
// 	rec.FldSevenBool = &f7

// 	err = Handle.Create(&rec)
// 	if err != nil {
// 		t.Errorf(err.Error())
// 	}
// }

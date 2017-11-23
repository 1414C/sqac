package sqac_test

import "testing"
import "time"
import "fmt"

import "github.com/1414C/sqac/common"

func TestNullString(t *testing.T) {

	type NString struct {
		NSKey                   int       `db:"ns_key" rgen:"primary_key:inc"`
		CreateDate              time.Time `db:"create_date" rgen:"nullable:false;default:now();"`
		StringDflt              string    `db:"string_dflt" rgen:"nullable:false;default:dflt_value"`
		StringDfltWithValue     string    `db:"string_dflt_with_value" rgen:"nullable:false;default:dflt_value2"`
		StringWithValue         string    `db:"string_with_value" rgen:"nullable:false"`
		NullStringDflt          *string   `db:"null_string_dflt" rgen:"nullable:true;default:dflt_value_for_nullable"`
		NullStringDfltWithValue *string   `db:"null_string_dflt_with_value" rgen:"nullable:true;default:dflt_value_for_nullable2"`
		NullStringWithValue     *string   `db:"null_string_with_value" rgen:"nullable:true"`
		NullString              *string   `db:"null_string" rgen:"nullable:true"`
	}

	err := Handle.CreateTables(NString{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(NString{})

	// expect that table depot exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s was not created", tn)
	}

	s1 := "n_crt_value1"
	s2 := "n_crt_value2"

	// create a new record via the CRUD Create call
	var nstring = NString{
		StringDflt:              "",
		StringDfltWithValue:     "crt_str_value1",
		StringWithValue:         "crt_str_value2",
		NullStringDflt:          nil,
		NullStringDfltWithValue: &s1,
		NullStringWithValue:     &s2,
		NullString:              nil,
	}

	if Handle.IsLog() {
		fmt.Printf("INSERTING: %v\n", nstring)
	}

	err = Handle.Create(&nstring)
	if err != nil {
		t.Errorf(err.Error())
	}
	if Handle.IsLog() {
		fmt.Printf("TEST GOT: %v\n", nstring)
	}
	fmt.Printf("TEST GOT: %v\n", nstring)

	Handle.Log(true)
	if nstring.StringDflt != "dflt_value" {
		t.Errorf("nstring expected %s for field 'StringDflt', got: %v", "dflt_value", nstring.StringDflt)
	} else {
		if Handle.IsLog() {
			fmt.Printf("nstring.StringDflt: %v\n", nstring.StringDflt)
		}
	}
	if nstring.StringDfltWithValue != "crt_str_value1" {
		t.Errorf("nstring expected %s for field 'StringDfltWithValue', got: %v", "crt_str_value1", nstring.StringDfltWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("nstring.StringDfltWithValue: %v\n", nstring.StringDfltWithValue)
		}
	}
	if nstring.StringWithValue != "crt_str_value2" {
		t.Errorf("nstring expected %s for field 'StringWithValue', got: %v", "crt_str_value2", nstring.StringDfltWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("nstring.StringWithValue: %v\n", nstring.StringWithValue)
		}
	}
	if *nstring.NullStringDflt != "dflt_value_for_nullable" {
		t.Errorf("nstring expected %s for field '*NullStringDflt', got: %v", "dflt_value_for_nullable", *nstring.NullStringDflt)
	} else {
		if Handle.IsLog() {
			fmt.Printf("*nstring.NullStringDflt: %v\n", *nstring.NullStringDflt)
		}
	}
	if *nstring.NullStringDfltWithValue != "n_crt_value1" {
		t.Errorf("nstring expected %s for field '*NullStringDfltWithValue', got: %v", "n_crt_value1", *nstring.NullStringDfltWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("*nstring.NullStringDfltWithValue: %v\n", *nstring.NullStringDfltWithValue)
		}
	}
	if *nstring.NullStringWithValue != "n_crt_value2" {
		t.Errorf("nstring expected %s for field '*NullStringWithValue', got: %v", "n_crt_value2", *nstring.NullStringWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("*nstring.NullStringWithValue: %v\n", *nstring.NullStringWithValue)
		}
	}
	if nstring.NullString != nil {
		t.Errorf("nstring expected <nil> for field 'NullString', got: %#v", *nstring.NullString)
	} else {
		if Handle.IsLog() {
			fmt.Printf("nstring.NullString: %v\n", nstring.NullString)
		}
	}
	Handle.Log(false)
}

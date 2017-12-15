package sqac_test

import "testing"
import "time"
import "fmt"

import "github.com/1414C/sqac/common"

func TestNullString(t *testing.T) {

	type NString struct {
		NSKey                   int       `db:"ns_key" sqac:"primary_key:inc"`
		CreateDate              time.Time `db:"create_date" sqac:"nullable:false;default:now();"`
		StringDflt              string    `db:"string_dflt" sqac:"nullable:false;default:dflt_value"`
		StringDfltWithValue     string    `db:"string_dflt_with_value" sqac:"nullable:false;default:dflt_value2"`
		StringWithValue         string    `db:"string_with_value" sqac:"nullable:false"`
		NullStringDflt          *string   `db:"null_string_dflt" sqac:"nullable:true;default:dflt_value_for_nullable"`
		NullStringDfltWithValue *string   `db:"null_string_dflt_with_value" sqac:"nullable:true;default:dflt_value_for_nullable2"`
		NullStringWithValue     *string   `db:"null_string_with_value" sqac:"nullable:true"`
		NullString              *string   `db:"null_string" sqac:"nullable:true"`
	}

	// create table if requied
	err := Handle.CreateTables(NString{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(NString{})

	// expect that table nstring exists
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
}

func TestNullInt(t *testing.T) {

	type NInt struct {
		NIKey                int       `db:"ni_key" sqac:"primary_key:inc"`
		CreateDate           time.Time `db:"create_date" sqac:"nullable:false;default:now();"`
		IntDflt              int       `db:"int_dflt" sqac:"nullable:false;default:1111"`
		IntDfltWithValue     int       `db:"int_dflt_with_value" sqac:"nullable:false;default:2222"`
		IntWithValue         int       `db:"int_with_value" sqac:"nullable:false"`
		NullIntDflt          *int      `db:"null_int_dflt" sqac:"nullable:true;default:5555"`
		NullIntDfltWithValue *int      `db:"null_int_dflt_with_value" sqac:"nullable:true;default:6666"`
		NullIntWithValue     *int      `db:"null_int_with_value" sqac:"nullable:true"`
		NullInt              *int      `db:"null_int" sqac:"nullable:true"`
	}

	// create table if requied
	err := Handle.CreateTables(NInt{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(NInt{})

	// expect that table nint exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s was not created", tn)
	}

	i1 := 100
	i2 := 200

	// create a new record via the CRUD Create call
	var nint = NInt{
		IntDflt:              0,
		IntDfltWithValue:     10,
		IntWithValue:         20,
		NullIntDflt:          nil,
		NullIntDfltWithValue: &i1,
		NullIntWithValue:     &i2,
		NullInt:              nil,
	}

	if Handle.IsLog() {
		fmt.Printf("INSERTING: %v\n", nint)
	}

	err = Handle.Create(&nint)
	if err != nil {
		t.Errorf(err.Error())
	}
	if Handle.IsLog() {
		fmt.Printf("TEST GOT: %v\n", nint)
	}

	if nint.IntDflt != 1111 {
		t.Errorf("nint expected %d for field 'IntDflt', got: %v", 1111, nint.IntDflt)
	} else {
		if Handle.IsLog() {
			fmt.Printf("nint.IntDflt: %v\n", nint.IntDflt)
		}
	}
	if nint.IntDfltWithValue != 10 {
		t.Errorf("nint expected %d for field 'IntDfltWithValue', got: %v", 10, nint.IntDfltWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("nint.IntDfltWithValue: %v\n", nint.IntDfltWithValue)
		}
	}
	if nint.IntWithValue != 20 {
		t.Errorf("nint expected %d for field 'IntWithValue', got: %v", 20, nint.IntWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("nint.IntWithValue : %v\n", nint.IntWithValue)
		}
	}
	if *nint.NullIntDflt != 5555 {
		t.Errorf("nint expected %d for field '*NullIntDflt', got: %v", 5555, *nint.NullIntDflt)
	} else {
		if Handle.IsLog() {
			fmt.Printf("*nint.NullIntDflt: %v\n", *nint.NullIntDflt)
		}
	}
	if *nint.NullIntDfltWithValue != 100 {
		t.Errorf("nint expected %d for field '*NullIntDfltWithValue', got: %v", 100, *nint.NullIntDfltWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("*nint.NullIntDfltWithValue: %v\n", *nint.NullIntDfltWithValue)
		}
	}
	if *nint.NullIntWithValue != 200 {
		t.Errorf("nint expected %d for field '*NullIntWithValue', got: %v", 200, *nint.NullIntWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("*nint.NullIntWithValue: %v\n", *nint.NullIntWithValue)
		}
	}
	if nint.NullInt != nil {
		t.Errorf("nint expected <nil> for field 'NullInt', got: %#v", *nint.NullInt)
	} else {
		if Handle.IsLog() {
			fmt.Printf("*nint.NullInt: %v\n", nint.NullInt)
		}
	}
}

func TestNullUint(t *testing.T) {

	type NUint struct {
		NIKey                 uint      `db:"ni_key" sqac:"primary_key:inc"`
		CreateDate            time.Time `db:"create_date" sqac:"nullable:false;default:now();"`
		UintDflt              uint      `db:"uint_dflt" sqac:"nullable:false;default:1111"`
		UintDfltWithValue     uint      `db:"uint_dflt_with_value" sqac:"nullable:false;default:2222"`
		UintWithValue         uint      `db:"uint_with_value" sqac:"nullable:false"`
		NullUintDflt          *uint     `db:"null_uint_dflt" sqac:"nullable:true;default:5555"`
		NullUintDfltWithValue *uint     `db:"null_uint_dflt_with_value" sqac:"nullable:true;default:6666"`
		NullUintWithValue     *uint     `db:"null_uint_with_value" sqac:"nullable:true"`
		NullUint              *uint     `db:"null_uint" sqac:"nullable:true"`
	}

	// create table if requied
	err := Handle.CreateTables(NUint{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(NUint{})

	// expect that table nuint exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s was not created", tn)
	}

	var u1 uint
	var u2 uint
	u1 = 100
	u2 = 200

	// create a new record via the CRUD Create call
	var nuint = NUint{
		UintDflt:              0,
		UintDfltWithValue:     10,
		UintWithValue:         20,
		NullUintDflt:          nil,
		NullUintDfltWithValue: &u1,
		NullUintWithValue:     &u2,
		NullUint:              nil,
	}

	if Handle.IsLog() {
		fmt.Printf("INSERTING: %v\n", nuint)
	}

	err = Handle.Create(&nuint)
	if err != nil {
		t.Errorf(err.Error())
	}
	if Handle.IsLog() {
		fmt.Printf("TEST GOT: %v\n", nuint)
	}

	if nuint.UintDflt != 1111 {
		t.Errorf("nuint expected %d for field 'UintDflt', got: %v", 1111, nuint.UintDflt)
	} else {
		if Handle.IsLog() {
			fmt.Printf("nint.IntDflt: %v\n", nuint.UintDflt)
		}
	}
	if nuint.UintDfltWithValue != 10 {
		t.Errorf("nuint expected %d for field 'UintDfltWithValue', got: %v", 10, nuint.UintDfltWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("nuint.UintDfltWithValue: %v\n", nuint.UintDfltWithValue)
		}
	}
	if nuint.UintWithValue != 20 {
		t.Errorf("nuint expected %d for field 'UintWithValue', got: %v", 20, nuint.UintWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("nuint.UintWithValue : %v\n", nuint.UintWithValue)
		}
	}
	if *nuint.NullUintDflt != 5555 {
		t.Errorf("nuint expected %d for field '*NullUintDflt', got: %v", 5555, *nuint.NullUintDflt)
	} else {
		if Handle.IsLog() {
			fmt.Printf("*nuint.NullUintDflt: %v\n", *nuint.NullUintDflt)
		}
	}
	if *nuint.NullUintDfltWithValue != 100 {
		t.Errorf("nuint expected %d for field '*NullUintDfltWithValue', got: %v", 100, *nuint.NullUintDfltWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("*nuint.NullUintDfltWithValue: %v\n", *nuint.NullUintDfltWithValue)
		}
	}
	if *nuint.NullUintWithValue != 200 {
		t.Errorf("nuint expected %d for field '*NullUintWithValue', got: %v", 200, *nuint.NullUintWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("*nuint.NullUintWithValue: %v\n", *nuint.NullUintWithValue)
		}
	}
	if nuint.NullUint != nil {
		t.Errorf("nuint expected <nil> for field 'NullUint', got: %#v", *nuint.NullUint)
	} else {
		if Handle.IsLog() {
			fmt.Printf("*nuint.NullUint: %v\n", nuint.NullUint)
		}
	}
}

func TestNullFloat(t *testing.T) {

	type NFloat struct {
		NIKey                  int       `db:"ni_key" sqac:"primary_key:inc"`
		CreateDate             time.Time `db:"create_date" sqac:"nullable:false;default:now();"`
		FloatDflt              float64   `db:"float_dflt" sqac:"nullable:false;default:1111.222"`
		FloatDfltWithValue     float64   `db:"float_dflt_with_value" sqac:"nullable:false;default:3333.444"`
		FloatWithValue         float64   `db:"float_with_value" sqac:"nullable:false"`
		NullFloatDflt          *float64  `db:"null_float_dflt" sqac:"nullable:true;default:6666.777"`
		NullFloatDfltWithValue *float64  `db:"null_float_dflt_with_value" sqac:"nullable:true;default:8888.999"`
		NullFloatWithValue     *float64  `db:"null_float_with_value" sqac:"nullable:true"`
		NullFloat              *float64  `db:"null_float" sqac:"nullable:true"`
	}

	// create table if requied
	err := Handle.CreateTables(NFloat{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(NFloat{})

	// expect that table nfloat exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s was not created", tn)
	}

	f1 := 100.4242
	f2 := 200.5656

	// create a new record via the CRUD Create call
	var nfloat = NFloat{
		FloatDflt:              0,
		FloatDfltWithValue:     10.701,
		FloatWithValue:         20.702,
		NullFloatDflt:          nil,
		NullFloatDfltWithValue: &f1,
		NullFloatWithValue:     &f2,
		NullFloat:              nil,
	}

	if Handle.IsLog() {
		fmt.Printf("INSERTING: %v\n", nfloat)
	}

	err = Handle.Create(&nfloat)
	if err != nil {
		t.Errorf(err.Error())
	}

	if Handle.IsLog() {
		fmt.Printf("TEST GOT: %v\n", nfloat)
	}

	if nfloat.FloatDflt != 1111.222 {
		t.Errorf("nfloat expected %f for field 'FloatDflt', got: %v", 1111.222, nfloat.FloatDflt)
	} else {
		if Handle.IsLog() {
			fmt.Printf("nfloat.FloattDflt: %f\n", nfloat.FloatDflt)
		}
	}
	if nfloat.FloatDfltWithValue != 10.701 {
		t.Errorf("nfloat expected %f for field 'FloatDfltWithValue', got: %v", 10.701, nfloat.FloatDfltWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("nfloat.FloatDfltWithValue: %f\n", nfloat.FloatDfltWithValue)
		}
	}
	if nfloat.FloatWithValue != 20.702 {
		t.Errorf("nfloat expected %f for field 'FloatWithValue', got: %v", 20.702, nfloat.FloatWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("nfloat.FloatWithValue : %f\n", nfloat.FloatWithValue)
		}
	}
	if *nfloat.NullFloatDflt != 6666.777 {
		t.Errorf("nfloat expected %f for field '*NullFloatDflt', got: %v", 6666.777, *nfloat.NullFloatDflt)
	} else {
		if Handle.IsLog() {
			fmt.Printf("*nfloat.NullFloatDflt: %f\n", *nfloat.NullFloatDflt)
		}
	}
	if *nfloat.NullFloatDfltWithValue != 100.4242 {
		t.Errorf("nfloat expected %f for field '*NullFloatDfltWithValue', got: %v", 100.4242, *nfloat.NullFloatDfltWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("*nfloat.NullFloatDfltWithValue: %f\n", *nfloat.NullFloatDfltWithValue)
		}
	}
	if *nfloat.NullFloatWithValue != 200.5656 {
		t.Errorf("nfloat expected %f for field '*NullFloatWithValue', got: %v", 200.5656, *nfloat.NullFloatWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("*nfloat.NullFloatWithValue: %f\n", *nfloat.NullFloatWithValue)
		}
	}
	if nfloat.NullFloat != nil {
		t.Errorf("nfloat expected <nil> for field 'NullFloat', got: %#v", *nfloat.NullFloat)
	} else {
		if Handle.IsLog() {
			fmt.Printf("*nfloat.NullFloat: %v\n", nfloat.NullFloat)
		}
	}
}

func TestNullBool(t *testing.T) {

	type NBool struct {
		NIKey                 int       `db:"ni_key" sqac:"primary_key:inc"`
		CreateDate            time.Time `db:"create_date" sqac:"nullable:false;default:now();"`
		BoolDfltWithValue     bool      `db:"bool_dflt_with_value" sqac:"nullable:false;default:true"`
		BoolWithValue         bool      `db:"bool_with_value" sqac:"nullable:false"`
		NullBoolDflt          *bool     `db:"null_bool_dflt" sqac:"nullable:true;default:true"`
		NullBoolDfltWithValue *bool     `db:"null_bool_dflt_with_value" sqac:"nullable:true;default:true"`
		NullBoolWithValue     *bool     `db:"null_bool_with_value" sqac:"nullable:true"`
		NullBool              *bool     `db:"null_bool" sqac:"nullable:true"`
	}

	// create table if requied
	err := Handle.CreateTables(NBool{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// determine the table name as per the table creation logic
	tn := common.GetTableName(NBool{})

	// expect that table nbool exists
	if !Handle.ExistsTable(tn) {
		t.Errorf("table %s was not created", tn)
	}

	b1 := false
	b2 := false

	// create a new record via the CRUD Create call
	var nbool = NBool{
		BoolDfltWithValue:     false,
		BoolWithValue:         false,
		NullBoolDflt:          nil,
		NullBoolDfltWithValue: &b1,
		NullBoolWithValue:     &b2,
		NullBool:              nil,
	}

	if Handle.IsLog() {
		fmt.Printf("INSERTING: %v\n", nbool)
	}

	err = Handle.Create(&nbool)
	if err != nil {
		t.Errorf(err.Error())
	}

	if Handle.IsLog() {
		fmt.Printf("TEST GOT: %v\n", nbool)
	}

	if nbool.BoolDfltWithValue != false {
		t.Errorf("nbool expected %t for field 'BoolDfltWithValue', got: %v", false, nbool.BoolDfltWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("nbool.BoolDfltWithValue: %v\n", nbool.BoolDfltWithValue)
		}
	}
	if nbool.BoolWithValue != false {
		t.Errorf("nbool expected %t for field 'BoolWithValue', got: %v", false, nbool.BoolWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("nbool.BoolWithValue : %v\n", nbool.BoolWithValue)
		}
	}
	if *nbool.NullBoolDflt != true {
		t.Errorf("nbool expected %t for field '*NullBoolDflt', got: %v", true, *nbool.NullBoolDflt)
	} else {
		if Handle.IsLog() {
			fmt.Printf("*nbool.NullBoolflt: %v\n", *nbool.NullBoolDflt)
		}
	}
	if *nbool.NullBoolDfltWithValue != false {
		t.Errorf("nbool expected %t for field '*NullBoolDfltWithValue', got: %v", false, *nbool.NullBoolDfltWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("*nbool.NullBoolDfltWithValue: %v\n", *nbool.NullBoolDfltWithValue)
		}
	}
	if *nbool.NullBoolWithValue != false {
		t.Errorf("nbool expected %t for field '*NullBoolWithValue', got: %v", false, *nbool.NullBoolWithValue)
	} else {
		if Handle.IsLog() {
			fmt.Printf("*nbool.NullBoolWithValue: %v\n", *nbool.NullBoolWithValue)
		}
	}
	if nbool.NullBool != nil {
		t.Errorf("nbool expected <nil> for field 'NullBool', got: %#v", *nbool.NullBool)
	} else {
		if Handle.IsLog() {
			fmt.Printf("*nbool.NullBool: %v\n", nbool.NullBool)
		}
	}
}

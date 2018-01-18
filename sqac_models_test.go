package sqac_test

import "time"

// Material has a single primary key
type Material struct {
	MaterialNum int `sqac:"primary_key"`
	CreatedAt   time.Time
	Description string
	Weight      float64
	WeightUOM   string
	Color       string
}

// Equipment has a compound primary key and
// creates a field with every presently
// supported go-type.  Inclusion of the
// Triplet structure demonstrates the use
// of flat sub-structures in entity
// definitions.
type Triplet struct {
	TripOne   string `db:"trip_one" sqac:"nullable:false"`
	TripTwo   int64  `db:"trip_two" sqac:"nullable:false;default:0"`
	TripThree string `db:"trip_three" sqac:"nullable:false;default:test_string"`
}

type Equipment struct {
	EquipmentNum   int64     `db:"equipment_num" sqac:"primary_key:inc;start:55550000"`
	ValidFrom      time.Time `db:"valid_from" sqac:"primary_key;nullable:false;default:now()"`
	ValidTo        time.Time `db:"valid_to" sqac:"primary_key;nullable:false;default:eot"`
	CreatedAt      time.Time `db:"created_at" sqac:"nullable:false;default:now()"`
	InspectionAt   time.Time `db:"inspection_at" sqac:"nullable:true"`
	MaterialNum    int       `db:"material_num" sqac:"index:idx_material_num_serial_num"`
	Description    string    `db:"description" sqac:"sqac:nullable:false"`
	SerialNum      string    `db:"serial_num" sqac:"index:idx_material_num_serial_num"`
	IntExample     int       `db:"int_example" sqac:"nullable:false;default:0"`
	Int64Example   int64     `db:"int64_example" sqac:"nullable:false;default:0"`
	Int32Example   int32     `db:"int32_example" sqac:"nullable:false;default:0"`
	Int16Example   int16     `db:"int16_example" sqac:"nullable:false;default:0"`
	Int8Example    int8      `db:"int8_example" sqac:"nullable:false;default:0"`
	UIntExample    uint      `db:"u_int_example" sqac:"nullable:false;default:0"`
	UInt64Example  uint64    `db:"u_int64_example" sqac:"nullable:false;default:0"`
	UInt32Example  uint32    `db:"u_int32_example" sqac:"nullable:false;default:0"`
	UInt16Example  uint16    `db:"u_int16_example" sqac:"nullable:false;default:0"`
	UInt8Example   uint8     `db:"u_int8_example" sqac:"nullable:false;default:0"`
	Float32Example float32   `db:"float32_example" sqac:"nullable:false;default:0.0"`
	Float64Example float64   `db:"float64_example" sqac:"nullable:false;default:0.0"`
	BoolExample    bool      `db:"bool_example" sqac:"nullable:false;default:false"`
	RuneExample    rune      `db:"rune_example" sqac:"nullable:true"`
	ByteExample    byte      `db:"byte_example" sqac:"nullable:true"`
	DoNotCreate    string    `db:"do_not_create" sqac:"-"`
	Triplet
}

type GetCmdTest struct {
	ID                  uint64    `db:"id" json:"id" sqac:"primary_key:inc;start:90000000"`
	FldOneInt           int       `db:"fld_one_int" json:"fld_one_int" sqac:"nullable:false;default:0"`
	TimeNow             time.Time `db:"time_now" json:"time_now" sqac:"nullable:false;default:now();index:nonUnique"`
	FldTwoString        string    `db:"fld_two_string" json:"fld_two_string" sqac:"nullable:false;default:YYC"`
	FldThreeFloat       float64   `db:"fld_three_float" json:"fld_three_float" sqac:"nullable:false;default:0.0"`
	FldFourBool         bool      `db:"fld_four_bool" json:"fld_four_bool"  sqac:"nullable:false;default:false"`
	NonPersistentColumn string    `db:"non_persistent_column" sqac:"-"`
	FldFiveString       *string   `db:"fld_five_string" json:"fld_five_string" sqac:"nullable:true"`
	FldSixFloat         *float64  `db:"fld_six_float" json:"fld_six_float" sqac:"nullable:true"`
	FldSevenBool        *bool     `db:"fld_seven_bool" json:"fld_seven_bool" sqac:"nullable:true"`
}

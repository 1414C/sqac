package sqac_test

import "time"

// Material has a single primary key
type Material struct {
	MaterialNum int `rgen:"primary_key"`
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
	TripOne   string `db:"trip_one" rgen:"nullable:false"`
	TripTwo   int64  `db:"trip_two" rgen:"nullable:false;default:0"`
	TripThree string `db:"trip_three" rgen:"nullable:false"`
}

type Equipment struct {
	EquipmentNum   int64     `db:"equipment_num" rgen:"primary_key:inc;start:55550000"`
	ValidFrom      time.Time `db:"valid_from" rgen:"primary_key;nullable:false;default:now()"`
	ValidTo        time.Time `db:"valid_to" rgen:"primary_key;nullable:false;default:eot"`
	CreatedAt      time.Time `db:"created_at" rgen:"nullable:false;default:now()"`
	InspectionAt   time.Time `db:"inspection_at" rgen:"nullable:true"`
	MaterialNum    int       `db:"material_num" rgen:"index:idx_material_num_serial_num"`
	Description    string    `db:"description" rgen:"rgen:nullable:false"`
	SerialNum      string    `db:"serial_num" rgen:"index:idx_material_num_serial_num"`
	IntExample     int       `db:"int_example" rgen:"nullable:false;default:0"`
	Int64Example   int64     `db:"int64_example" rgen:"nullable:false;default:0"`
	Int32Example   int32     `db:"int32_example" rgen:"nullable:false;default:0"`
	Int16Example   int16     `db:"int16_example" rgen:"nullable:false;default:0"`
	Int8Example    int8      `db:"int8_example" rgen:"nullable:false;default:0"`
	UIntExample    uint      `db:"u_int_example" rgen:"nullable:false;default:0"`
	UInt64Example  uint64    `db:"u_int64_example" rgen:"nullable:false;default:0"`
	UInt32Example  uint32    `db:"u_int32_example" rgen:"nullable:false;default:0"`
	UInt16Example  uint16    `db:"u_int16_example" rgen:"nullable:false;default:0"`
	UInt8Example   uint8     `db:"u_int8_example" rgen:"nullable:false;default:0"`
	Float32Example float32   `db:"float32_example" rgen:"nullable:false;default:0.0"`
	Float64Example float64   `db:"float64_example" rgen:"nullable:false;default:0.0"`
	BoolExample    bool      `db:"bool_example" rgen:"nullable:false;default:false"`
	RuneExample    rune      `db:"rune_example" rgen:"nullable:true"`
	ByteExample    byte      `db:"byte_example" rgen:"nullable:true"`
	Triplet
}

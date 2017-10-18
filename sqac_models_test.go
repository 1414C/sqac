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
type Equipment struct {
	BigKey1        int64     `rgen:"primary_key;"`
	ValidFrom      time.Time `rgen:"primary_key"`
	ValidTo        time.Time `rgen:"primary_key"`
	CreatedAt      time.Time
	InspectionAt   time.Time
	MaterialNum    int // `rgen:"index:idx_material_num_serial"`
	Description    string
	Serial         string // `rgen:"index:idx_material_num_serial"`
	IntExample     int
	Int64Example   int64
	Int32Example   int32
	Int16Example   int16
	Int8Example    int8
	UIntExample    uint
	UInt64Example  uint64
	UInt32Example  uint32
	UInt16Example  uint16
	UInt8Example   uint8
	Float32Example float32
	Float64Example float64
	BoolExample    bool
	RuneExample    rune
	ByteExample    byte
	Triplet
}

// Triplet is used as an embedded structure
type Triplet struct {
	Fiddle string
	Piddle int64
	Tiddle string
}

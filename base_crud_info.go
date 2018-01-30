package sqac

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/1414C/sqac/common"
)

// CrudInfo contains information used to perform CRUD
// activities.  Pre-call and post-call organization
// and formatting.
// v = Value (underlying struct of interface ptr ent)
type CrudInfo struct {
	ent        interface{}
	log        bool
	mode       string // "C" || "U"  || "D" == create or update or delete
	stype      reflect.Type
	flDef      []common.FieldDef
	tn         string
	fList      string
	vList      string
	fldMap     map[string]interface{} // string
	keyMap     map[string]interface{}
	incKeyName string
	entValue   reflect.Value
	resultMap  map[string]interface{}
}

// BuildComponents is used by each flavor to assemble the
// struct (entity) data for CRUD operations.  There is
// some redundancy in the structure for now, as it has
// recently been migrated into BaseFlavor.
func (bf *BaseFlavor) BuildComponents(inf *CrudInfo) error {

	inf.keyMap = make(map[string]interface{})
	inf.fldMap = make(map[string]interface{}) // string)
	inf.resultMap = make(map[string]interface{})

	// http://speakmy.name/2014/09/14/modifying-interfaced-go-struct/
	// get the underlying Type of the interface ptr
	inf.stype = reflect.TypeOf(inf.ent).Elem()
	if inf.log {
		log.Println("inf.stype:", inf.stype)
	}

	// check that the interface type passed in was a struct
	if inf.stype.Kind() != reflect.Struct {
		return fmt.Errorf("only struct{} types can be passed in for table creation.  got %s", inf.stype.Kind())
	}

	// read the tags for the struct underlying the interface ptr
	var err error
	inf.flDef, err = common.TagReader(inf.ent, inf.stype)
	if err != nil {
		log.Println("error reading model definition", err)
		return err
	}
	if inf.log {
		log.Println("inf.flDef:", inf.flDef)
	}

	// determine the table name as per the table creation logic
	inf.tn = common.GetTableName(inf.ent)

	inf.fList = "("
	inf.vList = "("

	// get the value that the interface ptr ent points to
	// i.e. the struct holding the data for insertion
	inf.entValue = reflect.ValueOf(inf.ent).Elem()
	if inf.log {
		log.Println("value of data in struct for insertion:", inf.entValue)
	}

	// what to do with sqac tags
	// primary key:inc - do not fill
	// primary key:""  - do nothing
	// default - DEFAULT keyword for field
	// nullable - if no and nil value, fill with default value for nullable type
	// insQuery = "INSERT INTO depot (depot_num, region, province) VALUES (DEFAULT,'YVR','AB');"
	// https: //stackoverflow.com/questions/18926303/iterate-through-the-fields-of-a-struct-in-go
	// entity-type in Create CRUD call: sqac_test.Depot
	// {depot_num  int false [{primary_key inc} {start 90000000}]}
	// {depot_bay  int false [{primary_key }]}
	// {create_date  time.Time false [{nullable false} {default now()} {index unique}]}
	// {region  string false [{nullable false} {default YYC}]}
	// {province  string false [{nullable false} {default AB}]}
	// {country  string false [{nullable true} {default CA}]}
	// {new_column1  string false [{nullable false}]}
	// {new_column2  int64 false [{nullable false}]}
	// {new_column3  float64 false [{nullable false} {default 0.0}]}
	// {non_persistent_column  string true []}

	// iterate over the entity-struct metadata
	for i, fd := range inf.flDef {
		if inf.log {
			log.Println(fd)
		}
		if fd.NoDB == true {
			continue
		}

		bDefault := false
		bPkeyInc := false
		bPkey := false
		bIsNull := false

		// set the field attribute indicators
		for _, t := range fd.SqacPairs {
			switch t.Name {
			case "primary_key":
				if t.Value == "inc" {
					bPkeyInc = true
					inf.incKeyName = fd.FName //MySQL :/
				} else {
					bPkey = true
				}
			case "default":
				bDefault = true
			case "nullable":
				// it is on the caller to know this
			default:

			}
		}

		// get the value of the current entity field
		fv := inf.entValue.Field(i).Interface()
		fvr := inf.entValue.Field(i)

		// is the struct member a pointer?
		if fvr.Kind() == reflect.Ptr {
			if fvr.IsNil() {
				bIsNull = true
			} else {
				fvr = fvr.Elem() // get the value
			}
		}

		switch fd.UnderGoType {
		case "int", "int8", "int16", "int32", "int64":
			if inf.mode == "C" {
				if bPkeyInc == true {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "DEFAULT")
					inf.fldMap[fd.FName] = "DEFAULT"
					continue
				}
				if bDefault == true && fv == 0 ||
					bDefault == true && bIsNull {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "DEFAULT")
					inf.fldMap[fd.FName] = "DEFAULT"
					continue
				}
			} else {
				if bPkeyInc == true || bPkey == true {
					inf.keyMap[fd.FName] = fvr.Int() // reflect.ValueOf(&fvr))??
					continue
				}
			}
			// in all other cases, just use the given value making the
			// assumption that the int-type field contains an int-type
			inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
			if !bIsNull {
				inf.vList = fmt.Sprintf("%s%v, ", inf.vList, fvr.Int())
				inf.fldMap[fd.FName] = fmt.Sprintf("%v", fvr.Int())
			} else {
				inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "NULL")
				inf.fldMap[fd.FName] = fmt.Sprintf("%s", "NULL")
			}
			continue

		case "uint", "uint8", "uint16", "uint32", "uint64", "rune", "byte":
			if inf.mode == "C" {
				if bPkeyInc == true {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "DEFAULT")
					inf.fldMap[fd.FName] = "DEFAULT"
					continue
				}
				if bDefault == true && reflect.DeepEqual(fv, reflect.Zero(reflect.TypeOf(fv)).Interface()) ||
					bDefault == true && bIsNull {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "DEFAULT")
					inf.fldMap[fd.FName] = "DEFAULT"
					continue
				}
			} else {
				if bPkeyInc == true || bPkey == true {
					inf.keyMap[fd.FName] = fvr.Uint()
					continue
				}
			}
			// in all other cases, just use the given value making the
			// assumption that the uint-type field contains a uint-type
			inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
			if !bIsNull {
				inf.vList = fmt.Sprintf("%s%d, ", inf.vList, fvr.Uint())
				inf.fldMap[fd.FName] = fmt.Sprintf("%d", fvr.Uint())
			} else {
				inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "NULL")
				inf.fldMap[fd.FName] = fmt.Sprintf("%s", "NULL")
			}
			continue

		case "float32", "float64":
			if inf.mode == "C" {
				if bPkeyInc == true {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "DEFAULT")
					inf.fldMap[fd.FName] = "DEFAULT"
					continue
				}
				if bDefault == true && reflect.DeepEqual(fv, reflect.Zero(reflect.TypeOf(fv)).Interface()) ||
					bDefault == true && bIsNull {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "DEFAULT")
					inf.fldMap[fd.FName] = "DEFAULT"
					continue
				}
			} else {
				if bPkeyInc == true || bPkey == true {
					inf.keyMap[fd.FName] = fvr.Float()
					continue
				}
			}
			// in all other cases, just use the given value making the
			// assumption that the float-type field contains a float-type
			inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
			if !bIsNull {
				inf.vList = fmt.Sprintf("%s%v, ", inf.vList, fvr.Float())
				inf.fldMap[fd.FName] = fmt.Sprintf("%v", fvr.Float())
			} else {
				inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "NULL")
				inf.fldMap[fd.FName] = fmt.Sprintf("%s", "NULL")
			}
			continue

		case "string":
			if inf.mode == "C" {
				if bPkeyInc == true {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "DEFAULT")
					inf.fldMap[fd.FName] = "DEFAULT"
					continue
				}
				if bDefault == true && fv == "" ||
					bDefault == true && bIsNull {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "DEFAULT")
					inf.fldMap[fd.FName] = "DEFAULT"
					continue
				}
			} else {
				if bPkeyInc == true || bPkey == true {
					inf.keyMap[fd.FName] = fvr.String()
					continue
				}
			}
			// in all other cases, just use the given value making the
			// assumption that the string-type field contains a string-type
			inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
			if !bIsNull {
				inf.vList = fmt.Sprintf("%s'%s', ", inf.vList, fvr.String())
				inf.fldMap[fd.FName] = fmt.Sprintf("'%s'", fvr.String())
			} else {
				inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "NULL")
				inf.fldMap[fd.FName] = fmt.Sprintf("%s", "NULL")
			}
			continue

		case "bool":
			if bDefault == true && fv == "" ||
				bDefault == true && bIsNull {
				inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
				inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "DEFAULT")
				inf.fldMap[fd.FName] = "DEFAULT"
				continue
			}

			// in all other cases, just use the given value making the
			// assumption that the string-type field contains a bool-type
			inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
			if !bIsNull {
				switch bf.GetDBDriverName() {
				case "sqlite3":
					i := bf.BoolToDBBool(fvr.Bool())
					inf.vList = fmt.Sprintf("%s%d, ", inf.vList, *i)
					inf.fldMap[fd.FName] = fmt.Sprintf("%d", *i)
				case "mssql":
					switch fvr.Bool() {
					case true:
						inf.vList = fmt.Sprintf("%s%d, ", inf.vList, 1)
						inf.fldMap[fd.FName] = fmt.Sprintf("%d", 1)
					case false:
						inf.vList = fmt.Sprintf("%s%d, ", inf.vList, 0)
						inf.fldMap[fd.FName] = fmt.Sprintf("%d", 0)
					default:

					}

				default:
					inf.vList = fmt.Sprintf("%s%t, ", inf.vList, fvr.Bool())
					inf.fldMap[fd.FName] = fmt.Sprintf("%t", fvr.Bool())
				}
			} else {
				inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "NULL")
				inf.fldMap[fd.FName] = fmt.Sprintf("%s", "NULL")
			}
			continue

		case "time.Time":

			if inf.mode == "C" {
				if bPkeyInc == true {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "DEFAULT")
					inf.keyMap[fd.FName] = "DEFAULT"
					continue
				}

				// only insert with DEFAULT if a zero-value time.Time was provided or
				// if a nil value was passed for a *time.Time
				bZzeroTime := false
				// if fv == reflect.Zero(reflect.TypeOf(fv)).Interface() {
				if reflect.DeepEqual(fv, reflect.Zero(reflect.TypeOf(fv)).Interface()) {
					bZzeroTime = true
				}

				if bDefault == true && bZzeroTime || // 0001-01-01 00:00:00 +0000 UTC
					bDefault == true && bIsNull {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "DEFAULT")
					inf.fldMap[fd.FName] = "DEFAULT"
					continue
				}
			} else {

				// deal with time keys, as they are immutable in update scenario
				if bPkeyInc == true || bPkey == true {
					// inf.keyMap[fd.FName] = fv.(time.Time).Format(time.RFC3339)
					inf.keyMap[fd.FName] = bf.TimeToFormattedString(fv) // fv.(time.Time).Format("2006-01-02 15:04:05.999999-07:00")
					continue
				}
			}
			inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
			if !bIsNull {
				inf.vList = fmt.Sprintf("%s'%v', ", inf.vList, bf.TimeToFormattedString(fv)) // fv.(time.Time).Format("2006-01-02 15:04:05.999999-07:00"))
				inf.fldMap[fd.FName] = fmt.Sprintf("'%s'", bf.TimeToFormattedString(fv))     // fv.(time.Time).Format("2006-01-02 15:04:05.999999-07:00")
			} else {
				inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "NULL")
				inf.fldMap[fd.FName] = fmt.Sprintf("%s", "NULL")
			}
			continue

		default:
			log.Printf("%s with go-type %s is unsupported\n", fd.FName, fd.GoType)
			continue

		}
	}
	inf.fList = strings.TrimSuffix(inf.fList, ", ")
	inf.fList = inf.fList + ")"
	inf.vList = strings.TrimSuffix(inf.vList, ", ")
	inf.vList = inf.vList + ")"
	return nil

}

// TimeToFormattedString is used to format the provided time.Time
// or *time.Time value in the string format required for the
// connected db insert or update operation.  This method is called
// from within the CRUD ops for each db flavor, and could be added
// to the flavor-specific Query / Exec methods at some point.
func (bf *BaseFlavor) TimeToFormattedString(i interface{}) string {

	var t time.Time

	if i == nil {
		panic(fmt.Errorf("nil value passed to TimeToFormattedString()"))
	}

	switch i.(type) {
	case time.Time:
		t = i.(time.Time)
	case *time.Time:
		iValPtr := reflect.ValueOf(i)
		iVal := iValPtr.Elem()
		t = iVal.Interface().(time.Time)
	default:
		panic(fmt.Errorf("type %v is not permitted in TimeToFormattedString()", reflect.TypeOf(i)))
	}

	switch bf.GetDBDriverName() {
	case "postgres":
		return t.Format("2006-01-02 15:04:05.999999-07:00")

	case "mysql":
		return t.Format("2006-01-02 15:04:05")

	case "sqlite":
		return t.Format("2006-01-02 15:04:05")

	case "mssql":
		return t.Format("2006-01-02 15:04:05.9999999")

	case "hdb":
		return t.Format("2006-01-02 15:04:05.9999999")

	default:
		// most db's will take this and convert to UTC
		return t.Format("2006-01-02 15:04:05")
	}
}

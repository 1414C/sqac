package sqac

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
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
		fmt.Println("inf.stype:", inf.stype)
	}

	// check that the interface type passed in was a struct
	if inf.stype.Kind() != reflect.Struct {
		return fmt.Errorf("only struct{} types can be passed in for table creation.  got %s", inf.stype.Kind())
	}

	// read the tags for the struct underlying the interface ptr
	var err error
	inf.flDef, err = common.TagReader(inf.ent, inf.stype)
	if err != nil {
		fmt.Println("error reading model definition", err)
		return err
	}
	if inf.log {
		fmt.Println("inf.flDef:", inf.flDef)
	}

	// determine the table name as per the table creation logic
	inf.tn = common.GetTableName(inf.ent)

	inf.fList = "("
	inf.vList = "("

	// get the value that the interface ptr ent points to
	// i.e. the struct holding the data for insertion
	inf.entValue = reflect.ValueOf(inf.ent).Elem()
	if inf.log {
		fmt.Println("value of data in struct for insertion:", inf.entValue)
	}

	// what to do with rgen tags
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
			fmt.Println(fd)
		}
		if fd.NoDB == true {
			continue
		}

		bDefault := false
		bPkeyInc := false
		bPkey := false
		bIsNull := false

		// set the field attribute indicators
		for _, t := range fd.RgenPairs {
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
					inf.keyMap[fd.FName] = fvr.Int()
					continue
				}
			}
			// in all other cases, just use the given value making the
			// assumption that the int-type field contains an int-type
			inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
			if !bIsNull {
				inf.vList = fmt.Sprintf("%s%d, ", inf.vList, fvr.Int())
				inf.fldMap[fd.FName] = fmt.Sprintf("%d", fvr.Int())
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
				if bDefault == true && fv == 0 ||
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
				inf.vList = fmt.Sprintf("%s%f, ", inf.vList, fvr.Float())
				inf.fldMap[fd.FName] = fmt.Sprintf("%f", fvr.Float())
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
					inf.keyMap[fd.FName] = reflect.ValueOf(&fvr) //fvr.String()
					continue
				}
			}
			// in all other cases, just use the given value making the
			// assumption that the string-type field contains a string-type
			inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
			if !bIsNull {
				inf.vList = fmt.Sprintf("%s'%s', ", inf.vList, reflect.ValueOf(&fvr)) // fvr.String())
				inf.fldMap[fd.FName] = fmt.Sprintf("'%s'", reflect.ValueOf(&fvr))     // fvr.String())
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
					bDefault == true && bIsNull { // nil pointer case
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
					inf.keyMap[fd.FName] = reflect.ValueOf(&fvr) //fvr.String()
					continue
				}
			}
			// in all other cases, just use the given value making the
			// assumption that the string-type field contains a string-type
			inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
			if !bIsNull {
				if fd.UnderGoType != "string" {
					inf.vList = fmt.Sprintf("%s%v, ", inf.vList, reflect.ValueOf(&fvr)) // fvr.Int();fvr.UInt();fvr.Float();fvr.Bool()
					inf.fldMap[fd.FName] = fmt.Sprintf("%v", reflect.ValueOf(&fvr))     // fvr.Int();fvr.UInt();fvr.Float();fvr.Bool()
					continue
				}
				inf.vList = fmt.Sprintf("%s'%s', ", inf.vList, reflect.ValueOf(&fvr)) // fvr.String())
				inf.fldMap[fd.FName] = fmt.Sprintf("'%s'", reflect.ValueOf(&fvr))     // fvr.String())
			} else {
				inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "NULL")
				inf.fldMap[fd.FName] = fmt.Sprintf("%s", "NULL")
			}
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

// FormatReturn is used by CRUD operations to format
// the result-data from INSERT/UPDATE/GET's back into
// the go-format.  Notably, timestamps are stored as
// UTC where the db supports it, and this is used to
// ensure that the returned timestamp is presented in
// server-local format.
func (bf *BaseFlavor) FormatReturn(inf *CrudInfo) error {

	values := make([]interface{}, inf.entValue.NumField())
	for i := 0; i < inf.entValue.NumField(); i++ {
		values[i] = inf.entValue.Field(i).Interface()

		fn := inf.stype.Field(i).Name                // GoName
		st := inf.stype.Field(i).Tag                 // structTag
		ft, _ := inf.stype.Field(i).Tag.Lookup("db") // snake_name
		tp := inf.stype.Field(i).Type.String()       // field-type as String

		if inf.log {
			fmt.Println("NAME:", fn)
			// fmt.Println("TAG:", st)
			fmt.Println("DB FIELD NAME:", ft)
			fmt.Println("FIELD-TYPE:", tp)
		}

		// get the reflect.Value of the current field in the ent struct
		fv := reflect.ValueOf(inf.ent).Elem().FieldByName(fn)
		if !fv.IsValid() {
			panic(fmt.Errorf("invalid field %s in struct %s", fn, st))
		}

		// deal with pointer members in the target struct
		if fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
		}

		// check if the reflect.Value can be updated and set the returned
		// db field value from the resultMap.
		if fv.CanSet() {
			bBlankField := false
			np, _ := inf.stype.Field(i).Tag.Lookup("rgen")
			if strings.Contains(np, "-") && len(np) == 1 {
				bBlankField = true
			}

			// this is where go is pedantic, as type-assertions rely on compile-time
			// constants in the .(<type>) expression.
			bByteVal := false
			switch vt := inf.resultMap[ft].(type) {
			case []byte:
				bByteVal = true
				_ = vt // go is awkward
			default:
				bByteVal = false
			}

			switch tp {
			case "int", "int8", "int16", "int32", "int64":
				if !bBlankField {
					// fmt.Printf("field-name: %s, go type: %s\n", ft, tp)
					if bByteVal {
						s := fmt.Sprintf("%s", inf.resultMap[ft].([]byte))
						f, _ := strconv.ParseInt(s, 10, 64)
						fv.SetInt(f)
					} else {
						fv.SetInt(inf.resultMap[ft].(int64))
					}
				} else {
					fv.SetInt(0)
				}

			case "uint", "uint8", "uint16", "uint32", "uint64":
				if !bBlankField {
					if bByteVal {
						s := fmt.Sprintf("%s", inf.resultMap[ft].([]byte))
						f, _ := strconv.ParseUint(s, 10, 64)
						fv.SetUint(f)
					} else {
						switch inf.resultMap[ft].(type) {
						case uint64:
							fv.SetUint(inf.resultMap[ft].(uint64))

						case int64:
							fv.SetUint(uint64(inf.resultMap[ft].(int64)))

						default:
							fv.SetUint(inf.resultMap[ft].(uint64))
						}
					}
				} else {
					fv.SetUint(0)
				}

			case "rune":
				if !bBlankField {
					fv.Set(reflect.ValueOf(inf.resultMap[ft].(rune)))
				} else {
					fv.SetUint(0)
				}

			case "byte":
				if !bBlankField {
					fv.Set(reflect.ValueOf(inf.resultMap[ft].(byte)))
				} else {
					fv.SetUint(0)
				}

			case "float32", "float64":
				if !bBlankField {
					if bByteVal {
						s := fmt.Sprintf("%s", inf.resultMap[ft].([]byte))
						f, _ := strconv.ParseFloat(s, 64)
						fv.SetFloat(f)
					} else {
						fv.SetFloat(inf.resultMap[ft].(float64))
					}
				} else {
					fv.SetFloat(0)
				}

			case "string", "*string":
				if !bBlankField {
					if bByteVal {
						s := fmt.Sprintf("%s", inf.resultMap[ft].([]byte))
						fv.SetString(s)
					} else {
						fv.SetString(inf.resultMap[ft].(string))
					}
				} else {
					fv.SetString("")
				}

			case "bool":
				if !bBlankField {
					if bByteVal {
						s := fmt.Sprintf("%s", inf.resultMap[ft].([]byte))
						switch s {
						case "0", "false", "FALSE":
							fv.SetBool(false)
						case "1", "true", "TRUE":
							fv.SetBool(true)
						default:

						}
					} else {
						switch bf.GetDBDriverName() {
						case "hdb":
							b := bf.DBBoolToBool(inf.resultMap[ft])
							fv.SetBool(b)
						default:
							fv.Set(reflect.ValueOf(inf.resultMap[ft].(bool)))
						}
					}
				} else {
					fv.SetBool(false) // nullable?
				}

			case "time.Time":
				if !bBlankField {
					// fv.Set(reflect.ValueOf(inf.resultMap[ft].(time.Time).Local()))
					fv.Set(reflect.ValueOf(inf.resultMap[ft].(time.Time)))
				} else {
					fv.SetInt(0)
				}

			case "*time.Time":
				if !bBlankField {
					fv.Set(reflect.ValueOf(inf.resultMap[ft].(*time.Time)))
				} else {
					fv.SetInt(0)
				}
			default:
				log.Printf("UNSUPPORTED TYPE:%s\n", tp)
				// one could try something like this:
				// fv.Set(reflect.ValueOf(resultMap[ft].(stype.Field(i).Type)))
			}
		} else {
			if inf.log {
				log.Printf("CANNOT SET %s:\n", fn)
			}
			log.Printf("CANNOT SET %s:\n", fn)
		}
	}
	if inf.log {
		fmt.Println("populated entity:", inf.ent)
	}
	return nil
}

// FormatReturn2 is used by CRUD operations to format
// the result-data from INSERT/UPDATE/GET's back into
// the go-format.  Notably, timestamps are stored as
// UTC where the db supports it, and this is used to
// ensure that the returned timestamp is presented in
// server-local format.
func (bf *BaseFlavor) FormatReturn2(inf *CrudInfo) error {

	for k, v := range inf.resultMap {
		fmt.Printf("k: %s, v: %v\n", k, v)
	}
	fmt.Printf("inf.ent: %v\n", inf.ent)

	for i, flDef := range inf.flDef {

		// get the value of the current entity field
		// tstfv := inf.entValue.Field(i).Interface()
		tstfvr := inf.entValue.Field(i)
		// fmt.Println("FV:", fv)
		// fmt.Println("FVR:", fvr)

		// is the struct member a pointer?
		if tstfvr.Kind() == reflect.Ptr {
			// fmt.Printf("%s is a pointer!\n", fd.FName)
			// bIsPointer = true
			if tstfvr.IsNil() {
				fmt.Println("tstfvr is nil!", flDef.FName)
				if inf.resultMap[flDef.FName] != nil {
					fmt.Printf("%s is not not nil: %v\n", flDef.FName, inf.resultMap[flDef.FName])
				}
				//var t time.Time
				// bIsNull = true
			} else {
				tstfvr = tstfvr.Elem() // get the value
			}
		}

		if inf.log {
			fmt.Println("======================================================")
			fmt.Println("flDef: ", flDef)
			fmt.Println("bf.FormatReturn")
			fmt.Println("GoFieldName:", flDef.GoName)
			fmt.Println("snake_name:", flDef.FName)
			fmt.Println("go-field-type:", flDef.GoType)
		}

		// get the reflect.Value of the current field in the ent struct
		fv := reflect.ValueOf(inf.ent).Elem().FieldByName(flDef.GoName)
		if !fv.IsValid() {
			panic(fmt.Errorf("invalid field %s in struct %s", flDef.GoName, "foo"))
		}
		// if flDef.GoName == "TimeNullWithDefault" {
		fmt.Printf("fv: %v, FName: %s\n", fv, flDef.FName)
		//}

		// deal with pointer members in the target struct
		if fv.Kind() == reflect.Ptr {
			fmt.Println("fv.Kind is pointer:", flDef.FName)
			fmt.Println("fv.CanSet()", fv.CanSet())
			fv = fv.Elem()
			fmt.Printf("fv: %v, FName: %s\n", fv, flDef.FName)
			fmt.Println("fv.CanSet()", fv.CanSet())
		}

		if fv.CanSet() {

			// this is where go is pedantic, as type-assertions rely on compile-time
			// constants in the .(<type>) expression.
			bByteVal := false
			switch vt := inf.resultMap[flDef.FName].(type) {
			case []byte:
				bByteVal = true
				_ = vt // go is awkward
			default:
				bByteVal = false
			}

			switch flDef.UnderGoType {
			case "int", "int8", "int16", "int32", "int64":
				if !flDef.NoDB {
					// fmt.Printf("field-name: %s, go type: %s\n", ft, tp)
					if bByteVal {
						s := fmt.Sprintf("%s", inf.resultMap[flDef.FName].([]byte))
						f, _ := strconv.ParseInt(s, 10, 64)
						fv.SetInt(f)
					} else {
						fv.SetInt(inf.resultMap[flDef.FName].(int64))
					}
				} else {
					fv.SetInt(0)
				}

			case "uint", "uint8", "uint16", "uint32", "uint64":
				if !flDef.NoDB {
					if bByteVal {
						s := fmt.Sprintf("%s", inf.resultMap[flDef.FName].([]byte))
						f, _ := strconv.ParseUint(s, 10, 64)
						fv.SetUint(f)
					} else {
						switch inf.resultMap[flDef.FName].(type) {
						case uint64:
							fv.SetUint(inf.resultMap[flDef.FName].(uint64))

						case int64:
							fv.SetUint(uint64(inf.resultMap[flDef.FName].(int64)))

						default:
							fv.SetUint(inf.resultMap[flDef.FName].(uint64))
						}
					}
				} else {
					fv.SetUint(0)
				}

			case "rune":
				if !flDef.NoDB {
					fv.Set(reflect.ValueOf(inf.resultMap[flDef.FName].(rune)))
				} else {
					fv.SetUint(0)
				}

			case "byte":
				if !flDef.NoDB {
					fv.Set(reflect.ValueOf(inf.resultMap[flDef.FName].(byte)))
				} else {
					fv.SetUint(0)
				}

			case "float32", "float64":
				if !flDef.NoDB {
					if bByteVal {
						s := fmt.Sprintf("%s", inf.resultMap[flDef.FName].([]byte))
						f, _ := strconv.ParseFloat(s, 64)
						fv.SetFloat(f)
					} else {
						fv.SetFloat(inf.resultMap[flDef.FName].(float64))
					}
				} else {
					fv.SetFloat(0)
				}

			case "string", "*string":
				if !flDef.NoDB {
					if bByteVal {
						s := fmt.Sprintf("%s", inf.resultMap[flDef.FName].([]byte))
						fv.SetString(s)
					} else {
						fv.SetString(inf.resultMap[flDef.FName].(string))
					}
				} else {
					fv.SetString("")
				}

			case "bool":
				if !flDef.NoDB {
					if bByteVal {
						s := fmt.Sprintf("%s", inf.resultMap[flDef.FName].([]byte))
						switch s {
						case "0", "false", "FALSE":
							fv.SetBool(false)
						case "1", "true", "TRUE":
							fv.SetBool(true)
						default:

						}
					} else {
						switch bf.GetDBDriverName() {
						case "hdb":
							b := bf.DBBoolToBool(inf.resultMap[flDef.FName])
							fv.SetBool(b)
						default:
							fv.Set(reflect.ValueOf(inf.resultMap[flDef.FName].(bool)))
						}
					}
				} else {
					fv.SetBool(false) // nullable?
				}

			case "time.Time":
				if !flDef.NoDB {
					// fv.Set(reflect.ValueOf(inf.resultMap[ft].(time.Time).Local()))
					fv.Set(reflect.ValueOf(inf.resultMap[flDef.FName].(time.Time)))
				} else {
					fv.SetInt(0)
				}

			case "*time.Time":
				if !flDef.NoDB {
					fv.Set(reflect.ValueOf(inf.resultMap[flDef.FName].(*time.Time)))
				} else {
					fv.SetInt(0)
				}
			default:
				log.Printf("UNSUPPORTED TYPE:%s\n", flDef.GoType)
				// one could try something like this:
				// fv.Set(reflect.ValueOf(resultMap[ft].(stype.Field(i).Type)))
			}
		} else {
			if inf.log {
				log.Printf("CANNOT SET %s:\n", flDef.GoName)
			}
			log.Printf("CANNOT SET %s:\n", flDef.GoName)
		}
	}
	if inf.log {
		fmt.Println("populated entity:", inf.ent)
	}
	return nil
}

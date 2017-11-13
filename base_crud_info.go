package sqac

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
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
	flDef      []FieldDef
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
	inf.flDef, err = TagReader(inf.ent, inf.stype)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if inf.log {
		fmt.Println("inf.flDef:", inf.flDef)
	}

	// determine the table name as per the table creation logic
	inf.tn = reflect.TypeOf(inf.ent).String()
	if strings.Contains(inf.tn, ".") {
		el := strings.Split(inf.tn, ".")
		inf.tn = strings.ToLower(el[len(el)-1])
	} else {
		inf.tn = strings.ToLower(inf.tn)
	}

	// insQuery := fmt.Sprintf("INSERT INTO %s", tn)
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
		bNullable := false

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
				if t.Value == "true" || t.Value == "TRUE" {
					bNullable = true
				}
			default:

			}
		}

		// get the value of the current entity field
		if bPkey {
			fmt.Println("")
		}
		fv := inf.entValue.Field(i).Interface()
		fvr := inf.entValue.Field(i)
		switch fd.GoType {
		case "int", "uint", "int8", "uint8", "int16", "uint16", "int32", "uint32", "int64", "uint64", "rune", "byte":
			if inf.mode == "C" {
				if bPkeyInc == true {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "DEFAULT")
					inf.fldMap[fd.FName] = "DEFAULT"
					continue
				}
				if bDefault == true && fv == 0 ||
					bDefault == true && fv == nil {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "DEFAULT")
					inf.fldMap[fd.FName] = "DEFAULT"
					continue
				}
				if bNullable == false && fv == nil {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%d, ", inf.vList, 0)
					inf.fldMap[fd.FName] = "0"
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
			inf.vList = fmt.Sprintf("%s%d, ", inf.vList, fvr.Int())
			inf.fldMap[fd.FName] = fmt.Sprintf("%d", fvr.Int())
			continue

		case "float32", "float64":
			if inf.mode == "C" {
				if bPkeyInc == true {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "DEFAULT")
					inf.fldMap[fd.FName] = "DEFAULT"
					continue
				}
				if bDefault == true && fv == 0 ||
					bDefault == true && fv == nil {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "DEFAULT")
					inf.fldMap[fd.FName] = "DEFAULT"
					continue
				}
				if bNullable == false && fv == nil {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%f, ", inf.vList, 0.0)
					inf.fldMap[fd.FName] = "0.0"
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
			inf.vList = fmt.Sprintf("%s%f, ", inf.vList, fvr.Float())
			inf.fldMap[fd.FName] = fmt.Sprintf("%f", fvr.Float())
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
					bDefault == true && fv == nil {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "DEFAULT")
					inf.fldMap[fd.FName] = "DEFAULT"
					continue
				}
				if bNullable == false && fv == nil {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "''")
					inf.fldMap[fd.FName] = "''"
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
			inf.vList = fmt.Sprintf("%s'%s', ", inf.vList, fvr.String())
			inf.fldMap[fd.FName] = fmt.Sprintf("'%s'", fvr.String())
			continue

		case "time.Time", "*time.Time":

			if inf.mode == "C" {
				if bPkeyInc == true {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "DEFAULT")
					inf.keyMap[fd.FName] = "DEFAULT"
					continue
				}

				// only insert with DEFAULT if a zero-value time.Time was provided
				bZzeroTime := false
				if fd.GoType == "time.Time" {
					var tt time.Time
					tt = fv.(time.Time)
					if tt.IsZero() {
						bZzeroTime = true
					}
				} else {
					var tt *time.Time
					tt = fv.(*time.Time)
					if tt.IsZero() {
						bZzeroTime = true
					}
				}

				if bDefault == true && bZzeroTime { // 0001-01-01 00:00:00 +0000 UTC
					// fmt.Printf("time.Time: %v\n", fv)
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "DEFAULT")
					inf.fldMap[fd.FName] = "DEFAULT"
					continue
				}
				if bNullable == false && fv == nil {
					inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
					inf.vList = fmt.Sprintf("%s%s, ", inf.vList, "bot") //"make_timestamptz(0000, 00, 00, 00, 00, 00.0")
					inf.fldMap[fd.FName] = "bot"                        //"make_timestamptz(0000, 00, 00, 00, 00, 00.0"
					continue
				}
			} else {

				// deal with time keys
				if bPkeyInc == true || bPkey == true {
					// fmt.Println("fv:", fv)
					if fd.GoType == "time.Time" {
						// inf.keyMap[fd.FName] = fv.(time.Time).Format(time.RFC3339)
						inf.keyMap[fd.FName] = bf.TimeToFormattedString(fv.(time.Time)) // fv.(time.Time).Format("2006-01-02 15:04:05.999999-07:00")
					} else if fd.GoType == "*time.Time" {
						tDRef := *fv.(*time.Time)
						inf.keyMap[fd.FName] = bf.TimeToFormattedString(tDRef) // fv.(*time.Time).Format("2006-01-02 15:04:05.999999-07:00")
					}
					continue
				}
			}
			inf.fList = fmt.Sprintf("%s%s, ", inf.fList, fd.FName)
			if fd.GoType == "time.Time" {
				inf.vList = fmt.Sprintf("%s'%v', ", inf.vList, bf.TimeToFormattedString(fv.(time.Time))) // fv.(time.Time).Format("2006-01-02 15:04:05.999999-07:00"))
				inf.fldMap[fd.FName] = fmt.Sprintf("'%s'", bf.TimeToFormattedString(fv.(time.Time)))     // fv.(time.Time).Format("2006-01-02 15:04:05.999999-07:00")
			} else {
				tDRef := *fv.(*time.Time)
				inf.vList = fmt.Sprintf("%s'%v', ", inf.vList, bf.TimeToFormattedString(tDRef)) // fv.(*time.Time).Format("2006-01-02 15:04:05.999999-07:00"))
				inf.fldMap[fd.FName] = fmt.Sprintf("'%s'", bf.TimeToFormattedString(tDRef))     // fv.(*time.Time).Format("2006-01-02 15:04:05.999999-07:00")
			}
			continue

		default:

		}
	}
	inf.fList = strings.TrimSuffix(inf.fList, ", ")
	inf.fList = inf.fList + ")"
	inf.vList = strings.TrimSuffix(inf.vList, ", ")
	inf.vList = inf.vList + ")"
	return nil

}

// TimeToFormattedString is used to format the provided time.Time
// value in the string format required for the connected db
// insert or update operation.  This method is called from
// within the CRUD ops for each db flavor, and could be added
// to the flavor-specific Query / Exec methods at some point.
func (bf *BaseFlavor) TimeToFormattedString(t time.Time) string {

	switch bf.GetDBDriverName() {
	case "postgres":
		return t.Format("2006-01-02 15:04:05.999999-07:00")

	case "mysql":
		return t.Format("2006-01-02 15:04:05")

	case "sqlite":
		return t.Format("2006-01-02 15:04:05")

	case "mssql":
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
			fmt.Println("TAG:", st)
			fmt.Println("DB FIELD NAME:", ft)
			fmt.Println("FIELD-TYPE:", tp)
		}

		// get the reflect.Value of the current field in the ent struct
		fv := reflect.ValueOf(inf.ent).Elem().FieldByName(fn)
		if !fv.IsValid() {
			panic(fmt.Errorf("invalid field %s in struct %s", fn, st))
		}

		// check if the reflect.Value can be updated and set the returned
		// db field value from the resultMap.
		if fv.CanSet() {
			bBlankField := false
			np, _ := inf.stype.Field(i).Tag.Lookup("rgen")
			if strings.Contains(np, "-") {
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
						fv.SetInt(inf.resultMap[ft].(int64))
					}
				} else {
					fv.SetInt(0)
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
			case "string":
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
				fmt.Printf("UNSUPPORTED TYPE:%s\n", tp)
				// try
				// fv.Set(reflect.ValueOf(resultMap[ft].(stype.Field(i).Type)))

			}
		} else {
			fmt.Printf("CANNOT SET %s:\n", fn)
		}

	}
	if inf.log {
		fmt.Println(values)
		fmt.Println("ENT:", inf.ent)
	}
	return nil
}

package sqac

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/1414C/sqac/common"
)

// CreateBak - Create the entity (single-row) on the database
func (pf *PostgresFlavor) CreateBak(ent interface{}) error {

	var info CrudInfo
	info.ent = ent
	info.log = true
	info.mode = "C"
	info.keyMap = make(map[string]interface{})
	err := pf.BuildComponents(&info)
	fmt.Println(info)
	fmt.Println(info.ent)
	fmt.Println(info.fList)
	fmt.Println(info.vList)
	for k, s := range info.keyMap {
		fmt.Printf("key: %s, value: %v\n", k, s)
	}
	os.Exit(0)

	// http://speakmy.name/2014/09/14/modifying-interfaced-go-struct/
	// get the underlying Type of the interface ptr
	stype := reflect.TypeOf(ent).Elem()
	if pf.IsLog() {
		fmt.Println("stype:", stype)
	}

	// check that the interface type passed in was a struct
	if stype.Kind() != reflect.Struct {
		return fmt.Errorf("only struct{} types can be passed in for table creation.  got %s", stype.Kind())
	}

	// read the tags for the struct underlying the interface ptr
	flDef, err := common.TagReader(ent, stype)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if pf.IsLog() {
		fmt.Println("flDef:", flDef)
	}

	// determine the table name as per the table creation logic
	tn := reflect.TypeOf(ent).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

	insQuery := fmt.Sprintf("INSERT INTO %s", tn)
	fList := "("
	vList := "("

	// get the value that the interface ptr ent points to
	// i.e. the struct holding the data for insertion
	v := reflect.ValueOf(ent).Elem()
	if pf.IsLog() {
		fmt.Println("value of data in struct for insertion:", v)
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
	for i, fd := range flDef {
		if pf.IsLog() {
			fmt.Println(fd)
		}
		if fd.NoDB == true {
			continue
		}
		bDefault := false
		bPkeyInc := false
		bNullable := false

		// set the field attribute indicators
		for _, t := range fd.SqacPairs {
			switch t.Name {
			case "primary_key":
				if t.Value == "inc" {
					bPkeyInc = true
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
		fv := v.Field(i).Interface()
		fvr := v.Field(i)
		switch fd.GoType {
		case "int", "uint", "int8", "uint8", "int16", "uint16", "int32", "uint32", "int64", "uint64", "rune", "byte":
			if bPkeyInc == true {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}
			if bDefault == true && fv == 0 ||
				bDefault == true && fv == nil {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}
			if bNullable == false && fv == nil {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%d, ", vList, 0)
				continue
			}
			// in all other cases, just use the given value making the
			// assumption that the int-type field contains an int-type
			fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
			vList = fmt.Sprintf("%s%d, ", vList, fvr.Int())
			continue

		case "float32", "float64":
			if bPkeyInc == true {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}
			if bDefault == true && fv == 0 ||
				bDefault == true && fv == nil {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}
			if bNullable == false && fv == nil {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%f, ", vList, 0.0)
				continue
			}
			// in all other cases, just use the given value making the
			// assumption that the float-type field contains a float-type
			fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
			vList = fmt.Sprintf("%s%f, ", vList, fvr.Float())
			continue

		case "string":
			if bPkeyInc == true {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}
			if bDefault == true && fv == "" ||
				bDefault == true && fv == nil {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}
			if bNullable == false && fv == nil {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "''")
				continue
			}
			// in all other cases, just use the given value making the
			// assumption that the string-type field contains a string-type
			fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
			vList = fmt.Sprintf("%s'%s', ", vList, fv)
			continue

		case "time.Time", "*time.Time":
			if bPkeyInc == true {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}

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
				fmt.Printf("time.Time: %v\n", fv)
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}
			if bNullable == false && fv == nil {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "make_timestamptz(0000, 00, 00, 00, 00, 00.0")
				continue
			}
			fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
			vList = fmt.Sprintf("%s%v, ", vList, fv)
			continue

		default:

		}
	}

	// build the insert query string
	fList = strings.TrimSuffix(fList, ", ")
	fList = fmt.Sprintf("%s%s", fList, ")")
	vList = strings.TrimSuffix(vList, ", ")
	vList = fmt.Sprintf("%s%s", vList, ") RETURNING *;") // depot_num
	insQuery = fmt.Sprintf("%s %s VALUES %s", insQuery, fList, vList)
	if pf.IsLog() {
		fmt.Println(insQuery)
	}

	// attempt the insert and read result back into resultMap
	resultMap := make(map[string]interface{})
	err = pf.db.QueryRowx(insQuery).MapScan(resultMap) // SliceScan
	if err != nil {
		return err
	}

	if pf.IsLog() {
		for k, r := range resultMap {
			fmt.Println(k, r)
		}
		fmt.Println("TYPEOF ent:", reflect.TypeOf(ent)) // sqac_test.Depot
	}

	values := make([]interface{}, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i).Interface()

		fn := stype.Field(i).Name                // GoName
		st := stype.Field(i).Tag                 // structTag
		ft, _ := stype.Field(i).Tag.Lookup("db") // snake_name
		tp := stype.Field(i).Type.String()       // field-type as String

		if pf.IsLog() {
			fmt.Println("NAME:", fn)
			fmt.Println("TAG:", st)
			fmt.Println("DB FIELD NAME:", ft)
			fmt.Println("FIELD-TYPE:", tp)
		}

		// get the reflect.Value of the current field in the ent struct
		fv := reflect.ValueOf(ent).Elem().FieldByName(fn)
		if !fv.IsValid() {
			panic(fmt.Errorf("invalid field %s in struct %s", fn, st))
		}

		// check if the reflect.Value can be updated and set the returned
		// db field value from the resultMap.
		if fv.CanSet() {
			bBlankField := false
			np, _ := stype.Field(i).Tag.Lookup("sqac")
			if strings.Contains(np, "-") {
				bBlankField = true
			}

			switch tp {
			case "int", "int8", "int16", "int32", "int64":
				if !bBlankField {
					fv.SetInt(resultMap[ft].(int64))
				} else {
					fv.SetInt(0)
				}

			case "uint", "uint8", "uint16", "uint32", "uint64", "rune", "byte":
				if !bBlankField {
					fv.SetUint(resultMap[ft].(uint64))
				} else {
					fv.SetInt(0)
				}

			case "float32", "float64":
				if !bBlankField {
					s := fmt.Sprintf("%s", resultMap[ft].([]byte))
					f, err := strconv.ParseFloat(s, 64)
					if err != nil {
						fmt.Printf("%s", err)
					}
					if pf.IsLog() {
						fmt.Println("float value:", f)
					}
					fv.SetFloat(f)
				} else {
					fv.SetFloat(0)
				}

			case "string":
				if !bBlankField {
					fv.SetString(resultMap[ft].(string))
				} else {
					fv.SetString("")
				}

			case "time.Time":
				if !bBlankField {
					fv.Set(reflect.ValueOf(resultMap[ft].(time.Time)))
				} else {
					fv.SetInt(0)
				}

			case "*time.Time":
				if !bBlankField {
					fv.Set(reflect.ValueOf(resultMap[ft].(*time.Time)))
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
	if pf.IsLog() {
		fmt.Println(values)
		fmt.Println("ENT:", ent)
	}
	return nil
}

// UpdateBak - Update an existing entity (single-row) on the database
func (pf *PostgresFlavor) UpdateBak(ent interface{}) error {

	// isolate key(s) for WHERE clause
	// update all non-key fields with their values from the incoming struct
	//   - still need to consider DEFAULT / nullable etc.

	// example update query
	// UPDATE weather SET (temp_lo, temp_hi, prcp) = (temp_lo+1, temp_lo+15, DEFAULT)
	//   WHERE city = 'San Francisco' AND date = '2003-07-03' RETURNING *;

	// http://speakmy.name/2014/09/14/modifying-interfaced-go-struct/
	// get the underlying Type of the interface ptr
	stype := reflect.TypeOf(ent).Elem()
	if pf.IsLog() {
		fmt.Println("stype:", stype)
	}

	// check that the interface type passed in was a struct
	if stype.Kind() != reflect.Struct {
		return fmt.Errorf("only struct{} types can be passed in for table creation.  got %s", stype.Kind())
	}

	// read the tags for the struct underlying the interface ptr
	flDef, err := common.TagReader(ent, stype)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if pf.IsLog() {
		fmt.Println("flDef:", flDef)
	}

	// determine the table name as per the table creation logic
	tn := reflect.TypeOf(ent).String()
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}

	updQuery := fmt.Sprintf("UPDATE %s SET", tn)
	keyMap := make(map[string]interface{})
	fList := "("
	vList := "("

	// get the value that the interface ptr ent points to
	// i.e. the struct holding the data for the update
	v := reflect.ValueOf(ent).Elem()
	if pf.IsLog() {
		fmt.Println("value of data in struct for update:", v)
	}

	// what to do with sqac tags
	// primary key:inc - do not fill - add to keyList
	// primary key:""  - do not fill - add to keyList
	// default - DEFAULT keyword for field
	// nullable - if no and nil value, fill with default value for nullable type

	// iterate over the entity-struct metadata
	for i, fd := range flDef {
		if pf.IsLog() {
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
		for _, t := range fd.SqacPairs {
			switch t.Name {
			case "primary_key":
				if t.Value == "inc" {
					bPkeyInc = true
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
		fv := v.Field(i).Interface()
		fvr := v.Field(i)
		switch fd.GoType {
		case "int", "uint", "int8", "uint8", "int16", "uint16", "int32", "uint32", "int64", "uint64", "rune", "byte":
			if bPkeyInc == true || bPkey == true {
				keyMap[fd.FName] = fvr.Int()
				continue
			}
			if bDefault == true && fv == 0 {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}
			if bNullable == false && fv == nil {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%d, ", vList, 0)
				continue
			}
			// in all other cases, just use the given value making the
			// assumption that the int-type field contains an int-type
			fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
			vList = fmt.Sprintf("%s%d, ", vList, fvr.Int())
			continue

		case "float32", "float64":
			if bPkeyInc == true || bPkey == true {
				keyMap[fd.FName] = fvr.Float()
				continue
			}
			if bDefault == true && fv == 0 {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}
			if bNullable == false && fv == nil {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%f, ", vList, 0.0)
				continue
			}
			// in all other cases, just use the given value making the
			// assumption that the float-type field contains a float-type
			fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
			vList = fmt.Sprintf("%s%f, ", vList, fvr.Float())
			continue

		case "string":
			if bPkeyInc == true || bPkey == true {
				keyMap[fd.FName] = fvr.String()
				continue
			}
			if bDefault == true && fv == "" {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}
			if bNullable == false && fv == nil {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "''")
				continue
			}
			// in all other cases, just use the given value making the
			// assumption that the string-type field contains a string-type
			fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
			vList = fmt.Sprintf("%s'%s', ", vList, fv)
			continue

		case "time.Time", "*time.Time":
			if bPkeyInc == true || bPkey == true {
				keyMap[fd.FName] = fv
				continue
			}
			if bDefault == true {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "DEFAULT")
				continue
			}
			if bNullable == false && fv == nil {
				fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
				vList = fmt.Sprintf("%s%s, ", vList, "make_timestamptz(0000, 00, 00, 00, 00, 00.0")
				continue
			}
			fList = fmt.Sprintf("%s%s, ", fList, fd.FName)
			vList = fmt.Sprintf("%s%v, ", vList, fv)
			continue

		default:

		}
	}

	// build the update query string
	// UPDATE weather SET (temp_lo, temp_hi, prcp) = (temp_lo+1, temp_lo+15, DEFAULT)
	//   WHERE city = 'San Francisco' AND date = '2003-07-03' RETURNING *;
	fList = strings.TrimSuffix(fList, ", ")
	fList = fmt.Sprintf("%s%s", fList, ")")
	vList = strings.TrimSuffix(vList, ", ")
	vList = fmt.Sprintf("%s%s", vList, ")")
	keyList := ""

	for k, s := range keyMap {

		fType := reflect.TypeOf(s).String()
		if pf.IsLog() {
			fmt.Printf("key: %v, value: %v\n", k, s)
			fmt.Println("TYPE:", fType)
		}

		if fType == "string" {
			keyList = fmt.Sprintf("%s %s = '%v' AND", keyList, k, s)
		} else {
			keyList = fmt.Sprintf("%s %s = %v AND", keyList, k, s)
		}
	}
	keyList = strings.TrimSuffix(keyList, " AND")
	keyList = keyList + " RETURNING *;"
	updQuery = fmt.Sprintf("%s %s = %s WHERE%s", updQuery, fList, vList, keyList)
	fmt.Println(updQuery)

	// attempt the update and read result back into resultMap
	resultMap := make(map[string]interface{})
	err = pf.db.QueryRowx(updQuery).MapScan(resultMap) // SliceScan
	if err != nil {
		fmt.Println(err) //?
	}

	if pf.IsLog() {
		for k, r := range resultMap {
			fmt.Println(k, r)
		}
		fmt.Println("TYPEOF ent:", reflect.TypeOf(ent)) // sqac_test.Depot
	}

	values := make([]interface{}, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i).Interface()

		fn := stype.Field(i).Name                // GoName
		st := stype.Field(i).Tag                 // structTag
		ft, _ := stype.Field(i).Tag.Lookup("db") // snake_name
		tp := stype.Field(i).Type.String()       // field-type as String

		if pf.IsLog() {
			fmt.Println("NAME:", fn)
			fmt.Println("TAG:", st)
			fmt.Println("DB FIELD NAME:", ft)
			fmt.Println("FIELD-TYPE:", tp)
		}

		// get the reflect.Value of the current field in the ent struct
		fv := reflect.ValueOf(ent).Elem().FieldByName(fn)
		if !fv.IsValid() {
			panic(fmt.Errorf("invalid field %s in struct %s", fn, st))
		}

		// check if the reflect.Value can be updated and set the returned
		// db field value from the resultMap.
		if fv.CanSet() {
			bBlankField := false
			np, _ := stype.Field(i).Tag.Lookup("sqac")
			if strings.Contains(np, "-") {
				bBlankField = true
			}

			switch tp {
			case "int", "int8", "int16", "int32", "int64":
				if !bBlankField {
					fv.SetInt(resultMap[ft].(int64))
				} else {
					fv.SetInt(0)
				}

			case "uint", "uint8", "uint16", "uint32", "uint64", "rune", "byte":
				if !bBlankField {
					fv.SetUint(resultMap[ft].(uint64))
				} else {
					fv.SetInt(0)
				}

			case "float32", "float64":
				if !bBlankField {
					s := fmt.Sprintf("%s", resultMap[ft].([]byte))
					f, err := strconv.ParseFloat(s, 64)
					if err != nil {
						fmt.Printf("%s", err)
					}
					if pf.IsLog() {
						fmt.Println("float value:", f)
					}
					fv.SetFloat(f)
				} else {
					fv.SetFloat(0)
				}

			case "string":
				if !bBlankField {
					fv.SetString(resultMap[ft].(string))
				} else {
					fv.SetString("")
				}

			case "time.Time":
				if !bBlankField {
					fv.Set(reflect.ValueOf(resultMap[ft].(time.Time)))
				} else {
					fv.SetInt(0)
				}

			case "*time.Time":
				if !bBlankField {
					fv.Set(reflect.ValueOf(resultMap[ft].(*time.Time)))
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
	if pf.IsLog() {
		fmt.Println(values)
		fmt.Println("ENT:", ent)
	}
	return nil
}

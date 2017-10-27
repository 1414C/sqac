package sqac

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/1414C/sqlxtest/dbgen/common"
)

// RgenPair holds name-value-pairs for db field attributes
type RgenPair struct {
	Name  string
	Value string
}

// FieldDef holds field-names and a slice of their db attributes
type FieldDef struct {
	FName     string
	FType     string
	GoType    string
	NoDB      bool
	RgenPairs []RgenPair
}

// TagReader reads the `db:`, `rgen:` and (maybe)`sql:` tags and returns
// an array of type FieldDef.
func TagReader(i interface{}, t reflect.Type) (fd []FieldDef, err error) {

	if t == nil {
		t = reflect.TypeOf(i) // ProfileHeader{} for example
	}

	// check that the interface type passed in was a struct
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("only struct{} types can be passed in for table creation.  got %s", t.Kind())
	}

	for i := 0; i < t.NumField(); i++ {

		// dealing with a basic field, or an embedded struct?
		// get the field-type as a string
		fts := t.Field(i).Type.String()
		if fts != "uint" && fts != "uint8" && fts != "uint16" && fts != "uint32" && fts != "uint64" &&
			fts != "int" && fts != "int8" && fts != "int16" && fts != "int32" && fts != "int64" &&
			fts != "rune" && fts != "byte" && fts != "string" && fts != "float32" && fts != "float64" &&
			fts != "bool" && fts != "time.Time" && fts != "*time.Time" {

			// embedded struct - recurse and append resulting field defs
			// get the Value from the StructField (t.Field(i))
			// for example: {Admin  exp.Admin  88 [7] true}
			fv := reflect.ValueOf(t.Field(i))

			// get the reflection.Type of the field:
			// for example: exp.Location
			// and then get the db: and rgen: tags
			ftr := t.Field(i).Type
			// for z := 0; z < ftr.NumField(); z++ {
			// 	fmt.Println("db:", ftr.Field(z).Tag.Get("db"))
			// 	fmt.Println("rgen:", ftr.Field(z).Tag.Get("rgen"))
			// }

			// recursively call the TagReader(interface{}, reflect.Type)
			es, err := TagReader(fv.Interface(), ftr)
			if err != nil {
				return nil, fmt.Errorf("unable to parse embedded struct of %s %s", fv.Type(), fv.Interface())
			}
			fd = append(fd, es...)
			continue
		}

		var fldDef FieldDef
		fldDef.RgenPairs = nil

		// fldDef.FName should be the same as: t.Field(i).Tag.Get("db")
		// based on the common use of the CamelToSnake function in the
		// go struct generation (tag `db:"field_name"`).  The function
		// is used here to make that point.
		fldDef.FName = common.CamelToSnake(t.Field(i).Name)
		fldDef.GoType = fts // go-type here

		// get the other field-level db attributes
		rgenTag := t.Field(i).Tag.Get("rgen")
		if rgenTag != "" {
			rgenTags := strings.Split(rgenTag, ";")
			for k := range rgenTags {
				rgenVars := strings.Split(rgenTags[k], ":")
				switch len(rgenVars) {
				case 2:
					p := RgenPair{
						Name:  rgenVars[0],
						Value: rgenVars[1],
					}
					fldDef.RgenPairs = append(fldDef.RgenPairs, p)
					fldDef.NoDB = false
					rgenVars = nil
				case 1:
					if rgenVars[0] == "-" {
						fldDef.NoDB = true
					}
				}
				// if len(rgenVars) == 2 {
				// 	p := RgenPair{
				// 		Name:  rgenVars[0],
				// 		Value: rgenVars[1],
				// 	}
				// 	fldDef.RgenPairs = append(fldDef.RgenPairs, p)
				// 	rgenVars = nil
				// } else {

				// }
			}
		}
		fd = append(fd, fldDef)
	}
	return fd, nil
}

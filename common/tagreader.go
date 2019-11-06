package common

import (
	"fmt"
	"reflect"
	"strings"
	// "github.com/1414C/sqlxtest/dbgen/common"
)

// SqacPair holds name-value-pairs for db field attributes
type SqacPair struct {
	Name  string
	Value string
}

// FieldDef holds field-names and a slice of their db attributes
type FieldDef struct {
	FName       string
	FType       string
	GoName      string
	GoType      string
	UnderGoType string // underlying go-type (strip *)
	NoDB        bool
	SqacPairs   []SqacPair
}

// TagReader reads the `db:`, `sqac:` and (maybe)`sql:` tags and returns
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
		ftu := strings.TrimPrefix(fts, "*")

		// this would be cleaner for embedded structs, but time.Time is a struct etc..
		// if t.Field(i).Type.Kind() == reflect.Struct {
		// }
		if ftu != "uint" && ftu != "uint8" && ftu != "uint16" && ftu != "uint32" && ftu != "uint64" &&
			ftu != "int" && ftu != "int8" && ftu != "int16" && ftu != "int32" && ftu != "int64" &&
			ftu != "rune" && ftu != "byte" && ftu != "string" && ftu != "float32" && ftu != "float64" &&
			ftu != "bool" && ftu != "time.Time" {

			// embedded struct - recurse and append resulting field defs
			// get the Value from the StructField (t.Field(i))
			// for example: {Admin  exp.Admin  88 [7] true}
			fv := reflect.ValueOf(t.Field(i))

			// get the reflection.Type of the field:
			// for example: exp.Location
			// and then get the db: and sqac: tags
			ftr := t.Field(i).Type
			// for z := 0; z < ftr.NumField(); z++ {
			// 	fmt.Println("db:", ftr.Field(z).Tag.Get("db"))
			// 	fmt.Println("sqac:", ftr.Field(z).Tag.Get("sqac"))
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
		fldDef.SqacPairs = nil

		// fldDef.FName should be the same as: t.Field(i).Tag.Get("db")
		// based on the common use of the CamelToSnake function in the
		// go struct generation (tag `db:"field_name"`).  The function
		// is used here to make that point.
		fldDef.FName = CamelToSnake(t.Field(i).Name)
		fldDef.GoName = t.Field(i).Name
		fldDef.GoType = fts      // go-type here
		fldDef.UnderGoType = ftu // underlying type of pointer

		// get the other field-level db attributes
		sqacTag := t.Field(i).Tag.Get("sqac")
		if sqacTag != "" {
			sqacTags := strings.Split(sqacTag, ";")
			for k := range sqacTags {
				sqacVars := strings.Split(sqacTags[k], ":")
				switch len(sqacVars) {
				case 2:
					p := SqacPair{
						Name:  sqacVars[0],
						Value: sqacVars[1],
					}
					fldDef.SqacPairs = append(fldDef.SqacPairs, p)
					fldDef.NoDB = false
					// sqacVars = nil
				case 1:
					if sqacVars[0] == "-" {
						fldDef.NoDB = true
					}
				default:
					// do nothing
				}
			}
		}
		fd = append(fd, fldDef)
	}
	return fd, nil
}

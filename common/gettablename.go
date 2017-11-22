package common

import (
	"reflect"
	"strings"
)

// GetTableName determines the db table-name based on interface{} i
func GetTableName(i interface{}) string {

	tn := reflect.TypeOf(i).String() // Profile{} for example
	if strings.Contains(tn, ".") {
		el := strings.Split(tn, ".")
		tn = strings.ToLower(el[len(el)-1])
	} else {
		tn = strings.ToLower(tn)
	}
	return tn
}

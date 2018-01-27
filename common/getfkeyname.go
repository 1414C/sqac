package common

import (
	"fmt"
)

// GetFKeyName can be used to determine the foreign-key name based on a set
// of input fields.  Note that this function does not guarantee or check
// for the existance of the foreign-key; it simply provides the name that
// would have been used for the given parameter values.
// i:  Model{}
// ft: From Table
// rt: Reference Table
// ff: From Field
// rf: Reference Field
func GetFKeyName(i interface{}, ft, rt, ff, rf string) (string, error) {

	// very simple checks
	if ft == "" && i != nil {
		ft = GetTableName(i)
	}

	if ft == "" || rt == "" || ff == "" || rf == "" {
		return "", fmt.Errorf("provide all required parameters for common.GetFKeyName: got ft: %s, rt: %s, ff: %s, rf: %s", ft, rt, ff, rf)
	}

	fkn := fmt.Sprintf("fk_%s_%s_%s", ft, rt, rf)
	return fkn, nil
}

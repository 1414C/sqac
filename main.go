package sqac

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

// Sqac is the main access structure for the
// sqac library.
type Sqac struct {
	DB   *sqlx.DB
	Log  bool
	Hndl PublicDB
}

// Open a sqlx connection to the specified database
func Open(flavor string, args ...interface{}) (db *sqlx.DB, err error) {

	if len(args) != 1 {
		return nil, fmt.Errorf("incorrect number of args detected in Open()")
	}
	db, err = sqlx.Connect(flavor, args[0].(string))
	if err != nil {
		return nil, err
	}
	return db, nil
}

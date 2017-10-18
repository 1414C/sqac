package dbgen

import (
// "fmt"
)

// SQLiteFlavor is a sqlite3-specific implementation.
// Methods defined in the PublicDB interface of struct-type
// BaseFlavor are called by default for SQLiteFlavor. If
// the method as it exists in the BaseFlavor implementation
// is not compatible with the schema-syntax required by
// SQLite, the method in question may be overridden.
// Overriding (redefining) a BaseFlavor method may be
// accomplished through the addition of a matching method
// signature and implementation on the SQLiteFlavor
// struct-type.
type SQLiteFlavor struct {
	BaseFlavor

	//================================================================
	// possible local SQLite-specific overrides
	//================================================================
	// GetDBDriverName() string
	// CreateTables(i ...interface{}) error
	// DropTables(i ...interface{}) error
	// AlterTables(i ...interface{}) error
	// ExistsTable(i interface{}) bool
	// ExistsColumn(tn string, cn string, ct string) bool
	// CreateIndex(tn string, in string) error
	// DropIndex(tn string, in string) error
	// ExistsIndex(tn string, in string) bool
	// CreateSequence(sn string, start string) error
	// DropSequence(sn string) error
	// ExistsSequence(sn string) bool
}

// var pg_schema = `
// DROP TABLE IF EXISTS person;
// DROP TABLE IF EXISTS films;
// DROP TABLE IF EXISTS distributors;
// DROP TABLE IF EXISTS place;

// DROP SEQUENCE IF EXISTS dist_serial;
// DROP SEQUENCE IF EXISTS person_serial;

// CREATE TABLE films (
//     code        integer PRIMARY KEY,
//     title       varchar(40) NOT NULL,
//     did         integer NOT NULL,
//     date_prod   date,
//     kind        varchar(10),
//     len         interval hour to minute
// );

// CREATE SEQUENCE dist_serial START 10;

package sqac

import (
	"fmt"
	"log"

	_ "github.com/SAP/go-hdb/driver"
	"github.com/jmoiron/sqlx"
)

// Sqac is the main access structure for the
// sqac library.
type Sqac struct {
	// DB   *sqlx.DB
	// Log  bool
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

func Create(flavor string, logFlag bool, connectionString string) (handle PublicDB) {

	switch flavor {
	case "postgres":
		pgh := new(PostgresFlavor)
		handle = pgh
		db, err := Open("postgres", "host=127.0.0.1 user=godev dbname=sqlx sslmode=disable password=gogogo123")
		if err != nil {
			log.Fatalf("%s\n", err.Error())
		}
		handle.SetDB(db)
		// defer db.Close()

	case "mysql":
		myh := new(MySQLFlavor)
		handle = myh
		db, err := Open("mysql", "stevem:gogogo123@tcp(192.168.1.50:3306)/sqlx?charset=utf8&parseTime=True&loc=Local")
		if err != nil {
			log.Fatalf("%s\n", err.Error())
			panic(err)
		}
		handle.SetDB(db)
		// defer db.Close()

	case "sqlite":
		sqh := new(SQLiteFlavor)
		handle = sqh
		db, err := Open("sqlite3", "testdb.sqlite")
		if err != nil {
			log.Fatalf("%s\n", err.Error())
			panic(err)
		}
		handle.SetDB(db)
		// defer db.Close()

	case "mssql":
		msh := new(MSSQLFlavor)
		handle = msh
		db, err := Open("mssql", "sqlserver://SA:Bunny123!!@localhost:1401?database=sqlx")
		if err != nil {
			log.Fatalf("%s\n", err.Error())
			panic(err)
		}
		handle.SetDB(db)
		err = db.Ping()
		if err != nil {
			log.Fatalf("%s\n", err.Error())
			panic(err)
		}
		// defer db.Close()

	case "db2":

	case "hdb":
		hdh := new(HDBFlavor)
		handle = hdh
		db, err := Open("hdb", "hdb://SMACLEOD:Blockhead1@clkhana01.lab.clockwork.ca:30047")
		if err != nil {
			log.Fatalf("%s\n", err.Error())
			panic(err)
		}
		handle.SetDB(db)
		err = db.Ping()
		if err != nil {
			log.Fatalf("%s\n", err.Error())
			panic(err)
		}

	default:

	}
	fmt.Printf("HANDLE: %v\n", handle)
	// detailed logging?
	if logFlag {
		handle.Log(true)
	} else {
		handle.Log(false)
	}
	return handle
}

package sqac_test

import (
	"flag"
	"fmt"
	"github.com/1414C/sqac"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"testing"
)

// // SessionData contains session management vars
// type SessionData struct {
// 	db  *sqlx.DB
// 	log bool
// }
type dbac struct {
	DB   *sqlx.DB
	Log  bool
	Hndl sqac.PublicDB
}

var (
	dbAccess dbac
)

func TestMain(m *testing.M) {

	// parse flags
	dbFlag := flag.String("db", "pg", "db to connect to")
	// logFlag := flag.Bool("l", false, "activate sqac logging")
	flag.Parse()

	var err error

	// var err error
	switch *dbFlag {
	case "pg":
		dbAccess.Hndl = new(sqac.PostgresFlavor)
		dbAccess.Log = true
		dbAccess.DB, err = sqlx.Connect("postgres", "host=127.0.0.1 user=godev dbname=sqlx sslmode=disable password=gogogo123")
		if err != nil {
			log.Fatalf("%s\n", err.Error())
		}
		// dbAccess.Hndl.InBase()
		// dbAccess.Hndl.InDB()

	case "mysql":
		dbAccess.Hndl = new(sqac.MySQLFlavor)
		dbAccess.Log = true
		dbAccess.DB, err = sqlx.Connect("mysql", "stevem:gogogo123@tcp(192.168.1.50:3306)/jsonddl?charset=utf8&parseTime=True&loc=Local")
		if err != nil {
			log.Fatalf("%s\n", err.Error())
		}

	case "sqlite":
		dbAccess.Hndl = new(sqac.MySQLFlavor)
		dbAccess.Log = true
		dbAccess.DB, err = sqlx.Connect("sqlite3", "testdb.sqlite")
		if err != nil {
			log.Fatalf("%s\n", err.Error())
		}

	case "db2":

	default:

	}

	// run the tests
	code := m.Run()

	// cleanup

	os.Exit(code)
}

// TestGetAccountHolders attempts to read all accountholders from the db
//
// GET /accountholders
func TestCreateTables(t *testing.T) {

	fmt.Println("this is a test")
	dbAccess.Hndl.InBase()
	dbAccess.Hndl.InDB()
	// url := "https://localhost:8080/accountholders"
	// url := sessionData.baseURL + "/accountholders"
	// jsonStr := []byte(`{}`)
	// req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	// req.Close = true
	// req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("Authorization", "Bearer "+sessionData.jwtToken)

	// if sessionData.log {
	// 	fmt.Println("GET /accountholders request Headers:", req.Header)
	// }

	// // client := &http.Client{}
	// resp, err := sessionData.client.Do(req)
	// if err != nil {
	// 	t.Errorf("Test was unable to GET /accountholders. Got %s.\n", err.Error())
	// }
	// defer resp.Body.Close()

	// if sessionData.log {
	// 	fmt.Println("GET /accountholders response Status:", resp.Status)
	// 	fmt.Println("GET /accountholders response Headers:", resp.Header)
	// 	body, _ := ioutil.ReadAll(resp.Body)
	// 	fmt.Println("GET /accountholders response Body:", string(body))
	// }
}

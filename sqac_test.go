package sqac_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/1414C/sqac"

	"github.com/jmoiron/sqlx"

	"github.com/gen_test2/models"
)

// SessionData contains session management vars
type SessionData struct {
	db  *sqlx.DB
	log bool
}

var (
	h *sqac.BaseFlavor
)

// var a appobj.AppObj

func TestMain(m *testing.M) {

	// parse flags
	dbFlag := flag.String("db", "pg", "db to connect to")
	logFlag := flag.Bool("l", false, "activate sqac logging")
	flag.Parse()

	var err error
	switch *dbFlag {
	case "pg":
		h = new(sqac.PostgresFlavor)

	case "mysql":

	case "sqlite":

	case "db2":

	default:

	}
	sessionData.db, err = sqlx.Connect("postgres", "host=127.0.0.1 user=godev dbname=sqlx sslmode=disable password=gogogo123")
	// db, err := sqlx.Connect("sqlite3", "testdb.sqlite")
	if err != nil {
		log.Fatalf("%s\n", err.Error())
	}

	sessionData.log = *logFlag

	// run the tests
	code := m.Run()

	// cleanup

	os.Exit(code)
}

// TestGetAccountHolders attempts to read all accountholders from the db
//
// GET /accountholders
func TestCreateTables(t *testing.T) {

	// url := "https://localhost:8080/accountholders"
	switch sessionData.db.DriverName() {
	case "postgres":

	case "mysql":

	case "sqlite":

	default:

	}
	url := sessionData.baseURL + "/accountholders"
	jsonStr := []byte(`{}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sessionData.jwtToken)

	if sessionData.log {
		fmt.Println("GET /accountholders request Headers:", req.Header)
	}

	// client := &http.Client{}
	resp, err := sessionData.client.Do(req)
	if err != nil {
		t.Errorf("Test was unable to GET /accountholders. Got %s.\n", err.Error())
	}
	defer resp.Body.Close()

	if sessionData.log {
		fmt.Println("GET /accountholders response Status:", resp.Status)
		fmt.Println("GET /accountholders response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("GET /accountholders response Body:", string(body))
	}
}

// TestCreateAccountHolder attempts to create a new AccountHolder on the db
//
// POST /accountholder
func TestCreateAccountHolder(t *testing.T) {

	// url := "https://localhost:8080/accountholder"
	url := sessionData.baseURL + "/accountholder"

	var jsonStr = []byte(`{"name":"string_value",
"age":500000,
"weight":1900.99}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sessionData.jwtToken)

	if sessionData.log {
		fmt.Println("POST request Headers:", req.Header)
	}

	resp, err := sessionData.client.Do(req)
	if err != nil {
		t.Errorf("Test was unable to POST /accountholder. Got %s.\n", err.Error())
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("GET /accountholder response Body:", string(body))
		t.Errorf("Test was unable to POST /accountholder. Got %s.\n", err.Error())
	}
	defer resp.Body.Close()

	var e models.AccountHolder
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&e); err != nil {
		t.Errorf("Test was unable to decode the result of POST /product. Got %s.\n", err.Error())
	}

	//============================================================================================
	// TODO: implement validation of the returned entity here
	//============================================================================================
	if e.Name != "string_value" ||
		e.Age != 500000 ||
		e.Weight != 1900.99 {
		t.Errorf("inconsistency detected in POST /accountholder.")
	} else {
		sessionData.ID = e.ID
	}
}

// TestGetAccountHolder attempts to read accountholder/{:id} from the db
// using the id created in this entity's TestCreate function.
//
// GET /accountholder/{:id}
func TestGetAccountHolder(t *testing.T) {

	idStr := fmt.Sprint(sessionData.ID)
	// url := "https://localhost:8080/accountholder/" + idStr
	url := sessionData.baseURL + "/accountholder/" + idStr
	jsonStr := []byte(`{}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sessionData.jwtToken)

	if sessionData.log {
		fmt.Println("GET /accountholder request Headers:", req.Header)
	}

	// client := &http.Client{}
	resp, err := sessionData.client.Do(req)
	if err != nil {
		t.Errorf("Test was unable to GET /accountholder/%d. Got %s.\n", sessionData.ID, err.Error())
	}
	defer resp.Body.Close()

	if sessionData.log {
		fmt.Println("GET /accountholder response Status:", resp.Status)
		fmt.Println("GET /accountholder response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("GET /accountholder response Body:", string(body))
	}
}

// TestUpdateAccountHolder attempts to update an existing AccountHolder on the db
//
// PUT /accountholder/{:id}
func TestUpdateAccountHolder(t *testing.T) {

	idStr := fmt.Sprint(sessionData.ID)
	// url := "https://localhost:8080/accountholder/" + idStr
	url := sessionData.baseURL + "/accountholder/" + idStr

	var jsonStr = []byte(`{"name":"string_update",
"age":999999,
"weight":8888.88}`)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sessionData.jwtToken)

	if sessionData.log {
		fmt.Println("POST request Headers:", req.Header)
	}

	resp, err := sessionData.client.Do(req)
	if err != nil {
		t.Errorf("Test was unable to PUT /accountholder/{:id}. Got %s.\n", err.Error())
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("PUT /accountholder{:id} expected http status code of 201 - got %d", resp.StatusCode)
	}
	defer resp.Body.Close()
}

// TestDeleteAccountHolder attempts to delete the new AccountHolder on the db
//
// DELETE /accountholder/{:id}
func TestDeleteAccountHolder(t *testing.T) {

	idStr := fmt.Sprint(sessionData.ID)
	// url := "https://localhost:8080/accountholder/" + idStr
	url := sessionData.baseURL + "/accountholder/" + idStr
	var jsonStr = []byte(`{}`)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sessionData.jwtToken)

	if sessionData.log {
		fmt.Println("sessionData.ID:", string(sessionData.ID))
		fmt.Println("DELETE URL:", url)
		fmt.Println("DELETE request Headers:", req.Header)
	}

	resp, err := sessionData.client.Do(req)
	if err != nil {
		t.Errorf("Test was unable to DELETE /accountholder/%d. Got %s.\n", sessionData.ID, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("DELETE /accountholder{:id} expected http status code of 201 - got %d", resp.StatusCode)
	}
}

func TestGetAccountHoldersByName(t *testing.T) {

	// http://127.0.0.1:<port>/accountholders/name(OP '<sel_string>')
	sessionData.testEndPoint = "/accountholders/name(EQ 'test_string')"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

	sessionData.testEndPoint = "/accountholders/name(LIKE 'test_string')"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

} // end func {Name string name  false  false false nonUnique EQ,LIKE gorm:"index;index:idx_age_name"   }

func TestGetAccountHoldersByAge(t *testing.T) {

	// http://127.0.0.1:<port>/accountholders/age(OP XXX)
	sessionData.testEndPoint = "/accountholders/age(EQ 77)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

	sessionData.testEndPoint = "/accountholders/age(LT 77)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

	sessionData.testEndPoint = "/accountholders/age(GT 77)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

} // end func {Age uint age  false  false false  EQ,LT,GT gorm:"index:idx_valid_license_age;index:idx_age_name"   }

func TestGetAccountHoldersByWeight(t *testing.T) {

	// http://127.0.0.1:<port>/accountholders/weight(OP xxx.yyy)
	sessionData.testEndPoint = "/accountholders/weight(EQ 55.44)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

	sessionData.testEndPoint = "/accountholders/weight(LT 55.44)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

	sessionData.testEndPoint = "/accountholders/weight(LE 55.44)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

	sessionData.testEndPoint = "/accountholders/weight(GT 55.44)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

	sessionData.testEndPoint = "/accountholders/weight(GE 55.44)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

} // end func {Weight float64 weight  false  false false  EQ,LT,LE,GT,GE    }

func TestGetAccountHoldersByValidLicense(t *testing.T) {

	// http://127.0.0.1:<port>/accountholders/valid_license(OP true|false)
	sessionData.testEndPoint = "/accountholders/valid_license(EQ true)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

	sessionData.testEndPoint = "/accountholders/valid_license(EQ false)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

	sessionData.testEndPoint = "/accountholders/valid_license(NE true)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

	sessionData.testEndPoint = "/accountholders/valid_license(NE false)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

} // end func {ValidLicense bool valid_license  false  false false nonUnique EQ,NE gorm:"index;index:idx_valid_license_age"   }

// TestGetAccounts attempts to read all accounts from the db
//
// GET /accounts
func TestGetAccounts(t *testing.T) {

	// url := "https://localhost:8080/accounts"
	url := sessionData.baseURL + "/accounts"
	jsonStr := []byte(`{}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sessionData.jwtToken)

	if sessionData.log {
		fmt.Println("GET /accounts request Headers:", req.Header)
	}

	// client := &http.Client{}
	resp, err := sessionData.client.Do(req)
	if err != nil {
		t.Errorf("Test was unable to GET /accounts. Got %s.\n", err.Error())
	}
	defer resp.Body.Close()

	if sessionData.log {
		fmt.Println("GET /accounts response Status:", resp.Status)
		fmt.Println("GET /accounts response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("GET /accounts response Body:", string(body))
	}
}

// TestCreateAccount attempts to create a new Account on the db
//
// POST /account
func TestCreateAccount(t *testing.T) {

	// url := "https://localhost:8080/account"
	url := sessionData.baseURL + "/account"

	var jsonStr = []byte(`{"accountholderid":500000,
"bankname":"string_value",
"banktransit":500000,
"street":"string_value",
"streetnumber":"string_value",
"postcode":"string_value",
"accounttype":"string_value",
"accountnumber":500000,
"accountbalance":1900.99}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sessionData.jwtToken)

	if sessionData.log {
		fmt.Println("POST request Headers:", req.Header)
	}

	resp, err := sessionData.client.Do(req)
	if err != nil {
		t.Errorf("Test was unable to POST /account. Got %s.\n", err.Error())
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("GET /account response Body:", string(body))
		t.Errorf("Test was unable to POST /account. Got %s.\n", err.Error())
	}
	defer resp.Body.Close()

	var e models.Account
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&e); err != nil {
		t.Errorf("Test was unable to decode the result of POST /product. Got %s.\n", err.Error())
	}

	//============================================================================================
	// TODO: implement validation of the returned entity here
	//============================================================================================
	if e.AccountHolderID != 500000 ||
		e.BankName != "string_value" ||
		e.BankTransit != 500000 ||
		e.Street != "string_value" ||
		e.StreetNumber != "string_value" ||
		e.PostCode != "string_value" ||
		e.AccountType != "string_value" ||
		e.AccountNumber != 500000 ||
		e.AccountBalance != 1900.99 {
		t.Errorf("inconsistency detected in POST /account.")
	} else {
		sessionData.ID = e.ID
	}
}

// TestGetAccount attempts to read account/{:id} from the db
// using the id created in this entity's TestCreate function.
//
// GET /account/{:id}
func TestGetAccount(t *testing.T) {

	idStr := fmt.Sprint(sessionData.ID)
	// url := "https://localhost:8080/account/" + idStr
	url := sessionData.baseURL + "/account/" + idStr
	jsonStr := []byte(`{}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sessionData.jwtToken)

	if sessionData.log {
		fmt.Println("GET /account request Headers:", req.Header)
	}

	// client := &http.Client{}
	resp, err := sessionData.client.Do(req)
	if err != nil {
		t.Errorf("Test was unable to GET /account/%d. Got %s.\n", sessionData.ID, err.Error())
	}
	defer resp.Body.Close()

	if sessionData.log {
		fmt.Println("GET /account response Status:", resp.Status)
		fmt.Println("GET /account response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("GET /account response Body:", string(body))
	}
}

// TestUpdateAccount attempts to update an existing Account on the db
//
// PUT /account/{:id}
func TestUpdateAccount(t *testing.T) {

	idStr := fmt.Sprint(sessionData.ID)
	// url := "https://localhost:8080/account/" + idStr
	url := sessionData.baseURL + "/account/" + idStr

	var jsonStr = []byte(`{"accountholderid":999999,
"bankname":"string_update",
"banktransit":999999,
"street":"string_update",
"streetnumber":"string_update",
"postcode":"string_update",
"accounttype":"string_update",
"accountnumber":999999,
"accountbalance":8888.88}`)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sessionData.jwtToken)

	if sessionData.log {
		fmt.Println("POST request Headers:", req.Header)
	}

	resp, err := sessionData.client.Do(req)
	if err != nil {
		t.Errorf("Test was unable to PUT /account/{:id}. Got %s.\n", err.Error())
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("PUT /account{:id} expected http status code of 201 - got %d", resp.StatusCode)
	}
	defer resp.Body.Close()
}

// TestDeleteAccount attempts to delete the new Account on the db
//
// DELETE /account/{:id}
func TestDeleteAccount(t *testing.T) {

	idStr := fmt.Sprint(sessionData.ID)
	// url := "https://localhost:8080/account/" + idStr
	url := sessionData.baseURL + "/account/" + idStr
	var jsonStr = []byte(`{}`)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sessionData.jwtToken)

	if sessionData.log {
		fmt.Println("sessionData.ID:", string(sessionData.ID))
		fmt.Println("DELETE URL:", url)
		fmt.Println("DELETE request Headers:", req.Header)
	}

	resp, err := sessionData.client.Do(req)
	if err != nil {
		t.Errorf("Test was unable to DELETE /account/%d. Got %s.\n", sessionData.ID, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("DELETE /account{:id} expected http status code of 201 - got %d", resp.StatusCode)
	}
}

func TestGetAccountsByAccountHolderID(t *testing.T) {

	// http://127.0.0.1:<port>/accounts/account_holder_id(OP XXX)
	sessionData.testEndPoint = "/accounts/account_holder_id(EQ 77)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

} // end func {AccountHolderID uint account_holder_id  false  false false nonUnique EQ gorm:"index"   }

func TestGetAccountsByBankName(t *testing.T) {

	// http://127.0.0.1:<port>/accounts/bank_name(OP '<sel_string>')
	sessionData.testEndPoint = "/accounts/bank_name(EQ 'test_string')"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

	sessionData.testEndPoint = "/accounts/bank_name(LIKE 'test_string')"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

} // end func {BankName string bank_name  false  true false nonUnique EQ,LIKE gorm:"not null;index"   }

func TestGetAccountsByBankTransit(t *testing.T) {

	// http://127.0.0.1:<port>/accounts/bank_transit(OP XXX)
	sessionData.testEndPoint = "/accounts/bank_transit(EQ 77)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

} // end func {BankTransit uint bank_transit  false  true false nonUnique EQ gorm:"not null;index"   }

func TestGetAccountsByStreet(t *testing.T) {

	// http://127.0.0.1:<port>/accounts/street(OP '<sel_string>')
	sessionData.testEndPoint = "/accounts/street(EQ 'test_string')"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

	sessionData.testEndPoint = "/accounts/street(LIKE 'test_string')"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

} // end func {Street string street  false  false false nonUnique EQ,LIKE gorm:"index"   }

func TestGetAccountsByPostCode(t *testing.T) {

	// http://127.0.0.1:<port>/accounts/post_code(OP '<sel_string>')
	sessionData.testEndPoint = "/accounts/post_code(EQ 'test_string')"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

	sessionData.testEndPoint = "/accounts/post_code(LIKE 'test_string')"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

} // end func {PostCode string post_code  false  false false nonUnique EQ,LIKE gorm:"index"   }

func TestGetAccountsByAccountType(t *testing.T) {

	// http://127.0.0.1:<port>/accounts/account_type(OP '<sel_string>')
	sessionData.testEndPoint = "/accounts/account_type(EQ 'test_string')"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

	sessionData.testEndPoint = "/accounts/account_type(LIKE 'test_string')"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

} // end func {AccountType string account_type  false  true false nonUnique EQ,LIKE gorm:"not null;index"   }

func TestGetAccountsByAccountNumber(t *testing.T) {

	// http://127.0.0.1:<port>/accounts/account_number(OP XXX)
	sessionData.testEndPoint = "/accounts/account_number(EQ 77)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

} // end func {AccountNumber uint account_number  false  true false nonUnique EQ gorm:"not null;index"   }

func TestGetAccountsByAccountBalance(t *testing.T) {

	// http://127.0.0.1:<port>/accounts/account_balance(OP xxx.yyy)
	sessionData.testEndPoint = "/accounts/account_balance(EQ 55.44)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

	sessionData.testEndPoint = "/accounts/account_balance(LT 55.44)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

	sessionData.testEndPoint = "/accounts/account_balance(LE 55.44)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

	sessionData.testEndPoint = "/accounts/account_balance(GT 55.44)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

	sessionData.testEndPoint = "/accounts/account_balance(GE 55.44)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

	sessionData.testEndPoint = "/accounts/account_balance(NE 55.44)"
	sessionData.testURL = sessionData.baseURL + sessionData.testEndPoint
	sessionData.testSelectableField(t)

} // end func {AccountBalance float64 account_balance  false  false false  EQ,LT,LE,GT,GE,NE    }

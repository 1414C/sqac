package sqac_test

import (
	"os"
	"testing"

	"github.com/1414C/sqac/common"
)

// TestForeignKeyCreate
//
// Test ForeignKeyCreate
func TestForeignKeyCreate(t *testing.T) {

	type Warehouse struct {
		ID       uint64 `db:"id" json:"id" sqac:"primary_key:inc;start:40000000"`
		City     string `db:"city" json:"city" sqac:"nullable:false;default:Calgary"`
		Quadrant string `db:"quadrant" json:"quadrant" sqac:"nullable:false;default:SE"`
	}

	type Product struct {
		ID          uint64 `db:"id" json:"id" sqac:"primary_key:inc;start:95000000"`
		ProductName string `db:"product_name" json:"product_name" sqac:"nullable:false;default:unknown"`
		ProductCode string `db:"product_code" json:"product_code" sqac:"nullable:false;default:0000-0000-00"`
		UOM         string `db:"uom" json:"uom" sqac:"nullable:false;default:EA"`
		WarehouseID uint64 `db:"warehouse_id" json:"warehouse_id" sqac:"nullable:false"`
	}

	// determine the table names as per the table creation logic
	wn := common.GetTableName(Warehouse{})
	pn := common.GetTableName(Product{})

	// verify tables do not exist
	err := Handle.DropTables(Product{})
	if err != nil {
		t.Errorf("failed to drop table %s", pn)
	}

	err = Handle.DropTables(Warehouse{})
	if err != nil {
		t.Errorf("failed to drop table %s", wn)
	}

	// create warehouse table
	err = Handle.CreateTables(Warehouse{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table warehouse exists
	if !Handle.ExistsTable(wn) {
		t.Errorf("table %s does not exist", wn)
	}

	// create a new record via the CRUD Create call
	var warehouse = Warehouse{
		City:     "Calgary",
		Quadrant: "SW",
	}

	err = Handle.Create(&warehouse)
	if err != nil {
		t.Errorf(err.Error())
	}

	// create product table
	err = Handle.CreateTables(Product{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table product exists
	if !Handle.ExistsTable(pn) {
		t.Errorf("table %s does not exist", pn)
	}

	// add a foreign-key to table product
	switch Handle.GetDBDriverName() {
	case "sqlite3":
		err = Handle.CreateForeignKey(Product{}, pn, wn, "warehouse_id", "id")
	default:
		err = Handle.CreateForeignKey(nil, pn, wn, "warehouse_id", "id")
	}
	if err != nil {
		t.Errorf("failed to create foreign-key; got: %s", err)
	}

	// try to add a product record with an illegal warehouse_id
	var prod = Product{
		ProductName: "Bad Product",
		ProductCode: "1111-1111-99",
		UOM:         "EA",
		WarehouseID: 55, // does not exist - should fail
	}

	err = Handle.Create(&prod)
	if err == nil {
		t.Errorf("product record %v was created in violation of warehouse foreign_key %v", prod.ID, prod.WarehouseID)
	}

	// try to add a product record with a good warehouse_id
	prod = Product{
		ProductName: "Good Product",
		ProductCode: "5555-5555-11",
		UOM:         "EA",
		WarehouseID: 40000000, // good warehouse id
	}

	err = Handle.Create(&prod)
	if err != nil {
		t.Errorf("product record failed to create with warehouse foreign_key %v", prod.WarehouseID)
	}

	err = Handle.DropTables(Product{})
	if err != nil {
		t.Errorf("failed to drop table %s", pn)
	}

	err = Handle.DropTables(Warehouse{})
	if err != nil {
		t.Errorf("failed to drop table %s", wn)
	}
}

// TestForeignKeyDrop
//
// Test ForeignKeyDrop
func TestForeignKeyDrop(t *testing.T) {

	type Warehouse struct {
		ID       uint64 `db:"id" json:"id" sqac:"primary_key:inc;start:40000000"`
		City     string `db:"city" json:"city" sqac:"nullable:false;default:Calgary"`
		Quadrant string `db:"quadrant" json:"quadrant" sqac:"nullable:false;default:SE"`
	}

	type Product struct {
		ID          uint64 `db:"id" json:"id" sqac:"primary_key:inc;start:95000000"`
		ProductName string `db:"product_name" json:"product_name" sqac:"nullable:false;default:unknown"`
		ProductCode string `db:"product_code" json:"product_code" sqac:"nullable:false;default:0000-0000-00"`
		UOM         string `db:"uom" json:"uom" sqac:"nullable:false;default:EA"`
		WarehouseID uint64 `db:"warehouse_id" json:"warehouse_id" sqac:"nullable:false"`
	}

	// determine the table names as per the table creation logic
	wn := common.GetTableName(Warehouse{})
	pn := common.GetTableName(Product{})

	// verify tables do not exist
	err := Handle.DropTables(Product{})
	if err != nil {
		t.Errorf("failed to drop table %s", pn)
	}

	err = Handle.DropTables(Warehouse{})
	if err != nil {
		t.Errorf("failed to drop table %s", wn)
	}

	// create warehouse table
	err = Handle.CreateTables(Warehouse{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table warehouse exists
	if !Handle.ExistsTable(wn) {
		t.Errorf("table %s does not exist", wn)
	}

	// create a new record via the CRUD Create call
	var warehouse = Warehouse{
		City:     "Calgary",
		Quadrant: "SW",
	}

	err = Handle.Create(&warehouse)
	if err != nil {
		t.Errorf(err.Error())
	}

	// create product table
	err = Handle.CreateTables(Product{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table product exists
	if !Handle.ExistsTable(pn) {
		t.Errorf("table %s does not exist", pn)
	}

	// add a foreign-key to table product
	switch Handle.GetDBDriverName() {
	case "sqlite3":
		err = Handle.CreateForeignKey(Product{}, pn, wn, "warehouse_id", "id")
	default:
		err = Handle.CreateForeignKey(nil, pn, wn, "warehouse_id", "id")
	}
	if err != nil {
		t.Errorf("failed to create foreign-key; got: %s", err)
	}

	// try to add a product record with an illegal warehouse_id
	var prod = Product{
		ProductName: "Bad Product",
		ProductCode: "1111-1111-99",
		UOM:         "EA",
		WarehouseID: 55, // does not exist - should fail
	}

	err = Handle.Create(&prod)
	if err == nil {
		t.Errorf("product record %v was created in violation of warehouse foreign_key %v", prod.ID, prod.WarehouseID)
	}

	// try to add a product record with a good warehouse_id
	prod = Product{
		ProductName: "Good Product",
		ProductCode: "5555-5555-11",
		UOM:         "EA",
		WarehouseID: 40000000, // good warehouse id
	}

	err = Handle.Create(&prod)
	if err != nil {
		t.Errorf("product record failed to create with warehouse foreign_key %v", prod.WarehouseID)
	}

	// attempt to drop a foreign-key that does not exist
	err = Handle.DropForeignKey(Product{}, pn, "fk_fake_fk")
	if err == nil {
		t.Errorf("DropForeignKey erroneously indicated success in dropping non-existent fk: 'fk_fake_fk' on table %s", pn)
	}

	os.Exit(0)

	// drop the foreign-key for real
	err = Handle.DropForeignKey(Product{}, pn, "fk_product_warehouse_id")
	if err != nil {
		t.Errorf("DropForeignKey failed for table %s foreign-key %s - got: %s", pn, "fk_product_warehouse_id", err)
	}

	err = Handle.DropTables(Product{})
	if err != nil {
		t.Errorf("failed to drop table %s", pn)
	}

	err = Handle.DropTables(Warehouse{})
	if err != nil {
		t.Errorf("failed to drop table %s", wn)
	}
}

// TestGetForeignKeyName
//
// Test GetForeignKeyName
func TestGetForeignKeyName(t *testing.T) {

	type Product struct {
		ID          uint64 `db:"id" json:"id" sqac:"primary_key:inc;start:95000000"`
		ProductName string `db:"product_name" json:"product_name" sqac:"nullable:false;default:unknown"`
		ProductCode string `db:"product_code" json:"product_code" sqac:"nullable:false;default:0000-0000-00"`
		UOM         string `db:"uom" json:"uom" sqac:"nullable:false;default:EA"`
		WarehouseID uint64 `db:"warehouse_id" json:"warehouse_id" sqac:"nullable:false"`
	}

	// construct foreign-key name (expect "fk_product_warehouse_id")
	fkn, err := common.GetFKeyName(nil, "product", "warehouse", "warehouse_id", "id")
	if err != nil {
		t.Errorf("failed to construct foreign-key name using '(nil, \"product\", \"warehouse\", \"warehouse_id\", \"id\")', got: %v", err.Error())
	}
	if fkn != "fk_product_warehouse_id" {
		t.Errorf("incorrect foreign-key name determined from '(nil, \"product\", \"warehouse\", \"warehouse_id\", \"id\")', got: %v", fkn)
	}

	// construct foreign-key name (expect "fk_product_warehouse_id")
	fkn, err = common.GetFKeyName(Product{}, "", "warehouse", "warehouse_id", "id")
	if err != nil {
		t.Errorf("failed to construct foreign-key name using '(Product{}, \"\", \"warehouse\", \"warehouse_id\", \"id\")', got: %v", err.Error())
	}
	if fkn != "fk_product_warehouse_id" {
		t.Errorf("incorrect foreign-key name determined from '(Product{}, \"\", \"warehouse\", \"warehouse_id\", \"id\")', got: %v", fkn)
	}

	// construct foreign-key name (expect "fk_product_warehouse_id")
	fkn, err = common.GetFKeyName(Product{}, "product", "warehouse", "warehouse_id", "id")
	if err != nil {
		t.Errorf("failed to construct foreign-key name using '(Product{}, \"product\", \"warehouse\", \"warehouse_id\", \"id\")', got: %v", err.Error())
	}
	if fkn != "fk_product_warehouse_id" {
		t.Errorf("incorrect foreign-key name determined from '(Product{}, \"product\", \"warehouse\", \"warehouse_id\", \"id\")', got: %v", fkn)
	}
}

// TestExistsForeignKeyByName
//
// Test ExistsForeignKeyByName
func TestExistsForeignKeyByName(t *testing.T) {

	type Warehouse struct {
		ID       uint64 `db:"id" json:"id" sqac:"primary_key:inc;start:40000000"`
		City     string `db:"city" json:"city" sqac:"nullable:false;default:Calgary"`
		Quadrant string `db:"quadrant" json:"quadrant" sqac:"nullable:false;default:SE"`
	}

	type Product struct {
		ID          uint64 `db:"id" json:"id" sqac:"primary_key:inc;start:95000000"`
		ProductName string `db:"product_name" json:"product_name" sqac:"nullable:false;default:unknown"`
		ProductCode string `db:"product_code" json:"product_code" sqac:"nullable:false;default:0000-0000-00"`
		UOM         string `db:"uom" json:"uom" sqac:"nullable:false;default:EA"`
		WarehouseID uint64 `db:"warehouse_id" json:"warehouse_id" sqac:"nullable:false"`
	}

	// determine the table names as per the table creation logic
	wn := common.GetTableName(Warehouse{})
	pn := common.GetTableName(Product{})

	// verify tables do not exist
	err := Handle.DropTables(Product{})
	if err != nil {
		t.Errorf("failed to drop table %s", pn)
	}

	err = Handle.DropTables(Warehouse{})
	if err != nil {
		t.Errorf("failed to drop table %s", wn)
	}

	// create warehouse table
	err = Handle.CreateTables(Warehouse{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table warehouse exists
	if !Handle.ExistsTable(wn) {
		t.Errorf("table %s does not exist", wn)
	}

	// create a new record via the CRUD Create call
	var warehouse = Warehouse{
		City:     "Calgary",
		Quadrant: "SW",
	}

	err = Handle.Create(&warehouse)
	if err != nil {
		t.Errorf(err.Error())
	}

	// create product table
	err = Handle.CreateTables(Product{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table product exists
	if !Handle.ExistsTable(pn) {
		t.Errorf("table %s does not exist", pn)
	}

	// add a foreign-key to table product
	switch Handle.GetDBDriverName() {
	case "sqlite3":
		err = Handle.CreateForeignKey(Product{}, pn, wn, "warehouse_id", "id")
	default:
		err = Handle.CreateForeignKey(nil, pn, wn, "warehouse_id", "id")
	}
	if err != nil {
		t.Errorf("failed to create foreign-key; got: %s", err)
	}

	// construct foreign-key name (expect "fk_product_warehouse_id")
	fkn, err := common.GetFKeyName(Product{}, "product", "warehouse", "warehouse_id", "id")
	if err != nil {
		t.Errorf("failed to construct foreign-key name, got: %v", err.Error())
		return
	}

	// check that the foreign-key exists by name
	kExists, err := Handle.ExistsForeignKeyByName(Product{}, fkn)
	if err != nil {
		t.Errorf(err.Error())
	}

	if !kExists {
		t.Errorf("foreign-key '%s' failed to be created via the model", fkn)
	}

}

// TestExistsForeignKeyByFields
//
// Test ExistsForeignKeyByFields
func TestExistsForeignKeyByFields(t *testing.T) {

	type Warehouse struct {
		ID       uint64 `db:"id" json:"id" sqac:"primary_key:inc;start:40000000"`
		City     string `db:"city" json:"city" sqac:"nullable:false;default:Calgary"`
		Quadrant string `db:"quadrant" json:"quadrant" sqac:"nullable:false;default:SE"`
	}

	type Product struct {
		ID          uint64 `db:"id" json:"id" sqac:"primary_key:inc;start:95000000"`
		ProductName string `db:"product_name" json:"product_name" sqac:"nullable:false;default:unknown"`
		ProductCode string `db:"product_code" json:"product_code" sqac:"nullable:false;default:0000-0000-00"`
		UOM         string `db:"uom" json:"uom" sqac:"nullable:false;default:EA"`
		WarehouseID uint64 `db:"warehouse_id" json:"warehouse_id" sqac:"nullable:false"`
	}

	// determine the table names as per the table creation logic
	wn := common.GetTableName(Warehouse{})
	pn := common.GetTableName(Product{})

	// verify tables do not exist
	err := Handle.DropTables(Product{})
	if err != nil {
		t.Errorf("failed to drop table %s", pn)
	}

	err = Handle.DropTables(Warehouse{})
	if err != nil {
		t.Errorf("failed to drop table %s", wn)
	}

	// create warehouse table
	err = Handle.CreateTables(Warehouse{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table warehouse exists
	if !Handle.ExistsTable(wn) {
		t.Errorf("table %s does not exist", wn)
	}

	// create a new record via the CRUD Create call
	var warehouse = Warehouse{
		City:     "Calgary",
		Quadrant: "SW",
	}

	err = Handle.Create(&warehouse)
	if err != nil {
		t.Errorf(err.Error())
	}

	// create product table
	err = Handle.CreateTables(Product{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table product exists
	if !Handle.ExistsTable(pn) {
		t.Errorf("table %s does not exist", pn)
	}

	// add a foreign-key to table product
	switch Handle.GetDBDriverName() {
	case "sqlite3":
		err = Handle.CreateForeignKey(Product{}, pn, wn, "warehouse_id", "id")
	default:
		err = Handle.CreateForeignKey(nil, pn, wn, "warehouse_id", "id")
	}
	if err != nil {
		t.Errorf("failed to create foreign-key; got: %s", err)
	}

	// check that the foreign-key exists by name
	kExists, err := Handle.ExistsForeignKeyByFields(Product{}, "product", "warehouse", "warehouse_id", "id")
	if err != nil {
		t.Errorf(err.Error())
	}

	if !kExists {
		t.Errorf("foreign-key '%s' failed to be created via the model", "fk_product_warehouse_id")
	}

}

// TestForeignKeyCreateFromModel
//
// Test ForeignKeyCreateFromModel
func TestForeignKeyCreateFromModel(t *testing.T) {

	type Warehouse struct {
		ID       uint64 `db:"id" json:"id" sqac:"primary_key:inc;start:40000000"`
		City     string `db:"city" json:"city" sqac:"nullable:false;default:Calgary"`
		Quadrant string `db:"quadrant" json:"quadrant" sqac:"nullable:false;default:SE"`
	}

	type Product struct {
		ID          uint64 `db:"id" json:"id" sqac:"primary_key:inc;start:95000000"`
		ProductName string `db:"product_name" json:"product_name" sqac:"nullable:false;default:unknown"`
		ProductCode string `db:"product_code" json:"product_code" sqac:"nullable:false;default:0000-0000-00"`
		UOM         string `db:"uom" json:"uom" sqac:"nullable:false;default:EA"`
		WarehouseID uint64 `db:"warehouse_id" json:"warehouse_id" sqac:"nullable:false;fkey:warehouse(id)"`
	}

	// determine the table names as per the table creation logic
	wn := common.GetTableName(Warehouse{})
	pn := common.GetTableName(Product{})

	// verify tables do not exist
	err := Handle.DropTables(Product{})
	if err != nil {
		t.Errorf("failed to drop table %s", pn)
	}

	err = Handle.DropTables(Warehouse{})
	if err != nil {
		t.Errorf("failed to drop table %s", wn)
	}

	// create warehouse table
	err = Handle.CreateTables(Warehouse{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table warehouse exists
	if !Handle.ExistsTable(wn) {
		t.Errorf("table %s does not exist", wn)
	}

	// create product table with its foreign-key definition
	err = Handle.CreateTables(Product{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table product exists
	if !Handle.ExistsTable(pn) {
		t.Errorf("table %s does not exist", wn)
	}

	// check that the foreign-key exists
	kExists, err := Handle.ExistsForeignKeyByName(Product{}, "fk_product_warehouse_id")
	if err != nil {
		t.Errorf(err.Error())
	}

	if !kExists {
		t.Errorf("foreign-key 'fk_product_warehouse_id' failed to be created via the model")
	}

	// check that the foreign-key exists via fields
	kExists, err = Handle.ExistsForeignKeyByFields(Product{}, "product", "warehouse", "warehouse_id", "id")
	if err != nil {
		t.Errorf(err.Error())
	}

	if !kExists {
		t.Errorf("foreign-key 'fk_product_warehouse_id' failed to be created via the model")
	}

	err = Handle.DropTables(Product{})
	if err != nil {
		t.Errorf("failed to drop table %s", pn)
	}

	err = Handle.DropTables(Warehouse{})
	if err != nil {
		t.Errorf("failed to drop table %s", wn)
	}
}

// TestForeignKeyCreateTwoFromModel
//
// Test ForeignKeyCreateTwoFromModel
func TestForeignKeyCreateTwoFromModel(t *testing.T) {

	type Warehouse struct {
		ID       uint64 `db:"id" json:"id" sqac:"primary_key:inc;start:40000000"`
		City     string `db:"city" json:"city" sqac:"nullable:false;default:Calgary"`
		Quadrant string `db:"quadrant" json:"quadrant" sqac:"nullable:false;default:SE"`
	}

	type UnitOfMeasure struct {
		Uom     string `db:"uom" json:"uom" sqac:"primary_key:"`
		UomText string `db:"uom_text" json:"uom_text" sqac:"nullable:false"`
	}

	type Product struct {
		ID          uint64 `db:"id" json:"id" sqac:"primary_key:inc;start:95000000"`
		ProductName string `db:"product_name" json:"product_name" sqac:"nullable:false;default:unknown"`
		ProductCode string `db:"product_code" json:"product_code" sqac:"nullable:false;default:0000-0000-00"`
		UOM         string `db:"uom" json:"uom" sqac:"nullable:false;default:EA;fkey:unitofmeasure(uom)"`
		WarehouseID uint64 `db:"warehouse_id" json:"warehouse_id" sqac:"nullable:false;fkey:warehouse(id)"`
	}

	// determine the table names as per the table creation logic
	wn := common.GetTableName(Warehouse{})
	un := common.GetTableName(UnitOfMeasure{})
	pn := common.GetTableName(Product{})

	// verify tables do not exist
	err := Handle.DropTables(Product{})
	if err != nil {
		t.Errorf("failed to drop table %s", pn)
	}

	err = Handle.DropTables(Warehouse{})
	if err != nil {
		t.Errorf("failed to drop table %s", wn)
	}

	err = Handle.DropTables(UnitOfMeasure{})
	if err != nil {
		t.Errorf("failed to drop table %s", wn)
	}

	// create warehouse table
	err = Handle.CreateTables(Warehouse{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table warehouse exists
	if !Handle.ExistsTable(wn) {
		t.Errorf("table %s does not exist", wn)
	}

	// create unitofmeasure table
	err = Handle.CreateTables(UnitOfMeasure{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table unitofmeasre exists
	if !Handle.ExistsTable(un) {
		t.Errorf("table %s does not exist", wn)
	}

	// create product table with its foreign-key definitions
	err = Handle.CreateTables(Product{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table product exists
	if !Handle.ExistsTable(pn) {
		t.Errorf("table %s does not exist", wn)
	}

	// check that the warehouse(id) foreign-key exists
	kExists, err := Handle.ExistsForeignKeyByName(Product{}, "fk_product_warehouse_id")
	if err != nil {
		t.Errorf(err.Error())
	}

	if !kExists {
		t.Errorf("foreign-key 'fk_product_warehouse_id' failed to be created via the model")
	}

	// check that the warehouse(id) foreign-key exists via fields
	kExists, err = Handle.ExistsForeignKeyByFields(Product{}, "product", "warehouse", "warehouse_id", "id")
	if err != nil {
		t.Errorf(err.Error())
	}

	if !kExists {
		t.Errorf("foreign-key 'fk_product_warehouse_id' failed to be created via the model")
	}

	// check that the unitofmeasure(uom) foreign-key exists
	kExists, err = Handle.ExistsForeignKeyByName(Product{}, "fk_product_unitofmeasure_uom")
	if err != nil {
		t.Errorf(err.Error())
	}

	if !kExists {
		t.Errorf("foreign-key 'fk_product_unitofmeasure_uom' failed to be created via the model")
	}

	// check that the unitofmeasure(uom) foreign-key exists via fields
	kExists, err = Handle.ExistsForeignKeyByFields(Product{}, "product", "unitofmeasure", "uom", "uom")
	if err != nil {
		t.Errorf(err.Error())
	}

	if !kExists {
		t.Errorf("foreign-key 'fk_product_unitofmeasure_uom' failed to be created via the model")
	}

	err = Handle.DropTables(Product{})
	if err != nil {
		t.Errorf("failed to drop table %s", pn)
	}

	err = Handle.DropTables(Warehouse{})
	if err != nil {
		t.Errorf("failed to drop table %s", wn)
	}

	err = Handle.DropTables(UnitOfMeasure{})
	if err != nil {
		t.Errorf("failed to drop table %s", un)
	}
}

// TestForeignKeyCreateTwoDelOneFromModel
//
// Test ForeignKeyCreateTwoDelOneFromModel
func TestForeignKeyCreateTwoDelOneFromModel(t *testing.T) {

	type Warehouse struct {
		ID       uint64 `db:"id" json:"id" sqac:"primary_key:inc;start:40000000"`
		City     string `db:"city" json:"city" sqac:"nullable:false;default:Calgary"`
		Quadrant string `db:"quadrant" json:"quadrant" sqac:"nullable:false;default:SE"`
	}

	type UnitOfMeasure struct {
		Uom     string `db:"uom" json:"uom" sqac:"primary_key:"`
		UomText string `db:"uom_text" json:"uom_text" sqac:"nullable:false"`
	}

	type Product struct {
		ID          uint64 `db:"id" json:"id" sqac:"primary_key:inc;start:95000000"`
		ProductName string `db:"product_name" json:"product_name" sqac:"nullable:false;default:unknown"`
		ProductCode string `db:"product_code" json:"product_code" sqac:"nullable:false;default:0000-0000-00"`
		UOM         string `db:"uom" json:"uom" sqac:"nullable:false;default:EA;fkey:unitofmeasure(uom)"`
		WarehouseID uint64 `db:"warehouse_id" json:"warehouse_id" sqac:"nullable:false;fkey:warehouse(id)"`
	}

	// determine the table names as per the table creation logic
	wn := common.GetTableName(Warehouse{})
	un := common.GetTableName(UnitOfMeasure{})
	pn := common.GetTableName(Product{})

	// verify tables do not exist
	err := Handle.DropTables(Product{})
	if err != nil {
		t.Errorf("failed to drop table %s", pn)
	}

	err = Handle.DropTables(Warehouse{})
	if err != nil {
		t.Errorf("failed to drop table %s", wn)
	}

	err = Handle.DropTables(UnitOfMeasure{})
	if err != nil {
		t.Errorf("failed to drop table %s", wn)
	}

	// create warehouse table
	err = Handle.CreateTables(Warehouse{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table warehouse exists
	if !Handle.ExistsTable(wn) {
		t.Errorf("table %s does not exist", wn)
	}

	// create unitofmeasure table
	err = Handle.CreateTables(UnitOfMeasure{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table unitofmeasure exists
	if !Handle.ExistsTable(un) {
		t.Errorf("table %s does not exist", wn)
	}

	// create product table with its foreign-key definitions
	err = Handle.CreateTables(Product{})
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// expect that table product exists
	if !Handle.ExistsTable(pn) {
		t.Errorf("table %s does not exist", wn)
	}

	// check that the warehouse(id) foreign-key exists
	kExists, err := Handle.ExistsForeignKeyByName(Product{}, "fk_product_warehouse_id")
	if err != nil {
		t.Errorf(err.Error())
	}

	if !kExists {
		t.Errorf("foreign-key 'fk_product_warehouse_id' failed to be created via the model")
	}

	// check that the warehouse(id) foreign-key exists via fields
	kExists, err = Handle.ExistsForeignKeyByFields(Product{}, "product", "warehouse", "warehouse_id", "id")
	if err != nil {
		t.Errorf(err.Error())
	}

	if !kExists {
		t.Errorf("foreign-key 'fk_product_warehouse_id' failed to be created via the model")
	}

	// check that the unitofmeasure(uom) foreign-key exists
	kExists, err = Handle.ExistsForeignKeyByName(Product{}, "fk_product_unitofmeasure_uom")
	if err != nil {
		t.Errorf(err.Error())
	}

	if !kExists {
		t.Errorf("foreign-key 'fk_product_unitofmeasure_uom' failed to be created via the model")
	}

	// check that the unitofmeasure(uom) foreign-key exists via fields
	kExists, err = Handle.ExistsForeignKeyByFields(Product{}, "product", "unitofmeasure", "uom", "uom")
	if err != nil {
		t.Errorf(err.Error())
	}

	if !kExists {
		t.Errorf("foreign-key 'fk_product_unitofmeasure_uom' failed to be created via the model")
	}

	// drop "fk_product_unitofmeasure_uom" on table product
	err = Handle.DropForeignKey(Product{}, "product", "fk_product_unitofmeasure_uom")
	if err != nil {
		t.Errorf("failed to drop foreign-key 'fk_product_unitofmeasure_uom', got: %s", err.Error())
	}

	// check that the unitofmeasure(uom) foreign-key has been dropped
	kExists, err = Handle.ExistsForeignKeyByName(Product{}, "fk_product_unitofmeasure_uom")
	if err != nil {
		t.Errorf(err.Error())
	}

	if kExists {
		t.Errorf("foreign-key 'fk_product_unitofmeasure_uom' failed to be dropped via direct call")
	}

	err = Handle.DropTables(Product{})
	if err != nil {
		t.Errorf("failed to drop table %s", pn)
	}

	err = Handle.DropTables(Warehouse{})
	if err != nil {
		t.Errorf("failed to drop table %s", wn)
	}

	err = Handle.DropTables(UnitOfMeasure{})
	if err != nil {
		t.Errorf("failed to drop table %s", un)
	}
}

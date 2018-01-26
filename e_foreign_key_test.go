package sqac_test

import (
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
	err = Handle.DropForeignKey(nil, pn, "fk_fake_fk")
	if err == nil {
		t.Errorf("DropForeignKey erroneously indicated success in dropping non-existent fk: 'fk_fake_fk' on table %s", pn)
	}

	// drop the foreign-key for real
	err = Handle.DropForeignKey(nil, pn, "fk_product_warehouse")
	if err != nil {
		t.Errorf("DropForeignKey failed for table %s foreign-key %s - got: %s", pn, "fk_product_warehouse", err)
	}

	// err = Handle.DropTables(Product{})
	// if err != nil {
	// 	t.Errorf("failed to drop table %s", pn)
	// }

	// err = Handle.DropTables(Warehouse{})
	// if err != nil {
	// 	t.Errorf("failed to drop table %s", wn)
	// }
}

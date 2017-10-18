package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/1414C/sqlxtest/dbgen/common"

	"github.com/1414C/sqlxtest/dbgen"
	"github.com/1414C/sqlxtest/exp"
	"github.com/1414C/sqlxtest/exp2"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	bBasic := flag.Bool("bBasic", false, "execute the basic sqlx example")
	bProfile := flag.Bool("bProfile", false, "execute the profile sqlx example")
	bTableTest := flag.Bool("bTableTest", false, "execute the create table test")

	// override testing
	bOverride := flag.Bool("bOverride", false, "run exp.TestOverrides()")

	// camel_to_snake_testing
	bCamel := flag.Bool("bCamel", false, "test camelCase to snake_case converter")
	sText := flag.String("s", "", "camelCase for conversion to snake_case")

	// table creation testing
	bPGCreate := flag.Bool("bPGCreate", false, "test creation of a table in postgres")

	// table drop testing
	bPGDrop := flag.Bool("bPGDrop", false, "test dropping a table in postgres")
	flag.Parse()

	// this Pings the database trying to connect, panics on error
	// use sqlx.Open() for sql.Open() semantics
	db, err := sqlx.Connect("postgres", "host=127.0.0.1 user=godev dbname=sqlx sslmode=disable password=gogogo123")
	// db, err := sqlx.Connect("sqlite3", "testdb.sqlite")
	if err != nil {
		log.Fatalln(err)
	}

	if !*bCamel {
		if *bBasic {
			exp.Basic(db)
		}

		if *bProfile {
			exp.CreateProfile(db)
		}

		if *bTableTest {
			exp2.CreateTableTest(db)
		}

		if *bOverride {
			exp.TestOverrides()
		}

		if *bPGCreate {
			var pg dbgen.PostgresFlavor

			type Location struct {
				Substation int    `db:"substation" rgen:"nullable:false;default:123"`
				Region     string `db:"region" rgen:"nullable:false;index:non-unique;default:YYC"`
				Province   string `db:"province" rgen:"nullable:false;default:AB"`
				Country    string `db:"country" rgen:"nullable:false;default:Canada"`
				Apples     int    `db:"apples" rgen:"index:idx_apples_oranges"`
				Oranges    int    `db:"oranges" rgen:"index:idx_apples_oranges;index:idx_oranges_peaches"`
				Peaches    int    `db:"peaches" rgen:"index:idx_oranges_peaches"`
				Pears      int    `db:"pears" rgen:"nullable:false;default:0;index:idx_pears_bananas"`
				Bananas    int    `db:"bananas" rgen:"nullable:false;default:4;index:idx_pears_bananas;index:idx_bananas_grapes"`
				Grapes     int64  `db:"grapes" rgen:"nullable:false;default:1000000000;index:idx_bananas_grapes"`
			}
			type Transformer struct {
				DeviceNum   int       `db:"device_num" rgen:"primary_key:inc;start:60000000"`
				NodeNum     int       `db:"node_num" rgen:"primary_key:inc;start:50"`
				MaterialNum int       `db:"material_num" rgen:"primary_key: "`
				CreateDate  time.Time `db:"create_date" rgen:"nullable:false;default:now()"`
				Ratio       int       `db:"ratio" rgen:"nullable:false;default:5"`
				Weight      int       `db:"decimals" rgen:"nullable:false;default:5"`
				Location
			}
			type Depot struct {
				DepotNum   int       `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
				CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
				Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
				Province   string    `db:"province" rgen:"nullable:false;default:AB"`
				Country    string    `db:"country" rgen:"nullable:true;"`
			}
			pg.DB = db
			pg.Log = false
			pg.AlterTables(Transformer{}, Depot{})
			pg.DB.Exec("INSERT INTO depot (depot_num, region, province, country) VALUES (DEFAULT, 'YVR', 'BC', 'CA') RETURNING depot_num;")
			result, err := pg.DB.Exec("INSERT INTO depot (depot_num, region, province) VALUES (DEFAULT, 'YVR','') RETURNING depot_num;")
			if err != nil {
				fmt.Println("insert error:", err)
			}
			ra, err := result.RowsAffected()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("%d rows affected.\n", ra)
			}

			type DepotN struct {
				DepotNum   int            `db:"depot_num"`
				CreateDate time.Time      `db:"create_date"`
				Region     string         `db:"region"`
				Province   string         `db:"province"`
				Country    sql.NullString `db:"country"`
			}

			d := []DepotN{}
			err = pg.DB.Select(&d, "SELECT * FROM depot;")
			if err != nil {
				fmt.Println("err reading Depot:", err)
				fmt.Println(d)
			}
			for _, v := range d {
				dx := Depot{
					DepotNum: v.DepotNum,
					Region:   v.Region,
					Province: v.Province,
					Country:  v.Country.String,
				}

				fmt.Printf("depot num:%d, country:%s, province:%s, region:%s\n", dx.DepotNum, dx.Country, dx.Province, dx.Region)
			}
		}
		if *bPGDrop {
			var pg dbgen.PostgresFlavor

			type Location struct {
				Substation int    `db:"substation" rgen:"nullable:false;default:123"`
				Region     string `db:"region" rgen:"nullable:false;default:YYC"`
				Province   string `db:"province" rgen:"nullable:false;default:AB"`
				Country    string `db:"country" rgen:"nullable:false;default:Canada"`
			}
			type Transformer struct {
				DeviceNum   int64 `db:"device_num" rgen:"primary_key:inc;start:60000000"`
				NodeNum     int   `db:"node_num" rgen:"primary_key:inc;start:50"`
				MaterialNum int   `db:"material_num" rgen:"primary_key: "`
				Ratio       int   `db:"ratio" rgen:"nullable:false;default:5"`
				Weight      int   `db:"decimals" rgen:"nullable:false;default:5"`
				Location
			}
			type Depot struct {
				DepotNum int    `db:"depot_num" rgen:"primary_key:inc;start:90000000"`
				Region   string `db:"region" rgen:"nullable:false;default:YYC"`
				Province string `db:"province" rgen:"nullable:false;default:AB"`
				Country  string `db:"country" rgen:"nullable:false;default:Canada"`
			}
			pg.DB = db
			pg.DropTables(Transformer{}, Depot{})
		}
	}
	if *bCamel {
		common.CamelToSnake(*sText)
	}

}

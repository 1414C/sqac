-- Ensure that we are using the correct database
USE sqlx;  
GO  

-- Create a new table called 'depot' in schema 'dbo'
-- Drop the table if it already exists
IF OBJECT_ID('dbo.depot', 'U') IS NOT NULL
DROP TABLE dbo.depot
GO
-- Create the table in the specified schema
CREATE TABLE dbo.depot
(
  depot_num INT NOT NULL IDENTITY (1,1) PRIMARY KEY, -- primary key column
  create_date DATETIME2 NOT NULL DEFAULT GETDATE(),
  region [NVARCHAR](50) NOT NULL DEFAULT 'YYC',
  province [NVARCHAR](50) NOT NULL DEFAULT 'AB',
  country [NVARCHAR](50) NOT NULL DEFAULT 'CA'
);
GO

-- Reseed the primary key
DBCC CHECKIDENT ('dbo.depot', RESEED, 50000000);  
GO  

	--CreateDate time.Time `db:"create_date" rgen:"nullable:false;default:now();index:unique"`
	--Region     string    `db:"region" rgen:"nullable:false;default:YYC"`
	--Province   string    `db:"province" rgen:"nullable:false;default:AB"`
	--Country    string    `db:"country" rgen:"nullable:false;default:CA"`

-- List columns in all tables whose name is like 'depot'
SELECT 
    TableName = tbl.table_schema + '.' + tbl.table_name, 
    ColumnName = col.column_name, 
    ColumnDataType = col.data_type
FROM information_schema.tables tbl
INNER JOIN information_schema.columns col 
    ON col.table_name = tbl.table_name
    AND col.table_schema = tbl.table_schema

WHERE tbl.table_type = 'base table' and tbl.table_name like '%depot%'
GO

-- Insert rows into table 'depot'
INSERT INTO depot
( -- columns to insert data into
 [region], [province]
)
VALUES
( -- first row: values for the columns in the list above
 'YYZ', 'ON'
),
( -- second row: values for the columns in the list above
 'YUL', 'PQ'
)
-- add more rows here
GO

-- Insert rows into table 'depot'
INSERT INTO depot
( -- columns to insert data into
 [region], [province]
)
VALUES
( -- first row: values for the columns in the list above
 'YVR', 'BC'
)
-- add more rows here
GO

-- Select rows from a Table or View 'depot' in schema 'dbo'
SELECT * FROM dbo.depot;
GO

-- https://docs.microsoft.com/en-us/sql/t-sql/statements/create-index-transact-sql
-- Create a nonclustered index on dbo.depot.region
CREATE INDEX idx_region ON dbo.depot (region); 
GO

-- verify that index idx_region exists
SELECT * FROM sys.indexes 
WHERE name='idx_region' AND object_id = OBJECT_ID('dbo.depot');
GO

-- DROP INDEX idx_region ON dbo.depot;
-- GO
CREATE INDEX idx_province_country on dbo.depot (province, country);
GO

-- verify that index idx_province_country exists
SELECT * FROM sys.indexes 
WHERE name='idx_province_country' AND object_id = OBJECT_ID('dbo.depot');
GO

-- check that column 'country' exists in table dbo.depot
-- List columns in all tables whose name is like 'depot'
SELECT 
    TableName = tbl.table_schema + '.' + tbl.table_name, 
    ColumnName = col.column_name, 
    ColumnDataType = col.data_type
FROM information_schema.tables tbl
INNER JOIN information_schema.columns col 
    ON col.table_name = tbl.table_name
    AND col.table_schema = tbl.table_schema

WHERE tbl.table_type = 'base table' and tbl.table_name like '%depot%' and col.COLUMN_NAME = 'country';
GO

-- Create a new table called 'equipment' in schema 'dbo'
-- Drop the table if it already exists
IF OBJECT_ID('dbo.equipment', 'U') IS NOT NULL
DROP TABLE dbo.equipment
GO
-- Create the table in the specified schema
CREATE TABLE dbo.equipment
(
  equipment_num INT NOT NULL IDENTITY (1,1) PRIMARY KEY, -- primary key column
  valid_from DATETIME2 NOT NULL DEFAULT GETDATE(),
  valid_to DATETIME2 NOT NULL DEFAULT '9999-12-31 23:59:59.999',
  created_at DATETIME2 NOT NULL DEFAULT GETDATE(),
  material_num INT,
  description [NVARCHAR](255) NOT NULL,
  serial_num  [NVARCHAR](255) NOT NULL,
  int_example INT NOT NULL DEFAULT 0,
  int64_example BIGINT NOT NULL DEFAULT 9999999,
  int32_example INT NOT NULL DEFAULT 888888,
  int16_example SMALLINT NOT NULL DEFAULT 777777,
  int8_example TINYINT NOT NULL DEFAULT 255,
  u_int_example INT NOT NULL DEFAULT 0,
  u_int64_example BIGINT NOT NULL DEFAULT 9999999,
  u_int32_example INT NOT NULL DEFAULT 888888,
  u_int16_example SMALLINT NOT NULL DEFAULT 777777,
  u_int8_example TINYINT NOT NULL DEFAULT 127,
  float32_example NUMERIC NOT NULL DEFAULT 22.333,
  float64_example NUMERIC NOT NULL DEFAULT 323523.335,
  bool_example BIT NOT NULL DEFAULT 0,
  bool_null_example BIT,
  rune_example TINYINT,
  byte_example TINYINT,
  region [NVARCHAR](50) NOT NULL DEFAULT 'YYC',
  province [NVARCHAR](50) NOT NULL DEFAULT 'AB',
  country [NVARCHAR](50) NOT NULL DEFAULT 'CA'
);
GO

-- Reseed the primary key
DBCC CHECKIDENT ('dbo.equipment', RESEED, 90000000);  
GO  


	--EquipmentNum   int64     `db:"equipment_num" rgen:"primary_key:inc;start:55550000"`
	--ValidFrom      time.Time `db:"valid_from" rgen:"primary_key;nullable:false;default:now()"`
	--ValidTo        time.Time `db:"valid_to" rgen:"primary_key;nullable:false;default:eot"`
	--CreatedAt      time.Time `db:"created_at" rgen:"nullable:false;default:now()"`
	--InspectionAt   time.Time `db:"inspection_at" rgen:"nullable:true"`
	--MaterialNum    int       `db:"material_num" rgen:"index:idx_material_num_serial_num"`
	--Description    string    `db:"description" rgen:"rgen:nullable:false"`
	--SerialNum      string    `db:"serial_num" rgen:"index:idx_material_num_serial_num"`
	--IntExample     int       `db:"int_example" rgen:"nullable:false;default:0"`
	--Int64Example   int64     `db:"int64_example" rgen:"nullable:false;default:0"`
	--Int32Example   int32     `db:"int32_example" rgen:"nullable:false;default:0"`
	--Int16Example   int16     `db:"int16_example" rgen:"nullable:false;default:0"`
	--Int8Example    int8      `db:"int8_example" rgen:"nullable:false;default:0"`
	--UIntExample    uint      `db:"u_int_example" rgen:"nullable:false;default:0"`
	--UInt64Example  uint64    `db:"u_int64_example" rgen:"nullable:false;default:0"`
	--UInt32Example  uint32    `db:"u_int32_example" rgen:"nullable:false;default:0"`
	--UInt16Example  uint16    `db:"u_int16_example" rgen:"nullable:false;default:0"`
	--UInt8Example   uint8     `db:"u_int8_example" rgen:"nullable:false;default:0"`
	--Float32Example float32   `db:"float32_example" rgen:"nullable:false;default:0.0"`
	--Float64Example float64   `db:"float64_example" rgen:"nullable:false;default:0.0"`
	--BoolExample    bool      `db:"bool_example" rgen:"nullable:false;default:false"`
	--RuneExample    rune      `db:"rune_example" rgen:"nullable:true"`
	--ByteExample    byte      `db:"byte_example" rgen:"nullable:true"`
	--DoNotCreate    string    `db:"do_not_create" rgen:"-"`

-- List columns in all tables whose name is like 'equipment'
SELECT 
    TableName = tbl.table_schema + '.' + tbl.table_name, 
    ColumnName = col.column_name, 
    ColumnDataType = col.data_type
FROM information_schema.tables tbl
INNER JOIN information_schema.columns col 
    ON col.table_name = tbl.table_name
    AND col.table_schema = tbl.table_schema

WHERE tbl.table_type = 'base table' and tbl.table_name like '%equipment%'
GO

-- Create a new table called 'depot' in schema 'dbo'
-- Drop the table if it already exists
IF OBJECT_ID('dbo.depot', 'U') IS NOT NULL
DROP TABLE dbo.depot
GO
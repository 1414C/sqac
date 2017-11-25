-- Ensure that we are using the correct database
USE sqlx;  
GO  

-- Create a new table called 'depot' in schema 'dbo'
-- Drop the table if it already exists
--IF OBJECT_ID('dbo.depotcreate', 'U') IS NOT NULL
--DROP TABLE dbo.depotcreate
--GO

--INSERT INTO depotcreate ( time_col,  int_zero_val_no_default,  new_column2,  new_column3,  depot_bay,  region,  new_column1) VALUES ( '2009-11-17 20:34:58',  0,  9999,  45.330000,  0,  'YYC',  'string_value');
--GO

--UPDATE depot SET  new_column2 = 1111,  new_column3 = 3333.555600,  fld_two = 0,  create_date = '2017-11-13 03:15:24',  region = 'YYZ',  province = 'ON',  new_column1 = 'updated_value',  country = 'CA',  fld_one = 0 WHERE  depot_num = 90000000 AND depot_bay = 0 AND test_key_date = '2017-11-13 03:15:24';
--GO

--DROP TABLE nfloat;

--SELECT * FROM nfloat;
INSERT INTO nbool ( bool_with_value,  null_bool_dflt_with_value,  null_bool_with_value,  null_bool,  bool_dflt_with_value) VALUES ( false,  false,  false,  NULL,  false);
GO
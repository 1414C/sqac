package sqac

// things to deal with:
// rgen:"primary_key:inc;start:55550000"
// rgen:"nullable:false"
// rgen:"default:0"
// rgen:"index:idx_material_num_serial_num
// rgen:"index:unique/non-unique"
// timestamp syntax and functions
// - pg now() equivalent
// - pg make_timestamptz(9999, 12, 31, 23, 59, 59.9) equivalent
// uint8  - TINYINT - range must be smaller than int8?
// uint16 - SMALLINT
// uint32 - INT
// uint64 - BIGINT

// int8  - TINYINT
// int16 - SMALLINT
// int32 - INT
// int64 - BIGINT

// float32 - DOUBLE
// float64 - DOUBLE

// bool - BOOLEAN - (alias for TINYINT(1))

// rune - INT (32-bits - unicode and stuff)
// byte - TINTINT - (8-bits and stuff)

// string - VARCHAR(255) - (uses 1-byte for length-prefix in record prefix)
// string - VARCHAR(256) - (uses 2-bytes for length-prefix; use for strings
//                      	that may exceed 255 bytes->out to max 65,535 bytes)

// TIMESTAMP - also look at YYYYMMDD format, which seems to be native

// autoincrement - https://mariadb.com/kb/en/library/auto_increment/
// spatial - POINT, MULTIPOINT, POLYGON (future)  https://mariadb.com/kb/en/library/geometry-types/

// CREATE TABLE `test_default_four` (
// 	`int16_key` bigint NOT NULL AUTO_INCREMENT,
// 	 `int32_field` int NOT NULL DEFAULT 0,
// 	`description` varchar(255) DEFAULT 'test',
// 	PRIMARY KEY (`int16_key`)
//   ) ENGINE=InnoDB DEFAULT CHARSET=latin1

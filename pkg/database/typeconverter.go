package database

import "strings"

type TypeConverter struct {
	driver   string
	fromJdbc map[string]map[string]string
	toJdbc   map[string]map[string]string
	valid    map[string]bool
}

func NewTypeConverter(driver string) *TypeConverter {
	fromJdbc := map[string]map[string]string{
		"mysql": {
			"clob":      "longtext",
			"boolean":   "tinyint(1)",
			"blob":      "longblob",
			"timestamp": "datetime",
		},
		"pgsql": {
			"clob":      "text",
			"blob":      "bytea",
			"float":     "real",
			"double":    "double precision",
			"varbinary": "bytea",
		},
		"sqlsrv": {
			"boolean":   "bit",
			"varchar":   "nvarchar",
			"clob":      "ntext",
			"blob":      "image",
			"time":      "time(0)",
			"timestamp": "datetime2(0)",
			"double":    "float",
			"float":     "real",
		},
	}

	toJdbc := map[string]map[string]string{
		"simplified": {
			"char":                    "varchar",
			"longvarchar":             "clob",
			"nchar":                   "varchar",
			"nvarchar":                "varchar",
			"longnvarchar":            "clob",
			"binary":                  "varbinary",
			"longvarbinary":           "blob",
			"tinyint":                 "integer",
			"smallint":                "integer",
			"real":                    "float",
			"numeric":                 "decimal",
			"nclob":                   "clob",
			"time_with_timezone":      "time",
			"timestamp_with_timezone": "timestamp",
		},
		"mysql": {
			"tinyint(1)": "boolean",
			"bit(1)":     "boolean",
			"tinyblob":   "blob",
			"mediumblob": "blob",
			"longblob":   "blob",
			"tinytext":   "clob",
			"mediumtext": "clob",
			"longtext":   "clob",
			"text":       "clob",
			"mediumint":  "integer",
			"int":        "integer",
			"polygon":    "geometry",
			"point":      "geometry",
			"datetime":   "timestamp",
			"year":       "integer",
			"enum":       "varchar",
			"set":        "varchar",
			"json":       "clob",
		},
		"pgsql": {
			"bigserial":         "bigint",
			"bit varying":       "bit",
			"box":               "geometry",
			"bytea":             "blob",
			"bpchar":            "char",
			"character varying": "varchar",
			"character":         "char",
			"cidr":              "varchar",
			"circle":            "geometry",
			"double precision":  "double",
			"inet":              "integer",
			//"interval [ fields ]"
			"json":                        "clob",
			"jsonb":                       "clob",
			"line":                        "geometry",
			"lseg":                        "geometry",
			"macaddr":                     "varchar",
			"money":                       "decimal",
			"path":                        "geometry",
			"point":                       "geometry",
			"polygon":                     "geometry",
			"real":                        "float",
			"serial":                      "integer",
			"text":                        "clob",
			"time without time zone":      "time",
			"time with time zone":         "time_with_timezone",
			"timestamp without time zone": "timestamp",
			"timestamp with time zone":    "timestamp_with_timezone",
			//"tsquery"=
			//"tsvector"
			//"txid_snapshot"
			"uuid": "char",
			"xml":  "clob",
		},
		// source: https://docs.microsoft.com/en-us/sql/connect/jdbc/using-basic-data-types?view=sql-server-2017
		"sqlsrv": {
			"varbinary":        "blob",
			"bit":              "boolean",
			"datetime":         "timestamp",
			"datetime2":        "timestamp",
			"float":            "double",
			"image":            "blob",
			"int":              "integer",
			"money":            "decimal",
			"ntext":            "clob",
			"smalldatetime":    "timestamp",
			"smallmoney":       "decimal",
			"text":             "clob",
			"timestamp":        "binary",
			"udt":              "varbinary",
			"uniqueidentifier": "char",
			"xml":              "clob",
		},
		"sqlite": {
			"tinytext":         "clob",
			"text":             "clob",
			"mediumtext":       "clob",
			"longtext":         "clob",
			"mediumint":        "integer",
			"int":              "integer",
			"bigint":           "bigint",
			"int2":             "smallint",
			"int4":             "integer",
			"int8":             "bigint",
			"double precision": "double",
			"datetime":         "timestamp",
		},
	}

	// source: https://docs.oracle.com/javase/9/docs/api/java/sql/Types.html
	valid := map[string]bool{
		//"array" : true,
		"bigint":  true,
		"binary":  true,
		"bit":     true,
		"blob":    true,
		"boolean": true,
		"char":    true,
		"clob":    true,
		//"datalink" : true,
		"date":    true,
		"decimal": true,
		//"distinct" : true,
		"double":  true,
		"float":   true,
		"integer": true,
		//"java_object" : true,
		"longnvarchar":  true,
		"longvarbinary": true,
		"longvarchar":   true,
		"nchar":         true,
		"nclob":         true,
		//"null" : true,
		"numeric":  true,
		"nvarchar": true,
		//"other" : true,
		"real": true,
		//"ref" : true,
		//"ref_cursor" : true,
		//"rowid" : true,
		"smallint": true,
		//"sqlxml" : true,
		//"struct" : true,
		"time":                    true,
		"time_with_timezone":      true,
		"timestamp":               true,
		"timestamp_with_timezone": true,
		"tinyint":                 true,
		"varbinary":               true,
		"varchar":                 true,
		// extra:
		"geometry": true,
	}
	return &TypeConverter{driver, fromJdbc, toJdbc, valid}
}

func (t *TypeConverter) ToJdbc(jdbcType string, size string) string {
	jdbcType = strings.ToLower(jdbcType)
	if val, ok := t.toJdbc[t.driver][jdbcType+"("+size+")"]; ok {
		jdbcType = val
	}
	if val, ok := t.toJdbc[t.driver][jdbcType]; ok {
		jdbcType = val
	}
	if val, ok := t.toJdbc["simplified"][jdbcType]; ok {
		jdbcType = val
	}
	if _, ok := t.valid[jdbcType]; !ok {
		jdbcType = "clob"
	}
	return jdbcType
}

func (t *TypeConverter) FromJdbc(jdbcType string) string {
	jdbcType = strings.ToLower(jdbcType)
	if _, ok := t.fromJdbc[t.driver][jdbcType]; ok {
		jdbcType = t.fromJdbc[t.driver][jdbcType]
	}
	return jdbcType
}

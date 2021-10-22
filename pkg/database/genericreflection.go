package database

import (
	"regexp"
)

type GenericReflection struct {
	pdo           *LazyPdo
	driver        string
	database      string
	tables        map[string]bool
	typeConverter *TypeConverter
}

func NewGenericReflection(pdo *LazyPdo, driver string, database string, tables map[string]bool) *GenericReflection {
	return &GenericReflection{pdo, driver, database, tables, NewTypeConverter(driver)}
}

func (r *GenericReflection) GetIgnoredTables() []string {
	switch r.driver {
	case "pgsql":
		return []string{"spatial_ref_sys", "raster_columns", "raster_overviews", "geography_columns", "geometry_columns"}
	default:
		return []string{}
	}
}

func (r *GenericReflection) getTablesSQL() string {
	switch r.driver {
	case `mysql`:
		return `SELECT "TABLE_NAME", "TABLE_TYPE" FROM "INFORMATION_SCHEMA"."TABLES" WHERE "TABLE_TYPE" IN (\BASE TABLE\' , \'VIEW\') AND "TABLE_SCHEMA" = ? ORDER BY BINARY "TABLE_NAME"`
	case `pgsql`:
		return `SELECT c.relname as "TABLE_NAME", c.relkind as "TABLE_TYPE" FROM pg_catalog.pg_class c LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace WHERE c.relkind IN ('r', 'v') AND n.nspname <> 'pg_catalog' AND n.nspname <> 'information_schema' AND n.nspname !~ '^pg_toast' AND pg_catalog.pg_table_is_visible(c.oid) AND '' <> ? ORDER BY "TABLE_NAME";`
	case `sqlsrv`:
		return `SELECT o.name as "TABLE_NAME", o.xtype as "TABLE_TYPE" FROM sysobjects o WHERE o.xtype IN ('U', 'V') ORDER BY "TABLE_NAME"`
	case `sqlite`:
		return `SELECT t.name as "TABLE_NAME", t.type as "TABLE_TYPE" FROM sqlite_master t WHERE t.type IN ('table', 'view') AND '' <> ? ORDER BY "TABLE_NAME"`
	default:
		return `SELECT 1=0`
	}
}

func (r *GenericReflection) getTableColumnsSQL() string {
	switch r.driver {
	case `mysql`:
		return `SELECT "COLUMN_NAME", "IS_NULLABLE", "DATA_TYPE", "CHARACTER_MAXIMUM_LENGTH" as "CHARACTER_MAXIMUM_LENGTH", "NUMERIC_PRECISION", "NUMERIC_SCALE", "COLUMN_TYPE" FROM "INFORMATION_SCHEMA"."COLUMNS" WHERE "TABLE_NAME" = ? AND "TABLE_SCHEMA" = ? ORDER BY "ORDINAL_POSITION"`
	case `pgsql`:
		return `SELECT a.attname AS "COLUMN_NAME", case when a.attnotnull then 'NO' else 'YES' end as "IS_NULLABLE", pg_catalog.format_type(a.atttypid, -1) as "DATA_TYPE", case when a.atttypmod < 0 then NULL else a.atttypmod-4 end as "CHARACTER_MAXIMUM_LENGTH", case when a.atttypid != 1700 then NULL else ((a.atttypmod - 4) >> 16) & 65535 end as "NUMERIC_PRECISION", case when a.atttypid != 1700 then NULL else (a.atttypmod - 4) & 65535 end as "NUMERIC_SCALE", '' AS "COLUMN_TYPE" FROM pg_attribute a JOIN pg_class pgc ON pgc.oid = a.attrelid WHERE pgc.relname = ? AND '' <> ? AND a.attnum > 0 AND NOT a.attisdropped ORDER BY a.attnum;`
	case `sqlsrv`:
		return `SELECT c.name AS "COLUMN_NAME", c.is_nullable AS "IS_NULLABLE", t.Name AS "DATA_TYPE", (c.max_length/2) AS "CHARACTER_MAXIMUM_LENGTH", c.precision AS "NUMERIC_PRECISION", c.scale AS "NUMERIC_SCALE", '' AS "COLUMN_TYPE" FROM sys.columns c INNER JOIN sys.types t ON c.user_type_id = t.user_type_id WHERE c.object_id = OBJECT_ID(?) AND '' <> ? ORDER BY c.column_id`
	case `sqlite`:
		return `SELECT "name" AS "COLUMN_NAME", case when "notnull"==1 then 'no' else 'yes' end as "IS_NULLABLE", lower("type") AS "DATA_TYPE", 2147483647 AS "CHARACTER_MAXIMUM_LENGTH", 0 AS "NUMERIC_PRECISION", 0 AS "NUMERIC_SCALE", '' AS "COLUMN_TYPE" FROM pragma_table_info(?) WHERE '' <> ? ORDER BY "cid"`
	default:
		return `SELECT 1=0`
	}
}

func (r *GenericReflection) getTablePrimaryKeysSQL() string {
	switch r.driver {
	case `mysql`:
		return `SELECT "COLUMN_NAME" FROM "INFORMATION_SCHEMA"."KEY_COLUMN_USAGE" WHERE "CONSTRAINT_NAME" = 'PRIMARY' AND "TABLE_NAME" = ? AND "TABLE_SCHEMA" = ?`
	case `pgsql`:
		return `SELECT a.attname AS "COLUMN_NAME" FROM pg_attribute a JOIN pg_constraint c ON (c.conrelid, c.conkey[1]) = (a.attrelid, a.attnum) JOIN pg_class pgc ON pgc.oid = a.attrelid WHERE pgc.relname = ? AND '' <> ? AND c.contype = 'p'`
	case `sqlsrv`:
		return `SELECT c.NAME as "COLUMN_NAME" FROM sys.key_constraints kc inner join sys.objects t on t.object_id = kc.parent_object_id INNER JOIN sys.index_columns ic ON kc.parent_object_id = ic.object_id and kc.unique_index_id = ic.index_id INNER JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id WHERE kc.type = 'PK' and t.object_id = OBJECT_ID(?) and '' <> ?`
	case `sqlite`:
		return `SELECT "name" as "COLUMN_NAME" FROM pragma_table_info(?) WHERE "pk"=1 AND '' <> ?`
	default:
		return `SELECT 1=0`
	}
}

func (r *GenericReflection) getTableForeignKeysSQL() string {
	switch r.driver {
	case `mysql`:
		return `SELECT "COLUMN_NAME", "REFERENCED_TABLE_NAME" FROM "INFORMATION_SCHEMA"."KEY_COLUMN_USAGE" WHERE "REFERENCED_TABLE_NAME" IS NOT NULL AND "TABLE_NAME" = ? AND "TABLE_SCHEMA" = ?`
	case `pgsql`:
		return `SELECT a.attname AS "COLUMN_NAME", c.confrelid::regclass::text AS "REFERENCED_TABLE_NAME" FROM pg_attribute a JOIN pg_constraint c ON (c.conrelid, c.conkey[1]) = (a.attrelid, a.attnum) JOIN pg_class pgc ON pgc.oid = a.attrelid WHERE pgc.relname = ? AND '' <> ? AND c.contype  = 'f'`
	case `sqlsrv`:
		return `SELECT COL_NAME(fc.parent_object_id, fc.parent_column_id) AS "COLUMN_NAME", OBJECT_NAME (f.referenced_object_id) AS "REFERENCED_TABLE_NAME" FROM sys.foreign_keys AS f INNER JOIN sys.foreign_key_columns AS fc ON f.OBJECT_ID = fc.constraint_object_id WHERE f.parent_object_id = OBJECT_ID(?) and '' <> ?`
	case `sqlite`:
		return `SELECT "from" AS "COLUMN_NAME", "table" AS "REFERENCED_TABLE_NAME" FROM pragma_foreign_key_list(?) WHERE '' <> ?`
	default:
		return `SELECT 1=0`
	}
}

func (r *GenericReflection) GetDatabaseName() string {
	return r.database
}

func (r *GenericReflection) GetTables() []map[string]interface{} {
	results := r.query(r.getTablesSQL(), r.database)
	tables := r.tables
	mapArr := map[string]string{}
	switch r.driver {
	case `mysql`:
		mapArr = map[string]string{`BASE TABLE`: `table`, `VIEW`: `view`}
	case `pgsql`:
		mapArr = map[string]string{`r`: `table`, `v`: `view`}
	case `sqlsrv`:
		mapArr = map[string]string{`U`: `table`, `V`: `view`}
	case `sqlite`:
		mapArr = map[string]string{`table`: `table`, `view`: `view`}
	}
	if len(tables) > 0 {
		_results := results
		for _, result := range results {
			if _, ok := tables[result["TABLE_NAME"].(string)]; ok {
				_results = append(_results, result)
			}
		}
		results = _results
	}
	for index := range results {
		results[index]["TABLE_TYPE"] = mapArr[results[index]["TABLE_TYPE"].(string)]
	}
	return results
}

func (r *GenericReflection) GetTableColumns(tableName string, viewType string) []map[string]interface{} {
	sql := r.getTableColumnsSQL()
	results := r.query(sql, tableName, r.database)
	if viewType == "view" {
		for index := range results {
			results[index]["IS_NULLABLE"] = false
		}
	}
	if r.driver == "mysql" {
		for index, result := range results {
			// mysql does not properly reflect display width of types
			re := regexp.MustCompile(`|([a-z]+)(\(([0-9]+)(,([0-9]+))?\))?|`)
			matches := re.FindStringSubmatch(result["DATA_TYPE"].(string))
			results[index]["DATA_TYPE"] = matches[1]
			if _, ok := results[index]["CHARACTER_MAXIMUM_LENGTH"]; ok {
				if matches[3] != "" {
					results[index]["NUMERIC_PRECISION"] = matches[3]
				}
				if matches[5] != "" {
					results[index]["NUMERIC_SCALE"] = matches[5]
				}
			}
		}
	}
	if r.driver == "sqlite" {
		for index, result := range results {
			// sqlite does not reflect types on view columns
			re := regexp.MustCompile(`|([a-z]+)(\(([0-9]+)(,([0-9]+))?\))?|`)
			matches := re.FindStringSubmatch(result["DATA_TYPE"].(string))
			if matches[1] != "" {
				results[index]["DATA_TYPE"] = matches[1]
			} else {
				results[index]["DATA_TYPE"] = "text"
			}
			if matches[5] != "" {
				results[index]["NUMERIC_PRECISION"] = matches[3]
				results[index]["NUMERIC_SCALE"] = matches[5]
			} else if matches[3] != "" {
				results[index]["CHARACTER_MAXIMUM_LENGTH"] = matches[3]
			}
		}
	}
	return results
}

func (r *GenericReflection) GetTablePrimaryKeys(tableName string) []string {
	sql := r.getTablePrimaryKeysSQL()
	results := r.query(sql, tableName, r.database)
	var primaryKeys []string
	for _, result := range results {
		primaryKeys = append(primaryKeys, result["COLUMN_NAME"].(string))
	}
	return primaryKeys
}

func (r *GenericReflection) GetTableForeignKeys(tableName string) map[string]string {
	sql := r.getTableForeignKeysSQL()
	results := r.query(sql, tableName, r.database)
	foreignKeys := map[string]string{}
	for _, result := range results {
		foreignKeys[result["COLUMN_NAME"].(string)] = result["REFERENCED_TABLE_NAME"].(string)
	}
	return foreignKeys
}

func (r *GenericReflection) ToJdbcType(jdbcType string, size string) string {
	return r.typeConverter.ToJdbc(jdbcType, size)
}

func (r *GenericReflection) query(sql string, parameters ...interface{}) []map[string]interface{} {
	var results []map[string]interface{}
	r.pdo.PDO().Raw(sql, parameters...).First(&results)
	return results
}

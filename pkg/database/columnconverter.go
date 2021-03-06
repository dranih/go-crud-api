package database

import "fmt"

type ColumnConverter struct {
	driver string
}

func NewColumnConverter(driver string) *ColumnConverter {
	return &ColumnConverter{driver}
}

func (cc *ColumnConverter) ConvertColumnValue(column *ReflectedColumn, parameters []interface{}) string {
	if column.IsBoolean() {
		switch cc.driver {
		case `mysql`:
			return "IFNULL(IF(?,TRUE,FALSE),NULL)"
		case `pgsql`:
			return fmt.Sprintf("$%d", len(parameters)+1)
		case `sqlsrv`:
			return fmt.Sprintf("@p%d", len(parameters)+1)
		}
	}
	if column.IsBinary() {
		switch cc.driver {
		case `mysql`:
			return "FROM_BASE64(?)"
		case `pgsql`:
			return fmt.Sprintf("decode($%d, 'base64')", len(parameters)+1)
		case `sqlsrv`:
			return fmt.Sprintf("CONVERT(XML, @p%d).value('.','varbinary(max)')", len(parameters)+1)
		}
	}
	if column.IsGeometry() {
		switch cc.driver {
		case `mysql`:
			return "ST_GeomFromText(?)"
		case `pgsql`:
			return fmt.Sprintf("ST_GeomFromText($%d)", len(parameters)+1)
		case `sqlsrv`:
			return fmt.Sprintf("geometry::STGeomFromText(@p%d,0)", len(parameters)+1)
		}
	}
	switch cc.driver {
	case `sqlsrv`:
		return fmt.Sprintf("@p%d", len(parameters)+1)
	case `pgsql`:
		return fmt.Sprintf("$%d", len(parameters)+1)
	default:
		return "?"
	}
}

func (cc *ColumnConverter) ConvertColumnName(column *ReflectedColumn, value string) string {
	if column.IsBinary() {
		switch cc.driver {
		case "mysql":
			return "TO_BASE64(" + value + ") as " + value
		case "pgsql":
			return "encode(" + value + "::bytea, 'base64') as " + value
		case "sqlsrv":
			return "CASE WHEN " + value + " IS NULL THEN NULL ELSE (SELECT CAST(" + value + " as varbinary(max)) FOR XML PATH(''), BINARY BASE64) END as " + value
		}
	}
	if column.IsGeometry() {
		switch cc.driver {
		case "mysql", "pgsql":
			return "ST_AsText(" + value + ") as " + value
		case "sqlsrv":
			return "REPLACE(" + value + ".STAsText(),' (','(') as " + value
		}
	}
	return value
}

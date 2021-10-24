package database

type ColumnConverter struct {
	driver string
}

func NewColumnConverter(driver string) *ColumnConverter {
	return &ColumnConverter{driver}
}

/*
public function convertColumnValue(ReflectedColumn $column): string
{
	if ($column->isBoolean()) {
		switch ($this->driver) {
			case 'mysql':
				return "IFNULL(IF(?,TRUE,FALSE),NULL)";
			case 'pgsql':
				return "?";
			case 'sqlsrv':
				return "?";
		}
	}
	if ($column->isBinary()) {
		switch ($this->driver) {
			case 'mysql':
				return "FROM_BASE64(?)";
			case 'pgsql':
				return "decode(?, 'base64')";
			case 'sqlsrv':
				return "CONVERT(XML, ?).value('.','varbinary(max)')";
		}
	}
	if ($column->isGeometry()) {
		switch ($this->driver) {
			case 'mysql':
			case 'pgsql':
				return "ST_GeomFromText(?)";
			case 'sqlsrv':
				return "geometry::STGeomFromText(?,0)";
		}
	}
	return '?';
}
*/
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
		case "mysql":
		case "pgsql":
			return "ST_AsText(" + value + ") as " + value
		case "sqlsrv":
			return "REPLACE(" + value + ".STAsText(),' (','(') as " + value
		}
	}
	return value
}

/*
public function convertColumnName(ReflectedColumn $column, $value): string
{
	if ($column->isBinary()) {
		switch ($this->driver) {
			case 'mysql':
				return "TO_BASE64($value) as $value";
			case 'pgsql':
				return "encode($value::bytea, 'base64') as $value";
			case 'sqlsrv':
				return "CASE WHEN $value IS NULL THEN NULL ELSE (SELECT CAST($value as varbinary(max)) FOR XML PATH(''), BINARY BASE64) END as $value";
		}
	}
	if ($column->isGeometry()) {
		switch ($this->driver) {
			case 'mysql':
			case 'pgsql':
				return "ST_AsText($value) as $value";
			case 'sqlsrv':
				return "REPLACE($value.STAsText(),' (','(') as $value";
		}
	}
	return $value;
}
*/

package database

import (
	"fmt"
	"strconv"
	"strings"
)

type DataConverter struct {
	driver string
}

func NewDataConverter(driver string) *DataConverter {
	return &DataConverter{driver}
}

// Should check conv errors
func (dc *DataConverter) convertRecordValue(conversion, value string) interface{} {
	args := strings.Split(conversion, "|")
	switch args[0] {
	case "boolean":
		res, _ := strconv.ParseBool(value)
		return res
	case "integer":
		res, _ := strconv.Atoi(value)
		return res
	case "float":
		res, _ := strconv.ParseFloat(value, 32)
		return res
	case "decimal":
		res, _ := strconv.ParseFloat(value, 32)
		return res
	}
	return value
}

func (dc *DataConverter) getRecordValueConversion(column *ReflectedColumn) string {
	if column.IsBoolean() {
		return "boolean"
	}
	switch column.GetType() {
	case "integer":
	case "bigint":
		return "integer"
	case "float":
	case "double":
		return "float"
	case "decimal":
		if dc.driver == "sqlite" {
			return "decimal|" + fmt.Sprint(column.GetScale())
		}
	}
	return "none"
}

func (dc *DataConverter) ConvertRecords(table *ReflectedTable, columnNames []string, records *[]map[string]interface{}) {
	for _, columnName := range columnNames {
		column := table.GetColumn(columnName)
		conversion := dc.getRecordValueConversion(column)
		if conversion != "none" {
			for i, record := range *records {
				value, ok := record[columnName]
				if !ok {
					continue
				}
				(*records)[i][columnName] = dc.convertRecordValue(conversion, value.(string))
			}
		}

	}
}

/*
private function convertInputValue($conversion, $value)
{
	switch ($conversion) {
		case 'boolean':
			return $value ? 1 : 0;
		case 'base64url_to_base64':
			return str_pad(strtr($value, '-_', '+/'), ceil(strlen($value) / 4) * 4, '=', STR_PAD_RIGHT);
	}
	return $value;
}

private function getInputValueConversion(ReflectedColumn $column): string
{
	if ($column->isBoolean()) {
		return 'boolean';
	}
	if ($column->isBinary()) {
		return 'base64url_to_base64';
	}
	return 'none';
}

public function convertColumnValues(ReflectedTable $table, array &$columnValues)
{
	$columnNames = array_keys($columnValues);
	foreach ($columnNames as $columnName) {
		$column = $table->getColumn($columnName);
		$conversion = $this->getInputValueConversion($column);
		if ($conversion != 'none') {
			$value = $columnValues[$columnName];
			if ($value !== null) {
				$columnValues[$columnName] = $this->convertInputValue($conversion, $value);
			}
		}
	}
}
*/

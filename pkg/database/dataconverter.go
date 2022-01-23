package database

import (
	"fmt"
	"math"
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

//Something nasty here in type conversion
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
				(*records)[i][columnName] = dc.convertRecordValue(conversion, fmt.Sprint(value))
			}
		}

	}
}

func (dc *DataConverter) convertInputValue(conversion, value string) interface{} {
	switch conversion {
	case `boolean`:
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		} else {
			return false
		}
	case `base64url_to_base64`:
		value = strings.ReplaceAll(value, `-`, `+`)
		value = strings.ReplaceAll(value, `_`, `/`)
		padLen := int(math.Ceil(float64(len(value))/4) * 4)
		return value + strings.Repeat(`=`, padLen-len(value))
	}
	return value
}

func (dc *DataConverter) getInputValueConversion(column *ReflectedColumn) string {
	if column.IsBoolean() {
		return `boolean`
	}
	if column.IsBinary() {
		return `base64url_to_base64`
	}
	return `none`
}

func (dc *DataConverter) ConvertColumnValues(table *ReflectedTable, columnValues *map[string]interface{}) {
	for columnName := range *columnValues {
		column := table.GetColumn(columnName)
		conversion := dc.getInputValueConversion(column)
		if conversion != `none` {
			if value, exists := (*columnValues)[columnName]; exists {
				if value == nil {
					(*columnValues)[columnName] = nil
				} else {
					(*columnValues)[columnName] = dc.convertInputValue(conversion, fmt.Sprintf("%v", value))
				}
			}
		}
	}
}

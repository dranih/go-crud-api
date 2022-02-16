package database

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/dranih/go-crud-api/pkg/utils"
)

type DataConverter struct {
	driver string
}

func NewDataConverter(driver string) *DataConverter {
	return &DataConverter{driver}
}

// Should check conv errors
func (dc *DataConverter) convertRecordValue(conversion string, value interface{}) interface{} {
	args := strings.Split(conversion, "|")
	switch args[0] {
	case "boolean":
		switch v := value.(type) {
		case string:
			//If we have only 1 byte, we test it
			if b := []byte(v); len(b) == 1 {
				return b[0] == byte(1)
			}
			res, _ := strconv.ParseBool(v)
			return res
		case bool:
			return v
		}
	case "integer":
		switch v := value.(type) {
		case string:
			res, _ := strconv.Atoi(v)
			return res
		case int, int64:
			return v
		}
	case "float":
		switch v := value.(type) {
		case string:
			res, _ := strconv.ParseFloat(v, 32)
			return res
		case float32, float64:
			return v
		}
	case "decimal":
		var flt *float64
		switch v := value.(type) {
		case string:
			val, _ := strconv.ParseFloat(v, 32)
			flt = &val
		case float64:
			flt = &v
		case float32:
			val := float64(v)
			flt = &val
		}
		if flt != nil {
			decimals := 0
			if len(args) == 2 {
				decimals, _ = strconv.Atoi(args[1])
			}
			return utils.NumberFormat(*flt, decimals, ".", "")
		}
	case "date":
		if t, ok := value.(time.Time); ok {
			return fmt.Sprintf("%d-%02d-%02d", t.Year(), int(t.Month()), t.Day())
		}
	case "time":
		if t, ok := value.(time.Time); ok {
			return fmt.Sprintf("%02d:%02d:%02d", t.Hour(), int(t.Minute()), t.Second())
		}
	case "timestamp":
		if t, ok := value.(time.Time); ok {
			return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", t.Year(), int(t.Month()), t.Day(), t.Hour(), int(t.Minute()), t.Second())
		}
	}
	return value
}

func (dc *DataConverter) getRecordValueConversion(column *ReflectedColumn) string {
	if column.IsBoolean() {
		return "boolean"
	}
	switch column.GetType() {
	case "integer", "bigint":
		return "integer"
	case "float", "double":
		return "float"
	case "decimal":
		if dc.driver == "sqlite" {
			return "decimal|" + fmt.Sprint(column.GetScale())
		}
	case "date", "time", "timestamp":
		return column.GetType()
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
				(*records)[i][columnName] = dc.convertRecordValue(conversion, value)
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

package database

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type ReflectedColumn struct {
	name       string
	columnType string
	length     int
	precision  int
	scale      int
	nullable   bool
	pk         bool
	fk         string
}

const (
	DEFAULT_LENGTH    = 255
	DEFAULT_PRECISION = 19
	DEFAULT_SCALE     = 4
)

// done
func NewReflectedColumn(name string, columnType string, length int, precision int, scale int, nullable bool, pk bool, fk string) *ReflectedColumn {
	r := &ReflectedColumn{name, columnType, length, precision, scale, nullable, pk, fk}
	r.sanitize()
	return r
}

// done
func parseColumnType(columnType string, length, precision, scale *int) {
	if columnType == "" {
		return
	}
	var err error
	pos := strings.Index(columnType, "(")
	if pos >= 0 {
		dataSize := strings.TrimRight(columnType[pos+1:], ")")
		if *length != -1 {
			*length, err = strconv.Atoi(dataSize)
			if err != nil {
				*length = -1
				log.Printf("Error parsing column type length : %v", err)
			}
		} else {
			pos = strings.Index(dataSize, ",")
			if pos >= 0 {
				*precision, err = strconv.Atoi(dataSize[:pos])
				if err != nil {
					*precision = -1
					log.Printf("Error parsing column type precision : %v", err)
				}
				*scale, err = strconv.Atoi(dataSize[pos+1:])
				if err != nil {
					*scale = -1
					log.Printf("Error parsing column type scale : %v", err)
				}
			} else {
				*precision, err = strconv.Atoi(dataSize)
				if err != nil {
					*precision = -1
					log.Printf("Error parsing column type precision : %v", err)
				}
				*scale = 0
			}
		}
	}
}

// done
func getDataSize(length, precision, scale int) string {
	dataSize := ""
	if length != -1 {
		dataSize = fmt.Sprint(length)
	} else if precision != -1 {
		if scale != -1 {
			dataSize = fmt.Sprint(precision) + "," + fmt.Sprint(scale)
		} else {
			dataSize = fmt.Sprint(precision)
		}

	}
	return dataSize
}

// done
func NewReflectedColumnFromReflection(reflection *GenericReflection, columnResult map[string]interface{}) *ReflectedColumn {
	name := columnResult["COLUMN_NAME"].(string)
	dataType := columnResult["DATA_TYPE"].(string)
	length, exists := columnResult["CHARACTER_MAXIMUM_LENGTH"].(int)
	if !exists {
		length = -1
	}
	precision, exists := columnResult["NUMERIC_PRECISION"].(int)
	if !exists {
		precision = -1
	}
	scale, exists := columnResult["NUMERIC_SCALE"].(int)
	if !exists {
		scale = -1
	}
	columnType := columnResult["COLUMN_TYPE"].(string)
	parseColumnType(columnType, &length, &precision, &scale)
	dataSize := getDataSize(length, precision, scale)
	jdbcType := reflection.ToJdbcType(dataType, dataSize)
	var nullable bool
	switch t := columnResult["IS_NULLABLE"].(type) {
	case bool:
		nullable = t
	case string:
		nullableList := map[string]bool{"TRUE": true, "YES": true, "T": true, "Y": true, "1": true}
		_, nullable = nullableList[strings.ToUpper(columnResult["IS_NULLABLE"].(string))]
	}
	pk := false
	fk := ""
	return &ReflectedColumn{name, jdbcType, length, precision, scale, nullable, pk, fk}
}

func NewReflectedColumnFromJson(json map[string]interface{}) *ReflectedColumn {
	name := fmt.Sprint(json["name"])
	columnType := fmt.Sprint(json["type"])
	length := 0
	if l, exists := json["length"]; exists {
		i, e := strconv.Atoi(fmt.Sprint(l))
		if e == nil {
			length = i
		}
	}
	precision := 0
	if l, exists := json["precision"]; exists {
		i, e := strconv.Atoi(fmt.Sprint(l))
		if e == nil {
			precision = i
		}
	}
	scale := 0
	if l, exists := json["scale"]; exists {
		i, e := strconv.Atoi(fmt.Sprint(l))
		if e == nil {
			scale = i
		}
	}
	nullable := false
	if l, exists := json["nullable"]; exists {
		i, e := strconv.ParseBool(fmt.Sprint(l))
		if e == nil {
			nullable = i
		}
	}
	pk := false
	if l, exists := json["pk"]; exists {
		i, e := strconv.ParseBool(fmt.Sprint(l))
		if e == nil {
			pk = i
		}
	}
	fk := ""
	if l, exists := json["fk"]; exists {
		fk = fmt.Sprint(l)
	}

	return NewReflectedColumn(name, columnType, length, precision, scale, nullable, pk, fk)
}

func (rc *ReflectedColumn) sanitize() {
	if rc.HasLength() {
		rc.length = rc.GetLength()
	} else {
		rc.length = 0
	}
	if rc.HasPrecision() {
		rc.precision = rc.GetPrecision()
	} else {
		rc.precision = 0
	}
	if rc.HasScale() {
		rc.scale = rc.GetScale()
	} else {
		rc.scale = 0
	}
}

func (rc *ReflectedColumn) GetName() string {
	return rc.name
}

func (rc *ReflectedColumn) GetNullable() bool {
	return rc.nullable
}

func (rc *ReflectedColumn) GetType() string {
	return rc.columnType
}

func (rc *ReflectedColumn) GetLength() int {
	if rc.length > 0 {
		return rc.length
	}
	return DEFAULT_LENGTH
}

func (rc *ReflectedColumn) GetPrecision() int {
	if rc.precision > 0 {
		return rc.precision
	}
	return DEFAULT_PRECISION
}

func (rc *ReflectedColumn) GetScale() int {
	if rc.scale != -1 {
		return rc.scale
	} else {
		return DEFAULT_SCALE
	}
}

func (rc *ReflectedColumn) HasLength() bool {
	return rc.columnType == "varchar" || rc.columnType == "varbinary"
}

func (rc *ReflectedColumn) HasPrecision() bool {
	return rc.columnType == "decimal"
}

func (rc *ReflectedColumn) HasScale() bool {
	return rc.columnType == "decimal"
}

func (rc *ReflectedColumn) IsBinary() bool {
	switch rc.columnType {
	case
		"blob",
		"varbinary":
		return true
	}
	return false
}

func (rc *ReflectedColumn) IsBoolean() bool {
	return rc.columnType == "boolean"
}

func (rc *ReflectedColumn) IsGeometry() bool {
	return rc.columnType == "geometry"
}

func (rc *ReflectedColumn) IsInteger() bool {
	switch rc.columnType {
	case
		"integer",
		"bigint",
		"smallint",
		"tinyint":
		return true
	}
	return false
}

func (rc *ReflectedColumn) SetPk(value bool) {
	rc.pk = value
}

func (rc *ReflectedColumn) GetPk() bool {
	return rc.pk
}

func (rc *ReflectedColumn) SetFk(value string) {
	rc.fk = value
}

func (rc *ReflectedColumn) GetFk() string {
	return rc.fk
}

func (rc *ReflectedColumn) Serialize() map[string]interface{} {
	return map[string]interface{}{
		"name":      rc.name,
		"type":      rc.columnType,
		"length":    rc.length,
		"precision": rc.precision,
		"scale":     rc.scale,
		"nullable":  rc.nullable,
		"pk":        rc.pk,
		"fk":        rc.fk,
	}
}

func (rc *ReflectedColumn) JsonSerialize() map[string]interface{} {
	return rc.Serialize()
}

// json marshaling for struct ReflectedTable
func (rt *ReflectedColumn) MarshalJSON() ([]byte, error) {
	return json.Marshal(rt.Serialize())
}

package database

import (
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
	//r.sanitize()
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

/*
private static function getDataSize(int $length, int $precision, int $scale): string
{
	$dataSize = '';
	if ($length) {
		$dataSize = $length;
	} elseif ($precision) {
		if ($scale) {
			$dataSize = $precision . ',' . $scale;
		} else {
			$dataSize = $precision;
		}
	}
	return $dataSize;
}
*/
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
	nullableList := map[string]bool{"TRUE": true, "YES": true, "T": true, "Y": true, "1": true}
	_, nullable := nullableList[strings.ToUpper(columnResult["IS_NULLABLE"].(string))]
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

/*
private function sanitize()
{
	$this->length = $this->hasLength() ? $this->getLength() : 0;
	$this->precision = $this->hasPrecision() ? $this->getPrecision() : 0;
	$this->scale = $this->hasScale() ? $this->getScale() : 0;
}
*/
func (rc *ReflectedColumn) GetName() string {
	return rc.name
}

/*
public function getNullable(): bool
{
	return $this->nullable;
}
*/
func (rc *ReflectedColumn) GetType() string {
	return rc.columnType
}

/*
public function getLength(): int
{
	return $this->length ?: self::DEFAULT_LENGTH;
}

public function getPrecision(): int
{
	return $this->precision ?: self::DEFAULT_PRECISION;
}
*/
func (rc *ReflectedColumn) GetScale() int {
	if rc.scale != -1 {
		return rc.scale
	} else {
		return DEFAULT_SCALE
	}
}

/*
public function hasLength(): bool
{
	return in_array($this->type, ['varchar', 'varbinary']);
}

public function hasPrecision(): bool
{
	return $this->type == 'decimal';
}

public function hasScale(): bool
{
	return $this->type == 'decimal';
}
*/
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

/*
public function isInteger(): bool
{
	return in_array($this->type, ['integer', 'bigint', 'smallint', 'tinyint']);
}
*/
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

/*
public function serialize()
{
	return [
		'name' => $this->name,
		'type' => $this->type,
		'length' => $this->length,
		'precision' => $this->precision,
		'scale' => $this->scale,
		'nullable' => $this->nullable,
		'pk' => $this->pk,
		'fk' => $this->fk,
	];
}

public function jsonSerialize()
{
	return array_filter($this->serialize());
}
*/

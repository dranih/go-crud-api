package database

import (
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

/*
private static function parseColumnType(string $columnType, int &$length, int &$precision, int &$scale)
{
	if (!$columnType) {
		return;
	}
	$pos = strpos($columnType, '(');
	if ($pos) {
		$dataSize = rtrim(substr($columnType, $pos + 1), ')');
		if ($length) {
			$length = (int) $dataSize;
		} else {
			$pos = strpos($dataSize, ',');
			if ($pos) {
				$precision = (int) substr($dataSize, 0, $pos);
				$scale = (int) substr($dataSize, $pos + 1);
			} else {
				$precision = (int) $dataSize;
				$scale = 0;
			}
		}
	}
}
*/
// done
func getDataSize(length, precision, scale int) string {
	dataSize := ""
	if length != -1 {
		dataSize = string(length)
	} else if precision != -1 {
		if scale != -1 {
			dataSize = string(precision) + "," + string(scale)
		} else {
			dataSize = string(precision)
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

/*
public static function fromReflection(GenericReflection $reflection, array $columnResult): ReflectedColumn
{
	$name = $columnResult['COLUMN_NAME'];
	$dataType = $columnResult['DATA_TYPE'];
	$length = (int) $columnResult['CHARACTER_MAXIMUM_LENGTH'];
	$precision = (int) $columnResult['NUMERIC_PRECISION'];
	$scale = (int) $columnResult['NUMERIC_SCALE'];
	$columnType = $columnResult['COLUMN_TYPE'];
	self::parseColumnType($columnType, $length, $precision, $scale);
	$dataSize = self::getDataSize($length, $precision, $scale);
	$type = $reflection->toJdbcType($dataType, $dataSize);
	$nullable = in_array(strtoupper($columnResult['IS_NULLABLE']), ['TRUE', 'YES', 'T', 'Y', '1']);
	$pk = false;
	$fk = '';
	return new ReflectedColumn($name, $type, $length, $precision, $scale, $nullable, $pk, $fk);
}

public static function fromJson($json): ReflectedColumn
{
	$name = $json->name;
	$type = $json->type;
	$length = isset($json->length) ? (int) $json->length : 0;
	$precision = isset($json->precision) ? (int) $json->precision : 0;
	$scale = isset($json->scale) ? (int) $json->scale : 0;
	$nullable = isset($json->nullable) ? (bool) $json->nullable : false;
	$pk = isset($json->pk) ? (bool) $json->pk : false;
	$fk = isset($json->fk) ? $json->fk : '';
	return new ReflectedColumn($name, $type, $length, $precision, $scale, $nullable, $pk, $fk);
}

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
public function getName(): string
{
	return $this->name;
}

public function getNullable(): bool
{
	return $this->nullable;
}
*/
func (rc *ReflectedColumn) GetType() string {
	return rc.columnType
}

/*
public function getType(): string
{
	return $this->type;
}

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
public function getScale(): int
{
	return $this->scale ?: self::DEFAULT_SCALE;
}

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

/*
public function isBinary(): bool
{
	return in_array($this->type, ['blob', 'varbinary']);
}
*/
func (rc *ReflectedColumn) IsBoolean() bool {
	return rc.columnType == "boolean"
}

/*
public function isBoolean(): bool
{
	return $this->type == 'boolean';
}
*/
func (rc *ReflectedColumn) IsGeometry() bool {
	return rc.columnType == "geometry"
}

/*
public function isGeometry(): bool
{
	return $this->type == 'geometry';
}

public function isInteger(): bool
{
	return in_array($this->type, ['integer', 'bigint', 'smallint', 'tinyint']);
}
*/
func (rc *ReflectedColumn) SetPk(value bool) {
	rc.pk = value
}

/*
public function setPk($value)
{
	$this->pk = $value;
}
*/
func (rc *ReflectedColumn) GetPk() bool {
	return rc.pk
}

/*
public function getPk(): bool
{
	return $this->pk;
}
*/
func (rc *ReflectedColumn) SetFk(value string) {
	rc.fk = value
} /*
public function setFk($value)
{
	$this->fk = $value;
}
*/
func (rc *ReflectedColumn) GetFk() string {
	return rc.fk
}

/*
public function getFk(): string
{
	return $this->fk;
}

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

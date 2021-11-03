package database

import (
	"strings"
)

type ColumnIncluder struct {
}

func (ci *ColumnIncluder) isMandatory(tableName, columnName string, params map[string][]string) bool {
	if _, exists := params["mandatory"]; exists {
		for _, param := range params["mandatory"] {
			if param == tableName+"."+columnName {
				return true
			}
		}
	}
	return false
}

func (ci *ColumnIncluder) selectColumn(tableName string, primaryTable bool, params map[string][]string, paramName string, columnNames []string, include bool) []string {
	if _, exists := params[paramName]; !exists {
		return columnNames
	}
	columns := map[string]bool{}
	for _, columnName := range strings.Split(params[paramName][0], ",") {
		columns[columnName] = true
	}
	result := []string{}
	for _, columnName := range columnNames {
		_, match := columns[`*.*`]
		if !match {
			_, match = columns[tableName+`.*`]
			if !match {
				_, match = columns[tableName+`.`+columnName]
			}
		}
		if primaryTable && !match {
			_, match = columns[tableName+`*`]
			if !match {
				_, match = columns[columnName]
			}
		}
		if match {
			if include || ci.isMandatory(tableName, columnName, params) {
				result = append(result, columnName)
			}
		} else if !include || ci.isMandatory(tableName, columnName, params) {
			result = append(result, columnName)
		}
	}
	return result
}

func (ci *ColumnIncluder) GetNames(table *ReflectedTable, primaryTable bool, params map[string][]string) []string {
	tableName := table.GetName()
	results := table.GetColumnNames()
	results = ci.selectColumn(tableName, primaryTable, params, "include", results, true)
	results = ci.selectColumn(tableName, primaryTable, params, "exclude", results, false)
	return results
}

// Not sure for property exists
func (ci *ColumnIncluder) GetValues(table *ReflectedTable, primaryTable bool, record map[string]interface{}, params map[string][]string) []interface{} {
	results := []interface{}{}
	columnNames := ci.GetNames(table, primaryTable, params)
	for _, columnName := range columnNames {
		if value, exists := record[columnName]; exists {
			results = append(results, value)
		}
	}
	return results
}

/*

public function getValues(ReflectedTable $table, bool $primaryTable, $record, array $params): array
{
	$results = array();
	$columnNames = $this->getNames($table, $primaryTable, $params);
	foreach ($columnNames as $columnName) {
		if (property_exists($record, $columnName)) {
			$results[$columnName] = $record->$columnName;
		}
	}
	return $results;
}
*/

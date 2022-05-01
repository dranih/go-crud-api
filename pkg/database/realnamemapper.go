package database

import "strings"

type RealNameMapper struct {
	tableMapping         map[string]string
	reverseTableMapping  map[string]string
	columnMapping        map[string]map[string]string
	reverseColumnMapping map[string]map[string]string
}

func NewRealNameMapper(mapping map[string]string) *RealNameMapper {
	tableMapping := map[string]string{}
	reverseTableMapping := map[string]string{}
	columnMapping := map[string]map[string]string{}
	reverseColumnMapping := map[string]map[string]string{}
	for name, realName := range mapping {
		if strings.Index(name, ".") >= 0 && strings.Index(realName, ".") >= 0 {
			nameSplit := strings.SplitN(name, ".", 2)
			realNameSplit := strings.SplitN(realName, ".", 2)
			tableMapping[nameSplit[0]] = realNameSplit[0]
			reverseTableMapping[realNameSplit[0]] = nameSplit[0]
			if _, exists := columnMapping[nameSplit[0]]; !exists {
				columnMapping[nameSplit[0]] = map[string]string{}
			}
			columnMapping[nameSplit[0]][nameSplit[1]] = realNameSplit[1]
			if _, exists := reverseColumnMapping[realNameSplit[0]]; !exists {
				reverseColumnMapping[realNameSplit[0]] = map[string]string{}
			}
			reverseColumnMapping[realNameSplit[0]][realNameSplit[1]] = nameSplit[1]
		} else {
			tableMapping[name] = realName
			reverseTableMapping[realName] = name
		}
	}
	return &RealNameMapper{tableMapping, reverseTableMapping, columnMapping, reverseColumnMapping}
}

func (rnm *RealNameMapper) GetColumnRealName(tableName, columnName string) string {
	if tm, exists := rnm.reverseColumnMapping[tableName]; exists {
		if cn, exists := tm[columnName]; exists {
			return cn
		}
	}
	return columnName
}

func (rnm *RealNameMapper) GetTableRealName(tableName string) string {
	if tn, exists := rnm.reverseTableMapping[tableName]; exists {
		return tn
	}
	return tableName
}

func (rnm *RealNameMapper) GetColumnName(tableRealName, columnRealName string) string {
	if tm, exists := rnm.columnMapping[tableRealName]; exists {
		if cn, exists := tm[columnRealName]; exists {
			return cn
		}
	}
	return columnRealName
}

func (rnm *RealNameMapper) GetTableName(tableRealName string) string {
	if tn, exists := rnm.tableMapping[tableRealName]; exists {
		return tn
	}
	return tableRealName
}

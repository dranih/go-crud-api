package database

import (
	"encoding/json"
)

type ReflectedDatabase struct {
	tableTypes     map[string]string
	tableRealNames map[string]string
}

func NewReflectedDatabase(tableTypes, tableRealNames map[string]string) *ReflectedDatabase {
	return &ReflectedDatabase{tableTypes, tableRealNames}
}

// done
func NewReflectedDatabaseFromReflection(reflection *GenericReflection) *ReflectedDatabase {
	tableTypes := map[string]string{}
	tableRealNames := map[string]string{}
	for _, table := range reflection.GetTables() {
		tableName := table["TABLE_NAME"]
		tableType := table["TABLE_TYPE"]
		found := false
		for _, ignoredTable := range reflection.GetIgnoredTables() {
			if tableName == ignoredTable {
				found = true
			}
		}
		if found {
			continue
		}
		tableTypes[tableName.(string)] = tableType.(string)
		tableRealNames[tableName.(string)] = table["TABLE_REAL_NAME"].(string)
	}
	return NewReflectedDatabase(tableTypes, tableRealNames)
}

func NewReflectedDatabaseFromJson(data string) *ReflectedDatabase {
	var wants struct {
		Types     map[string]string
		RealNames map[string]string
	}
	if err := json.Unmarshal([]byte(data), &wants); err != nil {
		return nil
	} else {
		return NewReflectedDatabase(wants.Types, wants.RealNames)
	}
}

/*

   public static function fromJson($json): ReflectedDatabase
   {
       $tableTypes = (array) $json->tables;
       return new ReflectedDatabase($tableTypes);
   }
*/

// done
func (rd *ReflectedDatabase) HasTable(tableName string) bool {
	_, isPresent := rd.tableTypes[tableName]
	return isPresent
}

func (rd *ReflectedDatabase) GetType(tableName string) string {
	if val, ok := rd.tableTypes[tableName]; ok {
		return val
	}
	return ""
}

func (rd *ReflectedDatabase) GetRealName(tableName string) string {
	if trns, exists := rd.tableRealNames[tableName]; exists {
		return trns
	}
	return ""
}

func (rd *ReflectedDatabase) GetTableNames() []string {
	i, keys := 0, make([]string, len(rd.tableTypes))
	for key := range rd.tableTypes {
		keys[i] = key
		i++
	}
	return keys
}

func (rd *ReflectedDatabase) RemoveTable(tableName string) bool {
	if _, exists := rd.tableTypes[tableName]; !exists {
		return false
	}
	delete(rd.tableTypes, tableName)
	delete(rd.tableRealNames, tableName)
	return true
}

func (rd *ReflectedDatabase) Serialize() map[string]interface{} {
	return map[string]interface{}{
		"types":     rd.tableTypes,
		"realNames": rd.tableRealNames,
	}
}

// json marshaling for struct ReflectedDatabase
func (rd *ReflectedDatabase) MarshalJSON() ([]byte, error) {
	return json.Marshal(rd.Serialize())
}

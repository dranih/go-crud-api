package database

import (
	"encoding/json"
)

type ReflectedDatabase struct {
	tableTypes map[string]string
}

func NewReflectedDatabase(tableTypes map[string]string) *ReflectedDatabase {
	return &ReflectedDatabase{tableTypes}
}

// done
func NewReflectedDatabaseFromReflection(reflection *GenericReflection) *ReflectedDatabase {
	tableTypes := map[string]string{}
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
	}
	return NewReflectedDatabase(tableTypes)
}

func NewReflectedDatabaseFromJson(data string) *ReflectedDatabase {
	var tableTypes map[string]string
	if err := json.Unmarshal([]byte(data), &tableTypes); err != nil {
		return nil
	} else {
		return NewReflectedDatabase(tableTypes)
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
	return true
}

func (rd *ReflectedDatabase) Serialize() map[string]interface{} {
	return map[string]interface{}{
		"tables": rd.tableTypes,
	}
}

// json marshaling for struct ReflectedDatabase
func (rd *ReflectedDatabase) MarshalJSON() ([]byte, error) {
	return json.Marshal(rd.Serialize())
}

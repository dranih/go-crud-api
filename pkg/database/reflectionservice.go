package database

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/dranih/go-crud-api/pkg/cache"
	"github.com/dranih/go-crud-api/pkg/utils"
)

type ReflectionService struct {
	db       *GenericDB
	cache    cache.Cache
	ttl      int32
	database *ReflectedDatabase
	tables   map[string]*ReflectedTable
}

func NewReflectionService(db *GenericDB, cache cache.Cache, ttl int32) *ReflectionService {
	return &ReflectionService{db, cache, ttl, nil, map[string]*ReflectedTable{}}
}

// done
func (rs *ReflectionService) getDatabase() *ReflectedDatabase {
	if rs.database != nil {
		return rs.database
	}
	rs.database = rs.loadDatabase(true)
	return rs.database
}

// to finish with cache
func (rs *ReflectionService) loadDatabase(useCache bool) *ReflectedDatabase {
	key := fmt.Sprintf("%s-ReflectedDatabase", rs.db.GetCacheKey())
	var data string
	var database *ReflectedDatabase
	if useCache {
		data = rs.cache.Get(key)
	}
	if data != "" {
		ungzipData, err := utils.GzUncompress(data)
		if err == nil {
			database = NewReflectedDatabaseFromJson(ungzipData)
		} else {
			log.Printf("Error cache uncompress : %v", err)
		}
	}
	if database == nil {
		database = NewReflectedDatabaseFromReflection(rs.db.Reflection())
		if jsonData, err := json.Marshal(database); err == nil {
			if data, err := utils.GzCompress(string(jsonData)); err == nil {
				rs.cache.Set(key, data, rs.ttl)
			} else {
				log.Printf("Error cache compress : %v", err)
			}
		} else {
			log.Printf("Error marshaling database for caching : %v", err)
		}
	}
	return database
}

// to finish with cache
func (rs *ReflectionService) loadTable(tableName string, useCache bool) *ReflectedTable {
	key := fmt.Sprintf("%s-ReflectedTable(%s)", rs.db.GetCacheKey(), tableName)
	var data string
	var table *ReflectedTable
	if useCache {
		data = rs.cache.Get(key)
	}
	if data != "" {
		ungzipData, err := utils.GzUncompress(data)
		if err == nil {
			var jsonData map[string]interface{}
			if err := json.Unmarshal([]byte(ungzipData), &jsonData); err != nil {
				log.Printf("Error cache table unmarshalling : %v", err)
			} else {
				table = NewReflectedTableFromJson(jsonData)
			}
		} else {
			log.Printf("Error cache uncompress : %v", err)
		}
	}
	if table == nil {
		tableType := rs.getDatabase().GetType(tableName)
		table = NewReflectedTableFromReflection(rs.db.Reflection(), tableName, tableType)
		if jsonData, err := json.Marshal(table); err == nil {
			if data, err := utils.GzCompress(string(jsonData)); err == nil {
				rs.cache.Set(key, data, rs.ttl)
			} else {
				log.Printf("Error cache compress : %v", err)
			}
		} else {
			log.Printf("Error marshaling table for caching : %v", err)
		}
	}
	return table
}

func (rs *ReflectionService) RefreshTables() {
	rs.database = rs.loadDatabase(false)
}

func (rs *ReflectionService) RefreshTable(tableName string) {
	rs.tables[tableName] = rs.loadTable(tableName, false)
}

func (rs *ReflectionService) HasTable(tableName string) bool {
	return rs.getDatabase().HasTable(tableName)
}

func (rs *ReflectionService) GetType(tableName string) string {
	return rs.getDatabase().GetType(tableName)
}

func (rs *ReflectionService) GetTable(tableName string) *ReflectedTable {
	if _, ok := rs.tables[tableName]; !ok {
		rs.tables[tableName] = rs.loadTable(tableName, true)
	}
	return rs.tables[tableName]
}

func (rs *ReflectionService) GetTableNames() []string {
	return rs.getDatabase().GetTableNames()
}

/*func (rs *ReflectionService) RemoveTable(tableName string) bool {
	delete(rs.tables, tableName)
	return rs.getDatabase().RemoveTable(tableName)
}*/

/*
       public function removeTable(string $tableName): bool
       {
           unset($this->tables[$tableName]);
           return $this->database()->removeTable($tableName);
       }
   }
*/

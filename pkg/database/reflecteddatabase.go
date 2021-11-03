package database

type ReflectedDatabase struct {
	tableTypes map[string]string
}

func NewReflectedDatabase(tableTypes map[string]string) *ReflectedDatabase {
	return &ReflectedDatabase{tableTypes}
}

/*
   public function __construct(array $tableTypes)
    {
        $this->tableTypes = $tableTypes;
    }

*/

// done
func NewReflectedDatabaseFromReflection(reflection *GenericReflection) *ReflectedDatabase {
	tableTypes := map[string]string{}
	//[]map[string]interface{}
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

/*
   public static function fromReflection(GenericReflection $reflection): ReflectedDatabase
   {
       $tableTypes = [];
       foreach ($reflection->getTables() as $table) {
           $tableName = $table['TABLE_NAME'];
           $tableType = $table['TABLE_TYPE'];
           if (in_array($tableName, $reflection->getIgnoredTables())) {
               continue;
           }
           $tableTypes[$tableName] = $tableType;
       }
       return new ReflectedDatabase($tableTypes);
   }

   public static function fromJson($json): ReflectedDatabase
   {
       $tableTypes = (array) $json->tables;
       return new ReflectedDatabase($tableTypes);
   }
*/

// done
func (r *ReflectedDatabase) HasTable(tableName string) bool {
	_, isPresent := r.tableTypes[tableName]
	return isPresent
}

/*
   public function hasTable(string $tableName): bool
   {
       return isset($this->tableTypes[$tableName]);
   }
*/
func (rd *ReflectedDatabase) GetType(tableName string) string {
	if val, ok := rd.tableTypes[tableName]; ok {
		return val
	}
	return ""
}

func (rc *ReflectedDatabase) GetTableNames() []string {
	i, keys := 0, make([]string, len(rc.tableTypes))
	for key := range rc.tableTypes {
		keys[i] = key
		i++
	}
	return keys
}

/*
       public function removeTable(string $tableName): bool
       {
           if (!isset($this->tableTypes[$tableName])) {
               return false;
           }
           unset($this->tableTypes[$tableName]);
           return true;
       }

       public function serialize()
       {
           return [
               'tables' => $this->tableTypes,
           ];
       }

       public function jsonSerialize()
       {
           return $this->serialize();
       }
   }
*/

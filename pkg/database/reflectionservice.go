package database

type ReflectionService struct {
	db       *GenericDB
	cache    interface{}
	ttl      int
	database *ReflectedDatabase
	tables   map[string]*ReflectedTable
}

func NewReflectionService(db *GenericDB, cache interface{}, ttl int) *ReflectionService {
	return &ReflectionService{db, cache, ttl, nil, map[string]*ReflectedTable{}}
}

/*
		public function __construct(GenericDB $db, Cache $cache, int $ttl)
        {
            $this->db = $db;
            $this->cache = $cache;
            $this->ttl = $ttl;
            $this->database = null;
            $this->tables = [];
        }
*/
// done
func (rs *ReflectionService) getDatabase() *ReflectedDatabase {
	if rs.database != nil {
		return rs.database
	}
	rs.database = rs.loadDatabase(true)
	return rs.database
}

/*
   private function database(): ReflectedDatabase
   {
       if ($this->database) {
           return $this->database;
       }
       $this->database = $this->loadDatabase(true);
       return $this->database;
   }
*/
// to finish with cache
func (rs *ReflectionService) loadDatabase(useCache bool) *ReflectedDatabase {
	database := NewReflectedDatabaseFromReflection(rs.db.Reflection())
	return database
}

/*
   private function loadDatabase(bool $useCache): ReflectedDatabase
   {
       $key = sprintf('%s-ReflectedDatabase', $this->db->getCacheKey());
       $data = $useCache ? $this->cache->get($key) : '';
       if ($data != '') {
           $database = ReflectedDatabase::fromJson(json_decode(gzuncompress($data)));
       } else {
           $database = ReflectedDatabase::fromReflection($this->db->reflection());
           $data = gzcompress(json_encode($database, JSON_UNESCAPED_UNICODE));
           $this->cache->set($key, $data, $this->ttl);
       }
       return $database;
   }
*/
// to finish with cache
func (rs *ReflectionService) loadTable(tableName string, useCache bool) *ReflectedTable {
	tableType := rs.getDatabase().GetType(tableName)
	table := NewReflectedTableFromReflection(rs.db.Reflection(), tableName, tableType)
	return table
}

/*
   private function loadTable(string $tableName, bool $useCache): ReflectedTable
   {
       $key = sprintf('%s-ReflectedTable(%s)', $this->db->getCacheKey(), $tableName);
       $data = $useCache ? $this->cache->get($key) : '';
       if ($data != '') {
           $table = ReflectedTable::fromJson(json_decode(gzuncompress($data)));
       } else {
           $tableType = $this->database()->getType($tableName);
           $table = ReflectedTable::fromReflection($this->db->reflection(), $tableName, $tableType);
           $data = gzcompress(json_encode($table, JSON_UNESCAPED_UNICODE));
           $this->cache->set($key, $data, $this->ttl);
       }
       return $table;
   }
*/
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

package column

import (
	"github.com/dranih/go-crud-api/pkg/database"
)

type ReflectionService struct {
	db       *database.GenericDB
	cache    string
	ttl      int
	database *ReflectedDatabase
	tables   []string
}

func NewReflectionService(db *database.GenericDB, cache string, ttl int) *ReflectionService {
	return &ReflectionService{db, cache, ttl, nil, []string{}}
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
func (r *ReflectionService) getDatabase() *ReflectedDatabase {
	if r.database != nil {
		return r.database
	}
	r.database = r.loadDatabase(true)
	return r.database
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
func (r *ReflectionService) loadDatabase(useCache bool) *ReflectedDatabase {
	database := FromReflection(r.db.Reflection())
	return database
}

/* to finish
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

   public function refreshTables()
   {
       $this->database = $this->loadDatabase(false);
   }

   public function refreshTable(string $tableName)
   {
       $this->tables[$tableName] = $this->loadTable($tableName, false);
   }
*/
// done
func (r *ReflectionService) HasTable(tableName string) bool {
	return r.getDatabase().HasTable(tableName)
}

/*
   public function hasTable(string $tableName): bool
   {
       return $this->database()->hasTable($tableName);
   }

   public function getType(string $tableName): string
   {
       return $this->database()->getType($tableName);
   }
*/
func (r *ReflectionService) GetTable(tableName string) *ReflectedTable {

}

/*
       public function getTable(string $tableName): ReflectedTable
       {
           if (!isset($this->tables[$tableName])) {
               $this->tables[$tableName] = $this->loadTable($tableName, true);
           }
           return $this->tables[$tableName];
       }

       public function getTableNames(): array
       {
           return $this->database()->getTableNames();
       }

       public function removeTable(string $tableName): bool
       {
           unset($this->tables[$tableName]);
           return $this->database()->removeTable($tableName);
       }
   }
*/

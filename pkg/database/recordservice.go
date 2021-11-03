package database

import (
	"github.com/dranih/go-crud-api/pkg/record"
)

type RecordService struct {
	db         *GenericDB
	reflection *ReflectionService
	columns    *ColumnIncluder
	joiner     *RelationJoiner
	filters    *FilterInfo
	ordering   string
	pagination string
}

func NewRecordService(db *GenericDB, reflection *ReflectionService) *RecordService {
	ci := &ColumnIncluder{}
	return &RecordService{db, reflection, ci, NewRelationJoiner(reflection, ci), &FilterInfo{}, "", ""}
}

/*
   public function __construct(GenericDB $db, ReflectionService $reflection)
   {
       $this->db = $db;
       $this->reflection = $reflection;
       $this->columns = new ColumnIncluder();
       $this->joiner = new RelationJoiner($reflection, $this->columns);
       $this->filters = new FilterInfo();
       $this->ordering = new OrderingInfo();
       $this->pagination = new PaginationInfo();
   }

   private function sanitizeRecord(string $tableName, $record, string $id)
   {
       $keyset = array_keys((array) $record);
       foreach ($keyset as $key) {
           if (!$this->reflection->getTable($tableName)->hasColumn($key)) {
               unset($record->$key);
           }
       }
       if ($id != '') {
           $pk = $this->reflection->getTable($tableName)->getPk();
           foreach ($this->reflection->getTable($tableName)->getColumnNames() as $key) {
               $field = $this->reflection->getTable($tableName)->getColumn($key);
               if ($field->getName() == $pk->getName()) {
                   unset($record->$key);
               }
           }
       }
   }
*/
func (r *RecordService) HasTable(table string) bool {
	return r.reflection.HasTable(table)
}

/*
   public function hasTable(string $table): bool
   {
       return $this->reflection->hasTable($table);
   }

   public function getType(string $table): string
   {
       return $this->reflection->getType($table);
   }

   public function beginTransaction()
   {
       $this->db->beginTransaction();
   }

   public function commitTransaction()
   {
       $this->db->commitTransaction();
   }

   public function rollBackTransaction()
       $this->db->rollBackTransaction();
   }

   public function create(string $tableName,$record, array $params)
   {
       $this->sanitizeRecord($tableName, $record, '');
       $table = $this->reflection->getTable($tableName);
       $columnValues = $this->columns->getValues($table, true, $record, $params);
       return $this->db->createSingle($table, $columnValues);
   }

   public function read(string $tableName, string $id, array $params)
   {
       $table = $this->reflection->getTable($tableName);
       $this->joiner->addMandatoryColumns($table, $params);
       $columnNames = $this->columns->getNames($table, true, $params);
       $record = $this->db->selectSingle($table, $columnNames, $id);
       if ($record == null) {
           return null;
       }
       $records = array($record);
       $this->joiner->addJoins($table, $records, $params, $this->db);
       return $records[0];
   }

   public function update(string $tableName, string $id, $record, array $params)
   {
       $this->sanitizeRecord($tableName, $record, $id);
       $table = $this->reflection->getTable($tableName);
       $columnValues = $this->columns->getValues($table, true, $record, $params);
       return $this->db->updateSingle($table, $columnValues, $id);
   }

   public function delete(string $tableName, string $id, array $params)
   {
       $table = $this->reflection->getTable($tableName);
       return $this->db->deleteSingle($table, $id);
   }

   public function increment(string $tableName, string $id, $record, array $params)
   {
       $this->sanitizeRecord($tableName, $record, $id);
       $table = $this->reflection->getTable($tableName);
       $columnValues = $this->columns->getValues($table, true, $record, $params);
       return $this->db->incrementSingle($table, $columnValues, $id);
   }
*/
// not finished
func (rs *RecordService) List(tableName string, params map[string][]string) *record.ListDocument {
	table := rs.reflection.GetTable(tableName)
	rs.joiner.AddMandatoryColumns(table, &params)
	columnNames := rs.columns.GetNames(table, true, params)
	condition := rs.filters.GetCombinedConditions(table, params)
	/*  $columnOrdering = $this->ordering->getColumnOrdering($table, $params);
	if (!$this->pagination->hasPage($params)) {
	    $offset = 0;
	    $limit = $this->pagination->getPageLimit($params);
	    $count = -1;
	} else {
	    $offset = $this->pagination->getPageOffset($params);
	    $limit = $this->pagination->getPageLimit($params);
	    $count = $this->db->selectCount($table, $condition);
	}
	$records = $this->db->selectAll($table, $columnNames, $condition, $columnOrdering, $offset, $limit);
	$this->joiner->addJoins($table, $records, $params, $this->db);
	return new ListDocument($records, $count);*/
	records := rs.db.SelectAll(table, columnNames, condition, []string{}, 0, 10)
	count := rs.db.SelectCount(table, condition)
	return record.NewListDocument(records, count)
}

/*
       public function _list(string $tableName, array $params): ListDocument
       {
           $table = $this->reflection->getTable($tableName);
           $this->joiner->addMandatoryColumns($table, $params);
           $columnNames = $this->columns->getNames($table, true, $params);
           $condition = $this->filters->getCombinedConditions($table, $params);
           $columnOrdering = $this->ordering->getColumnOrdering($table, $params);
           if (!$this->pagination->hasPage($params)) {
               $offset = 0;
               $limit = $this->pagination->getPageLimit($params);
               $count = -1;
           } else {
               $offset = $this->pagination->getPageOffset($params);
               $limit = $this->pagination->getPageLimit($params);
               $count = $this->db->selectCount($table, $condition);
           }
           $records = $this->db->selectAll($table, $columnNames, $condition, $columnOrdering, $offset, $limit);
           $this->joiner->addJoins($table, $records, $params, $this->db);
           return new ListDocument($records, $count);
       }

       public function ping(): int
       {
           return $this->db->ping();
       }
   }
*/

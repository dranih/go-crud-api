package database

import (
	"github.com/dranih/go-crud-api/pkg/record"
	"gorm.io/gorm"
)

type RecordService struct {
	db         *GenericDB
	reflection *ReflectionService
	columns    *ColumnIncluder
	joiner     *RelationJoiner
	filters    *FilterInfo
	ordering   *OrderingInfo
	pagination *PaginationInfo
}

func NewRecordService(db *GenericDB, reflection *ReflectionService) *RecordService {
	ci := &ColumnIncluder{}
	return &RecordService{db, reflection, ci, NewRelationJoiner(reflection, ci), &FilterInfo{}, &OrderingInfo{}, &PaginationInfo{}}
}

/*
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
func (rs *RecordService) HasTable(table string) bool {
	return rs.reflection.HasTable(table)
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
*/
func (rs *RecordService) BeginTransaction() *gorm.DB {
	return rs.db.BeginTransaction()
}

func (rs *RecordService) CommitTransaction(tx *gorm.DB) {
	rs.db.CommitTransaction(tx)
}

func (rs *RecordService) RollBackTransaction(tx *gorm.DB) {
	rs.db.RollBackTransaction(tx)
}

/*
   public function create(string $tableName,$record, array $params)
   {
       $this->sanitizeRecord($tableName, $record, '');
       $table = $this->reflection->getTable($tableName);
       $columnValues = $this->columns->getValues($table, true, $record, $params);
       return $this->db->createSingle($table, $columnValues);
   }
*/
func (rs *RecordService) Read(tableName, id string, params map[string][]string) (map[string]interface{}, error) {
	table := rs.reflection.GetTable(tableName)
	rs.joiner.AddMandatoryColumns(table, &params)
	columnNames := rs.columns.GetNames(table, true, params)
	records := rs.db.SelectSingle(table, columnNames, id)
	if records == nil || len(records) < 0 {
		return nil, nil
	}
	rs.joiner.AddJoins(table, &records, params, rs.db)
	return records[0], nil
}

/*
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
// done
func (rs *RecordService) List(tableName string, params map[string][]string) *record.ListDocument {
	table := rs.reflection.GetTable(tableName)
	rs.joiner.AddMandatoryColumns(table, &params)
	columnNames := rs.columns.GetNames(table, true, params)
	condition := rs.filters.GetCombinedConditions(table, params)
	columnOrdering := rs.ordering.GetColumnOrdering(table, params)
	var offset, limit, count int
	if !rs.pagination.HasPage(params) {
		offset = 0
		limit = rs.pagination.GetPageLimit(params)
		count = -1
	} else {
		offset = rs.pagination.GetPageOffset(params)
		limit = rs.pagination.GetPageLimit(params)
		count = rs.db.SelectCount(table, condition)
	}
	records := rs.db.SelectAll(table, columnNames, condition, columnOrdering, offset, limit)
	rs.joiner.AddJoins(table, &records, params, rs.db)
	return record.NewListDocument(records, count)
}

/*
       public function ping(): int
       {
           return $this->db->ping();
       }
   }
*/

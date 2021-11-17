package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/dranih/go-crud-api/pkg/record"
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

func (rs *RecordService) sanitizeRecord(tableName string, record interface{}, id string) map[string]interface{} {
	recordMap := map[string]interface{}{}
	//record type : map[string]interface {}
	if recordMap, ok := record.(map[string]interface{}); ok {
		for key := range recordMap {
			if !rs.reflection.GetTable(tableName).HasColumn(key) {
				delete(recordMap, key)
			}
		}
		if id != "" {
			pk := rs.reflection.GetTable(tableName).GetPk()
			for _, key := range rs.reflection.GetTable(tableName).GetColumnNames() {
				field := rs.reflection.GetTable(tableName).GetColumn(key)
				if field.GetName() == pk.GetName() {
					delete(recordMap, key)
				}
			}
		}
		return recordMap
	} else {
		log.Printf("Unable to assert record type : %T\n", record)
	}
	return recordMap
}

func (rs *RecordService) HasTable(table string) bool {
	return rs.reflection.HasTable(table)
}

func (rs *RecordService) GetType(table string) string {
	return rs.reflection.GetType(table)
}

func (rs *RecordService) BeginTransaction() (*sql.Tx, error) {
	return rs.db.BeginTransaction()
}

func (rs *RecordService) CommitTransaction(tx *sql.Tx) {
	rs.db.CommitTransaction(tx)
}

func (rs *RecordService) RollBackTransaction(tx *sql.Tx) {
	rs.db.RollBackTransaction(tx)
}

func (rs *RecordService) Create(tableName string, record interface{}, params map[string][]string) (map[string]interface{}, error) {
	recordMap := rs.sanitizeRecord(tableName, record, "")
	table := rs.reflection.GetTable(tableName)
	columnValues := rs.columns.GetValues(table, true, recordMap, params)
	return rs.db.CreateSingle(table, columnValues)
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
func (rs *RecordService) Read(tableName string, id interface{}, params map[string][]string) (map[string]interface{}, error) {
	table := rs.reflection.GetTable(tableName)
	rs.joiner.AddMandatoryColumns(table, &params)
	columnNames := rs.columns.GetNames(table, true, params)
	records := rs.db.SelectSingle(table, columnNames, fmt.Sprint(id))
	if records == nil || len(records) < 0 {
		return nil, nil
	}
	rs.joiner.AddJoins(table, &records, params, rs.db)
	return records[0], nil
}

/*
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

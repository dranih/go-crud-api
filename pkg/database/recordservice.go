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

func (rs *RecordService) Create(tableName string, params map[string][]string, record ...interface{}) (interface{}, error) {
	recordMap := rs.sanitizeRecord(tableName, record[0], "")
	table := rs.reflection.GetTable(tableName)
	columnValues := rs.columns.GetValues(table, true, recordMap, params)
	return rs.db.CreateSingle(table, columnValues)
}

func (rs *RecordService) Read(tableName string, params map[string][]string, id ...interface{}) (interface{}, error) {
	table := rs.reflection.GetTable(tableName)
	rs.joiner.AddMandatoryColumns(table, &params)
	columnNames := rs.columns.GetNames(table, true, params)
	records := rs.db.SelectSingle(table, columnNames, fmt.Sprint(id[0]))
	if records == nil || len(records) < 0 {
		return nil, nil
	}
	rs.joiner.AddJoins(table, &records, params, rs.db)
	return records[0], nil
}

func (rs *RecordService) Update(tableName string, params map[string][]string, args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return 0, fmt.Errorf("Not enought arguments : %v", args)
	}
	id := fmt.Sprint(args[0])
	record := args[1]
	recordMap := rs.sanitizeRecord(tableName, record, id)
	table := rs.reflection.GetTable(tableName)
	columnValues := rs.columns.GetValues(table, true, recordMap, params)
	return rs.db.UpdateSingle(table, columnValues, id)
}

func (rs *RecordService) Delete(tableName string, params map[string][]string, args ...interface{}) (interface{}, error) {
	table := rs.reflection.GetTable(tableName)
	return rs.db.DeleteSingle(table, fmt.Sprint(args[0]))
}

func (rs *RecordService) Increment(tableName string, params map[string][]string, args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return 0, fmt.Errorf("Not enought arguments : %v", args)
	}
	id := fmt.Sprint(args[0])
	record := args[1]
	recordMap := rs.sanitizeRecord(tableName, record, id)
	table := rs.reflection.GetTable(tableName)
	columnValues := rs.columns.GetValues(table, true, recordMap, params)
	return rs.db.IncrementSingle(table, columnValues, id)
}

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

func (rs *RecordService) Ping() int {
	return rs.db.Ping()
}

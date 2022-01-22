package database

import (
	"fmt"
	"strconv"
	"strings"
)

type RelationJoiner struct {
	reflection *ReflectionService
	ordering   *OrderingInfo
	columns    *ColumnIncluder
}

type HabtmValues struct {
	PkValues map[string][]map[string]interface{}
	FkValues map[string]map[string]interface{}
}

func NewRelationJoiner(reflection *ReflectionService, columns *ColumnIncluder) *RelationJoiner {
	return &RelationJoiner{reflection, &OrderingInfo{}, columns}
}

func (rj *RelationJoiner) AddMandatoryColumns(table *ReflectedTable, params *map[string][]string) {
	_, exists1 := (*params)["join"]
	_, exists2 := (*params)["include"]
	if !exists1 || !exists2 {
		return
	}
	(*params)["mandatory"] = []string{}
	for _, tableNames := range (*params)["join"] {
		t1 := table
		for _, tableName := range strings.Split(tableNames, ",") {
			if !rj.reflection.HasTable(tableName) {
				continue
			}
			t2 := rj.reflection.GetTable(tableName)
			fks1 := t1.GetFksTo(t2.GetName())
			t3 := rj.hasAndBelongsToMany(t1, t2)
			if t3 != nil && len(fks1) > 0 {
				(*params)["mandatory"] = append((*params)["mandatory"], t2.GetName()+"."+t2.GetPk().GetName())
			}
			for _, fk := range fks1 {
				(*params)["mandatory"] = append((*params)["mandatory"], t1.GetName()+"."+fk.GetName())
			}
			fks2 := t2.GetFksTo(t1.GetName())
			if t3 != nil && len(fks2) > 0 {
				(*params)["mandatory"] = append((*params)["mandatory"], t1.GetName()+"."+t1.GetPk().GetName())
			}
			for _, fk := range fks2 {
				(*params)["mandatory"] = append((*params)["mandatory"], t2.GetName()+"."+fk.GetName())
			}
			t1 = t2
		}
	}
}

func (rj *RelationJoiner) getJoinsAsPathTree(params map[string][]string) *PathTree {
	joins := NewPathTree(nil)
	if join, exists := params["join"]; exists {
		for _, tableNames := range join {
			path := []string{}
			for _, tableName := range strings.Split(tableNames, ",") {
				if !rj.reflection.HasTable(tableName) {
					continue
				}
				t := rj.reflection.GetTable(tableName)
				if t != nil {
					path = append(path, t.GetName())
				}
			}
			joins.Put(path, nil)
		}
	}
	return joins
}

func (rj *RelationJoiner) AddJoins(table *ReflectedTable, records *[]map[string]interface{}, params map[string][]string, db *GenericDB) {
	joins := rj.getJoinsAsPathTree(params)
	rj.addJoinsForTables(table, joins, records, params, db)
}

func (rj *RelationJoiner) hasAndBelongsToMany(t1, t2 *ReflectedTable) *ReflectedTable {
	for _, tableName := range rj.reflection.GetTableNames() {
		t3 := rj.reflection.GetTable(tableName)
		if len(t3.GetFksTo(t1.GetName())) > 0 && len(t3.GetFksTo(t2.GetName())) > 0 {
			return t3
		}
	}
	return nil
}

func (rj *RelationJoiner) addJoinsForTables(t1 *ReflectedTable, joins *PathTree, records *[]map[string]interface{}, params map[string][]string, db *GenericDB) {
	for _, t2Name := range joins.tree.GetKeys() {
		t2 := rj.reflection.GetTable(t2Name)
		belongsTo := len(t1.GetFksTo(t2.GetName())) > 0
		hasMany := len(t2.GetFksTo(t1.GetName())) > 0
		var t3 *ReflectedTable
		if !belongsTo && !hasMany {
			t3 = rj.hasAndBelongsToMany(t1, t2)
		} else {
			t3 = nil
		}
		hasAndBelongsToMany := (t3 != nil)
		newRecords := []map[string]interface{}{}
		var fkValues map[string]map[string]interface{}
		var pkValues map[string][]map[string]interface{}
		var habtmValues *HabtmValues
		if belongsTo {
			fkValues = rj.getFkEmptyValues(t1, t2, records)
			rj.addFkRecords(t2, fkValues, params, db, &newRecords)
		}
		if hasMany {
			pkValues = rj.getPkEmptyValues(t1, records)
			rj.addPkRecords(t1, t2, pkValues, params, db, &newRecords)
		}
		if hasAndBelongsToMany {
			habtmValues = rj.getHabtmEmptyValues(t1, t2, t3, db, records)
			rj.addFkRecords(t2, habtmValues.FkValues, params, db, &newRecords)
		}

		rj.addJoinsForTables(t2, joins.tree.Get(t2Name), &newRecords, params, db)

		if fkValues != nil {
			rj.fillFkValues(t2, newRecords, &fkValues)
			rj.setFkValues(t1, t2, records, fkValues)
		}
		if pkValues != nil {
			rj.fillPkValues(t1, t2, newRecords, &pkValues)
			rj.setPkValues(t1, t2, records, pkValues)
		}
		if habtmValues != nil {
			rj.fillFkValues(t2, newRecords, &(habtmValues.FkValues))
			rj.setHabtmValues(t1, t2, records, habtmValues)
		}
	}
}

func (rj *RelationJoiner) getFkEmptyValues(t1, t2 *ReflectedTable, records *[]map[string]interface{}) map[string]map[string]interface{} {
	fkValues := map[string]map[string]interface{}{}
	fks := t1.GetFksTo(t2.GetName())
	for _, fk := range fks {
		fkName := fk.GetName()
		for _, record := range *records {
			if fkValue, exists := record[fkName]; exists {
				fkValues[fmt.Sprint(fkValue)] = map[string]interface{}{}
			}
		}
	}
	return fkValues
}

func (rj *RelationJoiner) addFkRecords(t2 *ReflectedTable, fkValues map[string]map[string]interface{}, params map[string][]string, db *GenericDB, records *[]map[string]interface{}) {
	columnNames := rj.columns.GetNames(t2, false, params)
	fkIds := []string{}
	for key := range fkValues {
		fkIds = append(fkIds, key)
	}
	for _, record := range db.SelectMultiple(t2, columnNames, fkIds) {
		*records = append(*records, record)
	}
}

func (rj *RelationJoiner) fillFkValues(t2 *ReflectedTable, fkRecords []map[string]interface{}, fkValues *map[string]map[string]interface{}) {
	pkName := t2.GetPk().GetName()
	for _, fkRecord := range fkRecords {
		if pkValue, exists := fkRecord[pkName]; exists {
			(*fkValues)[fmt.Sprint(pkValue)] = fkRecord
		}
	}
}

func (rj *RelationJoiner) setFkValues(t1, t2 *ReflectedTable, records *[]map[string]interface{}, fkValues map[string]map[string]interface{}) {
	fks := t1.GetFksTo(t2.GetName())
	for _, fk := range fks {
		fkName := fk.GetName()
		for i, record := range *records {
			if key, exists := record[fkName]; exists {
				(*records)[i][fkName] = fkValues[fmt.Sprint(key)]
			}
		}
	}
}

func (rj *RelationJoiner) getPkEmptyValues(t1 *ReflectedTable, records *[]map[string]interface{}) map[string][]map[string]interface{} {
	pkValues := map[string][]map[string]interface{}{}
	pkName := t1.GetPk().GetName()
	for _, record := range *records {
		if pkValue, exists := record[pkName]; exists {
			pkValues[fmt.Sprint(pkValue)] = []map[string]interface{}{}
		}
	}
	return pkValues
}

func (rj *RelationJoiner) addPkRecords(t1, t2 *ReflectedTable, pkValues map[string][]map[string]interface{}, params map[string][]string, db *GenericDB, records *[]map[string]interface{}) {
	fks := t2.GetFksTo(t1.GetName())
	columnNames := rj.columns.GetNames(t2, false, params)

	pkIds := []string{}
	for key := range pkValues {
		pkIds = append(pkIds, key)
	}
	pkValueKeys := strings.Join(pkIds, `,`)
	conditions := []interface{ Condition }{}
	for _, fk := range fks {
		conditions = append(conditions, NewColumnCondition(fk, "in", pkValueKeys))
	}
	condition := OrConditionFromArray(conditions)
	columnOrdering := [][2]string{}
	limitInt := -1
	if limit := db.variablestore.Get("joinLimits.maxRecords"); limit != nil {
		columnOrdering = rj.ordering.GetDefaultColumnOrdering(t2)
		var err error
		if limitInt, err = strconv.Atoi(fmt.Sprint(limit)); err != nil {
			limitInt = -1
		}
	}
	for _, record := range db.SelectAll(t2, columnNames, condition, columnOrdering, 0, limitInt) {
		*records = append(*records, record)
	}
}

func (rj *RelationJoiner) fillPkValues(t1, t2 *ReflectedTable, pkRecords []map[string]interface{}, pkValues *map[string][]map[string]interface{}) {
	fks := t2.GetFksTo(t1.GetName())
	for _, fk := range fks {
		fkName := fk.GetName()
		for _, pkRecord := range pkRecords {
			key := fmt.Sprint(pkRecord[fkName])
			if _, exists := (*pkValues)[key]; exists {
				(*pkValues)[key] = append((*pkValues)[key], pkRecord)
			}
		}
	}
}

func (rj *RelationJoiner) setPkValues(t1, t2 *ReflectedTable, records *[]map[string]interface{}, pkValues map[string][]map[string]interface{}) {
	pkName := t1.GetPk().GetName()
	t2Name := t2.GetName()

	for i, record := range *records {
		key := fmt.Sprint(record[pkName])
		(*records)[i][t2Name] = pkValues[key]
	}
}

func (rj *RelationJoiner) getHabtmEmptyValues(t1, t2, t3 *ReflectedTable, db *GenericDB, records *[]map[string]interface{}) *HabtmValues {
	pkValues := rj.getPkEmptyValues(t1, records)
	fkValues := map[string]map[string]interface{}{}

	fk1 := t3.GetFksTo(t1.GetName())[0]
	fk2 := t3.GetFksTo(t2.GetName())[0]

	fk1Name := fk1.GetName()
	fk2Name := fk2.GetName()

	columnNames := []string{fk1Name, fk2Name}

	pkKeys := []string{}
	for key := range pkValues {
		pkKeys = append(pkKeys, key)
	}
	pkIds := strings.Join(pkKeys, `,`)
	condition := NewColumnCondition(t3.GetColumn(fk1Name), "in", pkIds)
	columnOrdering := [][2]string{}
	limitInt := -1
	if limit := db.variablestore.Get("joinLimits.maxRecords"); limit != nil {
		columnOrdering = rj.ordering.GetDefaultColumnOrdering(t3)
		if tempLimitInt, err := strconv.Atoi(fmt.Sprint(limit)); err == nil {
			limitInt = tempLimitInt
		}
	}
	for _, record := range db.SelectAll(t3, columnNames, condition, columnOrdering, 0, limitInt) {
		val1 := fmt.Sprint(record[fk1Name])
		val2 := fmt.Sprint(record[fk2Name])
		pkValues[val1] = append(pkValues[val1], map[string]interface{}{val2: ""})
		fkValues[val2] = map[string]interface{}{}
	}
	return &HabtmValues{pkValues, fkValues}
}

func (rj *RelationJoiner) setHabtmValues(t1, t2 *ReflectedTable, records *[]map[string]interface{}, habtmValues *HabtmValues) {
	pkName := t1.GetPk().GetName()
	t2Name := t2.GetName()
	for i, record := range *records {
		key := record[pkName]
		var val []map[string]interface{}
		fks := habtmValues.PkValues[fmt.Sprint(key)]
		for _, fkMap := range fks {
			for fk := range fkMap {
				val = append(val, habtmValues.FkValues[fmt.Sprint(fk)])
			}
		}
		(*records)[i][t2Name] = val
	}
}

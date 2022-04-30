package database

import (
	"encoding/json"
	"fmt"
	"sort"
)

type ReflectedTable struct {
	name      string
	realName  string
	tableType string
	columns   map[string]*ReflectedColumn
	pk        *ReflectedColumn
	fks       map[string]string
}

func NewReflectedTable(name, realName, tableType string, columns map[string]*ReflectedColumn) *ReflectedTable {
	r := &ReflectedTable{name, realName, tableType, map[string]*ReflectedColumn{}, nil, map[string]string{}}
	// set columns
	for _, column := range columns {
		columnName := column.GetName()
		r.columns[columnName] = column
	}
	// set primary key
	for _, column := range columns {
		if column.GetPk() {
			r.pk = column
		}
	}
	// set foreign keys
	for _, column := range columns {
		columnName := column.GetName()
		referencedTableName := column.GetFk()
		if referencedTableName != "" {
			r.fks[columnName] = referencedTableName
		}
	}
	return r
}

// done
func NewReflectedTableFromReflection(reflection *GenericReflection, name, realName, viewType string) *ReflectedTable {
	// set columns
	columns := map[string]*ReflectedColumn{}
	for _, tableColumn := range reflection.GetTableColumns(name, viewType) {
		column := NewReflectedColumnFromReflection(reflection, tableColumn)
		columns[column.GetName()] = column
	}
	// set primary key
	columnName := ""
	if viewType == "view" {
		columnName = "id"
	} else {
		columnNames := reflection.GetTablePrimaryKeys(name)
		if len(columnNames) == 1 {
			columnName = columnNames[0]
		}
	}
	if _, ok := columns[columnName]; columnName != "" && ok {
		columns[columnName].SetPk(true)
	}
	// set foreign keys
	if viewType == "view" {
		tables := reflection.GetTables()
		for columnName, column := range columns {
			if len(columnName)-3 >= 0 && columnName[len(columnName)-3:] == "_id" {
				for _, table := range tables {
					tableName := table["TABLE_NAME"].(string)
					suffix := tableName + "_id"
					if columnName[len(columnName)-len(suffix):] == suffix {
					}
					column.SetFk(tableName)
				}
			}
		}
	} else {
		fks := reflection.GetTableForeignKeys(name)
		for columnName, table := range fks {
			columns[columnName].SetFk(table)
		}
	}
	return NewReflectedTable(name, realName, viewType, columns)
}

func NewReflectedTableFromJson(json map[string]interface{}) *ReflectedTable {
	a, gotAlias := json["alias"]
	if n, exists := json["name"]; exists {
		name := fmt.Sprint(n)
		if gotAlias && a != nil {
			name = fmt.Sprint(a)
		}
		realName := fmt.Sprint(n)
		tableType := "table"
		if tt, exists := json["type"]; exists {
			tableType = fmt.Sprint(tt)
		}
		columns := map[string]*ReflectedColumn{}
		if jsonColumns, exists := json["columns"]; exists {
			switch c := jsonColumns.(type) {
			case []*ReflectedColumn:
				for _, column := range c {
					columns[column.GetName()] = column
				}
			case []interface{}:
				for _, column := range c {
					if tcolumn, ok := column.(map[string]interface{}); ok {
						rcolumn := NewReflectedColumnFromJson(tcolumn)
						columns[rcolumn.GetName()] = rcolumn
					}
				}
			}
		}
		return NewReflectedTable(name, realName, tableType, columns)
	}
	return nil
}

func (rt *ReflectedTable) HasColumn(columnName string) bool {
	_, exists := rt.columns[columnName]
	return exists
}

func (rt *ReflectedTable) HasPk() bool {
	return rt.pk != nil
}

func (rt *ReflectedTable) GetPk() *ReflectedColumn {
	return rt.pk
}

func (rt *ReflectedTable) GetName() string {
	return rt.name
}

func (rt *ReflectedTable) GetRealName() string {
	return rt.realName
}

func (rt *ReflectedTable) GetType() string {
	return rt.tableType
}

func (rt *ReflectedTable) GetColumnNames() []string {
	result := []string{}
	for key := range rt.columns {
		result = append(result, key)
	}
	return result
}

func (rt *ReflectedTable) GetColumn(columnName string) *ReflectedColumn {
	return rt.columns[columnName]
}

func (rt *ReflectedTable) GetFksTo(tableName string) []*ReflectedColumn {
	columns := []*ReflectedColumn{}
	for columnName, referencedTableName := range rt.fks {
		if _, exists := rt.columns[columnName]; tableName == referencedTableName && exists {
			columns = append(columns, rt.columns[columnName])
		}
	}
	return columns
}

func (rt *ReflectedTable) RemoveColumn(columnName string) bool {
	if _, exists := rt.columns[columnName]; !exists {
		return false
	}
	delete(rt.columns, columnName)
	return true
}

func (rt *ReflectedTable) Serialize() map[string]interface{} {
	var columns []*ReflectedColumn

	keys := make([]string, 0, len(rt.columns))
	for k := range rt.columns {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		columns = append(columns, rt.columns[k])
	}

	var a interface{}
	if rt.name != rt.realName {
		a = rt.name
	}

	return map[string]interface{}{
		"name":    rt.realName,
		"alias":   a,
		"type":    rt.tableType,
		"columns": columns,
	}
}

func (rt *ReflectedTable) JsonSerialize() map[string]interface{} {
	return rt.Serialize()
}

// json marshaling for struct ReflectedTable
func (rt *ReflectedTable) MarshalJSON() ([]byte, error) {
	return json.Marshal(rt.Serialize())
}

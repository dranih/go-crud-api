package database

import (
	"encoding/json"
	"fmt"
)

type ReflectedTable struct {
	name      string
	tableType string
	columns   map[string]*ReflectedColumn
	pk        *ReflectedColumn
	fks       map[string]string
}

func NewReflectedTable(name, tableType string, columns map[string]*ReflectedColumn) *ReflectedTable {
	r := &ReflectedTable{name, tableType, map[string]*ReflectedColumn{}, nil, map[string]string{}}
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
func NewReflectedTableFromReflection(reflection *GenericReflection, name, viewType string) *ReflectedTable {
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
	return NewReflectedTable(name, viewType, columns)
}

func NewReflectedTableFromJson(json map[string]interface{}) *ReflectedTable {
	if n, exists := json["name"]; exists {
		name := fmt.Sprint(n)
		tableType := "table"
		if tt, exists := json["type"]; exists {
			tableType = fmt.Sprint(tt)
		}
		columns := map[string]*ReflectedColumn{}
		if jsonColumns, exists := json["columns"]; exists {
			if c, ok := jsonColumns.([]*ReflectedColumn); ok {
				for _, column := range c {
					columns[column.GetName()] = column
				}
			}
		}
		return NewReflectedTable(name, tableType, columns)
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
	for _, c := range rt.columns {
		columns = append(columns, c)
	}
	return map[string]interface{}{
		"name":    rt.name,
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

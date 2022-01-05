package database

type DefinitionService struct {
	db         *GenericDB
	reflection *ReflectionService
}

func NewDefinitionService(db *GenericDB, reflection *ReflectionService) *DefinitionService {
	return &DefinitionService{db, reflection}
}

func (ds *DefinitionService) UpdateTable(tableName string, changes map[string]interface{}) bool {
	table := ds.reflection.GetTable(tableName)
	newTable := NewReflectedTableFromJson(mergeMaps(table.JsonSerialize(), changes))
	if table.GetName() != newTable.GetName() {
		if ds.db.Definition().RenameTable(table.GetName(), newTable.GetName()) != nil {
			return false
		}
		delete(ds.db.tables, table.GetName())
		ds.db.tables[newTable.GetName()] = true
		delete(ds.reflection.tables, table.GetName())
		ds.reflection.tables[newTable.GetName()] = newTable
	}
	return true
}

// mergeMaps naive merge m2 in m1 (if key exists in m1, override)
func mergeMaps(m1, m2 map[string]interface{}) map[string]interface{} {
	for key, val := range m2 {
		m1[key] = val
	}
	return m1
}

func (ds *DefinitionService) UpdateColumn(tableName, columnName string, changes map[string]interface{}) bool {
	table := ds.reflection.GetTable(tableName)
	column := table.GetColumn(columnName)

	// remove constraints on other column
	newColumn := NewReflectedColumnFromJson(mergeMaps(column.JsonSerialize(), changes))
	if newColumn.GetPk() != column.GetPk() && table.HasPk() {
		oldColumn := table.GetPk()
		if oldColumn.GetName() != columnName {
			oldColumn.SetPk(false)
			if ds.db.definition.RemoveColumnPrimaryKey(table.GetName(), oldColumn.GetName(), oldColumn) != nil {
				return false
			}
		}
	}

	// remove constraints
	newColumn = NewReflectedColumnFromJson(mergeMaps(column.JsonSerialize(), map[string]interface{}{"pk": false, "fk": false}))
	if newColumn.GetPk() != column.GetPk() && !newColumn.GetPk() {
		if ds.db.definition.RemoveColumnPrimaryKey(table.GetName(), column.GetName(), newColumn) != nil {
			return false
		}
	}
	if newColumn.GetFk() != column.GetFk() && newColumn.GetFk() != "" {
		if ds.db.definition.RemoveColumnForeignKey(table.GetName(), column.GetName(), newColumn) != nil {
			return false
		}
	}

	// name and type
	newColumn = NewReflectedColumnFromJson(mergeMaps(column.JsonSerialize(), changes))
	newColumn.SetPk(false)
	newColumn.SetFk("")
	if newColumn.GetName() != column.GetName() {
		if ds.db.definition.RenameColumn(table.GetName(), column.GetName(), newColumn) != nil {
			return false
		}
	}
	if newColumn.GetType() != column.GetType() ||
		newColumn.GetLength() != column.GetLength() ||
		newColumn.GetPrecision() != column.GetPrecision() ||
		newColumn.GetScale() != column.GetScale() {
		if ds.db.definition.RetypeColumn(table.GetName(), column.GetName(), newColumn) != nil {
			return false
		}
	}
	if newColumn.getNullable() != column.getNullable() {
		if ds.db.definition.SetColumnNullable(table.GetName(), column.GetName(), newColumn) != nil {
			return false
		}
	}

	// add constraints
	newColumn = NewReflectedColumnFromJson(mergeMaps(column.JsonSerialize(), changes))
	if newColumn.GetFk() != "" {
		if ds.db.definition.AddColumnForeignKey(table.GetName(), column.GetName(), newColumn) != nil {
			return false
		}
	}
	if newColumn.GetPk() {
		if ds.db.definition.AddColumnPrimaryKey(table.GetName(), column.GetName(), newColumn) != nil {
			return false
		}
	}

	return true
}

func (ds *DefinitionService) AddTable(definition map[string]interface{}) bool {
	newTable := NewReflectedTableFromJson(definition)
	if ds.db.definition.AddTable(newTable) != nil {
		return false
	}
	return true
}

func (ds *DefinitionService) AddColumn(tableName string, definition map[string]interface{}) bool {
	newColumn := NewReflectedColumnFromJson(definition)
	if ds.db.definition.AddColumn(tableName, newColumn) != nil {
		return false
	}
	if newColumn.GetFk() != "" {
		if ds.db.definition.AddColumnForeignKey(tableName, newColumn.GetName(), newColumn) != nil {
			return false
		}
	}
	if newColumn.GetPk() {
		if ds.db.definition.AddColumnPrimaryKey(tableName, newColumn.GetName(), newColumn) != nil {
			return false
		}
	}
	return true
}

func (ds *DefinitionService) RemoveTable(tableName string) bool {
	if ds.db.definition.RemoveTable(tableName) != nil {
		return false
	}
	return true
}

func (ds *DefinitionService) RemoveColumn(tableName, columnName string) bool {
	table := ds.reflection.GetTable(tableName)
	newColumn := table.GetColumn(columnName)
	if newColumn.GetPk() {
		newColumn.SetPk(false)
		if ds.db.definition.RemoveColumnPrimaryKey(table.GetName(), newColumn.GetName(), newColumn) != nil {
			return false
		}
	}
	if newColumn.GetFk() != "" {
		newColumn.SetFk("")
		if ds.db.definition.RemoveColumnForeignKey(tableName, newColumn.GetName(), newColumn) != nil {
			return false
		}
	}
	if ds.db.definition.RemoveColumn(tableName, columnName) != nil {
		return false
	}
	return true
}

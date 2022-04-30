package database

import "log"

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
	if table.GetRealName() != newTable.GetRealName() {
		if err := ds.db.Definition().RenameTable(table.GetRealName(), newTable.GetRealName()); err != nil {
			log.Printf("Error : %v", err)
			return false
		}
		if ds.db.tables != nil {
			delete(ds.db.tables, table.GetRealName())
			ds.db.tables[newTable.GetRealName()] = true
			delete(ds.reflection.tables, table.GetRealName())
			ds.reflection.tables[newTable.GetRealName()] = newTable
		}
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
		if oldColumn.GetRealName() != columnName {
			oldColumn.SetPk(false)
			if err := ds.db.definition.RemoveColumnPrimaryKey(table.GetRealName(), oldColumn.GetRealName(), oldColumn); err != nil {
				log.Printf("Error removing primary key for column %s : %v", oldColumn.GetRealName(), err)
				return false
			}
		}
	}

	// remove constraints
	newColumn = NewReflectedColumnFromJson(mergeMaps(column.JsonSerialize(), map[string]interface{}{"pk": false, "fk": ""}))
	if newColumn.GetPk() != column.GetPk() && !newColumn.GetPk() {
		if err := ds.db.definition.RemoveColumnPrimaryKey(table.GetRealName(), column.GetRealName(), newColumn); err != nil {
			log.Printf("Error removing primary key for column %s : %v", column.GetRealName(), err)
			return false
		}
	}
	if newColumn.GetFk() != column.GetFk() && newColumn.GetFk() == "" {
		if err := ds.db.definition.RemoveColumnForeignKey(table.GetRealName(), column.GetRealName(), newColumn); err != nil {
			log.Printf("Error removing foreign key for column %s : %v", column.GetRealName(), err)
			return false
		}
	}

	// name and type
	newColumn = NewReflectedColumnFromJson(mergeMaps(column.JsonSerialize(), changes))
	newColumn.SetPk(false)
	newColumn.SetFk("")
	if newColumn.GetRealName() != column.GetRealName() {
		if err := ds.db.definition.RenameColumn(table.GetRealName(), column.GetRealName(), newColumn); err != nil {
			log.Printf("Error rename column %s : %v", column.GetRealName(), err)
			return false
		}
	}
	if newColumn.GetType() != column.GetType() ||
		newColumn.GetLength() != column.GetLength() ||
		newColumn.GetPrecision() != column.GetPrecision() ||
		newColumn.GetScale() != column.GetScale() {
		if err := ds.db.definition.RetypeColumn(table.GetRealName(), newColumn.GetRealName(), newColumn); err != nil {
			log.Printf("Error changing type for column %s : %v", newColumn.GetRealName(), err)
			return false
		}
	}
	if newColumn.GetNullable() != column.GetNullable() {
		if err := ds.db.definition.SetColumnNullable(table.GetRealName(), newColumn.GetRealName(), newColumn); err != nil {
			log.Printf("Error changing nullable for column %s : %v", newColumn.GetRealName(), err)
			return false
		}
	}

	// add constraints
	newColumn = NewReflectedColumnFromJson(mergeMaps(column.JsonSerialize(), changes))
	if newColumn.GetFk() != "" {
		if err := ds.db.definition.AddColumnForeignKey(table.GetRealName(), newColumn.GetRealName(), newColumn); err != nil {
			log.Printf("Error adding foreign key for column %s : %v", newColumn.GetRealName(), err)
			return false
		}
	}
	if newColumn.GetPk() {
		if err := ds.db.definition.AddColumnPrimaryKey(table.GetRealName(), newColumn.GetRealName(), newColumn); err != nil {
			log.Printf("Error adding primary key for column %s : %v", newColumn.GetRealName(), err)
			return false
		}
	}

	return true
}

func (ds *DefinitionService) AddTable(definition map[string]interface{}) bool {
	newTable := NewReflectedTableFromJson(definition)
	if err := ds.db.definition.AddTable(newTable); err != nil {
		log.Printf("Error adding table %s : %v", newTable.GetRealName(), err)
		return false
	}
	if ds.db.tables != nil {
		ds.db.tables[newTable.GetRealName()] = true
		ds.reflection.tables[newTable.GetRealName()] = newTable
	}
	return true
}

func (ds *DefinitionService) AddColumn(tableName string, definition map[string]interface{}) bool {
	newColumn := NewReflectedColumnFromJson(definition)
	if err := ds.db.definition.AddColumn(tableName, newColumn); err != nil {
		log.Printf("Error adding column %s : %v", newColumn.GetRealName(), err)
		return false
	}
	if newColumn.GetFk() != "" {
		if err := ds.db.definition.AddColumnForeignKey(tableName, newColumn.GetRealName(), newColumn); err != nil {
			log.Printf("Error adding foreign key for column %s : %v", newColumn.GetRealName(), err)
			return false
		}
	}
	if newColumn.GetPk() {
		if err := ds.db.definition.AddColumnPrimaryKey(tableName, newColumn.GetRealName(), newColumn); err != nil {
			log.Printf("Error adding primary key for column %s : %v", newColumn.GetRealName(), err)
			return false
		}
	}
	return true
}

func (ds *DefinitionService) RemoveTable(tableName string) bool {
	if err := ds.db.definition.RemoveTable(tableName); err != nil {
		log.Printf("Error removing table %s : %v", tableName, err)
		return false
	}
	if ds.db.tables != nil {
		delete(ds.db.tables, tableName)
		delete(ds.reflection.tables, tableName)
	}
	return true
}

func (ds *DefinitionService) RemoveColumn(tableName, columnName string) bool {
	table := ds.reflection.GetTable(tableName)
	newColumn := table.GetColumn(columnName)
	if newColumn.GetPk() {
		newColumn.SetPk(false)
		if err := ds.db.definition.RemoveColumnPrimaryKey(table.GetRealName(), newColumn.GetRealName(), newColumn); err != nil {
			log.Printf("Error removing primary key for column %s : %v", newColumn.GetRealName(), err)
			return false
		}
	}
	if newColumn.GetFk() != "" {
		newColumn.SetFk("")
		if err := ds.db.definition.RemoveColumnForeignKey(tableName, newColumn.GetRealName(), newColumn); err != nil {
			log.Printf("Error removing foreign key for column %s : %v", newColumn.GetRealName(), err)
			return false
		}
	}
	if err := ds.db.definition.RemoveColumn(tableName, columnName); err != nil {
		log.Printf("Error removing column %s : %v", newColumn.GetRealName(), err)
		return false
	}
	return true
}

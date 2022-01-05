package controller

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/record"
	"github.com/gorilla/mux"
)

type ColumnController struct {
	responder  Responder
	reflection *database.ReflectionService
	definition *database.DefinitionService
}

func NewColumnController(router *mux.Router, responder Responder, reflection *database.ReflectionService, definition *database.DefinitionService) *ColumnController {
	cc := &ColumnController{responder, reflection, definition}
	router.HandleFunc("/columns", cc.getDatabase).Methods("GET")
	router.HandleFunc("/columns/{table}", cc.getTable).Methods("GET")
	router.HandleFunc("/columns/{table}/{column}", cc.getColumn).Methods("GET")
	router.HandleFunc("/columns/{table}", cc.updateTable).Methods("PUT")
	router.HandleFunc("/columns/{table}/{column}", cc.updateColumn).Methods("PUT")
	router.HandleFunc("/columns", cc.addTable).Methods("POST")
	router.HandleFunc("/columns/{table}", cc.addColumn).Methods("POST")
	router.HandleFunc("/columns/{table}", cc.removeTable).Methods("DELETE")
	router.HandleFunc("/columns/{table}/{column}", cc.removeColumn).Methods("DELETE")
	return cc
}

func (cc *ColumnController) getDatabase(w http.ResponseWriter, r *http.Request) {
	tables := []*database.ReflectedTable{}
	for _, table := range cc.reflection.GetTableNames() {
		tables = append(tables, cc.reflection.GetTable(table))
	}
	database := map[string][]*database.ReflectedTable{"tables": tables}
	cc.responder.Success(database, w)
}

func (cc *ColumnController) getTable(w http.ResponseWriter, r *http.Request) {
	tableName := mux.Vars(r)["table"]
	if !cc.reflection.HasTable(tableName) {
		cc.responder.Error(record.TABLE_NOT_FOUND, tableName, w, "")
		return
	}
	table := cc.reflection.GetTable(tableName)
	cc.responder.Success(table, w)
}

func (cc *ColumnController) getColumn(w http.ResponseWriter, r *http.Request) {
	tableName := mux.Vars(r)["table"]
	columnName := mux.Vars(r)["column"]
	if !cc.reflection.HasTable(tableName) {
		cc.responder.Error(record.TABLE_NOT_FOUND, tableName, w, "")
		return
	}
	table := cc.reflection.GetTable(tableName)
	if !table.HasColumn(columnName) {
		cc.responder.Error(record.COLUMN_NOT_FOUND, columnName, w, "")
		return
	}
	column := table.GetColumn(columnName)
	cc.responder.Success(column, w)
}

func (cc *ColumnController) updateTable(w http.ResponseWriter, r *http.Request) {
	tableName := mux.Vars(r)["table"]
	if !cc.reflection.HasTable(tableName) {
		cc.responder.Error(record.TABLE_NOT_FOUND, tableName, w, "")
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		cc.responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "")
		return
	}
	var jsonMap map[string]interface{}
	err = json.Unmarshal(b, &jsonMap)
	if err != nil {
		cc.responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "")
		return
	}
	success := cc.definition.UpdateTable(tableName, jsonMap)
	if success {
		cc.reflection.RefreshTables()
	}
	cc.responder.Success(success, w)
}

func (cc *ColumnController) updateColumn(w http.ResponseWriter, r *http.Request) {
	tableName := mux.Vars(r)["table"]
	columnName := mux.Vars(r)["column"]
	if !cc.reflection.HasTable(tableName) {
		cc.responder.Error(record.TABLE_NOT_FOUND, tableName, w, "")
		return
	}
	table := cc.reflection.GetTable(tableName)
	if !table.HasColumn(columnName) {
		cc.responder.Error(record.COLUMN_NOT_FOUND, columnName, w, "")
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		cc.responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "")
		return
	}
	var jsonMap map[string]interface{}
	err = json.Unmarshal(b, &jsonMap)
	if err != nil {
		cc.responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "")
		return
	}
	success := cc.definition.UpdateColumn(tableName, columnName, jsonMap)
	if success {
		cc.reflection.RefreshTable(tableName)
	}
	cc.responder.Success(success, w)
}

func (cc *ColumnController) addTable(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		cc.responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "")
		return
	}
	var jsonMap map[string]interface{}
	err = json.Unmarshal(b, &jsonMap)
	if err != nil {
		cc.responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "")
		return
	}
	if tableNameI, ok := jsonMap["name"]; ok {
		if tableName, ok := tableNameI.(string); ok {
			if cc.reflection.HasTable(tableName) {
				cc.responder.Error(record.TABLE_ALREADY_EXISTS, tableName, w, "")
				return
			}
		} else {
			cc.responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "Name argument not readable")
			return
		}
		success := cc.definition.AddTable(jsonMap)
		if success {
			cc.reflection.RefreshTables()
		}
		cc.responder.Success(success, w)
	} else {
		cc.responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "No name argument")
		return
	}
}

func (cc *ColumnController) addColumn(w http.ResponseWriter, r *http.Request) {
	tableName := mux.Vars(r)["table"]
	if !cc.reflection.HasTable(tableName) {
		cc.responder.Error(record.TABLE_NOT_FOUND, tableName, w, "")
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		cc.responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "")
		return
	}
	var jsonMap map[string]interface{}
	err = json.Unmarshal(b, &jsonMap)
	if err != nil {
		cc.responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "")
		return
	}
	table := cc.reflection.GetTable(tableName)
	if columnNameI, ok := jsonMap["name"]; ok {
		if columnName, ok := columnNameI.(string); ok {
			if table.HasColumn(columnName) {
				cc.responder.Error(record.COLUMN_ALREADY_EXISTS, columnName, w, "")
				return
			}
		} else {
			cc.responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "Name argument not readable")
			return
		}
		success := cc.definition.AddColumn(tableName, jsonMap)
		if success {
			cc.reflection.RefreshTable(tableName)
		}
		cc.responder.Success(success, w)
	} else {
		cc.responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "No name argument")
		return
	}
}

func (cc *ColumnController) removeTable(w http.ResponseWriter, r *http.Request) {
	tableName := mux.Vars(r)["table"]
	if !cc.reflection.HasTable(tableName) {
		cc.responder.Error(record.TABLE_NOT_FOUND, tableName, w, "")
		return
	}
	success := cc.definition.RemoveTable(tableName)
	if success {
		cc.reflection.RefreshTables()
	}
	cc.responder.Success(success, w)
}

func (cc *ColumnController) removeColumn(w http.ResponseWriter, r *http.Request) {
	tableName := mux.Vars(r)["table"]
	columnName := mux.Vars(r)["column"]
	if !cc.reflection.HasTable(tableName) {
		cc.responder.Error(record.TABLE_NOT_FOUND, tableName, w, "")
		return
	}
	table := cc.reflection.GetTable(tableName)
	if !table.HasColumn(columnName) {
		cc.responder.Error(record.COLUMN_NOT_FOUND, columnName, w, "")
		return
	}
	success := cc.definition.RemoveColumn(tableName, columnName)
	if success {
		cc.reflection.RefreshTable(tableName)
	}
	cc.responder.Success(success, w)
}

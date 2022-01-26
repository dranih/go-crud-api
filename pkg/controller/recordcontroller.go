package controller

import (
	"net/http"
	"strings"

	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/record"
	"github.com/dranih/go-crud-api/pkg/utils"
	"github.com/gorilla/mux"
)

type RecordController struct {
	service   *database.RecordService
	responder Responder
}

func NewRecordController(router *mux.Router, responder Responder, service *database.RecordService) *RecordController {
	rc := &RecordController{service, responder}
	router.HandleFunc("/records/{table}", rc.list).Methods("GET")
	router.HandleFunc("/records/{table}", rc.create).Methods("POST")
	router.HandleFunc("/records/{table}/{id}", rc.read).Methods("GET")
	router.HandleFunc("/records/{table}/{id}", rc.update).Methods("PUT")
	router.HandleFunc("/records/{table}/{id}", rc.delete).Methods("DELETE")
	router.HandleFunc("/records/{table}/{id}", rc.increment).Methods("PATCH")
	return rc
}

// List function lists a table
// Should return err error
func (rc *RecordController) list(w http.ResponseWriter, r *http.Request) {
	table := mux.Vars(r)["table"]
	params := utils.GetRequestParams(r)
	if !rc.service.HasTable(table) {
		rc.responder.Error(record.TABLE_NOT_FOUND, table, w, "")
		return
	}
	result := rc.service.List(table, params)
	rc.responder.Success(result, w)
	return
}

type argumentList struct {
	table   string
	payload []interface{}
	params  map[string][]string
}

// Should return err error
func (rc *RecordController) read(w http.ResponseWriter, r *http.Request) {
	table := mux.Vars(r)["table"]
	if !rc.service.HasTable(table) {
		rc.responder.Error(record.TABLE_NOT_FOUND, table, w, "")
		return
	}
	params := utils.GetRequestParams(r)
	id := mux.Vars(r)["id"]
	if strings.Index(id, ",") != -1 {
		ids := strings.Split(id, `,`)
		var argumentLists []*argumentList
		for i := 0; i < len(ids); i++ {
			argumentLists = append(argumentLists, &argumentList{table, []interface{}{ids[i]}, params})
		}
		result, errs := rc.multiCall(rc.service.Read, argumentLists)
		rc.responder.Multi(result, errs, w)
		return
	} else {
		response, err := rc.service.Read(table, params, id)
		if response == nil || err != nil {
			rc.responder.Error(record.RECORD_NOT_FOUND, id, w, "")
			return
		}
		rc.responder.Success(response, w)
	}
}

func (rc *RecordController) multiCall(callback func(string, map[string][]string, ...interface{}) (interface{}, error), argumentLists []*argumentList) (*[]interface{}, []error) {
	result := []interface{}{}
	var errs []error
	success := true
	tx, _ := rc.service.BeginTransaction()
	for _, arguments := range argumentLists {
		if tmp_result, err := callback(arguments.table, arguments.params, arguments.payload...); err == nil {
			result = append(result, tmp_result)
			errs = append(errs, nil)
		} else {
			success = false
			result = append(result, nil)
			errs = append(errs, err)
		}
	}
	if success {
		rc.service.CommitTransaction(tx)
	} else {
		rc.service.RollBackTransaction(tx)
	}
	return &result, errs
}

func (rc *RecordController) create(w http.ResponseWriter, r *http.Request) {
	table := mux.Vars(r)["table"]
	if !rc.service.HasTable(table) {
		rc.responder.Error(record.TABLE_NOT_FOUND, table, w, "")
		return
	}
	if rc.service.GetType(table) != "table" {
		rc.responder.Error(record.OPERATION_NOT_SUPPORTED, "create", w, "")
		return
	}
	jsonMap, err := utils.GetBodyData(r)
	if err != nil {
		rc.responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "")
		return
	}
	params := utils.GetRequestParams(r)

	// []interface{} or map[string]interface{}
	if records, isArray := jsonMap.([]interface{}); isArray {
		var argumentLists []*argumentList
		for _, record := range records {
			argumentLists = append(argumentLists, &argumentList{table, []interface{}{record}, params})
		}
		result, errs := rc.multiCall(rc.service.Create, argumentLists)
		rc.responder.Multi(result, errs, w)
		return
	} else {
		response, err := rc.service.Create(table, params, jsonMap)
		if response == nil || err != nil {
			rc.responder.Exception(err, w)
			return
		}
		rc.responder.Success(response, w)
	}
}

func (rc *RecordController) update(w http.ResponseWriter, r *http.Request) {
	table := mux.Vars(r)["table"]
	if !rc.service.HasTable(table) {
		rc.responder.Error(record.TABLE_NOT_FOUND, table, w, "")
		return
	}
	if rc.service.GetType(table) != "table" {
		rc.responder.Error(record.OPERATION_NOT_SUPPORTED, "update", w, "")
		return
	}
	jsonMap, err := utils.GetBodyData(r)
	if err != nil {
		rc.responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "")
		return
	}
	params := utils.GetRequestParams(r)
	id := mux.Vars(r)["id"]
	ids := strings.Split(id, `,`)

	if records, isArray := jsonMap.([]interface{}); isArray {
		if len(ids) != len(records) {
			rc.responder.Error(record.ARGUMENT_COUNT_MISMATCH, id, w, "")
			return
		}
		var argumentLists []*argumentList
		for i := 0; i < len(ids); i++ {
			argumentLists = append(argumentLists, &argumentList{table, []interface{}{ids[i], records[i]}, params})
		}
		result, errs := rc.multiCall(rc.service.Update, argumentLists)
		rc.responder.Multi(result, errs, w)
		return
	} else {
		if len(ids) != 1 {
			rc.responder.Error(record.ARGUMENT_COUNT_MISMATCH, id, w, "")
			return
		}
		response, err := rc.service.Update(table, params, id, jsonMap)
		if response == nil || err != nil {
			rc.responder.Exception(err, w)
			return
		}
		rc.responder.Success(response, w)
	}
}

func (rc *RecordController) delete(w http.ResponseWriter, r *http.Request) {
	table := mux.Vars(r)["table"]
	if !rc.service.HasTable(table) {
		rc.responder.Error(record.TABLE_NOT_FOUND, table, w, "")
		return
	}
	if rc.service.GetType(table) != "table" {
		rc.responder.Error(record.OPERATION_NOT_SUPPORTED, "delete", w, "")
		return
	}
	params := utils.GetRequestParams(r)
	id := mux.Vars(r)["id"]
	if strings.Index(id, ",") != -1 {
		ids := strings.Split(id, `,`)
		var argumentLists []*argumentList
		for i := 0; i < len(ids); i++ {
			argumentLists = append(argumentLists, &argumentList{table, []interface{}{ids[i]}, params})
		}
		result, errs := rc.multiCall(rc.service.Delete, argumentLists)
		rc.responder.Multi(result, errs, w)
		return
	} else {
		response, err := rc.service.Delete(table, params, id)
		if response == nil || err != nil {
			rc.responder.Exception(err, w)
			return
		}
		rc.responder.Success(response, w)
	}
}

func (rc *RecordController) increment(w http.ResponseWriter, r *http.Request) {
	table := mux.Vars(r)["table"]
	if !rc.service.HasTable(table) {
		rc.responder.Error(record.TABLE_NOT_FOUND, table, w, "")
		return
	}
	if rc.service.GetType(table) != "table" {
		rc.responder.Error(record.OPERATION_NOT_SUPPORTED, "update", w, "")
		return
	}
	jsonMap, err := utils.GetBodyData(r)
	if err != nil {
		rc.responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "")
		return
	}
	params := utils.GetRequestParams(r)
	id := mux.Vars(r)["id"]
	ids := strings.Split(id, `,`)

	if records, isArray := jsonMap.([]interface{}); isArray {
		if len(ids) != len(records) {
			rc.responder.Error(record.ARGUMENT_COUNT_MISMATCH, id, w, "")
			return
		}
		var argumentLists []*argumentList
		for i := 0; i < len(ids); i++ {
			argumentLists = append(argumentLists, &argumentList{table, []interface{}{ids[i], records[i]}, params})
		}
		result, errs := rc.multiCall(rc.service.Increment, argumentLists)
		rc.responder.Multi(result, errs, w)
		return
	} else {
		if len(ids) != 1 {
			rc.responder.Error(record.ARGUMENT_COUNT_MISMATCH, id, w, "")
			return
		}
		response, err := rc.service.Increment(table, params, id, jsonMap)
		if response == nil || err != nil {
			rc.responder.Exception(err, w)
			return
		}
		rc.responder.Success(response, w)
	}
}

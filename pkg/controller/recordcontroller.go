package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/record"
	"github.com/gorilla/mux"
)

type RecordController struct {
	service   *database.RecordService
	responder Responder
}

func NewRecordController(router *mux.Router, service *database.RecordService, debug bool) *RecordController {
	rc := &RecordController{service, NewJsonResponder(debug)}
	router.HandleFunc("/records/{table}", rc.list).Methods("GET")
	router.HandleFunc("/records/{table}", rc.create).Methods("POST")
	router.HandleFunc("/records/{table}/{id}", rc.read).Methods("GET")
	router.HandleFunc("/records/{table}/{id}", rc.update).Methods("PUT")
	return rc
}

/*
public function __construct(Router $router, Responder $responder, RecordService $service)
{
	$router->register('GET', '/records/*', array($this, '_list'));
	$router->register('POST', '/records/*', array($this, 'create'));
*/ //$router->register('GET', '/records/*/*', array($this, 'read'));
//$router->register('PUT', '/records/*/*', array($this, 'update'));
//$router->register('DELETE', '/records/*/*', array($this, 'delete'));
//$router->register('PATCH', '/records/*/*', array($this, 'increment'));
/*$this->service = $service;
	$this->responder = $responder;
}
*/
// List function lists a table
// Should return err error
func (rc *RecordController) list(w http.ResponseWriter, r *http.Request) {
	table := mux.Vars(r)["table"]
	params := getRequestParams(r)
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
	params := getRequestParams(r)
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

func (rc *RecordController) multiCall(callback func(string, map[string][]string, ...interface{}) (map[string]interface{}, error), argumentLists []*argumentList) (*[]map[string]interface{}, []error) {
	result := []map[string]interface{}{}
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
		rc.responder.Error(record.OPERATION_NOT_SUPPORTED, "Create", w, "")
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rc.responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "")
		return
	}
	var jsonMap interface{}
	err = json.Unmarshal(b, &jsonMap)
	if err != nil {
		rc.responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "")
		return
	}
	params := getRequestParams(r)

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
			rc.responder.Error(record.INTERNAL_SERVER_ERROR, fmt.Sprint(records), w, "")
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
		rc.responder.Error(record.OPERATION_NOT_SUPPORTED, "Create", w, "")
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rc.responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "")
		return
	}
	var jsonMap interface{}
	err = json.Unmarshal(b, &jsonMap)
	if err != nil {
		rc.responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "")
		return
	}
	params := getRequestParams(r)
	id := mux.Vars(r)["id"]
	ids := strings.Split(id, `,`)

	if records, isArray := jsonMap.([]interface{}); isArray {
		//if strings.Index(id, ",") != -1 {
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
		response, err := rc.service.Update(table, params, id)
		if response == nil || err != nil {
			rc.responder.Error(record.RECORD_NOT_FOUND, id, w, "")
			return
		}
		rc.responder.Success(response, w)
	}
}

/*
public function delete(ServerRequestInterface $request): ResponseInterface
{
	$table = RequestUtils::getPathSegment($request, 2);
	if (!$this->service->hasTable($table)) {
		return $this->responder->error(ErrorCode::TABLE_NOT_FOUND, $table);
	}
	if ($this->service->getType($table) != 'table') {
		return $this->responder->error(ErrorCode::OPERATION_NOT_SUPPORTED, __FUNCTION__);
	}
	$id = RequestUtils::getPathSegment($request, 3);
	$params = RequestUtils::getParams($request);
	$ids = explode(',', $id);
	if (count($ids) > 1) {
		$argumentLists = array();
		for ($i = 0; $i < count($ids); $i++) {
			$argumentLists[] = array($table, $ids[$i], $params);
		}
		return $this->responder->multi($this->multiCall([$this->service, 'delete'], $argumentLists));
	} else {
		return $this->responder->success($this->service->delete($table, $id, $params));
	}
}

public function increment(ServerRequestInterface $request): ResponseInterface
{
	$table = RequestUtils::getPathSegment($request, 2);
	if (!$this->service->hasTable($table)) {
		return $this->responder->error(ErrorCode::TABLE_NOT_FOUND, $table);
	}
	if ($this->service->getType($table) != 'table') {
		return $this->responder->error(ErrorCode::OPERATION_NOT_SUPPORTED, __FUNCTION__);
	}
	$id = RequestUtils::getPathSegment($request, 3);
	$record = $request->getParsedBody();
	if ($record === null) {
		return $this->responder->error(ErrorCode::HTTP_MESSAGE_NOT_READABLE, '');
	}
	$params = RequestUtils::getParams($request);
	$ids = explode(',', $id);
	if (is_array($record)) {
		if (count($ids) != count($record)) {
			return $this->responder->error(ErrorCode::ARGUMENT_COUNT_MISMATCH, $id);
		}
		$argumentLists = array();
		for ($i = 0; $i < count($ids); $i++) {
			$argumentLists[] = array($table, $ids[$i], $record[$i], $params);
		}
		return $this->responder->multi($this->multiCall([$this->service, 'increment'], $argumentLists));
	} else {
		if (count($ids) != 1) {
			return $this->responder->error(ErrorCode::ARGUMENT_COUNT_MISMATCH, $id);
		}
		return $this->responder->success($this->service->increment($table, $id, $record, $params));
	}
}
}*/

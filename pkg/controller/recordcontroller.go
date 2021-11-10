package controller

import (
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

func NewRecordController(router *mux.Router, service *database.RecordService) *RecordController {
	rc := &RecordController{service, NewJsonResponder(false)}
	router.HandleFunc("/records/{table}", rc.List).Methods("GET")
	router.HandleFunc("/records/{table}/{id}", rc.Read).Methods("GET")
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
func (rc *RecordController) List(w http.ResponseWriter, r *http.Request) {
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
	table  string
	id     string
	params map[string][]string
}

// Should return err error
func (rc *RecordController) Read(w http.ResponseWriter, r *http.Request) {
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
			argumentLists = append(argumentLists, &argumentList{table, ids[i], params})
		}
		rc.responder.Multi(rc.multiCall(rc.service.Read, argumentLists), w)
		return
	} else {
		response := rc.service.Read(table, id, params)
		if response == nil {
			rc.responder.Error(record.RECORD_NOT_FOUND, id, w, "")
			return
		}
		rc.responder.Success(response, w)
	}
}

/*
public function read(ServerRequestInterface $request): ResponseInterface
{
	$table = RequestUtils::getPathSegment($request, 2);
	if (!$this->service->hasTable($table)) {
		return $this->responder->error(ErrorCode::TABLE_NOT_FOUND, $table);
	}
	$id = RequestUtils::getPathSegment($request, 3);
	$params = RequestUtils::getParams($request);
	if (strpos($id, ',') !== false) {
		$ids = explode(',', $id);
		$argumentLists = array();
		for ($i = 0; $i < count($ids); $i++) {
			$argumentLists[] = array($table, $ids[$i], $params);
		}
		return $this->responder->multi($this->multiCall([$this->service, 'read'], $argumentLists));
	} else {
		$response = $this->service->read($table, $id, $params);
		if ($response === null) {
			return $this->responder->error(ErrorCode::RECORD_NOT_FOUND, $id);
		}
		return $this->responder->success($response);
	}
}

*/

// Use error instead of dealing with exceptions ?
func (rc *RecordController) multiCall(callback func(string, string, map[string][]string) map[string]interface{}, argumentLists []*argumentList) *[]map[string]interface{} {
	result := []map[string]interface{}{}
	success := true
	tx := rc.service.BeginTransaction()
	for _, arguments := range argumentLists {
		if tmp_result := callback(arguments.table, arguments.id, arguments.params); tmp_result != nil {
			result = append(result, tmp_result)
		}
	}
	if success {
		rc.service.CommitTransaction(tx)
	} else {
		rc.service.RollBackTransaction(tx)
	}
	return &result
}

/*
private function multiCall(callable $method, array $argumentLists): array
{
	$result = array();
	$success = true;
	$this->service->beginTransaction();
	foreach ($argumentLists as $arguments) {
		try {
			$result[] = call_user_func_array($method, $arguments);
		} catch (\Throwable $e) {
			$success = false;
			$result[] = $e;
		}
	}
	if ($success) {
		$this->service->commitTransaction();
	} else {
		$this->service->rollBackTransaction();
	}
	return $result;
}

public function create(ServerRequestInterface $request): ResponseInterface
{
	$table = RequestUtils::getPathSegment($request, 2);
	if (!$this->service->hasTable($table)) {
		return $this->responder->error(ErrorCode::TABLE_NOT_FOUND, $table);
	}
	if ($this->service->getType($table) != 'table') {
		return $this->responder->error(ErrorCode::OPERATION_NOT_SUPPORTED, __FUNCTION__);
	}
	$record = $request->getParsedBody();
	if ($record === null) {
		return $this->responder->error(ErrorCode::HTTP_MESSAGE_NOT_READABLE, '');
	}
	$params = RequestUtils::getParams($request);
	if (is_array($record)) {
		$argumentLists = array();
		foreach ($record as $r) {
			$argumentLists[] = array($table, $r, $params);
		}
		return $this->responder->multi($this->multiCall([$this->service, 'create'], $argumentLists));
	} else {
		return $this->responder->success($this->service->create($table, $record, $params));
	}
}

public function update(ServerRequestInterface $request): ResponseInterface
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
	$record = $request->getParsedBody();
	if ($record === null) {
		return $this->responder->error(ErrorCode::HTTP_MESSAGE_NOT_READABLE, '');
	}
	$ids = explode(',', $id);
	if (is_array($record)) {
		if (count($ids) != count($record)) {
			return $this->responder->error(ErrorCode::ARGUMENT_COUNT_MISMATCH, $id);
		}
		$argumentLists = array();
		for ($i = 0; $i < count($ids); $i++) {
			$argumentLists[] = array($table, $ids[$i], $record[$i], $params);
		}
		return $this->responder->multi($this->multiCall([$this->service, 'update'], $argumentLists));
	} else {
		if (count($ids) != 1) {
			return $this->responder->error(ErrorCode::ARGUMENT_COUNT_MISMATCH, $id);
		}
		return $this->responder->success($this->service->update($table, $id, $record, $params));
	}
}

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

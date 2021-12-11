package controller

import (
	"net/http"

	"github.com/dranih/go-crud-api/pkg/database"
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
	return cc
}

//public function __construct(Router $router, Responder $responder, ReflectionService $reflection, DefinitionService $definition)
//{
//$router->register('GET', '/columns', array($this, 'getDatabase'));
//$router->register('GET', '/columns/*', array($this, 'getTable'));
//$router->register('GET', '/columns/*/*', array($this, 'getColumn'));
//$router->register('PUT', '/columns/*', array($this, 'updateTable'));
//$router->register('PUT', '/columns/*/*', array($this, 'updateColumn'));
//$router->register('POST', '/columns', array($this, 'addTable'));
//$router->register('POST', '/columns/*', array($this, 'addColumn'));
//$router->register('DELETE', '/columns/*', array($this, 'removeTable'));
//$router->register('DELETE', '/columns/*/*', array($this, 'removeColumn'));
//$this->responder = $responder;
//$this->reflection = $reflection;
//$this->definition = $definition;
//}

func (cc *ColumnController) getDatabase(w http.ResponseWriter, r *http.Request) {
	tables := []*database.ReflectedTable{}
	for _, table := range cc.reflection.GetTableNames() {
		tables = append(tables, cc.reflection.GetTable(table))
	}
	database := map[string][]*database.ReflectedTable{"tables": tables}
	cc.responder.Success(database, w)
}

/*
public function getTable(ServerRequestInterface $request): ResponseInterface
{
	$tableName = RequestUtils::getPathSegment($request, 2);
	if (!$this->reflection->hasTable($tableName)) {
		return $this->responder->error(ErrorCode::TABLE_NOT_FOUND, $tableName);
	}
	$table = $this->reflection->getTable($tableName);
	return $this->responder->success($table);
}

public function getColumn(ServerRequestInterface $request): ResponseInterface
{
	$tableName = RequestUtils::getPathSegment($request, 2);
	$columnName = RequestUtils::getPathSegment($request, 3);
	if (!$this->reflection->hasTable($tableName)) {
		return $this->responder->error(ErrorCode::TABLE_NOT_FOUND, $tableName);
	}
	$table = $this->reflection->getTable($tableName);
	if (!$table->hasColumn($columnName)) {
		return $this->responder->error(ErrorCode::COLUMN_NOT_FOUND, $columnName);
	}
	$column = $table->getColumn($columnName);
	return $this->responder->success($column);
}

public function updateTable(ServerRequestInterface $request): ResponseInterface
{
	$tableName = RequestUtils::getPathSegment($request, 2);
	if (!$this->reflection->hasTable($tableName)) {
		return $this->responder->error(ErrorCode::TABLE_NOT_FOUND, $tableName);
	}
	$success = $this->definition->updateTable($tableName, $request->getParsedBody());
	if ($success) {
		$this->reflection->refreshTables();
	}
	return $this->responder->success($success);
}

public function updateColumn(ServerRequestInterface $request): ResponseInterface
{
	$tableName = RequestUtils::getPathSegment($request, 2);
	$columnName = RequestUtils::getPathSegment($request, 3);
	if (!$this->reflection->hasTable($tableName)) {
		return $this->responder->error(ErrorCode::TABLE_NOT_FOUND, $tableName);
	}
	$table = $this->reflection->getTable($tableName);
	if (!$table->hasColumn($columnName)) {
		return $this->responder->error(ErrorCode::COLUMN_NOT_FOUND, $columnName);
	}
	$success = $this->definition->updateColumn($tableName, $columnName, $request->getParsedBody());
	if ($success) {
		$this->reflection->refreshTable($tableName);
	}
	return $this->responder->success($success);
}

public function addTable(ServerRequestInterface $request): ResponseInterface
{
	$tableName = $request->getParsedBody()->name;
	if ($this->reflection->hasTable($tableName)) {
		return $this->responder->error(ErrorCode::TABLE_ALREADY_EXISTS, $tableName);
	}
	$success = $this->definition->addTable($request->getParsedBody());
	if ($success) {
		$this->reflection->refreshTables();
	}
	return $this->responder->success($success);
}

public function addColumn(ServerRequestInterface $request): ResponseInterface
{
	$tableName = RequestUtils::getPathSegment($request, 2);
	if (!$this->reflection->hasTable($tableName)) {
		return $this->responder->error(ErrorCode::TABLE_NOT_FOUND, $tableName);
	}
	$columnName = $request->getParsedBody()->name;
	$table = $this->reflection->getTable($tableName);
	if ($table->hasColumn($columnName)) {
		return $this->responder->error(ErrorCode::COLUMN_ALREADY_EXISTS, $columnName);
	}
	$success = $this->definition->addColumn($tableName, $request->getParsedBody());
	if ($success) {
		$this->reflection->refreshTable($tableName);
	}
	return $this->responder->success($success);
}

public function removeTable(ServerRequestInterface $request): ResponseInterface
{
	$tableName = RequestUtils::getPathSegment($request, 2);
	if (!$this->reflection->hasTable($tableName)) {
		return $this->responder->error(ErrorCode::TABLE_NOT_FOUND, $tableName);
	}
	$success = $this->definition->removeTable($tableName);
	if ($success) {
		$this->reflection->refreshTables();
	}
	return $this->responder->success($success);
}

public function removeColumn(ServerRequestInterface $request): ResponseInterface
{
	$tableName = RequestUtils::getPathSegment($request, 2);
	$columnName = RequestUtils::getPathSegment($request, 3);
	if (!$this->reflection->hasTable($tableName)) {
		return $this->responder->error(ErrorCode::TABLE_NOT_FOUND, $tableName);
	}
	$table = $this->reflection->getTable($tableName);
	if (!$table->hasColumn($columnName)) {
		return $this->responder->error(ErrorCode::COLUMN_NOT_FOUND, $columnName);
	}
	$success = $this->definition->removeColumn($tableName, $columnName);
	if ($success) {
		$this->reflection->refreshTable($tableName);
	}
	return $this->responder->success($success);
}
*/

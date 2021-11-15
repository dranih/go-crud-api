package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/dranih/go-crud-api/pkg/middleware"
)

type GenericDB struct {
	driver        string
	address       string
	port          int
	database      string
	tables        map[string]bool
	username      string
	password      string
	pdo           *LazyPdo
	reflection    *GenericReflection
	definition    string
	conditions    *ConditionsBuilder
	columns       *ColumnsBuilder
	converter     *DataConverter
	variablestore *middleware.VariableStore
}

func (g *GenericDB) getDsn() string {
	switch g.driver {
	case "mysql":
		return fmt.Sprintf("%s:host=%s;port=%d;dbname=%s;charset=utf8mb4", g.driver, g.address, g.port, g.database)
	case "pgsql":
		return fmt.Sprintf("%s:host=%s port=%d dbname=%s options=\"--client_encoding=UTF8\"", g.driver, g.address, g.port, g.database)
	case "sqlsrv":
		return fmt.Sprintf("%s:Server=%s,%d;Database=%s", g.driver, g.address, g.port, g.database)
	case "sqlite":
		return fmt.Sprintf("%s:%s", g.driver, g.address)
	default:
		return ""
	}
}

func (g *GenericDB) getCommands() []string {
	switch g.driver {
	case "mysql":
		return []string{
			"SET SESSION sql_warnings=1;",
			"SET NAMES utf8mb4;",
			"SET SESSION sql_mode = \"ANSI,TRADITIONAL\";",
		}
	case "pgsql":
		return []string{
			"SET NAMES \"UTF8\";",
		}
	case "sqlite":
		return []string{
			"PRAGMA foreign_keys = on;",
		}
	default:
		return []string{}
	}
}

func (g *GenericDB) getOptions() map[string]string {
	options := map[string]string{
		`\PDO::ATTR_ERRMODE`:            `\PDO::ERRMODE_EXCEPTION`,
		`\PDO::ATTR_DEFAULT_FETCH_MODE`: `\PDO::FETCH_ASSOC`,
	}
	switch g.driver {
	case "mysql":
		options[`\PDO::MYSQL_ATTR_FOUND_ROWS`] = "true"
		options[`\PDO::ATTR_PERSISTENT`] = "true"
	case "pgsql":
		options[`\PDO::ATTR_PERSISTENT`] = "true"
	}
	return options
}

// not finished
func (g *GenericDB) initPdo() bool {
	var result bool
	if g.pdo != nil {
		//result = $this->pdo->reconstruct($this->getDsn(), $this->username, $this->password, $this->getOptions());
		result = g.pdo.Reconstruct(g.getDsn(), g.username, g.password, g.getOptions())
	} else {
		g.pdo = NewLazyPdo(g.getDsn(), g.username, g.password, g.getOptions())
		result = true
	}
	commands := g.getCommands()
	for _, command := range commands {
		g.pdo.AddInitCommand(command)
	}

	g.reflection = NewGenericReflection(g.pdo, g.driver, g.database, g.tables)
	//$this->definition = new GenericDefinition($this->pdo, $this->driver, $this->database, $this->tables);
	g.conditions = NewConditionsBuilder(g.driver)
	g.columns = NewColumnsBuilder(g.driver)
	g.converter = NewDataConverter(g.driver)

	return result
}

func NewGenericDB(driver string, address string, port int, database string, tables map[string]bool, username string, password string) *GenericDB {
	g := &GenericDB{}
	g.driver = driver
	g.address = address
	g.port = port
	g.database = database
	g.tables = tables
	g.username = username
	g.password = password
	g.variablestore = &middleware.VariableStore{}
	g.initPdo()
	return g
}

/*
public function reconstruct(string $driver, string $address, int $port, string $database, array $tables, string $username, string $password): bool
        {
            if ($driver) {
                $this->driver = $driver;
            }
            if ($address) {
                $this->address = $address;
            }
            if ($port) {
                $this->port = $port;
            }
            if ($database) {
                $this->database = $database;
            }
            if ($tables) {
                $this->tables = $tables;
            }
            if ($username) {
                $this->username = $username;
            }
            if ($password) {
                $this->password = $password;
            }
            return $this->initPdo();
        }
*/
func (g *GenericDB) PDO() *LazyPdo {
	return g.pdo
}

func (g *GenericDB) Reflection() *GenericReflection {
	return g.reflection
}

/*
public function definition(): GenericDefinition
{
	return $this->definition;
}
*/
func (g *GenericDB) BeginTransaction() (*sql.Tx, error) {
	return g.pdo.BeginTransaction()
}

func (g *GenericDB) CommitTransaction(tx *sql.Tx) {
	g.pdo.Commit(tx)
}

func (g *GenericDB) RollBackTransaction(tx *sql.Tx) {
	g.pdo.RollBack(tx)
}

// Should type check
func (g *GenericDB) addMiddlewareConditions(tableName string, condition interface{ Condition }) interface{ Condition } {
	condition1 := g.variablestore.Get("authorization.conditions." + tableName)
	if condition1 != nil {
		condition = condition.And(condition1.(interface{ Condition }))
	}
	condition2 := g.variablestore.Get("multiTenancy.conditions." + tableName)
	if condition2 != nil {
		condition = condition.And(condition2.(interface{ Condition }))
	}
	return condition
}

/*

public function createSingle(ReflectedTable $table, array $columnValues)
{
	$this->converter->convertColumnValues($table, $columnValues);
	$insertColumns = $this->columns->getInsert($table, $columnValues);
	$tableName = $table->getName();
	$pkName = $table->getPk()->getName();
	$parameters = array_values($columnValues);
	$sql = 'INSERT INTO "' . $tableName . '" ' . $insertColumns;
	$stmt = $this->query($sql, $parameters);
	// return primary key value if specified in the input
	if (isset($columnValues[$pkName])) {
		return $columnValues[$pkName];
	}
	// work around missing "returning" or "output" in mysql
	switch ($this->driver) {
		case 'mysql':
			$stmt = $this->query('SELECT LAST_INSERT_ID()', []);
			break;
		case 'sqlite':
			$stmt = $this->query('SELECT LAST_INSERT_ROWID()', []);
			break;
	}
	$pkValue = $stmt->fetchColumn(0);
	if ($table->getPk()->getType() == 'bigint') {
		return (int) $pkValue;
	}
	if (in_array($table->getPk()->getType(), ['integer', 'bigint'])) {
		return (int) $pkValue;
	}
	return $pkValue;
}

*/
func (g *GenericDB) SelectSingle(table *ReflectedTable, columnNames []string, id string) []map[string]interface{} {
	records := []map[string]interface{}{}
	selectColumns := g.columns.GetSelect(table, columnNames)
	tableName := table.GetName()
	var condition interface{ Condition }
	condition = NewColumnCondition(table.GetPk(), `eq`, id)
	condition = g.addMiddlewareConditions(tableName, condition)
	parameters := []interface{}{}
	whereClause := g.conditions.GetWhereClause(condition, &parameters)
	sql := `SELECT ` + selectColumns + ` FROM "` + tableName + `" ` + whereClause
	records = g.query(sql, parameters...)
	if len(records) <= 0 {
		return nil
	}
	g.converter.ConvertRecords(table, columnNames, &records)
	return records[:1]
}

// done
func (g *GenericDB) SelectMultiple(table *ReflectedTable, columnNames, ids []string) []map[string]interface{} {
	records := []map[string]interface{}{}
	if len(ids) == 0 {
		return records
	}
	selectColumns := g.columns.GetSelect(table, columnNames)
	tableName := table.GetName()
	var condition interface{ Condition }
	condition = NewColumnCondition(table.GetPk(), `in`, strings.Join(ids, `,`))
	condition = g.addMiddlewareConditions(tableName, condition)
	parameters := []interface{}{}
	whereClause := g.conditions.GetWhereClause(condition, &parameters)
	sql := `SELECT ` + selectColumns + ` FROM "` + tableName + `" ` + whereClause
	records = g.query(sql, parameters...)
	g.converter.ConvertRecords(table, columnNames, &records)
	return records
}

// ok
func (g *GenericDB) SelectCount(table *ReflectedTable, condition interface{ Condition }) int {
	tableName := table.GetName()
	condition = g.addMiddlewareConditions(tableName, condition)
	parameters := []interface{}{}
	whereClause := g.conditions.GetWhereClause(condition, &parameters)
	sql := `SELECT COUNT(*) as c FROM "` + tableName + `"` + whereClause
	stmt := g.query(sql, parameters...)
	ret, ok := stmt[0]["c"].(int64)
	if !ok {
		log.Printf("Error converting count from table %s\n", tableName)
	}
	return int(ret)
}

// done
func (g *GenericDB) SelectAll(table *ReflectedTable, columnNames []string, condition interface{ Condition }, columnOrdering [][2]string, offset, limit int) []map[string]interface{} {
	if limit == 0 {
		return []map[string]interface{}{}
	}
	selectColumns := g.columns.GetSelect(table, columnNames)
	tableName := table.GetName()
	condition = g.addMiddlewareConditions(tableName, condition)
	parameters := []interface{}{}
	whereClause := g.conditions.GetWhereClause(condition, &parameters)
	orderBy := g.columns.GetOrderBy(table, columnOrdering)
	offsetLimit := g.columns.GetOffsetLimit(offset, limit)
	sql := "SELECT " + selectColumns + ` FROM "` + tableName + `"` + whereClause + orderBy + offsetLimit
	records := g.query(sql, parameters...)
	g.converter.ConvertRecords(table, columnNames, &records)
	return records
}

/*

public function updateSingle(ReflectedTable $table, array $columnValues, string $id)
{
	if (count($columnValues) == 0) {
		return 0;
	}
	$this->converter->convertColumnValues($table, $columnValues);
	$updateColumns = $this->columns->getUpdate($table, $columnValues);
	$tableName = $table->getName();
	$condition = new ColumnCondition($table->getPk(), 'eq', $id);
	$condition = $this->addMiddlewareConditions($tableName, $condition);
	$parameters = array_values($columnValues);
	$whereClause = $this->conditions->getWhereClause($condition, $parameters);
	$sql = 'UPDATE "' . $tableName . '" SET ' . $updateColumns . $whereClause;
	$stmt = $this->query($sql, $parameters);
	return $stmt->rowCount();
}

public function deleteSingle(ReflectedTable $table, string $id)
{
	$tableName = $table->getName();
	$condition = new ColumnCondition($table->getPk(), 'eq', $id);
	$condition = $this->addMiddlewareConditions($tableName, $condition);
	$parameters = array();
	$whereClause = $this->conditions->getWhereClause($condition, $parameters);
	$sql = 'DELETE FROM "' . $tableName . '" ' . $whereClause;
	$stmt = $this->query($sql, $parameters);
	return $stmt->rowCount();
}

public function incrementSingle(ReflectedTable $table, array $columnValues, string $id)
{
	if (count($columnValues) == 0) {
		return 0;
	}
	$this->converter->convertColumnValues($table, $columnValues);
	$updateColumns = $this->columns->getIncrement($table, $columnValues);
	$tableName = $table->getName();
	$condition = new ColumnCondition($table->getPk(), 'eq', $id);
	$condition = $this->addMiddlewareConditions($tableName, $condition);
	$parameters = array_values($columnValues);
	$whereClause = $this->conditions->getWhereClause($condition, $parameters);
	$sql = 'UPDATE "' . $tableName . '" SET ' . $updateColumns . $whereClause;
	$stmt = $this->query($sql, $parameters);
	return $stmt->rowCount();
}
*/

// Should check errors
func (g *GenericDB) query(sql string, parameters ...interface{}) []map[string]interface{} {
	rows, _ := g.pdo.connect().Query(sql, parameters...)
	results, _ := g.pdo.Rows2Map(rows)
	return results
}

/*
private function query(string $sql, array $parameters): \PDOStatement
{
	$stmt = $this->pdo->prepare($sql);
	//echo "- $sql -- " . json_encode($parameters, JSON_UNESCAPED_UNICODE) . "\n";
	$stmt->execute($parameters);
	return $stmt;
}

public function ping(): int
{
	$start = microtime(true);
	$stmt = $this->pdo->prepare('SELECT 1');
	$stmt->execute();
	return intval((microtime(true) - $start) * 1000000);
}

public function getCacheKey(): string
{
	return md5(json_encode([
		$this->driver,
		$this->address,
		$this->port,
		$this->database,
		$this->tables,
		$this->username,
	]));
}
}*/

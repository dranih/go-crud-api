package database

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/dranih/go-crud-api/pkg/utils"
)

type GenericDB struct {
	driver        string
	address       string
	port          int
	database      string
	tables        map[string]bool
	mapping       map[string]string
	username      string
	password      string
	pdo           *LazyPdo
	mapper        *RealNameMapper
	reflection    *GenericReflection
	definition    *GenericDefinition
	conditions    *ConditionsBuilder
	columns       *ColumnsBuilder
	converter     *DataConverter
	VariableStore *utils.VariableStore
}

func (g *GenericDB) getDsn() string {
	switch g.driver {
	case "mysql":
		//username:password@protocol(address)/dbname?param=value
		return fmt.Sprintf("%s:tcp(%s:%d)/%s?charset=utf8mb4&clientFoundRows=true", g.driver, g.address, g.port, g.database)
	case "pgsql":
		//return fmt.Sprintf("%s:host=%s port=%d dbname=%s options=\"--client_encoding=UTF8\"", g.driver, g.address, g.port, g.database)
		return fmt.Sprintf("%s:host=%s port=%d dbname=%s", g.driver, g.address, g.port, g.database)
	case "sqlsrv":
		return fmt.Sprintf("%s:server=%s;port=%d;database=%s", g.driver, g.address, g.port, g.database)
	case "sqlite":
		return fmt.Sprintf("%s:%s?_fk=1&defer_fk=1", g.driver, g.address)
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

func (g *GenericDB) initPdo() bool {
	var result bool
	if g.pdo != nil {
		result = g.pdo.Reconstruct(g.getDsn(), g.username, g.password, g.getOptions())
	} else {
		g.pdo = NewLazyPdo(g.getDsn(), g.username, g.password, g.getOptions())
		result = true
	}
	commands := g.getCommands()
	for _, command := range commands {
		g.pdo.AddInitCommand(command)
	}

	g.mapper = NewRealNameMapper(g.mapping)
	g.reflection = NewGenericReflection(g.pdo, g.driver, g.database, g.tables, g.mapper)
	g.definition = NewGenericDefinition(g.pdo, g.driver, g.database, g.tables, g.mapper)
	g.conditions = NewConditionsBuilder(g.driver)
	g.columns = NewColumnsBuilder(g.driver)
	g.converter = NewDataConverter(g.driver)

	return result
}

func NewGenericDB(driver string, address string, port int, database string, tables map[string]bool, mapping map[string]string, username string, password string) *GenericDB {
	g := &GenericDB{}
	g.driver = driver
	g.address = address
	g.port = port
	g.database = database
	g.tables = tables
	g.mapping = mapping
	g.username = username
	g.password = password
	g.VariableStore = utils.VStore
	g.initPdo()
	return g
}

func (g *GenericDB) Reconstruct(driver, address string, port int, database string, tables map[string]bool, mapping map[string]string, username string, password string) bool {
	if driver != "" {
		g.driver = driver
	}
	if address != "" {
		g.address = address
	}
	if port > 0 {
		g.port = port
	}
	if database != "" {
		g.database = database
	}
	if tables != nil {
		g.tables = tables
	}
	if mapping != nil {
		g.mapping = mapping
	}
	if username != "" {
		g.username = username
	}
	if password != "" {
		g.password = password
	}
	return g.initPdo()
}

func (g *GenericDB) PDO() *LazyPdo {
	return g.pdo
}

func (g *GenericDB) Reflection() *GenericReflection {
	return g.reflection
}

func (g *GenericDB) Definition() *GenericDefinition {
	return g.definition
}

func (g *GenericDB) BeginTransaction() (*sql.Tx, error) {
	return g.pdo.BeginTransaction()
}

func (g *GenericDB) CommitTransaction(tx *sql.Tx) error {
	return g.pdo.Commit(tx)
}

func (g *GenericDB) RollBackTransaction(tx *sql.Tx) error {
	return g.pdo.RollBack(tx)
}

// Should type check
func (g *GenericDB) addMiddlewareConditions(tableName string, condition interface{ Condition }) interface{ Condition } {
	condition1 := g.VariableStore.Get("authorization.conditions." + tableName)
	if condition1 != nil {
		condition = condition.And(condition1).(interface{ Condition })
	}
	condition2 := g.VariableStore.Get("multiTenancy.conditions." + tableName)
	if condition2 != nil {
		condition = condition.And(condition2).(interface{ Condition })
	}
	return condition
}

// getQuote returns the quote to use to escape columns and tables
func (g *GenericDB) getQuote() string {
	switch g.driver {
	case "mysql":
		return "`"
	default:
		return `"`
	}
}

func (g *GenericDB) CreateSingle(tx *sql.Tx, table *ReflectedTable, columnValues map[string]interface{}) (interface{}, error) {
	g.converter.ConvertColumnValues(table, &columnValues)
	insertColumns, parameters := g.columns.GetInsert(table, columnValues)
	tableRealName := table.GetRealName()
	pkName := table.GetPk().GetName()
	quote := g.getQuote()
	sql := fmt.Sprintf("INSERT INTO %s%s%s %s", quote, tableRealName, quote, insertColumns)
	//For pgsql and sqlsrv, get id from returning value
	if g.driver == "pgsql" || g.driver == "sqlsrv" {
		res, err := g.queryRowSingleColumn(tx, sql, parameters...)
		if err != nil {
			return nil, err
		}
		if table.GetPk().GetType() == `bigint` || table.GetPk().GetType() == `int` {
			if pkValueInt, ok := res.(int); ok {
				return pkValueInt, nil
			}
		}
		return res, nil
	} else {
		res, err := g.exec(tx, sql, parameters...)
		if err != nil {
			return nil, err
		}
		// return primary key value if specified in the input
		if pkValue, exists := columnValues[pkName]; exists {
			return pkValue, nil
		}
		id, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}
		return id, nil
	}
	/*
		// work around missing "returning" or "output" in mysql
		switch g.driver {
		case `mysql`:
			records, err = g.query("SELECT LAST_INSERT_ID()")
		case `sqlite`:
			records, err = g.query("SELECT LAST_INSERT_ROWID()")
		}
		if err != nil {
			return nil, err
		}
		for _, pkValue := range records[0] {
			if table.GetPk().GetType() == `bigint` || table.GetPk().GetType() == `int` {
				if pkValueInt, ok := pkValue.(int); ok {
					return map[string]interface{}{pkName: pkValueInt}, nil
				}
			}
			return map[string]interface{}{pkName: pkValue}, nil
		}
		return nil, errors.New("No Inserted ID")*/
}

// Should check error
func (g *GenericDB) SelectSingle(tx *sql.Tx, table *ReflectedTable, columnNames []string, id string) []map[string]interface{} {
	selectColumns := g.columns.GetSelect(table, columnNames)
	tableName := table.GetName()
	tableRealName := table.GetRealName()
	var condition interface{ Condition }
	condition = NewColumnCondition(table.GetPk(), `eq`, id)
	condition = g.addMiddlewareConditions(tableName, condition)
	parameters := []interface{}{}
	whereClause := g.conditions.GetWhereClause(condition, &parameters)
	quote := g.getQuote()
	sql := fmt.Sprintf("SELECT %s FROM %s%s%s %s", selectColumns, quote, tableRealName, quote, whereClause)
	records, _ := g.query(tx, sql, parameters...)
	if len(records) <= 0 {
		return nil
	}
	records = g.mapRecords(tableRealName, records)
	g.converter.ConvertRecords(table, columnNames, &records)
	return records[:1]
}

// Should check error
func (g *GenericDB) SelectMultiple(table *ReflectedTable, columnNames, ids []string) []map[string]interface{} {
	records := []map[string]interface{}{}
	if len(ids) == 0 {
		return records
	}
	selectColumns := g.columns.GetSelect(table, columnNames)
	tableName := table.GetName()
	tableRealName := table.GetRealName()
	var condition interface{ Condition }
	condition = NewColumnCondition(table.GetPk(), `in`, strings.Join(ids, `,`))
	condition = g.addMiddlewareConditions(tableName, condition)
	parameters := []interface{}{}
	whereClause := g.conditions.GetWhereClause(condition, &parameters)
	quote := g.getQuote()
	sql := fmt.Sprintf("SELECT %s FROM %s%s%s %s", selectColumns, quote, tableRealName, quote, whereClause)
	records, _ = g.query(nil, sql, parameters...)
	records = g.mapRecords(tableRealName, records)
	g.converter.ConvertRecords(table, columnNames, &records)
	return records
}

// Should check error
func (g *GenericDB) SelectCount(table *ReflectedTable, condition interface{ Condition }) int {
	tableName := table.GetName()
	tableRealName := table.GetRealName()
	condition = g.addMiddlewareConditions(tableName, condition)
	parameters := []interface{}{}
	whereClause := g.conditions.GetWhereClause(condition, &parameters)
	quote := g.getQuote()
	sql := fmt.Sprintf("SELECT COUNT(*) as c FROM %s%s%s %s", quote, tableRealName, quote, whereClause)
	stmt, _ := g.queryRowSingleColumn(nil, sql, parameters...)
	switch ct := stmt.(type) {
	case int:
		return ct
	case int64:
		return int(ct)
	case string:
		if i, err := strconv.Atoi(ct); err == nil {
			return i
		}
	case []byte:
		if i, err := strconv.Atoi(string(ct)); err == nil {
			return i
		}
	}
	log.Printf("Error processing count return value : %v of type % T from table %s\n", stmt, stmt, tableName)
	return 0
}

func (g *GenericDB) mapRecords(tableRealName string, records []map[string]interface{}) []map[string]interface{} {
	mappedRecords := []map[string]interface{}{}
	for _, record := range records {
		mappedRecord := map[string]interface{}{}
		for columRealName, columnValue := range record {
			mappedRecord[g.mapper.GetColumnName(tableRealName, columRealName)] = columnValue
		}
		mappedRecords = append(mappedRecords, mappedRecord)
	}
	return mappedRecords
}

// Should check error
func (g *GenericDB) SelectAll(table *ReflectedTable, columnNames []string, condition interface{ Condition }, columnOrdering [][2]string, offset, limit int) []map[string]interface{} {
	if limit == 0 {
		return []map[string]interface{}{}
	}
	selectColumns := g.columns.GetSelect(table, columnNames)
	tableName := table.GetName()
	tableRealName := table.GetRealName()
	condition = g.addMiddlewareConditions(tableName, condition)
	parameters := []interface{}{}
	whereClause := g.conditions.GetWhereClause(condition, &parameters)
	orderBy := g.columns.GetOrderBy(table, columnOrdering)
	offsetLimit := g.columns.GetOffsetLimit(offset, limit)
	quote := g.getQuote()
	sql := fmt.Sprintf("SELECT %s FROM %s%s%s %s %s %s", selectColumns, quote, tableRealName, quote, whereClause, orderBy, offsetLimit)
	records, _ := g.query(nil, sql, parameters...)
	records = g.mapRecords(tableRealName, records)
	g.converter.ConvertRecords(table, columnNames, &records)
	return records
}

func (g *GenericDB) UpdateSingle(tx *sql.Tx, table *ReflectedTable, columnValues map[string]interface{}, id string) (int64, error) {
	if len(columnValues) <= 0 {
		return 0, nil
	}
	g.converter.ConvertColumnValues(table, &columnValues)
	updateColumns, parameters := g.columns.GetUpdate(table, columnValues)
	tableName := table.GetName()
	tableRealName := table.GetRealName()
	var condition interface{ Condition }
	pk := table.GetPk()
	condition = NewColumnCondition(pk, `eq`, id)
	condition = g.addMiddlewareConditions(tableName, condition)
	whereClause := g.conditions.GetWhereClause(condition, &parameters)
	quote := g.getQuote()
	sql := fmt.Sprintf("UPDATE %s%s%s SET %s %s", quote, tableRealName, quote, updateColumns, whereClause)
	res, err := g.exec(tx, sql, parameters...)
	if err == nil {
		count, err := res.RowsAffected()
		if err != nil {
			return 0, err
		} else {
			return count, nil
		}
	} else {
		return 0, err
	}
}

func (g *GenericDB) DeleteSingle(tx *sql.Tx, table *ReflectedTable, id string) (int64, error) {
	tableName := table.GetName()
	tableRealName := table.GetRealName()
	var condition interface{ Condition }
	pk := table.GetPk()
	condition = NewColumnCondition(pk, `eq`, id)
	condition = g.addMiddlewareConditions(tableName, condition)
	parameters := []interface{}{}
	whereClause := g.conditions.GetWhereClause(condition, &parameters)
	quote := g.getQuote()
	sql := fmt.Sprintf("DELETE FROM %s%s%s %s", quote, tableRealName, quote, whereClause)
	res, err := g.exec(tx, sql, parameters...)
	if err == nil {
		count, err := res.RowsAffected()
		if err != nil {
			return 0, err
		} else {
			return count, nil
		}
	} else {
		return 0, err
	}
}

func (g *GenericDB) IncrementSingle(tx *sql.Tx, table *ReflectedTable, columnValues map[string]interface{}, id string) (int64, error) {
	if len(columnValues) <= 0 {
		return 0, nil
	}
	g.converter.ConvertColumnValues(table, &columnValues)
	updateColumns, parameters := g.columns.GetIncrement(table, columnValues)
	if updateColumns == "" {
		return 0, nil
	}
	tableName := table.GetName()
	tableRealName := table.GetRealName()
	var condition interface{ Condition }
	pk := table.GetPk()
	condition = NewColumnCondition(pk, `eq`, id)
	condition = g.addMiddlewareConditions(tableName, condition)
	whereClause := g.conditions.GetWhereClause(condition, &parameters)
	quote := g.getQuote()
	sql := fmt.Sprintf("UPDATE %s%s%s SET %s %s", quote, tableRealName, quote, updateColumns, whereClause)
	res, err := g.exec(tx, sql, parameters...)
	if err == nil {
		count, err := res.RowsAffected()
		if err != nil {
			return 0, err
		} else {
			return count, nil
		}
	} else {
		return 0, err
	}
}

func (g *GenericDB) queryRowSingleColumn(tx *sql.Tx, sql string, parameters ...interface{}) (interface{}, error) {
	return g.pdo.QueryRowSingleColumn(tx, sql, parameters...)
}

func (g *GenericDB) query(tx *sql.Tx, sql string, parameters ...interface{}) ([]map[string]interface{}, error) {
	return g.pdo.Query(tx, sql, parameters...)
}

func (g *GenericDB) exec(tx *sql.Tx, sql string, parameters ...interface{}) (sql.Result, error) {
	return g.pdo.Exec(tx, sql, parameters...)
}

func (g *GenericDB) Ping() int {
	start := time.Now()
	stmt, err := g.pdo.connect().Prepare("SELECT 1")
	if err != nil {
		return -1
	}
	_, err = stmt.Exec()
	if err != nil {
		return -1
	}
	t := time.Now()
	elapsed := t.Sub(start)
	return int(elapsed.Milliseconds())
}

func (g *GenericDB) GetCacheKey() string {
	gMap, _ := json.Marshal(map[string]interface{}{
		"driver":   g.driver,
		"address":  g.address,
		"port":     g.port,
		"database": g.database,
		"tables":   g.tables,
		"mapping":  g.mapping,
		"username": g.username,
	})
	return fmt.Sprintf("%x", md5.Sum(gMap))
}

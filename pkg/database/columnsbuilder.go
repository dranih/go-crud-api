package database

import (
	"fmt"
	"strings"
)

type ColumnsBuilder struct {
	driver    string
	converter *ColumnConverter
}

func NewColumnsBuilder(driver string) *ColumnsBuilder {
	return &ColumnsBuilder{driver, NewColumnConverter(driver)}
}

func (cb *ColumnsBuilder) GetOffsetLimit(offset, limit int) string {
	if limit < 0 || offset < 0 {
		return ``
	}
	switch cb.driver {
	case "mysql":
		return " LIMIT " + fmt.Sprint(offset) + ", " + fmt.Sprint(limit)
	case "pgsql":
		return " LIMIT " + fmt.Sprint(limit) + " OFFSET " + fmt.Sprint(offset)
	case "sqlsrv":
		return " OFFSET " + fmt.Sprint(offset) + " ROWS FETCH NEXT " + fmt.Sprint(limit) + " ROWS ONLY"
	case "sqlite":
		return " LIMIT " + fmt.Sprint(limit) + " OFFSET " + fmt.Sprint(offset)
	default:
		return ``
	}
}

func (cb *ColumnsBuilder) quoteColumnName(column *ReflectedColumn) string {
	return `"` + column.GetName() + `"`
}

func (cb *ColumnsBuilder) GetOrderBy(table *ReflectedTable, columnOrdering [][2]string) string {
	if len(columnOrdering) == 0 {
		return ``
	}
	results := []string{}
	for _, val := range columnOrdering {
		column := table.GetColumn(val[0])
		quotedColumnName := cb.quoteColumnName(column)
		results = append(results, quotedColumnName+` `+val[1])
	}
	return ` ORDER BY ` + strings.Join(results, `,`)
}

// done
func (cb *ColumnsBuilder) GetSelect(table *ReflectedTable, columnNames []string) string {
	results := []string{}
	for _, columnName := range columnNames {
		column := table.GetColumn(columnName)
		quotedColumnName := cb.quoteColumnName(column)
		quotedColumnName = cb.converter.ConvertColumnName(column, quotedColumnName)
		results = append(results, quotedColumnName)
	}
	return strings.Join(results, ",")
}

/*

public function getInsert(ReflectedTable $table, array $columnValues): string
{
	$columns = array();
	$values = array();
	foreach ($columnValues as $columnName => $columnValue) {
		$column = $table->getColumn($columnName);
		$quotedColumnName = $this->quoteColumnName($column);
		$columns[] = $quotedColumnName;
		$columnValue = $this->converter->convertColumnValue($column);
		$values[] = $columnValue;
	}
	$columnsSql = '(' . implode(',', $columns) . ')';
	$valuesSql = '(' . implode(',', $values) . ')';
	$outputColumn = $this->quoteColumnName($table->getPk());
	switch ($this->driver) {
		case 'mysql':
			return "$columnsSql VALUES $valuesSql";
		case 'pgsql':
			return "$columnsSql VALUES $valuesSql RETURNING $outputColumn";
		case 'sqlsrv':
			return "$columnsSql OUTPUT INSERTED.$outputColumn VALUES $valuesSql";
		case 'sqlite':
			return "$columnsSql VALUES $valuesSql";
	}
}

public function getUpdate(ReflectedTable $table, array $columnValues): string
{
	$results = array();
	foreach ($columnValues as $columnName => $columnValue) {
		$column = $table->getColumn($columnName);
		$quotedColumnName = $this->quoteColumnName($column);
		$columnValue = $this->converter->convertColumnValue($column);
		$results[] = $quotedColumnName . '=' . $columnValue;
	}
	return implode(',', $results);
}

public function getIncrement(ReflectedTable $table, array $columnValues): string
{
	$results = array();
	foreach ($columnValues as $columnName => $columnValue) {
		if (!is_numeric($columnValue)) {
			continue;
		}
		$column = $table->getColumn($columnName);
		$quotedColumnName = $this->quoteColumnName($column);
		$columnValue = $this->converter->convertColumnValue($column);
		$results[] = $quotedColumnName . '=' . $quotedColumnName . '+' . $columnValue;
	}
	return implode(',', $results);
}
*/

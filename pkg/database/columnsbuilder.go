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

// GetInsert return the insert request and the parameters to ensure to preserver column names and parameters order
func (cb *ColumnsBuilder) GetInsert(table *ReflectedTable, columnValues map[string]interface{}) (string, []interface{}) {
	columns := []string{}
	values := []string{}
	parameters := []interface{}{}
	for columnName, val := range columnValues {
		column := table.GetColumn(columnName)
		quotedColumnName := cb.quoteColumnName(column)
		quotedColumnName = cb.converter.ConvertColumnName(column, quotedColumnName)
		columns = append(columns, quotedColumnName)
		columnValue := cb.converter.ConvertColumnValue(column)
		values = append(values, columnValue)
		parameters = append(parameters, val)
	}
	columnsSql := `(` + strings.Join(columns, ",") + `)`
	valuesSql := `(` + strings.Join(values, ",") + `)`
	outputColumn := cb.quoteColumnName(table.GetPk())
	switch cb.driver {
	case `mysql`:
		return fmt.Sprintf("%s VALUES %s", columnsSql, valuesSql), parameters
	case `pgsql`:
		return fmt.Sprintf("%s VALUES %s RETURNING %s", columnsSql, valuesSql, outputColumn), parameters
	case `sqlsrv`:
		return fmt.Sprintf("%s OUTPUT INSERTED.%s VALUES %s", columnsSql, outputColumn, valuesSql), parameters
	case `sqlite`:
		return fmt.Sprintf("%s VALUES %s", columnsSql, valuesSql), parameters
	default:
		return "SELECT 1", nil
	}
}

func (cb *ColumnsBuilder) GetUpdate(table *ReflectedTable, columnValues map[string]interface{}) (string, []interface{}) {
	results := []string{}
	parameters := []interface{}{}
	for columnName, val := range columnValues {
		column := table.GetColumn(columnName)
		quotedColumnName := cb.quoteColumnName(column)
		columnValue := cb.converter.ConvertColumnValue(column)
		results = append(results, quotedColumnName+"="+columnValue)
		parameters = append(parameters, val)
	}
	return strings.Join(results, ","), parameters
}

/*
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

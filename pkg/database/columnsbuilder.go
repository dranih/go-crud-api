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
	switch cb.driver {
	case "mysql":
		return "`" + column.GetRealName() + "`"
	default:
		return `"` + column.GetRealName() + `"`
	}
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
		columns = append(columns, quotedColumnName)
		columnValue := cb.converter.ConvertColumnValue(column, parameters)
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
		columnValue := cb.converter.ConvertColumnValue(column, parameters)
		results = append(results, quotedColumnName+"="+columnValue)
		parameters = append(parameters, val)
	}
	return strings.Join(results, ","), parameters
}

func (cb *ColumnsBuilder) GetIncrement(table *ReflectedTable, columnValues map[string]interface{}) (string, []interface{}) {
	results := []string{}
	parameters := []interface{}{}
	for columnName, val := range columnValues {
		switch val.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		case float32, float64, complex64, complex128:
			column := table.GetColumn(columnName)
			quotedColumnName := cb.quoteColumnName(column)
			columnValue := cb.converter.ConvertColumnValue(column, parameters)
			results = append(results, quotedColumnName+"="+quotedColumnName+"+"+columnValue)
			parameters = append(parameters, val)
		}
	}
	return strings.Join(results, ","), parameters
}

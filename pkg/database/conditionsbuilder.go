package database

import (
	"fmt"
	"log"
	"strings"
)

type ConditionsBuilder struct {
	driver string
}

func NewConditionsBuilder(driver string) *ConditionsBuilder {
	return &ConditionsBuilder{driver}
}

func (cb *ConditionsBuilder) getConditionSql(condition interface{ Condition }, arguments *[]interface{}) string {
	switch v := condition.(type) {
	case *AndCondition:
		return cb.getAndConditionSql(condition.(*AndCondition), arguments)
	case *OrCondition:
		return cb.getOrConditionSql(condition.(*OrCondition), arguments)
	case *NotCondition:
		return cb.getNotConditionSql(condition.(*NotCondition), arguments)
	case *SpatialCondition:
		return cb.getSpatialConditionSql(condition.(*SpatialCondition), arguments)
	case *ColumnCondition:
		return cb.getColumnConditionSql(condition.(*ColumnCondition), arguments)
	default:
		log.Panicf("Unknown Condition: %T\n", v)
	}
	return ""
}

func (cb *ConditionsBuilder) getAndConditionSql(and *AndCondition, arguments *[]interface{}) string {
	parts := []string{}
	for _, condition := range and.GetConditions() {
		parts = append(parts, cb.getConditionSql(condition, arguments))
	}
	return "(" + strings.Join(parts, " AND ") + ")"
}

func (cb *ConditionsBuilder) getOrConditionSql(or *OrCondition, arguments *[]interface{}) string {
	parts := []string{}
	for _, condition := range or.GetConditions() {
		parts = append(parts, cb.getConditionSql(condition, arguments))
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

func (cb *ConditionsBuilder) getNotConditionSql(not *NotCondition, arguments *[]interface{}) string {
	condition := not.GetCondition()
	return "(NOT " + cb.getConditionSql(condition, arguments) + ")"
}

func (cb *ConditionsBuilder) quoteColumnName(column *ReflectedColumn) string {
	switch cb.driver {
	case "mysql":
		return "`" + column.GetName() + "`"
	default:
		return `"` + column.GetName() + `"`
	}
}

func (cb *ConditionsBuilder) escapeLikeValue(value string) string {
	return cb.addcslashes(value, "%_")
}

// From https://www.php2golang.com/method/function.addcslashes.html
// Addcslashes - Quote string with slashes in a C style
func (cb *ConditionsBuilder) addcslashes(s string, c string) string {
	var tmpRune []rune
	strRune := []rune(s)
	list := []rune(c)
	for _, ch := range strRune {
		for _, v := range list {
			if ch == v {
				tmpRune = append(tmpRune, '\\')
			}
		}
		tmpRune = append(tmpRune, ch)
	}
	return string(tmpRune)
}

func (cb *ConditionsBuilder) getColumnConditionSql(condition *ColumnCondition, arguments *[]interface{}) string {
	column := cb.quoteColumnName(condition.GetColumn())
	operator := condition.GetOperator()
	value := condition.GetValue()
	sql := "FALSE"
	switch operator {
	case `cs`:
		*arguments = append(*arguments, `%`+cb.escapeLikeValue(value)+`%`)
		sql = column + ` LIKE ` + cb.stmtOperator(len(*arguments))
	case `sw`:
		*arguments = append(*arguments, cb.escapeLikeValue(value)+`%`)
		sql = column + ` LIKE ` + cb.stmtOperator(len(*arguments))
	case `ew`:
		*arguments = append(*arguments, `%`+cb.escapeLikeValue(value))
		sql = column + ` LIKE ` + cb.stmtOperator(len(*arguments))
	case `eq`:
		*arguments = append(*arguments, value)
		sql = column + ` = ` + cb.stmtOperator(len(*arguments))
	case `lt`:
		*arguments = append(*arguments, value)
		sql = column + ` < ` + cb.stmtOperator(len(*arguments))
	case `le`:
		*arguments = append(*arguments, value)
		sql = column + ` <= ` + cb.stmtOperator(len(*arguments))
	case `ge`:
		*arguments = append(*arguments, value)
		sql = column + ` >= ` + cb.stmtOperator(len(*arguments))
	case `gt`:
		*arguments = append(*arguments, value)
		sql = column + ` > ` + cb.stmtOperator(len(*arguments))
	case `bt`:
		parts := strings.SplitN(value, `,`, 2)
		count := len(parts)
		if count == 2 {
			*arguments = append(*arguments, parts[0], parts[1])
			//sql = `(` + column + ` >= ? AND ` + column + ` <= ?)`
			sql = fmt.Sprintf("(column >= %s AND column <= %s)", cb.stmtOperator(len(*arguments)-1), cb.stmtOperator(len(*arguments)))
		} else {
			sql = "FALSE"
		}
	case `in`:
		parts := strings.Split(value, `,`)
		count := len(parts)
		if count > 0 {
			qmarks := cb.stmtOperator(len(*arguments) + 1)
			if count > 1 {
				//qmarks = strings.Repeat(`,?`, count)
				for i := 1; i <= count; i++ {
					qmarks = fmt.Sprintf("%s,%s", qmarks, cb.stmtOperator(len(*arguments)+count))
				}
			}
			sql = column + ` IN ( ` + qmarks + ` )`
			s := make([]interface{}, len(parts))
			for i, v := range parts {
				s[i] = v
			}
			*arguments = append(*arguments, s...)
		} else {
			sql = "FALSE"
		}
	case `is`:
		sql = column + ` IS NULL`
	}

	return sql
}

// stmtOperator returns the operator for the prepared statement : $x for psql else %
func (cb *ConditionsBuilder) stmtOperator(pos int) string {
	switch cb.driver {
	case "pgsql":
		return fmt.Sprintf("$%d", pos)
	case "sqlsrv":
		return fmt.Sprintf("@p%d", pos)
	default:
		return `?`
	}
}

func (cb *ConditionsBuilder) getSpatialFunctionName(operator string) string {
	switch operator {
	case `co`:
		return `ST_Contains`
	case `cr`:
		return `ST_Crosses`
	case `di`:
		return `ST_Disjoint`
	case `eq`:
		return `ST_Equals`
	case `in`:
		return `ST_Intersects`
	case `ov`:
		return `ST_Overlaps`
	case `to`:
		return `ST_Touches`
	case `wi`:
		return `ST_Within`
	case `ic`:
		return `ST_IsClosed`
	case `is`:
		return `ST_IsSimple`
	case `iv`:
		return `ST_IsValid`
	}
	return `FALSE`
}

func (cb *ConditionsBuilder) hasSpatialArgument(operator string) bool {
	return map[string]bool{`ic`: true, `is`: true, `iv`: true}[operator]
}

func (cb *ConditionsBuilder) getSpatialFunctionCall(functionName, column string, hasArgument bool, arguments *[]interface{}) string {
	argument := ""
	switch cb.driver {
	case `mysql`, `pgsql`:
		if hasArgument {
			argument = fmt.Sprintf("ST_GeomFromText(%s)", cb.stmtOperator(len(*arguments)+1))
		} else {
			argument = ``
		}
		return functionName + "(" + column + "," + argument + ")=TRUE"
	case `sqlsrv`:
		functionName = strings.Replace(functionName, `ST_`, `ST`, -1)
		if hasArgument {
			argument = fmt.Sprintf("geometry::STGeomFromText(%s,0)", cb.stmtOperator(len(*arguments)+1))
		} else {
			argument = ``
		}
		return column + "." + functionName + "(" + argument + ")=1"
	case `sqlite`:
		if hasArgument {
			argument = `?`
		} else {
			argument = `0`
		}
		return functionName + "(" + column + "," + argument + ")=1"
	}
	return ""
}

func (cb *ConditionsBuilder) getSpatialConditionSql(condition *SpatialCondition, arguments *[]interface{}) string {
	column := cb.quoteColumnName(condition.GetColumn())
	operator := condition.GetOperator()
	value := condition.GetValue()
	functionName := cb.getSpatialFunctionName(operator)
	hasArgument := cb.hasSpatialArgument(operator)
	sql := cb.getSpatialFunctionCall(functionName, column, hasArgument, arguments)
	if hasArgument {
		*arguments = append(*arguments, value)
	}
	return sql
}

func (cb *ConditionsBuilder) GetWhereClause(condition interface{ Condition }, arguments *[]interface{}) string {
	switch condition.(type) {
	case *NoCondition:
		return ``
	default:
		return ` WHERE ` + cb.getConditionSql(condition, arguments)
	}
}

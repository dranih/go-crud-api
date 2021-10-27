package record

import (
	"log"
	"strings"

	"github.com/dranih/go-crud-api/pkg/database"
)

type ConditionsBuilder struct {
	driver string
}

func NewConditionsBuilder(driver string) *ConditionsBuilder {
	return &ConditionsBuilder{driver}
}

func (cb *ConditionsBuilder) getConditionSql(condition interface{ Condition }, arguments *[]string) string {
	switch v := condition.(type) {
	case *AndCondition:
		return cb.getAndConditionSql(condition.(*AndCondition), arguments)
	case *OrCondition:
		return cb.getOrConditionSql(condition.(*OrCondition), arguments)
	case *NotCondition:
		return cb.getNotConditionSql(condition.(*NotCondition), arguments)
	/*case *SpatialCondition:
	return cb.getSpatialConditionSql(condition.(*SpatialCondition), arguments)*/
	case *ColumnCondition:
		return cb.getColumnConditionSql(condition.(*ColumnCondition), arguments)
	default:
		log.Panicf("Unknown Condition: %s\n", v)
	}
	return ""
}

func (cb *ConditionsBuilder) getAndConditionSql(and *AndCondition, arguments *[]string) string {
	parts := []string{}
	for _, condition := range and.GetConditions() {
		parts = append(parts, cb.getConditionSql(condition, arguments))
	}
	return "(" + strings.Join(parts, " AND ") + ")"
}

func (cb *ConditionsBuilder) getOrConditionSql(or *OrCondition, arguments *[]string) string {
	parts := []string{}
	for _, condition := range or.GetConditions() {
		parts = append(parts, cb.getConditionSql(condition, arguments))
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

func (cb *ConditionsBuilder) getNotConditionSql(not *NotCondition, arguments *[]string) string {
	condition := not.GetCondition()
	return "(NOT " + cb.getConditionSql(condition, arguments) + ")"
}

func (cb *ConditionsBuilder) quoteColumnName(column *database.ReflectedColumn) string {
	return `"` + column.GetName() + `"`
}

func (cb *ConditionsBuilder) escapeLikeValue(value string) string {
	return cb.Addcslashes(value, "%_")
}

// From https://www.php2golang.com/method/function.addcslashes.html
// Addcslashes - Quote string with slashes in a C style
func (cb *ConditionsBuilder) Addcslashes(s string, c string) string {
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

func (cb *ConditionsBuilder) getColumnConditionSql(condition *ColumnCondition, arguments *[]string) string {
	column := cb.quoteColumnName(condition.GetColumn())
	operator := condition.GetOperator()
	value := condition.GetValue()
	sql := "FALSE"
	switch operator {
	case `cs`:
		sql = column + ` LIKE ?`
		*arguments = append(*arguments, `%`+cb.escapeLikeValue(value)+`%`)
	case `sw`:
		sql = column + ` LIKE ?`
		*arguments = append(*arguments, cb.escapeLikeValue(value)+`%`)
	case `ew`:
		sql = column + ` LIKE ?`
		*arguments = append(*arguments, `%`+cb.escapeLikeValue(value))
	case `eq`:
		sql = column + ` = ?`
		*arguments = append(*arguments, value)
	case `lt`:
		sql = column + ` < ?`
		*arguments = append(*arguments, value)
	case `le`:
		sql = column + ` <= ?`
		*arguments = append(*arguments, value)
	case `ge`:
		sql = column + ` >= ?`
		*arguments = append(*arguments, value)
	case `gt`:
		sql = column + ` > ?`
		*arguments = append(*arguments, value)
	case `bt`:
		parts := strings.SplitN(value, `,`, 2)
		count := len(parts)
		if count == 2 {
			sql = `(` + column + ` >= ? AND ` + column + ` <= ?)`
			*arguments = append(*arguments, parts[0], parts[1])
		} else {
			sql = "FALSE"
		}
	case `in`:
		parts := strings.Split(value, `,`)
		count := len(parts)
		if count > 0 {
			qmarks := `?`
			if count > 1 {
				qmarks = strings.Repeat(`,?`, count)
			}
			sql = column + ` IN ( ` + qmarks + ` )`
			*arguments = append(*arguments, parts...)
		} else {
			sql = "FALSE"
		}
	case `is`:
		sql = column + ` IS NULL`
	}

	return sql
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

/*
   private function getSpatialFunctionCall(string $functionName, string $column, bool $hasArgument): string
   {
       switch ($this->driver) {
           case `mysql`:
           case `pgsql`:
               $argument = $hasArgument ? `ST_GeomFromText(?)` : ``;
               return "$functionName($column, $argument)=TRUE";
           case `sqlsrv`:
               $functionName = str_replace(`ST_`, `ST`, $functionName);
               $argument = $hasArgument ? `geometry::STGeomFromText(?,0)` : ``;
               return "$column.$functionName($argument)=1";
           case `sqlite`:
               $argument = $hasArgument ? `?` : `0`;
               return "$functionName($column, $argument)=1";
       }
   }

   private function getSpatialConditionSql(ColumnCondition $condition, array &$arguments): string
   {
       $column = $this->quoteColumnName($condition->getColumn());
       $operator = $condition->getOperator();
       $value = $condition->getValue();
       $functionName = $this->getSpatialFunctionName($operator);
       $hasArgument = $this->hasSpatialArgument($operator);
       $sql = $this->getSpatialFunctionCall($functionName, $column, $hasArgument);
       if ($hasArgument) {
           $arguments[] = $value;
       }
       return $sql;
   }
*/
func (cb *ConditionsBuilder) GetWhereClause(condition interface{ Condition }, arguments *[]string) string {
	switch condition.(type) {
	case *NoCondition:
		return ``
	default:
		return ` WHERE ` + cb.getConditionSql(condition, arguments)
	}
}

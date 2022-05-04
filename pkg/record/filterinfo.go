package record

import (
	"regexp"

	"github.com/dranih/go-crud-api/pkg/database"
)

type FilterInfo struct {
}

func (ft *FilterInfo) getConditionsAsPathTree(table *database.ReflectedTable, params map[string][]string) *PathTree {
	conditions := NewPathTree(nil)
	for key, filters := range params {
		if len(key) >= 6 && key[0:6] == `filter` {
			re := regexp.MustCompile(`\[.*\]`)
			key = re.ReplaceAllString(key, "")
			re = regexp.MustCompile(`\d+|\D+`)
			matches := re.FindAllString(key[6:], -1)
			path := []string{"filter"}
			if len(matches) > 0 {
				path = append(path, matches[0])
			}
			for _, filter := range filters {
				condition := database.ConditionFromString(table, filter)
				switch condition.(type) {
				case *database.NoCondition:
					continue
				default:
					conditions.Put(path, condition)
				}
			}
		}
	}
	return conditions
}

func (ft *FilterInfo) combinePathTreeOfConditions(tree *PathTree) interface{ database.Condition } {
	andConditions := tree.tree.GetValues()
	and := database.AndConditionFromArray(andConditions)
	orConditions := []interface{ database.Condition }{}
	for _, p := range tree.tree.GetKeys() {
		if pt := tree.tree.Get(p); pt.tree != nil {
			orConditions = append(orConditions, ft.combinePathTreeOfConditions(pt))
		}
	}
	or := database.OrConditionFromArray(orConditions)
	cond := and.And(or)
	return cond.(interface{ database.Condition })
}

func (ft *FilterInfo) GetCombinedConditions(table *database.ReflectedTable, params map[string][]string) interface{ database.Condition } {
	return ft.combinePathTreeOfConditions(ft.getConditionsAsPathTree(table, params))
}

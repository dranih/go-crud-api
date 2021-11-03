package database

import (
	"regexp"
)

type FilterInfo struct {
}

func (ft *FilterInfo) getConditionsAsPathTree(table *ReflectedTable, params map[string][]string) *PathTree {
	conditions := NewPathTree(nil)
	for key, filters := range params {
		if len(key) >= 6 && key[0:6] == `filter` {
			re := regexp.MustCompile(`\d+|\D+`)
			matches := re.FindAllString(key[6:], -1)
			match := ``
			if len(matches) > 0 {
				match = matches[0]
			}
			path := []string{match}
			for _, filter := range filters {
				condition := ConditionFromString(table, filter)
				switch condition.(type) {
				case *NoCondition:
					continue
				default:
					conditions.Put(path, condition)
				}
			}
		}
	}
	return conditions
}

func (ft *FilterInfo) combinePathTreeOfConditions(tree *PathTree) interface{ Condition } {
	andConditions := tree.tree.GetValues()
	and := AndConditionFromArray(andConditions)
	orConditions := []interface{ Condition }{}
	for _, p := range tree.tree.GetKeys() {
		orConditions = append(orConditions, ft.combinePathTreeOfConditions(tree.tree.Get(p)))
	}
	or := OrConditionFromArray(orConditions)
	cond := and.And(or)
	return cond
}

func (ft *FilterInfo) GetCombinedConditions(table *ReflectedTable, params map[string][]string) interface{ Condition } {
	return ft.combinePathTreeOfConditions(ft.getConditionsAsPathTree(table, params))
}

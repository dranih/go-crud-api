package database

import (
	"log"
	"regexp"
)

type FilterInfo struct {
}

func (ft *FilterInfo) getConditionsAsPathTree(table *ReflectedTable, params map[string][]string) *PathTree {
	conditions := NewPathTree(nil)
	for key, filters := range params {
		log.Printf("key : %v\n", key)
		log.Printf("filters : %v\n", filters)
		if key[0:6] == `filter` {
			re := regexp.MustCompile(`\d+|\D+`)
			matches := re.FindAllString(key[6:], -1)
			match := ``
			if len(matches) > 0 {
				match = matches[0]
			}
			path := []string{match}
			for _, filter := range filters {
				log.Printf("table : %v\n", table)
				log.Printf("filter : %v\n", filter)
				condition := GenericConditionFromString(table, filter)
				log.Printf("condition : %v\n", condition)
				switch condition.(type) {
				case *NoCondition:
					continue
				default:
					conditions.Put(path, condition)
				}
			}
			log.Println(conditions)
		}
	}
	return conditions
}

func (ft *FilterInfo) combinePathTreeOfConditions(tree *PathTree) interface{ Condition } {
	andConditions := tree.GetValues()
	and := AndConditionFromArray(andConditions)
	orConditions := []interface{ Condition }{}
	for _, p := range tree.GetKeys() {
		orConditions = append(orConditions, ft.combinePathTreeOfConditions(tree.Get(p)))
	}
	or := OrConditionFromArray(orConditions)
	return and.And(or)
}

func (ft *FilterInfo) GetCombinedConditions(table *ReflectedTable, params map[string][]string) interface{ Condition } {
	return ft.combinePathTreeOfConditions(ft.getConditionsAsPathTree(table, params))
}

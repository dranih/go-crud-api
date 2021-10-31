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
				condition := ConditionFromString(table, filter)
				log.Printf("condition : %v\n", condition)
				switch condition.(type) {
				case *NoCondition:
					continue
				default:
					conditions.Put(path, condition)
				}
			}
			log.Printf("Conditions : %v \n", conditions)
		}
	}
	return conditions
}

func (ft *FilterInfo) combinePathTreeOfConditions(tree *PathTree) interface{ Condition } {
	log.Printf("*** Combine tree : %v\n", tree.tree)
	andConditions := tree.tree.GetValues()
	log.Printf("*** Combine andConditions : %v \n", andConditions)
	and := AndConditionFromArray(andConditions)
	log.Printf("*** Combine and : %v / %T\n", and, and)
	orConditions := []interface{ Condition }{}
	for _, p := range tree.tree.GetKeys() {
		orConditions = append(orConditions, ft.combinePathTreeOfConditions(tree.tree.Get(p)))
	}
	or := OrConditionFromArray(orConditions)
	log.Printf("*** Combine or : %v\n", or)
	log.Printf("*** Combine type or : %T\n", or)
	log.Printf("*** Combine type and : %T\n", and)
	cond := and.And(or)
	log.Printf("*** Combine cond : %v\n", cond)
	log.Printf("*** Combine type cond : %T\n", cond)
	return cond
}

func (ft *FilterInfo) GetCombinedConditions(table *ReflectedTable, params map[string][]string) interface{ Condition } {
	return ft.combinePathTreeOfConditions(ft.getConditionsAsPathTree(table, params))
}

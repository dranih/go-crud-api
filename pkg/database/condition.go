package database

import (
	"log"
	"strings"
)

// Condition interface
type Condition interface {
	GetCondition() interface{ Condition }
	And(condition interface{ Condition }) interface{ Condition }
	Or(condition interface{ Condition }) interface{ Condition }
	Not() interface{ Condition }
}

type GenericCondition struct {
	condition interface{ Condition }
}

func (gc *GenericCondition) And(condition interface{ Condition }) interface{ Condition } {
	log.Println("Coucou")
	log.Printf("AAA gc : %v\n", gc.condition)
	log.Printf("AAA gc type : %T\n", gc.condition)
	log.Printf("AAA gc : %v\n", gc)
	log.Printf("AAA gc type : %T\n", gc)
	switch condition.(type) {
	case *NoCondition:
		log.Println("Coucou1")
		return gc.condition
	default:
		log.Println("Coucou2")
		return NewAndCondition(gc.condition, condition)
	}
}

func (gc *GenericCondition) Or(condition interface{ Condition }) interface{ Condition } {
	switch condition.(type) {
	case *NoCondition:
		return condition
	default:
		return NewOrCondition(gc.condition, condition)
	}
}

func (gc *GenericCondition) Not() interface{ Condition } {
	return NewNotCondition(gc.condition)
}

func (gc *GenericCondition) GetCondition() interface{ Condition } {
	return nil
}

func ConditionFromString(table *ReflectedTable, value string) interface{ Condition } {
	var condition interface{ Condition }
	condition = NewNoCondition()
	parts := strings.SplitN(value, ",", 3)
	log.Printf("Parts : %v\n", parts)
	if (len(parts)) < 2 {
		return condition
	}
	if len(parts) < 3 {
		parts = append(parts, "")
	}
	field := table.GetColumn(parts[0])
	log.Printf("field : %v\n", field)
	command := parts[1]
	negate := false
	spatial := false
	if len(command) > 2 {
		if command[0:1] == "n" {
			negate = true
			command = command[1:]
		} else if command[0:1] == "n" {
			spatial = true
			command = command[1:]
		}
	}
	log.Printf("command : %v\n", command)
	if spatial {
		if map[string]bool{"co": true, "cr": true, "di": true, "eq": true, "in": true, "ov": true, "to": true, "wi": true, "ic": true, "is": true, "iv": true}[command] {
			condition = NewSpatialCondition(field, command, parts[2])
		}
	} else {
		if map[string]bool{"cs": true, "sw": true, "ew": true, "eq": true, "lt": true, "le": true, "ge": true, "gt": true, "bt": true, "in": true, "is": true}[command] {
		}
		condition = NewColumnCondition(field, command, parts[2])
	}
	if negate {
		condition = condition.Not()
	}
	log.Printf("#################### from string : %v // %T", condition, condition)
	return condition
}

// end Condition interface

// NoCondition struct
type NoCondition struct {
}

func NewNoCondition() *NoCondition {
	return &NoCondition{}

}

func (nc *NoCondition) GetCondition() interface{ Condition } {
	return nil
}

func (nc *NoCondition) And(condition interface{ Condition }) interface{ Condition } {
	return condition
}

func (nc *NoCondition) Or(condition interface{ Condition }) interface{ Condition } {
	return condition
}

func (nc *NoCondition) Not() interface{ Condition } {
	return nc
}

// end NoCondition struct

// NotCondition struct
type NotCondition struct {
	condition interface{ Condition }
	GenericCondition
}

func NewNotCondition(condition interface{ Condition }) *NotCondition {
	return &NotCondition{condition, GenericCondition{condition}}
}

func (nc *NotCondition) GetCondition() interface{ Condition } {
	return nc.condition
}

// end NotCondition

// OrCondition struct
type OrCondition struct {
	conditions []interface{ Condition }
	GenericCondition
}

func NewOrCondition(condition1, condition2 interface{ Condition }) *OrCondition {
	return &OrCondition{[]interface{ Condition }{condition1, condition2}, GenericCondition{}}
}

func (oc *OrCondition) Or(condition interface{ Condition }) interface{ Condition } {
	switch condition.(type) {
	case *NoCondition:
		return condition
	default:
		oc.conditions = append(oc.conditions, condition)
		return oc
	}
}

func (ac *OrCondition) GetConditions() []interface{ Condition } {
	return ac.conditions
}

func OrConditionFromArray(conditions []interface{ Condition }) interface{ Condition } {
	var condition interface{ Condition }
	condition = NewNoCondition()
	for _, c := range conditions {
		condition = condition.Or(c)
	}
	return condition
}

// end OrCondition

// AndCondition struct
type AndCondition struct {
	conditions []interface{ Condition }
	GenericCondition
}

func NewAndCondition(condition1, condition2 interface{ Condition }) *AndCondition {
	return &AndCondition{[]interface{ Condition }{condition1, condition2}, GenericCondition{}}
}

func (ac *AndCondition) And(condition interface{ Condition }) interface{ Condition } {
	log.Println("33333333333333333333333")
	switch condition.(type) {
	case *NoCondition:
		return ac
	default:
		ac.conditions = append(ac.conditions, condition)
		return ac
	}
}

func (ac *AndCondition) GetConditions() []interface{ Condition } {
	return ac.conditions
}

func AndConditionFromArray(conditions []interface{ Condition }) interface{ Condition } {
	var condition interface{ Condition }
	condition = NewNoCondition()
	for _, c := range conditions {
		condition = condition.And(c)
	}
	return condition
}

// end AndCondition

// ColumnCondition struct
type ColumnCondition struct {
	column   *ReflectedColumn
	operator string
	value    string
	GenericCondition
}

// Ugly
func NewColumnCondition(column *ReflectedColumn, operator, value string) *ColumnCondition {
	condition := &ColumnCondition{column, operator, value, GenericCondition{}}
	condition.GenericCondition = GenericCondition{condition}
	//return &ColumnCondition{column, operator, value, GenericCondition{&NoCondition{}}}
	return condition
}

func (cc *ColumnCondition) GetColumn() *ReflectedColumn {
	return cc.column
}
func (cc *ColumnCondition) GetOperator() string {
	return cc.operator
}

func (cc *ColumnCondition) GetValue() string {
	return cc.value
}

// end ColumnCondition

// SpatialCondition struct
type SpatialCondition struct {
	ColumnCondition
}

//Ugly
func NewSpatialCondition(column *ReflectedColumn, operator, value string) *SpatialCondition {
	condition := &ColumnCondition{column, operator, value, GenericCondition{}}
	condition.GenericCondition = GenericCondition{condition}
	return &SpatialCondition{*condition}
}

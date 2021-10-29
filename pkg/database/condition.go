package database

import (
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
}

func (gc *GenericCondition) And(condition interface{ Condition }) interface{ Condition } {
	switch condition.(type) {
	case *NoCondition:
		return condition
	default:
		return NewAndCondition(gc, condition)
	}
}

func (gc *GenericCondition) Or(condition interface{ Condition }) interface{ Condition } {
	switch condition.(type) {
	case *NoCondition:
		return condition
	default:
		return NewOrCondition(gc, condition)
	}
}

func (gc *GenericCondition) Not() interface{ Condition } {
	return NewNotCondition(gc)
}

func (gc *GenericCondition) GetCondition() interface{ Condition } {
	return nil
}

func GenericConditionFromString(table *ReflectedTable, value string) interface{ Condition } {
	var condition interface{ Condition }
	condition = NewNoCondition()
	parts := strings.SplitN(value, ",", 3)
	if (len(parts)) < 2 {
		return condition
	}
	if len(parts) < 3 {
		parts = append(parts, "")
	}
	field := table.GetColumn(parts[0])
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
	if spatial {
		if map[string]bool{"co": true, "cr": true, "di": true, "eq": true, "in": true, "ov": true, "to": true, "wi": true, "ic": true, "is": true, "iv": true}[command] {
			condition = NewSpatialCondition(field, command, parts[2])
		} else {
			if map[string]bool{"cs": true, "sw": true, "ew": true, "eq": true, "lt": true, "le": true, "ge": true, "gt": true, "bt": true, "in": true, "is": true}[command] {
			}
			condition = NewColumnCondition(field, command, parts[2])
		}
	}
	if negate {
		condition = condition.Not()
	}
	return condition
}

// end Condition interface

// NoCondition struct
type NoCondition struct {
	GenericCondition
}

func NewNoCondition() *NoCondition {
	return &NoCondition{}
}

/*func (nc *NoCondition) GetCondition() interface{ Condition } {
	return nil
}*/

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
	return &NotCondition{condition, GenericCondition{}}
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

func NewOrCondition(condition1, condition2 interface{ Condition }) *AndCondition {
	return &AndCondition{[]interface{ Condition }{condition1, condition2}, GenericCondition{}}
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
	switch condition.(type) {
	case *NoCondition:
		return condition
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

func NewColumnCondition(column *ReflectedColumn, operator, value string) *ColumnCondition {
	return &ColumnCondition{column, operator, value, GenericCondition{}}
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

func NewSpatialCondition(column *ReflectedColumn, operator, value string) *SpatialCondition {
	return &SpatialCondition{ColumnCondition{column, operator, value, GenericCondition{}}}
}

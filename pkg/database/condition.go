package database

import (
	"strings"
)

// Condition interface
type Condition interface {
	GetCondition() interface{}
	And(condition interface{}) interface{}
	Or(condition interface{}) interface{}
	Not() interface{}
}

type GenericCondition struct {
	condition interface{ Condition }
}

func (gc *GenericCondition) And(condition interface{}) interface{} {
	switch c := condition.(type) {
	case *NoCondition:
		return gc.condition
	case interface{ Condition }:
		return NewAndCondition(gc.condition, c)
	}
	return nil
}

func (gc *GenericCondition) Or(condition interface{}) interface{} {
	switch c := condition.(type) {
	case *NoCondition:
		return gc.condition
	case interface{ Condition }:
		return NewOrCondition(gc.condition, c)
	}
	return nil
}

func (gc *GenericCondition) Not() interface{} {
	return NewNotCondition(gc.condition)
}

func (gc *GenericCondition) GetCondition() interface{} {
	return nil
}

func ConditionFromString(table *ReflectedTable, value string) interface{ Condition } {
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
	if field == nil {
		return condition
	}
	command := parts[1]
	negate := false
	spatial := false
	if len(command) > 2 {
		if command[0:1] == "n" {
			negate = true
			command = command[1:]
		} else if command[0:1] == "s" {
			spatial = true
			command = command[1:]
		}
	}
	if spatial {
		if map[string]bool{"co": true, "cr": true, "di": true, "eq": true, "in": true, "ov": true, "to": true, "wi": true, "ic": true, "is": true, "iv": true}[command] {
			condition = NewSpatialCondition(field, command, parts[2])
		}
	} else if map[string]bool{"cs": true, "sw": true, "ew": true, "eq": true, "lt": true, "le": true, "ge": true, "gt": true, "bt": true, "in": true, "is": true}[command] {
		condition = NewColumnCondition(field, command, parts[2])
	}
	if negate {
		condition = condition.Not().(interface{ Condition })
	}
	return condition
}

// end Condition interface

// NoCondition struct
type NoCondition struct {
}

func NewNoCondition() *NoCondition {
	return &NoCondition{}

}

func (nc *NoCondition) GetCondition() interface{} {
	return nil
}

func (nc *NoCondition) And(condition interface{}) interface{} {
	return condition
}

func (nc *NoCondition) Or(condition interface{}) interface{} {
	return condition
}

func (nc *NoCondition) Not() interface{} {
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

func (nc *NotCondition) And(condition interface{}) interface{} {
	switch c := condition.(type) {
	case *NoCondition:
		return nc
	case interface{ Condition }:
		return NewAndCondition(nc, c)
	}
	return nil
}

func (nc *NotCondition) Or(condition interface{}) interface{} {
	switch c := condition.(type) {
	case *NoCondition:
		return nc
	case interface{ Condition }:
		return NewOrCondition(nc, c)
	}
	return nil
}

func (nc *NotCondition) GetCondition() interface{} {
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

func (oc *OrCondition) Or(condition interface{}) interface{} {
	switch c := condition.(type) {
	case *NoCondition:
		return oc
	case interface{ Condition }:
		oc.conditions = append(oc.conditions, c)
		return oc
	}
	return nil
}

func (oc *OrCondition) And(condition interface{}) interface{} {
	switch c := condition.(type) {
	case *NoCondition:
		return oc
	case interface{ Condition }:
		return NewAndCondition(oc, c)
	}
	return nil
}

func (ac *OrCondition) GetConditions() []interface{ Condition } {
	return ac.conditions
}

func OrConditionFromArray(conditions []interface{ Condition }) interface{ Condition } {
	var condition interface{ Condition }
	condition = NewNoCondition()
	for _, c := range conditions {
		if ct, ok := condition.Or(c).(interface{ Condition }); ok {
			condition = ct
		}
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

func (ac *AndCondition) And(condition interface{}) interface{} {
	switch c := condition.(type) {
	case *NoCondition:
		return ac
	case interface{ Condition }:
		ac.conditions = append(ac.conditions, c)
		return ac
	}
	return nil
}

func (ac *AndCondition) Or(condition interface{}) interface{} {
	switch c := condition.(type) {
	case *NoCondition:
		return ac
	case interface{ Condition }:
		return NewOrCondition(ac, c)
	}
	return nil
}

func (ac *AndCondition) GetConditions() []interface{ Condition } {
	return ac.conditions
}

func AndConditionFromArray(conditions []interface{ Condition }) interface{ Condition } {
	var condition interface{ Condition }
	condition = NewNoCondition()
	for _, c := range conditions {
		if ct, ok := condition.And(c).(interface{ Condition }); ok {
			condition = ct
		}
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

func (sc *SpatialCondition) And(condition interface{}) interface{} {
	switch c := condition.(type) {
	case *NoCondition:
		return sc
	case interface{ Condition }:
		return NewAndCondition(sc, c)
	}
	return nil
}

func (sc *SpatialCondition) Or(condition interface{}) interface{} {
	switch c := condition.(type) {
	case *NoCondition:
		return sc
	case interface{ Condition }:
		return NewOrCondition(sc, c)
	}
	return nil
}

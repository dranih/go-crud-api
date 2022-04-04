package record

import (
	"strings"

	"github.com/dranih/go-crud-api/pkg/database"
)

type OrderingInfo struct{}

func NewOrderingInfo() *OrderingInfo {
	return &OrderingInfo{}
}

func (oi *OrderingInfo) GetColumnOrdering(table *database.ReflectedTable, params map[string][]string) [][2]string {
	fields := [][2]string{}
	if orders, exists := params["order"]; exists {
		for _, order := range orders {
			parts := strings.SplitN(order, ",", 3)
			columnName := parts[0]
			if !table.HasColumn(columnName) {
				continue
			}
			ascending := `ASC`
			if len(parts) > 1 {
				if len(parts[1]) == 4 && strings.ToUpper(parts[1])[:4] == `DESC` {
					ascending = `DESC`
				}
			}
			fields = append(fields, [2]string{columnName, ascending})
		}
	}
	if len(fields) == 0 {
		return oi.GetDefaultColumnOrdering(table)
	}
	return fields
}

func (oi *OrderingInfo) GetDefaultColumnOrdering(table *database.ReflectedTable) [][2]string {
	fields := [][2]string{}
	pk := table.GetPk()
	if pk != nil {
		fields = append(fields, [2]string{pk.GetName(), `ASC`})
	} else {
		for _, columnName := range table.GetColumnNames() {
			fields = append(fields, [2]string{columnName, `ASC`})
		}
	}
	return fields
}

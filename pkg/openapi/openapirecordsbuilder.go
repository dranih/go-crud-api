package openapi

import (
	"fmt"
	"math"
	"unicode"

	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/utils"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type OpenApiRecordsBuilder struct {
	openapi    *OpenApiDefinition
	operations map[string]string
	reflection *database.ReflectionService
	types      map[string]map[string]interface{}
}

func (oab *OpenApiBuilder) NewOpenApiRecordsBuilder(openapi *OpenApiDefinition, reflection *database.ReflectionService) *OpenApiRecordsBuilder {
	return &OpenApiRecordsBuilder{
		openapi,
		map[string]string{
			"list":      "get",
			"create":    "post",
			"read":      "get",
			"update":    "put",
			"delete":    "delete",
			"increment": "patch",
		},
		reflection,
		map[string]map[string]interface{}{
			"integer":   {"type": "integer", "format": "int32"},
			"bigint":    {"type": "integer", "format": "int64"},
			"varchar":   {"type": "string"},
			"clob":      {"type": "string", "format": "large-string"}, //custom format
			"varbinary": {"type": "string", "format": "byte"},
			"blob":      {"type": "string", "format": "large-byte"}, //custom format
			"decimal":   {"type": "string", "format": "decimal"},    //custom format
			"float":     {"type": "number", "format": "float"},
			"double":    {"type": "number", "format": "double"},
			"date":      {"type": "string", "format": "date"},
			"time":      {"type": "string", "format": "time"}, //custom format
			"timestamp": {"type": "string", "format": "date-time"},
			"geometry":  {"type": "string", "format": "geometry"}, //custom format
			"boolean":   {"type": "boolean"},
		},
	}
}

func (oarb *OpenApiRecordsBuilder) normalize(value string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	if result, _, err := transform.String(t, value); err != nil {
		//If normalization failed, return initial string
		return value
	} else {
		return result
	}
}

func (oarb *OpenApiRecordsBuilder) getAllTableReferences() map[string][]string {
	tableReferences := map[string][]string{}
	for _, tableName := range oarb.reflection.GetTableNames() {
		table := oarb.reflection.GetTable(tableName)
		for _, columnName := range table.GetColumnNames() {
			column := table.GetColumn(columnName)
			referencedTableName := column.GetFk()
			if referencedTableName != "" {
				if _, exists := tableReferences[referencedTableName]; !exists {
					tableReferences[referencedTableName] = []string{}
				}
				tableReferences[referencedTableName] = append(tableReferences[referencedTableName], fmt.Sprintf("%s.%s", tableName, columnName))
			}
		}
	}
	return tableReferences
}

func (oarb *OpenApiRecordsBuilder) Build() {
	tableNames := oarb.reflection.GetTableNames()
	for _, tableName := range tableNames {
		oarb.setPath(tableName)
	}
	oarb.openapi.Set("components|responses|pk_integer|description", "inserted primary key value (integer)")
	oarb.openapi.Set("components|responses|pk_integer|content|application/json|schema|type", "integer")
	oarb.openapi.Set("components|responses|pk_integer|content|application/json|schema|format", "int64")
	oarb.openapi.Set("components|responses|pk_string|description", "inserted primary key value (string)")
	oarb.openapi.Set("components|responses|pk_string|content|application/json|schema|type", "string")
	oarb.openapi.Set("components|responses|pk_string|content|application/json|schema|format", "uuid")
	oarb.openapi.Set("components|responses|rows_affected|description", "number of rows affected (integer)")
	oarb.openapi.Set("components|responses|rows_affected|content|application/json|schema|type", "integer")
	oarb.openapi.Set("components|responses|rows_affected|content|application/json|schema|format", "int64")
	tableReferences := oarb.getAllTableReferences()
	for _, tableName := range tableNames {
		var references []string
		if _references, exists := tableReferences[tableName]; exists {
			references = _references
		}
		oarb.setComponentSchema(tableName, references)
		oarb.setComponentResponse(tableName)
		oarb.setComponentRequestBody(tableName)
	}
	oarb.setComponentParameters()
	for index, tableName := range tableNames {
		oarb.setTag(index, tableName)
	}
}

//Should try to pass a func as a middleware property to see if this works
func (oarb *OpenApiRecordsBuilder) isOperationOnTableAllowed(operation, tableName string) bool {
	if tableHandler := utils.VStore.Get("authorization.tableHandler"); tableHandler == nil {
		return true
	} else {
		if tableHandlerFunc, ok := tableHandler.(func(string, string) bool); ok {
			return tableHandlerFunc(operation, tableName)
		} else {
			return true
		}
	}
}

func (oarb *OpenApiRecordsBuilder) isOperationOnColumnAllowed(operation, tableName, columnName string) bool {
	if columnHandler := utils.VStore.Get("authorization.columnHandler"); columnHandler == nil {
		return true
	} else {
		if columnHandlerFunc, ok := columnHandler.(func(string, string, string) bool); ok {
			return columnHandlerFunc(operation, tableName, columnName)
		} else {
			return true
		}
	}
}

func (oarb *OpenApiRecordsBuilder) setPath(tableName string) {
	normalizedTableName := oarb.normalize(tableName)
	table := oarb.reflection.GetTable(tableName)
	tableType := table.GetType()
	pk := table.GetPk()
	pkName := ""
	if pk != nil {
		pkName = pk.GetName()
	}
	for operation, method := range oarb.operations {
		if pkName == "" && operation != "list" {
			continue
		}
		if tableType != "table" && operation != "list" {
			continue
		}
		if !oarb.isOperationOnTableAllowed(operation, tableName) {
			continue
		}
		var parameters []string
		var path string
		if operation == "list" || operation == "create" {
			path = fmt.Sprintf("/records/%s", tableName)
			if operation == "list" {
				parameters = []string{"filter", "include", "exclude", "order", "size", "page", "join"}
			}
		} else {
			path = fmt.Sprintf("/records/%s/{id}", tableName)
			if operation == "read" {
				parameters = []string{"pk", "include", "exclude", "join"}
			} else {
				parameters = []string{"pk"}
			}
		}
		for p, parameter := range parameters {
			oarb.openapi.Set(fmt.Sprintf(`paths|%s|%s|parameters|%d|$ref`, path, method, p), fmt.Sprintf("#/components/parameters/%s", parameter))
		}
		if _, ok := map[string]bool{"create": true, "update": true, "increment": true}[operation]; ok {
			oarb.openapi.Set(fmt.Sprintf(`paths|%s|%s|requestBody|$ref`, path, method), fmt.Sprintf("#/components/requestBodies/%s-%s", operation, normalizedTableName))
		}
		oarb.openapi.Set(fmt.Sprintf("paths|%s|%s|tags|0", path, method), tableName)
		oarb.openapi.Set(fmt.Sprintf("paths|%s|%s|operationId", path, method), fmt.Sprintf("%s_%s", operation, normalizedTableName))
		oarb.openapi.Set(fmt.Sprintf("paths|%s|%s|description", path, method), fmt.Sprintf("%s %s", operation, tableName))
		switch operation {
		case "list":
			oarb.openapi.Set(fmt.Sprintf(`paths|%s|%s|responses|200|$ref`, path, method), fmt.Sprintf("/components/responses/%s-%s", operation, normalizedTableName))
		case "create":
			if pk.GetType() == "integer" {
				oarb.openapi.Set(fmt.Sprintf(`paths|%s|%s|responses|200|$ref`, path, method), "/components/responses/pk_integer")
			} else {
				oarb.openapi.Set(fmt.Sprintf(`paths|%s|%s|responses|200|$ref`, path, method), "/components/responses/pk_string")
			}
		case "read":
			oarb.openapi.Set(fmt.Sprintf(`paths|%s|%s|responses|200|$ref`, path, method), fmt.Sprintf("/components/responses/%s-%s", operation, normalizedTableName))
		case "update", "delete", "increment":
			oarb.openapi.Set(fmt.Sprintf(`paths|%s|%s|responses|200|$ref`, path, method), "/components/responses/rows_affected")
		}
	}
}

func (oarb *OpenApiRecordsBuilder) getPattern(column *database.ReflectedColumn) string {
	switch column.GetType() {
	case "integer":
		return "^-?[0-9]{1,10}$"
	case "bigint":
		return "^-?[0-9]{1,19}$"
	case "varchar":
		l := column.GetLength()
		return fmt.Sprintf("^.{0,%d}$", l)
	case "clob":
		return "^.*$"
	case "varbinary":
		l := column.GetLength()
		b := 4 * math.Ceil(float64(l)/3)
		return fmt.Sprintf("^[A-Za-z0-9+/]{0,%d}=*$", int(b))
	case "blob":
		return "^[A-Za-z0-9+/]*=*$"
	case "decimal":
		p := column.GetPrecision()
		s := column.GetScale()
		return fmt.Sprintf(`^-?[0-9]{1,%d}(\.[0-9]{1,%d})?$`, p-s, s)
	case "float":
		return `^-?[0-9]+(\.[0-9]+)?([eE]-?[0-9]+)?$`
	case "double":
		return `^-?[0-9]+(\.[0-9]+)?([eE]-?[0-9]+)?$`
	case "date":
		return "^[0-9]{4}-[0-9]{2}-[0-9]{2}$"
	case "time":
		return "^[0-9]{2}:[0-9]{2}:[0-9]{2}$"
	case "timestamp":
		return "^[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}$"
	case "geometry":
		return `^(POINT|LINESTRING|POLYGON|MULTIPOINT|MULTILINESTRING|MULTIPOLYGON)\s*\(.*$`
	case "boolean":
		return "^(true|false)$"
	}
	return ""
}

func (oarb *OpenApiRecordsBuilder) setComponentSchema(tableName string, references []string) {
	normalizedTableName := oarb.normalize(tableName)
	table := oarb.reflection.GetTable(tableName)
	tableType := table.GetType()
	pk := table.GetPk()
	pkName := ""
	if pk != nil {
		pkName = pk.GetName()
	}
	for operation := range oarb.operations {
		if pkName == "" && operation != "list" {
			continue
		}
		if tableType == "view" && (operation != "read" && operation != "list") {
			continue
		}
		if tableType == "view" && pkName == "" && operation == "read" {
			continue
		}
		if !oarb.isOperationOnTableAllowed(operation, tableName) {
			continue
		}

		if operation == "list" {
			var prefix string
			if operation == "list" {
				oarb.openapi.Set(fmt.Sprintf("components|schemas|%s-%s|type", operation, normalizedTableName), "object")
				oarb.openapi.Set(fmt.Sprintf("components|schemas|%s-%s|properties|results|type", operation, normalizedTableName), "integer")
				oarb.openapi.Set(fmt.Sprintf("components|schemas|%s-%s|properties|results|format", operation, normalizedTableName), "int64")
				oarb.openapi.Set(fmt.Sprintf("components|schemas|%s-%s|properties|records|type", operation, normalizedTableName), "array")
				prefix = fmt.Sprintf("components|schemas|%s-%s|properties|records|items", operation, normalizedTableName)
			} else {
				prefix = fmt.Sprintf("components|schemas|%s-%s", operation, normalizedTableName)
			}
			oarb.openapi.Set(fmt.Sprintf("%s|type", prefix), "object")
			for _, columnName := range table.GetColumnNames() {
				if !oarb.isOperationOnColumnAllowed(operation, tableName, columnName) {
					continue
				}
				column := table.GetColumn(columnName)
				properties := oarb.types[column.GetType()]
				if column.HasLength() {
					properties["maxLength"] = column.GetLength()
				} else {
					properties["maxLength"] = 0
				}
				properties["nullable"] = column.GetNullable()
				properties["pattern"] = oarb.getPattern(column)
				for key, value := range properties {
					if value != "" {
						oarb.openapi.Set(fmt.Sprintf("%s|properties|%s|%s", prefix, columnName, key), value)
					}
				}
				if column.GetPk() {
					oarb.openapi.Set(fmt.Sprintf("%s|properties|%s|x-primary-key", prefix, columnName), true)
					oarb.openapi.Set(fmt.Sprintf("%s|properties|%s|x-referenced", prefix, columnName), references)
				}
				if fk := column.GetFk(); fk != "" {
					oarb.openapi.Set(fmt.Sprintf("%s|properties|%s|x-references", prefix, columnName), fk)
				}
			}
		}
	}
}

func (oarb *OpenApiRecordsBuilder) setComponentResponse(tableName string) {
	normalizedTableName := oarb.normalize(tableName)
	table := oarb.reflection.GetTable(tableName)
	tableType := table.GetType()
	pk := table.GetPk()
	pkName := ""
	if pk != nil {
		pkName = pk.GetName()
	}
	for operation := range map[string]bool{"list": true, "read": true} {
		if pkName == "" && operation != "list" {
			continue
		}
		if tableType != "table" && operation != "list" {
			continue
		}
		if !oarb.isOperationOnTableAllowed(operation, tableName) {
			continue
		}

		if operation == "list" {
			if operation == "list" {
				oarb.openapi.Set(fmt.Sprintf("components|responses|%s-%s|description", operation, normalizedTableName), fmt.Sprintf("list of %s records", tableName))
			} else {
				oarb.openapi.Set(fmt.Sprintf("components|responses|%s-%s|description", operation, normalizedTableName), fmt.Sprintf("single %s record", tableName))
			}
			oarb.openapi.Set(fmt.Sprintf("components|responses|%s-%s|content|application/json|schema|$ref", operation, normalizedTableName), fmt.Sprintf("#/components/schemas/$%s-%s", operation, normalizedTableName))
		}
	}
}

func (oarb *OpenApiRecordsBuilder) setComponentRequestBody(tableName string) {
	normalizedTableName := oarb.normalize(tableName)
	table := oarb.reflection.GetTable(tableName)
	tableType := table.GetType()
	pk := table.GetPk()
	if pk != nil {
		if pkName := pk.GetName(); pkName != "" && tableType == "table" {
			for operation := range map[string]bool{"create": true, "update": true, "increment": true} {
				if !oarb.isOperationOnTableAllowed(operation, tableName) {
					continue
				}
				oarb.openapi.Set(fmt.Sprintf("components|requestBodies|%s-%s|description", operation, normalizedTableName), fmt.Sprintf("single %s record", tableName))
				oarb.openapi.Set(fmt.Sprintf("components|requestBodies|%s-%s|content|application/json|schema|$ref", operation, normalizedTableName), fmt.Sprintf("#/components/schemas/$%s-%s", operation, normalizedTableName))
			}
		}
	}
}

func (oarb OpenApiRecordsBuilder) setComponentParameters() {
	oarb.openapi.Set("components|parameters|pk|name", "id")
	oarb.openapi.Set("components|parameters|pk|in", "path")
	oarb.openapi.Set("components|parameters|pk|schema|type", "string")
	oarb.openapi.Set("components|parameters|pk|description", "primary key value")
	oarb.openapi.Set("components|parameters|pk|required", true)

	oarb.openapi.Set("components|parameters|filter|name", "filter")
	oarb.openapi.Set("components|parameters|filter|in", "query")
	oarb.openapi.Set("components|parameters|filter|schema|type", "array")
	oarb.openapi.Set("components|parameters|filter|schema|items|type", "string")
	oarb.openapi.Set("components|parameters|filter|description", "Filters to be applied. Each filter consists of a column, an operator and a value (comma separated). Example: id,eq,1")
	oarb.openapi.Set("components|parameters|filter|required", false)

	oarb.openapi.Set("components|parameters|include|name", "include")
	oarb.openapi.Set("components|parameters|include|in", "query")
	oarb.openapi.Set("components|parameters|include|schema|type", "string")
	oarb.openapi.Set("components|parameters|include|description", "Columns you want to include in the output (comma separated). Example: posts.*,categories.name")
	oarb.openapi.Set("components|parameters|include|required", false)

	oarb.openapi.Set("components|parameters|exclude|name", "exclude")
	oarb.openapi.Set("components|parameters|exclude|in", "query")
	oarb.openapi.Set("components|parameters|exclude|schema|type", "string")
	oarb.openapi.Set("components|parameters|exclude|description", "Columns you want to exclude from the output (comma separated). Example: posts.content")
	oarb.openapi.Set("components|parameters|exclude|required", false)

	oarb.openapi.Set("components|parameters|order|name", "order")
	oarb.openapi.Set("components|parameters|order|in", "query")
	oarb.openapi.Set("components|parameters|order|schema|type", "array")
	oarb.openapi.Set("components|parameters|order|schema|items|type", "string")
	oarb.openapi.Set("components|parameters|order|description", "Column you want to sort on and the sort direction (comma separated). Example: id,desc")
	oarb.openapi.Set("components|parameters|order|required", false)

	oarb.openapi.Set("components|parameters|size|name", "size")
	oarb.openapi.Set("components|parameters|size|in", "query")
	oarb.openapi.Set("components|parameters|size|schema|type", "string")
	oarb.openapi.Set("components|parameters|size|description", "Maximum number of results (for top lists). Example: 10")
	oarb.openapi.Set("components|parameters|size|required", false)

	oarb.openapi.Set("components|parameters|page|name", "page")
	oarb.openapi.Set("components|parameters|page|in", "query")
	oarb.openapi.Set("components|parameters|page|schema|type", "string")
	oarb.openapi.Set("components|parameters|page|description", "Page number and page size (comma separated). Example: 1,10")
	oarb.openapi.Set("components|parameters|page|required", false)

	oarb.openapi.Set("components|parameters|join|name", "join")
	oarb.openapi.Set("components|parameters|join|in", "query")
	oarb.openapi.Set("components|parameters|join|schema|type", "array")
	oarb.openapi.Set("components|parameters|join|schema|items|type", "string")
	oarb.openapi.Set("components|parameters|join|description", "Paths (comma separated) to related entities that you want to include. Example: comments,users")
	oarb.openapi.Set("components|parameters|join|required", false)
}

func (oarb OpenApiRecordsBuilder) setTag(index int, tableName string) {
	oarb.openapi.Set(fmt.Sprintf("tags|%d|name", index), tableName)
	oarb.openapi.Set(fmt.Sprintf("tags|%d|description", index), fmt.Sprintf("%s operations", tableName))
}

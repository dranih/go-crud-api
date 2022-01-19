package openapi

import (
	"fmt"
	"strings"
)

type OpenApiColumnsBuilder struct {
	openapi    *OpenApiDefinition
	operations map[string]map[string]string
}

func (oab *OpenApiBuilder) NewOpenApiColumnsBuilder(openapi *OpenApiDefinition) *OpenApiColumnsBuilder {
	return &OpenApiColumnsBuilder{
		openapi,
		map[string]map[string]string{
			"database": {
				"read": "get",
			},
			"table": {
				"create": "post",
				"read":   "get",
				"update": "put", //rename
				"delete": "delete",
			},
			"column": {
				"create": "post",
				"read":   "get",
				"update": "put",
				"delete": "delete",
			},
		},
	}
}

func (oacb *OpenApiColumnsBuilder) Build() {
	oacb.setPaths()
	oacb.openapi.Set("components|responses|bool-success|description", "boolean indicating success or failure")
	oacb.openapi.Set("components|responses|bool-success|content|application/json|schema|type", "boolean")
	oacb.setComponentSchema()
	oacb.setComponentResponse()
	oacb.setComponentRequestBody()
	oacb.setComponentParameters()
	i := 0
	for key := range oacb.operations {
		oacb.setTag(i, key)
		i++
	}
}

func (oacb *OpenApiColumnsBuilder) setPaths() {
	for optype, operationPair := range oacb.operations {
		for operation, method := range operationPair {
			var path string
			var parameters []string
			switch optype {
			case "database":
				path = "/columns"
			case "table":
				if operation == "create" {
					path = "/columns"
				} else {
					path = "/columns/{table}"
				}
			case "column":
				if operation == "create" {
					path = "/columns/{table}"
				} else {
					path = "/columns/{table}/{column}"
				}
			}
			if strings.Index(path, "{table}") >= 0 {
				parameters = append(parameters, "table")
			}
			if strings.Index(path, "{column}") >= 0 {
				parameters = append(parameters, "column")
			}
			for p, parameter := range parameters {
				oacb.openapi.Set(fmt.Sprintf("paths|%s|%s|parameters|%d|$ref", path, method, p), fmt.Sprintf("#/components/parameters/%s", parameter))
			}
			if _, exists := map[string]bool{"create": true, "update": true}[operation]; exists {
				oacb.openapi.Set(fmt.Sprintf("paths|%s|%s|requestBody|$ref", path, method), fmt.Sprintf("#/components/requestBodies/%s-%s", operation, optype))
			}
			oacb.openapi.Set(fmt.Sprintf("paths|%s|%s|tags|0", path, method), optype)
			oacb.openapi.Set(fmt.Sprintf("paths|%s|%s|operationId", path, method), fmt.Sprintf("%s_%s", operation, optype))
			if operation == "update" && optype == "table" {
				oacb.openapi.Set(fmt.Sprintf("paths|%s|%s|description", path, method), "rename table")
			} else {
				oacb.openapi.Set(fmt.Sprintf("paths|%s|%s|description", path, method), fmt.Sprintf("%s %s", operation, optype))
			}
			switch operation {
			case "read":
				oacb.openapi.Set(fmt.Sprintf(`paths|%s|%s|responses|200|$ref`, path, method), fmt.Sprintf("#/components/responses/%s-%s", operation, optype))
			case "create", "update", "delete":
				oacb.openapi.Set(fmt.Sprintf(`paths|%s|%s|responses|200|$ref`, path, method), "#/components/responses/bool-success")
			}
		}
	}
}

func (oacb *OpenApiColumnsBuilder) setComponentSchema() {
	for optype, operationPair := range oacb.operations {
		for operation := range operationPair {
			if operation == "delete" {
				continue
			}
			prefix := fmt.Sprintf("components|schemas|%s-%s", operation, optype)
			oacb.openapi.Set(fmt.Sprintf("%s|type", prefix), "object")
			switch operation {
			case "database":
				oacb.openapi.Set(fmt.Sprintf("%s|properties|tables|type", prefix), "array")
				oacb.openapi.Set(fmt.Sprintf("%s|properties|tables|items|$ref", prefix), "#/components/schemas/read-table")
			case "table":
				if operation == "update" {
					oacb.openapi.Set(fmt.Sprintf("%s|required", prefix), []string{"name"})
					oacb.openapi.Set(fmt.Sprintf("%s|properties|name|type", prefix), "string")
				} else {
					oacb.openapi.Set(fmt.Sprintf("%s|properties|name|type", prefix), "string")
					if operation == "read" {
						oacb.openapi.Set(fmt.Sprintf("%s|properties|type|type", prefix), "string")
					}
					oacb.openapi.Set(fmt.Sprintf("%s|properties|columns|type", prefix), "array")
					oacb.openapi.Set(fmt.Sprintf("%s|properties|columns|items|$ref", prefix), "#/components/schemas/read-column")
				}
			case "column":
				oacb.openapi.Set(fmt.Sprintf("%s|required", prefix), []string{"name", "type"})
				oacb.openapi.Set(fmt.Sprintf("%s|properties|name|type", prefix), "string")
				oacb.openapi.Set(fmt.Sprintf("%s|properties|type|type", prefix), "string")
				oacb.openapi.Set(fmt.Sprintf("%s|properties|length|type", prefix), "integer")
				oacb.openapi.Set(fmt.Sprintf("%s|properties|length|format", prefix), "int64")
				oacb.openapi.Set(fmt.Sprintf("%s|properties|precision|type", prefix), "integer")
				oacb.openapi.Set(fmt.Sprintf("%s|properties|precision|format", prefix), "int64")
				oacb.openapi.Set(fmt.Sprintf("%s|properties|scale|type", prefix), "integer")
				oacb.openapi.Set(fmt.Sprintf("%s|properties|scale|format", prefix), "int64")
				oacb.openapi.Set(fmt.Sprintf("%s|properties|nullable|type", prefix), "boolean")
				oacb.openapi.Set(fmt.Sprintf("%s|properties|pk|type", prefix), "boolean")
				oacb.openapi.Set(fmt.Sprintf("%s|properties|fk|type", prefix), "string")
				break
			}
		}
	}
}

func (oacb *OpenApiColumnsBuilder) setComponentResponse() {
	for optype, operationPair := range oacb.operations {
		for operation := range operationPair {
			if operation != "read" {
				continue
			}
			oacb.openapi.Set(fmt.Sprintf("components|responses|%s-%s|description", operation, optype), fmt.Sprintf("single %s record", operation))
			oacb.openapi.Set(fmt.Sprintf(`components|responses|%s-%s|content|application/json|schema|$ref`, operation, optype), fmt.Sprintf("#/components/schemas/%s-%s", operation, optype))
		}
	}
}

func (oacb *OpenApiColumnsBuilder) setComponentRequestBody() {
	for optype, operationPair := range oacb.operations {
		for operation := range operationPair {
			if _, exists := map[string]bool{"create": true, "update": true}[operation]; !exists {
				continue
			}
			oacb.openapi.Set(fmt.Sprintf("components|requestBodies|%s-%s|description", operation, optype), fmt.Sprintf("single %s record", optype))
			oacb.openapi.Set(fmt.Sprintf(`components|requestBodies|%s-%s|content|application/json|schema|$ref`, operation, optype), fmt.Sprintf("#/components/schemas/%s-%s", operation, optype))
		}
	}
}

func (oacb *OpenApiColumnsBuilder) setComponentParameters() {
	oacb.openapi.Set("components|parameters|table|name", "table")
	oacb.openapi.Set("components|parameters|table|in", "path")
	oacb.openapi.Set("components|parameters|table|schema|type", "string")
	oacb.openapi.Set("components|parameters|table|description", "table name")
	oacb.openapi.Set("components|parameters|table|required", true)

	oacb.openapi.Set("components|parameters|column|name", "column")
	oacb.openapi.Set("components|parameters|column|in", "path")
	oacb.openapi.Set("components|parameters|column|schema|type", "string")
	oacb.openapi.Set("components|parameters|column|description", "column name")
	oacb.openapi.Set("components|parameters|column|required", true)
}

func (oacb *OpenApiColumnsBuilder) setTag(index int, optype string) {
	oacb.openapi.Set(fmt.Sprintf("tags|%d|name", index), optype)
	oacb.openapi.Set(fmt.Sprintf("tags|%d|description", index), fmt.Sprintf("%s operations", optype))
}

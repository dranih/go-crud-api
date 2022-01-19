package openapi

import "fmt"

type OpenApiStatusBuilder struct {
	openapi    *OpenApiDefinition
	operations map[string]map[string]string
}

func (oab *OpenApiBuilder) NewOpenApiStatusBuilder(openapi *OpenApiDefinition) *OpenApiStatusBuilder {
	return &OpenApiStatusBuilder{
		openapi,
		map[string]map[string]string{"status": {"ping": "get"}},
	}
}

func (oasb *OpenApiStatusBuilder) Build() {
	oasb.setPaths()
	oasb.setComponentSchema()
	oasb.setComponentResponse()
	i := 0
	for key := range oasb.operations {
		oasb.setTag(i, key)
		i++
	}
}

func (oasb *OpenApiStatusBuilder) setPaths() {
	for optype, operationPair := range oasb.operations {
		for operation, method := range operationPair {
			path := fmt.Sprintf("/%s/%s", optype, operation)
			oasb.openapi.Set(fmt.Sprintf("paths|%s|%s|tags|0", path, method), optype)
			oasb.openapi.Set(fmt.Sprintf("paths|%s|%s|operationId", path, method), fmt.Sprintf("%s_%s", operation, optype))
			oasb.openapi.Set(fmt.Sprintf("paths|%s|%s|description", path, method), fmt.Sprintf("Request API '%s' status", operation))
			oasb.openapi.Set(fmt.Sprintf(`paths|%s|%s|responses|200|$ref`, path, method), fmt.Sprintf("#/components/responses/%s-%s", operation, optype))
		}
	}
}

func (oasb *OpenApiStatusBuilder) setComponentSchema() {
	for optype, operationPair := range oasb.operations {
		for operation := range operationPair {
			prefix := fmt.Sprintf("components|schemas|%s-%s", operation, optype)
			oasb.openapi.Set(fmt.Sprintf("%s|type", prefix), "object")
			switch operation {
			case "ping":
				oasb.openapi.Set(fmt.Sprintf("%s|required", prefix), []string{"db", "cache"})
				oasb.openapi.Set(fmt.Sprintf("%s|properties|db|type", prefix), "integer")
				oasb.openapi.Set(fmt.Sprintf("%s|properties|db|format", prefix), "int64")
				oasb.openapi.Set(fmt.Sprintf("%s|properties|cache|type", prefix), "integer")
				oasb.openapi.Set(fmt.Sprintf("%s|properties|cache|format", prefix), "int64")
			}
		}
	}
}

func (oasb *OpenApiStatusBuilder) setComponentResponse() {
	for optype, operationPair := range oasb.operations {
		for operation := range operationPair {
			oasb.openapi.Set(fmt.Sprintf("components|responses|%s-%s|description", operation, optype), fmt.Sprintf("%s status record", operation))
			oasb.openapi.Set(fmt.Sprintf(`components|responses|%s-%s|content|application/json|schema|$ref`, operation, optype), fmt.Sprintf("#/components/schemas/%s-%s", operation, optype))
		}
	}
}

func (oasb *OpenApiStatusBuilder) setTag(index int, optype string) {
	oasb.openapi.Set(fmt.Sprintf("tags|%d|name", index), optype)
	oasb.openapi.Set(fmt.Sprintf("tags|%d|description", index), fmt.Sprintf("%s operations", optype))
}

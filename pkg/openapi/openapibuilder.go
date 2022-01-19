package openapi

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/dranih/go-crud-api/pkg/database"
)

type Builder interface {
	Build()
}

type OpenApiBuilder struct {
	openapi  *OpenApiDefinition
	records  *OpenApiRecordsBuilder
	columns  *OpenApiColumnsBuilder
	status   *OpenApiStatusBuilder
	builders []Builder
}

func NewOpenApiBuilder(reflection *database.ReflectionService, base map[string]interface{}, controllers, builders map[string]bool) *OpenApiBuilder {
	oab := &OpenApiBuilder{}
	oab.openapi = &OpenApiDefinition{base}
	if controllers["records"] {
		oab.records = oab.NewOpenApiRecordsBuilder(oab.openapi, reflection)
	}
	if controllers["columns"] {
		oab.columns = oab.NewOpenApiColumnsBuilder(oab.openapi)
	}
	if controllers["status"] {
		oab.status = oab.NewOpenApiStatusBuilder(oab.openapi)
	}
	for builder := range builders {
		// We try to call func New{custom builder name}
		// The MyCustomBuilder needs to have a function OpenApiBuilder.NewMyCustomBuilder(openapi,reflection)
		f := reflect.ValueOf(oab).MethodByName(fmt.Sprintf("New%s", builder))
		if f.IsValid() {
			res := f.Call([]reflect.Value{reflect.ValueOf(oab.openapi), reflect.ValueOf(reflection)})
			result := res[0].Interface()
			if buiderInterface, ok := result.(Builder); ok {
				oab.builders = append(oab.builders, buiderInterface)
			}
		}
	}
	return oab
}

func (oab *OpenApiBuilder) getServerUrl(r *http.Request) string {
	if r.URL.IsAbs() {
		host := r.Host
		if i := strings.Index(host, ":"); i != -1 {
			host = host[:i]
		}
		return host
	}
	return r.URL.Host
}

func (oab *OpenApiBuilder) Build(r *http.Request) *OpenApiDefinition {
	oab.openapi.Set("openapi", "3.0.0")
	if !oab.openapi.Has("servers") {
		oab.openapi.Set("servers|0|url", oab.getServerUrl(r))
	}
	if oab.records != nil {
		oab.records.Build()
	}
	if oab.columns != nil {
		oab.columns.Build()
	}
	if oab.status != nil {
		oab.status.Build()
	}
	for _, builder := range oab.builders {
		builder.Build()
	}
	return oab.openapi
}

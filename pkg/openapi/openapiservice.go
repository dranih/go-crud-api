package openapi

import (
	"net/http"

	"github.com/dranih/go-crud-api/pkg/database"
)

type OpenApiService struct {
	builder *OpenApiBuilder
}

func NewOpenApiService(reflection *database.ReflectionService, base map[string]interface{}, controllers, customBuilders map[string]bool) *OpenApiService {
	return &OpenApiService{NewOpenApiBuilder(reflection, base, controllers, customBuilders)}
}

func (oas *OpenApiService) Get(r *http.Request) *OpenApiDefinition {
	return oas.builder.Build(r)
}

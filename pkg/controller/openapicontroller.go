package controller

import (
	"net/http"

	"github.com/dranih/go-crud-api/pkg/openapi"
	"github.com/gorilla/mux"
)

type OpenApiController struct {
	openapiService *openapi.OpenApiService
	responder      Responder
}

func NewOpenApiController(router *mux.Router, responder Responder, service *openapi.OpenApiService) *OpenApiController {
	oac := &OpenApiController{service, responder}
	router.HandleFunc("/openapi", oac.openapi).Methods("GET")
	return oac
}

func (oac *OpenApiController) openapi(w http.ResponseWriter, r *http.Request) {
	oac.responder.Success(oac.openapiService.Get(r), w)
}

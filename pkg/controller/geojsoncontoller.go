package controller

import (
	"net/http"
	"strings"

	"github.com/dranih/go-crud-api/pkg/geojson"
	"github.com/dranih/go-crud-api/pkg/record"
	"github.com/dranih/go-crud-api/pkg/utils"
	"github.com/gorilla/mux"
)

type GeoJsonController struct {
	service   *geojson.Service
	responder Responder
}

func NewGeoJsonController(router *mux.Router, responder Responder, service *geojson.Service) *GeoJsonController {
	gc := &GeoJsonController{service, responder}
	router.HandleFunc("/geojson/{table}", gc.list).Methods("GET")
	router.HandleFunc("/geojson/{table}/{id}", gc.read).Methods("GET")
	return gc
}

func (gc *GeoJsonController) list(w http.ResponseWriter, r *http.Request) {
	table := mux.Vars(r)["table"]
	params := utils.GetRequestParams(r)
	if !gc.service.HasTable(table) {
		gc.responder.Error(record.TABLE_NOT_FOUND, table, w, "")
		return
	}
	result := gc.service.List(table, params)
	gc.responder.Success(result, w)
	return
}

func (gc *GeoJsonController) read(w http.ResponseWriter, r *http.Request) {
	table := mux.Vars(r)["table"]
	if !gc.service.HasTable(table) {
		gc.responder.Error(record.TABLE_NOT_FOUND, table, w, "")
		return
	}
	if gc.service.GetType(table) != "table" {
		gc.responder.Error(record.OPERATION_NOT_SUPPORTED, "read", w, "")
		return
	}
	params := utils.GetRequestParams(r)
	id := mux.Vars(r)["id"]
	if strings.Index(id, ",") != -1 {
		ids := strings.Split(id, `,`)
		results := struct {
			Type     string `json:"type"`
			features []*geojson.Feature
		}{"FeatureCollection", nil}
		for i := 0; i < len(ids); i++ {
			results.features = append(results.features, gc.service.Read(table, ids[i], params))
		}
		gc.responder.Success(results, w)
		return
	} else {
		response := gc.service.Read(table, id, params)
		if response == nil {
			gc.responder.Error(record.RECORD_NOT_FOUND, id, w, "")
			return
		}
		gc.responder.Success(response, w)
	}
}

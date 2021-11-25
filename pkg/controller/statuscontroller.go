package controller

import (
	"net/http"

	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/gorilla/mux"
)

type StatusController struct {
	db        *database.GenericDB
	cache     interface{}
	responder Responder
}

func NewStatusController(router *mux.Router, responder Responder, cache interface{}, db *database.GenericDB) *StatusController {
	sc := &StatusController{db, cache, responder}
	router.HandleFunc("/status/ping", sc.ping).Methods("GET")
	return sc
}

func (sc *StatusController) ping(w http.ResponseWriter, r *http.Request) {
	result := map[string]int{"db": sc.db.Ping(), "cache": 0}
	sc.responder.Success(result, w)
}

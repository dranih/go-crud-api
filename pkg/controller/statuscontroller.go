package controller

import (
	"net/http"

	"github.com/dranih/go-crud-api/pkg/cache"
	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/gorilla/mux"
)

type StatusController struct {
	db        *database.GenericDB
	cache     cache.Cache
	responder Responder
}

func NewStatusController(router *mux.Router, responder Responder, cache cache.Cache, db *database.GenericDB) *StatusController {
	sc := &StatusController{db, cache, responder}
	router.HandleFunc("/status/ping", sc.ping).Methods("GET")
	return sc
}

func (sc *StatusController) ping(w http.ResponseWriter, r *http.Request) {
	result := map[string]int{"db": sc.db.Ping(), "cache": sc.cache.Ping()}
	sc.responder.Success(result, w)
}

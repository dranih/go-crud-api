package controller

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dranih/go-crud-api/pkg/cache"
	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/gorilla/mux"
)

type StatusController struct {
	db        *database.GenericDB
	cache     cache.Cache
	responder Responder
}

func NewStatusController(router *mux.Router, responder Responder, lcache cache.Cache, db *database.GenericDB) *StatusController {
	if lcache == nil {
		prefix := fmt.Sprintf("gocrudapi-%d-", os.Getpid())
		lcache = cache.Create("TempFile", prefix, "")
	}
	sc := &StatusController{db, lcache, responder}
	router.HandleFunc("/status/ping", sc.ping).Methods("GET")
	return sc
}

func (sc *StatusController) ping(w http.ResponseWriter, r *http.Request) {
	result := map[string]int{"db": sc.db.Ping(), "cache": sc.cache.Ping()}
	sc.responder.Success(result, w)
}

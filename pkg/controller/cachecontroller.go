package controller

import (
	"net/http"

	"github.com/dranih/go-crud-api/pkg/cache"
	"github.com/gorilla/mux"
)

type CacheController struct {
	cache     cache.Cache
	responder Responder
}

func NewCacheController(router *mux.Router, responder Responder, cache cache.Cache) *CacheController {
	cc := &CacheController{cache, responder}
	router.HandleFunc("/cache/clear", cc.clear).Methods("GET")
	return cc
}

func (cc *CacheController) clear(w http.ResponseWriter, r *http.Request) {
	result := cc.cache.Clear()
	cc.responder.Success(result, w, r)
}

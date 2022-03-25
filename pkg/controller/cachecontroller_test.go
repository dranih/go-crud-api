package controller

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dranih/go-crud-api/pkg/cache"
	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/utils"
	"github.com/gorilla/mux"
)

func TestCacheController(t *testing.T) {
	db_path := utils.SelectConfig(true)
	db := database.NewGenericDB(
		"sqlite",
		db_path,
		0,
		"go-crud-api",
		nil,
		"go-crud-api",
		"go-crud-api",
	)
	defer db.PDO().CloseConn()
	prefix := fmt.Sprintf("gocrudapi-%d-", os.Getpid())
	cache := cache.Create("Gocache", prefix, "")
	reflection := database.NewReflectionService(db, cache, 10)
	responder := NewJsonResponder(false)
	definition := database.NewDefinitionService(db, reflection)
	router := mux.NewRouter()
	NewColumnController(router, responder, reflection, definition)
	NewCacheController(router, responder, cache)
	ts := httptest.NewServer(router)
	defer ts.Close()

	//https://ieftimov.com/post/testing-in-go-testing-http-servers/
	//https://stackoverflow.com/questions/42474259/golang-how-to-live-test-an-http-server
	tt := []utils.Test{
		{
			Name:       "get tables and columns ",
			Method:     http.MethodGet,
			Uri:        "/columns/barcodes",
			Body:       ``,
			WantJson:   `{"name":"barcodes","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"product_id","type":"integer","fk":"products"},{"name":"hex","type":"varchar","length":255},{"name":"bin","type":"blob"},{"name":"ip_address","type":"varchar","length":15,"nullable":true}]}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "clear cache",
			Method:     http.MethodGet,
			Uri:        "/cache/clear",
			Body:       ``,
			Want:       `true`,
			StatusCode: http.StatusOK,
		},
	}
	utils.RunTests(t, ts.URL, tt)
	if err := os.Remove(db_path); err != nil {
		panic(err)
	}
}

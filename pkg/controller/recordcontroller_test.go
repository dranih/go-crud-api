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

// TestRecordController should unit test the controller, not the whole api
// To rewrite
func TestRecordController(t *testing.T) {
	db := database.NewGenericDB(
		"sqlite",
		"../../test/test.db",
		0,
		"test",
		nil,
		"",
		"",
	)
	prefix := fmt.Sprintf("gocrudapi-%d-", os.Getpid())
	cache := cache.Create("TempFile", prefix, "")
	reflection := database.NewReflectionService(db, cache, 10)
	records := database.NewRecordService(db, reflection)
	responder := NewJsonResponder(false)
	router := mux.NewRouter()
	NewRecordController(router, responder, records)
	NewStatusController(router, responder, cache, db)
	ts := httptest.NewServer(router)
	defer ts.Close()

	//https://ieftimov.com/post/testing-in-go-testing-http-servers/
	//https://stackoverflow.com/questions/42474259/golang-how-to-live-test-an-http-server
	tt := []utils.Test{
		{
			Name:       "ping ",
			Method:     http.MethodGet,
			Uri:        "/status/ping",
			Body:       ``,
			WantRegex:  `{"cache":[0-9]+,"db":[0-9]+}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "get table ",
			Method:     http.MethodGet,
			Uri:        "/records/sharks",
			Body:       ``,
			WantRegex:  `"sharktype":"Megaladon"`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "get unique id ",
			Method:     http.MethodGet,
			Uri:        "/records/sharks/3",
			Body:       ``,
			Want:       `{"id":3,"length":1800,"name":"Himari","sharktype":"Megaladon"}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "get multiple ids ",
			Method:     http.MethodGet,
			Uri:        "/records/sharks/1,3",
			Body:       ``,
			Want:       `[{"id":1,"length":427,"name":"Sammy","sharktype":"Greenland Shark"},{"id":3,"length":1800,"name":"Himari","sharktype":"Megaladon"}]`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "post unique ",
			Method:     http.MethodPost,
			Uri:        "/records/sharks",
			Body:       `{"id":99,"name":"Tomy","length": "100","sharktype": "Great White Shark"}`,
			WantRegex:  `{"id":[0-9]+}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "put unique ",
			Method:     http.MethodPut,
			Uri:        "/records/sharks/99",
			Body:       `{"length": 2000}`,
			Want:       `{"RowsAffected":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "patch unique ",
			Method:     http.MethodPatch,
			Uri:        "/records/sharks/99",
			Body:       `{"length": 10}`,
			Want:       `{"RowsAffected":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "delete unique ",
			Method:     http.MethodDelete,
			Uri:        "/records/sharks/99",
			Body:       ``,
			Want:       `{"RowsAffected":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "post multiple ",
			Method:     http.MethodPost,
			Uri:        "/records/sharks",
			Body:       `[{"id":99,"name":"Tomy","length": "100","sharktype": "Great White Shark"},{"id":999,"name":"Barbara","length": "150","sharktype": "Hammer head"}]`,
			WantRegex:  `[{"id":[0-9]+},{"id":[0-9]+}]`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "put multiples ",
			Method:     http.MethodPut,
			Uri:        "/records/sharks/99,999",
			Body:       `[{"length": 2000},{"name": "Barbara3","length": 1000}]`,
			Want:       `[{"RowsAffected":1},{"RowsAffected":1}]`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "patch multiple ",
			Method:     http.MethodPatch,
			Uri:        "/records/sharks/99,999",
			Body:       `[{"length": 10},{"length": 50}]`,
			Want:       `[{"RowsAffected":1},{"RowsAffected":1}]`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "delete multiple ",
			Method:     http.MethodDelete,
			Uri:        "/records/sharks/99,999",
			Body:       ``,
			Want:       `[{"RowsAffected":1},{"RowsAffected":1}]`,
			StatusCode: http.StatusOK,
		},
	}
	utils.RunTests(t, ts.URL, tt)
}

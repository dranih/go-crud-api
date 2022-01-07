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

func TestColumnController(t *testing.T) {
	db := database.NewGenericDB(
		"sqlite",
		"../../test/test.db",
		0,
		"test",
		map[string]bool{"cows": true, "sharks": true},
		"",
		"",
	)
	prefix := fmt.Sprintf("gocrudapi-%d-", os.Getpid())
	cache := cache.Create("TempFile", prefix, "")
	reflection := database.NewReflectionService(db, cache, 10)
	responder := NewJsonResponder(false)
	definition := database.NewDefinitionService(db, reflection)
	router := mux.NewRouter()
	NewColumnController(router, responder, reflection, definition)
	ts := httptest.NewServer(router)
	defer ts.Close()

	//https://ieftimov.com/post/testing-in-go-testing-http-servers/
	//https://stackoverflow.com/questions/42474259/golang-how-to-live-test-an-http-server
	tt := []utils.Test{
		{
			Name:       "get tables and columns ",
			Method:     http.MethodGet,
			Uri:        "/columns",
			Body:       ``,
			WantRegex:  `\{"tables":\[\{"columns":.*\}\]\}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "get 1 table and columns ",
			Method:     http.MethodGet,
			Uri:        "/columns/cows",
			Body:       ``,
			WantRegex:  `\{"columns":\[.*\],"name":"cows","type":"table"\}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "get inexistant table ",
			Method:     http.MethodGet,
			Uri:        "/columns/doesnotexists",
			Body:       ``,
			Want:       "{\"code\":1001,\"details\":\"\",\"message\":\"Table `doesnotexists` not found\"}",
			StatusCode: http.StatusNotFound,
		},
		{
			Name:       "get 1 table 1 column ",
			Method:     http.MethodGet,
			Uri:        "/columns/cows/length",
			Body:       ``,
			WantRegex:  `\{.*"name":"length",.*\}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "get inexistant column ",
			Method:     http.MethodGet,
			Uri:        "/columns/cows/doesnotexists",
			Body:       ``,
			Want:       "{\"code\":1005,\"details\":\"\",\"message\":\"Column `doesnotexists` not found\"}",
			StatusCode: http.StatusNotFound,
		},
		{
			Name:       "update column ",
			Method:     http.MethodPut,
			Uri:        "/columns/cows",
			Body:       `{"name":"cows2"}`,
			Want:       "true",
			StatusCode: http.StatusOK,
		},
		{
			Name:       "update (back) column - test refresh tables",
			Method:     http.MethodPut,
			Uri:        "/columns/cows2",
			Body:       `{"name":"cows"}`,
			Want:       "true",
			StatusCode: http.StatusOK,
		},
		{
			Name:       "create table cows2",
			Method:     http.MethodPost,
			Uri:        "/columns",
			Body:       `{"name":"cows2","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"shark","type":"integer","fk":"sharks"},{"name":"name","type":"varchar","length":255},{"name":"cowtype","type":"varchar","length":15,"nullable":true}]}`,
			Want:       "true",
			StatusCode: http.StatusOK,
		},
		{
			Name:       "get table cows2",
			Method:     http.MethodGet,
			Uri:        "/columns/cows2",
			Body:       ``,
			WantRegex:  `\{"columns":\[.*\],"name":"cows2","type":"table"\}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "create column cows2 length",
			Method:     http.MethodPost,
			Uri:        "/columns/cows2",
			Body:       `{"name":"length","type":"integer"}`,
			Want:       "true",
			StatusCode: http.StatusOK,
		},
		{
			Name:       "get column cows2 length",
			Method:     http.MethodGet,
			Uri:        "/columns/cows2/length",
			Body:       ``,
			WantRegex:  `\{.*"name":"length",.*\}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "delete cows2 length",
			Method:     http.MethodDelete,
			Uri:        "/columns/cows2/length",
			Body:       ``,
			Want:       "true",
			StatusCode: http.StatusOK,
		},
		{
			Name:       "delete cows2",
			Method:     http.MethodDelete,
			Uri:        "/columns/cows2",
			Body:       ``,
			Want:       "true",
			StatusCode: http.StatusOK,
		},
	}
	utils.RunTests(t, ts, tt)
}

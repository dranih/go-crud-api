package controller

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dranih/go-crud-api/pkg/cache"
	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/record"
	"github.com/dranih/go-crud-api/pkg/utils"
	"github.com/gorilla/mux"
)

// TestRecordController should unit test the controller, not the whole api
// To rewrite
func TestRecordController(t *testing.T) {
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
	cache := cache.Create("TempFile", prefix, "")
	reflection := database.NewReflectionService(db, cache, 10)
	records := record.NewRecordService(db, reflection)
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
			Uri:        "/records/posts",
			Body:       ``,
			WantJson:   `{"records":[{"category_id":1,"content":"blog started","id":1,"user_id":1},{"category_id":2,"content":"It works!","id":2,"user_id":1}]}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "get unique id ",
			Method:     http.MethodGet,
			Uri:        "/records/posts/2",
			Body:       ``,
			WantJson:   `{"category_id":2,"content":"It works!","id":2,"user_id":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "get multiple ids ",
			Method:     http.MethodGet,
			Uri:        "/records/posts/1,2",
			Body:       ``,
			WantJson:   `[{"category_id":1,"content":"blog started","id":1,"user_id":1},{"category_id":2,"content":"It works!","id":2,"user_id":1}]`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "post unique ",
			Method:     http.MethodPost,
			Uri:        "/records/posts",
			Body:       `{"user_id":1,"category_id":1,"content":"test"}`,
			Want:       `3`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "put unique ",
			Method:     http.MethodPut,
			Uri:        "/records/posts/3",
			Body:       `{"user_id":1,"category_id":1,"content":"test (edited)"}`,
			Want:       `1`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "patch unique ",
			Method:     http.MethodPatch,
			Uri:        "/records/posts/3",
			Body:       `{"category_id":1}`,
			Want:       `1`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "delete unique ",
			Method:     http.MethodDelete,
			Uri:        "/records/posts/3",
			Body:       ``,
			Want:       `1`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "post multiple ",
			Method:     http.MethodPost,
			Uri:        "/records/posts",
			Body:       `[{"user_id":1,"category_id":1,"content":"test"},{"user_id":1,"category_id":1,"content":"test"}]`,
			Want:       `[4,5]`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "put multiples ",
			Method:     http.MethodPut,
			Uri:        "/records/posts/4,5",
			Body:       `[{"category_id":2},{"category_id":2}]`,
			Want:       `[1,1]`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "patch multiple ",
			Method:     http.MethodPatch,
			Uri:        "/records/posts/4,5",
			Body:       `[{"category_id":1},{"category_id":1}]`,
			Want:       `[1,1]`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "delete multiple ",
			Method:     http.MethodDelete,
			Uri:        "/records/posts/4,5",
			Body:       ``,
			Want:       `[1,1]`,
			StatusCode: http.StatusOK,
		},
	}
	utils.RunTests(t, ts.URL, tt)
	if err := os.Remove(db_path); err != nil {
		panic(err)
	}
}

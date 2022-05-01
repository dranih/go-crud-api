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
	db_path := utils.SelectConfig(true)
	mapping := map[string]string{
		"abc_posts.abc_id":          "posts.id",
		"abc_posts.abc_user_id":     "posts.user_id",
		"abc_posts.abc_category_id": "posts.category_id",
		"abc_posts.abc_content":     "posts.content"}
	db := database.NewGenericDB(
		"sqlite",
		db_path,
		0,
		"go-crud-api",
		nil,
		mapping,
		"go-crud-api",
		"go-crud-api",
	)
	defer db.PDO().CloseConn()
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
			WantJson:   `{"tables":[{"columns":[{"length":36,"name":"id","pk":true,"type":"varchar"}],"name":"invisibles","type":"table"},{"columns":[{"name":"created_at","type":"timestamp"},{"name":"deleted_at","nullable":true,"type":"timestamp"},{"name":"id","pk":true,"type":"integer"},{"length":255,"name":"name","type":"varchar"},{"name":"price","precision":10,"scale":2,"type":"decimal"},{"name":"properties","type":"clob"}],"name":"products","type":"table"},{"columns":[{"length":255,"name":"api_key","nullable":true,"type":"varchar"},{"name":"id","pk":true,"type":"integer"},{"name":"location","nullable":true,"type":"clob"},{"length":255,"name":"password","type":"varchar"},{"length":255,"name":"username","type":"varchar"}],"name":"users","type":"table"},{"columns":[{"name":"bin","type":"blob"},{"length":255,"name":"hex","type":"varchar"},{"name":"id","pk":true,"type":"integer"},{"length":15,"name":"ip_address","nullable":true,"type":"varchar"},{"fk":"products","name":"product_id","type":"integer"}],"name":"barcodes","type":"table"},{"columns":[{"name":"id","pk":true,"type":"integer"},{"length":255,"name":"name","type":"varchar"},{"name":"shape","type":"clob"}],"name":"countries","type":"table"},{"columns":[{"name":"id","pk":true,"type":"integer"},{"fk":"posts","name":"post_id","type":"integer"},{"fk":"tags","name":"tag_id","type":"integer"}],"name":"post_tags","type":"table"},{"columns":[{"name":"count","type":"clob"},{"name":"id","pk":true,"type":"integer"},{"length":255,"name":"name","type":"varchar"}],"name":"tag_usage","type":"view"},{"alias":"posts","columns":[{"alias":"category_id","fk":"categories","name":"abc_category_id","type":"integer"},{"alias":"content","length":255,"name":"abc_content","type":"varchar"},{"alias":"id","name":"abc_id","pk":true,"type":"integer"},{"alias":"user_id","fk":"users","name":"abc_user_id","type":"integer"}],"name":"abc_posts","type":"table"},{"columns":[{"fk":"categories","name":"category_id","type":"integer"},{"name":"id","pk":true,"type":"integer"},{"length":255,"name":"message","type":"varchar"},{"fk":"posts","name":"post_id","type":"integer"}],"name":"comments","type":"table"},{"columns":[{"length":36,"name":"id","type":"varchar"}],"name":"nopk","type":"table"},{"columns":[{"name":"id","pk":true,"type":"integer"},{"name":"is_important","type":"boolean"},{"length":255,"name":"name","type":"varchar"}],"name":"tags","type":"table"},{"columns":[{"name":"icon","nullable":true,"type":"blob"},{"name":"id","pk":true,"type":"integer"},{"length":255,"name":"name","type":"varchar"}],"name":"categories","type":"table"},{"columns":[{"name":"datetime","nullable":true,"type":"timestamp"},{"name":"id","pk":true,"type":"integer"},{"length":255,"name":"name","type":"varchar"},{"name":"visitors","nullable":true,"type":"bigint"}],"name":"events","type":"table"},{"columns":[{"name":"Umlauts ä_ö_ü-COUNT","type":"integer"},{"length":36,"name":"id","pk":true,"type":"varchar"},{"length":36,"name":"invisible","nullable":true,"type":"varchar"},{"fk":"invisibles","length":36,"name":"invisible_id","nullable":true,"type":"varchar"},{"fk":"users","name":"user_id","type":"integer"}],"name":"kunsthåndværk","type":"table"}]}`,
			StatusCode: http.StatusOK,
			Driver:     "sqlite",
			SkipFor:    map[string]bool{"mysql": true, "pgsql": true, "sqlsrv": true},
		},
		{
			Name:       "get 1 table and columns ",
			Method:     http.MethodGet,
			Uri:        "/columns/barcodes",
			Body:       ``,
			WantJson:   `{"name":"barcodes","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"product_id","type":"integer","fk":"products"},{"name":"hex","type":"varchar","length":255},{"name":"bin","type":"blob"},{"name":"ip_address","type":"varchar","length":15,"nullable":true}]}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "get inexistant table ",
			Method:     http.MethodGet,
			Uri:        "/columns/doesnotexists",
			Body:       ``,
			Want:       `{"code":1001,"message":"Table 'doesnotexists' not found"}`,
			StatusCode: http.StatusNotFound,
		},
		{
			Name:       "get 1 table 1 column ",
			Method:     http.MethodGet,
			Uri:        "/columns/barcodes/id",
			Body:       ``,
			WantJson:   `{"name":"id","type":"integer","pk":true}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "get inexistant column ",
			Method:     http.MethodGet,
			Uri:        "/columns/barcodes/doesnotexists",
			Body:       ``,
			Want:       `{"code":1005,"message":"Column 'doesnotexists' not found"}`,
			StatusCode: http.StatusNotFound,
		},
		{
			Name:       "update column ",
			Method:     http.MethodPut,
			Uri:        "/columns/barcodes/id",
			Body:       `{"name":"id2","type":"bigint"}`,
			Want:       `true`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "update (back) column - test refresh tables",
			Method:     http.MethodPut,
			Uri:        "/columns/barcodes/id2",
			Body:       `{"name":"id","type":"integer"}`,
			Want:       `true`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "create table barcodes2",
			Method:     http.MethodPost,
			Uri:        "/columns",
			Body:       `{"name":"barcodes2","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"product_id","type":"integer","fk":"products"},{"name":"hex","type":"varchar","length":255},{"name":"bin","type":"blob"},{"name":"ip_address","type":"varchar","length":15,"nullable":true}]}`,
			Want:       `true`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "get table barcodes2",
			Method:     http.MethodGet,
			Uri:        "/columns/barcodes2",
			Body:       ``,
			WantJson:   `{"name":"barcodes2","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"product_id","type":"integer","fk":"products"},{"name":"hex","type":"varchar","length":255},{"name":"bin","type":"blob"},{"name":"ip_address","type":"varchar","length":15,"nullable":true}]}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "create column barcodes2 test",
			Method:     http.MethodPost,
			Uri:        "/columns/barcodes2",
			Body:       `{"name":"test","type":"integer"}`,
			Want:       "true",
			StatusCode: http.StatusOK,
		},
		{
			Name:       "get column barcodes2 test",
			Method:     http.MethodGet,
			Uri:        "/columns/barcodes2/test",
			WantJson:   `{"name":"test","type":"integer"}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "delete barcodes2 test",
			Method:     http.MethodDelete,
			Uri:        "/columns/barcodes2/test",
			Want:       "true",
			StatusCode: http.StatusOK,
		},
		{
			Name:       "delete barcodes2",
			Method:     http.MethodDelete,
			Uri:        "/columns/barcodes2",
			Want:       "true",
			StatusCode: http.StatusOK,
		},
	}
	utils.RunTests(t, ts.URL, tt)
	if err := os.Remove(db_path); err != nil {
		panic(err)
	}
}

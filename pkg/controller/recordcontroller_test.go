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

func TestRecordController(t *testing.T) {
	db := database.NewGenericDB(
		"sqlite",
		"../../test/tests.db",
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
			Name:       "001_list_posts",
			Method:     http.MethodGet,
			Uri:        "/records/posts",
			Body:       ``,
			Want:       `{"records":[{"category_id":1,"content":"blog started","id":1,"user_id":1},{"category_id":2,"content":"It works!","id":2,"user_id":1}]}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "002_list_post_columns",
			Method:     http.MethodGet,
			Uri:        "/records/posts?include=id,content",
			Body:       ``,
			Want:       `{"records":[{"content":"blog started","id":1},{"content":"It works!","id":2}]}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "003_read_post_A",
			Method:     http.MethodGet,
			Uri:        "/records/posts/2",
			Body:       ``,
			Want:       `{"category_id":2,"content":"It works!","id":2,"user_id":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "003_read_post_B",
			Method:     http.MethodGet,
			Uri:        "/records/posts/0",
			Body:       ``,
			Want:       `{"code":1003,"details":"","message":"Record '0' not found"}`,
			StatusCode: http.StatusNotFound,
		},
		{
			Name:       "004_read_posts",
			Method:     http.MethodGet,
			Uri:        "/records/posts/1,2",
			Body:       ``,
			Want:       `[{"category_id":1,"content":"blog started","id":1,"user_id":1},{"category_id":2,"content":"It works!","id":2,"user_id":1}]`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "005_read_post_columns",
			Method:     http.MethodGet,
			Uri:        "/records/posts/2?include=id,content",
			Body:       ``,
			Want:       `{"content":"It works!","id":2}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "006_add_post",
			Method:     http.MethodPost,
			Uri:        "/records/posts",
			Body:       `{"user_id":1,"category_id":1,"content":"test"}`,
			Want:       `{"id":3}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "007_edit_post_A",
			Method:     http.MethodPut,
			Uri:        "/records/posts/3",
			Body:       `{"user_id":1,"category_id":1,"content":"test (edited)"}`,
			Want:       `{"RowsAffected":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "007_edit_post_B",
			Method:     http.MethodGet,
			Uri:        "/records/posts/3",
			Body:       ``,
			Want:       `{"category_id":1,"content":"test (edited)","id":3,"user_id":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:   "007_edit_post_C",
			Method: http.MethodPut,
			Uri:    "/records/posts/3",
			Body: `    {
				"user_id": 1,
				"category_id": 1,
				"content": "test (edited 1)"
			}`,
			Want:       `{"RowsAffected":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "007_edit_post_D",
			Method:     http.MethodGet,
			Uri:        "/records/posts/3",
			Body:       ``,
			Want:       `{"category_id":1,"content":"test (edited 1)","id":3,"user_id":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "008_edit_post_columns_missing_field_A",
			Method:     http.MethodPut,
			Uri:        "/records/posts/3?include=id,content",
			Body:       `{"content":"test (edited 2)"}`,
			Want:       `{"RowsAffected":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "008_edit_post_columns_missing_field_B",
			Method:     http.MethodGet,
			Uri:        "/records/posts/3",
			Body:       ``,
			Want:       `{"category_id":1,"content":"test (edited 2)","id":3,"user_id":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "009_edit_post_columns_extra_field_A",
			Method:     http.MethodPut,
			Uri:        "/records/posts/3?include=id,content",
			Body:       `{"user_id":2,"content":"test (edited 3)"}`,
			Want:       `{"RowsAffected":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "009_edit_post_columns_extra_field_B",
			Method:     http.MethodGet,
			Uri:        "/records/posts/3",
			Body:       ``,
			Want:       `{"category_id":1,"content":"test (edited 3)","id":3,"user_id":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "010_edit_post_with_utf8_content_A",
			Method:     http.MethodPut,
			Uri:        "/records/posts/2",
			Body:       `{"content":"🤗 Grüßgott, Вiтаю, dobrý deň, hyvää päivää, გამარჯობა, Γεια σας, góðan dag, здравствуйте"}`,
			Want:       `{"RowsAffected":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "010_edit_post_with_utf8_content_B",
			Method:     http.MethodGet,
			Uri:        "/records/posts/2",
			Body:       ``,
			Want:       `{"category_id":2,"content":"🤗 Grüßgott, Вiтаю, dobrý deň, hyvää päivää, გამარჯობა, Γεια σας, góðan dag, здравствуйте","id":2,"user_id":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:        "011_edit_post_with_utf8_content_with_post_A",
			Method:      http.MethodPut,
			Uri:         "/records/posts/2",
			Body:        `content=%F0%9F%A6%80%E2%82%AC%20Gr%C3%BC%C3%9Fgott%2C%20%D0%92i%D1%82%D0%B0%D1%8E%2C%20dobr%C3%BD%20de%C5%88%2C%20hyv%C3%A4%C3%A4%20p%C3%A4iv%C3%A4%C3%A4%2C%20%E1%83%92%E1%83%90%E1%83%9B%E1%83%90%E1%83%A0%E1%83%AF%E1%83%9D%E1%83%91%E1%83%90%2C%20%CE%93%CE%B5%CE%B9%CE%B1%20%CF%83%CE%B1%CF%82%2C%20g%C3%B3%C3%B0an%20dag%2C%20%D0%B7%D0%B4%D1%80%D0%B0%D0%B2%D1%81%D1%82%D0%B2%D1%83%D0%B9%D1%82%D0%B5`,
			ContentType: "application/x-www-form-urlencoded",
			Want:        `{"RowsAffected":1}`,
			StatusCode:  http.StatusOK,
		},
		{
			Name:       "011_edit_post_with_utf8_content_with_post_B",
			Method:     http.MethodGet,
			Uri:        "/records/posts/2",
			Body:       ``,
			Want:       `{"category_id":2,"content":"🦀€ Grüßgott, Вiтаю, dobrý deň, hyvää päivää, გამარჯობა, Γεια σας, góðan dag, здравствуйте","id":2,"user_id":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "012_delete_post_A",
			Method:     http.MethodDelete,
			Uri:        "/records/posts/3",
			Body:       ``,
			Want:       `{"RowsAffected":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "012_delete_post_B",
			Method:     http.MethodGet,
			Uri:        "/records/posts/3",
			Body:       ``,
			Want:       `{"code":1003,"details":"","message":"Record '3' not found"}`,
			StatusCode: http.StatusNotFound,
		},
	}
	utils.RunTests(t, ts, tt)
}

func TestRecordControllerOld(t *testing.T) {
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
	utils.RunTests(t, ts, tt)
}

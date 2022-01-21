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
			Want:       `{"code":1003,"message":"Record '0' not found"}`,
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
			Want:       `3`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "007_edit_post_A",
			Method:     http.MethodPut,
			Uri:        "/records/posts/3",
			Body:       `{"user_id":1,"category_id":1,"content":"test (edited)"}`,
			Want:       `1`,
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
			Want:       `1`,
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
			Want:       `1`,
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
			Want:       `1`,
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
			Want:       `1`,
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
			Want:        `1`,
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
			Want:       `1`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "012_delete_post_B",
			Method:     http.MethodGet,
			Uri:        "/records/posts/3",
			Body:       ``,
			Want:       `{"code":1003,"message":"Record '3' not found"}`,
			StatusCode: http.StatusNotFound,
		},
		{
			Name:        "013_add_post_with_post",
			Method:      http.MethodPost,
			Uri:         "/records/posts",
			Body:        `user_id=1&category_id=1&content=test`,
			ContentType: "application/x-www-form-urlencoded",
			Want:        `4`,
			StatusCode:  http.StatusOK,
		},
		{
			Name:        "014_edit_post_with_post_A",
			Method:      http.MethodPut,
			Uri:         "/records/posts/4",
			Body:        `user_id=1&category_id=1&content=test+(edited)`,
			ContentType: "application/x-www-form-urlencoded",
			Want:        `1`,
			StatusCode:  http.StatusOK,
		},
		{
			Name:       "014_edit_post_with_post_B",
			Method:     http.MethodGet,
			Uri:        "/records/posts/4",
			Body:       ``,
			Want:       `{"category_id":1,"content":"test (edited)","id":4,"user_id":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "015_delete_post_ignore_columns_A",
			Method:     http.MethodDelete,
			Uri:        "/records/posts/4?include=id,content",
			Body:       ``,
			Want:       `1`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "015_delete_post_ignore_columns_B",
			Method:     http.MethodGet,
			Uri:        "/records/posts/4",
			Body:       ``,
			Want:       `{"code":1003,"message":"Record '4' not found"}`,
			StatusCode: http.StatusNotFound,
		},
		{
			Name:       "016_list_with_paginate_A",
			Method:     http.MethodPost,
			Uri:        "/records/posts",
			Body:       `{"user_id":1,"category_id":1,"content":"#1"}`,
			Want:       `5`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "016_list_with_paginate_B",
			Method:     http.MethodPost,
			Uri:        "/records/posts",
			Body:       `{"user_id":1,"category_id":1,"content":"#2"}`,
			Want:       `6`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "016_list_with_paginate_C",
			Method:     http.MethodPost,
			Uri:        "/records/posts",
			Body:       `{"user_id":1,"category_id":1,"content":"#3"}`,
			Want:       `7`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "016_list_with_paginate_D",
			Method:     http.MethodPost,
			Uri:        "/records/posts",
			Body:       `{"user_id":1,"category_id":1,"content":"#4"}`,
			Want:       `8`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "016_list_with_paginate_E",
			Method:     http.MethodPost,
			Uri:        "/records/posts",
			Body:       `{"user_id":1,"category_id":1,"content":"#5"}`,
			Want:       `9`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "016_list_with_paginate_F",
			Method:     http.MethodPost,
			Uri:        "/records/posts",
			Body:       `{"user_id":1,"category_id":1,"content":"#6"}`,
			Want:       `10`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "016_list_with_paginate_G",
			Method:     http.MethodPost,
			Uri:        "/records/posts",
			Body:       `{"user_id":1,"category_id":1,"content":"#7"}`,
			Want:       `11`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "016_list_with_paginate_H",
			Method:     http.MethodPost,
			Uri:        "/records/posts",
			Body:       `{"user_id":1,"category_id":1,"content":"#8"}`,
			Want:       `12`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "016_list_with_paginate_I",
			Method:     http.MethodPost,
			Uri:        "/records/posts",
			Body:       `{"user_id":1,"category_id":1,"content":"#9"}`,
			Want:       `13`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "016_list_with_paginate_J",
			Method:     http.MethodPost,
			Uri:        "/records/posts",
			Body:       `{"user_id":1,"category_id":1,"content":"#10"}`,
			Want:       `14`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "016_list_with_paginate_K",
			Method:     http.MethodGet,
			Uri:        "/records/posts?page=2,2&order=id",
			Body:       ``,
			Want:       `{"records":[{"category_id":1,"content":"#1","id":5,"user_id":1},{"category_id":1,"content":"#2","id":6,"user_id":1}],"results":12}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "017_edit_post_primary_key",
			Method:     http.MethodPut,
			Uri:        "/records/posts/2",
			Body:       `{"id":1}`,
			Want:       `0`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "018_add_post_missing_field",
			Method:     http.MethodPost,
			Uri:        "/records/posts",
			Body:       `{"category_id":1,"content":"test"}`,
			Want:       `{"code":1010,"message":"Data integrity violation"}`,
			StatusCode: http.StatusConflict,
		},
		{
			Name:       "019_list_with_paginate_in_multiple_order",
			Method:     http.MethodGet,
			Uri:        "/records/posts?page=1,2&order=category_id,asc&order=id,desc",
			Body:       ``,
			Want:       `{"records":[{"category_id":1,"content":"#10","id":14,"user_id":1},{"category_id":1,"content":"#9","id":13,"user_id":1}],"results":12}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "020_list_with_paginate_in_descending_order",
			Method:     http.MethodGet,
			Uri:        "/records/posts?page=2,2&order=id,desc",
			Body:       ``,
			Want:       `{"records":[{"category_id":1,"content":"#8","id":12,"user_id":1},{"category_id":1,"content":"#7","id":11,"user_id":1}],"results":12}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "021_list_with_size",
			Method:     http.MethodGet,
			Uri:        "/records/posts?order=id&size=1",
			Body:       ``,
			Want:       `{"records":[{"category_id":1,"content":"blog started","id":1,"user_id":1}]}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "022_list_with_zero_page_size_A",
			Method:     http.MethodGet,
			Uri:        "/records/posts?order=id&page=1,0",
			Body:       ``,
			Want:       `{"records":[],"results":12}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "022_list_with_zero_page_size_B",
			Method:     http.MethodGet,
			Uri:        "/records/posts?filter=id,eq,0&page=1,0",
			Body:       ``,
			Want:       `{"records":[],"results":0}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "023_list_with_zero_size",
			Method:     http.MethodGet,
			Uri:        "/records/posts?order=id&size=0",
			Body:       ``,
			Want:       `{"records":[]}`,
			StatusCode: http.StatusOK,
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

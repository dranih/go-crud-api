package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/record"
	"github.com/dranih/go-crud-api/pkg/utils"
	"github.com/gorilla/mux"
)

func TestAuthorizationMiddleware(t *testing.T) {
	properties := map[string]interface{}{
		"columnHandler": "{{ if eq .ColumnName \"invisible\"}} false {{ else }} true {{ end }}",
		"tableHandler":  "{{ if eq .TableName \"invisibles\" }} false {{ else }} true {{ end }}",
		"recordHandler": "{{ if eq .TableName \"comments\"}} filter=message,neq,invisible {{ else }}{{ end }}",
	}

	db_path := utils.SelectConfig(true)
	db := database.NewGenericDB(
		"sqlite",
		db_path,
		0,
		"go-crud-api",
		nil,
		nil,
		"go-crud-api",
		"go-crud-api")
	defer db.PDO().CloseConn()
	reflection := database.NewReflectionService(db, nil, 0)
	router := mux.NewRouter()
	responder := controller.NewJsonResponder(false)
	amMiddle := NewAuthorizationMiddleware(responder, properties, reflection)
	records := record.NewRecordService(db, reflection)
	controller.NewRecordController(router, responder, records)
	router.Use(amMiddle.Process)
	ts := httptest.NewServer(router)
	defer ts.Close()

	tt := []utils.Test{
		{
			Name:       "forbidden ",
			Method:     http.MethodGet,
			Uri:        "/records/invisibles",
			Want:       `{"code":1001,"message":"Table 'invisibles' not found"}`,
			StatusCode: http.StatusNotFound,
		},
		{
			Name:       "allowed ",
			Method:     http.MethodGet,
			Uri:        "/records/events",
			WantJson:   `{"records":[{"datetime":"2016-01-01 13:01:01","id":1,"name":"Launch","visitors":0}]}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "update_invisble_column_kunsthandvaerk_A",
			Method:     http.MethodPost,
			Uri:        "/records/kunsthåndværk",
			Body:       `{"id":"b55decba-8eb5-436b-af3e-148f7b4eacda","Umlauts ä_ö_ü-COUNT":4,"user_id":1}`,
			Want:       `"b55decba-8eb5-436b-af3e-148f7b4eacda"`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "update_invisble_column_kunsthandvaerk_B",
			Method:     http.MethodGet,
			Uri:        "/records/kunsthåndværk/b55decba-8eb5-436b-af3e-148f7b4eacda",
			Want:       `{"Umlauts ä_ö_ü-COUNT":4,"id":"b55decba-8eb5-436b-af3e-148f7b4eacda","invisible_id":null,"user_id":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "update_invisble_column_kunsthandvaerk_C",
			Method:     http.MethodPut,
			Uri:        "/records/kunsthåndværk/b55decba-8eb5-436b-af3e-148f7b4eacda",
			Body:       `{"id":"b55decba-8eb5-436b-af3e-148f7b4eacda","Umlauts ä_ö_ü-COUNT":3,"invisible":"test"}`,
			Want:       `1`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "update_invisble_column_kunsthandvaerk_D",
			Method:     http.MethodGet,
			Uri:        "/records/kunsthåndværk/b55decba-8eb5-436b-af3e-148f7b4eacda",
			Body:       ``,
			Want:       `{"Umlauts ä_ö_ü-COUNT":3,"id":"b55decba-8eb5-436b-af3e-148f7b4eacda","invisible_id":null,"user_id":1}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "add_comment_with_invisible_record_A",
			Method:     http.MethodPost,
			Uri:        "/records/comments",
			Body:       `{"user_id":1,"post_id":2,"message":"invisible","category_id":3}`,
			Want:       `5`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "add_comment_with_invisible_record_B",
			Method:     http.MethodGet,
			Uri:        "/records/comments/5",
			Body:       ``,
			Want:       `{"code":1003,"message":"Record '5' not found"}`,
			StatusCode: http.StatusNotFound,
		},
	}
	utils.RunTests(t, ts.URL, tt)
	if err := os.Remove(db_path); err != nil {
		panic(err)
	}
}

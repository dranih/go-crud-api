package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/utils"
	"github.com/gorilla/mux"
)

func TestApiKeyDbAuth(t *testing.T) {
	properties := map[string]interface{}{
		"mode":  "required",
		"realm": "GoCrudApi : Api key required",
	}

	db_path := utils.SelectConfig(true)
	db := database.NewGenericDB(
		"sqlite",
		db_path,
		0,
		"go-crud-api",
		nil,
		"go-crud-api",
		"go-crud-api")
	defer db.PDO().CloseConn()
	reflection := database.NewReflectionService(db, nil, 0)
	router := mux.NewRouter()
	responder := controller.NewJsonResponder(false)
	akdamMiddle := NewApiKeyDbAuth(responder, properties, reflection, db)
	router.HandleFunc("/", allowedTest).Methods("GET")
	router.Use(akdamMiddle.Process)
	ts := httptest.NewServer(router)
	defer ts.Close()

	tt := []utils.Test{
		{
			Name:       "forbidden ",
			Method:     http.MethodGet,
			Uri:        "/",
			Want:       `{"code":1011,"details":"GoCrudApi : Api key required","message":"Authentication required"}`,
			StatusCode: http.StatusUnauthorized,
		},
		{
			Name:          "bad key",
			Method:        http.MethodGet,
			Uri:           "/",
			RequestHeader: map[string]string{"X-API-Key": "1234"},
			Want:          `{"code":1012,"message":"Authentication failed for '1234'"}`,
			StatusCode:    http.StatusForbidden,
		},
		{
			Name:          "auth ok",
			Method:        http.MethodGet,
			Uri:           "/",
			RequestHeader: map[string]string{"X-API-Key": "123456789abc"},
			Want:          `Allowed`,
			StatusCode:    http.StatusOK,
		},
	}
	utils.RunTests(t, ts.URL, tt)
	if err := os.Remove(db_path); err != nil {
		panic(err)
	}
}

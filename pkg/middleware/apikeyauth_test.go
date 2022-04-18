package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/utils"
	"github.com/gorilla/mux"
)

func TestApiKeyAuth(t *testing.T) {
	properties := map[string]interface{}{
		"mode":  "required",
		"realm": "GoCrudApi : Api key required",
		"keys":  "123456789abc,123456789def",
	}

	router := mux.NewRouter()
	responder := controller.NewJsonResponder(false)
	akamMiddle := NewApiKeyAuth(responder, properties)
	router.HandleFunc("/", utils.AllowedTest).Methods("GET")
	router.Use(akamMiddle.Process)
	ts := httptest.NewServer(router)
	defer ts.Close()

	tt := []utils.Test{
		{
			Name:       "forbidden ",
			Method:     http.MethodGet,
			Uri:        "/",
			WantRegex:  `{"code":1011,"details":"GoCrudApi : Api key required","message":"Authentication required.*`,
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
			Name:          "auth ok 1",
			Method:        http.MethodGet,
			Uri:           "/",
			RequestHeader: map[string]string{"X-API-Key": "123456789abc"},
			Want:          `Allowed`,
			StatusCode:    http.StatusOK,
		},
		{
			Name:          "auth ok 2",
			Method:        http.MethodGet,
			Uri:           "/",
			RequestHeader: map[string]string{"X-API-Key": "123456789def"},
			Want:          `Allowed`,
			StatusCode:    http.StatusOK,
		},
	}
	utils.RunTests(t, ts.URL, tt)
}

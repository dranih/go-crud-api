package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/utils"
	"github.com/gorilla/mux"
)

func TestAjaxOnlyMiddleware(t *testing.T) {
	properties := map[string]interface{}{}

	router := mux.NewRouter()
	responder := controller.NewJsonResponder(false)
	aoMiddle := NewAjaxOnlyMiddleware(responder, properties)
	router.HandleFunc("/", allowedTest).Methods("GET", "POST")
	router.Use(aoMiddle.Process)
	ts := httptest.NewServer(router)
	defer ts.Close()

	tt := []utils.Test{
		{
			Name:       "forbidden ",
			Method:     http.MethodPost,
			Uri:        "/",
			Want:       `{"code":1018,"message":"Only AJAX requests allowed for 'POST'"}`,
			StatusCode: http.StatusForbidden,
		},
		{
			Name:       "allowed get ",
			Method:     http.MethodGet,
			Uri:        "/",
			WantRegex:  `Allowed`,
			StatusCode: http.StatusOK,
		},
		{
			Name:          "allowed post with header ",
			Method:        http.MethodPost,
			Uri:           "/",
			RequestHeader: map[string]string{"X-Requested-With": "XMLHttpRequest"},
			Want:          `Allowed`,
			StatusCode:    http.StatusOK,
		},
	}
	utils.RunTests(t, ts.URL, tt)
}

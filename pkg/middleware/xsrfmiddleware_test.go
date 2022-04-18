package middleware

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/utils"
	"github.com/gorilla/mux"
)

func TestXsrfMiddleware(t *testing.T) {
	properties := map[string]interface{}{}

	router := mux.NewRouter()
	responder := controller.NewJsonResponder(false)
	xMiddle := NewXsrfMiddleware(responder, properties)
	router.HandleFunc("/", utils.AllowedTest).Methods("GET", "POST")
	router.Use(xMiddle.Process)
	ts := httptest.NewServer(router)
	defer ts.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Errorf("Got error while creating cookie jar %s", err.Error())
	}

	tt := []utils.Test{
		{
			Name:       "forbidden ",
			Method:     http.MethodPost,
			Uri:        "/",
			Want:       `{"code":1017,"message":"Bad or missing XSRF token"}`,
			Jar:        jar,
			StatusCode: http.StatusForbidden,
		},
		{
			Name:       "get cookie ",
			Method:     http.MethodGet,
			Uri:        "/",
			WantRegex:  `Allowed`,
			Jar:        jar,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "forbidden ",
			Method:     http.MethodPost,
			Uri:        "/",
			Want:       `{"code":1017,"message":"Bad or missing XSRF token"}`,
			Jar:        jar,
			StatusCode: http.StatusForbidden,
		},
	}
	utils.RunTests(t, ts.URL, tt)

	rheader := map[string]string{}
	u, _ := url.Parse(ts.URL)
	for _, cookie := range jar.Cookies(u) {
		if cookie.Name == "XSRF-TOKEN" {
			rheader["X-XSRF-TOKEN"] = cookie.Value
		}
	}

	tt2 := []utils.Test{
		{
			Name:          "allowed ",
			Method:        http.MethodPost,
			Uri:           "/",
			Want:          `Allowed`,
			Jar:           jar,
			RequestHeader: rheader,
			StatusCode:    http.StatusOK,
		},
		{
			Name:       "forbidden ",
			Method:     http.MethodPost,
			Uri:        "/",
			Want:       `{"code":1017,"message":"Bad or missing XSRF token"}`,
			Jar:        jar,
			StatusCode: http.StatusForbidden,
		},
	}
	utils.RunTests(t, ts.URL, tt2)
}

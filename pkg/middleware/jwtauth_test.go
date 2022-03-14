package middleware

import (
	"encoding/gob"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/utils"
	"github.com/gorilla/mux"
)

func TestJwtAuth(t *testing.T) {
	properties := map[string]interface{}{
		"mode":    "required",
		"realm":   "GoCrudApi : JWT required",
		"time":    "1538207605",
		"secrets": "axpIrCGNGqxzx2R9dtXLIPUSqPo778uhb8CA0F4Hx",
	}

	router := mux.NewRouter()
	responder := controller.NewJsonResponder(false)
	jaMiddle := NewJwtAuth(responder, properties)
	saveSessionMiddle := NewSaveSession(responder, nil)
	router.HandleFunc("/", allowedTest).Methods("GET")
	router.Use(jaMiddle.Process)
	router.Use(saveSessionMiddle.Process)
	gob.Register(map[string]interface{}{})
	ts := httptest.NewServer(router)
	defer ts.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Errorf("Got error while creating cookie jar %s", err.Error())
	}

	tt := []utils.Test{
		{
			Name:       "forbidden ",
			Method:     http.MethodGet,
			Uri:        "/",
			Want:       `{"code":1011,"details":"GoCrudApi : JWT required","message":"Authentication required"}`,
			StatusCode: http.StatusUnauthorized,
		},
		{
			Name:          "bad key",
			Method:        http.MethodGet,
			Uri:           "/",
			RequestHeader: map[string]string{"X-Authorization": "Bearer invalid"},
			Jar:           jar,
			Want:          `{"code":1012,"message":"Authentication failed for 'JWT'"}`,
			StatusCode:    http.StatusForbidden,
		},
		{
			Name:          "auth ok 1",
			Method:        http.MethodGet,
			Uri:           "/",
			RequestHeader: map[string]string{"X-Authorization": "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6IjE1MzgyMDc2MDUiLCJleHAiOjE1MzgyMDc2MzV9.Z5px_GT15TRKhJCTHhDt5Z6K6LRDSFnLj8U5ok9l7gw"},
			Jar:           jar,
			Want:          `Allowed`,
			StatusCode:    http.StatusOK,
		},
		{
			Name:       "auth ok 2",
			Method:     http.MethodGet,
			Uri:        "/",
			Jar:        jar,
			Want:       `Allowed`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "no cookie ;-(",
			Method:     http.MethodGet,
			Uri:        "/",
			Body:       ``,
			Want:       `{"code":1011,"details":"GoCrudApi : JWT required","message":"Authentication required"}`,
			StatusCode: http.StatusUnauthorized,
		},
	}
	utils.RunTests(t, ts.URL, tt)
}

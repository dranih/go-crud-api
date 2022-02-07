package middleware

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/utils"
	"github.com/gorilla/mux"
)

func TestBasicAuth(t *testing.T) {
	properties := map[string]interface{}{
		"mode":         "required",
		"realm":        "GoCrudApi : Username and password required",
		"passwordFile": "../../test/test.pwd",
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Errorf("Got error while creating cookie jar %s", err.Error())
	}

	router := mux.NewRouter()
	responder := controller.NewJsonResponder(false)
	bamMiddle := NewBasicAuth(responder, properties)
	router.HandleFunc("/", allowedTest).Methods("GET")
	router.Use(bamMiddle.Process)
	ts := httptest.NewServer(router)
	defer ts.Close()

	tt := []utils.Test{
		{
			Name:       "forbidden ",
			Method:     http.MethodGet,
			Uri:        "/",
			Body:       ``,
			WantRegex:  `{"code":1011,"details":"GoCrudApi : Username and password required","message":"Authentication required.*`,
			StatusCode: http.StatusUnauthorized,
		},
		{
			Name:       "bad user",
			Method:     http.MethodGet,
			Uri:        "/",
			Body:       ``,
			Want:       `{"code":1012,"message":"Authentication failed for 'user10'"}`,
			AuthMethod: `basicauth`,
			Username:   `user10`,
			Password:   `MyPwd01`,
			StatusCode: http.StatusForbidden,
		},
		{
			Name:       "bad pwd",
			Method:     http.MethodGet,
			Uri:        "/",
			Body:       ``,
			WantRegex:  `{"code":1012,"message":"Authentication failed for 'user1'"}`,
			AuthMethod: `basicauth`,
			Username:   `user1`,
			Password:   `MyPwd011`,
			StatusCode: http.StatusForbidden,
		},
		{
			Name:       "auth ok",
			Method:     http.MethodGet,
			Uri:        "/",
			Body:       ``,
			Want:       `Allowed`,
			AuthMethod: `basicauth`,
			Username:   `user1`,
			Password:   `MyPwd01`,
			Jar:        jar,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "using cookie",
			Method:     http.MethodGet,
			Uri:        "/",
			Body:       ``,
			Want:       `Allowed`,
			Jar:        jar,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "no cookie ;-(",
			Method:     http.MethodGet,
			Uri:        "/",
			Body:       ``,
			WantRegex:  `{"code":1011,"details":"GoCrudApi : Username and password required","message":"Authentication required.*`,
			StatusCode: http.StatusUnauthorized,
		},
	}
	utils.RunTests(t, ts.URL, tt)
}

func allowedTest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Allowed")
}

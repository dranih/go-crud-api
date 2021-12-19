package controller

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/gorilla/mux"
)

func TestColumnController(t *testing.T) {
	db := database.NewGenericDB(
		"sqlite",
		"../../test/test.db",
		0,
		"test",
		map[string]bool{"cows": true, "sharks": true},
		"",
		"",
	)
	var cache interface{}
	reflection := database.NewReflectionService(db, cache, 0)
	responder := NewJsonResponder(false)
	definition := database.NewDefinitionService(db, reflection)
	router := mux.NewRouter()
	NewColumnController(router, responder, reflection, definition)
	ts := httptest.NewServer(router)
	defer ts.Close()

	//https://ieftimov.com/post/testing-in-go-testing-http-servers/
	//https://stackoverflow.com/questions/42474259/golang-how-to-live-test-an-http-server
	tt := []struct {
		name       string
		method     string
		uri        string
		body       string
		want       string
		wantRegex  string
		statusCode int
	}{
		{
			name:       "get tables and columns ",
			method:     http.MethodGet,
			uri:        "/columns",
			body:       ``,
			wantRegex:  `\{"tables":\[\{"columns":.*\}\]\}`,
			statusCode: http.StatusOK,
		},
		{
			name:       "get 1 table and columns ",
			method:     http.MethodGet,
			uri:        "/columns/cows",
			body:       ``,
			wantRegex:  `\{"columns":\[.*\],"name":"cows","type":"table"\}`,
			statusCode: http.StatusOK,
		},
		{
			name:       "get inexistant table ",
			method:     http.MethodGet,
			uri:        "/columns/doesnotexists",
			body:       ``,
			want:       "{\"code\":1001,\"details\":\"\",\"message\":\"Table `doesnotexists` not found\"}",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "get 1 table 1 column ",
			method:     http.MethodGet,
			uri:        "/columns/cows/length",
			body:       ``,
			wantRegex:  `\{.*"name":"length",.*\}`,
			statusCode: http.StatusOK,
		},
		{
			name:       "get inexistant column ",
			method:     http.MethodGet,
			uri:        "/columns/cows/doesnotexists",
			body:       ``,
			want:       "{\"code\":1005,\"details\":\"\",\"message\":\"Column `doesnotexists` not found\"}",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "update column ",
			method:     http.MethodPut,
			uri:        "/columns/cows",
			body:       `{"name":"cows2"}`,
			want:       "true",
			statusCode: http.StatusOK,
		},
		{
			name:       "update (back) column - test refresh tables",
			method:     http.MethodPut,
			uri:        "/columns/cows2",
			body:       `{"name":"cows"}`,
			want:       "true",
			statusCode: http.StatusOK,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request, err := http.NewRequest(tc.method, ts.URL+tc.uri, strings.NewReader(tc.body))
			if err != nil {
				t.Fatal(err)
			}

			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.statusCode {
				t.Errorf("Want status '%d', got '%d' at url '%s'", tc.statusCode, resp.StatusCode, resp.Request.URL)
			}
			b, err := io.ReadAll(resp.Body)
			if tc.wantRegex != "" {
				re, _ := regexp.Compile(tc.wantRegex)
				if !re.Match(b) {
					t.Errorf("Regex '%s' not matching, got '%s'", tc.wantRegex, b)
				}
			} else if strings.TrimSpace(string(b)) != tc.want {
				t.Errorf("Want '%s', got '%s'", tc.want, b)
			}
		})
	}
}

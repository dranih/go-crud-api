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

func TestRecordController(t *testing.T) {
	db := database.NewGenericDB(
		"sqlite",
		"../../test/test.db",
		0,
		"test",
		nil,
		"",
		"",
	)
	var cache interface{}
	reflection := database.NewReflectionService(db, cache, 0)
	records := database.NewRecordService(db, reflection)
	responder := NewJsonResponder(false)
	router := mux.NewRouter()
	NewRecordController(router, responder, records)
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
			name:       "get table ",
			method:     http.MethodGet,
			uri:        "/records/sharks",
			body:       ``,
			wantRegex:  `"sharktype":"Megaladon"`,
			statusCode: http.StatusOK,
		},
		{
			name:       "get unique id ",
			method:     http.MethodGet,
			uri:        "/records/sharks/3",
			body:       ``,
			want:       `{"id":3,"length":1800,"name":"Himari","sharktype":"Megaladon"}`,
			statusCode: http.StatusOK,
		},
		{
			name:       "get multiple ids ",
			method:     http.MethodGet,
			uri:        "/records/sharks/1,3",
			body:       ``,
			want:       `[{"id":1,"length":427,"name":"Sammy","sharktype":"Greenland Shark"},{"id":3,"length":1800,"name":"Himari","sharktype":"Megaladon"}]`,
			statusCode: http.StatusOK,
		},
		{
			name:       "post unique ",
			method:     http.MethodPost,
			uri:        "/records/sharks",
			body:       `{"name":"Tomy","length": "100","sharktype": "Great White Shark"}`,
			wantRegex:  `{"id":[0-9]+}`,
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
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, resp.StatusCode)
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

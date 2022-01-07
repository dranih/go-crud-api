package controller

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/dranih/go-crud-api/pkg/cache"
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
			name:       "ping ",
			method:     http.MethodGet,
			uri:        "/status/ping",
			body:       ``,
			wantRegex:  `{"cache":[0-9]+,"db":[0-9]+}`,
			statusCode: http.StatusOK,
		},
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
			body:       `{"id":99,"name":"Tomy","length": "100","sharktype": "Great White Shark"}`,
			wantRegex:  `{"id":[0-9]+}`,
			statusCode: http.StatusOK,
		},
		{
			name:       "put unique ",
			method:     http.MethodPut,
			uri:        "/records/sharks/99",
			body:       `{"length": 2000}`,
			want:       `{"RowsAffected":1}`,
			statusCode: http.StatusOK,
		},
		{
			name:       "patch unique ",
			method:     http.MethodPatch,
			uri:        "/records/sharks/99",
			body:       `{"length": 10}`,
			want:       `{"RowsAffected":1}`,
			statusCode: http.StatusOK,
		},
		{
			name:       "delete unique ",
			method:     http.MethodDelete,
			uri:        "/records/sharks/99",
			body:       ``,
			want:       `{"RowsAffected":1}`,
			statusCode: http.StatusOK,
		},
		{
			name:       "post multiple ",
			method:     http.MethodPost,
			uri:        "/records/sharks",
			body:       `[{"id":99,"name":"Tomy","length": "100","sharktype": "Great White Shark"},{"id":999,"name":"Barbara","length": "150","sharktype": "Hammer head"}]`,
			wantRegex:  `[{"id":[0-9]+},{"id":[0-9]+}]`,
			statusCode: http.StatusOK,
		},
		{
			name:       "put multiples ",
			method:     http.MethodPut,
			uri:        "/records/sharks/99,999",
			body:       `[{"length": 2000},{"name": "Barbara3","length": 1000}]`,
			want:       `[{"RowsAffected":1},{"RowsAffected":1}]`,
			statusCode: http.StatusOK,
		},
		{
			name:       "patch multiple ",
			method:     http.MethodPatch,
			uri:        "/records/sharks/99,999",
			body:       `[{"length": 10},{"length": 50}]`,
			want:       `[{"RowsAffected":1},{"RowsAffected":1}]`,
			statusCode: http.StatusOK,
		},
		{
			name:       "delete multiple ",
			method:     http.MethodDelete,
			uri:        "/records/sharks/99,999",
			body:       ``,
			want:       `[{"RowsAffected":1},{"RowsAffected":1}]`,
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

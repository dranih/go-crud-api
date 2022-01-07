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

func TestCacheController(t *testing.T) {
	db := database.NewGenericDB(
		"sqlite",
		"../../test/test.db",
		0,
		"test",
		map[string]bool{"cows": true, "sharks": true},
		"",
		"",
	)
	prefix := fmt.Sprintf("gocrudapi-%d-", os.Getpid())
	cache := cache.Create("Gocache", prefix, "")
	reflection := database.NewReflectionService(db, cache, 10)
	responder := NewJsonResponder(false)
	definition := database.NewDefinitionService(db, reflection)
	router := mux.NewRouter()
	NewColumnController(router, responder, reflection, definition)
	NewCacheController(router, responder, cache)
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
			name:       "clear cache",
			method:     http.MethodGet,
			uri:        "/cache/clear",
			body:       ``,
			want:       `true`,
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

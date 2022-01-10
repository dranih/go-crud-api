package controller

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dranih/go-crud-api/pkg/cache"
	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/geojson"
	"github.com/dranih/go-crud-api/pkg/utils"
	"github.com/gorilla/mux"
)

//No geometry with sqllite, should mock DB
func TestGeoJsonController(t *testing.T) {
	db := database.NewGenericDB(
		"sqlite",
		"../../test/test.db",
		0,
		"test",
		map[string]bool{"countries": true},
		"",
		"",
	)
	prefix := fmt.Sprintf("gocrudapi-%d-", os.Getpid())
	cache := cache.Create("TempFile", prefix, "")
	reflection := database.NewReflectionService(db, cache, 10)
	responder := NewJsonResponder(false)
	router := mux.NewRouter()
	records := database.NewRecordService(db, reflection)
	geoJson := geojson.NewGeoJsonService(reflection, records)
	NewGeoJsonController(router, responder, geoJson)
	ts := httptest.NewServer(router)
	defer ts.Close()

	tt := []utils.Test{
		{
			Name:       "get geojson ",
			Method:     http.MethodGet,
			Uri:        "/geojson/countries/3",
			Body:       ``,
			Want:       `{"type":"Feature","id":3,"properties":{"name":"Point"},"geometry":{"type":"Point","coordinates":[30,10]}}`,
			StatusCode: http.StatusOK,
		},
	}
	utils.RunTests(t, ts, tt)
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/utils"
	"github.com/gorilla/mux"
)

func TestFirewallMiddleware(t *testing.T) {
	properties := map[string]interface{}{
		"allowedIpAddresses": "1.2.3.4,5.6.7.8,172.17.0.0/16",
	}

	router := mux.NewRouter()
	responder := controller.NewJsonResponder(false)
	fwMiddle := NewFirewallMiddleware(responder, properties)
	router.HandleFunc("/", allowedTest).Methods("GET")
	router.Use(fwMiddle.Process)
	ts := httptest.NewServer(router)
	defer ts.Close()

	properties2 := map[string]interface{}{
		"allowedIpAddresses": "127.0.0.0/8",
	}

	router2 := mux.NewRouter()
	responder2 := controller.NewJsonResponder(false)
	fwMiddle2 := NewFirewallMiddleware(responder2, properties2)
	router2.HandleFunc("/", allowedTest).Methods("GET")
	router2.Use(fwMiddle2.Process)
	ts2 := httptest.NewServer(router2)
	defer ts2.Close()

	tt := []utils.Test{
		{
			Name:       "forbidden ",
			Method:     http.MethodGet,
			Uri:        "/",
			Want:       `{"code":1016,"message":"Temporary or permanently blocked"}`,
			StatusCode: http.StatusForbidden,
		},
	}

	tt2 := []utils.Test{
		{
			Name:       "forbidden ",
			Method:     http.MethodGet,
			Uri:        "/",
			Want:       `Allowed`,
			StatusCode: http.StatusOK,
		},
	}

	utils.RunTests(t, ts.URL, tt)
	utils.RunTests(t, ts2.URL, tt2)
}

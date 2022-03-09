package apiserver

import (
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/dranih/go-crud-api/pkg/utils"
)

// Global API tests for records
// To check compatibility with php-crud-api
func TestCacheApi(t *testing.T) {
	utils.SelectConfig()
	config := ReadConfig()
	config.Init()
	serverStarted := new(sync.WaitGroup)
	serverStarted.Add(1)
	api := NewApi(config)
	go api.Handle(serverStarted)
	//Waiting http server to start
	serverStarted.Wait()
	serverUrlHttps := fmt.Sprintf("https://%s:%d", config.Server.Address, config.Server.HttpsPort)

	//https://ieftimov.com/post/testing-in-go-testing-http-servers/
	//https://stackoverflow.com/questions/42474259/golang-how-to-live-test-an-http-server
	tt := []utils.Test{
		{
			Name:       "001_clear_cache",
			Method:     http.MethodGet,
			Uri:        "/cache/clear",
			Body:       ``,
			Want:       `true`,
			StatusCode: http.StatusOK,
		},
	}
	utils.RunTests(t, serverUrlHttps, tt)
}

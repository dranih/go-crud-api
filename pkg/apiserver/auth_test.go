package apiserver

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"testing"

	"github.com/dranih/go-crud-api/pkg/utils"
)

// Global API tests for records
// To check compatibility with php-crud-api
func TestAuthApi(t *testing.T) {
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
	//serverUrlHttp := fmt.Sprintf("http://%s:%d", config.Server.Address, config.Server.HttpPort)
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Errorf("Got error while creating cookie jar %s", err.Error())
	}

	//https://ieftimov.com/post/testing-in-go-testing-http-servers/
	//https://stackoverflow.com/questions/42474259/golang-how-to-live-test-an-http-server
	tt := []utils.Test{
		{
			Name:       "002_basic_auth_A",
			Method:     http.MethodGet,
			Uri:        "/records/invisibles/e42c77c6-06a4-4502-816c-d112c7142e6d",
			Body:       ``,
			AuthMethod: `basicauth`,
			Username:   `username1`,
			Password:   `password1`,
			Jar:        jar,
			Want:       `{"id":"e42c77c6-06a4-4502-816c-d112c7142e6d"}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "002_basic_auth_B",
			Method:     http.MethodGet,
			Uri:        "/records/invisibles/e42c77c6-06a4-4502-816c-d112c7142e6d",
			Body:       ``,
			Jar:        jar,
			Want:       `{"id":"e42c77c6-06a4-4502-816c-d112c7142e6d"}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "002_basic_auth_C",
			Method:     http.MethodGet,
			Uri:        "/records/invisibles/e42c77c6-06a4-4502-816c-d112c7142e6d",
			Body:       ``,
			AuthMethod: `basicauth`,
			Username:   `invaliduser`,
			Password:   `invalidpass`,
			Want:       `{"code":1012,"message":"Authentication failed for 'invaliduser'"}`,
			StatusCode: http.StatusForbidden,
		},
		{
			Name:       "002_basic_auth_D",
			Method:     http.MethodGet,
			Uri:        "/records/invisibles/e42c77c6-06a4-4502-816c-d112c7142e6d",
			Body:       ``,
			Want:       `{"code":1001,"message":"Table 'invisibles' not found"}`,
			StatusCode: http.StatusNotFound,
		},
	}
	utils.RunTests(t, serverUrlHttps, tt)

}

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
			Name:          "001_jwt_auth_A",
			Method:        http.MethodGet,
			Uri:           "/records/invisibles/e42c77c6-06a4-4502-816c-d112c7142e6d",
			Body:          ``,
			RequestHeader: map[string]string{"X-Authorization": "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6IjE1MzgyMDc2MDUiLCJleHAiOjE1MzgyMDc2MzV9.Z5px_GT15TRKhJCTHhDt5Z6K6LRDSFnLj8U5ok9l7gw"},
			Jar:           jar,
			Want:          `{"id":"e42c77c6-06a4-4502-816c-d112c7142e6d"}`,
			StatusCode:    http.StatusOK,
		},
		{
			Name:       "001_jwt_auth_B",
			Method:     http.MethodGet,
			Uri:        "/records/invisibles/e42c77c6-06a4-4502-816c-d112c7142e6d",
			Body:       ``,
			Jar:        jar,
			Want:       `{"id":"e42c77c6-06a4-4502-816c-d112c7142e6d"}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:          "001_jwt_auth_C",
			Method:        http.MethodGet,
			Uri:           "/records/invisibles/e42c77c6-06a4-4502-816c-d112c7142e6d",
			Body:          ``,
			RequestHeader: map[string]string{"X-Authorization": "Bearer invalid"},
			Jar:           jar,
			Want:          `{"code":1012,"message":"Authentication failed for 'JWT'"}`,
			StatusCode:    http.StatusForbidden,
		},
		{
			Name:       "001_jwt_auth_D",
			Method:     http.MethodGet,
			Uri:        "/records/invisibles/e42c77c6-06a4-4502-816c-d112c7142e6d",
			Body:       ``,
			Jar:        jar,
			Want:       `{"code":1001,"message":"Table 'invisibles' not found"}`,
			StatusCode: http.StatusNotFound,
		},
		{
			Name:          "001_jwt_auth_E",
			Method:        http.MethodOptions,
			Uri:           "/records/invisibles/e42c77c6-06a4-4502-816c-d112c7142e6d",
			Body:          ``,
			Jar:           jar,
			RequestHeader: map[string]string{"Access-Control-Request-Method": "POST", "Access-Control-Request-Headers": "X-PINGOTHER, Content-Type"},
			WantHeader: map[string]string{"Access-Control-Allow-Headers": "Content-Type, X-XSRF-TOKEN, X-Authorization",
				"Access-Control-Allow-Methods":     "OPTIONS, GET, PUT, POST, DELETE, PATCH",
				"Access-Control-Allow-Credentials": "true",
				"Access-Control-Max-Age":           "1728000"},
			StatusCode: http.StatusOK,
		},
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
		{
			Name:       "003_db_auth_A",
			Method:     http.MethodGet,
			Uri:        "/records/invisibles/e42c77c6-06a4-4502-816c-d112c7142e6d",
			Want:       `{"code":1001,"message":"Table 'invisibles' not found"}`,
			StatusCode: http.StatusNotFound,
		},
		{
			Name:       "003_db_auth_B",
			Method:     http.MethodPost,
			Uri:        "/login",
			Body:       `{"username":"user2","password":"pass2"}`,
			Want:       `{"id":2,"username":"user2"}`,
			Jar:        jar,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "003_db_auth_C",
			Method:     http.MethodGet,
			Uri:        "/me",
			Want:       `{"id":2,"username":"user2"}`,
			Jar:        jar,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "003_db_auth_D",
			Method:     http.MethodGet,
			Uri:        "/records/invisibles/e42c77c6-06a4-4502-816c-d112c7142e6d",
			Want:       `{"id":"e42c77c6-06a4-4502-816c-d112c7142e6d"}`,
			Jar:        jar,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "003_db_auth_E",
			Method:     http.MethodPost,
			Uri:        "/login",
			Body:       `{"username":"user2","password":"incorect password"}`,
			Want:       `{"code":1012,"message":"Authentication failed for 'user2'"}`,
			Jar:        jar,
			StatusCode: http.StatusForbidden,
		},
		{
			Name:       "003_db_auth_F",
			Method:     http.MethodGet,
			Uri:        "/records/invisibles/e42c77c6-06a4-4502-816c-d112c7142e6d",
			Want:       `{"id":"e42c77c6-06a4-4502-816c-d112c7142e6d"}`,
			Jar:        jar,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "003_db_auth_G",
			Method:     http.MethodPost,
			Uri:        "/logout",
			Want:       `{"id":2,"username":"user2"}`,
			Jar:        jar,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "003_db_auth_H",
			Method:     http.MethodGet,
			Uri:        "/records/invisibles/e42c77c6-06a4-4502-816c-d112c7142e6d",
			Want:       `{"code":1001,"message":"Table 'invisibles' not found"}`,
			Jar:        jar,
			StatusCode: http.StatusNotFound,
		},
		{
			Name:       "003_db_auth_I",
			Method:     http.MethodPost,
			Uri:        "/logout",
			Want:       `{"code":1011,"message":"Authentication required"}`,
			Jar:        jar,
			StatusCode: http.StatusUnauthorized,
		},
		{
			Name:       "003_db_auth_J",
			Method:     http.MethodPost,
			Uri:        "/register",
			Body:       `{"username":"user2","password":""}`,
			Want:       `{"code":1021,"message":"Password too short (<4 characters)"}`,
			Jar:        jar,
			StatusCode: http.StatusUnprocessableEntity,
		},
		{
			Name:       "003_db_auth_K",
			Method:     http.MethodPost,
			Uri:        "/register",
			Body:       `{"username":"user2","password":"pass2"}`,
			Want:       `{"code":1020,"message":"User 'user2' already exists"}`,
			Jar:        jar,
			StatusCode: http.StatusConflict,
		},
		{
			Name:       "003_db_auth_L",
			Method:     http.MethodPost,
			Uri:        "/register",
			Body:       `{"username":"user3","password":"pass3"}`,
			Want:       `{"id":3,"username":"user3"}`,
			Jar:        jar,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "003_db_auth_M",
			Method:     http.MethodPost,
			Uri:        "/login",
			Body:       `{"username":"user3","password":"pass3"}`,
			Want:       `{"id":3,"username":"user3"}`,
			Jar:        jar,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "003_db_auth_N",
			Method:     http.MethodGet,
			Uri:        "/me",
			Want:       `{"id":3,"username":"user3"}`,
			Jar:        jar,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "003_db_auth_O",
			Method:     http.MethodPost,
			Uri:        "/password",
			Body:       `{"username":"user3","password":"pass3","newPassword":"secret3"}`,
			Want:       `{"id":3,"username":"user3"}`,
			Jar:        jar,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "003_db_auth_P",
			Method:     http.MethodPost,
			Uri:        "/logout",
			Want:       `{"id":3,"username":"user3"}`,
			Jar:        jar,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "003_db_auth_Q",
			Method:     http.MethodPost,
			Uri:        "/login",
			Body:       `{"username":"user3","password":"secret3"}`,
			Want:       `{"id":3,"username":"user3"}`,
			Jar:        jar,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "003_db_auth_R",
			Method:     http.MethodGet,
			Uri:        "/me",
			Want:       `{"id":3,"username":"user3"}`,
			Jar:        jar,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "003_db_auth_S",
			Method:     http.MethodPost,
			Uri:        "/logout",
			Want:       `{"id":3,"username":"user3"}`,
			Jar:        jar,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "003_db_auth_T",
			Method:     http.MethodDelete,
			Uri:        "/records/users/3",
			Want:       `1`,
			Jar:        jar,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "004_api_key_auth_A",
			Method:     http.MethodGet,
			Uri:        "/records/invisibles/e42c77c6-06a4-4502-816c-d112c7142e6d",
			Want:       `{"code":1001,"message":"Table 'invisibles' not found"}`,
			StatusCode: http.StatusNotFound,
		},
		{
			Name:          "004_api_key_auth_B",
			Method:        http.MethodGet,
			Uri:           "/records/invisibles/e42c77c6-06a4-4502-816c-d112c7142e6d",
			RequestHeader: map[string]string{"X-API-Key": "123456789abc"},
			Want:          `{"id":"e42c77c6-06a4-4502-816c-d112c7142e6d"}`,
			StatusCode:    http.StatusOK,
		},
		{
			Name:       "005_api_key_db_auth_A",
			Method:     http.MethodGet,
			Uri:        "/records/invisibles/e42c77c6-06a4-4502-816c-d112c7142e6d",
			Want:       `{"code":1001,"message":"Table 'invisibles' not found"}`,
			StatusCode: http.StatusNotFound,
		},
		{
			Name:          "005_api_key_db_auth_B",
			Method:        http.MethodGet,
			Uri:           "/records/invisibles/e42c77c6-06a4-4502-816c-d112c7142e6d",
			RequestHeader: map[string]string{"X-API-Key-DB": "123456789abc"},
			Want:          `{"id":"e42c77c6-06a4-4502-816c-d112c7142e6d"}`,
			StatusCode:    http.StatusOK,
		},
	}
	utils.RunTests(t, serverUrlHttps, tt)

}

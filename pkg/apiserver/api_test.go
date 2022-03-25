package apiserver

import (
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/dranih/go-crud-api/pkg/utils"
)

func TestNewApi(t *testing.T) {
	db_path := utils.SelectConfig(false)
	config := ReadConfig()
	config.Init()
	if db_path != "" && config.Api.Driver == "sqlite" {
		config.Api.Address = db_path
	}
	serverStarted := new(sync.WaitGroup)
	serverStarted.Add(1)
	api := NewApi(config)
	go api.Handle(serverStarted)
	//Waiting http server to start
	serverStarted.Wait()
	serverUrlHttps := fmt.Sprintf("https://%s:%d", config.Server.Address, config.Server.HttpsPort)

	tt := []utils.Test{
		{
			Name:       "ping ",
			Method:     http.MethodGet,
			Uri:        "/status/ping",
			Body:       ``,
			WantRegex:  `{"cache":[0-9]+,"db":[0-9]+}`,
			StatusCode: http.StatusOK,
		},
	}
	utils.RunTests(t, serverUrlHttps, tt)
}

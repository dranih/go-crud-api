package apiserver

import (
	"sync"
	"testing"
)

func TestNewApi(t *testing.T) {
	config := ReadConfig("../../test/")
	config.Init()
	serverStarted := new(sync.WaitGroup)
	serverStarted.Add(1)
	api := NewApi(config.Api)
	go api.Handle(config.Server, serverStarted)
	//Waiting http server to start
	serverStarted.Wait()
	/*
		url := "http://" + config.Server.Address + ":" + fmt.Sprint(config.Server.Port)
		response := map[string]int{}
		testutils.TestGetUrl(t, url, &response)
		if val, exists := response["db"]; !exists {
			t.Error("No db field in response data")
		} else if val > 50 {
			t.Error("DB ping too long")
		}*/

}

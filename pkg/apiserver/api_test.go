package apiserver

import (
	"sync"
	"testing"

	"github.com/dranih/go-crud-api/pkg/utils"
)

func TestNewApi(t *testing.T) {
	utils.SelectConfig()
	config := ReadConfig()
	config.Init()
	serverStarted := new(sync.WaitGroup)
	serverStarted.Add(1)
	api := NewApi(config)
	go api.Handle(serverStarted)
	//Waiting http server to start
	serverStarted.Wait()
}

package apiserver

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

func TestNewApi(t *testing.T) {
	config := ReadConfig("../../test/")
	config.Init()
	api := NewApi(config.Api)
	url := "http://" + config.Server.Address + ":" + fmt.Sprint(config.Server.Port)
	log.Printf(url)

	go func() { api.Handle(config.Server) }()
	res, err := http.Get(url + "/status/ping")
	if err != nil {
		log.Printf("RES : %v", res)
		log.Printf("ERR : %v", err)
		t.Error(err, "unable to complete Get request")
		return
	}
	defer res.Body.Close()
	out, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err, "unable to read response data")
	}
	log.Printf("here")
	log.Printf("%v", out)
	//Stopchan <- syscall.SIGINT
	log.Printf("there")
}

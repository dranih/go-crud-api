package main

import (
	"github.com/dranih/go-crud-api/pkg/apiserver"
)

func main() {
	config := apiserver.ReadConfig()
	config.Init()
	api := apiserver.NewApi(config.Api)
	api.Handle(config.Server)
}

package main

import (
	"github.com/dranih/go-crud-api/pkg/apiserver"
)

func main() {
	config := &apiserver.Config{
		Api: &apiserver.ApiConfig{
			Driver:   "sqlite",
			Address:  "../../test/test.db",
			Database: "test",
			Tables:   "sharks",
		},
	}
	config.SetDefaults()
	api := apiserver.NewApi(config.Api)
	api.Handle(config.Server)
}

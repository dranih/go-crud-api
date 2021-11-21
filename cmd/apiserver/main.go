package main

import (
	"github.com/dranih/go-crud-api/pkg/apiserver"
)

func main() {
	config := apiserver.NewConfig()
	config.Driver = "sqlite"
	config.Address = "../../test/test.db"
	config.Database = "test"
	config.Tables = "sharks"
	api := apiserver.NewApi(config)
	api.Handle()
}

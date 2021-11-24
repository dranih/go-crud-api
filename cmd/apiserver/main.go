package main

import (
	"fmt"
	"strings"

	"github.com/dranih/go-crud-api/pkg/apiserver"
	"github.com/spf13/viper"
)

func main() {

	viper.SetConfigName("gcaconfig")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")
	viper.SetConfigType("yml")
	viper.SetEnvPrefix("gca")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()
	var config apiserver.Config

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}

	// Set undefined variables
	viper.SetDefault("api.driver", "mysql")
	viper.SetDefault("api.middlewares", "records,geojson,openapi,status")
	viper.SetDefault("api.cachetype", "TempFile")
	viper.SetDefault("api.cachetime", 10)
	viper.SetDefault("api.openapibase", `{"info":{"title":"GO-CRUD-API","version":"0.0.1"}}`)
	viper.SetDefault("server.address", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.gracefultimeout", 15)
	viper.SetDefault("server.writetimeout", 15)
	viper.SetDefault("server.readtimeout", 15)
	viper.SetDefault("server.idletimeout", 60)

	err := viper.Unmarshal(&config)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}

	config.Api.SetDriverDefaults()
	api := apiserver.NewApi(config.Api)
	api.Handle(config.Server)
}

package apiserver

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Api    *ApiConfig
	Server *ServerConfig
}

type ApiConfig struct {
	Driver                string
	Address               string
	Port                  int
	Username              string
	Password              string
	Database              string
	Tables                string
	Middlewares           map[string]map[string]interface{}
	Controllers           string
	CustomControllers     string
	CustomOpenApiBuilders string
	CacheType             string
	CachePath             string
	CacheTime             int32
	Debug                 bool
	BasePath              string
	OpenApiBase           string
}

type ServerConfig struct {
	Address         string
	Port            int
	GracefulTimeout int
	WriteTimeout    int
	ReadTimeout     int
	IdleTimeout     int
}

func ReadConfig(configPaths ...string) *Config {
	for _, configPath := range configPaths {
		viper.AddConfigPath(configPath)
	}
	viper.SetConfigName("gcaconfig")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")
	viper.SetConfigType("yml")
	viper.SetEnvPrefix("gca")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()
	var config Config

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}

	// Set undefined variables
	viper.SetDefault("api.driver", "mysql")
	viper.SetDefault("api.controllers", "records,geojson,openapi,status")
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
		panic(fmt.Sprintf("Unable to decode into struct, %v", err))
	}

	return &config
}

func (ac *ApiConfig) getDefaultPort(driver string) int {
	switch driver {
	case "mysql":
		return 3306
	case "pgsql":
		return 5432
	case "sqlsrv":
		return 1433
	case "sqlite":
		return 0
	default:
		return -1
	}
}

func (ac *ApiConfig) getDefaultAddress(driver string) string {
	switch driver {
	case "mysql":
		return "localhost"
	case "pgsql":
		return "localhost"
	case "sqlsrv":
		return "localhost"
	case "sqlite":
		return "data.db"
	default:
		return ""
	}
}

func (ac *ApiConfig) getDriverDefaults(driver string) map[string]interface{} {
	return map[string]interface{}{
		"driver":  driver,
		"address": ac.getDefaultAddress(driver),
		"port":    ac.getDefaultPort(driver),
	}
}

func (ac *ApiConfig) setDriverDefaults() {
	defaults := ac.getDriverDefaults(ac.Driver)
	if ac.Address == "" {
		ac.Address = fmt.Sprint(defaults["address"])
	}
	if ac.Port == 0 {
		ac.Port, _ = defaults["port"].(int)
	}
}

func (c *Config) Init() {
	c.Api.setDriverDefaults()
	c.Api.initMiddlewares()
}

func (ac *ApiConfig) initMiddlewares() {
	defaultMiddlewares := "cors,errors"
	if ac.Middlewares == nil {
		ac.Middlewares = map[string]map[string]interface{}{}
	}
	for _, defaultMiddleware := range strings.Split(defaultMiddlewares, ",") {
		if _, exists := ac.Middlewares[defaultMiddleware]; !exists {
			ac.Middlewares[defaultMiddleware] = nil
		}
	}
}

/*
   public function getDriver(): string
   {
       return $this->values['driver'];
   }

   public function getAddress(): string
   {
       return $this->values['address'];
   }

   public function getPort(): int
   {
       return $this->values['port'];
   }

   public function getUsername(): string
   {
       return $this->values['username'];
   }

   public function getPassword(): string
   {
       return $this->values['password'];
   }

   public function getDatabase(): string
   {
       return $this->values['database'];
   }
*/
func (ac *ApiConfig) GetTables() map[string]bool {
	result := map[string]bool{}
	for _, table := range strings.Split(ac.Tables, ",") {
		result[table] = true
	}
	return result
}

/*
   public function getMiddlewares(): array
   {
       return $this->values['middlewares'];
   }
*/
func (ac *ApiConfig) GetControllers() []string {
	return strings.Split(ac.Controllers, ",")
}

/*
    public function getCustomControllers(): array
    {
        return array_filter(array_map('trim', explode(',', $this->values['customControllers'])));
    }

    public function getCustomOpenApiBuilders(): array
    {
        return array_filter(array_map('trim', explode(',', $this->values['customOpenApiBuilders'])));
    }

    public function getCacheType(): string
    {
        return $this->values['cacheType'];
    }

    public function getCachePath(): string
    {
        return $this->values['cachePath'];
    }

    public function getCacheTime(): int
    {
        return $this->values['cacheTime'];
    }

    public function getDebug(): bool
    {
        return $this->values['debug'];
    }

    public function getBasePath(): string
    {
        return $this->values['basePath'];
    }

    public function getOpenApiBase(): array
    {
        return json_decode($this->values['openApiBase'], true);
    }
}*/

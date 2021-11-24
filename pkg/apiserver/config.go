package apiserver

import (
	"fmt"
	"strings"
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
	Middlewares           map[string][]map[string]interface{}
	Controllers           string
	CustomControllers     string
	CustomOpenApiBuilders string
	CacheType             string
	CachePath             string
	CacheTime             int
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

/*
   private function applyEnvironmentVariables(array $values): array
   {
       $newValues = array();
       foreach ($values as $key => $value) {
           $environmentKey = 'PHP_CRUD_API_' . strtoupper(preg_replace('/(?<!^)[A-Z]/', '_$0', str_replace('.', '_', $key)));
           $newValues[$key] = getenv($environmentKey, true) ?: $value;
       }
       return $newValues;
   }
*/

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

/*
   public function __construct(array $values)
   {
       $driver = $this->getDefaultDriver($values);
       $defaults = $this->getDriverDefaults($driver);
       $newValues = array_merge($this->values, $defaults, $values);
       $newValues = $this->parseMiddlewares($newValues);
       $diff = array_diff_key($newValues, $this->values);
       if (!empty($diff)) {
           $key = array_keys($diff)[0];
           throw new \Exception("Config has invalid value '$key'");
       }
       $newValues = $this->applyEnvironmentVariables($newValues);
       $this->values = $newValues;
   }

*/

func (ac *ApiConfig) initMiddlewares() {
	defaultMiddlewares := "cors,errors"
	for _, defaultMiddleware := range strings.Split(defaultMiddlewares, ",") {
		if _, exists := ac.Middlewares[defaultMiddleware]; !exists {
			ac.Middlewares[defaultMiddleware] = nil
		}
	}
}

/*
   private function parseMiddlewares(array $values): array
   {
       $newValues = array();
       $properties = array();
       $middlewares = array_map('trim', explode(',', $values['middlewares']));
       foreach ($middlewares as $middleware) {
           $properties[$middleware] = [];
       }
       foreach ($values as $key => $value) {
           if (strpos($key, '.') === false) {
               $newValues[$key] = $value;
           } else {
               list($middleware, $key2) = explode('.', $key, 2);
               if (isset($properties[$middleware])) {
                   $properties[$middleware][$key2] = $value;
               } else {
                   throw new \Exception("Config has invalid value '$key'");
               }
           }
       }
       $newValues['middlewares'] = array_reverse($properties, true);
       return $newValues;
   }

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
    public function getControllers(): array
    {
        return array_filter(array_map('trim', explode(',', $this->values['controllers'])));
    }

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

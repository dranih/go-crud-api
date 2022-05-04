# GO-CRUD-API
[![Build Status](https://github.com/dranih/go-crud-api/workflows/Build%20and%20test/badge.svg)](https://github.com/dranih/go-crud-api/actions?workflow=Build%20and%20test)

:warning: Work in progress :warning: 

Adding a REST API to a MySQL/MariaDB, PostgreSQL, SQL Server or SQLite database.

This is an attempt to port [php-crud-api](https://github.com/mevdschee/php-crud-api) to golang.  
Therefore, parts of this documentation refers to or copies the orginal projet readme.

## Installation

- Download latest release from Github and start it
```
./gocrudapi
```

- Or clone it and compile :
```
git@github.com:dranih/go-crud-api.git
cd go-crud-api/cmd/gocrudapi/
go install
./gocrudapi
```

## Configuration

**gocrudapi** looks for a **gcaconfig.yml** config file in current dir then $HOME if not found.  
The **GCA_CONFIG_FILE** environnement variable can also be set to the path of yaml configuration file.

Config file example :
```yaml
server:
  https: true

api:
  driver: "sqlite"
  controllers: "records,columns,cache,openapi,geojson,status"
  address: "/tmp/gocrudtests.db"
  database: "go-crud-api"
  username: "go-crud-api"
  password: "go-crud-api"

  middlewares:
  - basicAuth:
    - mode: "optional"
    - realm: "GoCrudApi : Username and password required"
    - passwordFile: "../../test/test.pwd"
```

These are all the configuration options and their default value :
- **server** block :

  |Option|Description|Default value|
  | --- | --- | --- |
  | address | Address the web server will be listening to | `:http` or `:https` |
  | http | Start the http server (boolean) | `true` |
  | https | Start the https server (boolean) | `false` |
  | httpPort | Address the http server will be listening to (int) | `8080` |
  | httpsPort | Address the https server will be listening to (int) | `8443` |
  | HttpsCertFile | Path to the PEM cert file for tls | will generate a self-signed certificate if https on |
  | HttpsKeyFile | Path to the PEM key file for tls | will generate a self-signed certificate if https on |
  | gracefulTimeout | Duration in seconds the web server will try to gracefully stop (int) | `15` |
  | writeTimeout | See [http.Server](https://pkg.go.dev/net/http?#Server) (int) | `15` |
  | readTimeout | See [http.Server](https://pkg.go.dev/net/http?#Server) (int) | `15` |
  | idleTimeout | See [http.Server](https://pkg.go.dev/net/http?#Server) (int) | `60` |

- **api** block :

  |Option|Description|Default value|
  | --- | --- | --- |
  | driver | `mysql`, `pgsql`, `sqlsrv` or `sqlite` | `mysql` |
  | address | Hostname (or filename) of the database server | `localhost` |
  | port | TCP port of the database server (int) | defaults to driver default |
  | username | Username of the user connecting to the database | no default |
  | password | Username of the user connecting to the database | no default |
  | database | Database the connecting is made to | no default |
  | tables | Comma separated list of tables to publish | defaults to 'all' |
  | mapping | List of table/column mappings | no mapping |
  | middlewares | List of middlewares to load (see  [Middlewares](#middlewares) for configuration) | `cors` |
  | controllers | List of controllers to load | `records,geojson,openapi,status` |
  | customControllers | Not implemented yet | N/A |
  | openApiBase | OpenAPI info | `{"info": {"title": "GO-CRUD-API", "version": "0.0.1"}}` |
  | cacheType | `TempFile`, `Redis`, `Memcache`, `Memcached` or `NoCache` | `TempFile` |
  | cachePath | Path/address of the cache  | defaults to system's temp directory |
  | cacheTime | Number of seconds the cache is valid (int)  | `10` |
  | debug | Show errors in the "X-Exception" headers (boolean) | `false` |
  | basePath | Not implemented yet | N/A |

All configuration options are also available as environment variables. Write the config option with capitals, a "GCA_" prefix and underscores for word breakes, so for instance:

- GCA_SERVER_HTTPS=true
- GCA_SERVER_HTTPSPORT=443
- GCA_API_DRIVER=mysql
- GCA_API_ADDRESS=localhost
- GCA_API_PORT=3306
- GCA_API_DATABASE=php-crud-api
- GCA_API_USERNAME=php-crud-api
- GCA_API_PASSWORD=php-crud-api
- GCA_API_DEBUG=1

The environment variables take precedence over the yaml file configuration.

## Limitations
See [php-crud-api#limitations](https://github.com/mevdschee/php-crud-api#limitations).

In addition, this golang implementation has some more limitations :
- Not actively used in production
- Using [gotemplate](https://golangdocs.com/templates-in-golang) for handlers instead of pure php code.

## Features
See [php-crud-api#features](https://github.com/mevdschee/php-crud-api#features).

Missing features : customControllers, basePath

## API usage
See [php-crud-api#treeql-a-pragmatic-graphql](https://github.com/mevdschee/php-crud-api#treeql-a-pragmatic-graphql)

## Middlewares
See [php-crud-api#middleware](https://github.com/mevdschee/php-crud-api#middleware)

One of the main drawback of **GO-CRUD-API** compared to **PHP-CRUD-API** is the use of [gotemplate](https://golangdocs.com/templates-in-golang) syntax for the user provided handlers which is more complex and less powerful than php (or go) code.

The middlewares options have to be configured in the yaml configuration file, ex :
```yaml
api:
  middlewares:
 - basicAuth:
    - mode: "optional"
    - realm: "GoCrudApi : Username and password required"
    - passwordFile: "../../test/test.pwd"
  - json:
    - controllers: "records"
    - tables: "products"
    - columns: "properties"
  - xml:
  - cors:
  - validation:
    - handler: "{{ if and (eq .Column.GetName \"post_id\") (and (not (kindIs \"float64\" .Value)) (not (kindIs \"int\" .Value))) }}must be numeric{{ else }}true{{ end }}"
```

## OpenAPI specification
See [php-crud-api#openapi-specification](https://github.com/mevdschee/php-crud-api#openapi-specification)

## Cache
See [php-crud-api#cache](https://github.com/mevdschee/php-crud-api#cache)

## Types
See [php-crud-api#types](https://github.com/mevdschee/php-crud-api#types)

## Errors
See [php-crud-api#errors](https://github.com/mevdschee/php-crud-api#errors)

## Status
See [php-crud-api#status](https://github.com/mevdschee/php-crud-api#status)

## Tests
Functional tests from [PHP-CRUD-API](https://github.com/mevdschee/php-crud-api/tree/main/tests/functional) had been implemented in the [apiserver package](./pkg/apiserver/).

More unit tests are needed in all the packages.

The [test folder](./test/) contains the procedure and the configuration files used to launch the tests with all four kind of databases (mysql, pgsql, sqlite and sqlserver).

The build and test github action only starts tests with sqlite database at the moment.

## Dependencies
Those go packages are used :
|Package|Usage|
| --- | --- |
| [github.com/gorilla/mux](https://github.com/gorilla/mux) | Routing requests |
| [github.com/clbanning/mxj/v2](https://github.com/clbanning/mxj) | XML decode/encode |
| [github.com/gorilla/sessions](https://github.com/gorilla/sessions) | Handling sessions |
| [github.com/spf13/viper](https://github.com/spf13/viper) | Reading configuration |
| [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) | Sqlite database driver |
| [github.com/denisenkom/go-mssqldb](https://github.com/denisenkom/go-mssqldb) | Mssql database driver |
| [github.com/lib/pq](https://github.com/lib/pq) | Postgresql database driver |
| [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) | Mysql database driver |
| [github.com/Masterminds/sprig/v3](https://github.com/Masterminds/sprig) | Adds functions to go template |
| [github.com/carmo-evan/strtotime](https://github.com/carmo-evan/strtotime) | Parsing text to date |
| [github.com/bradfitz/gomemcache](https://github.com/bradfitz/gomemcache) | Memcache client library |
| [github.com/go-redis/redis/v8](https://github.com/go-redis/redis) | Redis client library |
| [github.com/patrickmn/go-cache](github.com/patrickmn/go-cache) | In-memory cache (not maintained ?) |

## To-do
- [X] Fix bug with seq in renaming tables in sqlserver (records test 080_add_barcode_with_ip_address_A fail after columns tests)
- [ ] Tests : 
  - [ ] more unit tests
  - [X] implement php-crud-api tests
- [X] Other drivers (only sqlite now)
- [X] Cache mecanism
- [X] Finishing controllers
- [ ] Custom controller (compile extra go code at launch like https://github.com/benhoyt/prig ?)
- [X] Finishing middlewares
- [ ] Add a github workflow
  - [X] Init
  - [X] Add pgsql, mysql and sqlserver testing
  - [X] Find why somes linters are not working  : linked to https://github.com/golang/go/issues/26863, should rewrite condition interface. See also https://github.com/golang/go/issues/28254#issuecomment-996822778
  - [X] Release pipeline
- [ ] Add an alter table function for sqlite (create new table, copy data, drop old table)
- [X] Review packages structure
- [ ] Logger options
- [X] https
- [X] Write a README
- [ ] Comment code
- [ ] :tada:
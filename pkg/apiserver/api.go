package apiserver

import (
	"context"
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/mux"

	"github.com/dranih/go-crud-api/pkg/cache"
	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/geojson"
	"github.com/dranih/go-crud-api/pkg/middleware"
	"github.com/dranih/go-crud-api/pkg/openapi"
	"github.com/dranih/go-crud-api/pkg/record"
	"github.com/dranih/go-crud-api/pkg/utils"
)

type Api struct {
	router *mux.Router
	config *Config
}

//todo : cache
func NewApi(globalConfig *Config) *Api {
	config := globalConfig.Api
	db := database.NewGenericDB(
		config.Driver,
		config.Address,
		config.Port,
		config.Database,
		config.GetTables(),
		config.Username,
		config.Password)
	prefix := fmt.Sprintf("gocrudapi-%d-", os.Getpid())
	cache := cache.Create(config.CacheType, prefix, config.CachePath)
	reflection := database.NewReflectionService(db, cache, config.CacheTime)
	responder := controller.NewJsonResponder(config.Debug)
	router := mux.NewRouter()
	//Consistent middle order :
	//sslRedirect,cors,*firewall*,*xsrf*,xml,json,reconnect,apiKeyAuth,apiKeyDbAuth,dbAuth,*jwtAuth*,basicAuth,authorization,sanitation,validation,ipAddress,multiTenancy,pageLimits,joinLimits,customization
	if properties, exists := config.Middlewares["sslRedirect"]; exists {
		sslMiddle := middleware.NewSslRedirectMiddleware(responder, properties, globalConfig.Server.HttpsPort)
		router.Use(sslMiddle.Process)
	}
	if properties, exists := config.Middlewares["cors"]; exists {
		router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}).Methods("OPTIONS")
		corsMiddleware := middleware.NewCorsMiddleware(responder, properties, config.Debug)
		router.Use(corsMiddleware.Process)
	}
	if properties, exists := config.Middlewares["xml"]; exists {
		xmlMiddle := middleware.NewXmlMiddleware(responder, properties)
		router.Use(xmlMiddle.Process)
	}
	if properties, exists := config.Middlewares["json"]; exists {
		jsonMiddle := middleware.NewJsonMiddleware(responder, properties)
		router.Use(jsonMiddle.Process)
	}
	if properties, exists := config.Middlewares["reconnect"]; exists {
		reconnectMiddle := middleware.NewReconnectMiddleware(responder, properties, reflection, db)
		router.Use(reconnectMiddle.Process)
	}
	if properties, exists := config.Middlewares["apiKeyAuth"]; exists {
		akamMiddle := middleware.NewApiKeyAuth(responder, properties)
		router.Use(akamMiddle.Process)
	}
	if properties, exists := config.Middlewares["apiKeyDbAuth"]; exists {
		akdamMiddle := middleware.NewApiKeyDbAuth(responder, properties, reflection, db)
		router.Use(akdamMiddle.Process)
	}
	if properties, exists := config.Middlewares["dbAuth"]; exists {
		damMiddle := middleware.NewDbAuth(responder, properties, reflection, db)
		router.Use(damMiddle.Process)
	}
	if properties, exists := config.Middlewares["basicAuth"]; exists {
		bamMiddle := middleware.NewBasicAuth(responder, properties)
		router.Use(bamMiddle.Process)
	}
	if properties, exists := config.Middlewares["authorization"]; exists {
		authMiddle := middleware.NewAuthorizationMiddleware(responder, properties, reflection)
		router.Use(authMiddle.Process)
	}
	if properties, exists := config.Middlewares["sanitation"]; exists {
		sanitationMiddle := middleware.NewSanitationMiddleware(responder, properties, reflection)
		router.Use(sanitationMiddle.Process)
	}
	if properties, exists := config.Middlewares["validation"]; exists {
		validationMiddle := middleware.NewValidationMiddleware(responder, properties, reflection)
		router.Use(validationMiddle.Process)
	}
	if properties, exists := config.Middlewares["ipAddress"]; exists {
		ipAddressMiddle := middleware.NewIpAddressMiddleware(responder, properties, reflection)
		router.Use(ipAddressMiddle.Process)
	}
	if properties, exists := config.Middlewares["multiTenancy"]; exists {
		multiTenancyMiddle := middleware.NewMultiTenancyMiddleware(responder, properties, reflection)
		router.Use(multiTenancyMiddle.Process)
	}
	if properties, exists := config.Middlewares["pageLimits"]; exists {
		pageLimitsMiddle := middleware.NewPageLimitsMiddleware(responder, properties, reflection)
		router.Use(pageLimitsMiddle.Process)
	}
	if properties, exists := config.Middlewares["joinLimits"]; exists {
		joinLimitsMiddle := middleware.NewJoinLimitsMiddleware(responder, properties, reflection)
		router.Use(joinLimitsMiddle.Process)
	}
	if properties, exists := config.Middlewares["customization"]; exists {
		gob.Register(map[string]interface{}{})
		customizationMiddle := middleware.NewCustomizationMiddleware(responder, properties, reflection)
		router.Use(customizationMiddle.Process)
	}

	//Save session after all middlewares
	//Session should not be altered by the controllers
	saveSessionMiddle := middleware.NewSaveSession(responder, nil)
	router.Use(saveSessionMiddle.Process)

	for ctrl := range config.GetControllers() {
		switch ctrl {
		case "records":
			records := database.NewRecordService(db, reflection)
			controller.NewRecordController(router, responder, records)
		case "columns":
			definition := database.NewDefinitionService(db, reflection)
			controller.NewColumnController(router, responder, reflection, definition)
		case "cache":
			controller.NewCacheController(router, responder, cache)
		case "openapi":
			openapi := openapi.NewOpenApiService(reflection, config.OpenApiBase, config.GetControllers(), config.GetCustomOpenApiBuilders())
			controller.NewOpenApiController(router, responder, openapi)
		case "geojson":
			records := database.NewRecordService(db, reflection)
			geoJson := geojson.NewGeoJsonService(reflection, records)
			controller.NewGeoJsonController(router, responder, geoJson)
		case "status":
			controller.NewStatusController(router, responder, cache, db)
		}
	}

	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responder.Error(record.ROUTE_NOT_FOUND, r.RequestURI, w, "")
	}).Methods("OPTIONS", "GET", "PUT", "POST", "DELETE", "PATCH")

	return &Api{router, globalConfig}
}

func (a *Api) Handle(wg *sync.WaitGroup) {
	config := a.config.Server
	//From https://golangexample.com/a-powerful-http-router-and-url-matcher-for-building-go-web-servers/
	var srvHttp, srvHttps *http.Server
	if config.Http {
		srvHttp = &http.Server{
			Addr: fmt.Sprintf("%s:%d", config.Address, config.HttpPort),
			// Good practice to set timeouts to avoid Slowloris attacks.
			WriteTimeout: time.Second * time.Duration(config.WriteTimeout),
			ReadTimeout:  time.Second * time.Duration(config.ReadTimeout),
			IdleTimeout:  time.Second * time.Duration(config.IdleTimeout),
			Handler:      a.router, // Pass our instance of gorilla/mux in.
		}
		if wg != nil {
			wg.Add(1)
		}
		// Run our server in a goroutine so that it doesn't block.
		go func() {
			addr := srvHttp.Addr
			if addr == "" {
				addr = ":http"
			}
			ln, err := net.Listen("tcp", addr)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Started http server at %s", addr)
			if wg != nil {
				wg.Done()
			}
			if err := srvHttp.Serve(ln); err != nil {
				log.Fatal(err)
			}
		}()
	}

	if config.Https {
		var serverTLSConf *tls.Config
		if config.HttpsCertFile == "" || config.HttpsKeyFile == "" {
			var err error
			serverTLSConf, _, err = utils.CertSetup(config.Address)

			if err != nil {
				log.Fatal(err)
			}
		} else {
			serverCert, err := tls.LoadX509KeyPair(config.HttpsCertFile, config.HttpsKeyFile)
			if err != nil {
				log.Fatal(err)
			}
			serverTLSConf = &tls.Config{
				Certificates: []tls.Certificate{serverCert},
			}
		}

		srvHttps = &http.Server{
			Addr: fmt.Sprintf("%s:%d", config.Address, config.HttpsPort),
			// Good practice to set timeouts to avoid Slowloris attacks.
			WriteTimeout: time.Second * time.Duration(config.WriteTimeout),
			ReadTimeout:  time.Second * time.Duration(config.ReadTimeout),
			IdleTimeout:  time.Second * time.Duration(config.IdleTimeout),
			TLSConfig:    serverTLSConf,
			Handler:      a.router, // Pass our instance of gorilla/mux in.
		}
		if wg != nil {
			wg.Add(1)
		}
		// Run our server in a goroutine so that it doesn't block.
		go func() {
			addr := srvHttps.Addr
			if addr == "" {
				addr = ":https"
			}
			ln, err := net.Listen("tcp", addr)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Started https server at %s", addr)
			if wg != nil {
				wg.Done()
			}
			if err := srvHttps.ServeTLS(ln, "", ""); err != nil {
				log.Fatal(err)
			}
		}()
	}

	if wg != nil {
		wg.Done()
	}
	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(config.GracefulTimeout))
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	if srvHttp != nil {
		srvHttp.Shutdown(ctx)
	}
	if srvHttps != nil {
		srvHttps.Shutdown(ctx)
	}
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}

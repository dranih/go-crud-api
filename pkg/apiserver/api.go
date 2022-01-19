package apiserver

import (
	"context"
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
)

type Api struct {
	router *mux.Router
	debug  bool
}

//todo : cache
func NewApi(config *ApiConfig) *Api {
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
	for middle, properties := range config.Middlewares {
		switch middle {
		case "basicAuth":
			bamMiddle := middleware.NewBasicAuth(responder, properties)
			router.Use(bamMiddle.Process)
		}
	}

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
	return &Api{router, config.Debug}
}

func (a *Api) Handle(config *ServerConfig, wg *sync.WaitGroup) {
	//From https://golangexample.com/a-powerful-http-router-and-url-matcher-for-building-go-web-servers/
	srv := &http.Server{
		Addr: fmt.Sprintf("%s:%d", config.Address, config.Port),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * time.Duration(config.WriteTimeout),
		ReadTimeout:  time.Second * time.Duration(config.ReadTimeout),
		IdleTimeout:  time.Second * time.Duration(config.IdleTimeout),
		Handler:      a.router, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		/*defer wg.Done()
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}*/
		addr := srv.Addr
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
		if err := srv.Serve(ln); err != nil {
			log.Fatal(err)
		}
	}()

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
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}

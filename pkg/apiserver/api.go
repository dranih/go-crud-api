package apiserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/database"
)

type Api struct {
	router *mux.Router
	debug  bool
}

//todo : cache
func NewApi(config *ApiConfig) *Api {
	router := mux.NewRouter()
	dbClient := database.NewGenericDB(
		config.Driver,
		config.Address,
		config.Port,
		config.Database,
		config.GetTables(),
		config.Username,
		config.Password)
	//$prefix = sprintf('phpcrudapi-%s-', substr(md5(__FILE__), 0, 8));
	//$cache = CacheFactory::create($config->getCacheType(), $prefix, $config->getCachePath());
	reflection := database.NewReflectionService(dbClient, "", 0)
	for _, ctrl := range config.GetControllers() {
		switch ctrl {
		case "records":
			records := database.NewRecordService(dbClient, reflection)
			controller.NewRecordController(router, records, config.Debug)
		case "columns":
			//$definition = new DefinitionService($db, $reflection);
			//new ColumnController($router, $responder, $reflection, $definition);
		case "cache":
			//new CacheController($router, $responder, $cache);
		case "openapi":
			//$openApi = new OpenApiService($reflection, $config->getOpenApiBase(), $config->getControllers(), $config->getCustomOpenApiBuilders());
			//new OpenApiController($router, $responder, $openApi);
		case "geojson":
			//$records = new RecordService($db, $reflection);
			//$geoJson = new GeoJsonService($reflection, $records);
			//new GeoJsonController($router, $responder, $geoJson);
		case "status":
			//new StatusController($router, $responder, $cache, $db);
		}
	}
	return &Api{router, config.Debug}
}

func (a *Api) Handle(config *ServerConfig) {
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
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
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

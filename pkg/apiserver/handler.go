package apiserver

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"

	"github.com/dranih/go-crud-api/pkg/database"
)

var dbClient database.Client
var dbReflection database.GenericReflection

func Handle() {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	//Try DB connection
	connectDB()

	router := mux.NewRouter()

	router.HandleFunc("/status/ping", getPing).Methods("GET")
	router.HandleFunc("/records/{table}/{row}", read).Methods("GET")

	srv := &http.Server{
		Addr: "0.0.0.0:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router, // Pass our instance of gorilla/mux in.
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
	ctx, cancel := context.WithTimeout(context.Background(), wait)
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

func getPing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

func read(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Table: %v\n", vars["table"])
	fmt.Fprintf(w, "Row: %v\n", vars["row"])
	m := make(map[string]interface{})
	m["name"] = vars["row"]
	var results []map[string]interface{}
	dbClient.Client.Select("*").Where(m).Table(vars["table"]).Find(&results)
	fmt.Fprintf(w, "Result: %v\n", results)
	dbTables := dbReflection.GetTables()
	fmt.Fprintf(w, "Tables: %v\n", dbTables)
	dbColumns := dbReflection.GetTableColumns("cows", "")
	fmt.Fprintf(w, "Columns: %v\n", dbColumns)
}

func connectDB() {
	if err := dbClient.Connect(); err != nil {
		log.Fatalf("Connection to database failed : %v", err)
		os.Exit(1)
	}
	dbReflection = *database.NewGenericReflection(dbClient, "", "sqlite", "test", map[string]bool{"sharks": true}, "")
}

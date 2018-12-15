package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"time"

	"log"

	"github.com/gorilla/mux"

	// we do a blank import here because this packages init() function simply registers the package as the SQL driver for mysql(mariaDB), init() runs before main()
	_ "github.com/go-sql-driver/mysql"
)

var (
	db *sql.DB // global variable so we can access the DB from our handlers
)

func main() {

	// first we connect to the SQL DB, so resources are available upon the first request to the server
	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		log.Fatalf("Failed to open SQL database: %s", err)
	}

	// and close the DB when this function ends
	defer db.Close()

	// Setup the mux router and register our routes
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	// r.HandleFunc("/second-route", CoolNewHandler)

	// if we want to do something to each request before entering the route handler, for example checking an auth token
	// we can implement a middleware
	r.Use(loggingMiddleware, authMiddleware)

	// Now lets setup the server with a graceful shutddown
	srv := &http.Server{
		Addr: "0.0.0.0:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	srv.Shutdown(ctx)
	log.Println("shutting down")
	os.Exit(0)
}

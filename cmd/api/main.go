package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// version represents the version of this application.
const version = "1.0.0"

// config holds all the configuration settings for this application.
type config struct {
	port int
	env  string
}

// application holds the dependencies for HTTP handlers, helpers, loggers and middlewares.
type application struct {
	config config
	logger *log.Logger
}

func main() {
	// Initialize a config to store settings from flags.
	var cfg config

	// Parse flags and store settings into config.
	flag.IntVar(&cfg.port, "port", 8080, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// Initialize a new logger for application.
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Create an application with config and logger.
	app := &application{
		config: cfg,
		logger: logger,
	}

	// Create a servemux and add the temporary first page.
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	// Create a HTTP server.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start the HTTP server.
	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err := srv.ListenAndServe()
	logger.Fatal(err)
}

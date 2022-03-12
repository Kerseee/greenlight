package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"greenlight.kerseeehuang.com/internal/data"
	"greenlight.kerseeehuang.com/internal/jsonlog"
	"greenlight.kerseeehuang.com/internal/mailer"
)

var (
	buildTime string // buildTime stores the building time when using go build command.
	version   string // version represents the version of this application.
)

// config holds all the configuration settings for this application.
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	// limiter holds configuration settings for the rate limiter.
	limiter struct {
		rps     float64 // rate limit per second
		burst   int
		enabled bool
	}
	// smtp holds configuration settings for the SMTP server.
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	// cors holds trusted origins
	cors struct {
		trustedOrigins []string
	}
}

// application holds the dependencies for HTTP handlers, helpers, loggers and middlewares.
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	// Initialize a config to store settings from flags.
	var cfg config

	// Parse flags and store settings into config.
	flag.IntVar(&cfg.port, "port", 8080, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second ")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate Limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "7dd615752ea0d8", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "e4faa190df6fd5", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@example.com>", "SMTP sender")

	flag.Func("cors-trustedOrigins", "Trusted CORS origins (space separated)", func(s string) error {
		cfg.cors.trustedOrigins = strings.Fields(s)
		return nil
	})

	displayVersion := flag.Bool("version", false, "Display application version and exit")

	flag.Parse()

	// Display the version and exit if flag "version" is set.
	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		fmt.Printf("Build time:\t%s\n", buildTime)
		os.Exit(0)
	}

	// Initialize a new logger for application.
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Create the DB connection pool.
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err.Error(), nil)
	}
	defer db.Close()
	logger.PrintInfo("database connection pool established", nil)

	// Publish metrics information to expvar handler.
	expvar.NewString("version").Set(version)
	expvar.Publish("goroutine", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("database", expvar.Func(func() interface{} {
		return db.Stats()
	}))
	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	// Create an application with config and logger.
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	// Create a server and serve.
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err.Error(), nil)
	}
}

// openDB creates a connection pool of DB, establish the first connection,
// and returns the DB connection pool.
func openDB(cfg config) (*sql.DB, error) {
	// Open an empty connection pool.
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set the parameter of limits on connections
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(int(duration))

	// Create a dummy time-out-context for the first connection.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Establish the first connection to DB.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

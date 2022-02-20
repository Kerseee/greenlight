package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const shutDownTimeout = 5 * time.Second // 5 seconds

// serve create a http server and call server.ListenAndServe().
// It only return non-nil error from server.ListenAndServe()
func (app *application) serve() error {
	// Create a HTTP server.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		ErrorLog:     log.New(app.logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Create a channel for shutdown error.
	shutdownError := make(chan error)

	// Listen to quit signals in the background.
	go func() {
		// Wait for catching signals.
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		// Inform admin about shutdown.
		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})

		// Shutdown the http server and send shutdown error to the main goroutine.
		ctx, cancel := context.WithTimeout(context.Background(), shutDownTimeout)
		defer cancel()
		shutdownError <- srv.Shutdown(ctx)
	}()

	// Start the HTTP server.
	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	// Run the HTTP server.
	// If the error is not http.ErrServerClosed, then return the error directly,
	// since that indicate there is wrong with the graceful shutdown.
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Handle the shutdown error.
	err = <-shutdownError
	if err != nil {
		return err
	}

	// Inform the admin.
	app.logger.PrintInfo("server was stopped", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}

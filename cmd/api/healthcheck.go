package main

import (
	"fmt"
	"net/http"
)

// healthcheckHandler shows the health status to client.
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Create a fixed format JSON message.
	js := `{"status": "available", "environment": %q, "version": "%q"}`
	js = fmt.Sprintf(js, app.config.env, version)

	// Set the "Content-Type" response header to JSON.
	w.Header().Set("Content-Type", "application/json")

	// Write the response body.
	w.Write([]byte(js))
}

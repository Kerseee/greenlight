package main

import (
	"net/http"
)

// healthcheckHandler shows the health status to client.
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {

	data := map[string]string{
		"status":      "available",
		"environment": "somthing",
		"version":     version,
	}

	// Write JSON.
	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "Server Error! Cannot write JSON!", http.StatusInternalServerError)
	}

}

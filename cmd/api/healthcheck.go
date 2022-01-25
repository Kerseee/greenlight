package main

import (
	"net/http"
)

// healthcheckHandler shows the health status to client.
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {

	data := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	// Write JSON.
	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "Server Error! Cannot write JSON!", http.StatusInternalServerError)
	}

}

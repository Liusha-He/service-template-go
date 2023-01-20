package main

import (
	"logger-service/data"
	"net/http"
)

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	// read json into a var
	var requestPayload JSONPayload
	_ = app.ReadJson(w, r, &requestPayload)

	// insert data
	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}

	err := app.Models.LogEntry.Insert(event)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	resp := jsonResponse{
		Error:   false,
		Message: "Logged",
	}

	app.WriteJson(w, http.StatusAccepted, resp)
}

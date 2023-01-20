package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

func (app *Config) AuthHandler(w http.ResponseWriter, r *http.Request) {
	var RequestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.ReadJson(w, r, &RequestPayload)
	if err != nil {
		app.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	// validate the user against DB
	user, err := app.Models.User.GetByEmail(RequestPayload.Email)
	if err != nil {
		app.ErrorJson(w, errors.New("User not exists"), http.StatusBadRequest)
		return
	}

	valid, err := user.PasswordMatches(RequestPayload.Password)
	if err != nil || !valid {
		app.ErrorJson(w, errors.New("Password invalid..."))
		return
	}

	// log authentication
	err = app.LogRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	app.WriteJson(w, http.StatusAccepted, payload)
}

func (app *Config) LogRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceUrl := os.Getenv("LOGGER-URL")

	request, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	client := &http.Client{}

	_, err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}

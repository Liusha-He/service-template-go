package main

import (
	"broker/event"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

var AUTH_URL = os.Getenv("AUTH-URL")
var LOG_URL = os.Getenv("LOGGER-URL")
var MAIL_URL = os.Getenv("MAIL-URL")

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the Broker",
	}

	_ = app.WriteJson(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmit(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.ReadJson(w, r, &requestPayload)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.Authenticate(w, requestPayload.Auth)
	case "log":
		app.LogEventViaRabbit(w, requestPayload.Log)
	case "mail":
		app.SendMail(w, requestPayload.Mail)
	default:
		app.ErrorJson(w, errors.New("Unknown Action..."))
	}
}

func (app *Config) Authenticate(w http.ResponseWriter, a AuthPayload) {
	// create json body we will send to auth service
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	// call the auth service
	request, err := http.NewRequest("POST", AUTH_URL, bytes.NewBuffer(jsonData))
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	defer response.Body.Close()

	// make sure we get expected status code
	if response.StatusCode == http.StatusUnauthorized {
		app.ErrorJson(w, errors.New("Invalid Authentication..."))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.ErrorJson(w, errors.New(fmt.Sprintf("Error calling authentication service... %s", response.Status)))
		return
	}

	var jsonFromService jsonResponse
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	if jsonFromService.Error {
		app.ErrorJson(w, err, http.StatusUnauthorized)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated!"
	payload.Data = jsonFromService.Data

	app.WriteJson(w, http.StatusAccepted, payload)
}

func (app *Config) LogItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	request, err := http.NewRequest("POST", LOG_URL, bytes.NewBuffer(jsonData))
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.ErrorJson(w, errors.New("error calling logging service..."))
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Logged"

	app.WriteJson(w, http.StatusAccepted, payload)
}

func (app *Config) LogEventViaRabbit(w http.ResponseWriter, l LogPayload) {
	err := app.PushToQueue(l.Name, l.Data)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged via RabbitMQ"

	app.WriteJson(w, http.StatusAccepted, payload)
}

func (app *Config) PushToQueue(qname, msg string) error {
	emitter, err := event.NewEventEmitter(app.RabbitConn)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: qname,
		Data: msg,
	}

	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}

	return nil
}

func (app *Config) SendMail(w http.ResponseWriter, mail MailPayload) {
	jsonData, _ := json.MarshalIndent(mail, "", "\t")

	request, err := http.NewRequest("POST", MAIL_URL, bytes.NewBuffer(jsonData))
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.ErrorJson(w, errors.New("error calling mail service..."))
		return
	}
	var payload jsonResponse
	payload.Error = false
	payload.Message = "Mail sent to " + mail.To

	app.WriteJson(w, http.StatusAccepted, payload)
}

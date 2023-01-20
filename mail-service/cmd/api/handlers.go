package main

import (
	"net/http"
)

func (app *Config) SendMail(w http.ResponseWriter, r *http.Request) {
	type MailMessage struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	var requestPayload MailMessage
	err := app.ReadJson(w, r, &requestPayload)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	msg := Message{
		From:    requestPayload.From,
		To:      requestPayload.To,
		Subject: requestPayload.Subject,
		Data:    requestPayload.Message,
	}

	err = app.Mailer.SendSMTPMessage(msg)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "semt to " + requestPayload.To,
	}

	app.WriteJson(w, http.StatusAccepted, payload)
}

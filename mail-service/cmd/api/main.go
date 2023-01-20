package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Config struct {
	Mailer Mail
}

const webPort = "80"

func main() {
	app := Config{
		Mailer: CreateMail(),
	}

	log.Println("Starting mail service on port", webPort)

	service := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.Routes(),
	}

	err := service.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}

}

func CreateMail() Mail {
	port, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
	m := Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Host:        os.Getenv("MAIL_HOST"),
		Port:        port,
		Username:    os.Getenv("MAIL_USERNAME"),
		Password:    os.Getenv("MAIL_PASSWORD"),
		Encryption:  os.Getenv("MAIL_ENCRYPTION"),
		FromName:    os.Getenv("MAIL_FROMNAME"),
		FromAddress: os.Getenv("MAIL_FROMADDRESS"),
	}

	return m
}

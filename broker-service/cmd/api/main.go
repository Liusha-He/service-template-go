package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "80"

var (
	CONNECTION_STRING = os.Getenv("CONNECTION-STRING")
)

type Config struct {
	RabbitConn *amqp.Connection
}

func main() {
	rabbitConn, err := Connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitConn.Close()

	app := Config{
		RabbitConn: rabbitConn,
	}

	log.Printf("Starting Broker service on port %s\n", webPort)

	// define http server
	servoce := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.Routes(),
	}

	// start server
	err = servoce.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func Connect() (*amqp.Connection, error) {
	var counts int64
	var backoff = 1 * time.Second
	var connection *amqp.Connection

	for {
		conn, err := amqp.Dial(CONNECTION_STRING)
		if err != nil {
			log.Println("RabbitMQ not ready...")
			counts++
		} else {
			log.Println("Connected to RabbitMQ!")
			connection = conn
			break
		}

		if counts > 5 {
			log.Panic(err)
			return nil, err
		}

		backoff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off...")
		time.Sleep(backoff)
		continue
	}

	return connection, nil
}

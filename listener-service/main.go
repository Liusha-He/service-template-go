package main

import (
	"listener-service/event"
	"log"
	"math"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	n_retry = 5
)

var (
	CONNECTION_STRING = os.Getenv("CONNECTION-STRING")
)

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

func main() {
	rabbitConn, err := Connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitConn.Close()

	// start listening to messages
	log.Println("Listening for and consuming RabbitMQ messages")

	// create consumer
	consumer, err := event.NewConsumer(rabbitConn)
	if err != nil {
		log.Panic(err)
	}

	// watch the queue and consume events
	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})
	if err != nil {
		log.Panic(err)
	}
}

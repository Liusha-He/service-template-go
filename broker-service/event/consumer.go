package event

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

var LOG_URL = os.Getenv("LOGGER-URL")

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}

	err := consumer.Setup()
	if err != nil {
		return Consumer{}, err
	}

	return consumer, nil
}

func (c *Consumer) Setup() error {
	channel, err := c.conn.Channel()
	if err != nil {
		return err
	}
	return DeclareExchange(channel)
}

func (c *Consumer) Listen(topics []string) error {
	channel, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	q, err := DeclareRandomQueue(channel)
	if err != nil {
		return err
	}

	for _, s := range topics {
		err := channel.QueueBind(
			q.Name,
			s,
			"logs_topic",
			false,
			nil,
		)

		if err != nil {
			return err
		}
	}

	messages, err := channel.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	forever := make(chan bool)
	go func() {
		for d := range messages {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)

			go HandlePayload(payload)

		}
	}()

	log.Printf("Waiting for message [Exchange, Queue] - [logs_topic, %s]", q.Name)
	<-forever

	return nil
}

func HandlePayload(payload Payload) {
	switch payload.Name {
	case "log", "event":
		// log whatever we get
		err := LogEvent(payload)
		if err != nil {
			log.Println(err)
		}
	case "auth":
		// authenticate
	default:
		err := LogEvent(payload)
		if err != nil {
			log.Println(err)
		}
	}
}

func LogEvent(entry Payload) error {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	client := &http.Client{}

	request, err := http.NewRequest("POST", LOG_URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusAccepted {
		return errors.New("error calling logging service...")
	}

	return nil
}

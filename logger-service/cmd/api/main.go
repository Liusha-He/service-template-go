package main

import (
	"context"
	"fmt"
	"log"
	"logger-service/data"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort  = "80"
	rpcPort  = "5001"
	grpcPort = "50001"
)

var (
	mongoUrl = os.Getenv("MONGO-URL")
	client   *mongo.Client
)

type Config struct {
	Models data.Models
}

func ConnectToMongo() (*mongo.Client, error) {
	// create connection options
	clientOptions := options.Client().ApplyURI(mongoUrl)
	clientOptions.SetAuth(options.Credential{
		Username: os.Getenv("USERNAME"),
		Password: os.Getenv("PASSWORD"),
	})

	// connect to the db
	conn, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("Error connecting:", err)
		return nil, err
	}

	return conn, err
}

// func (app *Config) Serve() {
// 	server := &http.Server{
// 		Addr:    fmt.Sprintf(":%s", webPort),
// 		Handler: app.Routes(),
// 	}

// 	err := server.ListenAndServe()
// 	if err != nil {
// 		log.Panic()
// 	}
// }

func main() {
	// connect to mongodb
	mongoClient, err := ConnectToMongo()
	if err != nil {
		log.Panic(err)
	}
	client = mongoClient

	// create a context in order to disconnect
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// close connection
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	// start web server
	// go app.Serve()
	log.Println("Starting service on port", webPort)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.Routes(),
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Panic()
	}
}

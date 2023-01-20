package main

import (
	"auth-service/data"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	webPort = "80"
	tries   = 10
)

var counts int64

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()

	if err != nil {
		return nil, err
	}

	return db, err
}

func ConnectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		conn, err := OpenDB(dsn)

		if err != nil {
			log.Println("DB Driver not Ready...")
			counts++
		} else {
			log.Println("Connected to DB!")
			return conn
		}

		if counts > tries {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for 2 seconds...")
		time.Sleep(2 * time.Second)
		continue
	}
}

func main() {
	log.Println("Starting Authentication Service...")

	// connect to database
	conn := ConnectToDB()
	if conn == nil {
		log.Panic("Can't connect to DB!")
	}

	// setup config
	app := Config{
		DB:     conn,
		Models: data.New(conn),
	}

	service := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.Routes(),
	}

	err := service.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}

}

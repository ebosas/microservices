package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ebosas/microservices/internal/config"
	"github.com/ebosas/microservices/internal/models"
	"github.com/ebosas/microservices/internal/rabbit"
	"github.com/jackc/pgx/v4"
	"github.com/streadway/amqp"
)

var conf = config.New()

func main() {
	fmt.Println("Running database service")

	// Postgres connection
	connP, err := pgx.Connect(context.Background(), conf.PostgresURL)
	if err != nil {
		log.Fatalf("postgres connection: %s", err)
	}
	defer connP.Close(context.Background())

	// Rabbit connection
	connR, err := rabbit.GetConn(conf.RabbitURL)
	if err != nil {
		log.Fatalf("rabbit connection: %s", err)
	}
	defer connR.Close()

	err = connR.DeclareTopicExchange(conf.Exchange)
	if err != nil {
		log.Fatalf("declare exchange: %s", err)
	}

	// Start a Rabbit consumer with a message processing handler.
	connR.StartConsumer(conf.Exchange, conf.QueueDB, conf.KeyDB, func(d amqp.Delivery) bool {
		return insertToDB(d, connP)
	})

	select {}
}

// insertToDB inserts a Rabbit message into a Postgres database.
func insertToDB(d amqp.Delivery, connP *pgx.Conn) bool {
	var message models.Message
	err := json.Unmarshal(d.Body, &message)
	if err != nil {
		log.Fatalf("unmarshal message: %s", err)
	}

	_, err = connP.Exec(context.Background(), "insert into messages (message, created) values ($1, to_timestamp($2))", message.Text, message.Time/1000)
	if err != nil {
		log.Fatalf("insert into database: %s", err)
	}

	return true
}

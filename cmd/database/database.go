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
	fmt.Println("[Database service]")

	// Postgres connection
	connPG, err := pgx.Connect(context.Background(), conf.PostgresURL)
	if err != nil {
		log.Fatalf("postgres connection: %s", err)
	}
	defer connPG.Close(context.Background())

	// RabbitMQ connection
	connMQ, err := rabbit.GetConn(conf.RabbitURL)
	if err != nil {
		log.Fatalf("rabbit connection: %s", err)
	}
	defer connMQ.Close()

	err = connMQ.DeclareTopicExchange(conf.Exchange)
	if err != nil {
		log.Fatalf("declare exchange: %s", err)
	}

	// Start a Rabbit consumer with a message processing handler.
	connMQ.StartConsumer(conf.Exchange, conf.QueueDB, conf.KeyDB, func(d amqp.Delivery) bool {
		return insertToDB(d, connPG)
	})

	select {}
}

// insertToDB inserts a Rabbit message into a Postgres database.
func insertToDB(d amqp.Delivery, c *pgx.Conn) bool {
	var message models.Message
	err := json.Unmarshal(d.Body, &message)
	if err != nil {
		log.Fatalf("unmarshal message: %s", err)
	}

	_, err = c.Exec(context.Background(), "insert into messages (message, created) values ($1, to_timestamp($2))", message.Text, message.Time/1000)
	if err != nil {
		log.Fatalf("insert into database: %s", err)
	}

	// An alternative query that returns the id of the inserted row.
	// var id int64
	// err = c.QueryRow(context.Background(), "insert into messages (message, created) values ($1, to_timestamp($2)) returning id", message.Text, message.Time/1000).Scan(&id)
	// if err != nil {
	// 	log.Fatalf("insert into database: %s", err)
	// }
	// fmt.Println(id)

	// For cache, could send messages from here instead
	// of doing it from the server and backend services.
	// err = <Rabbit conn>.Publish(conf.Exchange, conf.KeyCache, d.Body)

	return true
}

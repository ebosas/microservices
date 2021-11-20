package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ebosas/microservices/internal/config"
	"github.com/ebosas/microservices/internal/models"
	"github.com/ebosas/microservices/internal/rabbit"

	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
)

var conf = config.New()

func main() {
	fmt.Println("[Database service]")

	// Postgres connection
	connPG, err := sql.Open("postgres", conf.PostgresURL+"?sslmode=disable")
	if err != nil {
		log.Fatalf("postgres connection: %s", err)
	}
	defer connPG.Close()

	// The table is not created when deployed on AWS RDS.
	_, err = connPG.Exec("create table if not exists messages (id serial primary key, message text not null, created timestamp not null)")
	if err != nil {
		log.Fatalf("create table: %s", err)
	}

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
func insertToDB(d amqp.Delivery, c *sql.DB) bool {
	var message models.Message
	err := json.Unmarshal(d.Body, &message)
	if err != nil {
		log.Fatalf("unmarshal message: %s", err)
	}

	_, err = c.Exec("insert into messages (message, created) values ($1, to_timestamp($2))", message.Text, message.Time/1000)
	if err != nil {
		log.Fatalf("insert into database: %s", err)
	}

	return true
}

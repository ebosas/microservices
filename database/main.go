package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ebosas/microservices/config"
	"github.com/jackc/pgx/v4"
	"github.com/streadway/amqp"
)

type Message struct {
	Text   string `json:"text"`
	Source string `json:"source"`
	Time   int64  `json:"time"`
}

var conf = config.New()

func main() {
	// Postgres connection
	connP, err := pgx.Connect(context.Background(), conf.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %s", err)
	}
	defer connP.Close(context.Background())

	// Amqp connection
	connA, err := amqp.Dial(conf.RabbitURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer connA.Close()

	ch, err := connA.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		conf.Exchange, // name
		"topic",       // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare an exchange: %s", err)
	}

	_, err = ch.QueueDeclare(
		conf.QueueDB, // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	err = ch.QueueBind(
		conf.QueueDB,                      // queue name
		fmt.Sprintf("#.%s.#", conf.KeyDB), // routing key
		conf.Exchange,                     // exchange
		false,                             // no-wait
		nil,                               // arguments
	)
	if err != nil {
		log.Fatalf("Failed to bind a queue: %s", err)
	}

	msgs, err := ch.Consume(
		conf.QueueDB, // queue name
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %s", err)
	}

	for msg := range msgs {
		var message Message
		err := json.Unmarshal(msg.Body, &message)
		if err != nil {
			log.Printf("Failed to unmarshal a message: %s", err)
			break
		}

		// Insert a message from Rabbit to Postgres
		_, err = connP.Exec(context.Background(), "insert into messages (message, created) values ($1, to_timestamp($2))", message.Text, message.Time/1000)
		if err != nil {
			log.Printf("Failed to insert into a database: %s", err)
			break
		}

		msg.Ack(false)
	}
}

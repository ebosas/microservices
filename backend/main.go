package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ebosas/microservices/config"
	"github.com/streadway/amqp"
)

type Message struct {
	Text   string `json:"text"`
	Source string `json:"source"`
	Time   int64  `json:"time"`
}

var conf = config.New()

func main() {
	log.SetFlags(0)

	conn, err := amqp.Dial(conf.RabbitURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
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
		conf.QueueBack, // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	err = ch.QueueBind(
		conf.QueueBack,                      // queue name
		fmt.Sprintf("#.%s.#", conf.KeyBack), // routing key
		conf.Exchange,                       // exchange
		false,                               // no-wait
		nil,                                 // arguments
	)
	if err != nil {
		log.Fatalf("Failed to bind a queue: %s", err)
	}

	msgs, err := ch.Consume(
		conf.QueueBack, // queue name
		"",             // consumer
		false,          // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %s", err)
	}

	go func() {
		for msg := range msgs {
			var message Message
			err := json.Unmarshal(msg.Body, &message)
			if err != nil {
				log.Fatalf("Failed to unmarshal a message: %s", err)
			}

			log.Printf("[Received] %s", string(message.Text))
			msg.Ack(false)
		}
	}()

	publishInput(conn)
}

// publishInput reads user input from stdin,
// marshals as json messages, and publishes them
// to a RabbitMQ exchange
func publishInput(c *amqp.Connection) {
	ch, err := c.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer ch.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		message, err := json.Marshal(
			Message{
				input,
				"back",
				time.Now().UnixNano() / int64(1e6),
			},
		)
		if err != nil {
			log.Fatalf("Failed to marshal a message: %s", err)
		}

		err = ch.Publish(
			conf.Exchange,                // exchane name
			conf.KeyFront+"."+conf.KeyDB, // routing key
			false,                        // mandatory
			false,                        // immediate
			amqp.Publishing{
				Timestamp:   time.Now(),
				ContentType: "text/plain",
				Body:        message,
			},
		)
		if err != nil {
			log.Fatalf("Failed to publish a message: %s", err)
		}

	}
}

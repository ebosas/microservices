package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ebosas/microservices/internal/config"
	"github.com/ebosas/microservices/internal/models"
	"github.com/ebosas/microservices/internal/rabbit"
	"github.com/streadway/amqp"
)

var conf = config.New()

func main() {
	fmt.Println("[Backend service]")

	// Establish a Rabbit connection.
	conn, err := rabbit.GetConn(conf.RabbitURL)
	if err != nil {
		log.Fatalf("rabbit connection: %s", err)
	}
	defer conn.Close()

	err = conn.DeclareTopicExchange(conf.Exchange)
	if err != nil {
		log.Fatalf("declare exchange: %s", err)
	}

	// Start a Rabbit consumer with a handler for printing messages.
	conn.StartConsumer(conf.Exchange, conf.QueueBack, conf.KeyBack, printMessages)

	publishInput(conn)
}

// printMessages prints messages to stdout.
func printMessages(d amqp.Delivery) bool {
	var message models.Message
	err := json.Unmarshal(d.Body, &message)
	if err != nil {
		log.Fatalf("unmarshal message: %s", err)
	}

	fmt.Printf("> %s\n", string(message.Text))

	return true
}

// publishInput reads user input, marshals to json, and publishes to
// a Rabbit exchange with the front-end and database routing keys.
func publishInput(c *rabbit.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		inputTime := time.Now().UnixNano() / int64(1e6) // in miliseconds
		inputMsg := models.Message{Text: input, Source: "back", Time: inputTime}
		message, err := json.Marshal(inputMsg)
		if err != nil {
			log.Fatalf("marshal message: %s", err)
		}

		key := conf.KeyFront + "." + conf.KeyDB + "." + conf.KeyCache
		err = c.Publish(conf.Exchange, key, message)
		if err != nil {
			log.Fatalf("publish message: %s", err)
		}
	}
}

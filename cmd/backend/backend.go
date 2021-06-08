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
	fmt.Println("Running backend service")

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

	// Start a Rabbit consumer with a message processing handler.
	conn.StartConsumer(conf.Exchange, conf.QueueBack, conf.KeyBack, receiveMessages)

	publishInput(conn)
}

// receiveMessages prints messages to stdout.
func receiveMessages(d amqp.Delivery) bool {
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
func publishInput(conn *rabbit.Conn) {
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
		err = conn.Publish(conf.Exchange, conf.KeyFront+"."+conf.KeyDB, message)
		if err != nil {
			log.Fatalf("publish message: %s", err)
		}
	}
}

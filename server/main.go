package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/ebosas/microservices/config"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

//go:embed template.html
var files embed.FS

//go:embed static
var static embed.FS

var (
	conf     = config.New()
	upgrader = websocket.Upgrader{} // use default options
)

func main() {
	log.SetFlags(0)
	log.Print("Running server")

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
		log.Fatalf("Failed to declare a backend queue: %s", err)
	}

	err = ch.QueueBind(
		conf.QueueBack,                      // queue name
		fmt.Sprintf("#.%s.#", conf.KeyBack), // routing key
		conf.Exchange,                       // exchange
		false,                               // no-wait
		nil,                                 // arguments
	)
	if err != nil {
		log.Fatalf("Failed to bind a backend queue: %s", err)
	}

	http.Handle("/static/", http.FileServer(http.FS(static)))
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWs(w, r, conn)
	})
	log.Fatal(http.ListenAndServe(conf.ServerAddr, nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFS(files, "template.html")
	t.Execute(w, nil)
}

func handleWs(w http.ResponseWriter, r *http.Request, c *amqp.Connection) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %s", err)
		return
	}
	defer ws.Close()

	done := make(chan bool)

	go wsWriter(ws, c, done)
	go wsReader(ws, c, done)

	<-done
}

// wsWriter reads messages from RabbitMQ
// and writes to websocket
func wsWriter(ws *websocket.Conn, c *amqp.Connection, done chan bool) {
	defer func() {
		done <- true
	}()

	ch, err := c.Channel()
	if err != nil {
		log.Printf("Failed to open a channel: %s", err)
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Printf("Failed to create a frontend queue: %s", err)
		return
	}

	err = ch.QueueBind(
		q.Name,                               // queue name
		fmt.Sprintf("#.%s.#", conf.KeyFront), // routing key
		conf.Exchange,                        // exchange
		false,                                // no-wait
		nil,                                  // arguments
	)
	if err != nil {
		log.Printf("Failed to bind a frontend queue: %s", err)
		return
	}

	msgs, err := ch.Consume(
		q.Name, // queue name
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Printf("Failed to register a consumer: %s", err)
		return
	}

	for msg := range msgs {
		ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
		err = ws.WriteMessage(websocket.TextMessage, []byte(msg.Body))
		if err != nil {
			log.Printf("Failed to write to WebSocket: %s", err)
			break
		}
	}

}

// wsReader reads messages from websocket
// and publishes to RabbitMQ
func wsReader(ws *websocket.Conn, c *amqp.Connection, done chan bool) {
	defer func() {
		done <- true
	}()

	ch, err := c.Channel()
	if err != nil {
		log.Printf("Failed to open a channel: %s", err)
		return
	}
	defer ch.Close()

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Failed to read a message: %s", err)
			break
		}

		err = ch.Publish(
			conf.Exchange,               // exchane name
			conf.KeyBack+"."+conf.KeyDB, // routing key
			false,                       // mandatory
			false,                       // immediate
			amqp.Publishing{
				Timestamp:   time.Now(),
				ContentType: "text/plain",
				Body:        []byte(message),
			},
		)
		if err != nil {
			log.Printf("Failed to publish a message: %s", err)
			break
		}

	}
}

package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/ebosas/microservices/internal/config"
	"github.com/ebosas/microservices/internal/rabbit"
	iwebsocket "github.com/ebosas/microservices/internal/websocket"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

//go:embed template
var filesTempl embed.FS

//go:embed static
var filesStatic embed.FS

var (
	conf     = config.New()
	upgrader = websocket.Upgrader{} // use default options
)

func main() {
	fmt.Println("[Server]")

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

	http.Handle("/static/", http.FileServer(http.FS(filesStatic)))
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/ws", handleWebsocketConn(conn))
	log.Fatal(http.ListenAndServe(conf.ServerAddr, nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		handleNotFound(w)
		return
	}
	t, _ := template.ParseFS(filesTempl, "template/template.html")
	t.Execute(w, nil)
}

// handleWebsocketConn passes a Rabbit connection to the Websocket handler.
func handleWebsocketConn(conn *rabbit.Conn) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handleWebsocket(w, r, conn)
	}
}

// handleWebsocket starts two message readers (consumers), one consuming
// a Rabbit queue and another reading from a Websocket connection.
// Each consumer receives a message handler to relay messages –
// from Rabbit to Websocket and vice versa.
func handleWebsocket(w http.ResponseWriter, r *http.Request, conn *rabbit.Conn) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade websocket: %s", err)
		return
	}
	defer ws.Close()

	// A separate channel for a publisher in a go routine.
	ch, err := conn.Connection.Channel()
	if err != nil {
		log.Printf("open channel: %s", err)
		return
	}
	defer ch.Close()

	// done and cancel() makes sure all spawned go routines are
	// terminated if any one of them is finished.
	done := make(chan bool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start a Rabbit consumer
	err = conn.StartConsumerTemp(ctx, done, conf.Exchange, conf.KeyFront, handleWriteWebsocket(ws))
	if err != nil {
		log.Printf("start temp consumer: %s", err)
		return
	}

	// Start a websocket reader (consumer)
	err = iwebsocket.StartReader(ctx, done, ws, handlePublishRabbit(ch))
	if err != nil {
		log.Printf("start websocket reader: %s", err)
		return
	}

	<-done
}

// handleWriteWebsocket writes a Rabbit message to Websocket.
// A Rabbit consumer only passes a message. So, a Websocket connection is
// additionally passed using a closure.
func handleWriteWebsocket(ws *websocket.Conn) func(d amqp.Delivery) error {
	return func(d amqp.Delivery) error {
		ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
		// TODO: check msg
		err := ws.WriteMessage(websocket.TextMessage, []byte(d.Body))
		if err != nil {
			return fmt.Errorf("write websocket: %v", err)
		}
		return nil
	}
}

// handlePublishRabbit publishes a Websocket message to a Rabbit
// exchange with the the back-end and database routing keys.
// A Websocket reader only passes a message. So, a Rabbit channel is
// additionally passed using a closure.
func handlePublishRabbit(ch *amqp.Channel) func(msg []byte) error {
	return func(msg []byte) error {
		// TODO: check msg
		err := rabbit.PublishInChannel(ch, conf.Exchange, conf.KeyBack+"."+conf.KeyDB, msg)
		if err != nil {
			// TODO: log error
			return fmt.Errorf("publish rabbit: %v", err)
		}
		return nil
	}
}

// handleNotFound handles 404
func handleNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	t, _ := template.ParseFS(filesTempl, "template/404.html")
	t.Execute(w, nil)
}

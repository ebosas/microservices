package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/ebosas/microservices/internal/cache"
	"github.com/ebosas/microservices/internal/config"
	"github.com/ebosas/microservices/internal/rabbit"
	iwebsocket "github.com/ebosas/microservices/internal/websocket"
	"github.com/go-redis/redis/v8"
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
	connMQ, err := rabbit.GetConn(conf.RabbitURL)
	if err != nil {
		log.Fatalf("rabbit connection: %s", err)
	}
	defer connMQ.Close()

	err = connMQ.DeclareTopicExchange(conf.Exchange)
	if err != nil {
		log.Fatalf("declare exchange: %s", err)
	}

	// Redis connection
	connR := redis.NewClient(&redis.Options{
		Addr:     conf.RedisURL,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	http.Handle("/static/", http.FileServer(http.FS(filesStatic)))
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/messages", handleMessages(connR))
	http.HandleFunc("/ws", handleWebsocketConn(connMQ))
	http.HandleFunc("/api/cache", handleAPICache(connR)) // defined in api.go
	log.Fatal(http.ListenAndServe(conf.ServerAddr, nil))
}

// handleHome handles the home page.
func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		handleNotFound(w)
		return
	}
	t := template.Must(template.ParseFS(filesTempl, "template/template.html", "template/navbar.html", "template/home.html"))
	t.ExecuteTemplate(w, "layout", map[string]string{"Page": "home"})
}

// handleMessages handles the messages page.
func handleMessages(cr *redis.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		cacheData, err := cache.GetCache(cr)
		if err != nil {
			log.Printf("get cache: %s", err)
			return
		}

		cacheJSON, err := json.Marshal(cacheData)
		if err != nil {
			log.Printf("marshal cache: %s", err)
			return
		}

		data := map[string]interface{}{
			"Data": cacheData,
			"Json": string(cacheJSON),
			"Page": "messages",
		}

		// funcMap := template.FuncMap{"ftime": timeutil.FormatDuration}
		t := template.Must(template.New(""). /*Funcs(funcMap).*/ ParseFS(filesTempl, "template/template.html", "template/navbar.html", "template/messages.html"))
		t.ExecuteTemplate(w, "layout", data)
	}
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
		key := conf.KeyBack + "." + conf.KeyDB + "." + conf.KeyCache
		err := rabbit.PublishInChannel(ch, conf.Exchange, key, msg)
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

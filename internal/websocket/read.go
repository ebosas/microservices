package websocket

import (
	"context"

	"github.com/gorilla/websocket"
)

// StartReader reads messages from a Websocket connection and passes
// them to a supplied handler for processing.
func StartReader(ctx context.Context, done chan<- bool, ws *websocket.Conn, handler func([]byte) error) error {
	msgs := make(chan []byte)
	go func() {
	Reader:
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				// log.Printf("read websocket: %s", err)
				done <- true
				break Reader
			}
			msgs <- message
		}
	}()
	go func() {
	Consumer:
		for {
			select {
			case msg := <-msgs:
				err := handler(msg)
				if err != nil {
					done <- true
					break Consumer
				}
			case <-ctx.Done():
				break Consumer
			}
		}
	}()
	return nil
}

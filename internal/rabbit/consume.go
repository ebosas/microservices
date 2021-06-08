package rabbit

import (
	"context"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// StartConsumer consumes messages from a Rabbit queue with a specified
// routing key and passes them to a supplied handler for processing.
// The queue is created (or connected to, if exists) and bound to an exchange.
// Used for durable queues in the main go routine.
func (conn *Conn) StartConsumer(exch, qName, rKey string, handler func(amqp.Delivery) bool) error {
	// Declare a durable queue
	_, err := conn.Channel.QueueDeclare(qName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("queue declare: %v", err)
	}

	err = conn.Channel.QueueBind(qName, "#."+rKey+".#", exch, false, nil)
	if err != nil {
		return fmt.Errorf("queue bind: %v", err)
	}

	// Set prefetchCount above zero to limit unacknowledged messages.
	err = conn.Channel.Qos(0, 0, false)
	if err != nil {
		return err
	}

	// Consume with explicit ack
	msgs, err := conn.Channel.Consume(qName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %v", err)
	}

	go func() {
		for msg := range msgs {
			if handler(msg) {
				msg.Ack(false)
			} else {
				msg.Nack(false, true)
			}
		}
		log.Fatalf("consumer closed")
	}()

	return nil
}

// StartConsumerTemp consumes messages with a specified routing key
// and passes them to a supplied handler for processing.
// Creates a separate channel and a temporary queue that will be deleted
// when processing ends (i.e. Websocket connection closes).
// Used in go routines such as each Websocket handler established
// by a front end user.
func (conn *Conn) StartConsumerTemp(ctx context.Context, done chan<- bool, exch, rKey string, handler func(amqp.Delivery) error) error {
	// A separate channel for a consumer in a go routine
	ch, err := conn.Connection.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %v", err)
	}

	// Declare a non-durable, auto-deleted, exlusive queue with
	// a generated name.
	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		return fmt.Errorf("queue declare: %v", err)
	}

	err = ch.QueueBind(q.Name, "#."+rKey+".#", exch, false, nil)
	if err != nil {
		return fmt.Errorf("queue bind: %v", err)
	}

	// Consume with auto-ack
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %v", err)
	}

	go func() {
		defer ch.Close()
	Consumer:
		for {
			select {
			case msg := <-msgs:
				if err := handler(msg); err != nil {
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

package rabbit

import (
	"time"

	"github.com/streadway/amqp"
)

// Publish publishes a message to a Rabbit exchange using the main channel.
// For use in the main go routine.
func (conn Conn) Publish(exch, rKey string, message []byte) error {
	return PublishInChannel(conn.Channel, exch, rKey, message)
}

// PublishInChannel publishes a message to a Rabbit exchange using
// a provided channel. For use in go routines.
func PublishInChannel(ch *amqp.Channel, exch, rKey string, message []byte) error {
	return ch.Publish(
		exch,  // exchane name
		rKey,  // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         message,
		},
	)
}

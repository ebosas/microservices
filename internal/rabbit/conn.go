package rabbit

import "github.com/streadway/amqp"

// Conn returns a Rabbit connecton. Also, a channel to be used
// in the main go routine.
type Conn struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

// GetConn established a Rabbit connection.
func GetConn(rabbitURL string) (*Conn, error) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return &Conn{}, err
	}
	ch, err := conn.Channel()
	return &Conn{
		Connection: conn,
		Channel:    ch,
	}, err
}

// Close closes the Rabbit connection.
// All resources associated with the connection, including channels,
// will also be closed.
func (conn *Conn) Close() error {
	return conn.Connection.Close()
}

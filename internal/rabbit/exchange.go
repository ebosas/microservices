package rabbit

// DeclareTopicExchange declares an exchange of type 'topic'.
func (conn *Conn) DeclareTopicExchange(name string) error {
	return conn.Channel.ExchangeDeclare(
		name,    // name
		"topic", // type
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
}

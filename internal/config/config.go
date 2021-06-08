package config

import "os"

type Config struct {
	ServerAddr  string
	PostgresURL string
	RabbitURL   string
	Exchange    string // Rabbit exchange name
	QueueBack   string // queue name
	QueueDB     string
	KeyFront    string // routing key name
	KeyBack     string
	KeyDB       string
}

// New returns configuration variables from the environment.
// These are passed by Docker from the .env file.
func New() *Config {
	return &Config{
		ServerAddr:  getEnv("SERVER_ADDR", "localhost:8080"),
		PostgresURL: getEnv("POSTGRES_URL", "postgres://postgres:demopsw@localhost:5432/microservices"),
		RabbitURL:   getEnv("RABBIT_URL", "amqp://guest:guest@localhost:5672"),
		Exchange:    getEnv("EXCHANGE", "main_exchange"),
		QueueBack:   getEnv("QUEUE_BACK", "backend_queue"),
		QueueDB:     getEnv("QUEUE_DB", "db_queue"),
		KeyFront:    getEnv("KEY_FRONT", "frontend_key"),
		KeyBack:     getEnv("KEY_BACK", "backend_key"),
		KeyDB:       getEnv("KEY_DB", "db_key"),
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

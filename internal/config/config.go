package config

import "os"

// Config stores configuration
type Config struct {
	// Connection strings
	ServerAddr  string
	PostgresURL string
	RabbitURL   string
	RedisURL    string
	// Rabbit exchange name
	Exchange string
	// Queue names
	QueueBack  string
	QueueDB    string
	QueueCache string
	// Routing key names
	KeyFront string
	KeyBack  string
	KeyDB    string
	KeyCache string
}

// New returns configuration variables from the environment.
// These are passed by Docker from the .env file.
func New() *Config {
	return &Config{
		ServerAddr:  getEnv("SERVER_ADDR", "localhost:8080"),
		PostgresURL: getEnv("POSTGRES_URL", "postgres://postgres:demopsw@localhost:5432/microservices"),
		RabbitURL:   getEnv("RABBIT_URL", "amqp://guest:guest@localhost:5672"),
		RedisURL:    getEnv("REDIS_URL", "localhost:6379"),
		Exchange:    getEnv("EXCHANGE", "main_exchange"),
		QueueBack:   getEnv("QUEUE_BACK", "backend_queue"),
		QueueDB:     getEnv("QUEUE_DB", "db_queue"),
		QueueCache:  getEnv("QUEUE_CACHE", "cache_queue"),
		KeyFront:    getEnv("KEY_FRONT", "frontend_key"),
		KeyBack:     getEnv("KEY_BACK", "backend_key"),
		KeyDB:       getEnv("KEY_DB", "db_key"),
		KeyCache:    getEnv("KEY_CACHE", "cache_key"),
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

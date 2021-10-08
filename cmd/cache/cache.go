package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ebosas/microservices/internal/config"
	"github.com/ebosas/microservices/internal/rabbit"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

var conf = config.New()

var ctx = context.Background()

func main() {
	fmt.Println("[Cache service]")

	// Redis connection
	connR := redis.NewClient(&redis.Options{
		Addr:     conf.RedisURL,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// RabbitMQ connection
	connMQ, err := rabbit.GetConn(conf.RabbitURL)
	if err != nil {
		log.Fatalf("rabbit connection: %s", err)
	}
	defer connMQ.Close()

	err = connMQ.DeclareTopicExchange(conf.Exchange)
	if err != nil {
		log.Fatalf("declare exchange: %s", err)
	}

	// Start a Rabbit consumer with a message processing handler.
	connMQ.StartConsumer(conf.Exchange, conf.QueueCache, conf.KeyCache, func(d amqp.Delivery) bool {
		return updateRedis(d, connR)
	})

	select {}
}

// updateRedis updates Redis with a new Rabbit message.
func updateRedis(d amqp.Delivery, c *redis.Client) bool {
	// Add a message, limit to 10 in cache, increment total count.
	if _, err := c.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.LPush(ctx, "messages", d.Body)
		pipe.LTrim(ctx, "messages", 0, 9)
		pipe.Incr(ctx, "total")
		return nil
	}); err != nil {
		log.Fatalf("update redis: %s", err)
	}

	return true
}

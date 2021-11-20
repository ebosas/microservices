package main

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/ebosas/microservices/internal/models"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

// TestUpdateRedis tests a message insertion into cache
func TestUpdateRedis(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis connection: %s", err)
	}
	defer s.Close()

	c := redis.NewClient(&redis.Options{Addr: s.Addr()})

	now := time.Now().UnixMilli()
	var tests = []struct {
		message string
		source  string
		time    int64
	}{
		{"Hello", "back", now},
		{"Another test!", "back", now},
		{"1", "front", now},
		{" ", "back", now - 60*60*1000},
	}
	for _, test := range tests {
		d := testArguments(t, test.message, test.source, test.time)
		updateRedis(*d, c)
	}

	if got, err := s.Get("total"); err != nil || got != strconv.Itoa(len(tests)) {
		t.Error("'total' has the wrong value")
	}

	list, err := s.List("messages")
	if err != nil {
		t.Errorf("list 'messages': %s", err)
	}

	if len(list) != len(tests) {
		t.Error("'messages' has wrong length")
	}

	// TODO: compare each message
}

// testArguments produces arguments for the function being tested
func testArguments(t *testing.T, message, source string, time int64) *amqp.Delivery {
	inputMsg := models.Message{Text: message, Source: source, Time: time}
	messageJson, err := json.Marshal(inputMsg)
	if err != nil {
		t.Fatalf("marshal message: %s", err)
	}
	d := &amqp.Delivery{Body: messageJson}

	return d
}

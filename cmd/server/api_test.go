package main

// http.HandleFunc("/api/cache", handleAPICache(connR))

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/ebosas/microservices/internal/models"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// TestAPICache tests the cache API
func TestAPICache(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis connection: %s", err)
	}
	defer s.Close()

	c := redis.NewClient(&redis.Options{Addr: s.Addr()})

	testMsg := "This is a test!"
	testUpdateRedis(t, c, testMsg)

	req := httptest.NewRequest(http.MethodGet, "/api/cache", nil)
	w := httptest.NewRecorder()
	handler := handleAPICache(c)
	handler(w, req)
	res := w.Result()
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	if !strings.Contains(string(data), testMsg) {
		t.Errorf("test message not found: %q", testMsg)
	}
}

// testUpdateRedis inserts a marshalled message into mock redis
func testUpdateRedis(t *testing.T, c *redis.Client, message string) {
	time := time.Now().UnixMilli()
	inputMsg := models.Message{Text: message, Source: "back", Time: time}
	messageJson, err := json.Marshal(inputMsg)
	if err != nil {
		t.Fatalf("marshal message: %s", err)
	}

	if _, err := c.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.LPush(ctx, "messages", messageJson)
		pipe.LTrim(ctx, "messages", 0, 9)
		pipe.Incr(ctx, "total")
		return nil
	}); err != nil {
		t.Fatalf("update redis: %s", err)
	}
}

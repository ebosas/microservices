package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ebosas/microservices/internal/models"
	"github.com/ebosas/microservices/internal/timeutil"
	"github.com/go-redis/redis/v8"
)

// Message adds additional fields to models.Message.
type Message struct {
	models.Message
	TimeFmt string `json:"timefmt"`
}

// Cache is used in the messages.html template.
type Cache struct {
	Count    int64     `json:"count"`
	Total    int64     `json:"total"`
	Messages []Message `json:"messages"`
}

// GetCache gets cached data from Redis.
func GetCache(c *redis.Client) (*Cache, error) {
	messages, err := c.LRange(context.Background(), "messages", 0, -1).Result()
	if err != nil {
		return &Cache{}, fmt.Errorf("lrange redis: %v", err)
	}

	total, err := c.Get(context.Background(), "total").Int64()
	if err == redis.Nil {
		total = 0
	} else if err != nil {
		return &Cache{}, fmt.Errorf("get redis: %v", err)
	}

	msgsCache := make([]Message, 0) // avoid null in JSON when empty
	for _, messageJSON := range messages {
		var message models.Message
		err = json.Unmarshal([]byte(messageJSON), &message)
		if err != nil {
			return &Cache{}, fmt.Errorf("unmarshal cache: %v", err)
		}

		msgsCache = append(msgsCache, Message{
			Message: message,
			TimeFmt: timeutil.FormatDuration(message.Time),
		})
	}

	cache := &Cache{
		Count:    int64(len(messages)),
		Total:    total,
		Messages: msgsCache,
	}

	return cache, nil
}

// GetCacheJSON marshals cached data into JSON,
// calls GetCache to get the Cache struct.
func GetCacheJSON(c *redis.Client) (string, error) {
	cacheData, err := GetCache(c)
	if err != nil {
		return "", fmt.Errorf("get cache: %s", err)
	}

	cacheJSON, err := json.Marshal(cacheData)
	if err != nil {
		return "", fmt.Errorf("marshal cache: %s", err)

	}

	return string(cacheJSON), nil
}

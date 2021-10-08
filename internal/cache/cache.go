package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ebosas/microservices/internal/models"
	"github.com/go-redis/redis/v8"
)

// Cache is a data structure for the cache API template.
type Cache struct {
	Count    int              `json:"count"`
	Total    int              `json:"total"`
	Messages []models.Message `json:"messages"`
}

// GetCache gets cached messages from Redis,
// returns both a Cache struct and a JSON string.
func GetCache(c *redis.Client) (*Cache, string, error) {
	cacheJSON, err := GetCacheJSON(c)
	if err != nil {
		return &Cache{}, "", fmt.Errorf("get cache json: %v", err)
	}

	var cache Cache
	err = json.Unmarshal([]byte(cacheJSON), &cache)
	if err != nil {
		return &Cache{}, "", fmt.Errorf("unmarshal cache: %v", err)
	}

	return &cache, cacheJSON, nil
}

// GetCacheJSON reads cached messages from Redis, returns JSON.
func GetCacheJSON(c *redis.Client) (string, error) {
	messages, err := c.LRange(context.Background(), "messages", 0, -1).Result()
	if err != nil {
		return "", fmt.Errorf("lrange redis: %v", err)
	}

	total, err := c.Get(context.Background(), "total").Result()
	if err == redis.Nil {
		total = "0"
	} else if err != nil {
		return "", fmt.Errorf("get redis: %v", err)
	}

	cacheJSON := "{\"count\":" + fmt.Sprint(len(messages)) + ",\"total\":" + total + ",\"messages\":[" + strings.Join(messages, ",") + "]}"

	return cacheJSON, nil
}
